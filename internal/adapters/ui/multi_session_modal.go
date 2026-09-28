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

// TabInfo stores metadata for a top-level tab in MultiSessionModal.
type TabInfo struct {
	ID    string
	Title string
	Type  string // "dashboard", "session", "split"
	Alias string
}

// MultiSessionModal provides a multi-session engine with an overview dashboard and interactive tabs.
type MultiSessionModal struct {
	*tview.Flex
	app           *tview.Application
	service       ports.ServerService
	targetServers []domain.Server
	command       string

	// Tab system
	tabBar     *tview.TextView
	pages      *tview.Pages
	hintBar    *tview.TextView
	currentTab string
	tabs       []TabInfo

	// State
	resultsMu sync.RWMutex
	results   map[string]*ServerExecResult
	outputs   map[string]*strings.Builder

	// Dashboard (Tab 0)
	dashboardTable *tview.Table
	dashboardHead  *tview.TextView
	checkedRows    map[string]bool // server alias -> checked

	// Session tabs
	sessionViews map[string]*tview.TextView

	// Split view tab
	splitFlex    *tview.Flex
	splitPanes   []*tview.TextView
	splitServers []domain.Server
	focusedSplit int

	onClose       func()
	onStatus      func(msg, color string)
	onInteractive func(alias, cmd string)
}

// NewMultiSessionModal creates an interactive multi-session dashboard and tabbed runner dialog.
func NewMultiSessionModal(
	app *tview.Application,
	service ports.ServerService,
	targets []domain.Server,
	command string,
	onClose func(),
	onStatus func(msg, color string),
	onInteractive func(alias, cmd string),
) *MultiSessionModal {
	m := &MultiSessionModal{
		Flex:           tview.NewFlex().SetDirection(tview.FlexRow),
		app:            app,
		service:        service,
		targetServers:  targets,
		command:        command,
		tabBar:         tview.NewTextView().SetDynamicColors(true),
		pages:          tview.NewPages(),
		hintBar:        tview.NewTextView().SetDynamicColors(true),
		results:        make(map[string]*ServerExecResult),
		outputs:        make(map[string]*strings.Builder),
		sessionViews:   make(map[string]*tview.TextView),
		checkedRows:    make(map[string]bool),
		dashboardTable: tview.NewTable().SetSelectable(true, false),
		dashboardHead:  tview.NewTextView().SetDynamicColors(true),
		splitFlex:      tview.NewFlex(),
		onClose:        onClose,
		onStatus:       onStatus,
		onInteractive:  onInteractive,
	}

	for _, s := range targets {
		m.results[s.Alias] = &ServerExecResult{
			Server:  s,
			Running: command != "",
		}
		m.outputs[s.Alias] = &strings.Builder{}
	}

	m.setupUI()
	if command != "" {
		m.RunCommand(command)
	}
	return m
}

func (m *MultiSessionModal) setupUI() {
	m.SetBorder(true).
		SetTitle(fmt.Sprintf(" 🖥️ Multi-Session Engine (%d servers) ", len(m.targetServers))).
		SetTitleAlign(tview.AlignLeft)

	// Tab Bar styling
	m.tabBar.SetBackgroundColor(CurrentTheme.StatusBarBackground)
	m.tabBar.SetTextAlign(tview.AlignLeft)

	// Hint Bar styling
	m.hintBar.SetBackgroundColor(CurrentTheme.StatusBarBackground)
	m.hintBar.SetTextAlign(tview.AlignCenter)

	m.setupDashboard()
	m.setupSessionTabs()

	// Initial active tab
	m.currentTab = "tab-dashboard"
	m.updateTabBar()
	m.updateHintBar()

	m.AddItem(m.tabBar, 1, 0, false)
	m.AddItem(m.pages, 0, 1, true)
	m.AddItem(m.hintBar, 1, 0, false)

	// Global key capture for switching tabs
	m.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Modifiers()&tcell.ModAlt != 0 {
			switch event.Rune() {
			case '0':
				m.switchToTab("tab-dashboard")
				return nil
			case '1', '2', '3', '4', '5', '6', '7', '8', '9':
				idx := int(event.Rune() - '0')
				if idx < len(m.tabs) {
					m.switchToTab(m.tabs[idx].ID)
					return nil
				}
			}
		}

		if event.Key() == tcell.KeyEscape {
			if m.currentTab != "tab-dashboard" {
				m.switchToTab("tab-dashboard")
				return nil
			}
			if m.onClose != nil {
				m.onClose()
			}
			return nil
		}
		return event
	})
}

