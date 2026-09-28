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
	"strings"
	"testing"
	"time"

	"github.com/WhiteRoseLK/neossh/internal/core/domain"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"go.uber.org/zap"
)

type mockReadOnlyService struct {
	servers       []domain.Server
	addCalled     bool
	updateCalled  bool
	deleteCalled  bool
	copyKeyCalled bool
	importCalled  bool
}

func (m *mockReadOnlyService) ListServers(query string) ([]domain.Server, error) {
	return m.servers, nil
}

func (m *mockReadOnlyService) UpdateServer(server domain.Server, newServer domain.Server) error {
	m.updateCalled = true
	return nil
}

func (m *mockReadOnlyService) AddServer(server domain.Server) error {
	m.addCalled = true
	return nil
}

func (m *mockReadOnlyService) DeleteServer(server domain.Server) error {
	m.deleteCalled = true
	return nil
}

func (m *mockReadOnlyService) SetPinned(alias string, pinned bool) error {
	return nil
}

func (m *mockReadOnlyService) SetHidden(alias string, hidden bool) error {
	return nil
}

func (m *mockReadOnlyService) SSH(alias string) error {
	return nil
}

func (m *mockReadOnlyService) SSHWithArgs(alias string, extraArgs []string) error {
	return nil
}

func (m *mockReadOnlyService) CopySSHKey(alias string) error {
	m.copyKeyCalled = true
	return nil
}

func (m *mockReadOnlyService) StartForward(alias string, extraArgs []string) (int, error) {
	return 0, nil
}

func (m *mockReadOnlyService) StopForwarding(alias string) error {
	return nil
}

func (m *mockReadOnlyService) IsForwarding(alias string) bool {
	return false
}

func (m *mockReadOnlyService) Ping(server domain.Server) (bool, time.Duration, error) {
	return true, 10 * time.Millisecond, nil
}

func (m *mockReadOnlyService) DiscoverKnownHosts(string) ([]domain.Server, domain.ImportResult, error) {
	return nil, domain.ImportResult{}, nil
}

func (m *mockReadOnlyService) ImportKnownHosts(string) (domain.ImportResult, error) {
	m.importCalled = true
	return domain.ImportResult{}, nil
}

func (m *mockReadOnlyService) ListKnownHostRecords(string) ([]domain.KnownHostRecord, error) {
	return nil, nil
}

func (m *mockReadOnlyService) RemoveKnownHost(string, string, int) (string, int, error) {
	return "known_hosts.old", 1, nil
}

func (m *mockReadOnlyService) RemoveKnownHostByLine(string, int) (string, error) {
	return "known_hosts.old", nil
}

func (m *mockReadOnlyService) ScanAndAddKnownHost(string, string, int) (*domain.KnownHostRecord, error) {
	return &domain.KnownHostRecord{HostPattern: "1.1.1.1", KeyType: "ssh-ed25519", Fingerprint: "SHA256:abc"}, nil
}

func (m *mockReadOnlyService) GetTheme() (string, error) {
	return "dark", nil
}

func (m *mockReadOnlyService) SaveTheme(string) error {
	return nil
}

func (m *mockReadOnlyService) ListActiveSessions(string) ([]domain.Server, error) {
	return nil, nil
}

func (m *mockReadOnlyService) KillActiveSessions(domain.Server) (int, error) {
	return 0, nil
}

func (m *mockReadOnlyService) ResolveConfigServer(server domain.Server) (domain.Server, bool, error) {
	return server, true, nil
}

func (m *mockReadOnlyService) GetDefaultIdentityKey() (string, error) {
	return "", nil
}

func (m *mockReadOnlyService) SaveDefaultIdentityKey(string) error {
	return nil
}

func (m *mockReadOnlyService) ReloadServers() error {
	return nil
}

func (m *mockReadOnlyService) SFTP(alias string) error {
	return nil
}

func (m *mockReadOnlyService) LaunchFileManager(alias string, customTool string) error {
	return nil
}

func (m *mockReadOnlyService) UpdateServerPing(string, string, time.Duration) {}

func (m *mockReadOnlyService) GetSSHAgentStatus() domain.SSHAgentStatus {
	return domain.SSHAgentStatus{Available: true, KeyCount: 2, Type: domain.AgentTypeOpenSSH}
}

