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

package ui

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/WhiteRoseLK/neossh/internal/core/domain"
	"github.com/WhiteRoseLK/neossh/internal/core/ports"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type mockGitServiceForUI struct {
	isRepo        bool
	rootPath      string
	pushRemote    string
	pushURL       string
	keys          []domain.SSHKey
	currentConfig string
	configuredKey string
	scope         string
}

func (m *mockGitServiceForUI) IsGitRepository(path string) bool { return m.isRepo }
func (m *mockGitServiceForUI) GetGitRootPath(path string) (string, error) {
	return m.rootPath, nil
}
func (m *mockGitServiceForUI) GetPushRemoteURL(repoPath string) (string, string, error) {
	return m.pushRemote, m.pushURL, nil
}
func (m *mockGitServiceForUI) ListSSHKeys(sshDir string, serverRepo ports.ServerRepository) ([]domain.SSHKey, error) {
	return m.keys, nil
}
func (m *mockGitServiceForUI) ConfigureGitSSHKey(repoPath, keyPath, scope string) error {
	m.configuredKey = keyPath
	m.scope = scope
	return nil
}
func (m *mockGitServiceForUI) GetCurrentGitSSHConfig(repoPath string) (string, error) {
	return m.currentConfig, nil
}
func (m *mockGitServiceForUI) ClearGitSSHConfig(repoPath, scope string) error {
	m.currentConfig = ""
	return nil
}
func (m *mockGitServiceForUI) GetLoadedAgentKeys() ([]string, error) {
	return []string{}, nil
}
func (m *mockGitServiceForUI) ListAllSSHKeys(serverRepo ports.ServerRepository) ([]domain.SSHKey, error) {
	return m.keys, nil
}
func (m *mockGitServiceForUI) ListSSHKeysFromConfig(serverRepo ports.ServerRepository) ([]domain.SSHKey, error) {
	return m.keys, nil
}
func (m *mockGitServiceForUI) ListSSHKeysFromAgent() ([]domain.SSHKey, error) {
	return m.keys, nil
}
func (m *mockGitServiceForUI) LoadKeyToAgent(_ string) error                { return nil }
func (m *mockGitServiceForUI) UnloadKeyFromAgent(_ string) error            { return nil }
func (m *mockGitServiceForUI) UpdateKeyComment(_, _ string) error           { return nil }
func (m *mockGitServiceForUI) SetServerRepository(_ ports.ServerRepository) {}

type mockServerRepoForUI struct{}

func (m *mockServerRepoForUI) ListServers(_ string) ([]domain.Server, error) { return nil, nil }
func (m *mockServerRepoForUI) UpdateServer(_, _ domain.Server) error         { return nil }
func (m *mockServerRepoForUI) AddServer(_ domain.Server) error               { return nil }
func (m *mockServerRepoForUI) AddServers(_ []domain.Server) error            { return nil }
func (m *mockServerRepoForUI) DeleteServer(_ domain.Server) error            { return nil }
func (m *mockServerRepoForUI) SetPinned(_ string, _ bool) error              { return nil }
func (m *mockServerRepoForUI) SetHidden(_ string, _ bool) error              { return nil }
func (m *mockServerRepoForUI) RecordSSH(_ string) error                      { return nil }
func (m *mockServerRepoForUI) GetConfigFile() string                         { return "" }
func (m *mockServerRepoForUI) GetConfigFiles() ([]string, error)             { return []string{""}, nil }
func (m *mockServerRepoForUI) GetMetadataFile() string                       { return "" }
func (m *mockServerRepoForUI) DiscoverKnownHosts(_ string) ([]domain.Server, domain.ImportResult, error) {
	return nil, domain.ImportResult{}, nil
}
func (m *mockServerRepoForUI) ImportKnownHosts(_ string) (domain.ImportResult, error) {
	return domain.ImportResult{}, nil
}
func (m *mockServerRepoForUI) ListKnownHostRecords(_ string) ([]domain.KnownHostRecord, error) {
	return nil, nil
}
func (m *mockServerRepoForUI) RemoveKnownHost(_, _ string, _ int) (string, int, error) {
	return "", 0, nil
}
func (m *mockServerRepoForUI) RemoveKnownHostByLine(_ string, _ int) (string, error) {
	return "", nil
}
func (m *mockServerRepoForUI) ScanAndAddKnownHost(_, _ string, _ int) (*domain.KnownHostRecord, error) {
	return nil, nil
}
func (m *mockServerRepoForUI) GetTheme() (string, error)              { return "", nil }
func (m *mockServerRepoForUI) SaveTheme(_ string) error               { return nil }
func (m *mockServerRepoForUI) GetPreConnectCommand() (string, error)  { return "", nil }
func (m *mockServerRepoForUI) SavePreConnectCommand(_ string) error   { return nil }
func (m *mockServerRepoForUI) GetDefaultIdentityKey() (string, error) { return "", nil }
func (m *mockServerRepoForUI) SaveDefaultIdentityKey(_ string) error  { return nil }
func (m *mockServerRepoForUI) GetFileManager() (string, error)        { return "", nil }
func (m *mockServerRepoForUI) SaveFileManager(_ string) error         { return nil }

