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
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const (
	SortFilesByName = 0
	SortFilesBySize = 1
	SortFilesByDate = 2
)

// SFTPManager represents the dual-pane SFTP and local file manager.
type SFTPManager struct {
	*tview.Flex

	app         *tview.Application
	leftSource  FileSource
	rightSource FileSource

	header     *tview.TextView
	middle     *tview.Flex
	leftTable  *tview.Table
	rightTable *tview.Table
	statusText *tview.TextView
	hintText   *tview.TextView

	leftEntries  []FileEntry
	rightEntries []FileEntry

	activePane int // 0 = Left, 1 = Right
	sortMode   int // 0 = Name, 1 = Size, 2 = Date

	onClose func()

	transferMu   sync.Mutex
	transferring bool
}

// NewSFTPManager creates a new dual-pane SFTP file manager.
func NewSFTPManager(app *tview.Application, leftSource, rightSource FileSource) *SFTPManager {
	m := &SFTPManager{
		Flex:        tview.NewFlex().SetDirection(tview.FlexRow),
		app:         app,
		leftSource:  leftSource,
		rightSource: rightSource,
		header:      tview.NewTextView(),
		middle:      tview.NewFlex().SetDirection(tview.FlexColumn),
		leftTable:   tview.NewTable(),
		rightTable:  tview.NewTable(),
		statusText:  tview.NewTextView(),
		hintText:    tview.NewTextView(),
		activePane:  0,
		sortMode:    SortFilesByName,
	}

	m.setupUI()
	return m
}

func (m *SFTPManager) setupUI() {
	m.header.
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft).
		SetBorderPadding(0, 0, 1, 1)

	m.statusText.
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft).
		SetText("[green]Ready[-]")
	m.statusText.SetBorderPadding(0, 0, 1, 1)

	m.hintText.
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft).
		SetText(fmt.Sprintf(
			"[%s]Tab[-] Switch Pane  • [%s]↑↓/jk[-] Move  • [%s]Enter[-] Open  • [%s]Backspace[-] Up  • "+
				"[%s]u[-] Upload (L->R)  • [%s]d[-] Download (R->L)  • [%s]s[-] Sort  • [%s]r[-] Refresh  • [%s]q/Esc[-] Close",
			CurrentTheme.HintKey, CurrentTheme.HintKey, CurrentTheme.HintKey, CurrentTheme.HintKey,
			CurrentTheme.HintKey, CurrentTheme.HintKey, CurrentTheme.HintKey, CurrentTheme.HintKey, CurrentTheme.HintKey,
		))
	m.hintText.SetBorderPadding(0, 0, 1, 1)

	m.setupTable(m.leftTable, " Local Files ")
	m.setupTable(m.rightTable, " Remote Files ")

	m.middle.
		AddItem(m.leftTable, 0, 1, true).
		AddItem(m.rightTable, 0, 1, false)

	m.AddItem(m.header, 2, 0, false).
		AddItem(m.middle, 0, 1, true).
		AddItem(m.statusText, 1, 0, false).
		AddItem(m.hintText, 1, 0, false)

	m.SetBorder(true).
		SetTitle(" SFTP Dual-Pane File Manager ").
		SetTitleAlign(tview.AlignCenter)

	m.updateHeader()
	m.refreshPane(0)
	m.refreshPane(1)
	m.updateFocus()
}

func (m *SFTPManager) setupTable(table *tview.Table, title string) {
	table.
		SetSelectable(true, false).
		SetBorders(false).
		SetBorder(true).
		SetTitle(title).
		SetTitleAlign(tview.AlignLeft)

	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		return m.handleKey(event)
	})
}

// OnClose sets the callback when the file manager is closed.
func (m *SFTPManager) OnClose(fn func()) *SFTPManager {
	m.onClose = fn
	return m
}

func (m *SFTPManager) updateHeader() {
	leftTitle := "Left"
	if m.leftSource != nil {
		leftTitle = m.leftSource.Title()
	}
	rightTitle := "Right"
	if m.rightSource != nil {
		rightTitle = m.rightSource.Title()
	}

	sortName := "Name"
	switch m.sortMode {
	case SortFilesBySize:
		sortName = "Size ↓"
	case SortFilesByDate:
		sortName = "Date ↓"
	}

	m.header.SetText(fmt.Sprintf(
		"[dodgerblue::b]◀ %s[-::-]   [gray]|[-]   [yellow::b]▶ %s[-::-]   [gray](Sort: [white]%s[gray])[-]",
		leftTitle, rightTitle, sortName,
	))
}

func (m *SFTPManager) updateFocus() {
	focusedBorderColor := CurrentTheme.BorderColorFocused
	normalBorderColor := CurrentTheme.BorderColorUnfocused

	if m.activePane == 0 {
		m.leftTable.SetBorderColor(focusedBorderColor)
		m.rightTable.SetBorderColor(normalBorderColor)
		m.app.SetFocus(m.leftTable)
	} else {
		m.leftTable.SetBorderColor(normalBorderColor)
		m.rightTable.SetBorderColor(focusedBorderColor)
		m.app.SetFocus(m.rightTable)
	}
}

