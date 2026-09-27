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
	"testing"
	"time"

	"github.com/WhiteRoseLK/neossh/internal/core/domain"
	"github.com/WhiteRoseLK/neossh/internal/core/ports"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"go.uber.org/zap"
)

type mockServerService struct {
	ports.ServerService
	servers            []domain.Server
	sshCalled          bool
	sshAlias           string
	sshWithArgsCalled  bool
	sshWithArgsAlias   string
	sshWithArgsArgs    []string
	listServersCalled  bool
	sftpCalled         bool
	sftpAlias          string
	launchFMToolCalled bool
	launchFMAlias      string
	launchFMTool       string
}

func (m *mockServerService) SSH(alias string) error {
	m.sshCalled = true
	m.sshAlias = alias
	return nil
}

func (m *mockServerService) SSHWithArgs(alias string, extraArgs []string) error {
	m.sshWithArgsCalled = true
	m.sshWithArgsAlias = alias
	m.sshWithArgsArgs = extraArgs
	return nil
}

func (m *mockServerService) ListServers(query string) ([]domain.Server, error) {
	m.listServersCalled = true
	return m.servers, nil
}

func (m *mockServerService) IsForwarding(alias string) bool {
	return false
}

func (m *mockServerService) ReloadServers() error {
	return nil
}

func (m *mockServerService) UpdateServerPing(string, string, time.Duration) {}

func (m *mockServerService) DiscoverKnownHosts(string) ([]domain.Server, domain.ImportResult, error) {
	return nil, domain.ImportResult{}, nil
}

func (m *mockServerService) ImportKnownHosts(string) (domain.ImportResult, error) {
	return domain.ImportResult{}, nil
}

func (m *mockServerService) SetHidden(alias string, hidden bool) error {
	for i := range m.servers {
		if m.servers[i].Alias == alias {
			m.servers[i].Hidden = hidden
			break
		}
	}
	return nil
}

func (m *mockServerService) ListActiveSessions(string) ([]domain.Server, error) {
	return nil, nil
}

func (m *mockServerService) KillActiveSessions(domain.Server) (int, error) {
	return 0, nil
}

func (m *mockServerService) ResolveConfigServer(server domain.Server) (domain.Server, bool, error) {
	return server, true, nil
}

func (m *mockServerService) GetDefaultIdentityKey() (string, error) {
	return "", nil
}

func (m *mockServerService) SaveDefaultIdentityKey(string) error {
	return nil
}

func (m *mockServerService) SFTP(alias string) error {
	m.sftpCalled = true
	m.sftpAlias = alias
	return nil
}

func (m *mockServerService) LaunchFileManager(alias string, customTool string) error {
	m.launchFMToolCalled = true
	m.launchFMAlias = alias
	m.launchFMTool = customTool
	return nil
}

func TestNewTUI_ExitOnDisconnectConfig(t *testing.T) {
	logger := zap.NewNop().Sugar()
	svc := &mockServerService{}

	// Case 1: Default when no Config argument is provided
	appDef := NewTUI(logger, svc, "v1.0.0", "abc1234").(*tui)
	if appDef.ExitOnDisconnect() {
		t.Errorf("expected exitOnDisconnect to be false by default, got true")
	}

	// Case 2: Explicit false
	appFalse := NewTUI(logger, svc, "v1.0.0", "abc1234", Config{ExitOnDisconnect: false}).(*tui)
	if appFalse.ExitOnDisconnect() {
		t.Errorf("expected exitOnDisconnect to be false, got true")
	}

	// Case 3: Explicit true
	appTrue := NewTUI(logger, svc, "v1.0.0", "abc1234", Config{ExitOnDisconnect: true}).(*tui)
	if !appTrue.ExitOnDisconnect() {
		t.Errorf("expected exitOnDisconnect to be true, got false")
	}
}

func setupTestTUI(exitOnDisconnect bool) (*tui, *mockServerService, chan error) {
	logger := zap.NewNop().Sugar()
	srv := domain.Server{Alias: "srv1", Host: "1.2.3.4", Port: 22, User: "root"}
	svc := &mockServerService{servers: []domain.Server{srv}}

	appInstance := NewTUI(logger, svc, "v1.0.0", "abc", Config{ExitOnDisconnect: exitOnDisconnect}).(*tui)

	simScreen := tcell.NewSimulationScreen("UTF-8")
	_ = simScreen.Init()
	appInstance.app.SetScreen(simScreen)

	// Build the minimal components needed for handler execution
	appInstance.buildComponents()
	appInstance.serverList.UpdateServers(svc.servers)
	appInstance.root = tview.NewFlex()
	appInstance.app.SetRoot(appInstance.root, true)

	runDone := make(chan error, 1)
	go func() {
		runDone <- appInstance.app.Run()
	}()

	// Wait briefly for the application loop to be running
	time.Sleep(50 * time.Millisecond)

	// Reset listServersCalled flag since setup might have called it
	svc.listServersCalled = false

	return appInstance, svc, runDone
}

