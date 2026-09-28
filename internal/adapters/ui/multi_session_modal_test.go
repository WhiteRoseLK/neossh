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
)

func TestMultiSessionModal_CreationAndTabs(t *testing.T) {
	app := tview.NewApplication()
	mockSvc := newMockSnippetService()

	targets := []domain.Server{
		{Alias: "srv-alpha", Host: "10.0.0.1"},
		{Alias: "srv-beta", Host: "10.0.0.2"},
		{Alias: "srv-gamma", Host: "10.0.0.3"},
	}

	modal := NewMultiSessionModal(app, mockSvc, targets, "uptime", func() {}, nil, nil)
	if modal == nil {
		t.Fatal("expected non-nil modal")
	}

	// 1 dashboard + 3 session tabs = 4 total tabs
	if len(modal.tabs) != 4 {
		t.Fatalf("expected 4 tabs, got %d", len(modal.tabs))
	}

	if modal.tabs[0].ID != "tab-dashboard" {
		t.Fatalf("expected first tab to be tab-dashboard, got %s", modal.tabs[0].ID)
	}

	if modal.currentTab != "tab-dashboard" {
		t.Fatalf("expected initial currentTab to be tab-dashboard, got %s", modal.currentTab)
	}

	// Dashboard table should have 3 rows + 1 header = 4 rows
	if modal.dashboardTable.GetRowCount() != 4 {
		t.Fatalf("expected 4 rows in dashboard table, got %d", modal.dashboardTable.GetRowCount())
	}
}

func TestMultiSessionModal_TabSwitching(t *testing.T) {
	app := tview.NewApplication()
	mockSvc := newMockSnippetService()

	targets := []domain.Server{
		{Alias: "srv-1", Host: "10.0.0.1"},
		{Alias: "srv-2", Host: "10.0.0.2"},
	}

	modal := NewMultiSessionModal(app, mockSvc, targets, "", func() {}, nil, nil)

	// Switch to session tab for srv-1
	modal.switchToTab("tab-session-srv-1")
	if modal.currentTab != "tab-session-srv-1" {
		t.Fatalf("expected currentTab to be tab-session-srv-1, got %s", modal.currentTab)
	}

	// Switch back to dashboard
	modal.switchToTab("tab-dashboard")
	if modal.currentTab != "tab-dashboard" {
		t.Fatalf("expected currentTab to be tab-dashboard, got %s", modal.currentTab)
	}
}

func TestMultiSessionModal_DashboardDrillDownAndSplit(t *testing.T) {
	app := tview.NewApplication()
	mockSvc := newMockSnippetService()

	targets := []domain.Server{
		{Alias: "web-01", Host: "10.0.1.1"},
		{Alias: "web-02", Host: "10.0.1.2"},
		{Alias: "db-01", Host: "10.0.2.1"},
		{Alias: "db-02", Host: "10.0.2.2"},
	}

	modal := NewMultiSessionModal(app, mockSvc, targets, "", func() {}, nil, nil)

	// 1. Single row drill-down (no checkboxes checked):
	// Cursor on row 3 (db-01)
	modal.dashboardTable.Select(3, 0)
	modal.handleDashboardEnter()
	if modal.currentTab != "tab-session-db-01" {
		t.Fatalf("expected drill-down to tab-session-db-01, got %s", modal.currentTab)
	}

	// Return to dashboard
	modal.switchToTab("tab-dashboard")

	// 2. Select 2 servers using Space
	// Check web-01 (row 1)
	modal.dashboardTable.Select(1, 0)
	modal.toggleDashboardRowCheck()
	// Check web-02 (row 2)
	modal.dashboardTable.Select(2, 0)
	modal.toggleDashboardRowCheck()

	if len(modal.checkedRows) < 2 || !modal.checkedRows["web-01"] || !modal.checkedRows["web-02"] {
		t.Fatalf("expected web-01 and web-02 to be checked, got %+v", modal.checkedRows)
	}

	// Press Enter -> opens Split View Tab with 2 panes!
	modal.handleDashboardEnter()
	if modal.currentTab != "tab-split" {
		t.Fatalf("expected currentTab to be tab-split, got %s", modal.currentTab)
	}
	if len(modal.splitPanes) != 2 {
		t.Fatalf("expected 2 split panes, got %d", len(modal.splitPanes))
	}

	// 3. Select 4 servers and launch 2x2 quadrant split view
	modal.switchToTab("tab-dashboard")
	modal.dashboardTable.Select(3, 0)
	modal.toggleDashboardRowCheck() // db-01
	modal.dashboardTable.Select(4, 0)
	modal.toggleDashboardRowCheck() // db-02

	modal.handleDashboardEnter()
	if modal.currentTab != "tab-split" {
		t.Fatalf("expected currentTab to be tab-split for 4 servers, got %s", modal.currentTab)
	}
	if len(modal.splitPanes) != 4 {
		t.Fatalf("expected 4 split panes (2x2 grid), got %d", len(modal.splitPanes))
	}
}

