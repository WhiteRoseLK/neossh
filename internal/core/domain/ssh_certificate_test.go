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
	"crypto/ed25519"
	"crypto/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
)

func generateTestCertificate(t *testing.T, validAfter, validBefore uint64, keyID string, principals []string) ([]byte, ed25519.PublicKey) {
	t.Helper()

	// CA key
	caPub, caPriv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate CA key: %v", err)
	}
	caSigner, err := ssh.NewSignerFromKey(caPriv)
	if err != nil {
		t.Fatalf("failed to create CA signer: %v", err)
	}

	// User key
	userPub, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate user key: %v", err)
	}
	sshUserPub, err := ssh.NewPublicKey(userPub)
	if err != nil {
		t.Fatalf("failed to convert user pubkey: %v", err)
	}

	cert := &ssh.Certificate{
		Key:             sshUserPub,
		Serial:          42,
		CertType:        ssh.UserCert,
		KeyId:           keyID,
		ValidPrincipals: principals,
		ValidAfter:      validAfter,
		ValidBefore:     validBefore,
		Permissions: ssh.Permissions{
			Extensions: map[string]string{
				"permit-pty": "",
			},
		},
	}

	if err := cert.SignCert(rand.Reader, caSigner); err != nil {
		t.Fatalf("failed to sign test certificate: %v", err)
	}

	_ = caPub
	return ssh.MarshalAuthorizedKey(cert), userPub
}

func TestCalculateExpiringSoonThreshold(t *testing.T) {
	tests := []struct {
		name     string
		lifetime time.Duration
		expected time.Duration
	}{
		{
			name:     "30-day certificate capped at 24h",
			lifetime: 30 * 24 * time.Hour,
			expected: 24 * time.Hour,
		},
		{
			name:     "10-day certificate capped at 24h",
			lifetime: 10 * 24 * time.Hour,
			expected: 24 * time.Hour,
		},
		{
			name:     "7-day certificate",
			lifetime: 7 * 24 * time.Hour,
			expected: 16*time.Hour + 48*time.Minute,
		},
		{
			name:     "24-hour certificate",
			lifetime: 24 * time.Hour,
			expected: 2*time.Hour + 24*time.Minute,
		},
		{
			name:     "1-hour certificate is 6m, not 24h",
			lifetime: 1 * time.Hour,
			expected: 6 * time.Minute,
		},
		{
			name:     "10-minute certificate minimum 1m",
			lifetime: 10 * time.Minute,
			expected: 1 * time.Minute,
		},
		{
			name:     "5-minute certificate",
			lifetime: 5 * time.Minute,
			expected: 30 * time.Second,
		},
		{
			name:     "0 or negative fallback",
			lifetime: 0,
			expected: 5 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateExpiringSoonThreshold(tt.lifetime)
			if got != tt.expected {
				t.Errorf("CalculateExpiringSoonThreshold(%v) = %v, want %v", tt.lifetime, got, tt.expected)
			}
		})
	}
}

func TestEvaluateCertStatus(t *testing.T) {
	now := time.Date(2026, 9, 27, 12, 0, 0, 0, time.UTC)

	// 1-hour cert: 11:30 -> 12:30. Lifetime = 60m. Threshold = 6m.
	validAfter := now.Add(-30 * time.Minute)
	validBefore := now.Add(30 * time.Minute)
	lifetime := validBefore.Sub(validAfter)

	// Case 1: Valid (remaining 30m > 6m threshold)
	status, rem, expAgo := EvaluateCertStatus(validAfter, validBefore, lifetime, now)
	if status != CertStatusValid {
		t.Errorf("expected CertStatusValid, got %v", status)
	}
	if rem != 30*time.Minute || expAgo != 0 {
		t.Errorf("unexpected remaining/expiredAgo: %v, %v", rem, expAgo)
	}

	// Case 2: Expiring soon (now at 12:26, remaining 4m <= 6m threshold)
	nearExpiry := validBefore.Add(-4 * time.Minute)
	status, rem, _ = EvaluateCertStatus(validAfter, validBefore, lifetime, nearExpiry)
	if status != CertStatusExpiringSoon {
		t.Errorf("expected CertStatusExpiringSoon, got %v", status)
	}
	if rem != 4*time.Minute {
		t.Errorf("expected 4m remaining, got %v", rem)
	}

	// Case 3: Expired (now at 12:35, expired 5m ago)
	afterExpiry := validBefore.Add(5 * time.Minute)
	status, rem, expAgo = EvaluateCertStatus(validAfter, validBefore, lifetime, afterExpiry)
	if status != CertStatusExpired {
		t.Errorf("expected CertStatusExpired, got %v", status)
	}
	if expAgo != 5*time.Minute || rem != 0 {
		t.Errorf("expected 5m expired ago, got %v", expAgo)
	}

	// Case 4: Not yet valid (now before validAfter)
	beforeStart := validAfter.Add(-10 * time.Minute)
	status, _, _ = EvaluateCertStatus(validAfter, validBefore, lifetime, beforeStart)
	if status != CertStatusNotYetValid {
		t.Errorf("expected CertStatusNotYetValid, got %v", status)
	}
}

