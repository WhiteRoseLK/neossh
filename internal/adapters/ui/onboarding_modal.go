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
	"time"

	"github.com/WhiteRoseLK/neossh/internal/core/domain"
	"github.com/WhiteRoseLK/neossh/internal/core/ports"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// OnboardingModal is an interactive setup checklist modal for companion tools.
type OnboardingModal struct {
	*tview.Flex
	app              *tview.Application
	companionService ports.CompanionService
	table            *tview.Table
	headerText       *tview.TextView
	infoText         *tview.TextView
	statusText       *tview.TextView
	footerText       *tview.TextView

	tools          []domain.CompanionTool
	packageManager domain.PackageManagerInfo

	isInstalling    bool
	installDone     bool
	installDoneChan chan struct{}
	spinnerIdx      int
	spinnerStop     chan struct{}
	mu              sync.Mutex

	updater  func(func())
	onFinish func()
	onSkip   func()
}

// NewOnboardingModal creates a new OnboardingModal dialog.
func NewOnboardingModal(
	app *tview.Application,
	companionService ports.CompanionService,
	onFinish func(),
	onSkip func(),
) *OnboardingModal {
	m := &OnboardingModal{
		Flex:             tview.NewFlex().SetDirection(tview.FlexRow),
		app:              app,
		companionService: companionService,
		table:            tview.NewTable().SetSelectable(true, false),
		headerText:       tview.NewTextView().SetDynamicColors(true),
		infoText:         tview.NewTextView().SetDynamicColors(true),
		statusText:       tview.NewTextView().SetDynamicColors(true),
		footerText:       tview.NewTextView().SetDynamicColors(true),
		onFinish:         onFinish,
		onSkip:           onSkip,
	}

	m.updater = func(f func()) {
		if m.app != nil {
			m.app.QueueUpdateDraw(f)
		} else {
			f()
		}
	}

	if companionService != nil {
		m.packageManager = companionService.DetectPackageManager()
		m.tools = companionService.CheckCompanionTools()
	}

	m.setupUI()
	return m
}

func (m *OnboardingModal) setupUI() {
	box := tview.NewFlex().SetDirection(tview.FlexRow)
	box.SetBorder(true).
		SetTitle(" 🚀 Welcome to neossh — Companion Tools Setup ").
		SetTitleAlign(tview.AlignCenter).
		SetBorderColor(tcell.ColorDarkCyan)

	// Header
	pmDisplay := "[yellow]None detected[-]"
	if m.packageManager.Type != domain.PackageManagerNone {
		pmDisplay = fmt.Sprintf("[green::b]%s[-]", m.packageManager.DisplayName)
	}

	headerContent := fmt.Sprintf(
		"[white::b]Companion tools unlock advanced native features in neossh:[-]\n"+
			"Package Manager: %s\n",
		pmDisplay,
	)
	m.headerText.SetText(headerContent)
	box.AddItem(m.headerText, 3, 0, false)

	// Table Checklist
	m.table.SetBorder(false)
	m.renderTable()
	m.table.SetSelectionChangedFunc(func(row, column int) {
		m.updateInfoText(row)
	})
	box.AddItem(m.table, len(m.tools)+1, 0, true)

	// Detailed Info Text
	m.infoText.SetBorder(true).
		SetTitle(" Feature Details ").
		SetTitleAlign(tview.AlignLeft).
		SetBorderColor(tcell.ColorDarkSlateGray)
	m.updateInfoText(0)
	box.AddItem(m.infoText, 5, 0, false)

	// Status / Progress Text
	m.statusText.SetText("[gray]Select companion tools to install, then press [Enter].[--]")
	box.AddItem(m.statusText, 2, 0, false)

	// Footer Help
	m.footerText.SetText("[yellow::b][Space][-] Toggle  [green::b][Enter][-] Install / Continue  [white::b][Esc][-] Skip to neossh")
	box.AddItem(m.footerText, 1, 0, false)

	// Center modal on screen
	centered := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(box, 20, 1, true).
			AddItem(nil, 0, 1, false), 76, 1, true).
		AddItem(nil, 0, 1, false)

	m.AddItem(centered, 0, 1, true)
	m.setupKeyBindings()
}

