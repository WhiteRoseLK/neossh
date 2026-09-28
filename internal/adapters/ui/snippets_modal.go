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
	"time"

	"github.com/WhiteRoseLK/neossh/internal/core/domain"
	"github.com/WhiteRoseLK/neossh/internal/core/ports"
	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// SnippetsModal is an interactive modal for managing and running command snippets.
type SnippetsModal struct {
	*tview.Flex
	app           *tview.Application
	service       ports.ServerService
	targetServers []domain.Server
	snippets      []domain.Snippet
	filtered      []domain.Snippet

	headerText  *tview.TextView
	filterInput *tview.InputField
	table       *tview.Table
	detailsView *tview.TextView
	helpText    *tview.TextView

	onClose          func()
	onStatus         func(msg, color string)
	onInteractiveRun func(alias, command string)
}

// NewSnippetsModal creates a new command snippets manager and quick runner dialog.
func NewSnippetsModal(
	app *tview.Application,
	service ports.ServerService,
	targets []domain.Server,
	onClose func(),
	onStatus func(msg, color string),
	onInteractiveRun func(alias, command string),
) *SnippetsModal {
	m := &SnippetsModal{
		Flex:             tview.NewFlex().SetDirection(tview.FlexRow),
		app:              app,
		service:          service,
		targetServers:    targets,
		headerText:       tview.NewTextView().SetDynamicColors(true),
		filterInput:      tview.NewInputField(),
		table:            tview.NewTable().SetSelectable(true, false),
		detailsView:      tview.NewTextView().SetDynamicColors(true).SetScrollable(true),
		helpText:         tview.NewTextView().SetDynamicColors(true),
		onClose:          onClose,
		onStatus:         onStatus,
		onInteractiveRun: onInteractiveRun,
	}

	m.setupUI()
	m.LoadSnippets()
	return m
}

func (m *SnippetsModal) setupUI() {
	m.SetBorder(true).
		SetTitle(" ⚡ Command Snippets & Quick Runner ").
		SetTitleAlign(tview.AlignLeft)

	m.renderHeader()

	// Filter Input
	m.filterInput.
		SetLabel(" 🔍 Filter / Search: ").
		SetLabelColor(tcell.ColorYellow).
		SetFieldBackgroundColor(tcell.ColorBlack).
		SetFieldTextColor(tcell.ColorWhite).
		SetChangedFunc(func(text string) {
			m.applyFilter(text)
		})

	m.filterInput.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter || event.Key() == tcell.KeyDown {
			m.app.SetFocus(m.table)
			return nil
		}
		if event.Key() == tcell.KeyEscape {
			if m.filterInput.GetText() != "" {
				m.filterInput.SetText("")
				return nil
			}
			if m.onClose != nil {
				m.onClose()
			}
			return nil
		}
		if event.Key() == tcell.KeyCtrlR {
			m.runAdHocCommand(m.filterInput.GetText())
			return nil
		}
		return event
	})

	// Snippets Table setup
	m.table.SetBorder(true).SetTitle(" Saved Snippets ")
	m.table.SetSelectionChangedFunc(func(row, column int) {
		m.updateDetailsView(row)
	})
	m.table.SetSelectedFunc(func(row, column int) {
		m.handleRunSelected(false)
	})

	m.table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'r', 'R':
			m.handleRunSelected(false)
			return nil
		case 'i', 'I':
			m.handleRunSelected(true)
			return nil
		case 'a', 'A':
			m.showAddSnippetModal()
			return nil
		case 'e', 'E':
			m.handleEditSelected()
			return nil
		case 'd', 'D':
			m.handleDeleteSelected()
			return nil
		case 'c', 'C':
			m.handleCopySelected()
			return nil
		case 'x', 'X':
			m.showAdHocCommandModal()
			return nil
		case '/':
			m.app.SetFocus(m.filterInput)
			return nil
		case 'q', 'Q':
			if m.onClose != nil {
				m.onClose()
			}
			return nil
		}

		if event.Key() == tcell.KeyDelete {
			m.handleDeleteSelected()
			return nil
		}
		if event.Key() == tcell.KeyEscape {
			if m.onClose != nil {
				m.onClose()
			}
			return nil
		}
		if event.Key() == tcell.KeyTab {
			m.app.SetFocus(m.filterInput)
			return nil
		}
		return event
	})

	// Details View
	m.detailsView.SetBorder(true).SetTitle(" Snippet Details ")

	// Split body
	bodyFlex := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(m.table, 0, 3, true).
		AddItem(m.detailsView, 0, 2, false)

	// Help Bar
	m.helpText.SetText(" [yellow]Enter/r[-]: Run  [yellow]i[-]: Interactive  [yellow]a[-]: Add  [yellow]e[-]: Edit  [yellow]d[-]: Delete  [yellow]c[-]: Copy  [yellow]x[-]: Ad-hoc  [yellow]/[-]: Search  [yellow]Esc[-]: Close")

	m.AddItem(m.headerText, 2, 0, false)
	m.AddItem(m.filterInput, 1, 0, false)
	m.AddItem(bodyFlex, 0, 1, true)
	m.AddItem(m.helpText, 1, 0, false)
}