func TestHandleServerConnect_ExitOnDisconnectTrue(t *testing.T) {
	appInstance, svc, runDone := setupTestTUI(true)

	appInstance.handleServerConnect()

	if !svc.sshCalled {
		t.Fatalf("expected SSH to be called")
	}
	if svc.sshAlias != "srv1" {
		t.Fatalf("expected SSH alias 'srv1', got %q", svc.sshAlias)
	}

	// Application should have stopped and runDone should receive within 1 second
	select {
	case err := <-runDone:
		if err != nil {
			t.Fatalf("unexpected app.Run error: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Fatalf("timed out waiting for app.Run to stop when ExitOnDisconnect is true")
	}

	// RefreshServerList should NOT have been called when exiting on disconnect
	if svc.listServersCalled {
		t.Errorf("expected ListServers not to be called after exit on disconnect")
	}
}

func TestHandleServerConnect_ExitOnDisconnectFalse(t *testing.T) {
	appInstance, svc, runDone := setupTestTUI(false)
	defer func() {
		appInstance.app.Stop()
		<-runDone
	}()

	appInstance.handleServerConnect()

	if !svc.sshCalled {
		t.Fatalf("expected SSH to be called")
	}

	// Application should NOT have stopped
	select {
	case err := <-runDone:
		t.Fatalf("app.Run stopped unexpectedly when ExitOnDisconnect is false: %v", err)
	case <-time.After(100 * time.Millisecond):
		// Expected: app is still running
	}

	// RefreshServerList SHOULD have been called
	if !svc.listServersCalled {
		t.Errorf("expected ListServers to be called after disconnect when ExitOnDisconnect is false")
	}
}

func TestShowPortForwardForm_ForwardSSH_ExitOnDisconnectTrue(t *testing.T) {
	appInstance, svc, runDone := setupTestTUI(true)
	srv := svc.servers[0]

	form := appInstance.showPortForwardForm(srv)

	form.GetFormItem(1).(*tview.InputField).SetText("8080")
	form.GetFormItem(3).(*tview.InputField).SetText("80")
	form.GetFormItem(5).(*tview.DropDown).SetCurrentOption(1) // ForwardModeForwardSSH

	// Trigger "Start" button
	startBtn := form.GetButton(0)
	startBtn.InputHandler()(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), nil)

	if !svc.sshWithArgsCalled {
		t.Fatalf("expected SSHWithArgs to be called")
	}
	if svc.sshWithArgsAlias != "srv1" {
		t.Fatalf("expected alias 'srv1', got %q", svc.sshWithArgsAlias)
	}

	// Application should have stopped
	select {
	case err := <-runDone:
		if err != nil {
			t.Fatalf("unexpected app.Run error: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Fatalf("timed out waiting for app.Run to stop when ExitOnDisconnect is true")
	}
}

func TestShowPortForwardForm_ForwardSSH_ExitOnDisconnectFalse(t *testing.T) {
	appInstance, svc, runDone := setupTestTUI(false)
	defer func() {
		appInstance.app.Stop()
		<-runDone
	}()

	srv := svc.servers[0]
	form := appInstance.showPortForwardForm(srv)

	form.GetFormItem(1).(*tview.InputField).SetText("8080")
	form.GetFormItem(3).(*tview.InputField).SetText("80")
	form.GetFormItem(5).(*tview.DropDown).SetCurrentOption(1) // ForwardModeForwardSSH

	// Trigger "Start" button
	startBtn := form.GetButton(0)
	startBtn.InputHandler()(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), nil)

	if !svc.sshWithArgsCalled {
		t.Fatalf("expected SSHWithArgs to be called")
	}

	// Application should NOT have stopped
	select {
	case err := <-runDone:
		t.Fatalf("app.Run stopped unexpectedly when ExitOnDisconnect is false: %v", err)
	case <-time.After(100 * time.Millisecond):
		// Expected: app is still running
	}
}

func TestNewTUI_InitialFilterConfig(t *testing.T) {
	logger := zap.NewNop().Sugar()
	srv := domain.Server{Alias: "srv-prod", Host: "1.2.3.4", Port: 22, User: "root"}
	svc := &mockServerService{servers: []domain.Server{srv}}

	app := NewTUI(logger, svc, "v1.0.0", "abc", Config{
		InitialFilter: "srv-prod",
	}).(*tui)

	if app.InitialFilter() != "srv-prod" {
		t.Errorf("expected InitialFilter to be 'srv-prod', got %q", app.InitialFilter())
	}

	app.buildComponents()
	app.loadPreferences()
	app.buildLayout()
	app.bindEvents()
	app.loadInitialData()

	// Verify searchBar text is set
	if app.searchBar.GetText() != "srv-prod" {
		t.Errorf("expected searchBar text to be 'srv-prod', got %q", app.searchBar.GetText())
	}

	// Verify server list has been populated
	if app.serverList.GetItemCount() == 0 {
		t.Errorf("expected server list to have at least 1 server item")
	}
}

func TestIsGUIFileManager(t *testing.T) {
	guiTools := []string{"filezilla", "FileZilla", "cyberduck", "nautilus", "dolphin"}
	for _, tool := range guiTools {
		if !isGUIFileManager(tool) {
			t.Errorf("expected %s to be recognized as GUI file manager", tool)
		}
	}

	cliTools := []string{"sftp", "yazi", "ranger", "internal", "custom-sh"}
	for _, tool := range cliTools {
		if isGUIFileManager(tool) {
			t.Errorf("expected %s to NOT be recognized as GUI file manager", tool)
		}
	}
}

func TestHandleSFTPFileManager_ExternalLaunch(t *testing.T) {
	appInstance, svc, _ := setupTestTUI(false)
	defer appInstance.app.Stop()

	// Select server
	appInstance.serverList.SetCurrentItem(0)

	// Test external launch (default tool is "sftp")
	appInstance.handleSFTPFileManager()

	if !svc.launchFMToolCalled {
		t.Fatalf("expected LaunchFileManager to be called")
	}
	if svc.launchFMAlias != "srv1" {
		t.Errorf("expected alias 'srv1', got %q", svc.launchFMAlias)
	}
	if svc.launchFMTool != "sftp" {
		t.Errorf("expected tool 'sftp', got %q", svc.launchFMTool)
	}
}
