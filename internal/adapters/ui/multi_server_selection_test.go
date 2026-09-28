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

func TestServerList_MultiSelectionBasics(t *testing.T) {
	sl := NewServerList()
	servers := []domain.Server{
		{Alias: "srv-alpha", Host: "10.0.0.1", User: "root", Group: "Production"},
		{Alias: "srv-beta", Host: "10.0.0.2", User: "admin", Group: "Production"},
		{Alias: "srv-gamma", Host: "10.0.0.3", User: "ubuntu", Group: "Staging"},
	}
	sl.UpdateServers(servers)

	if sl.GetMultiSelectionCount() != 0 {
		t.Fatalf("expected initial count 0, got %d", sl.GetMultiSelectionCount())
	}
	if sl.IsServerSelected("srv-alpha") {
		t.Errorf("expected srv-alpha not to be selected")
	}

	// 1. Toggle selection
	var lastCount int
	sl.OnMultiSelectionChange(func(count int) {
		lastCount = count
	})

	sl.ToggleServerSelection("srv-alpha")
	if !sl.IsServerSelected("srv-alpha") {
		t.Errorf("expected srv-alpha to be selected")
	}
	if sl.GetMultiSelectionCount() != 1 || lastCount != 1 {
		t.Errorf("expected count 1, got %d (lastCount: %d)", sl.GetMultiSelectionCount(), lastCount)
	}

	// 2. SetServerSelected
	sl.SetServerSelected("srv-beta", true)
	if sl.GetMultiSelectionCount() != 2 || lastCount != 2 {
		t.Errorf("expected count 2, got %d", sl.GetMultiSelectionCount())
	}

	selected := sl.GetMultiSelectedServers()
	if len(selected) != 2 {
		t.Fatalf("expected 2 selected servers, got %d", len(selected))
	}
	if selected[0].Alias != "srv-alpha" || selected[1].Alias != "srv-beta" {
		t.Errorf("unexpected selected aliases: %v", selected)
	}

	// 3. ClearSelection
	sl.ClearSelection()
	if sl.GetMultiSelectionCount() != 0 || lastCount != 0 {
		t.Errorf("expected count 0 after clear, got %d", sl.GetMultiSelectionCount())
	}

	// 4. SelectAllDisplayed & DeselectAllDisplayed
	sl.SelectAllDisplayed()
	if sl.GetMultiSelectionCount() != 3 {
		t.Errorf("expected count 3 after SelectAllDisplayed, got %d", sl.GetMultiSelectionCount())
	}

	sl.ToggleSelectAll() // should deselect all when all are selected
	if sl.GetMultiSelectionCount() != 0 {
		t.Errorf("expected count 0 after ToggleSelectAll deselect, got %d", sl.GetMultiSelectionCount())
	}

	sl.ToggleSelectAll() // should select all when none are selected
	if sl.GetMultiSelectionCount() != 3 {
		t.Errorf("expected count 3 after ToggleSelectAll select, got %d", sl.GetMultiSelectionCount())
	}

	sl.DeselectAllDisplayed()
	if sl.GetMultiSelectionCount() != 0 {
		t.Errorf("expected count 0 after DeselectAllDisplayed, got %d", sl.GetMultiSelectionCount())
	}
}

func TestServerList_ToggleSelectGroup(t *testing.T) {
	sl := NewServerList()
	now := time.Now()
	servers := []domain.Server{
		{Alias: "pinned-1", Host: "10.0.0.1", PinnedAt: now},
		{Alias: "prod-1", Host: "10.0.0.2", Group: "Production"},
		{Alias: "prod-2", Host: "10.0.0.3", Group: "Production"},
		{Alias: "ungrouped-1", Host: "10.0.0.4"},
	}
	sl.UpdateServers(servers)

	// Toggle Production group
	sl.ToggleSelectGroup("Production")
	if sl.GetMultiSelectionCount() != 2 {
		t.Errorf("expected 2 servers selected in Production, got %d", sl.GetMultiSelectionCount())
	}
	if !sl.IsServerSelected("prod-1") || !sl.IsServerSelected("prod-2") {
		t.Errorf("expected prod-1 and prod-2 to be selected")
	}

	// Toggle Production group off
	sl.ToggleSelectGroup("Production")
	if sl.GetMultiSelectionCount() != 0 {
		t.Errorf("expected 0 servers selected after toggling off Production, got %d", sl.GetMultiSelectionCount())
	}

	// Toggle Pinned group
	sl.ToggleSelectGroup("Pinned")
	if sl.GetMultiSelectionCount() != 1 || !sl.IsServerSelected("pinned-1") {
		t.Errorf("expected pinned-1 to be selected")
	}
	sl.ClearSelection()

	// Toggle Ungrouped
	sl.ToggleSelectGroup("Ungrouped")
	if sl.GetMultiSelectionCount() != 1 || !sl.IsServerSelected("ungrouped-1") {
		t.Errorf("expected ungrouped-1 to be selected")
	}
}