func (m *SnippetsModal) renderHeader() {
	var targetBadge string
	switch len(m.targetServers) {
	case 0:
		targetBadge = "[red::b]None (Select server first to execute)[-::-]"
	case 1:
		srv := m.targetServers[0]
		targetBadge = fmt.Sprintf("[green::b]%s[-::-] (%s)", srv.Alias, srv.Host)
	default:
		aliases := make([]string, 0, len(m.targetServers))
		for _, s := range m.targetServers {
			aliases = append(aliases, s.Alias)
		}
		if len(aliases) > 4 {
			aliases = append(aliases[:4], fmt.Sprintf("+%d more", len(aliases)-4))
		}
		targetBadge = fmt.Sprintf("[yellow::b]%d servers selected[-::-] [%s]", len(m.targetServers), strings.Join(aliases, ", "))
	}

	m.headerText.SetText(fmt.Sprintf(" [yellow::b]Target:[-::-] %s\n [gray]Execute repetitive DevOps/sysadmin tasks, with variable interpolation & multi-server support[-]",
		targetBadge))
}

// LoadSnippets reloads snippets from the underlying repository.
func (m *SnippetsModal) LoadSnippets() {
	if m.service == nil {
		return
	}
	snippets, err := m.service.GetSnippets()
	if err != nil {
		if m.onStatus != nil {
			m.onStatus("Failed to load snippets: "+err.Error(), "#FF6B6B")
		}
		return
	}
	m.snippets = snippets
	m.applyFilter(m.filterInput.GetText())
}

func (m *SnippetsModal) applyFilter(query string) {
	q := strings.ToLower(strings.TrimSpace(query))
	m.filtered = nil

	for _, snip := range m.snippets {
		if q == "" {
			m.filtered = append(m.filtered, snip)
			continue
		}

		match := strings.Contains(strings.ToLower(snip.Name), q) ||
			strings.Contains(strings.ToLower(snip.Description), q) ||
			strings.Contains(strings.ToLower(snip.Command), q)

		if !match {
			for _, tg := range snip.Tags {
				if strings.Contains(strings.ToLower(tg), q) {
					match = true
					break
				}
			}
		}

		if match {
			m.filtered = append(m.filtered, snip)
		}
	}

	m.renderTable()
}