func (m *MultiSessionModal) setupDashboard() {
	// Tab 0 metadata
	m.tabs = append(m.tabs, TabInfo{
		ID:    "tab-dashboard",
		Title: "[0] Dashboard",
		Type:  "dashboard",
	})

	dashFlex := tview.NewFlex().SetDirection(tview.FlexRow)

	m.dashboardHead.SetDynamicColors(true)
	m.updateDashboardHeader()

	m.dashboardTable.SetBorder(true).SetTitle(" Target Servers (Space: Select, Enter: Drill-down / Split) ")
	m.dashboardTable.SetInputCapture(m.handleDashboardKey)
	m.dashboardTable.SetSelectedFunc(func(row, column int) {
		m.handleDashboardEnter()
	})

	dashFlex.AddItem(m.dashboardHead, 3, 0, false)
	dashFlex.AddItem(m.dashboardTable, 0, 1, true)

	m.pages.AddPage("tab-dashboard", dashFlex, true, true)
	m.renderDashboardTable()
}

func (m *MultiSessionModal) setupSessionTabs() {
	for i, s := range m.targetServers {
		tabID := fmt.Sprintf("tab-session-%s", s.Alias)
		tabTitle := fmt.Sprintf("[%d] %s", i+1, s.Alias)

		m.tabs = append(m.tabs, TabInfo{
			ID:    tabID,
			Title: tabTitle,
			Type:  "session",
			Alias: s.Alias,
		})

		sessFlex := tview.NewFlex().SetDirection(tview.FlexRow)

		topInfo := tview.NewTextView().SetDynamicColors(true)
		topInfo.SetText(fmt.Sprintf(" [yellow::b]Server:[-::-] [white]%s (%s)[-]  •  [yellow]i[-]: Interactive SSH  •  [yellow]r[-]: Re-run  •  [yellow]c[-]: Copy  •  [yellow]Esc[-]: Dashboard",
			s.Alias, s.Host))

		outView := tview.NewTextView().SetDynamicColors(true).SetScrollable(true)
		outView.SetBorder(true).SetTitle(fmt.Sprintf(" %s Output ", s.Alias))
		m.sessionViews[s.Alias] = outView

		outView.SetInputCapture(func(srv domain.Server) func(event *tcell.EventKey) *tcell.EventKey {
			return func(event *tcell.EventKey) *tcell.EventKey {
				switch event.Rune() {
				case 'i', 'I':
					if m.onInteractive != nil {
						m.onInteractive(srv.Alias, m.command)
					}
					return nil
				case 'r', 'R':
					if m.command != "" {
						m.runOnSingleServer(srv, m.command)
					}
					return nil
				case 'c', 'C':
					m.copyOutputForServer(srv.Alias)
					return nil
				}
				if event.Key() == tcell.KeyEscape {
					m.switchToTab("tab-dashboard")
					return nil
				}
				return event
			}
		}(s))

		sessFlex.AddItem(topInfo, 1, 0, false)
		sessFlex.AddItem(outView, 0, 1, true)

		m.pages.AddPage(tabID, sessFlex, true, false)
	}
}

