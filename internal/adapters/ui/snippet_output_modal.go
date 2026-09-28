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
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/WhiteRoseLK/neossh/internal/core/domain"
	"github.com/WhiteRoseLK/neossh/internal/core/ports"
	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ServerExecResult holds the execution telemetry and result for a single remote server.
type ServerExecResult struct {
	Server   domain.Server
	Output   string
	Duration time.Duration
	Err      error
	Running  bool
}

// SnippetOutputModal displays live execution status and stdout/stderr output across target servers.
type SnippetOutputModal struct {
	*tview.Flex
	app           *tview.Application
	service       ports.ServerService
	targetServers []domain.Server
	command       string

	resultsMu sync.RWMutex
	results   map[string]*ServerExecResult

	headerText  *tview.TextView
	serverTable *tview.Table
	outputView  *tview.TextView
	helpText    *tview.TextView

	onClose  func()
	onStatus func(msg string, color string)
}

// NewSnippetOutputModal initializes the runner modal and kicks off concurrent execution.
func NewSnippetOutputModal(
	app *tview.Application,
	service ports.ServerService,
	targets []domain.Server,
	command string,
	onClose func(),
	onStatus func(msg, color string),
) *SnippetOutputModal {
	m := &SnippetOutputModal{
		Flex:          tview.NewFlex().SetDirection(tview.FlexRow),
		app:           app,
		service:       service,
		targetServers: targets,
		command:       command,
		results:       make(map[string]*ServerExecResult),
		headerText:    tview.NewTextView().SetDynamicColors(true),
		serverTable:   tview.NewTable().SetSelectable(true, false),
		outputView:    tview.NewTextView().SetDynamicColors(true).SetScrollable(true),
		helpText:      tview.NewTextView().SetDynamicColors(true),
		onClose:       onClose,
		onStatus:      onStatus,
	}

	for _, s := range targets {
		m.results[s.Alias] = &ServerExecResult{
			Server:  s,
			Running: true,
		}
	}

	m.setupUI()
	m.RunAll()
	return m
}

func (m *SnippetOutputModal) setupUI() {
	m.SetBorder(true).
		SetTitle(fmt.Sprintf(" ⚡ Execution Output (%d servers) ", len(m.targetServers))).
		SetTitleAlign(tview.AlignLeft)

	// Header banner
	escapedCmd := tview.Escape(m.command)
	m.headerText.SetText(fmt.Sprintf(" [yellow::b]Command:[-::-] [aqua]%s[-]\n [gray]Targets: %d server(s)[-]",
		escapedCmd, len(m.targetServers)))

	// Server Table (left)
	m.serverTable.SetBorder(true).SetTitle(" Target Servers ")
	m.serverTable.SetSelectionChangedFunc(func(row, column int) {
		m.updateOutputView(row)
	})

	// Output Viewer (right)
	m.outputView.SetBorder(true).SetTitle(" Output (stdout / stderr) ")

	// Main split
	mainSplit := tview.NewFlex().SetDirection(tview.FlexColumn)
	if len(m.targetServers) > 1 {
		mainSplit.AddItem(m.serverTable, 30, 0, true)
	}
	mainSplit.AddItem(m.outputView, 0, 1, len(m.targetServers) <= 1)

	// Help bar
	m.helpText.SetText(" [yellow]↑/↓[-]: Select server  [yellow]r[-]: Re-run  [yellow]c[-]: Copy output  [yellow]Esc/q[-]: Close")

	// Key bindings
	inputHandler := func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'q', 'Q':
			if m.onClose != nil {
				m.onClose()
			}
			return nil
		case 'r', 'R':
			m.RunAll()
			return nil
		case 'c', 'C':
			m.copyCurrentOutput()
			return nil
		}
		if event.Key() == tcell.KeyEscape {
			if m.onClose != nil {
				m.onClose()
			}
			return nil
		}
		return event
	}

	m.serverTable.SetInputCapture(inputHandler)
	m.outputView.SetInputCapture(inputHandler)
	m.SetInputCapture(inputHandler)

	m.AddItem(m.headerText, 3, 0, false)
	m.AddItem(mainSplit, 0, 1, true)
	m.AddItem(m.helpText, 1, 0, false)

	m.renderServerTable()
}

func (m *SnippetOutputModal) renderServerTable() {
	m.serverTable.Clear()

	startRow := 0
	if len(m.targetServers) > 1 {
		// "All Servers" summary row
		allCell := tview.NewTableCell(" 🌐 [All Servers (Combined)]").
			SetTextColor(tcell.ColorYellow).
			SetSelectable(true)
		m.serverTable.SetCell(0, 0, allCell)
		startRow = 1
	}

	m.resultsMu.RLock()
	defer m.resultsMu.RUnlock()

	for i, s := range m.targetServers {
		row := startRow + i
		res := m.results[s.Alias]

		var text string
		var color tcell.Color

		switch {
		case res == nil || res.Running:
			text = fmt.Sprintf(" ⏳ %s (running...)", s.Alias)
			color = tcell.ColorYellow
		case res.Err == nil:
			text = fmt.Sprintf(" ✓ %s (%.2fs)", s.Alias, res.Duration.Seconds())
			color = tcell.ColorGreen
		default:
			text = fmt.Sprintf(" ✗ %s (error)", s.Alias)
			color = tcell.ColorRed
		}

		cell := tview.NewTableCell(text).
			SetTextColor(color).
			SetSelectable(true)
		m.serverTable.SetCell(row, 0, cell)
	}

	row, _ := m.serverTable.GetSelection()
	m.updateOutputView(row)
}