func (m *OnboardingModal) renderTable() {
	m.mu.Lock()
	toolsCopy := make([]domain.CompanionTool, len(m.tools))
	copy(toolsCopy, m.tools)
	m.mu.Unlock()

	m.table.Clear()

	// Header row
	m.table.SetCell(0, 0, tview.NewTableCell(" Status").SetTextColor(tcell.ColorGray).SetSelectable(false))
	m.table.SetCell(0, 1, tview.NewTableCell("Tool").SetTextColor(tcell.ColorGray).SetSelectable(false))
	m.table.SetCell(0, 2, tview.NewTableCell("Unlocked Feature").SetTextColor(tcell.ColorGray).SetSelectable(false))

	for i, tool := range toolsCopy {
		row := i + 1

		var statusCell *tview.TableCell
		switch {
		case tool.Installed:
			statusCell = tview.NewTableCell(" [✓] Installed").SetTextColor(tcell.ColorGreen)
		case tool.Selected:
			statusCell = tview.NewTableCell(" [*] Install").SetTextColor(tcell.ColorYellow)
		default:
			statusCell = tview.NewTableCell(" [ ] Skip").SetTextColor(tcell.ColorDarkGray)
		}

		nameCell := tview.NewTableCell(tool.Name).SetTextColor(tcell.ColorWhite).SetAttributes(tcell.AttrBold)
		descCell := tview.NewTableCell(fmt.Sprintf("%s (%s)", tool.Description, tool.Shortcut)).SetTextColor(tcell.ColorCadetBlue)

		m.table.SetCell(row, 0, statusCell)
		m.table.SetCell(row, 1, nameCell)
		m.table.SetCell(row, 2, descCell)
	}

	if len(toolsCopy) > 0 {
		m.table.Select(1, 0)
	}
}

func (m *OnboardingModal) updateInfoText(tableRow int) {
	idx := tableRow - 1
	m.mu.Lock()
	if idx < 0 || idx >= len(m.tools) {
		m.mu.Unlock()
		m.infoText.SetText("")
		return
	}

	tool := m.tools[idx]
	m.mu.Unlock()

	var details string
	switch tool.ID {
	case "chezmoi":
		details = "[yellow::b]chezmoi[-] (Dotfiles Sync)\n" +
			"Synchronize your personal dotfiles & shell configurations to remote servers in 1 click (Key: [white::b]D[-]).\n" +
			"Uses local streaming tar archives with zero software dependencies required on remote hosts."
	case "yazi":
		details = "[yellow::b]yazi[-] (SFTP Terminal File Manager)\n" +
			"Modern, blazing-fast async terminal file manager written in Rust.\n" +
			"Enables visual dual-pane SFTP browsing, directory navigation, and file transfers (Key: [white::b]F[-])."
	case "ssh-copy-id":
		details = "[yellow::b]ssh-copy-id[-] (1-Click Key Deploy)\n" +
			"Official OpenSSH public key installation utility.\n" +
			"Enables 1-click deployment of your identity keys into remote authorized_keys files."
	default:
		details = fmt.Sprintf("[white::b]%s[-]: %s", tool.Name, tool.Description)
	}

	m.infoText.SetText(details)
}

func (m *OnboardingModal) setupKeyBindings() {
	m.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		m.mu.Lock()
		installing := m.isInstalling
		done := m.installDone
		m.mu.Unlock()

		if installing {
			return nil
		}

		if done {
			if event.Key() == tcell.KeyEnter || event.Key() == tcell.KeyEscape || event.Rune() == 'q' {
				if m.onFinish != nil {
					m.onFinish()
				}
				return nil
			}
			return event
		}

		if event.Key() == tcell.KeyEscape {
			if m.onSkip != nil {
				m.onSkip()
			}
			return nil
		}
		if event.Key() == tcell.KeyEnter {
			m.handleEnter()
			return nil
		}
		if event.Key() == tcell.KeyRune {
			if event.Rune() == ' ' {
				m.toggleCurrentSelection()
				return nil
			}
			if event.Rune() == 'q' {
				if m.onSkip != nil {
					m.onSkip()
				}
				return nil
			}
		}

		return event
	})
}

func (m *OnboardingModal) toggleCurrentSelection() {
	row, _ := m.table.GetSelection()
	idx := row - 1
	m.mu.Lock()
	if idx < 0 || idx >= len(m.tools) {
		m.mu.Unlock()
		return
	}

	if m.tools[idx].Installed {
		name := m.tools[idx].Name
		m.mu.Unlock()
		m.statusText.SetText(fmt.Sprintf("[yellow]%s is already installed on this machine.[-]", name))
		return
	}

	m.tools[idx].Selected = !m.tools[idx].Selected
	m.mu.Unlock()

	m.renderTable()
	m.table.Select(row, 0)
}

