// Copyright 2025.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package domain

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// CertStatus represents the validity state of an SSH certificate.
type CertStatus string

const (
	CertStatusValid        CertStatus = "valid"
	CertStatusExpiringSoon CertStatus = "expiring_soon"
	CertStatusExpired      CertStatus = "expired"
	CertStatusNotYetValid  CertStatus = "not_yet_valid"
)

// SSHCertificate encapsulates parsed OpenSSH certificate details and its live validity status.
type SSHCertificate struct {
	Path           string        // Path to certificate file on disk (tilde or expanded)
	KeyID          string        // Key ID / principal identity comment
	Principals     []string      // Valid principals (e.g. root, ubuntu)
	ValidAfter     time.Time     // Validity start time
	ValidBefore    time.Time     // Validity end time (or infinite)
	Lifetime       time.Duration // Total validity window (ValidBefore - ValidAfter)
	Serial         uint64        // Certificate serial number
	CertType       string        // "user" or "host"
	Status         CertStatus    // Current evaluated status (valid, expiring_soon, expired, not_yet_valid)
	TimeRemaining  time.Duration // Remaining validity if valid or expiring soon
	TimeExpiredAgo time.Duration // Time elapsed since expiration if expired
	IsImplicit     bool          // True if discovered via <IdentityFile>-cert.pub rather than explicit CertificateFile
	FileExists     bool          // True if certificate file exists on disk
}

// ResolveServerCertificatePath discovers the certificate file for a server.
// Priority:
// 1. Explicit CertificateFile configured on the server.
// 2. Implicit <IdentityFile>-cert.pub matching OpenSSH behavior for each configured IdentityFile.
func ResolveServerCertificatePath(server Server) (path string, isImplicit bool) {
	if server.CertificateFile != "" {
		return server.CertificateFile, false
	}

	for _, idFile := range server.IdentityFiles {
		if idFile == "" {
			continue
		}
		// OpenSSH standard implicit cert pattern: <IdentityFile>-cert.pub
		candidate := idFile + "-cert.pub"
		expanded := ExpandTilde(candidate)
		if fi, err := os.Stat(expanded); err == nil && !fi.IsDir() {
			return candidate, true
		}
	}

	return "", false
}

// ParseSSHCertificateFile reads and parses an OpenSSH certificate from disk at the given path.
func ParseSSHCertificateFile(path string, isImplicit bool, now ...time.Time) (*SSHCertificate, error) {
	expanded := ExpandTilde(path)
	data, err := os.ReadFile(filepath.Clean(expanded)) // #nosec G304 -- User requested SSH certificate file read
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file %q: %w", path, err)
	}

	currentTime := time.Now()
	if len(now) > 0 {
		currentTime = now[0]
	}

	return ParseSSHCertificateBytes(data, path, isImplicit, currentTime)
}

// ParseSSHCertificateBytes parses OpenSSH certificate data.
func ParseSSHCertificateBytes(data []byte, path string, isImplicit bool, now time.Time) (*SSHCertificate, error) {
	//nolint:dogsled // OpenSSH authorized key format returns multiple metadata fields
	pubKey, _, _, _, err := ssh.ParseAuthorizedKey(data)
	if err != nil {
		return nil, fmt.Errorf("invalid SSH certificate format: %w", err)
	}

	cert, ok := pubKey.(*ssh.Certificate)
	if !ok {
		return nil, fmt.Errorf("key in %q is a regular public key, not an SSH certificate", path)
	}

	var validAfter, validBefore time.Time
	if cert.ValidAfter > 0 {
		if cert.ValidAfter <= uint64(math.MaxInt64) {
			validAfter = time.Unix(int64(cert.ValidAfter), 0)
		} else {
			validAfter = time.Unix(math.MaxInt64, 0)
		}
	}

	isInfinite := cert.ValidBefore == ssh.CertTimeInfinity || cert.ValidBefore > uint64(math.MaxInt64)
	if !isInfinite && cert.ValidBefore > 0 {
		if cert.ValidBefore <= uint64(math.MaxInt64) {
			validBefore = time.Unix(int64(cert.ValidBefore), 0)
		} else {
			validBefore = time.Unix(math.MaxInt64, 0)
		}
	}

	certTypeStr := "user"
	if cert.CertType == ssh.HostCert {
		certTypeStr = "host"
	}

	res := &SSHCertificate{
		Path:        path,
		KeyID:       cert.KeyId,
		Principals:  cert.ValidPrincipals,
		ValidAfter:  validAfter,
		ValidBefore: validBefore,
		Serial:      cert.Serial,
		CertType:    certTypeStr,
		IsImplicit:  isImplicit,
		FileExists:  true,
	}

	if isInfinite {
		res.Status = CertStatusValid
		res.Lifetime = time.Duration(math.MaxInt64)
		res.TimeRemaining = time.Duration(math.MaxInt64)
		return res, nil
	}

	if !validBefore.IsZero() && !validAfter.IsZero() && validBefore.After(validAfter) {
		res.Lifetime = validBefore.Sub(validAfter)
	}

	status, remaining, expiredAgo := EvaluateCertStatus(validAfter, validBefore, res.Lifetime, now)
	res.Status = status
	res.TimeRemaining = remaining
	res.TimeExpiredAgo = expiredAgo

	return res, nil
}