func TestServerList_StaleSelectionsCleaned(t *testing.T) {
	sl := NewServerList()
	servers := []domain.Server{
		{Alias: "srv-1", Host: "10.0.0.1"},
		{Alias: "srv-2", Host: "10.0.0.2"},
	}
	sl.UpdateServers(servers)
	sl.SetServerSelected("srv-1", true)
	sl.SetServerSelected("srv-2", true)
	if sl.GetMultiSelectionCount() != 2 {
		t.Fatalf("expected count 2, got %d", sl.GetMultiSelectionCount())
	}

	// Remove srv-2
	newServers := []domain.Server{
		{Alias: "srv-1", Host: "10.0.0.1"},
	}
	sl.UpdateServers(newServers)
	if sl.GetMultiSelectionCount() != 1 {
		t.Errorf("expected count 1 after srv-2 removed, got %d", sl.GetMultiSelectionCount())
	}
	if sl.IsServerSelected("srv-2") {
		t.Errorf("expected srv-2 to no longer be selected")
	}
}

func TestServerList_GetTargetServers(t *testing.T) {
	sl := NewServerList()
	servers := []domain.Server{
		{Alias: "srv-1", Host: "10.0.0.1"},
		{Alias: "srv-2", Host: "10.0.0.2"},
	}
	sl.UpdateServers(servers)

	// Case 1: None selected, fallback to focused item
	targets := sl.GetTargetServers()
	if len(targets) != 1 || targets[0].Alias != "srv-1" {
		t.Errorf("expected fallback to focused server srv-1, got %v", targets)
	}

	// Case 2: Multi-selected
	sl.SetServerSelected("srv-2", true)
	targets = sl.GetTargetServers()
	if len(targets) != 1 || targets[0].Alias != "srv-2" {
		t.Errorf("expected multi-selected target srv-2, got %v", targets)
	}
}

func TestServerList_KeyEvents(t *testing.T) {
	sl := NewServerList()
	servers := []domain.Server{
		{Alias: "srv-1", Host: "10.0.0.1", Group: "Datacenter"},
		{Alias: "srv-2", Host: "10.0.0.2", Group: "Datacenter"},
	}
	sl.UpdateServers(servers)

	// 1. Esc when selections exist clears selections and does not trigger onReturnToSearch
	returnToSearchCalled := false
	sl.OnReturnToSearch(func() {
		returnToSearchCalled = true
	})

	sl.SetServerSelected("srv-1", true)
	evEsc := tcell.NewEventKey(tcell.KeyESC, 0, tcell.ModNone)
	res := sl.handleKeyInput(evEsc)
	if res != nil {
		t.Errorf("expected Esc with active selections to return nil (consumed)")
	}
	if sl.GetMultiSelectionCount() != 0 {
		t.Errorf("expected Esc to clear selections, got %d", sl.GetMultiSelectionCount())
	}
	if returnToSearchCalled {
		t.Errorf("expected onReturnToSearch not to be called on first Esc")
	}

	// Esc when no selections exist triggers onReturnToSearch
	res = sl.handleKeyInput(evEsc)
	if res != nil {
		t.Errorf("expected Esc to return nil")
	}
	if !returnToSearchCalled {
		t.Errorf("expected onReturnToSearch to be called on second Esc")
	}

	// 2. Ctrl+A toggles select all
	evCtrlA := tcell.NewEventKey(tcell.KeyCtrlA, 0, tcell.ModNone)
	sl.handleKeyInput(evCtrlA)
	if sl.GetMultiSelectionCount() != 2 {
		t.Errorf("expected count 2 after Ctrl+A, got %d", sl.GetMultiSelectionCount())
	}
	sl.handleKeyInput(evCtrlA)
	if sl.GetMultiSelectionCount() != 0 {
		t.Errorf("expected count 0 after second Ctrl+A, got %d", sl.GetMultiSelectionCount())
	}

	// 3. Space on server item toggles selection
	// Find index of srv-1
	srv1Idx := -1
	for i, item := range sl.displayedItems {
		if item != nil && item.Alias == "srv-1" {
			srv1Idx = i
			break
		}
	}
	if srv1Idx >= 0 {
		sl.List.SetCurrentItem(srv1Idx)
		evSpace := tcell.NewEventKey(tcell.KeyRune, ' ', tcell.ModNone)
		sl.handleKeyInput(evSpace)
		if !sl.IsServerSelected("srv-1") {
			t.Errorf("expected Space to select srv-1")
		}
		sl.handleKeyInput(evSpace)
		if sl.IsServerSelected("srv-1") {
			t.Errorf("expected Space to deselect srv-1")
		}
	}

	// 4. Asterisk '*' on group header or server item
	evStar := tcell.NewEventKey(tcell.KeyRune, '*', tcell.ModNone)
	sl.handleKeyInput(evStar)
	if sl.GetMultiSelectionCount() != 2 {
		t.Errorf("expected '*' to select all servers in Datacenter group, got %d", sl.GetMultiSelectionCount())
	}
}