func TestParseSSHCertificateBytes(t *testing.T) {
	now := time.Now()
	after := uint64(now.Add(-10 * time.Minute).Unix())
	before := uint64(now.Add(50 * time.Minute).Unix())
	certBytes, _ := generateTestCertificate(t, after, before, "user@corp.internal", []string{"ubuntu", "admin"})

	cert, err := ParseSSHCertificateBytes(certBytes, "/tmp/id_test-cert.pub", false, now)
	if err != nil {
		t.Fatalf("unexpected error parsing certificate: %v", err)
	}

	if cert.KeyID != "user@corp.internal" {
		t.Errorf("expected KeyID 'user@corp.internal', got %q", cert.KeyID)
	}
	if len(cert.Principals) != 2 || cert.Principals[0] != "ubuntu" || cert.Principals[1] != "admin" {
		t.Errorf("unexpected principals: %v", cert.Principals)
	}
	if cert.Serial != 42 {
		t.Errorf("expected serial 42, got %d", cert.Serial)
	}
	if cert.CertType != "user" {
		t.Errorf("expected user cert, got %q", cert.CertType)
	}
	if cert.Status != CertStatusValid {
		t.Errorf("expected valid status, got %v", cert.Status)
	}

	badge, desc := FormatCertStatusBadge(cert)
	if !strings.Contains(badge, "Valid") {
		t.Errorf("expected valid badge, got %q", badge)
	}
	if !strings.Contains(desc, "expires in") {
		t.Errorf("expected expires in description, got %q", desc)
	}
}

func TestResolveServerCertificatePath(t *testing.T) {
	tempDir := t.TempDir()
	idKeyPath := filepath.Join(tempDir, "id_ed25519")
	idCertPath := filepath.Join(tempDir, "id_ed25519-cert.pub")

	if err := os.WriteFile(idKeyPath, []byte("dummy private key"), 0o600); err != nil {
		t.Fatal(err)
	}

	// 1. Neither CertificateFile nor implicit exists
	server := Server{
		Alias:         "host1",
		IdentityFiles: []string{idKeyPath},
	}
	path, isImplicit := ResolveServerCertificatePath(server)
	if path != "" || isImplicit {
		t.Errorf("expected no cert path, got %q (implicit=%v)", path, isImplicit)
	}

	// 2. Implicit cert file exists
	if err := os.WriteFile(idCertPath, []byte("dummy cert"), 0o644); err != nil {
		t.Fatal(err)
	}

	path, isImplicit = ResolveServerCertificatePath(server)
	if path != idCertPath || !isImplicit {
		t.Errorf("expected implicit cert %q, got %q (implicit=%v)", idCertPath, path, isImplicit)
	}

	// 3. Explicit CertificateFile overrides implicit
	explicitCert := filepath.Join(tempDir, "custom-cert.pub")
	server.CertificateFile = explicitCert
	path, isImplicit = ResolveServerCertificatePath(server)
	if path != explicitCert || isImplicit {
		t.Errorf("expected explicit cert %q, got %q (implicit=%v)", explicitCert, path, isImplicit)
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		d        time.Duration
		expected string
	}{
		{50 * time.Hour, "2d 2h"},
		{48 * time.Hour, "2d"},
		{25 * time.Hour, "1d 1h"},
		{2 * time.Hour, "2h"},
		{2*time.Hour + 15*time.Minute, "2h 15m"},
		{45 * time.Minute, "45m"},
		{30 * time.Second, "30s"},
		{0, "< 1s"},
	}

	for _, tt := range tests {
		got := FormatDuration(tt.d)
		if got != tt.expected {
			t.Errorf("FormatDuration(%v) = %q, want %q", tt.d, got, tt.expected)
		}
	}
}
