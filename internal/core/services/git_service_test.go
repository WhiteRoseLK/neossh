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

package services

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/WhiteRoseLK/neossh/internal/core/domain"
	"go.uber.org/zap"
)

type mockServerRepo struct {
	servers []domain.Server
}

func (m *mockServerRepo) ListServers(query string) ([]domain.Server, error) {
	return m.servers, nil
}
func (m *mockServerRepo) UpdateServer(server, newServer domain.Server) error { return nil }
func (m *mockServerRepo) AddServer(server domain.Server) error               { return nil }
func (m *mockServerRepo) AddServers(servers []domain.Server) error           { return nil }
func (m *mockServerRepo) DeleteServer(server domain.Server) error            { return nil }
func (m *mockServerRepo) SetPinned(alias string, pinned bool) error          { return nil }
func (m *mockServerRepo) SetHidden(alias string, hidden bool) error          { return nil }
func (m *mockServerRepo) RecordSSH(alias string) error                       { return nil }
func (m *mockServerRepo) GetConfigFile() string                              { return "" }
func (m *mockServerRepo) GetConfigFiles() ([]string, error)                  { return []string{""}, nil }
func (m *mockServerRepo) GetMetadataFile() string                            { return "" }
func (m *mockServerRepo) DiscoverKnownHosts(path string) ([]domain.Server, domain.ImportResult, error) {
	return nil, domain.ImportResult{}, nil
}
func (m *mockServerRepo) ImportKnownHosts(path string) (domain.ImportResult, error) {
	return domain.ImportResult{}, nil
}
func (m *mockServerRepo) ListKnownHostRecords(string) ([]domain.KnownHostRecord, error) {
	return nil, nil
}
func (m *mockServerRepo) RemoveKnownHost(string, string, int) (string, int, error) {
	return "", 0, nil
}
func (m *mockServerRepo) RemoveKnownHostByLine(string, int) (string, error) {
	return "", nil
}
func (m *mockServerRepo) ScanAndAddKnownHost(string, string, int) (*domain.KnownHostRecord, error) {
	return nil, nil
}
func (m *mockServerRepo) GetTheme() (string, error)              { return "", nil }
func (m *mockServerRepo) SaveTheme(theme string) error           { return nil }
func (m *mockServerRepo) GetPreConnectCommand() (string, error)  { return "", nil }
func (m *mockServerRepo) SavePreConnectCommand(cmd string) error { return nil }
func (m *mockServerRepo) GetDefaultIdentityKey() (string, error) { return "", nil }
func (m *mockServerRepo) SaveDefaultIdentityKey(_ string) error  { return nil }
func (m *mockServerRepo) GetFileManager() (string, error)        { return "", nil }
func (m *mockServerRepo) SaveFileManager(_ string) error         { return nil }

func TestGitService_IsGitRepository(t *testing.T) {
	logger := zap.NewNop().Sugar()
	gs := NewGitService(logger)

	tempDir := t.TempDir()
	if gs.IsGitRepository(tempDir) {
		t.Fatalf("expected non-git repo for empty directory %s", tempDir)
	}

	cmd := exec.Command("git", "init", tempDir)
	if err := cmd.Run(); err != nil {
		t.Skip("git not installed or failed to init")
	}

	if !gs.IsGitRepository(tempDir) {
		t.Fatalf("expected git repo for initialized directory %s", tempDir)
	}

	rootPath, err := gs.GetGitRootPath(tempDir)
	if err != nil {
		t.Fatalf("failed to get git root path: %v", err)
	}
	// On macOS /private/var vs /var symlink can happen
	evalTemp, _ := filepath.EvalSymlinks(tempDir)
	evalRoot, _ := filepath.EvalSymlinks(rootPath)
	if evalTemp != evalRoot {
		t.Fatalf("root path mismatch: expected %s, got %s", evalTemp, evalRoot)
	}
}