func (m *SnippetsModal) renderTable() {
	m.table.Clear()

	// Table Headers
	headers := []string{"Name", "Tags", "Command"}
	for col, h := range headers {
		cell := tview.NewTableCell(" " + h + " ").
			SetTextColor(tcell.ColorYellow).
			SetSelectable(false)
		m.table.SetCell(0, col, cell)
	}

	for i, snip := range m.filtered {
		row := i + 1

		nameCell := tview.NewTableCell(" " + snip.Name).
			SetTextColor(tcell.ColorWhite).
			SetSelectable(true)

		tagsStr := strings.Join(snip.Tags, ", ")
		if tagsStr == "" {
			tagsStr = "-"
		}
		tagsCell := tview.NewTableCell(" " + tagsStr).
			SetTextColor(tcell.ColorTeal).
			SetSelectable(true)

		cmdPreview := snip.Command
		if len(cmdPreview) > 40 {
			cmdPreview = cmdPreview[:37] + "..."
		}
		cmdCell := tview.NewTableCell(" " + cmdPreview).
			SetTextColor(tcell.ColorGray).
			SetSelectable(true)

		m.table.SetCell(row, 0, nameCell)
		m.table.SetCell(row, 1, tagsCell)
		m.table.SetCell(row, 2, cmdCell)
	}

	if len(m.filtered) > 0 {
		m.table.Select(1, 0)
		m.updateDetailsView(1)
	} else {
		m.detailsView.SetText("[gray]No snippets match the filter.[-]")
	}
}

func (m *SnippetsModal) updateDetailsView(row int) {
	if row < 1 || row > len(m.filtered) {
		m.detailsView.SetText("")
		return
	}

	snip := m.filtered[row-1]
	placeholders := domain.ExtractPlaceholders(snip.Command)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("[yellow::b]Name:[-::-] [white]%s[-]\n\n", tview.Escape(snip.Name)))
	sb.WriteString(fmt.Sprintf("[yellow::b]Description:[-::-] [gray]%s[-]\n\n", tview.Escape(snip.Description)))

	tagsStr := strings.Join(snip.Tags, ", ")
	if tagsStr == "" {
		tagsStr = "none"
	}
	sb.WriteString(fmt.Sprintf("[yellow::b]Tags:[-::-] [teal]%s[-]\n\n", tview.Escape(tagsStr)))

	if len(placeholders) > 0 {
		params := strings.Join(placeholders, ", ")
		sb.WriteString(fmt.Sprintf("[lime::b]Parameters Detected:[-::-] [yellow]{%s}[-]\n\n", tview.Escape(params)))
	}

	sb.WriteString("[yellow::b]Command:[-::-]\n")
	sb.WriteString(fmt.Sprintf("[aqua]%s[-]\n", tview.Escape(snip.Command)))

	m.detailsView.SetText(sb.String())
	m.detailsView.ScrollToBeginning()
}

func (m *SnippetsModal) getSelectedSnippet() (*domain.Snippet, bool) {
	row, _ := m.table.GetSelection()
	if row < 1 || row > len(m.filtered) {
		return nil, false
	}
	return &m.filtered[row-1], true
}

func (m *SnippetsModal) handleRunSelected(interactive bool) {
	snip, ok := m.getSelectedSnippet()
	if !ok {
		return
	}

	if len(m.targetServers) == 0 {
		if m.onStatus != nil {
			m.onStatus("No target server selected. Select a server before running.", "#FF6B6B")
		}
		return
	}

	placeholders := domain.ExtractPlaceholders(snip.Command)
	if len(placeholders) > 0 {
		m.showParameterPromptModal(*snip, interactive)
		return
	}

	m.executeCommand(snip.Command, interactive)
}