func (m *SFTPManager) currentSourceAndTable(pane int) (FileSource, *tview.Table, *[]FileEntry) {
	if pane == 0 {
		return m.leftSource, m.leftTable, &m.leftEntries
	}
	return m.rightSource, m.rightTable, &m.rightEntries
}

func (m *SFTPManager) refreshPane(pane int) {
	src, table, entriesPtr := m.currentSourceAndTable(pane)
	if src == nil {
		return
	}

	rawEntries, err := src.List()
	if err != nil {
		m.showStatus(fmt.Sprintf("Failed to list files: %v", err), "#FF6B6B")
		return
	}

	// Sort entries
	m.sortEntries(rawEntries)
	*entriesPtr = rawEntries

	table.Clear()

	// Header row
	headers := []string{"Name", "Size", "Permissions", "Modified"}
	for col, h := range headers {
		cell := tview.NewTableCell(h).
			SetTextColor(tcell.ColorYellow).
			SetSelectable(false).
			SetAttributes(tcell.AttrBold)
		table.SetCell(0, col, cell)
	}

	// Parent directory entry ".."
	row := 1
	dotDotCell := tview.NewTableCell("📁 ..").
		SetTextColor(tcell.ColorDodgerBlue).
		SetAttributes(tcell.AttrBold)
	table.SetCell(row, 0, dotDotCell)
	table.SetCell(row, 1, tview.NewTableCell("<DIR>").SetTextColor(tcell.ColorGray))
	table.SetCell(row, 2, tview.NewTableCell("drwxr-xr-x").SetTextColor(tcell.ColorGray))
	table.SetCell(row, 3, tview.NewTableCell("").SetTextColor(tcell.ColorGray))
	row++

	for _, e := range rawEntries {
		nameIcon := "📄 "
		color := tcell.ColorWhite
		if e.IsDir {
			nameIcon = "📁 "
			color = tcell.ColorDodgerBlue
		}

		sizeStr := FormatFileSize(e.Size)
		if e.IsDir {
			sizeStr = "<DIR>"
		}

		nameCell := tview.NewTableCell(nameIcon + e.Name).SetTextColor(color)
		sizeCell := tview.NewTableCell(sizeStr).SetTextColor(tcell.ColorLightGray)
		permCell := tview.NewTableCell(e.Mode.String()).SetTextColor(tcell.ColorGray)
		dateCell := tview.NewTableCell(e.ModTime.Format("2006-01-02 15:04")).SetTextColor(tcell.ColorGray)

		table.SetCell(row, 0, nameCell)
		table.SetCell(row, 1, sizeCell)
		table.SetCell(row, 2, permCell)
		table.SetCell(row, 3, dateCell)
		row++
	}

	if table.GetRowCount() > 1 {
		table.Select(1, 0)
	}
	m.updateHeader()
}

func (m *SFTPManager) sortEntries(entries []FileEntry) {
	sort.SliceStable(entries, func(i, j int) bool {
		// Directories always first
		if entries[i].IsDir != entries[j].IsDir {
			return entries[i].IsDir
		}
		switch m.sortMode {
		case SortFilesBySize:
			return entries[i].Size > entries[j].Size
		case SortFilesByDate:
			return entries[i].ModTime.After(entries[j].ModTime)
		default: // SortFilesByName
			return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
		}
	})
}

func (m *SFTPManager) handleKey(event *tcell.EventKey) *tcell.EventKey {
	//nolint:exhaustive // Only relevant keys are handled
	switch event.Key() {
	case tcell.KeyTab, tcell.KeyBacktab:
		m.activePane = 1 - m.activePane
		m.updateFocus()
		return nil

	case tcell.KeyEscape:
		m.close()
		return nil

	case tcell.KeyBackspace, tcell.KeyBackspace2:
		m.navigateUp()
		return nil

	case tcell.KeyEnter:
		m.navigateEnter()
		return nil

	case tcell.KeyRune:
		switch event.Rune() {
		case 'q', 'Q':
			m.close()
			return nil
		case 's', 'S':
			m.sortMode = (m.sortMode + 1) % 3
			m.refreshPane(0)
			m.refreshPane(1)
			return nil
		case 'r', 'R':
			m.refreshPane(m.activePane)
			return nil
		case 'u', 'U':
			m.startTransfer(0, 1) // Upload: Left -> Right
			return nil
		case 'd', 'D':
			m.startTransfer(1, 0) // Download: Right -> Left
			return nil
		}
	}
	return event
}

func (m *SFTPManager) navigateUp() {
	src, _, _ := m.currentSourceAndTable(m.activePane)
	if src == nil {
		return
	}
	if err := src.SetPath(".."); err != nil {
		m.showStatus(fmt.Sprintf("Cannot navigate up: %v", err), "#FF6B6B")
		return
	}
	m.refreshPane(m.activePane)
}