func (m *SnippetOutputModal) updateOutputView(row int) {
	if len(m.targetServers) == 0 {
		m.outputView.SetText("[gray]No servers targeted.[-]")
		return
	}

	m.resultsMu.RLock()
	defer m.resultsMu.RUnlock()

	// If multiple servers and row 0 is "[All Servers]"
	if len(m.targetServers) > 1 && row == 0 {
		var sb strings.Builder
		sb.WriteString("[yellow::b]=== Combined Output Across All Servers ===[-::-]\n\n")
		for _, s := range m.targetServers {
			res := m.results[s.Alias]
			if res == nil {
				continue
			}
			var statusBadge string
			switch {
			case res.Running:
				statusBadge = "[yellow]⏳ RUNNING...[-]"
			case res.Err != nil:
				statusBadge = fmt.Sprintf("[red]✗ FAILED (%v)[-]", res.Err)
			default:
				statusBadge = "[green]✓ SUCCESS[-]"
			}

			sb.WriteString(fmt.Sprintf("[teal::b]─── %s (%s) • %s ───[-::-]\n", s.Alias, s.Host, statusBadge))
			switch {
			case res.Running:
				sb.WriteString("[gray](Waiting for output...)[-]\n\n")
			case strings.TrimSpace(res.Output) == "":
				sb.WriteString("[gray](No output produced)[-]\n\n")
			default:
				sb.WriteString(tview.Escape(res.Output))
				if !strings.HasSuffix(res.Output, "\n") {
					sb.WriteString("\n")
				}
				sb.WriteString("\n")
			}
		}
		m.outputView.SetText(sb.String())
		m.outputView.ScrollToBeginning()
		return
	}

	// Single server or specific server row
	targetIdx := row
	if len(m.targetServers) > 1 {
		targetIdx = row - 1
	}
	if targetIdx < 0 || targetIdx >= len(m.targetServers) {
		targetIdx = 0
	}

	srv := m.targetServers[targetIdx]
	res := m.results[srv.Alias]
	if res == nil {
		m.outputView.SetText("[gray]No data.[-]")
		return
	}

	var sb strings.Builder
	var statusBadge string
	switch {
	case res.Running:
		statusBadge = "[yellow]⏳ RUNNING...[-]"
	case res.Err != nil:
		statusBadge = fmt.Sprintf("[red]✗ FAILED: %s[-]", tview.Escape(res.Err.Error()))
	default:
		statusBadge = fmt.Sprintf("[green]✓ SUCCESS (in %.2fs)[-]", res.Duration.Seconds())
	}

	sb.WriteString(fmt.Sprintf("[yellow::b]Server:[-::-] [white]%s (%s)[-]\n", srv.Alias, srv.Host))
	sb.WriteString(fmt.Sprintf("[yellow::b]Status:[-::-] %s\n", statusBadge))
	sb.WriteString("[gray]──────────────────────────────────────────────[-]\n\n")

	switch {
	case res.Running:
		sb.WriteString("[gray]Executing remote command, please wait...[-]\n")
	case strings.TrimSpace(res.Output) == "":
		sb.WriteString("[gray](Command finished with no output)[-]\n")
	default:
		sb.WriteString(tview.Escape(res.Output))
	}

	m.outputView.SetText(sb.String())
	m.outputView.ScrollToBeginning()
}

// RunAll kicks off concurrent execution across all target servers.
func (m *SnippetOutputModal) RunAll() {
	if len(m.targetServers) == 0 {
		return
	}

	m.resultsMu.Lock()
	for _, s := range m.targetServers {
		m.results[s.Alias] = &ServerExecResult{
			Server:  s,
			Running: true,
		}
	}
	m.resultsMu.Unlock()

	m.renderServerTable()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	var wg sync.WaitGroup
	for _, srv := range m.targetServers {
		wg.Add(1)
		go func(s domain.Server) {
			defer wg.Done()
			start := time.Now()

			type execRes struct {
				out string
				err error
			}
			resChan := make(chan execRes, 1)

			go func() {
				out, err := m.service.ExecuteRemoteCommand(s.Alias, m.command)
				resChan <- execRes{out: out, err: err}
			}()

			var out string
			var err error

			select {
			case <-ctx.Done():
				err = ctx.Err()
				out = fmt.Sprintf("execution timed out after 30s: %v", err)
			case r := <-resChan:
				out = r.out
				err = r.err
			}

			dur := time.Since(start)

			m.resultsMu.Lock()
			if res, ok := m.results[s.Alias]; ok {
				res.Running = false
				res.Output = out
				res.Duration = dur
				res.Err = err
			}
			m.resultsMu.Unlock()

			if m.app != nil {
				m.app.QueueUpdateDraw(func() {
					m.renderServerTable()
				})
			}
		}(srv)
	}

	go func() {
		wg.Wait()
		cancel()
		if m.app != nil {
			m.app.QueueUpdateDraw(func() {
				m.renderServerTable()
			})
		}
	}()
}

func (m *SnippetOutputModal) copyCurrentOutput() {
	text := m.outputView.GetText(true)
	if text == "" {
		return
	}
	if err := clipboard.WriteAll(text); err == nil {
		if m.onStatus != nil {
			m.onStatus("Output copied to clipboard", "#50FA7B")
		}
	} else if m.onStatus != nil {
		m.onStatus("Failed to copy output: "+err.Error(), "#FF6B6B")
	}
}