func (m *SnippetsModal) executeCommand(command string, interactive bool) {
	if len(m.targetServers) == 0 {
		if m.onStatus != nil {
			m.onStatus("No target server selected.", "#FF6B6B")
		}
		return
	}

	if interactive {
		if len(m.targetServers) > 1 {
			if m.onStatus != nil {
				m.onStatus("Interactive mode is only supported on a single target server.", "#FF6B6B")
			}
			return
		}
		if m.onInteractiveRun != nil {
			m.onInteractiveRun(m.targetServers[0].Alias, command)
		}
		return
	}

	// Multi-session runner modal with Dashboard and Tabs
	runner := NewMultiSessionModal(
		m.app,
		m.service,
		m.targetServers,
		command,
		func() {
			m.app.SetRoot(m, true)
			m.app.SetFocus(m.table)
		},
		m.onStatus,
		m.onInteractiveRun,
	)

	m.app.SetRoot(runner, true)
	m.app.SetFocus(runner)
}

func (m *SnippetsModal) showParameterPromptModal(snippet domain.Snippet, interactive bool) {
	placeholders := domain.ExtractPlaceholders(snippet.Command)
	form := tview.NewForm()
	form.SetBorder(true).
		SetTitle(fmt.Sprintf(" Enter Parameters: %s ", snippet.Name)).
		SetTitleAlign(tview.AlignCenter)

	inputs := make(map[string]*tview.InputField)
	for _, p := range placeholders {
		input := tview.NewInputField().
			SetLabel(fmt.Sprintf(" %s: ", p)).
			SetFieldWidth(35)
		form.AddFormItem(input)
		inputs[p] = input
	}

	form.AddButton("Run", func() {
		params := make(map[string]string)
		for p, inp := range inputs {
			params[p] = strings.TrimSpace(inp.GetText())
		}
		interpolated := domain.InterpolateSnippet(snippet.Command, params)
		m.app.SetRoot(m, true)
		m.app.SetFocus(m.table)
		m.executeCommand(interpolated, interactive)
	})

	form.AddButton("Cancel", func() {
		m.app.SetRoot(m, true)
		m.app.SetFocus(m.table)
	})

	form.SetCancelFunc(func() {
		m.app.SetRoot(m, true)
		m.app.SetFocus(m.table)
	})

	m.app.SetRoot(form, true)
	m.app.SetFocus(form)
}

func (m *SnippetsModal) runAdHocCommand(cmd string) {
	trimmed := strings.TrimSpace(cmd)
	if trimmed == "" {
		m.showAdHocCommandModal()
		return
	}
	m.executeCommand(trimmed, false)
}

func (m *SnippetsModal) showAdHocCommandModal() {
	form := tview.NewForm()
	form.SetBorder(true).
		SetTitle(" 🚀 Run Ad-hoc Command ").
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
		m.app.SetFocus(m.table)
		m.executeCommand(cmd, false)
	})

	form.AddButton("Cancel", func() {
		m.app.SetRoot(m, true)
		m.app.SetFocus(m.table)
	})

	form.SetCancelFunc(func() {
		m.app.SetRoot(m, true)
		m.app.SetFocus(m.table)
	})

	m.app.SetRoot(form, true)
	m.app.SetFocus(form)
}

func (m *SnippetsModal) showAddSnippetModal() {
	m.showSnippetForm(domain.Snippet{}, false)
}

func (m *SnippetsModal) handleEditSelected() {
	snip, ok := m.getSelectedSnippet()
	if !ok {
		return
	}
	m.showSnippetForm(*snip, true)
}