func (m *SFTPManager) navigateEnter() {
	src, table, entriesPtr := m.currentSourceAndTable(m.activePane)
	if src == nil || table == nil || entriesPtr == nil {
		return
	}

	selectedRow, _ := table.GetSelection()
	if selectedRow <= 0 {
		return
	}

	// Row 1 is ".."
	if selectedRow == 1 {
		m.navigateUp()
		return
	}

	entryIdx := selectedRow - 2
	entries := *entriesPtr
	if entryIdx < 0 || entryIdx >= len(entries) {
		return
	}

	entry := entries[entryIdx]
	if !entry.IsDir {
		m.showStatus(fmt.Sprintf("File selected: %s (%s)", entry.Name, FormatFileSize(entry.Size)), "#51CF66")
		return
	}

	if err := src.SetPath(entry.Name); err != nil {
		m.showStatus(fmt.Sprintf("Cannot enter directory: %v", err), "#FF6B6B")
		return
	}
	m.refreshPane(m.activePane)
}

func (m *SFTPManager) startTransfer(fromPane, toPane int) {
	m.transferMu.Lock()
	if m.transferring {
		m.transferMu.Unlock()
		m.showStatus("Transfer already in progress…", "#FFD43B")
		return
	}
	m.transferring = true
	m.transferMu.Unlock()

	src, srcTable, srcEntriesPtr := m.currentSourceAndTable(fromPane)
	dst, _, _ := m.currentSourceAndTable(toPane)
	if src == nil || dst == nil || srcTable == nil || srcEntriesPtr == nil {
		m.setTransferDone()
		return
	}

	selectedRow, _ := srcTable.GetSelection()
	if selectedRow <= 1 {
		m.setTransferDone()
		m.showStatus("Please select a file to transfer", "#FFD43B")
		return
	}

	entryIdx := selectedRow - 2
	entries := *srcEntriesPtr
	if entryIdx < 0 || entryIdx >= len(entries) {
		m.setTransferDone()
		return
	}

	entry := entries[entryIdx]
	go m.executeTransfer(src, dst, entry, false, toPane)
}

func (m *SFTPManager) executeTransfer(src, dst FileSource, entry FileEntry, overwrite bool, toPane int) {
	direction := "Uploading"
	if toPane == 0 {
		direction = "Downloading"
	}

	startTime := time.Now()
	err := TransferFile(src, dst, entry, overwrite, func(copied, total int64) {
		pct := 0
		if total > 0 {
			pct = int((copied * 100) / total)
		}
		speed := ""
		elapsed := time.Since(startTime).Seconds()
		if elapsed > 0.1 && copied > 0 {
			bps := float64(copied) / elapsed
			speed = fmt.Sprintf(" @ %s/s", FormatFileSize(int64(bps)))
		}

		m.app.QueueUpdateDraw(func() {
			m.statusText.SetText(fmt.Sprintf(
				"[yellow]%s %s... (%d%% - %s / %s%s)[-]",
				direction, entry.Name, pct, FormatFileSize(copied), FormatFileSize(total), speed,
			))
		})
	})

	m.setTransferDone()

	if err != nil {
		if errors.Is(err, ErrFileExists) {
			m.app.QueueUpdateDraw(func() {
				m.showOverwriteModal(src, dst, entry, toPane)
			})
			return
		}
		m.app.QueueUpdateDraw(func() {
			m.showStatus(fmt.Sprintf("Transfer failed: %v", err), "#FF6B6B")
		})
		return
	}

	m.app.QueueUpdateDraw(func() {
		m.refreshPane(toPane)
		m.showStatus(fmt.Sprintf("✓ Successfully transferred %s (%s)", entry.Name, FormatFileSize(entry.Size)), "#51CF66")
	})
}

func (m *SFTPManager) showOverwriteModal(src, dst FileSource, entry FileEntry, toPane int) {
	modal := tview.NewModal().
		SetText(fmt.Sprintf("File %q already exists at destination.\nDo you want to overwrite it?", entry.Name)).
		AddButtons([]string{"Overwrite", "Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			m.app.SetRoot(m, true)
			m.updateFocus()
			if buttonLabel == "Overwrite" {
				m.transferMu.Lock()
				m.transferring = true
				m.transferMu.Unlock()
				go m.executeTransfer(src, dst, entry, true, toPane)
			} else {
				m.showStatus("Transfer canceled", "#FFD43B")
			}
		})

	m.app.SetRoot(modal, true)
	m.app.SetFocus(modal)
}

func (m *SFTPManager) setTransferDone() {
	m.transferMu.Lock()
	m.transferring = false
	m.transferMu.Unlock()
}

func (m *SFTPManager) showStatus(msg string, color string) {
	m.app.QueueUpdateDraw(func() {
		m.statusText.SetText(fmt.Sprintf("[%s]%s[-]", color, msg))
	})
}

func (m *SFTPManager) close() {
	if m.leftSource != nil {
		_ = m.leftSource.Close()
	}
	if m.rightSource != nil {
		_ = m.rightSource.Close()
	}
	if m.onClose != nil {
		m.onClose()
	}
}