func TestFormatServerLine_CheckboxRendering(t *testing.T) {
	srv := domain.Server{
		Alias: "web-01",
		Host:  "192.168.1.50",
	}

	// 1. Unchecked (isSelected=false)
	primaryUnchecked, _ := formatServerLine(srv, 10, 80, false)
	if !strings.Contains(primaryUnchecked, "[ []") {
		t.Errorf("expected '[ []' (escaped [ ]) in unchecked line, got: %q", primaryUnchecked)
	}

	// 2. Checked (isSelected=true)
	primaryChecked, _ := formatServerLine(srv, 10, 80, true)
	if !strings.Contains(primaryChecked, "[✓]") {
		t.Errorf("expected '[✓]' in checked line, got: %q", primaryChecked)
	}

	// 3. Omitted isSelected: no checkbox rendered
	primaryNoCheck, _ := formatServerLine(srv, 10, 80)
	if strings.Contains(primaryNoCheck, "[✓]") || strings.Contains(primaryNoCheck, "[ []") {
		t.Errorf("expected no checkbox when isSelected omitted, got: %q", primaryNoCheck)
	}
}

func TestStatusBar_MultiSelectionBadge(t *testing.T) {
	app := &tui{
		app:        tview.NewApplication(),
		statusBar:  tview.NewTextView().SetDynamicColors(true),
		serverList: NewServerList(),
	}

	servers := []domain.Server{
		{Alias: "srv-1", Host: "10.0.0.1"},
		{Alias: "srv-2", Host: "10.0.0.2"},
	}
	app.serverList.UpdateServers(servers)

	// 0 selected: badge is empty
	if b := app.renderMultiSelectionBadge(); b != "" {
		t.Errorf("expected empty badge for 0 selected, got %q", b)
	}

	// 1 selected: [1 server selected]
	app.serverList.SetServerSelected("srv-1", true)
	b1 := app.renderMultiSelectionBadge()
	if !strings.Contains(b1, "1 server selected") {
		t.Errorf("expected '1 server selected', got %q", b1)
	}

	// 2 selected: [2 servers selected]
	app.serverList.SetServerSelected("srv-2", true)
	b2 := app.renderMultiSelectionBadge()
	if !strings.Contains(b2, "2 servers selected") {
		t.Errorf("expected '2 servers selected', got %q", b2)
	}

	// Verify status text includes the badge
	statusText := app.defaultStatusTextLocked()
	if !strings.Contains(statusText, "2 servers selected") {
		t.Errorf("expected status text to include badge, got %q", statusText)
	}

	// Callback updates status bar
	app.handleMultiSelectionChange(2)
	if !strings.Contains(app.statusBar.GetText(false), "2 servers selected") {
		t.Errorf("expected statusBar to show badge after callback, got %q", app.statusBar.GetText(false))
	}
}

func TestBulkOperations_PingAndTags(t *testing.T) {
	logger := zap.NewNop().Sugar()
	svc := &mockPingService{
		mockServerService: mockServerService{
			servers: []domain.Server{
				{Alias: "srv-1", Host: "10.0.0.1", Tags: []string{"web"}},
				{Alias: "srv-2", Host: "10.0.0.2", Tags: []string{"api"}},
				{Alias: "srv-3", Host: "10.0.0.3"},
			},
		},
	}
	app := NewTUI(logger, svc, "v1.0.0", "abc").(*tui)
	app.buildComponents()
	app.buildLayout()
	app.bindEvents()
	app.serverList.UpdateServers(svc.servers)
	app.app.SetRoot(app.root, true)
	app.app.SetFocus(app.serverList)

	// 1. Select srv-1 and srv-2
	app.serverList.SetServerSelected("srv-1", true)
	app.serverList.SetServerSelected("srv-2", true)

	// Trigger handlePingAll (should ping only the 2 selected)
	app.handlePingAll()
	if !strings.Contains(app.statusBar.GetText(false), "2 selected servers") {
		t.Errorf("expected status bar to say 'Pinging 2 selected servers', got %q", app.statusBar.GetText(false))
	}

	// 2. handleTagsEdit with multi-selection launches showBulkEditTagsForm
	app.handleTagsEdit()
	focused := app.app.GetFocus()
	if focused == nil {
		t.Fatalf("expected focus not to be nil after handleTagsEdit")
	}
	// In tview, focusing a Form delegates to its first InputField or Form itself
	_, isInputField := focused.(*tview.InputField)
	_, isForm := focused.(*tview.Form)
	if !isInputField && !isForm {
		t.Errorf("expected focus to be Form or InputField, got %T", focused)
	}

	// 3. Test form cancel returns to main
	app.returnToMain()
	if app.app.GetFocus() != app.serverList {
		t.Errorf("expected focus restored to serverList after returnToMain")
	}
}