func (m *MultiSessionModal) setupSplitTab(servers []domain.Server) {
	if len(servers) < 2 || len(servers) > 4 {
		return
	}

	splitTabID := "tab-split"
	splitTitle := fmt.Sprintf("[S] Split (%d)", len(servers))

	// Remove old split tab if exists
	foundIdx := -1
	for i, t := range m.tabs {
		if t.ID == splitTabID {
			foundIdx = i
			break
		}
	}
	if foundIdx != -1 {
		m.tabs = append(m.tabs[:foundIdx], m.tabs[foundIdx+1:]...)
	}

	m.tabs = append(m.tabs, TabInfo{
		ID:    splitTabID,
		Title: splitTitle,
		Type:  "split",
	})

	m.splitServers = servers
	m.splitPanes = make([]*tview.TextView, len(servers))
	m.focusedSplit = 0

	for i, s := range servers {
		pane := tview.NewTextView().SetDynamicColors(true).SetScrollable(true)
		pane.SetBorder(true).
			SetTitle(fmt.Sprintf(" %s (%s) ", s.Alias, s.Host))
		m.updatePaneOutput(pane, s.Alias)
		m.splitPanes[i] = pane
	}

	m.splitFlex = tview.NewFlex().SetDirection(tview.FlexRow)

	switch len(servers) {
	case 2:
		cols := tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(m.splitPanes[0], 0, 1, true).
			AddItem(m.splitPanes[1], 0, 1, false)
		m.splitFlex.AddItem(cols, 0, 1, true)
	case 3:
		topCols := tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(m.splitPanes[0], 0, 1, true).
			AddItem(m.splitPanes[1], 0, 1, false)
		m.splitFlex.AddItem(topCols, 0, 1, true).
			AddItem(m.splitPanes[2], 0, 1, false)
	case 4:
		topCols := tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(m.splitPanes[0], 0, 1, true).
			AddItem(m.splitPanes[1], 0, 1, false)
		botCols := tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(m.splitPanes[2], 0, 1, false).
			AddItem(m.splitPanes[3], 0, 1, false)
		m.splitFlex.AddItem(topCols, 0, 1, true).
			AddItem(botCols, 0, 1, false)
	}

	m.splitFlex.SetInputCapture(m.handleSplitKey)

	m.pages.AddPage(splitTabID, m.splitFlex, true, false)
	m.updateSplitBorders()
	m.switchToTab(splitTabID)
}

func (m *MultiSessionModal) updateSplitBorders() {
	for i, pane := range m.splitPanes {
		if i == m.focusedSplit {
			pane.SetBorderColor(tcell.ColorYellow)
			pane.SetTitleColor(tcell.ColorYellow)
			m.app.SetFocus(pane)
		} else {
			pane.SetBorderColor(tcell.ColorGray)
			pane.SetTitleColor(tcell.ColorWhite)
		}
	}
}

func (m *MultiSessionModal) handleSplitKey(event *tcell.EventKey) *tcell.EventKey {
	//nolint:exhaustive // Only relevant navigation keys handled
	switch event.Key() {
	case tcell.KeyTab:
		m.focusedSplit = (m.focusedSplit + 1) % len(m.splitPanes)
		m.updateSplitBorders()
		return nil
	case tcell.KeyBacktab:
		m.focusedSplit = (m.focusedSplit - 1 + len(m.splitPanes)) % len(m.splitPanes)
		m.updateSplitBorders()
		return nil
	case tcell.KeyEscape:
		m.switchToTab("tab-dashboard")
		return nil
	default:
	}

	switch event.Rune() {
	case 'i', 'I':
		if m.focusedSplit < len(m.splitServers) && m.onInteractive != nil {
			m.onInteractive(m.splitServers[m.focusedSplit].Alias, m.command)
		}
		return nil
	case 'c', 'C':
		if m.focusedSplit < len(m.splitServers) {
			m.copyOutputForServer(m.splitServers[m.focusedSplit].Alias)
		}
		return nil
	case 'r', 'R':
		if m.command != "" {
			for _, s := range m.splitServers {
				m.runOnSingleServer(s, m.command)
			}
		}
		return nil
	}
	return event
}