func TestTUI_ReadOnlyState(t *testing.T) {
	logger := zap.NewNop().Sugar()
	mockSvc := &mockReadOnlyService{}

	app := NewTUI(logger, mockSvc, "v1.0.0", "abcdef1", Config{
		ReadOnly: true,
	}).(*tui)

	if !app.IsReadOnly() {
		t.Fatal("expected app to be readonly")
	}

	app.SetReadOnly(false)
	if app.IsReadOnly() {
		t.Fatal("expected app to not be readonly after SetReadOnly(false)")
	}

	app.SetReadOnly(true)
	if !app.IsReadOnly() {
		t.Fatal("expected app to be readonly after SetReadOnly(true)")
	}
}

func TestStatusText_ReadOnlyVsDefault(t *testing.T) {
	defaultText := DefaultStatusText()
	readonlyText := ReadonlyStatusText()

	// In default mode, modifying actions should be present
	for _, action := range []string{"Add", "Edit", "Delete", "Clone", "Paste SSH", "Install Key"} {
		if !strings.Contains(defaultText, action) {
			t.Errorf("default status text missing action %q", action)
		}
	}
	if strings.Contains(defaultText, "[READONLY]") {
		t.Error("default status text should not contain [READONLY]")
	}

	// In readonly mode, modifying actions should NOT be present
	for _, action := range []string{"Add", "Edit", "Delete", "Clone", "Paste SSH", "Install Key"} {
		if strings.Contains(readonlyText, action) {
			t.Errorf("readonly status text unexpectedly contains action %q", action)
		}
	}
	if !strings.Contains(readonlyText, "[READONLY]") {
		t.Error("readonly status text should contain [READONLY]")
	}

	// StatusText helper function check
	if StatusText(false) != defaultText {
		t.Error("StatusText(false) should equal DefaultStatusText()")
	}
	if StatusText(true) != readonlyText {
		t.Error("StatusText(true) should equal ReadonlyStatusText()")
	}
}

func TestAppHeader_ReadOnlyVisualIndicator(t *testing.T) {
	normalHeader := NewAppHeader("v1.0.0", "1234567", "https://github.com/example", false)
	roHeader := NewAppHeader("v1.0.0", "1234567", "https://github.com/example", true)

	// Left section (stylized app name)
	leftNormal := normalHeader.buildLeftSection(tcell.ColorBlack)
	if strings.Contains(leftNormal.GetText(false), "READONLY") {
		t.Error("normal header left section should not contain READONLY")
	}

	leftRO := roHeader.buildLeftSection(tcell.ColorBlack)
	if !strings.Contains(leftRO.GetText(false), "READONLY") {
		t.Error("readonly header left section should contain READONLY")
	}

	// Center section (chips)
	centerNormal := normalHeader.buildCenterSection(tcell.ColorBlack)
	if strings.Contains(centerNormal.GetText(false), "READONLY") {
		t.Error("normal header center section should not contain READONLY chip")
	}

	centerRO := roHeader.buildCenterSection(tcell.ColorBlack)
	if !strings.Contains(centerRO.GetText(false), "READONLY") {
		t.Error("readonly header center section should contain READONLY chip")
	}
}

func TestServerDetails_ReadOnlyCommandsDisplay(t *testing.T) {
	srv := domain.Server{
		Alias: "web-01",
		Host:  "192.168.1.10",
		User:  "ubuntu",
		Port:  22,
	}

	normalDetails := NewServerDetails(false)
	normalDetails.UpdateServer(srv)
	normalText := normalDetails.GetText(false)

	if !strings.Contains(normalText, "Add new server") || !strings.Contains(normalText, "Edit entry") {
		t.Error("normal details should list Add and Edit commands")
	}
	if strings.Contains(normalText, "Modifications disabled") {
		t.Error("normal details should not mention modifications disabled")
	}

	roDetails := NewServerDetails(true)
	roDetails.UpdateServer(srv)
	roText := roDetails.GetText(false)

	if !strings.Contains(roText, "Modifications disabled (readonly mode)") {
		t.Error("readonly details should mention modifications disabled")
	}
	if strings.Contains(roText, "Add new server") || strings.Contains(roText, "Edit entry") {
		t.Error("readonly details should not list active Add or Edit commands")
	}
}

