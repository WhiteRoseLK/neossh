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

package ssh_config_file

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/WhiteRoseLK/neossh/internal/core/domain"
	"golang.org/x/crypto/ssh"
)

// ListKnownHostRecords returns all parsed entries in the given known_hosts file.
func (r *Repository) ListKnownHostRecords(knownHostsPath string) ([]domain.KnownHostRecord, error) {
	resolved := resolveKnownHostsPath(knownHostsPath)
	// #nosec G304 -- path is user-specified known_hosts file path
	file, err := os.Open(filepath.Clean(resolved))
	if err != nil {
		if os.IsNotExist(err) {
			return []domain.KnownHostRecord{}, nil
		}
		return nil, fmt.Errorf("failed to open known_hosts: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	var records []domain.KnownHostRecord
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		raw := scanner.Text()
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		fields := strings.Fields(trimmed)
		if len(fields) < 2 {
			continue
		}

		offset := 0
		if strings.HasPrefix(fields[0], "@") {
			offset = 1
		}
		if len(fields) < offset+2 {
			continue
		}

		hostPattern := fields[offset]
		keyType := fields[offset+1]
		var pubKeyBase64 string
		comment := ""
		if len(fields) > offset+2 {
			pubKeyBase64 = fields[offset+2]
		}
		if len(fields) > offset+3 {
			comment = strings.Join(fields[offset+3:], " ")
		}

		fingerprint := ""
		if pubKeyBase64 != "" {
			if keyBytes, err := base64.StdEncoding.DecodeString(pubKeyBase64); err == nil {
				if parsedKey, err := ssh.ParsePublicKey(keyBytes); err == nil {
					fingerprint = ssh.FingerprintSHA256(parsedKey)
				}
			}
		}

		records = append(records, domain.KnownHostRecord{
			LineNumber:  lineNum,
			HostPattern: hostPattern,
			KeyType:     keyType,
			Fingerprint: fingerprint,
			Comment:     comment,
			IsHashed:    strings.HasPrefix(hostPattern, "|1|"),
			RawLine:     raw,
		})
	}

	return records, scanner.Err()
}

// RemoveKnownHost deletes host entries matching host and port from known_hosts, creating a backup.
func (r *Repository) RemoveKnownHost(knownHostsPath, host string, port int) (string, int, error) {
	resolved := resolveKnownHostsPath(knownHostsPath)
	if _, err := os.Stat(resolved); os.IsNotExist(err) {
		return "", 0, fmt.Errorf("known_hosts file not found: %s", resolved)
	}

	backupPath, err := createKnownHostsBackup(resolved)
	if err != nil {
		return "", 0, fmt.Errorf("failed to create known_hosts backup: %w", err)
	}

	target := host
	if port > 0 && port != 22 {
		target = fmt.Sprintf("[%s]:%d", host, port)
	}

	// Try ssh-keygen -R first if available
	if sshKeygen, err := exec.LookPath("ssh-keygen"); err == nil && sshKeygen != "" {
		// #nosec G204 -- sshKeygen is trusted binary path from LookPath
		cmd := exec.Command(sshKeygen, "-R", target, "-f", resolved)
		_ = cmd.Run()
		if port > 0 && port != 22 {
			// #nosec G204 -- sshKeygen is trusted binary path from LookPath
			cmd2 := exec.Command(sshKeygen, "-R", host, "-f", resolved)
			_ = cmd2.Run()
		}
	}

	// Also perform safe direct line removal to handle edge cases or custom formats
	removed, err := purgeMatchingLines(resolved, host, target)
	if err != nil {
		return backupPath, removed, err
	}

	return backupPath, removed, nil
}

// RemoveKnownHostByLine removes the entry at the specified 1-based line number.
func (r *Repository) RemoveKnownHostByLine(knownHostsPath string, lineNumber int) (string, error) {
	resolved := resolveKnownHostsPath(knownHostsPath)
	// #nosec G304 -- resolved is user-specified known_hosts file path
	data, err := os.ReadFile(filepath.Clean(resolved))
	if err != nil {
		return "", fmt.Errorf("failed to read known_hosts: %w", err)
	}

	backupPath, err := createKnownHostsBackup(resolved)
	if err != nil {
		return "", fmt.Errorf("failed to create backup: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	if lineNumber < 1 || lineNumber > len(lines) {
		return backupPath, fmt.Errorf("line number %d out of range (1-%d)", lineNumber, len(lines))
	}

	newLines := make([]string, 0, len(lines)-1)
	for i, l := range lines {
		if i+1 == lineNumber {
			continue
		}
		newLines = append(newLines, l)
	}

	if err := os.WriteFile(resolved, []byte(strings.Join(newLines, "\n")), 0o600); err != nil {
		return backupPath, fmt.Errorf("failed to write updated known_hosts: %w", err)
	}

	return backupPath, nil
}

// ScanAndAddKnownHost connects to host:port, captures its host key, appends it to known_hosts, and returns the record.
func (r *Repository) ScanAndAddKnownHost(knownHostsPath, host string, port int) (*domain.KnownHostRecord, error) {
	if port <= 0 {
		port = 22
	}
	resolved := resolveKnownHostsPath(knownHostsPath)

	capturedKey, err := fetchRemoteHostKey(host, port, 4*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to scan host key from %s:%d: %w", host, port, err)
	}

	target := host
	if port != 22 {
		target = fmt.Sprintf("[%s]:%d", host, port)
	}

	b64Key := base64.StdEncoding.EncodeToString(capturedKey.Marshal())
	fp := ssh.FingerprintSHA256(capturedKey)
	rawLine := fmt.Sprintf("%s %s %s\n", target, capturedKey.Type(), b64Key)

	// Ensure parent dir exists
	if dir := filepath.Dir(resolved); dir != "" {
		_ = os.MkdirAll(dir, 0o700)
	}

	// #nosec G304 -- resolved is user-specified known_hosts file path
	f, err := os.OpenFile(filepath.Clean(resolved), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return nil, fmt.Errorf("failed to open known_hosts for writing: %w", err)
	}
	defer func() {
		_ = f.Close()
	}()

	if _, err := f.WriteString(rawLine); err != nil {
		return nil, fmt.Errorf("failed to write host key to known_hosts: %w", err)
	}

	return &domain.KnownHostRecord{
		HostPattern: target,
		KeyType:     capturedKey.Type(),
		Fingerprint: fp,
		RawLine:     strings.TrimSpace(rawLine),
	}, nil
}

func fetchRemoteHostKey(host string, port int, timeout time.Duration) (ssh.PublicKey, error) {
	var capturedKey ssh.PublicKey
	conf := &ssh.ClientConfig{
		User: "neossh-scan",
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			capturedKey = key
			return nil
		},
		Timeout: timeout,
	}

	addr := net.JoinHostPort(host, strconv.Itoa(port))
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = conn.Close()
	}()

	sshConn, chans, reqs, _ := ssh.NewClientConn(conn, addr, conf)
	_ = chans
	_ = reqs
	if sshConn != nil {
		_ = sshConn.Close()
	}

	if capturedKey == nil {
		return nil, fmt.Errorf("no public key received during SSH handshake")
	}
	return capturedKey, nil
}