func (m *MultiSessionModal) updatePaneOutput(pane *tview.TextView, alias string) {
	m.resultsMu.RLock()
	defer m.resultsMu.RUnlock()

	res := m.results[alias]
	if res == nil {
		pane.SetText("[gray]No output.[-]")
		return
	}

	var sb strings.Builder
	switch {
	case res.Running:
		sb.WriteString("[yellow]⏳ Running...[-]\n\n")
	case res.Err != nil:
		sb.WriteString(fmt.Sprintf("[red]✗ Error: %s[-]\n\n", tview.Escape(res.Err.Error())))
	default:
		sb.WriteString(fmt.Sprintf("[green]✓ Succeeded in %.2fs[-]\n\n", res.Duration.Seconds()))
	}

	if strings.TrimSpace(res.Output) != "" {
		sb.WriteString(tview.Escape(res.Output))
	}
	pane.SetText(sb.String())
}

func (m *MultiSessionModal) updateTabBar() {
	var sb strings.Builder
	for i, tab := range m.tabs {
		if i > 0 {
			sb.WriteString(" ")
		}
		if tab.ID == m.currentTab {
			sb.WriteString(fmt.Sprintf("[black:yellow:b] %s [-:-:-]", tview.Escape(tab.Title)))
		} else {
			sb.WriteString(fmt.Sprintf("[gray] %s [-]", tview.Escape(tab.Title)))
		}
	}
	m.tabBar.SetText(sb.String())
}

func (m *MultiSessionModal) updateHintBar() {
	switch {
	case m.currentTab == "tab-dashboard":
		m.hintBar.SetText(" [yellow]Space[-]: Check  [yellow]Enter[-]: Drill-down/Split  [yellow]Alt+0..9[-]: Tab  [yellow]r[-]: Re-run all  [yellow]x[-]: Ad-hoc  [yellow]Esc[-]: Close")
	case m.currentTab == "tab-split":
		m.hintBar.SetText(" [yellow]Tab[-]: Next Pane  [yellow]i[-]: Interactive SSH  [yellow]r[-]: Re-run split  [yellow]c[-]: Copy  [yellow]Esc[-]: Dashboard")
	default:
		m.hintBar.SetText(" [yellow]i[-]: Interactive SSH Shell  [yellow]r[-]: Re-run  [yellow]c[-]: Copy output  [yellow]Alt+0[-]: Dashboard  [yellow]Esc[-]: Back")
	}
}

func (m *MultiSessionModal) switchToTab(tabID string) {
	m.currentTab = tabID
	m.pages.SwitchToPage(tabID)
	m.updateTabBar()
	m.updateHintBar()

	switch tabID {
	case "tab-dashboard":
		m.app.SetFocus(m.dashboardTable)
	case "tab-split":
		m.updateSplitBorders()
	default:
		// Find session alias
		for _, t := range m.tabs {
			if t.ID == tabID && t.Alias != "" {
				if view, ok := m.sessionViews[t.Alias]; ok {
					m.app.SetFocus(view)
				}
				break
			}
		}
	}
}

func (m *MultiSessionModal) updateDashboardHeader() {
	m.resultsMu.RLock()
	var successCount, failCount, runningCount int
	for _, res := range m.results {
		switch {
		case res.Running:
			runningCount++
		case res.Err != nil:
			failCount++
		default:
			successCount++
		}
	}
	m.resultsMu.RUnlock()

	var cmdDisplay string
	if m.command != "" {
		cmdDisplay = fmt.Sprintf("Command: [aqua]%s[-] • ", tview.Escape(m.command))
	} else {
		cmdDisplay = "Overview • "
	}

	selectedCount := 0
	for _, checked := range m.checkedRows {
		if checked {
			selectedCount++
		}
	}

	selectedBadge := ""
	if selectedCount > 0 {
		selectedBadge = fmt.Sprintf(" • [yellow::b]%d selected for split[-::-]", selectedCount)
	}

	m.dashboardHead.SetText(fmt.Sprintf(" [yellow::b]%s[-::-]Total: %d servers [gray]([green]✓ %d Done[-] [yellow]⏳ %d Running[-] [red]✗ %d Failed[-])[gray]%s[-]",
		cmdDisplay, len(m.targetServers), successCount, runningCount, failCount, selectedBadge))
}