func (m *SnippetsModal) showSnippetForm(initial domain.Snippet, isEdit bool) {
	title := " ➕ Add Snippet "
	if isEdit {
		title = fmt.Sprintf(" ✏️ Edit Snippet: %s ", initial.Name)
	}

	form := tview.NewForm()
	form.SetBorder(true).
		SetTitle(title).
		SetTitleAlign(tview.AlignCenter)

	nameInput := tview.NewInputField().
		SetLabel(" Name: ").
		SetText(initial.Name).
		SetFieldWidth(40)

	cmdInput := tview.NewInputField().
		SetLabel(" Command: ").
		SetText(initial.Command).
		SetFieldWidth(40)

	descInput := tview.NewInputField().
		SetLabel(" Description: ").
		SetText(initial.Description).
		SetFieldWidth(40)

	tagsInput := tview.NewInputField().
		SetLabel(" Tags (comma-separated): ").
		SetText(strings.Join(initial.Tags, ", ")).
		SetFieldWidth(40)

	form.AddFormItem(nameInput)
	form.AddFormItem(cmdInput)
	form.AddFormItem(descInput)
	form.AddFormItem(tagsInput)

	form.AddButton("Save", func() {
		name := strings.TrimSpace(nameInput.GetText())
		cmd := strings.TrimSpace(cmdInput.GetText())
		if name == "" || cmd == "" {
			if m.onStatus != nil {
				m.onStatus("Name and Command are required fields", "#FF6B6B")
			}
			return
		}

		var tags []string
		for _, part := range strings.Split(tagsInput.GetText(), ",") {
			if p := strings.TrimSpace(part); p != "" {
				tags = append(tags, p)
			}
		}

		id := initial.ID
		if id == "" {
			id = fmt.Sprintf("snippet-%d", time.Now().UnixNano())
		}

		snip := domain.Snippet{
			ID:          id,
			Name:        name,
			Command:     cmd,
			Description: strings.TrimSpace(descInput.GetText()),
			Tags:        tags,
		}

		if err := m.service.SaveSnippet(snip); err != nil {
			if m.onStatus != nil {
				m.onStatus("Failed to save snippet: "+err.Error(), "#FF6B6B")
			}
			return
		}

		m.LoadSnippets()
		m.app.SetRoot(m, true)
		m.app.SetFocus(m.table)
		if m.onStatus != nil {
			m.onStatus(fmt.Sprintf("Saved snippet %q", name), "#50FA7B")
		}
	})

	form.AddButton("Cancel", func() {
		m.app.SetRoot(m, true)
		m.app.SetFocus(m.table)
	})

	form.SetCancelFunc(func() {
		m.app.SetRoot(m, true)
		m.app.SetFocus(m.table)
	})

	m.app.SetRoot(form, true)
	m.app.SetFocus(form)
}

func (m *SnippetsModal) handleDeleteSelected() {
	snip, ok := m.getSelectedSnippet()
	if !ok {
		return
	}

	confirmText := fmt.Sprintf("Are you sure you want to delete snippet %q?", snip.Name)
	confirmModal := tview.NewModal().
		SetText(confirmText).
		AddButtons([]string{"[yellow]C[-]ancel", "[red]D[-]elete"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonIndex == 1 {
				if err := m.service.DeleteSnippet(snip.ID); err != nil {
					if m.onStatus != nil {
						m.onStatus("Failed to delete snippet: "+err.Error(), "#FF6B6B")
					}
				} else {
					if m.onStatus != nil {
						m.onStatus(fmt.Sprintf("Deleted snippet %q", snip.Name), "#50FA7B")
					}
					m.LoadSnippets()
				}
			}
			m.app.SetRoot(m, true)
			m.app.SetFocus(m.table)
		})

	confirmModal.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'd' || event.Rune() == 'D' {
			confirmModal.SetFocus(1)
			return nil
		}
		if event.Key() == tcell.KeyEscape || event.Rune() == 'c' || event.Rune() == 'C' {
			m.app.SetRoot(m, true)
			m.app.SetFocus(m.table)
			return nil
		}
		return event
	})

	m.app.SetRoot(confirmModal, true)
	m.app.SetFocus(confirmModal)
}

func (m *SnippetsModal) handleCopySelected() {
	snip, ok := m.getSelectedSnippet()
	if !ok {
		return
	}

	if err := clipboard.WriteAll(snip.Command); err == nil {
		if m.onStatus != nil {
			m.onStatus(fmt.Sprintf("Copied snippet %q command to clipboard", snip.Name), "#50FA7B")
		}
	} else if m.onStatus != nil {
		m.onStatus("Failed to copy command: "+err.Error(), "#FF6B6B")
	}
}
