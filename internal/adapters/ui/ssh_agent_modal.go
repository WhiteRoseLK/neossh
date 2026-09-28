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

	"github.com/WhiteRoseLK/neossh/internal/core/domain"
	"github.com/WhiteRoseLK/neossh/internal/core/ports"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// SSHAgentModal displays active keys currently loaded inside the system's SSH agent.
type SSHAgentModal struct {
	*tview.Flex
	app        *tview.Application
	service    ports.ServerService
	table      *tview.Table
	headerText *tview.TextView
	helpText   *tview.TextView
	status     domain.SSHAgentStatus

	onClose  func()
	onStatus func(msg string, color string)
}

// NewSSHAgentModal creates an interactive SSH agent inspector dialog.
func NewSSHAgentModal(app *tview.Application, service ports.ServerService, onClose func(), onStatus func(msg string, color string)) *SSHAgentModal {
	m := &SSHAgentModal{
		Flex:       tview.NewFlex().SetDirection(tview.FlexRow),
		app:        app,
		service:    service,
		table:      tview.NewTable().SetSelectable(true, false),
		headerText: tview.NewTextView().SetDynamicColors(true),
		helpText:   tview.NewTextView().SetDynamicColors(true),
		onClose:    onClose,
		onStatus:   onStatus,
	}

	m.setupUI()
	m.Refresh()
	return m
}

func (m *SSHAgentModal) setupUI() {
	m.SetBorder(true).
		SetTitle(" 🔑 SSH Agent Inspector ").
		SetTitleAlign(tview.AlignLeft)

	m.headerText.SetDynamicColors(true)
	m.helpText.SetDynamicColors(true).
		SetText(" [yellow]↑/↓[-]: Navigate  [yellow]r[-]: Refresh  [yellow]Esc/q[-]: Close")

	m.table.SetBorder(true).SetTitle(" Active Keys in Agent ")

	m.table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'r', 'R':
			m.Refresh()
			return nil
		case 'q', 'Q':
			if m.onClose != nil {
				m.onClose()
			}
			return nil
		}
		if event.Key() == tcell.KeyEscape {
			if m.onClose != nil {
				m.onClose()
			}
			return nil
		}
		return event
	})

	m.AddItem(m.headerText, 3, 0, false)
	m.AddItem(m.table, 0, 1, true)
	m.AddItem(m.helpText, 1, 0, false)
}

// Refresh re-queries the agent for active keys.
func (m *SSHAgentModal) Refresh() {
	if m.service == nil {
		return
	}
	m.status = m.service.GetSSHAgentStatus()

	var stateText string
	if m.status.Available {
		stateText = fmt.Sprintf(" Agent: [green::b]Active[-] (%s) • Socket: [gray]%s[-] • Loaded Keys: [yellow::b]%d[-]",
			m.status.Type, m.status.SocketPath, m.status.KeyCount)
	} else {
		errMsg := m.status.Error
		if errMsg == "" {
			errMsg = "Agent unreachable or SSH_AUTH_SOCK not found"
		}
		stateText = fmt.Sprintf(" Agent: [red::b]Inactive[-] • Reason: [gray]%s[-]", errMsg)
	}
	m.headerText.SetText(stateText)

	m.table.Clear()
	headers := []string{"#", "Algorithm", "SHA256 Fingerprint", "Comment / Path"}
	for col, h := range headers {
		cell := tview.NewTableCell("[yellow::b]" + h + "[-::-]").
			SetSelectable(false).
			SetAlign(tview.AlignLeft)
		m.table.SetCell(0, col, cell)
	}

	for i, k := range m.status.Keys {
		row := i + 1
		numCell := tview.NewTableCell(fmt.Sprintf("%d", row)).SetTextColor(tcell.ColorCadetBlue)
		algoCell := tview.NewTableCell(k.Format).SetTextColor(tcell.ColorDarkOrange)
		fpCell := tview.NewTableCell(k.Fingerprint).SetTextColor(tcell.ColorGreen)
		comment := k.Comment
		if comment == "" {
			comment = "[gray](no comment)[-]"
		}
		commentCell := tview.NewTableCell(comment).SetTextColor(tcell.ColorWhite)

		m.table.SetCell(row, 0, numCell)
		m.table.SetCell(row, 1, algoCell)
		m.table.SetCell(row, 2, fpCell)
		m.table.SetCell(row, 3, commentCell)
	}

	if len(m.status.Keys) > 0 {
		m.table.Select(1, 0)
	}
}