func (m *MultiSessionModal) renderDashboardTable() {
	m.dashboardTable.Clear()

	headers := []string{"Sel", "Server", "Host", "Status", "Duration", "Output Preview"}
	for col, h := range headers {
		cell := tview.NewTableCell(" " + h + " ").
			SetTextColor(tcell.ColorYellow).
			SetSelectable(false)
		m.dashboardTable.SetCell(0, col, cell)
	}

	m.resultsMu.RLock()
	defer m.resultsMu.RUnlock()

	for i, s := range m.targetServers {
		row := i + 1
		checked := m.checkedRows[s.Alias]

		checkStr := "[ ]"
		checkColor := tcell.ColorGray
		if checked {
			checkStr = "[✓]"
			checkColor = tcell.ColorTeal
		}

		res := m.results[s.Alias]

		var statusStr string
		var statusColor tcell.Color
		var durStr string
		var previewStr string

		switch {
		case res == nil || res.Running:
			statusStr = "⏳ Running"
			statusColor = tcell.ColorYellow
			durStr = "-"
			previewStr = "(executing...)"
		case res.Err != nil:
			statusStr = "✗ Failed"
			statusColor = tcell.ColorRed
			durStr = fmt.Sprintf("%.2fs", res.Duration.Seconds())
			previewStr = res.Err.Error()
		default:
			statusStr = "✓ Success"
			statusColor = tcell.ColorGreen
			durStr = fmt.Sprintf("%.2fs", res.Duration.Seconds())
			lines := strings.Split(strings.TrimSpace(res.Output), "\n")
			if len(lines) > 0 {
				previewStr = lines[len(lines)-1]
			}
		}

		if len(previewStr) > 40 {
			previewStr = previewStr[:37] + "..."
		}

		m.dashboardTable.SetCell(row, 0, tview.NewTableCell(" "+checkStr).SetTextColor(checkColor).SetSelectable(true))
		m.dashboardTable.SetCell(row, 1, tview.NewTableCell(" "+s.Alias).SetTextColor(tcell.ColorWhite).SetSelectable(true))
		m.dashboardTable.SetCell(row, 2, tview.NewTableCell(" "+s.Host).SetTextColor(tcell.ColorGray).SetSelectable(true))
		m.dashboardTable.SetCell(row, 3, tview.NewTableCell(" "+statusStr).SetTextColor(statusColor).SetSelectable(true))
		m.dashboardTable.SetCell(row, 4, tview.NewTableCell(" "+durStr).SetTextColor(tcell.ColorLightGray).SetSelectable(true))
		m.dashboardTable.SetCell(row, 5, tview.NewTableCell(" "+previewStr).SetTextColor(tcell.ColorLightSlateGray).SetSelectable(true))
	}

	if len(m.targetServers) > 0 {
		m.dashboardTable.Select(1, 0)
	}
}

func (m *MultiSessionModal) handleDashboardKey(event *tcell.EventKey) *tcell.EventKey {
	switch event.Rune() {
	case ' ':
		m.toggleDashboardRowCheck()
		return nil
	case '*':
		m.toggleDashboardSelectAll()
		return nil
	case 'r', 'R':
		if m.command != "" {
			m.RunCommand(m.command)
		}
		return nil
	case 'x', 'X':
		m.showAdHocCommandModal()
		return nil
	}

	if event.Key() == tcell.KeyCtrlA {
		m.toggleDashboardSelectAll()
		return nil
	}
	return event
}

func (m *MultiSessionModal) toggleDashboardRowCheck() {
	row, _ := m.dashboardTable.GetSelection()
	if row < 1 || row > len(m.targetServers) {
		return
	}
	srv := m.targetServers[row-1]
	m.checkedRows[srv.Alias] = !m.checkedRows[srv.Alias]
	m.renderDashboardTable()
	m.updateDashboardHeader()
	m.dashboardTable.Select(row, 0)
}