// EvaluateCertStatus evaluates whether the certificate is valid, expiring soon, expired, or not yet valid.
// The "expiring soon" threshold is relative to the certificate's lifetime (e.g. 10% of lifetime,
// capped at a maximum of 24h for long-lived certificates, and scaling down gracefully for short-lived ones).
func EvaluateCertStatus(validAfter, validBefore time.Time, lifetime time.Duration, now time.Time) (CertStatus, time.Duration, time.Duration) {
	if validBefore.IsZero() {
		return CertStatusValid, 0, 0
	}

	if !validAfter.IsZero() && now.Before(validAfter) {
		return CertStatusNotYetValid, validBefore.Sub(now), 0
	}

	if now.After(validBefore) || now.Equal(validBefore) {
		return CertStatusExpired, 0, now.Sub(validBefore)
	}

	remaining := validBefore.Sub(now)

	// Calculate proportional threshold based on certificate lifetime
	threshold := CalculateExpiringSoonThreshold(lifetime)

	if remaining <= threshold {
		return CertStatusExpiringSoon, remaining, 0
	}

	return CertStatusValid, remaining, 0
}

// CalculateExpiringSoonThreshold calculates the relative threshold for "expiring soon".
// A 24h warning makes sense for a 30-day certificate, not for a 1h one.
// Rule:
// - 10% of the total lifetime.
// - Capped at a maximum of 24 hours.
// - Minimum of 1 minute if total lifetime is >= 10 minutes.
func CalculateExpiringSoonThreshold(lifetime time.Duration) time.Duration {
	if lifetime <= 0 {
		return 5 * time.Minute
	}

	threshold := lifetime / 10

	if threshold > 24*time.Hour {
		threshold = 24 * time.Hour
	} else if threshold < 1*time.Minute && lifetime >= 10*time.Minute {
		threshold = 1 * time.Minute
	}

	return threshold
}

// FormatDuration formats a time duration into a compact human-readable string (e.g. "4d 12h", "2h 15m", "8m").
func FormatDuration(d time.Duration) string {
	if d < 0 {
		d = -d
	}

	if d >= 48*time.Hour {
		days := int(d / (24 * time.Hour))
		hours := int((d % (24 * time.Hour)) / time.Hour)
		if hours > 0 {
			return fmt.Sprintf("%dd %dh", days, hours)
		}
		return fmt.Sprintf("%dd", days)
	}

	if d >= 24*time.Hour {
		days := int(d / (24 * time.Hour))
		hours := int((d % (24 * time.Hour)) / time.Hour)
		return fmt.Sprintf("%dd %dh", days, hours)
	}

	if d >= 1*time.Hour {
		hours := int(d / time.Hour)
		mins := int((d % time.Hour) / time.Minute)
		if mins > 0 {
			return fmt.Sprintf("%dh %dm", hours, mins)
		}
		return fmt.Sprintf("%dh", hours)
	}

	if d >= 1*time.Minute {
		mins := int(d / time.Minute)
		return fmt.Sprintf("%dm", mins)
	}

	secs := int(d / time.Second)
	if secs <= 0 {
		return "< 1s"
	}
	return fmt.Sprintf("%ds", secs)
}

// FormatCertStatusBadge returns a UI-formatted badge and description for server details.
func FormatCertStatusBadge(cert *SSHCertificate) (badge string, description string) {
	if cert == nil {
		return "", ""
	}

	if !cert.FileExists {
		return "[red::b]⚠ Missing[-]", "file not found on disk"
	}

	switch cert.Status {
	case CertStatusValid:
		rem := FormatDuration(cert.TimeRemaining)
		return "[green::b]✓ Valid[-]", fmt.Sprintf("expires in %s", rem)
	case CertStatusExpiringSoon:
		rem := FormatDuration(cert.TimeRemaining)
		return "[yellow::b]⚠ Expiring soon[-]", fmt.Sprintf("expires in %s", rem)
	case CertStatusExpired:
		ago := FormatDuration(cert.TimeExpiredAgo)
		return "[red::b]✗ Expired[-]", fmt.Sprintf("expired %s ago", ago)
	case CertStatusNotYetValid:
		return "[blue::b]⏱ Not yet valid[-]", fmt.Sprintf("valid starting %s", cert.ValidAfter.Format("2006-01-02 15:04:05"))
	default:
		return "[dim]Unknown[-]", ""
	}
}

// InspectServerCertificate discovers and parses the certificate for the given server if available.
func InspectServerCertificate(server Server, now ...time.Time) *SSHCertificate {
	path, isImplicit := ResolveServerCertificatePath(server)
	if path == "" {
		return nil
	}

	expanded := ExpandTilde(path)
	if fi, err := os.Stat(expanded); err != nil || fi.IsDir() {
		return &SSHCertificate{
			Path:       path,
			IsImplicit: isImplicit,
			FileExists: false,
		}
	}

	cert, err := ParseSSHCertificateFile(path, isImplicit, now...)
	if err != nil {
		return &SSHCertificate{
			Path:       path,
			IsImplicit: isImplicit,
			FileExists: true,
			KeyID:      strings.TrimSpace(err.Error()),
		}
	}

	return cert
}