func TestEditKeyComment(t *testing.T) {
	app := tview.NewApplication()
	editor := NewEditKeyComment(app)
	editor.SetKey("id_rsa", "/path/to/id_rsa", "my-comment")

	var savedComment string
	editor.OnSave(func(c string) {
		savedComment = c
	})

	var canceled bool
	editor.OnCancel(func() {
		canceled = true
	})

	if err := editor.Show(); err != nil {
		t.Fatalf("Show failed: %v", err)
	}

	// Trigger cancel
	editor.cancel()
	if !canceled {
		t.Errorf("expected canceled callback to be called")
	}

	// Test save
	editor.save()
	if savedComment != "my-comment" {
		t.Errorf("expected saved comment to be 'my-comment', got %s", savedComment)
	}
}

func TestGitSSHSetup(t *testing.T) {
	app := tview.NewApplication()
	tempDir := t.TempDir()
	keyPath := filepath.Join(tempDir, "id_rsa")
	_ = os.WriteFile(keyPath, []byte("key"), 0o600)

	mockGit := &mockGitServiceForUI{
		isRepo:        true,
		rootPath:      tempDir,
		pushRemote:    "origin",
		pushURL:       "git@github.com:WhiteRoseLK/neossh.git",
		currentConfig: "ssh -i " + keyPath,
		keys: []domain.SSHKey{
			{
				Name:          "id_rsa",
				Path:          keyPath,
				Comment:       "test-key",
				LoadedInAgent: true,
				HasPublicKey:  true,
			},
		},
	}

	setup := NewGitSSHSetup(app, mockGit, nil)
	var doneCalled bool
	setup.OnDone(func() {
		doneCalled = true
	})

	var cancelCalled bool
	setup.OnCancel(func() {
		cancelCalled = true
	})

	if err := setup.Show(); err != nil {
		t.Fatalf("Show failed: %v", err)
	}

	setup.handleConfigure()
	if mockGit.configuredKey != keyPath {
		t.Errorf("expected configured key %s, got %s", keyPath, mockGit.configuredKey)
	}

	// Test Esc key capture
	event := setup.GetInputCapture()(tcell.NewEventKey(tcell.KeyEsc, 0, tcell.ModNone))
	if event != nil {
		t.Errorf("expected Esc key to be consumed")
	}
	if !cancelCalled {
		t.Errorf("expected cancel handler to be called on Esc")
	}

	_ = doneCalled
}

func TestServerDetails_SSHKeyDisplay(t *testing.T) {
	details := NewServerDetails()
	mockGit := &mockGitServiceForUI{
		keys: []domain.SSHKey{
			{
				Name:          "id_rsa",
				Path:          "/home/user/.ssh/id_rsa",
				Type:          "rsa",
				Size:          4096,
				Fingerprint:   "SHA256:abc123xyz",
				Comment:       "work-laptop",
				LoadedInAgent: true,
				IsEncrypted:   true,
			},
		},
	}
	details.SetGitService(mockGit, &mockServerRepoForUI{})

	server := domain.Server{
		Alias:         "prod-server",
		Host:          "192.168.1.1",
		User:          "root",
		Port:          22,
		IdentityFiles: []string{"/home/user/.ssh/id_rsa"},
	}

	details.UpdateServer(server)
	text := details.GetText(false)
	if text == "" {
		t.Fatalf("details text is empty")
	}

	// Verify key details are present
	expectedParts := []string{"SSH Key Details:", "work-laptop", "SHA256:abc123xyz", "4096 bits", "rsa"}
	for _, part := range expectedParts {
		if !containsSubstr(text, part) {
			t.Errorf("expected details to contain %q", part)
		}
	}
}

func containsSubstr(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
