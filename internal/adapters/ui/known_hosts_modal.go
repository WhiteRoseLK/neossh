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
	"path/filepath"
	"strings"

	"github.com/WhiteRoseLK/neossh/internal/core/domain"
	"github.com/WhiteRoseLK/neossh/internal/core/ports"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// KnownHostsModal is a modal dialog for viewing, searching, and managing ~/.ssh/known_hosts.
type KnownHostsModal struct {
	*tview.Flex
	app         *tview.Application
	service     ports.ServerService
	table       *tview.Table
	filterInput *tview.InputField
	infoText    *tview.TextView
	records     []domain.KnownHostRecord
	filtered    []domain.KnownHostRecord

	onClose  func()
	onStatus func(msg string, color string)
}

// NewKnownHostsModal creates a new KnownHostsModal.
func NewKnownHostsModal(app *tview.Application, service ports.ServerService, onClose func(), onStatus func(msg string, color string)) *KnownHostsModal {
	m := &KnownHostsModal{
		Flex:        tview.NewFlex().SetDirection(tview.FlexRow),
		app:         app,
		service:     service,
		table:       tview.NewTable().SetSelectable(true, false),
		filterInput: tview.NewInputField(),
		infoText:    tview.NewTextView().SetDynamicColors(true),
		onClose:     onClose,
		onStatus:    onStatus,
	}

	m.setupUI()
	m.LoadRecords()
	return m
}

func (m *KnownHostsModal) setupUI() {
	m.SetBorder(true).
		SetTitle(" 🔑 Known Hosts Manager (~/.ssh/known_hosts) ").
		SetTitleAlign(tview.AlignLeft)

	// Filter Input
	m.filterInput.
		SetLabel(" 🔍 Filter: ").
		SetLabelColor(tcell.ColorYellow).
		SetFieldBackgroundColor(tcell.ColorBlack).
		SetFieldTextColor(tcell.ColorWhite).
		SetChangedFunc(func(text string) {
			m.applyFilter(text)
		})

	// Info / Help Bar
	m.infoText.SetText(" [yellow]↑/↓[-]: Navigate  [yellow]d[-]: Delete entry  [yellow]r[-]: Refresh  [yellow]/[-]: Search  [yellow]Tab[-]: Switch focus  [yellow]Esc[-]: Close")

	// Table setup
	m.table.SetBorder(true).SetTitle(" Host Key Records ")
	m.table.SetSelectedFunc(func(row, column int) {
		m.handleDeletePrompt()
	})

	m.table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'd', 'D':
			m.handleDeletePrompt()
			return nil
		case 'r', 'R':
			m.LoadRecords()
			return nil
		case '/':
			m.app.SetFocus(m.filterInput)
			return nil
		}
		if event.Key() == tcell.KeyDelete {
			m.handleDeletePrompt()
			return nil
		}
		return event
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
		return event
	})

	m.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			if m.app.GetFocus() == m.filterInput {
				m.app.SetFocus(m.table)
			} else {
				m.app.SetFocus(m.filterInput)
			}
			return nil
		}
		if event.Key() == tcell.KeyEscape && m.app.GetFocus() != m.filterInput {
			if m.onClose != nil {
				m.onClose()
			}
			return nil
		}
		return event
	})

	m.AddItem(m.filterInput, 1, 0, false)
	m.AddItem(m.table, 0, 1, true)
	m.AddItem(m.infoText, 1, 0, false)
}

// LoadRecords reloads entries from known_hosts.
func (m *KnownHostsModal) LoadRecords() {
	records, err := m.service.ListKnownHostRecords("")
	if err != nil {
		if m.onStatus != nil {
			m.onStatus("Failed to read known_hosts: "+err.Error(), "#FF6B6B")
		}
		return
	}
	m.records = records
	m.applyFilter(m.filterInput.GetText())
}

func (m *KnownHostsModal) applyFilter(query string) {
	q := strings.ToLower(strings.TrimSpace(query))
	m.filtered = nil

	for _, rec := range m.records {
		if q == "" {
			m.filtered = append(m.filtered, rec)
			continue
		}
		if strings.Contains(strings.ToLower(rec.HostPattern), q) ||
			strings.Contains(strings.ToLower(rec.KeyType), q) ||
			strings.Contains(strings.ToLower(rec.Fingerprint), q) ||
			strings.Contains(strings.ToLower(rec.Comment), q) {
			m.filtered = append(m.filtered, rec)
		}
	}

	m.renderTable()
}

func (m *KnownHostsModal) renderTable() {
	m.table.Clear()

	// Header
	headers := []string{"Line", "Host / Pattern", "Key Type", "SHA256 Fingerprint", "Details / Comment"}
	for col, h := range headers {
		cell := tview.NewTableCell("[yellow::b]" + h + "[-::-]").
			SetSelectable(false).
			SetAlign(tview.AlignLeft)
		m.table.SetCell(0, col, cell)
	}

	for i, rec := range m.filtered {
		row := i + 1
		lineCell := tview.NewTableCell(fmt.Sprintf("%d", rec.LineNumber)).SetTextColor(tcell.ColorCadetBlue)

		patternText := rec.HostPattern
		if rec.IsHashed {
			patternText = "[gray]|1| (Hashed)[-]"
		}
		hostCell := tview.NewTableCell(patternText).SetTextColor(tcell.ColorWhite)
		typeCell := tview.NewTableCell(rec.KeyType).SetTextColor(tcell.ColorDarkOrange)

		fp := rec.Fingerprint
		if fp == "" {
			fp = "[gray](unavailable)[-]"
		}
		fpCell := tview.NewTableCell(fp).SetTextColor(tcell.ColorGreen)

		comment := rec.Comment
		if rec.IsHashed && comment == "" {
			comment = "[gray]Hashed hostname[-]"
		}
		commentCell := tview.NewTableCell(comment).SetTextColor(tcell.ColorLightSlateGray)

		m.table.SetCell(row, 0, lineCell)
		m.table.SetCell(row, 1, hostCell)
		m.table.SetCell(row, 2, typeCell)
		m.table.SetCell(row, 3, fpCell)
		m.table.SetCell(row, 4, commentCell)
	}

	if len(m.filtered) > 0 {
		m.table.Select(1, 0)
	}
}

func (m *KnownHostsModal) handleDeletePrompt() {
	row, _ := m.table.GetSelection()
	if row < 1 || row > len(m.filtered) {
		return
	}
	selected := m.filtered[row-1]

	displayHost := selected.HostPattern
	if selected.IsHashed {
		displayHost = "Hashed Host Key"
	}

	confirmText := fmt.Sprintf("Remove host key at line %d (%s)?\n\nA backup copy (~/.ssh/known_hosts.old) will be created automatically.",
		selected.LineNumber, displayHost)

	confirmModal := tview.NewModal().
		SetText(confirmText).
		AddButtons([]string{"[yellow]C[-]ancel", "[red]D[-]elete"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonIndex == 1 {
				bak, err := m.service.RemoveKnownHostByLine("", selected.LineNumber)
				if err != nil {
					if m.onStatus != nil {
						m.onStatus("Failed to delete entry: "+err.Error(), "#FF6B6B")
					}
				} else {
					if m.onStatus != nil {
						m.onStatus(fmt.Sprintf("Removed entry from line %d (backup: %s)", selected.LineNumber, filepath.Base(bak)), "#50FA7B")
					}
					m.LoadRecords()
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