func createKnownHostsBackup(khPath string) (string, error) {
	// #nosec G304 -- khPath is user-specified known_hosts file path
	data, err := os.ReadFile(filepath.Clean(khPath))
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	bakPath := khPath + ".old"
	if err := os.WriteFile(bakPath, data, 0o600); err != nil {
		return "", err
	}
	return bakPath, nil
}

func purgeMatchingLines(khPath, host, target string) (int, error) {
	// #nosec G304 -- khPath is user-specified known_hosts file path
	data, err := os.ReadFile(filepath.Clean(khPath))
	if err != nil {
		return 0, err
	}

	lines := strings.Split(string(data), "\n")
	var remaining []string
	removed := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			remaining = append(remaining, line)
			continue
		}

		fields := strings.Fields(trimmed)
		if len(fields) > 0 {
			pattern := fields[0]
			if strings.HasPrefix(pattern, "@") && len(fields) > 1 {
				pattern = fields[1]
			}
			tokens := strings.Split(pattern, ",")
			matches := false
			for _, tok := range tokens {
				tok = strings.TrimSpace(tok)
				if strings.EqualFold(tok, host) || strings.EqualFold(tok, target) {
					matches = true
					break
				}
			}
			if matches {
				removed++
				continue
			}
		}
		remaining = append(remaining, line)
	}

	if removed > 0 {
		if err := os.WriteFile(khPath, []byte(strings.Join(remaining, "\n")), 0o600); err != nil {
			return 0, err
		}
	}
	return removed, nil
}