func TestGitService_ConfigureAndClearGitSSH(t *testing.T) {
	logger := zap.NewNop().Sugar()
	gs := NewGitService(logger)

	tempDir := t.TempDir()
	cmd := exec.Command("git", "init", tempDir)
	if err := cmd.Run(); err != nil {
		t.Skip("git not available")
	}

	dummyKeyPath := filepath.Join(tempDir, "id_ed25519")
	if err := os.WriteFile(dummyKeyPath, []byte("-----BEGIN OPENSSH PRIVATE KEY-----\ntest\n-----END OPENSSH PRIVATE KEY-----\n"), 0o600); err != nil {
		t.Fatalf("failed to create dummy key: %v", err)
	}

	// Test invalid scope
	if err := gs.ConfigureGitSSHKey(tempDir, dummyKeyPath, "invalid"); err == nil {
		t.Fatalf("expected error for invalid scope")
	}

	// Test local configuration
	if err := gs.ConfigureGitSSHKey(tempDir, dummyKeyPath, ScopeLocal); err != nil {
		t.Fatalf("ConfigureGitSSHKey failed: %v", err)
	}

	cfg, err := gs.GetCurrentGitSSHConfig(tempDir)
	if err != nil {
		t.Fatalf("GetCurrentGitSSHConfig failed: %v", err)
	}
	if cfg != "ssh -i "+dummyKeyPath+" -o IdentitiesOnly=yes" {
		t.Fatalf("unexpected ssh command: %s", cfg)
	}

	// Test clearing configuration
	if err := gs.ClearGitSSHConfig(tempDir, ScopeLocal); err != nil {
		t.Fatalf("ClearGitSSHConfig failed: %v", err)
	}

	cfgAfter, _ := gs.GetCurrentGitSSHConfig(tempDir)
	if cfgAfter != "" {
		t.Fatalf("expected empty config after clear, got %s", cfgAfter)
	}
}

func TestGitService_ListSSHKeys_And_Parsing(t *testing.T) {
	logger := zap.NewNop().Sugar()
	gs := NewGitService(logger)

	tempDir := t.TempDir()
	rsaKey := filepath.Join(tempDir, "id_rsa")
	rsaPub := filepath.Join(tempDir, "id_rsa.pub")

	_ = os.WriteFile(rsaKey, []byte("-----BEGIN RSA PRIVATE KEY-----\nProc-Type: 4,ENCRYPTED\ntest\n-----END RSA PRIVATE KEY-----\n"), 0o600)
	_ = os.WriteFile(rsaPub, []byte("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC dummy rsa-key-comment\n"), 0o644)

	edKey := filepath.Join(tempDir, "id_ed25519")
	edPub := filepath.Join(tempDir, "id_ed25519.pub")
	_ = os.WriteFile(edKey, []byte("-----BEGIN OPENSSH PRIVATE KEY-----\ntest\n-----END OPENSSH PRIVATE KEY-----\n"), 0o600)
	_ = os.WriteFile(edPub, []byte("ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAA dummy ed25519-key-comment\n"), 0o644)

	repo := &mockServerRepo{
		servers: []domain.Server{
			{
				Alias:         "server1",
				IdentityFiles: []string{rsaKey},
			},
		},
	}
	gs.SetServerRepository(repo)

	keys, err := gs.ListSSHKeys(tempDir, repo)
	if err != nil {
		t.Fatalf("ListSSHKeys failed: %v", err)
	}

	if len(keys) < 2 {
		t.Fatalf("expected at least 2 keys, got %d", len(keys))
	}

	var foundRSA, foundED bool
	for _, k := range keys {
		if k.Path == rsaKey {
			foundRSA = true
			if !k.IsEncrypted {
				t.Errorf("expected rsa key to be marked encrypted")
			}
			if !k.HasPublicKey {
				t.Errorf("expected rsa key to have pub key")
			}
		}
		if k.Path == edKey {
			foundED = true
			if !k.HasPublicKey {
				t.Errorf("expected ed key to have pub key")
			}
		}
	}

	if !foundRSA || !foundED {
		t.Fatalf("keys not properly listed: foundRSA=%v, foundED=%v", foundRSA, foundED)
	}
}