func (m *MultiSessionModal) toggleDashboardSelectAll() {
	allSelected := true
	for _, s := range m.targetServers {
		if !m.checkedRows[s.Alias] {
			allSelected = false
			break
		}
	}
	for _, s := range m.targetServers {
		m.checkedRows[s.Alias] = !allSelected
	}
	m.renderDashboardTable()
	m.updateDashboardHeader()
}

func (m *MultiSessionModal) handleDashboardEnter() {
	var checkedServers []domain.Server
	for _, s := range m.targetServers {
		if m.checkedRows[s.Alias] {
			checkedServers = append(checkedServers, s)
		}
	}

	switch len(checkedServers) {
	case 0:
		// No checkboxes checked: drill-down into focused row
		row, _ := m.dashboardTable.GetSelection()
		if row >= 1 && row <= len(m.targetServers) {
			target := m.targetServers[row-1]
			m.switchToTab(fmt.Sprintf("tab-session-%s", target.Alias))
		}
	case 1:
		// 1 checked server: drill-down into it
		m.switchToTab(fmt.Sprintf("tab-session-%s", checkedServers[0].Alias))
	case 2, 3, 4:
		// 2 to 4 servers: launch Split View Tab!
		m.setupSplitTab(checkedServers)
	default:
		// > 4 servers
		if m.onStatus != nil {
			m.onStatus("Split view supports up to 4 servers. Please select 2 to 4 servers.", "#FF6B6B")
		}
	}
}

// RunCommand initiates concurrent execution across all target servers.
func (m *MultiSessionModal) RunCommand(command string) {
	m.command = command

	m.resultsMu.Lock()
	for _, s := range m.targetServers {
		m.results[s.Alias] = &ServerExecResult{
			Server:  s,
			Running: true,
		}
		m.outputs[s.Alias].Reset()
	}
	m.resultsMu.Unlock()

	m.updateDashboardHeader()
	m.renderDashboardTable()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)

	var wg sync.WaitGroup
	for _, s := range m.targetServers {
		wg.Add(1)
		go func(srv domain.Server) {
			defer wg.Done()
			m.executeOnServer(srv, command, ctx)
		}(s)
	}

	go func() {
		wg.Wait()
		cancel()
		if m.app != nil {
			m.app.QueueUpdateDraw(func() {
				m.updateDashboardHeader()
				m.renderDashboardTable()
			})
		}
	}()
}

func (m *MultiSessionModal) runOnSingleServer(srv domain.Server, command string) {
	m.resultsMu.Lock()
	if res, ok := m.results[srv.Alias]; ok {
		res.Running = true
	}
	m.resultsMu.Unlock()

	m.updateDashboardHeader()
	m.renderDashboardTable()

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		m.executeOnServer(srv, command, ctx)
		if m.app != nil {
			m.app.QueueUpdateDraw(func() {
				m.updateDashboardHeader()
				m.renderDashboardTable()
			})
		}
	}()
}

func (m *MultiSessionModal) executeOnServer(srv domain.Server, command string, ctx context.Context) {
	start := time.Now()

	type execRes struct {
		out string
		err error
	}
	resChan := make(chan execRes, 1)

	go func() {
		out, err := m.service.ExecuteRemoteCommand(srv.Alias, command)
		resChan <- execRes{out: out, err: err}
	}()

	var out string
	var err error

	select {
	case <-ctx.Done():
		err = ctx.Err()
		out = fmt.Sprintf("execution timed out: %v", err)
	case r := <-resChan:
		out = r.out
		err = r.err
	}

	dur := time.Since(start)

	m.resultsMu.Lock()
	if res, ok := m.results[srv.Alias]; ok {
		res.Running = false
		res.Output = out
		res.Duration = dur
		res.Err = err
	}
	m.resultsMu.Unlock()

	if m.app != nil {
		m.app.QueueUpdateDraw(func() {
			m.updateSessionViewText(srv.Alias)
			for i, splitSrv := range m.splitServers {
				if splitSrv.Alias == srv.Alias && i < len(m.splitPanes) {
					m.updatePaneOutput(m.splitPanes[i], srv.Alias)
				}
			}
			m.updateDashboardHeader()
			m.renderDashboardTable()
		})
	}
}

