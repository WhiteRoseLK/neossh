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
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestKnownHostsOperations(t *testing.T) {
	tempDir := t.TempDir()
	khPath := filepath.Join(tempDir, "known_hosts")

	sampleContent := `# Test known_hosts
github.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl
192.168.1.100 ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBNf3rN6o5g2X2aM6hK/nE449yLzX2E8s4bM5gL1j0kG+t8h4vK6wM3nF2v4b8u3l9s1x7z5q6w8e0r1t2y3u4i5o= old key
[192.168.1.200]:2222 ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIOMqqnkVzrm0SdG6UOoqKLsabgH5C9okWi0dh2l9GKJl
|1|F1A9b... hashed-entry-example
`
	if err := os.WriteFile(khPath, []byte(sampleContent), 0o600); err != nil {
		t.Fatalf("failed to write sample known_hosts: %v", err)
	}

	repo := &Repository{}

	// Test 1: ListKnownHostRecords
	records, err := repo.ListKnownHostRecords(khPath)
	if err != nil {
		t.Fatalf("unexpected error listing records: %v", err)
	}
	if len(records) != 4 {
		t.Fatalf("expected 4 records, got %d", len(records))
	}
	if records[0].HostPattern != "github.com" || records[0].KeyType != "ssh-ed25519" {
		t.Errorf("record 0 mismatch: %+v", records[0])
	}
	if !strings.HasPrefix(records[0].Fingerprint, "SHA256:") {
		t.Errorf("expected SHA256 fingerprint, got %q", records[0].Fingerprint)
	}
	if !records[3].IsHashed {
		t.Errorf("expected record 3 to be hashed, got %+v", records[3])
	}

	// Test 2: RemoveKnownHost for 192.168.1.100
	backupPath, removed, err := repo.RemoveKnownHost(khPath, "192.168.1.100", 22)
	if err != nil {
		t.Fatalf("unexpected error removing known host: %v", err)
	}
	if removed == 0 {
		t.Errorf("expected at least 1 removed entry, got %d", removed)
	}
	if backupPath == "" || !strings.HasSuffix(backupPath, ".old") {
		t.Errorf("expected valid backup path, got %q", backupPath)
	}
	// Verify backup file exists
	if _, err := os.Stat(backupPath); err != nil {
		t.Errorf("expected backup file to exist: %v", err)
	}

	// Check updated records
	recordsAfterRemove, err := repo.ListKnownHostRecords(khPath)
	if err != nil {
		t.Fatalf("unexpected error listing records after remove: %v", err)
	}
	for _, r := range recordsAfterRemove {
		if strings.Contains(r.HostPattern, "192.168.1.100") {
			t.Errorf("found removed host in records: %+v", r)
		}
	}

	// Test 3: RemoveKnownHostByLine
	// Remove line of [192.168.1.200]:2222
	var targetLine int
	for _, r := range recordsAfterRemove {
		if strings.Contains(r.HostPattern, "192.168.1.200") {
			targetLine = r.LineNumber
			break
		}
	}
	if targetLine == 0 {
		t.Fatal("target line for 192.168.1.200 not found")
	}

	_, err = repo.RemoveKnownHostByLine(khPath, targetLine)
	if err != nil {
		t.Fatalf("unexpected error removing line: %v", err)
	}

	recordsFinal, err := repo.ListKnownHostRecords(khPath)
	if err != nil {
		t.Fatalf("unexpected error listing final records: %v", err)
	}
	for _, r := range recordsFinal {
		if strings.Contains(r.HostPattern, "192.168.1.200") {
			t.Errorf("found removed host by line in records: %+v", r)
		}
	}
}
