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
	"sync"
	"testing"

	"github.com/WhiteRoseLK/neossh/internal/core/domain"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type mockCompanionServiceForUI struct {
	mu            sync.Mutex
	pm            domain.PackageManagerInfo
	tools         []domain.CompanionTool
	installedIDs  []string
	installErr    error
	firstRun      bool
	markCompleted bool
}

func (m *mockCompanionServiceForUI) DetectPackageManager() domain.PackageManagerInfo {
	return m.pm
}

func (m *mockCompanionServiceForUI) CheckCompanionTools() []domain.CompanionTool {
	m.mu.Lock()
	defer m.mu.Unlock()
	res := make([]domain.CompanionTool, len(m.tools))
	copy(res, m.tools)
	return res
}

func (m *mockCompanionServiceForUI) InstallCompanionTools(toolIDs []string, onProgress func(tool domain.CompanionTool, status string, err error)) error {
	m.mu.Lock()
	m.installedIDs = append(m.installedIDs, toolIDs...)
	err := m.installErr
	m.mu.Unlock()

	if onProgress != nil {
		for _, id := range toolIDs {
			tool := domain.CompanionTool{ID: id, Name: id}
			onProgress(tool, "installing", nil)
			if err != nil {
				onProgress(tool, "failed", err)
				return err
			}
			onProgress(tool, "installed", nil)
		}
	}
	return err
}

func (m *mockCompanionServiceForUI) GetInstalledIDs() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	res := make([]string, len(m.installedIDs))
	copy(res, m.installedIDs)
	return res
}

func (m *mockCompanionServiceForUI) IsFirstRun() bool {
	return m.firstRun
}

func (m *mockCompanionServiceForUI) MarkFirstRunCompleted() error {
	m.markCompleted = true
	return nil
}

func TestNewOnboardingModal_InitialState(t *testing.T) {
	app := tview.NewApplication()
	mockSvc := &mockCompanionServiceForUI{
		pm: domain.PackageManagerInfo{
			Type:        domain.PackageManagerBrew,
			DisplayName: "Homebrew (brew)",
		},
		tools: []domain.CompanionTool{
			{ID: "chezmoi", Name: "chezmoi", Description: "Dotfiles sync", Shortcut: "D", Installed: true, Selected: false},
			{ID: "yazi", Name: "yazi", Description: "SFTP manager", Shortcut: "F", Installed: false, Selected: true},
			{ID: "ssh-copy-id", Name: "ssh-copy-id", Description: "Key deploy", Shortcut: "1-click", Installed: false, Selected: true},
		},
	}

	modal := NewOnboardingModal(app, mockSvc, nil, nil)
	if modal == nil {
		t.Fatal("expected modal not to be nil")
	}

	// Row 0 is header, rows 1..3 are tools
	if modal.table.GetRowCount() != 4 {
		t.Errorf("expected 4 rows in table, got %d", modal.table.GetRowCount())
	}

	// Check installed row cell
	cell0 := modal.table.GetCell(1, 0)
	if cell0.Text != " [✓] Installed" {
		t.Errorf("expected row 1 cell 0 to be '[✓] Installed', got %q", cell0.Text)
	}

	// Check selected row cell
	cell1 := modal.table.GetCell(2, 0)
	if cell1.Text != " [*] Install" {
		t.Errorf("expected row 2 cell 0 to be ' [*] Install', got %q", cell1.Text)
	}
}

func TestOnboardingModal_ToggleSelection(t *testing.T) {
	app := tview.NewApplication()
	mockSvc := &mockCompanionServiceForUI{
		pm: domain.PackageManagerInfo{Type: domain.PackageManagerBrew, DisplayName: "Homebrew (brew)"},
		tools: []domain.CompanionTool{
			{ID: "chezmoi", Name: "chezmoi", Installed: false, Selected: true},
		},
	}

	modal := NewOnboardingModal(app, mockSvc, nil, nil)
	modal.table.Select(1, 0)

	// Press Space to deselect
	event := tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone)
	modal.GetInputCapture()(event)

	if modal.tools[0].Selected {
		t.Errorf("expected tool to be deselected after Space")
	}
	cell := modal.table.GetCell(1, 0)
	if cell.Text != " [ ] Skip" {
		t.Errorf("expected cell text to be ' [ ] Skip', got %q", cell.Text)
	}

	// Press Space again to re-select
	modal.GetInputCapture()(event)
	if !modal.tools[0].Selected {
		t.Errorf("expected tool to be selected after second Space")
	}
	cell = modal.table.GetCell(1, 0)
	if cell.Text != " [*] Install" {
		t.Errorf("expected cell text to be ' [*] Install', got %q", cell.Text)
	}
}

func TestOnboardingModal_Skip(t *testing.T) {
	app := tview.NewApplication()
	mockSvc := &mockCompanionServiceForUI{
		pm: domain.PackageManagerInfo{Type: domain.PackageManagerBrew, DisplayName: "Homebrew (brew)"},
		tools: []domain.CompanionTool{
			{ID: "chezmoi", Name: "chezmoi", Installed: false, Selected: true},
		},
	}

	skipped := false
	modal := NewOnboardingModal(app, mockSvc, nil, func() {
		skipped = true
	})

	// Press Escape
	eventEsc := tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone)
	modal.GetInputCapture()(eventEsc)

	if !skipped {
		t.Errorf("expected onSkip to be called on Escape")
	}

	skipped = false
	// Press 'q'
	eventQ := tcell.NewEventKey(tcell.KeyRune, 'q', tcell.ModNone)
	modal.GetInputCapture()(eventQ)
	if !skipped {
		t.Errorf("expected onSkip to be called on 'q'")
	}
}

func TestOnboardingModal_Enter_AllInstalledProceeds(t *testing.T) {
	app := tview.NewApplication()
	mockSvc := &mockCompanionServiceForUI{
		pm: domain.PackageManagerInfo{Type: domain.PackageManagerBrew, DisplayName: "Homebrew (brew)"},
		tools: []domain.CompanionTool{
			{ID: "chezmoi", Name: "chezmoi", Installed: true, Selected: false},
		},
	}

	finished := false
	modal := NewOnboardingModal(app, mockSvc, func() {
		finished = true
	}, nil)

	// Press Enter when no tools need installing
	eventEnter := tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone)
	modal.GetInputCapture()(eventEnter)

	if !finished {
		t.Errorf("expected onFinish to be called immediately when no tools to install")
	}
}

func TestOnboardingModal_Enter_RunsInstallation(t *testing.T) {
	app := tview.NewApplication()
	mockSvc := &mockCompanionServiceForUI{
		pm: domain.PackageManagerInfo{Type: domain.PackageManagerBrew, DisplayName: "Homebrew (brew)"},
		tools: []domain.CompanionTool{
			{ID: "yazi", Name: "yazi", Installed: false, Selected: true},
		},
	}

	finished := false
	modal := NewOnboardingModal(app, mockSvc, func() {
		finished = true
	}, nil)
	modal.SetUpdater(func(f func()) { f() })

	// Press Enter to start installation
	eventEnter := tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone)
	modal.GetInputCapture()(eventEnter)

	// Wait for worker goroutine to complete
	modal.WaitInstallDone()

	installed := mockSvc.GetInstalledIDs()
	if len(installed) != 1 || installed[0] != "yazi" {
		t.Errorf("expected yazi to be installed, got %v", installed)
	}

	// Press Enter after install done to proceed
	modal.GetInputCapture()(eventEnter)
	if !finished {
		t.Errorf("expected onFinish to be called after installation")
	}
}