func TestHandlers_ReadOnlyBlocking(t *testing.T) {
	logger := zap.NewNop().Sugar()
	mockSvc := &mockReadOnlyService{
		servers: []domain.Server{
			{Alias: "prod-01", Host: "10.0.0.1", User: "root", Port: 22},
		},
	}

	appInstance := tview.NewApplication()
	uiApp := NewTUI(logger, mockSvc, "v1.0.0", "abcdef1", Config{
		ReadOnly: true,
	}).(*tui)
	uiApp.app = appInstance
	uiApp.buildComponents()
	uiApp.buildLayout()

	// Verify updateListTitle includes [READONLY]
	uiApp.updateListTitle()
	if !strings.Contains(uiApp.serverList.GetTitle(), "[READONLY]") {
		t.Fatalf("expected serverList title to contain [READONLY], got %q", uiApp.serverList.GetTitle())
	}

	// Test handleGlobalKeys for modifying action keys in readonly mode
	blockedKeys := []rune{'a', 'e', 'd', 'c', 'p', 'K', 'y', 'C', 'v', 't', 'i', 'I'}
	for _, k := range blockedKeys {
		t.Run(string(k), func(t *testing.T) {
			ev := tcell.NewEventKey(tcell.KeyRune, k, tcell.ModNone)
			res := uiApp.handleGlobalKeys(ev)
			if res != nil {
				t.Fatalf("expected key %c to be consumed by readonly interceptor", k)
			}
			btn, ok := uiApp.app.GetFocus().(*tview.Button)
			if !ok || btn == nil || btn.GetLabel() != "Close" {
				t.Fatalf("key %c: expected modal 'Close' button to be focused, got %T (%v)", k, uiApp.app.GetFocus(), uiApp.app.GetFocus())
			}
			statusText := uiApp.statusBar.GetText(false)
			if !strings.Contains(statusText, ReadonlyMessage) {
				t.Fatalf("key %c: expected statusBar to contain ReadonlyMessage, got %q", k, statusText)
			}
		})
	}

	// Verify that mockService modification methods were NEVER called
	if mockSvc.addCalled {
		t.Error("AddServer was unexpectedly called")
	}
	if mockSvc.updateCalled {
		t.Error("UpdateServer was unexpectedly called")
	}
	if mockSvc.deleteCalled {
		t.Error("DeleteServer was unexpectedly called")
	}
	if mockSvc.copyKeyCalled {
		t.Error("CopySSHKey was unexpectedly called")
	}

	// Test calling handlers directly in readonly mode
	t.Run("DirectHandlerAdd", func(t *testing.T) {
		uiApp.handleServerAdd()
		if mockSvc.addCalled {
			t.Error("AddServer was called directly")
		}
	})

	t.Run("DirectHandlerEdit", func(t *testing.T) {
		uiApp.handleServerEdit()
		if mockSvc.updateCalled {
			t.Error("UpdateServer was called directly")
		}
	})

	t.Run("DirectHandlerDelete", func(t *testing.T) {
		uiApp.handleServerDelete()
		if mockSvc.deleteCalled {
			t.Error("DeleteServer was called directly")
		}
	})

	t.Run("DirectHandlerClone", func(t *testing.T) {
		uiApp.handleServerClone()
		if mockSvc.addCalled {
			t.Error("AddServer was called via clone")
		}
	})

	t.Run("DirectHandlerPaste", func(t *testing.T) {
		uiApp.handlePasteCommand()
		if mockSvc.addCalled {
			t.Error("AddServer was called via paste")
		}
	})

	t.Run("DirectHandlerInstallKey", func(t *testing.T) {
		uiApp.handleInstallSSHKey()
		if mockSvc.copyKeyCalled {
			t.Error("CopySSHKey was called directly")
		}
	})

	t.Run("DirectHandlerSave", func(t *testing.T) {
		srv := domain.Server{Alias: "srv"}
		uiApp.handleServerSave(srv, nil)
		if mockSvc.addCalled || mockSvc.updateCalled {
			t.Error("Server save modified service in readonly mode")
		}
	})

	t.Run("DirectHandlerImportKnownHosts", func(t *testing.T) {
		uiApp.handleImportKnownHosts()
		if mockSvc.importCalled {
			t.Error("ImportKnownHosts was called in readonly mode")
		}
	})
}