func TestMultiSessionModal_SplitPaneCycling(t *testing.T) {
	app := tview.NewApplication()
	mockSvc := newMockSnippetService()

	targets := []domain.Server{
		{Alias: "srv-a", Host: "10.0.0.1"},
		{Alias: "srv-b", Host: "10.0.0.2"},
	}

	modal := NewMultiSessionModal(app, mockSvc, targets, "", func() {}, nil, nil)
	modal.setupSplitTab(targets)

	if modal.focusedSplit != 0 {
		t.Fatalf("expected focusedSplit to start at 0, got %d", modal.focusedSplit)
	}

	// Press Tab -> focus moves to pane 1
	tabEv := tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone)
	modal.handleSplitKey(tabEv)
	if modal.focusedSplit != 1 {
		t.Fatalf("expected focusedSplit to be 1 after Tab, got %d", modal.focusedSplit)
	}

	// Press Tab again -> wraps around to pane 0
	modal.handleSplitKey(tabEv)
	if modal.focusedSplit != 0 {
		t.Fatalf("expected focusedSplit to wrap to 0, got %d", modal.focusedSplit)
	}

	// Press Esc -> returns to dashboard
	escEv := tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone)
	modal.handleSplitKey(escEv)
	if modal.currentTab != "tab-dashboard" {
		t.Fatalf("expected Esc in split view to return to dashboard, got %s", modal.currentTab)
	}
}

func TestMultiSessionModal_ExecutionAndInteractive(t *testing.T) {
	app := tview.NewApplication()
	mockSvc := newMockSnippetService()

	targets := []domain.Server{
		{Alias: "prod-web", Host: "10.0.0.1"},
	}

	var interactiveAlias, interactiveCmd string
	onInteractive := func(alias, cmd string) {
		interactiveAlias = alias
		interactiveCmd = cmd
	}

	modal := NewMultiSessionModal(app, mockSvc, targets, "df -h", func() {}, nil, onInteractive)

	// Wait briefly for execution
	time.Sleep(100 * time.Millisecond)

	res := modal.GetResult("prod-web")
	if res == nil || res.Running {
		t.Fatal("expected result for prod-web to be completed")
	}
	if !strings.Contains(res.Output, "prod-web") {
		t.Fatalf("expected result output to contain prod-web, got: %s", res.Output)
	}

	modal.RefreshViews()
	view := modal.sessionViews["prod-web"]
	if view == nil {
		t.Fatal("expected session view for prod-web")
	}

	txt := view.GetText(true)
	if !strings.Contains(txt, "prod-web") {
		t.Fatalf("expected session view text to contain prod-web output, got: %s", txt)
	}

	// Press 'i' on session tab to trigger interactive SSH
	inputHandler := view.GetInputCapture()
	if inputHandler != nil {
		iEv := tcell.NewEventKey(tcell.KeyRune, 'i', tcell.ModNone)
		inputHandler(iEv)
		if interactiveAlias != "prod-web" {
			t.Fatalf("expected onInteractive called with prod-web, got: %s", interactiveAlias)
		}
	}
	_ = interactiveCmd
}

func TestMultiSessionModal_SelectAllDashboard(t *testing.T) {
	app := tview.NewApplication()
	mockSvc := newMockSnippetService()

	targets := []domain.Server{
		{Alias: "srv1", Host: "10.0.0.1"},
		{Alias: "srv2", Host: "10.0.0.2"},
		{Alias: "srv3", Host: "10.0.0.3"},
	}

	modal := NewMultiSessionModal(app, mockSvc, targets, "", func() {}, nil, nil)

	// Toggle select all (Space on all)
	modal.toggleDashboardSelectAll()
	if len(modal.checkedRows) != 3 || !modal.checkedRows["srv1"] || !modal.checkedRows["srv2"] || !modal.checkedRows["srv3"] {
		t.Fatalf("expected all 3 servers checked, got %+v", modal.checkedRows)
	}

	// Toggle select all again -> all unchecked
	modal.toggleDashboardSelectAll()
	if modal.checkedRows["srv1"] || modal.checkedRows["srv2"] || modal.checkedRows["srv3"] {
		t.Fatalf("expected all servers unchecked, got %+v", modal.checkedRows)
	}
}