func (m *MultiSessionModal) updateSessionViewText(alias string) {
	m.resultsMu.RLock()
	res := m.results[alias]
	m.resultsMu.RUnlock()

	if res == nil {
		return
	}

	view, ok := m.sessionViews[alias]
	if !ok || view == nil {
		return
	}

	var sb strings.Builder
	switch {
	case res.Running:
		sb.WriteString("[yellow]⏳ Running...[-]\n\n")
	case res.Err != nil:
		sb.WriteString(fmt.Sprintf("[red]✗ Execution failed (%v)[-]\n\n", res.Err))
	default:
		sb.WriteString(fmt.Sprintf("[green]✓ Succeeded in %.2fs[-]\n\n", res.Duration.Seconds()))
	}
	if res.Output != "" {
		sb.WriteString(tview.Escape(res.Output))
	}
	view.SetText(sb.String())
	view.ScrollToBeginning()
}

// GetResult returns a copy of the execution result for a server alias.
func (m *MultiSessionModal) GetResult(alias string) *ServerExecResult {
	m.resultsMu.RLock()
	defer m.resultsMu.RUnlock()
	res := m.results[alias]
	if res == nil {
		return nil
	}
	cp := *res
	return &cp
}

// RefreshViews forces a refresh of all session and dashboard views on the calling goroutine.
func (m *MultiSessionModal) RefreshViews() {
	m.resultsMu.RLock()
	aliases := make([]string, 0, len(m.results))
	for a := range m.results {
		aliases = append(aliases, a)
	}
	m.resultsMu.RUnlock()

	for _, a := range aliases {
		m.updateSessionViewText(a)
	}
	for i, splitSrv := range m.splitServers {
		if i < len(m.splitPanes) {
			m.updatePaneOutput(m.splitPanes[i], splitSrv.Alias)
		}
	}
	m.updateDashboardHeader()
	m.renderDashboardTable()
}

func (m *MultiSessionModal) showAdHocCommandModal() {
	form := tview.NewForm()
	form.SetBorder(true).
		SetTitle(" 🚀 Run Ad-hoc Command Across Target Servers ").
		SetTitleAlign(tview.AlignCenter)

	cmdInput := tview.NewInputField().
		SetLabel(" Command: ").
		SetFieldWidth(45)
	form.AddFormItem(cmdInput)

	form.AddButton("Run", func() {
		cmd := strings.TrimSpace(cmdInput.GetText())
		if cmd == "" {
			return
		}
		m.app.SetRoot(m, true)
		m.app.SetFocus(m.dashboardTable)
		m.RunCommand(cmd)
	})

	form.AddButton("Cancel", func() {
		m.app.SetRoot(m, true)
		m.app.SetFocus(m.dashboardTable)
	})

	form.SetCancelFunc(func() {
		m.app.SetRoot(m, true)
		m.app.SetFocus(m.dashboardTable)
	})

	m.app.SetRoot(form, true)
	m.app.SetFocus(form)
}

func (m *MultiSessionModal) copyOutputForServer(alias string) {
	m.resultsMu.RLock()
	res := m.results[alias]
	m.resultsMu.RUnlock()

	if res == nil || res.Output == "" {
		return
	}

	if err := clipboard.WriteAll(res.Output); err == nil {
		if m.onStatus != nil {
			m.onStatus(fmt.Sprintf("Copied %s output to clipboard", alias), "#50FA7B")
		}
	} else if m.onStatus != nil {
		m.onStatus("Failed to copy output: "+err.Error(), "#FF6B6B")
	}
}