func (m *OnboardingModal) handleEnter() {
	var toInstall []string
	m.mu.Lock()
	for _, t := range m.tools {
		if !t.Installed && t.Selected {
			toInstall = append(toInstall, t.ID)
		}
	}
	m.mu.Unlock()

	if len(toInstall) == 0 {
		// Nothing to install, proceed directly
		if m.onFinish != nil {
			m.onFinish()
		}
		return
	}

	if m.packageManager.Type == domain.PackageManagerNone {
		m.statusText.SetText("[yellow]No package manager detected. Please install missing tools manually, or press [Esc] to continue.[-]")
		return
	}

	m.startInstallation(toInstall)
}

func (m *OnboardingModal) startInstallation(toolIDs []string) {
	m.mu.Lock()
	m.isInstalling = true
	m.spinnerStop = make(chan struct{})
	m.installDoneChan = make(chan struct{})
	m.mu.Unlock()

	m.footerText.SetText("[darkgray]Installing tools... please wait[-]")
	m.statusText.SetText(fmt.Sprintf("[yellow]Starting installation of %d tool(s) via %s...[-]", len(toolIDs), m.packageManager.DisplayName))

	// Animated spinner goroutine
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-m.spinnerStop:
				return
			case <-ticker.C:
				m.mu.Lock()
				m.spinnerIdx = (m.spinnerIdx + 1) % len(spinnerFrames)
				curSpinner := spinnerFrames[m.spinnerIdx]
				m.mu.Unlock()

				m.queueUpdate(func() {
					currentText := m.statusText.GetText(false)
					if strings.Contains(currentText, "Installing") {
						for _, f := range spinnerFrames {
							currentText = strings.ReplaceAll(currentText, f, curSpinner)
						}
						m.statusText.SetText(currentText)
					}
				})
			}
		}
	}()

	// Worker goroutine
	go func() {
		err := m.companionService.InstallCompanionTools(toolIDs, func(tool domain.CompanionTool, status string, toolErr error) {
			m.queueUpdate(func() {
				m.mu.Lock()
				curSpinner := spinnerFrames[m.spinnerIdx]
				switch status {
				case "installing":
					m.statusText.SetText(fmt.Sprintf("[yellow]%s Installing %s via %s...[-]", curSpinner, tool.Name, m.packageManager.DisplayName))
				case "installed":
					for i := range m.tools {
						if m.tools[i].ID == tool.ID {
							m.tools[i].Installed = true
							m.tools[i].Selected = false
						}
					}
					m.statusText.SetText(fmt.Sprintf("[green]✓ %s installed successfully![-]", tool.Name))
				case "failed":
					m.statusText.SetText(fmt.Sprintf("[red]✗ %s installation failed: %v[-]", tool.Name, toolErr))
				}
				m.mu.Unlock()
				m.renderTable()
			})
		})

		m.mu.Lock()
		close(m.spinnerStop)
		m.isInstalling = false
		m.installDone = true
		if m.installDoneChan != nil {
			close(m.installDoneChan)
		}
		m.mu.Unlock()

		m.queueUpdate(func() {
			if err != nil {
				m.statusText.SetText(fmt.Sprintf("[red::b]✗ Installation failed: %v[-] Press [Enter] to continue.", err))
			} else {
				m.statusText.SetText("[green::b]✓ All selected tools installed successfully! Press [Enter] to continue to neossh.[-]")
			}
			m.footerText.SetText("[green::b][Enter][-] Continue to neossh")
		})
	}()
}

// SetUpdater overrides the UI update dispatcher (defaults to app.QueueUpdateDraw).
func (m *OnboardingModal) SetUpdater(u func(func())) {
	m.mu.Lock()
	m.updater = u
	m.mu.Unlock()
}

func (m *OnboardingModal) queueUpdate(fn func()) {
	m.mu.Lock()
	u := m.updater
	m.mu.Unlock()
	switch {
	case u != nil:
		u(fn)
	case m.app != nil:
		m.app.QueueUpdateDraw(fn)
	default:
		fn()
	}
}

// WaitInstallDone waits until background installation completes.
func (m *OnboardingModal) WaitInstallDone() {
	m.mu.Lock()
	ch := m.installDoneChan
	m.mu.Unlock()
	if ch != nil {
		<-ch
	}
}
