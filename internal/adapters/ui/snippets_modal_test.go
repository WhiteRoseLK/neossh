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
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/WhiteRoseLK/neossh/internal/core/domain"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type mockSnippetService struct {
	mockReadOnlyService
	mu           sync.Mutex
	snippets     []domain.Snippet
	execCalls    map[string]int
	execOutputs  map[string]string
	execErrors   map[string]error
	interactCall string
}

func newMockSnippetService() *mockSnippetService {
	return &mockSnippetService{
		snippets: []domain.Snippet{
			{
				ID:          "snip-1",
				Name:        "System Stats",
				Command:     "uname -a && uptime",
				Description: "View kernel version and system load",
				Tags:        []string{"sysadmin", "monitoring"},
			},
			{
				ID:          "snip-2",
				Name:        "Restart Service",
				Command:     "sudo systemctl restart {{service}}",
				Description: "Restart specified systemd daemon",
				Tags:        []string{"systemd", "ops"},
			},
			{
				ID:          "snip-3",
				Name:        "Disk Free",
				Command:     "df -h",
				Description: "Inspect filesystem space usage",
				Tags:        []string{"storage"},
			},
		},
		execCalls:   make(map[string]int),
		execOutputs: make(map[string]string),
		execErrors:  make(map[string]error),
	}
}

func (m *mockSnippetService) GetSnippets() ([]domain.Snippet, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	res := make([]domain.Snippet, len(m.snippets))
	copy(res, m.snippets)
	return res, nil
}

func (m *mockSnippetService) SaveSnippet(s domain.Snippet) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, existing := range m.snippets {
		if existing.ID == s.ID {
			m.snippets[i] = s
			return nil
		}
	}
	m.snippets = append(m.snippets, s)
	return nil
}

func (m *mockSnippetService) DeleteSnippet(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, s := range m.snippets {
		if s.ID == id {
			m.snippets = append(m.snippets[:i], m.snippets[i+1:]...)
			return nil
		}
	}
	return nil
}

func (m *mockSnippetService) ExecuteRemoteCommand(alias string, command string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.execCalls[alias]++
	if err, ok := m.execErrors[alias]; ok && err != nil {
		return m.execOutputs[alias], err
	}
	if out, ok := m.execOutputs[alias]; ok {
		return out, nil
	}
	return fmt.Sprintf("output from %s: %s", alias, command), nil
}

func (m *mockSnippetService) RunInteractiveRemoteCommand(alias string, command string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.interactCall = fmt.Sprintf("%s:%s", alias, command)
	return nil
}

func TestSnippetsModal_CreationAndLoad(t *testing.T) {
	app := tview.NewApplication()
	mockSvc := newMockSnippetService()

	targets := []domain.Server{
		{Alias: "srv-prod-1", Host: "10.0.0.1"},
		{Alias: "srv-prod-2", Host: "10.0.0.2"},
	}

	var statusMsg string
	onStatus := func(msg, color string) {
		statusMsg = msg
	}

	modal := NewSnippetsModal(app, mockSvc, targets, func() {}, onStatus, nil)
	if modal == nil {
		t.Fatal("expected non-nil modal")
	}

	// Check table has 3 snippets + 1 header = 4 rows
	if modal.table.GetRowCount() != 4 {
		t.Fatalf("expected 4 table rows, got %d", modal.table.GetRowCount())
	}

	// Verify header text mentions 2 servers selected
	header := modal.headerText.GetText(true)
	if !strings.Contains(header, "2 servers selected") {
		t.Fatalf("expected header to contain '2 servers selected', got: %s", header)
	}

	// Verify details view has content for selected row
	details := modal.detailsView.GetText(true)
	if !strings.Contains(details, "System Stats") {
		t.Fatalf("expected details to contain 'System Stats', got: %s", details)
	}

	_ = statusMsg
}

func TestSnippetsModal_Filter(t *testing.T) {
	app := tview.NewApplication()
	mockSvc := newMockSnippetService()

	modal := NewSnippetsModal(app, mockSvc, nil, func() {}, nil, nil)

	// Filter by tag "monitoring"
	modal.applyFilter("monitoring")
	if len(modal.filtered) != 1 || modal.filtered[0].Name != "System Stats" {
		t.Fatalf("expected 1 match for 'monitoring', got %d", len(modal.filtered))
	}

	// Filter by command "systemctl"
	modal.applyFilter("systemctl")
	if len(modal.filtered) != 1 || modal.filtered[0].Name != "Restart Service" {
		t.Fatalf("expected 1 match for 'systemctl', got %d", len(modal.filtered))
	}

	// Filter with non-matching query
	modal.applyFilter("nonexistent123")
	if len(modal.filtered) != 0 {
		t.Fatalf("expected 0 matches, got %d", len(modal.filtered))
	}

	// Reset filter
	modal.applyFilter("")
	if len(modal.filtered) != 3 {
		t.Fatalf("expected 3 items after clearing filter, got %d", len(modal.filtered))
	}
}

func TestSnippetsModal_PlaceholdersPrompt(t *testing.T) {
	app := tview.NewApplication()
	mockSvc := newMockSnippetService()

	target := []domain.Server{{Alias: "srv1", Host: "1.1.1.1"}}
	var interactiveCalled bool
	onInteractive := func(alias, command string) {
		interactiveCalled = true
	}

	modal := NewSnippetsModal(app, mockSvc, target, func() {}, nil, onInteractive)

	// Select row 2 ("Restart Service" which has {{service}})
	modal.table.Select(2, 0)
	modal.handleRunSelected(false)

	// Should have transitioned focus to parameter prompt form
	if form, ok := app.GetFocus().(*tview.Form); ok {
		if form.GetButtonCount() == 0 {
			t.Fatal("expected buttons in parameter prompt form")
		}
	}
	_ = interactiveCalled
}

func TestSnippetsModal_CRUD(t *testing.T) {
	app := tview.NewApplication()
	mockSvc := newMockSnippetService()

	modal := NewSnippetsModal(app, mockSvc, nil, func() {}, nil, nil)

	// 1. Add snippet
	newSnip := domain.Snippet{
		ID:          "snip-new",
		Name:        "Docker PS",
		Command:     "docker ps -a",
		Description: "List containers",
		Tags:        []string{"docker"},
	}
	if err := mockSvc.SaveSnippet(newSnip); err != nil {
		t.Fatalf("failed to save snippet: %v", err)
	}
	modal.LoadSnippets()

	if len(modal.snippets) != 4 {
		t.Fatalf("expected 4 snippets after add, got %d", len(modal.snippets))
	}

	// 2. Edit snippet
	newSnip.Name = "Docker Containers"
	if err := mockSvc.SaveSnippet(newSnip); err != nil {
		t.Fatalf("failed to update snippet: %v", err)
	}
	modal.LoadSnippets()

	found := false
	for _, s := range modal.snippets {
		if s.ID == "snip-new" && s.Name == "Docker Containers" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected edited snippet name 'Docker Containers'")
	}

	// 3. Delete snippet
	if err := mockSvc.DeleteSnippet("snip-new"); err != nil {
		t.Fatalf("failed to delete snippet: %v", err)
	}
	modal.LoadSnippets()

	if len(modal.snippets) != 3 {
		t.Fatalf("expected 3 snippets after delete, got %d", len(modal.snippets))
	}
}

func TestSnippetOutputModal_ConcurrentExecution(t *testing.T) {
	app := tview.NewApplication()
	mockSvc := newMockSnippetService()

	targets := []domain.Server{
		{Alias: "srv-alpha", Host: "10.0.1.1"},
		{Alias: "srv-beta", Host: "10.0.1.2"},
		{Alias: "srv-gamma", Host: "10.0.1.3"},
	}

	mockSvc.execOutputs["srv-alpha"] = "Linux alpha 5.15\n"
	mockSvc.execOutputs["srv-beta"] = "Linux beta 5.15\n"
	mockSvc.execErrors["srv-gamma"] = fmt.Errorf("connection refused")

	var statusMsgs []string
	onStatus := func(msg, color string) {
		statusMsgs = append(statusMsgs, msg)
	}

	var closed bool
	onClose := func() {
		closed = true
	}

	outputModal := NewSnippetOutputModal(app, mockSvc, targets, "uname -a", onClose, onStatus)
	if outputModal == nil {
		t.Fatal("expected non-nil output modal")
	}

	// Wait briefly for goroutines to finish
	time.Sleep(100 * time.Millisecond)

	// Verify all 3 servers were called
	mockSvc.mu.Lock()
	if mockSvc.execCalls["srv-alpha"] != 1 || mockSvc.execCalls["srv-beta"] != 1 || mockSvc.execCalls["srv-gamma"] != 1 {
		t.Fatalf("expected 1 call per server, got %+v", mockSvc.execCalls)
	}
	mockSvc.mu.Unlock()

	// Verify Combined view (row 0)
	outputModal.updateOutputView(0)
	allText := outputModal.outputView.GetText(true)
	if !strings.Contains(allText, "srv-alpha") || !strings.Contains(allText, "srv-beta") || !strings.Contains(allText, "srv-gamma") {
		t.Fatalf("expected combined output to contain all servers, got: %s", allText)
	}

	// Verify Single server view (row 1 is srv-alpha)
	outputModal.updateOutputView(1)
	alphaText := outputModal.outputView.GetText(true)
	if !strings.Contains(alphaText, "Linux alpha 5.15") {
		t.Fatalf("expected srv-alpha output, got: %s", alphaText)
	}

	// Verify Failed server view (row 3 is srv-gamma)
	outputModal.updateOutputView(3)
	gammaText := outputModal.outputView.GetText(true)
	if !strings.Contains(gammaText, "FAILED") {
		t.Fatalf("expected srv-gamma output to show FAILED, got: %s", gammaText)
	}

	// Test close
	ev := tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone)
	if handler := outputModal.serverTable.GetInputCapture(); handler != nil {
		handler(ev)
	}
	if !closed {
		t.Fatal("expected modal to be closed on Escape")
	}
}

func TestSnippetsModal_InputCaptureKeys(t *testing.T) {
	app := tview.NewApplication()
	mockSvc := newMockSnippetService()
	var closed bool

	modal := NewSnippetsModal(app, mockSvc, nil, func() { closed = true }, nil, nil)

	tableHandler := modal.table.GetInputCapture()
	if tableHandler == nil {
		t.Fatal("expected table input capture handler")
	}

	// Pressing '/' focuses filterInput
	slashEv := tcell.NewEventKey(tcell.KeyRune, '/', tcell.ModNone)
	res := tableHandler(slashEv)
	if res != nil {
		t.Fatal("expected nil event (consumed)")
	}

	// Pressing Esc calls onClose
	escEv := tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone)
	tableHandler(escEv)
	if !closed {
		t.Fatal("expected closed to be true on Escape")
	}
}
