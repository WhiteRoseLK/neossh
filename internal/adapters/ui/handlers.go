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
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/WhiteRoseLK/neossh/internal/core/domain"
	"github.com/WhiteRoseLK/neossh/internal/core/services"
	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// =============================================================================
// Event Handlers (handle user input/events)
// =============================================================================
const (
	ForwardTypeLocal   = "Local"
	ForwardTypeRemote  = "Remote"
	ForwardTypeDynamic = "Dynamic"

	ForwardModeOnlyForward = "Only forward"
	ForwardModeForwardSSH  = "Forward + SSH"

	ReadonlyMessage = "Readonly mode: SSH configuration modifications are disabled."
)

// commandKey normalizes runes only in command contexts. Text inputs, search,
// and dropdown filtering intentionally consume the original event unchanged.
func commandKey(event *tcell.EventKey) rune {
	return normalizeGlobalHotkey(event.Rune())
}

// hotkeyMap maps each rune to its normalized command rune.
// Keys that must be case-sensitive have separate entries; keys that fold
// to lowercase share one target. This replaces a large switch so that
// the cyclomatic complexity stays within the gocyclo budget.
var hotkeyMap = map[rune]rune{
	'q': 'q', 'Q': 'q',
	'/': '/',
	'a': 'a', 'A': 'a',
	'e': 'e', 'E': 'e',
	'd': 'd', 'D': 'd',
	'p': 'p',
	'P': 'P',
	'l': 'l', 'L': 'l',
	'u': 'u', 'U': 'u',
	's': 's',
	'S': 'S',
	'c': 'c',
	'C': 'C',
	'o': 'o', 'O': 'o',
	'v': 'v', 'V': 'v',
	'y': 'y', 'Y': 'y',
	'h': 'h',
	'H': 'H',
	'm': 'm',
	'M': 'M',
	'g': 'g',
	'G': 'G',
	'r': 'r', 'R': 'r',
	't': 't',
	'T': 'T',
	'f': 'f', 'F': 'f',
	'x': 'x', 'X': 'x',
	'j': 'j', 'J': 'j',
	'k': 'k',
	'K': 'K',
}

// normalizeGlobalHotkey preserves command semantics for ASCII keys while
// keeping all other runes untouched.
func normalizeGlobalHotkey(key rune) rune {
	if mapped, ok := hotkeyMap[key]; ok {
		return mapped
	}
	return key
}

func (t *tui) handleNavigationKey(key tcell.Key) bool {
	//nolint:exhaustive // We only handle navigation keys and pass through others
	switch key {
	case tcell.KeyTab:
		t.handleNextPanel()
		return true
	case tcell.KeyBacktab:
		t.handlePrevPanel()
		return true
	case tcell.KeyEnter:
		t.handleServerConnect()
		return true
	default:
		return false
	}
}

func (t *tui) handleFocusKeys(cmd rune) bool {
	switch cmd {
	case '0', '/':
		t.handleSearchFocus()
		return true
	case '1':
		t.handleServerListFocus()
		return true
	case '2':
		t.handleActiveListFocus()
		return true
	case '3':
		t.handleDetailsFocus()
		return true
	default:
		return false
	}
}

func (t *tui) handleGlobalKeys(event *tcell.EventKey) *tcell.EventKey {
	// Don't handle global keys when search has focus
	if t.app.GetFocus() == t.searchBar {
		return event
	}

	if t.handleNavigationKey(event.Key()) {
		return nil
	}

	if event.Key() == tcell.KeyCtrlG {
		t.handleGitSSHSetup()
		return nil
	}

	if event.Key() == tcell.KeyCtrlP || event.Key() == tcell.KeyCtrlW {
		t.handleToggleAutoPing()
		return nil
	}

	cmd := commandKey(event)
	if t.readonly {
		switch cmd {
		case 'a', 'e', 'd', 'C', 'y', 'p', 'v', 'K', 't', 'i', 'I', 'm', 'l', 'u':
			t.showReadonlyModal()
			return nil
		}
	}

	if t.handleFocusKeys(cmd) {
		return nil
	}

	if t.handleActionKeys(cmd) {
		return nil
	}

	return event
}

func (t *tui) handleClipboardKeys(cmd rune) bool {
	switch cmd {
	case 'c':
		t.handleCopyCommand()
		return true
	case 'h':
		t.handleCopyHost()
		return true
	case 'o':
		t.handleSCPCommandGenerator()
		return true
	case 'M':
		t.handleSSHFSCommandGenerator()
		return true
	case 'v':
		t.handlePasteCommand()
		return true
	case 'y':
		t.handleServerClone()
		return true
	default:
		return false
	}
}

func (t *tui) handleActionKeys(cmd rune) bool {
	if t.handleClipboardKeys(cmd) {
		return true
	}

	switch cmd {
	case 'q':
		t.handleQuit()
		return true
	case 'a':
		t.handleServerAdd()
		return true
	case 'e':
		t.handleServerEdit()
		return true
	case 'd':
		t.handleServerDelete()
		return true
	case 'p':
		t.handleServerPin()
		return true
	case 'P':
		t.handleGitSSHSetup()
		return true
	case 'l':
		t.handleLoadServerKeyToAgent()
		return true
	case 'u':
		t.handleUnloadServerKeyFromAgent()
		return true
	case 'm':
		t.handleToggleServerHidden()
		return true
	case 'H':
		t.handleToggleShowHidden()
		return true
	case 's':
		t.handleSortToggle()
		return true
	case 'S':
		t.handleSortReverse()
		return true
	case 'C':
		t.handleEditServerKeyComment()
		return true
	case 'g':
		t.handlePingSelected()
		return true
	case 'G':
		t.handlePingAll()
		return true
	case 'W', 'w':
		t.handleToggleAutoPing()
		return true
	case 'r':
		t.handleRefreshBackground()
		return true
	case 't':
		t.handleTagsEdit()
		return true
	case 'f':
		t.handlePortForward()
		return true
	case 'x':
		t.handleStopForwarding()
		return true
	case 'j':
		t.handleNavigateDown()
		return true
	case 'k':
		t.handleNavigateUp()
		return true
	case 'K':
		if t.isActiveListFocused() {
			t.handleKillActiveSessions()
		} else {
			t.handleInstallSSHKey()
		}
		return true
	case 'i', 'I':
		t.handleImportKnownHosts()
		return true
	case 'T':
		t.handleThemeToggle()
		return true
	default:
		return false
	}
}

func (t *tui) handleQuit() {
	t.stopAutoPing()
	t.app.Stop()
}

func (t *tui) handleThemeToggle() {
	var newTheme string
	switch CurrentThemeMode {
	case ThemeDark:
		newTheme = ThemeLight
	case ThemeLight:
		newTheme = ThemeSystem
	default:
		newTheme = ThemeDark
	}

	if t.settings != nil {
		_ = t.settings.SaveTheme(newTheme)
	}
	if t.serverService != nil {
		_ = t.serverService.SaveTheme(newTheme)
	}

	if t.themeWatcher != nil {
		if newTheme == ThemeSystem {
			t.themeWatcher.Start()
		} else {
			t.themeWatcher.Stop()
		}
	}

	SetTheme(newTheme)
	ApplyTheme()
	t.rebuildUI()

	modeLabel := newTheme
	if newTheme == ThemeSystem {
		modeLabel = fmt.Sprintf("system (%s)", CurrentTheme.Name)
	}
	t.showStatusTemp("Theme: " + modeLabel)
}

func (t *tui) handleServerPin() {
	if server, ok := t.serverList.GetSelectedServer(); ok {
		pinned := server.PinnedAt.IsZero()
		_ = t.serverService.SetPinned(server.Alias, pinned)
		t.refreshServerList()
	}
}

func (t *tui) handleToggleServerHidden() {
	if server, ok := t.serverList.GetSelectedServer(); ok {
		newHidden := !server.Hidden
		if err := t.serverService.SetHidden(server.Alias, newHidden); err != nil {
			t.showStatusTempColor("Failed to toggle hidden: "+err.Error(), "#FF6B6B")
			return
		}
		if newHidden {
			t.showStatusTemp("Hidden: " + server.Alias)
		} else {
			t.showStatusTemp("Unhidden: " + server.Alias)
		}
		t.refreshServerList()
		t.updateListTitle()
	}
}

func (t *tui) handleToggleShowHidden() {
	t.showHidden = !t.showHidden
	if t.showHidden {
		t.showStatusTemp("Showing hidden servers")
	} else {
		t.showStatusTemp("Hiding hidden servers")
	}
	t.refreshServerList()
	t.updateListTitle()
}

func (t *tui) handleSortToggle() {
	t.sortMode = t.sortMode.ToggleField()
	t.showStatusTemp("Sort: " + t.sortMode.String())
	t.updateListTitle()
	t.persistSortMode()
	t.refreshServerList()
}

func (t *tui) handleSortReverse() {
	t.sortMode = t.sortMode.Reverse()
	t.showStatusTemp("Sort: " + t.sortMode.String())
	t.updateListTitle()
	t.persistSortMode()
	t.refreshServerList()
}

func (t *tui) handleCopyCommand() {
	if server, ok := t.serverList.GetSelectedServer(); ok {
		cmd := BuildSSHCommand(server)
		if err := clipboard.WriteAll(cmd); err == nil {
			t.showStatusTemp("Copied: " + cmd)
		} else {
			t.showStatusTemp("Failed to copy to clipboard")
		}
	}
}

func (t *tui) handleCopyHost() {
	if server, ok := t.serverList.GetSelectedServer(); ok {
		host := server.Host
		if err := clipboard.WriteAll(host); err == nil {
			t.showStatusTemp("Copied: " + host)
		} else {
			t.showStatusTemp("Failed to copy to clipboard")
		}
	}
}

func (t *tui) handleSCPCommandGenerator() {
	server, ok := t.serverList.GetSelectedServer()
	if !ok {
		t.showStatusTemp("No server selected")
		return
	}

	modal := NewSCPModal(t.app, server).
		OnCopied(func(cmd string) {
			t.app.SetRoot(t.root, true)
			t.app.SetFocus(t.serverList)
			t.showStatusTemp("Copied SCP: " + cmd)
		}).
		OnCancel(func() {
			t.app.SetRoot(t.root, true)
			t.app.SetFocus(t.serverList)
		})

	_ = modal.Show()
}

func (t *tui) handleSSHFSCommandGenerator() {
	server, ok := t.serverList.GetSelectedServer()
	if !ok {
		t.showStatusTemp("No server selected")
		return
	}

	modal := NewSSHFSModal(t.app, server).
		OnCopied(func(cmd string) {
			t.app.SetRoot(t.root, true)
			t.app.SetFocus(t.serverList)
			t.showStatusTemp("Copied SSHFS: " + cmd)
		}).
		OnCancel(func() {
			t.app.SetRoot(t.root, true)
			t.app.SetFocus(t.serverList)
		})

	_ = modal.Show()
}

func (t *tui) handlePasteCommand() {
	if t.readonly {
		t.showReadonlyModal()
		return
	}

	// Read from clipboard
	clipContent, err := clipboard.ReadAll()
	if err != nil {
		t.showStatusTemp("Failed to read from clipboard")
		return
	}

	// Try to parse as SSH command
	server, err := ParseSSHCommand(clipContent)
	if err != nil {
		t.showStatusTemp("Invalid SSH command in clipboard: " + err.Error())
		return
	}

	// Check for duplicate alias and auto-adjust if necessary
	existingAliases := t.getExistingAliases()
	server.Alias = GenerateUniqueAlias(server.Alias, existingAliases)

	// Show the server form with parsed data
	// Note: For Add mode, original should be nil. We'll set initial data separately.
	form := NewServerForm(ServerFormAdd, nil).
		SetInitialData(server).
		SetDefaultIdentityKey(t.getDefaultIdentityKey()).
		SetApp(t.app).
		SetVersionInfo(t.version, t.commit).
		OnSave(t.handleServerSave).
		OnCancel(t.handleFormCancel).
		SetExistingAliases(existingAliases)
	t.app.SetRoot(form, true)
}

// getExistingAliases returns all existing server aliases
func (t *tui) getExistingAliases() []string {
	servers, err := t.serverService.ListServers("")
	if err != nil {
		return []string{}
	}

	aliases := make([]string, 0)
	for _, s := range servers {
		if len(s.Aliases) > 0 {
			aliases = append(aliases, s.Aliases...)
		} else if s.Alias != "" {
			aliases = append(aliases, s.Alias)
		}
	}
	return aliases
}

// getExistingAliasesExcept returns all existing server aliases except those belonging to exclude
func (t *tui) getExistingAliasesExcept(exclude domain.Server) []string {
	servers, err := t.serverService.ListServers("")
	if err != nil {
		return []string{}
	}

	aliases := make([]string, 0)
	for _, s := range servers {
		if s.Alias == exclude.Alias || slices.Contains(s.Aliases, exclude.Alias) {
			continue
		}
		if len(s.Aliases) > 0 {
			aliases = append(aliases, s.Aliases...)
		} else if s.Alias != "" {
			aliases = append(aliases, s.Alias)
		}
	}
	return aliases
}

func (t *tui) handleTagsEdit() {
	if t.readonly {
		t.showReadonlyModal()
		return
	}
	if server, ok := t.serverList.GetSelectedServer(); ok {
		t.showEditTagsForm(server)
	}
}

func (t *tui) targetNavigationList() *ServerList {
	if t.isActiveListFocused() {
		return t.activeList
	}
	if t.isServerListFocused() {
		return t.serverList
	}
	return nil
}

func (t *tui) handleNavigateDown() {
	list := t.targetNavigationList()
	if list == nil {
		return
	}
	currentIdx := list.GetCurrentItem()
	itemCount := list.GetItemCount()
	if itemCount == 0 {
		return
	}
	if currentIdx < itemCount-1 {
		list.SetCurrentItem(currentIdx + 1)
	} else {
		list.SetCurrentItem(0)
	}
}

func (t *tui) handleNavigateUp() {
	list := t.targetNavigationList()
	if list == nil {
		return
	}
	currentIdx := list.GetCurrentItem()
	itemCount := list.GetItemCount()
	if itemCount == 0 {
		return
	}
	if currentIdx > 0 {
		list.SetCurrentItem(currentIdx - 1)
	} else {
		list.SetCurrentItem(itemCount - 1)
	}
}

func (t *tui) handleSearchInput(query string) {
	filtered, _ := t.serverService.ListServers(query)
	if strings.TrimSpace(query) == "" {
		sortServersForUI(filtered, t.sortMode)
	}
	displayServers := t.filterServersForDisplay(filtered)
	t.serverList.UpdateServers(displayServers)
	if t.activeList != nil {
		active, _ := t.serverService.ListActiveSessions(query)
		t.activeList.UpdateServers(active)
	}
	if len(displayServers) == 0 {
		t.details.ShowEmpty()
	}
}

func (t *tui) isServerListFocused() bool {
	if t == nil || t.app == nil || t.serverList == nil {
		return false
	}
	focus := t.app.GetFocus()
	return focus == t.serverList || focus == t.serverList.List
}

func (t *tui) isActiveListFocused() bool {
	if t == nil || t.app == nil || t.activeList == nil {
		return false
	}
	focus := t.app.GetFocus()
	return focus == t.activeList || focus == t.activeList.List
}

func (t *tui) handleSearchFocus() {
	if t.app != nil && t.searchBar != nil {
		t.app.SetFocus(t.searchBar)
		t.updateFocusBorders()
	}
}

func (t *tui) handleServerListFocus() {
	if t.app != nil && t.serverList != nil {
		t.app.SetFocus(t.serverList)
		t.updateFocusBorders()
	}
}

func (t *tui) handleActiveListFocus() {
	if t.app != nil && t.activeList != nil {
		t.app.SetFocus(t.activeList)
		t.updateFocusBorders()
	}
}

func (t *tui) handleDetailsFocus() {
	if t.app != nil && t.details != nil {
		t.app.SetFocus(t.details)
		t.updateFocusBorders()
	}
}

func (t *tui) handleNextPanel() {
	focus := t.app.GetFocus()
	switch {
	case t.searchBar != nil && (focus == t.searchBar || t.searchBar.HasFocus()):
		t.handleServerListFocus()
	case t.serverList != nil && (focus == t.serverList || focus == t.serverList.List || t.serverList.HasFocus()):
		t.handleActiveListFocus()
	case t.activeList != nil && (focus == t.activeList || focus == t.activeList.List || t.activeList.HasFocus()):
		t.handleDetailsFocus()
	default:
		t.handleSearchFocus()
	}
}

func (t *tui) handlePrevPanel() {
	focus := t.app.GetFocus()
	switch {
	case t.details != nil && (focus == t.details || focus == t.details.TextView || t.details.HasFocus()):
		t.handleActiveListFocus()
	case t.activeList != nil && (focus == t.activeList || focus == t.activeList.List || t.activeList.HasFocus()):
		t.handleServerListFocus()
	case t.serverList != nil && (focus == t.serverList || focus == t.serverList.List || t.serverList.HasFocus()):
		t.handleSearchFocus()
	default:
		t.handleDetailsFocus()
	}
}

func (t *tui) handleSearchNavigate(direction int) {
	if t.serverList != nil {
		t.app.SetFocus(t.serverList)

		currentIdx := t.serverList.GetCurrentItem()
		itemCount := t.serverList.GetItemCount()

		if itemCount == 0 {
			return
		}

		if direction > 0 {
			if currentIdx < itemCount-1 {
				t.serverList.SetCurrentItem(currentIdx + 1)
			} else {
				t.serverList.SetCurrentItem(0)
			}
		} else {
			if currentIdx > 0 {
				t.serverList.SetCurrentItem(currentIdx - 1)
			} else {
				t.serverList.SetCurrentItem(itemCount - 1)
			}
		}

		if server, ok := t.serverList.GetSelectedServer(); ok {
			t.details.UpdateServer(server)
		}
	}
}

func (t *tui) handleReturnToSearch() {
	if t.searchBar != nil {
		t.app.SetFocus(t.searchBar)
	}
}

func (t *tui) handleServerConnect() {
	var server domain.Server
	var ok bool
	if t.isActiveListFocused() {
		server, ok = t.activeList.GetSelectedServer()
	} else if t.serverList != nil {
		server, ok = t.serverList.GetSelectedServer()
	}
	if !ok {
		return
	}
	if server.IsWildcardServer() {
		t.showErrorModal("SSH Connection Warning", "Cannot initiate direct SSH connection to a wildcard pattern block")
		return
	}
	var sshErr error
	t.app.Suspend(func() {
		if err := t.serverService.SSH(server.Alias); err != nil {
			sshErr = err
			t.logger.Errorw("ssh session error", "alias", server.Alias, "error", err)
		}
	})
	if t.exitOnDisconnect {
		t.app.Stop()
		return
	}
	services.SetTerminalTitle("neossh")
	t.app.Sync()
	t.refreshServerList()
	if sshErr != nil {
		t.showSSHErrorModal(server.Alias, sshErr.Error())
	}
}

func (t *tui) handleInstallSSHKey() {
	if t.readonly {
		t.showReadonlyModal()
		return
	}
	if server, ok := t.serverList.GetSelectedServer(); ok {
		if server.IsWildcardServer() {
			t.showErrorModal("SSH Key Installation Warning", "Cannot install SSH key to a wildcard pattern block")
			return
		}
		alias := server.Alias
		t.showStatusTemp(fmt.Sprintf("Installing key to %s…", alias))
		var copyErr error
		t.app.Suspend(func() {
			if err := t.serverService.CopySSHKey(alias); err != nil {
				copyErr = err
				t.logger.Errorw("failed to install ssh key", "alias", alias, "error", err)
			}
		})
		t.app.Sync()
		t.refreshServerList()
		if copyErr != nil {
			t.showErrorModal(fmt.Sprintf("Failed to install SSH key to %q", alias), copyErr.Error())
		}
	}
}

func (t *tui) handleServerSelectionChange(server domain.Server) {
	t.details.UpdateServer(server)
}

func (t *tui) handleImportKnownHosts() {
	if t.readonly {
		t.showReadonlyModal()
		return
	}

	home, err := os.UserHomeDir()
	if err != nil {
		t.showErrorModal("Import Known Hosts Error", "Failed to determine user home directory: "+err.Error())
		return
	}
	khPath := filepath.Join(home, ".ssh", "known_hosts")

	candidates, result, err := t.serverService.DiscoverKnownHosts(khPath)
	if err != nil {
		t.showErrorModal("Import Known Hosts", err.Error())
		return
	}

	if len(candidates) == 0 {
		t.showStatusTemp(fmt.Sprintf("No new hosts in %s (%d already configured)", khPath, result.Skipped))
		return
	}

	msg := fmt.Sprintf("Import %d new host(s) from %s into SSH config?\n\n(%d host(s) already configured and will be skipped)",
		len(candidates), khPath, result.Skipped)

	doImport := func() {
		res, err := t.serverService.ImportKnownHosts(khPath)
		if err != nil {
			t.showStatusTempColor("Import failed: "+err.Error(), "#FF6B6B")
			t.handleModalClose()
			return
		}
		t.refreshServerList()
		t.handleModalClose()
		t.showStatusTemp(fmt.Sprintf("Imported %d new host(s) from %s", res.Imported, khPath))
	}

	modal := tview.NewModal().
		SetText(msg).
		AddButtons([]string{"[yellow]C[-]ancel", "[yellow]I[-]mport"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonIndex == 1 {
				doImport()
				return
			}
			t.handleModalClose()
		})

	modal.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'c', 'C':
			t.handleModalClose()
			return nil
		case 'i', 'I':
			doImport()
			return nil
		}
		return event
	})

	t.app.SetRoot(modal, true)
}

func (t *tui) handleServerAdd() {
	if t.readonly {
		t.showReadonlyModal()
		return
	}
	defaultKey := t.getDefaultIdentityKey()
	if t.isActiveListFocused() {
		if server, ok := t.activeList.GetSelectedServer(); ok {
			prefill := server
			prefill.Alias = ""
			prefill.Tags = nil
			form := NewServerForm(ServerFormAdd, nil).
				SetInitialData(&prefill).
				SetDefaultIdentityKey(defaultKey).
				SetApp(t.app).
				SetVersionInfo(t.version, t.commit).
				OnSave(t.handleServerSave).
				OnCancel(t.handleFormCancel).
				SetExistingAliases(t.getExistingAliases()).
				SetExistingGroups(t.getUniqueGroups())
			t.app.SetRoot(form, true)
			return
		}
	}
	form := NewServerForm(ServerFormAdd, nil).
		SetDefaultIdentityKey(defaultKey).
		SetApp(t.app).
		SetVersionInfo(t.version, t.commit).
		OnSave(t.handleServerSave).
		OnCancel(t.handleFormCancel).
		SetExistingAliases(t.getExistingAliases()).
		SetExistingGroups(t.getUniqueGroups())
	t.app.SetRoot(form, true)
}

func (t *tui) handleServerEdit() {
	if t.readonly {
		t.showReadonlyModal()
		return
	}
	if t.isActiveListFocused() {
		t.showStatusTemp("Edit disabled for active sessions. Use 'a' to add.")
		return
	}
	if server, ok := t.serverList.GetSelectedServer(); ok {
		form := NewServerForm(ServerFormEdit, &server).
			SetApp(t.app).
			SetVersionInfo(t.version, t.commit).
			OnSave(t.handleServerSave).
			OnCancel(t.handleFormCancel).
			SetExistingAliases(t.getExistingAliasesExcept(server)).
			SetExistingGroups(t.getUniqueGroups())
		t.app.SetRoot(form, true)
	}
}

func (t *tui) handleServerClone() {
	if t.readonly {
		t.showReadonlyModal()
		return
	}
	if server, ok := t.serverList.GetSelectedServer(); ok {
		cloned := cloneServer(server)
		existingAliases := t.getExistingAliases()
		cloned.Alias = GenerateUniqueAlias(server.Alias, existingAliases)
		cloned.Aliases = []string{cloned.Alias}
		cloned.PinnedAt = time.Time{}
		cloned.LastSeen = time.Time{}
		cloned.PingStatus = ""
		cloned.PingLatency = 0

		form := NewServerForm(ServerFormAdd, nil).
			SetInitialData(&cloned).
			SetApp(t.app).
			SetVersionInfo(t.version, t.commit).
			OnSave(t.handleServerSave).
			OnCancel(t.handleFormCancel).
			SetExistingAliases(existingAliases).
			SetExistingGroups(t.getUniqueGroups())
		t.app.SetRoot(form, true)
	}
}

func cloneServer(s domain.Server) domain.Server {
	cloned := s
	cloned.Aliases = nil
	if s.IdentityFiles != nil {
		cloned.IdentityFiles = make([]string, len(s.IdentityFiles))
		copy(cloned.IdentityFiles, s.IdentityFiles)
	}
	if s.Tags != nil {
		cloned.Tags = make([]string, len(s.Tags))
		copy(cloned.Tags, s.Tags)
	}
	if s.LocalForward != nil {
		cloned.LocalForward = make([]string, len(s.LocalForward))
		copy(cloned.LocalForward, s.LocalForward)
	}
	if s.RemoteForward != nil {
		cloned.RemoteForward = make([]string, len(s.RemoteForward))
		copy(cloned.RemoteForward, s.RemoteForward)
	}
	if s.DynamicForward != nil {
		cloned.DynamicForward = make([]string, len(s.DynamicForward))
		copy(cloned.DynamicForward, s.DynamicForward)
	}
	if s.SendEnv != nil {
		cloned.SendEnv = make([]string, len(s.SendEnv))
		copy(cloned.SendEnv, s.SendEnv)
	}
	if s.SetEnv != nil {
		cloned.SetEnv = make([]string, len(s.SetEnv))
		copy(cloned.SetEnv, s.SetEnv)
	}
	return cloned
}

func (t *tui) handleServerSave(server domain.Server, original *domain.Server) {
	if t.readonly {
		t.showReadonlyModal()
		return
	}
	var err error
	if original != nil {
		// Edit mode
		base := *original
		if resolved, ok, resolveErr := t.serverService.ResolveConfigServer(*original); resolveErr != nil {
			err = resolveErr
		} else if ok {
			base = resolved
		}
		if err == nil {
			err = t.serverService.UpdateServer(base, server)
		}
	} else {
		// Add mode
		err = t.serverService.AddServer(server)
	}
	if err != nil {
		var ambig *domain.ErrAmbiguousHost
		if errors.As(err, &ambig) && original != nil {
			t.showFileChoiceModal(ambig.Alias, ambig.Candidates, "Save", func(chosen string) {
				origCopy := *original
				origCopy.SourceFile = chosen
				newCopy := server
				newCopy.SourceFile = chosen
				if err := t.serverService.UpdateServer(origCopy, newCopy); err != nil {
					t.showStatusTempColor("Save failed: "+err.Error(), "#FF6B6B")
					return
				}
				_ = t.serverService.SetHidden(newCopy.Alias, newCopy.Hidden)
				t.showStatusTemp(fmt.Sprintf("Updated %s in %s", newCopy.Alias, chosen))
				t.refreshServerList()
				t.handleFormCancel()
			})
			return
		}
		// Stay on form; show a small modal with the error
		modal := tview.NewModal().
			SetText(fmt.Sprintf("Save failed: %v", err)).
			AddButtons([]string{"Close"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) { t.handleModalClose() })
		t.app.SetRoot(modal, true)
		return
	}

	_ = t.serverService.SetHidden(server.Alias, server.Hidden)

	if server.SourceFile != "" {
		verb := "Added"
		if original != nil {
			verb = "Updated"
		}
		t.showStatusTemp(fmt.Sprintf("%s %s in %s", verb, server.Alias, server.SourceFile))
	}
	t.refreshServerList()
	t.handleFormCancel()
}

func (t *tui) handleServerDelete() {
	if t.readonly {
		t.showReadonlyModal()
		return
	}
	if t.isActiveListFocused() {
		t.showStatusTemp("Cannot delete active session. Use 'K' to terminate.")
		return
	}
	if server, ok := t.serverList.GetSelectedServer(); ok {
		t.showDeleteConfirmModal(server)
	}
}

func (t *tui) handleFormCancel() {
	t.returnToMain()
}

func (t *tui) handlePingSelected() {
	if server, ok := t.serverList.GetSelectedServer(); ok {
		if server.IsWildcardServer() {
			t.showStatusTemp("Cannot ping a wildcard pattern block")
			return
		}
		alias := server.Alias

		// Set checking status
		server.PingStatus = StatusChecking
		t.pingStatuses[alias] = server
		t.serverService.UpdateServerPing(alias, StatusChecking, 0)
		t.updateServerListWithPingStatus()

		t.showStatusTemp(fmt.Sprintf("Pinging %s…", alias))
		go func() {
			up, dur, err := t.serverService.Ping(server)
			t.app.QueueUpdateDraw(func() {
				if ps, ok := t.pingStatuses[alias]; ok {
					if err != nil || !up {
						ps.PingStatus = StatusDown
						ps.PingLatency = 0
						t.showStatusTempColor(fmt.Sprintf("Ping %s: DOWN", alias), "#FF6B6B")
					} else {
						ps.PingStatus = StatusUp
						ps.PingLatency = dur
						t.showStatusTempColor(fmt.Sprintf("Ping %s: UP (%s)", alias, dur), "#A0FFA0")
					}
					t.pingStatuses[alias] = ps
					t.serverService.UpdateServerPing(alias, ps.PingStatus, ps.PingLatency)
					t.updateServerListWithPingStatus()
				}
			})
		}()
	}
}

func (t *tui) updateServerListWithPingStatus() {
	if t.serverList != nil {
		t.serverList.UpdatePingStatuses(t.pingStatuses)
	}
}

func (t *tui) handlePingAll() {
	servers := t.serverList.GetServers()
	if len(servers) == 0 {
		t.showStatusTemp("No servers to ping")
		return
	}

	t.showStatusTemp(fmt.Sprintf("Pinging all %d servers…", len(servers)))

	// Set all servers to checking status
	t.pingStatuses = make(map[string]domain.Server)
	for _, server := range servers {
		if server.IsWildcardServer() {
			continue
		}
		s := server
		s.PingStatus = StatusChecking
		t.pingStatuses[s.Alias] = s
		t.serverService.UpdateServerPing(s.Alias, StatusChecking, 0)
	}
	t.updateServerListWithPingStatus()

	// Ping all servers concurrently
	for _, server := range servers {
		if server.IsWildcardServer() {
			continue
		}
		go func(srv domain.Server) {
			up, dur, err := t.serverService.Ping(srv)
			t.app.QueueUpdateDraw(func() {
				if ps, ok := t.pingStatuses[srv.Alias]; ok {
					if err != nil || !up {
						ps.PingStatus = StatusDown
						ps.PingLatency = 0
					} else {
						ps.PingStatus = StatusUp
						ps.PingLatency = dur
					}
					t.pingStatuses[srv.Alias] = ps
					t.serverService.UpdateServerPing(srv.Alias, ps.PingStatus, ps.PingLatency)
					t.updateServerListWithPingStatus()
				}
			})
		}(server)
	}

	// Show summary after 3 seconds
	go func() {
		time.Sleep(3 * time.Second)
		t.app.QueueUpdateDraw(func() {
			upCount := 0
			downCount := 0
			for _, ps := range t.pingStatuses {
				if ps.PingStatus == StatusUp {
					upCount++
				} else if ps.PingStatus == StatusDown {
					downCount++
				}
			}
			t.showStatusTempColor(fmt.Sprintf("Ping completed: %d UP, %d DOWN", upCount, downCount), "#A0FFA0")
		})
	}()
}

func (t *tui) handleModalClose() {
	t.returnToMain()
}

// handleRefreshBackground refreshes the server list in the background without leaving the current screen.
// It preserves the current search query and selection, shows transient status, and avoids concurrent runs.
func (t *tui) handleRefreshBackground() {
	currentIdx := t.serverList.GetCurrentItem()
	query := ""
	if t.searchBar != nil {
		query = t.searchBar.InputField.GetText()
	}

	t.showStatusTemp("Refreshing…")

	go func(prevIdx int, q string) {
		_ = t.serverService.ReloadServers()
		servers, err := t.serverService.ListServers(q)
		if err != nil {
			t.app.QueueUpdateDraw(func() {
				t.showStatusTempColor(fmt.Sprintf("Refresh failed: %v", err), "#FF6B6B")
			})
			return
		}
		var active []domain.Server
		if t.activeList != nil {
			var activeErr error
			active, activeErr = t.serverService.ListActiveSessions(q)
			if activeErr != nil {
				t.app.QueueUpdateDraw(func() {
					t.showStatusTempColor(fmt.Sprintf("Active refresh failed: %v", activeErr), "#FF6B6B")
				})
				return
			}
		}
		if strings.TrimSpace(q) == "" {
			sortServersForUI(servers, t.sortMode)
		}
		t.app.QueueUpdateDraw(func() {
			t.serverList.UpdateServers(servers)
			if t.activeList != nil {
				t.activeList.UpdateServers(active)
			}
			// Try to restore selection if still valid
			if prevIdx >= 0 && prevIdx < t.serverList.List.GetItemCount() {
				t.serverList.SetCurrentItem(prevIdx)
				if srv, ok := t.serverList.GetSelectedServer(); ok {
					t.details.UpdateServer(srv)
				}
			}
			t.showStatusTemp(fmt.Sprintf("Refreshed %d servers", len(servers)))
		})
	}(currentIdx, query)
}

// =============================================================================
// UI Display Functions (show UI elements/modals)
// =============================================================================

func (t *tui) showReadonlyModal() {
	modal := tview.NewModal().
		SetText(ReadonlyMessage).
		AddButtons([]string{"Close"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			t.handleModalClose()
		})
	modal.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape || event.Rune() == 'q' {
			t.handleModalClose()
			return nil
		}
		return event
	})
	if t.app != nil {
		t.app.SetRoot(modal, true)
		t.app.SetFocus(modal)
	}
	t.showStatusTempColor(ReadonlyMessage, "#FF6B6B")
}

func (t *tui) showDeleteConfirmModal(server domain.Server) {
	if t.readonly {
		t.showReadonlyModal()
		return
	}
	msg := fmt.Sprintf("Delete server %s (%s@%s:%d)?\n\nThis action cannot be undone.",
		server.Alias, server.User, server.Host, server.Port)

	modal, pages := t.newDeleteConfirmationOverlay(server, msg)
	t.app.SetRoot(pages, true)
	t.app.SetFocus(modal)
}

func (t *tui) newDeleteConfirmationOverlay(server domain.Server, msg string) (*tview.Modal, *tview.Pages) {
	doDelete := func() {
		err := t.serverService.DeleteServer(server)
		if err != nil {
			var ambig *domain.ErrAmbiguousHost
			if errors.As(err, &ambig) {
				t.showFileChoiceModal(ambig.Alias, ambig.Candidates, "Delete", func(chosen string) {
					srv := server
					srv.SourceFile = chosen
					if err := t.serverService.DeleteServer(srv); err != nil {
						t.showStatusTempColor("Delete failed: "+err.Error(), "#FF6B6B")
						return
					}
					t.showStatusTemp(fmt.Sprintf("Deleted %s from %s", srv.Alias, chosen))
					t.refreshServerList()
					t.handleModalClose()
				})
				return
			}
			t.showStatusTempColor("Delete failed: "+err.Error(), "#FF6B6B")
			t.handleModalClose()
			return
		}
		t.showStatusTemp(fmt.Sprintf("Deleted %s", server.Alias))
		t.refreshServerList()
		t.handleModalClose()
	}

	modal := tview.NewModal().
		SetText(msg).
		AddButtons([]string{"[yellow]C[-]ancel", "[red]D[-]elete"}).
		SetBackgroundColor(tcell.Color235).
		SetTextColor(tcell.Color252).
		SetButtonStyle(tcell.StyleDefault.Foreground(tcell.Color252).Background(tcell.Color232)).
		SetButtonActivatedStyle(tcell.StyleDefault.Foreground(tcell.Color232).Background(tcell.Color252)).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonIndex == 1 {
				doDelete()
				return
			}
			t.handleModalClose()
		})
	modal.SetBorderColor(tcell.ColorRed)
	modal.SetTitle(" Confirm Deletion ")
	modal.SetTitleAlign(tview.AlignCenter)
	modal.SetTitleColor(tcell.ColorRed)
	modal.SetFocus(0)

	modal.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			t.handleModalClose()
			return nil
		}
		switch commandKey(event) {
		case 'c':
			t.handleModalClose()
			return nil
		case 'd':
			doDelete()
			return nil
		}
		return event
	})

	pages := tview.NewPages().
		AddPage("main", t.root, true, true).
		AddPage("delete-confirmation", modal, true, true)
	pages.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
		if modal.InRect(event.Position()) {
			return action, event
		}
		return tview.MouseConsumed, nil
	})

	return modal, pages
}

// showFileChoiceModal asks the user which config file to apply an action to
// when the same alias is defined in multiple files. action is a verb shown
// on the confirmation buttons (e.g. "Save", "Delete"). onChoose is called
// with the chosen absolute file path; Cancel closes the modal.
func (t *tui) showFileChoiceModal(alias string, candidates []string, action string, onChoose func(path string)) {
	msg := fmt.Sprintf("Host %q is defined in multiple files.\nWhich file should %s use?", alias, action)
	buttons := append([]string{}, candidates...)
	buttons = append(buttons, "Cancel")

	modal := tview.NewModal().
		SetText(msg).
		AddButtons(buttons).
		SetDoneFunc(func(idx int, label string) {
			if idx < 0 || idx >= len(candidates) {
				t.handleModalClose()
				return
			}
			t.handleModalClose()
			onChoose(candidates[idx])
		})
	modal.SetBorderColor(BorderColorUnfocused)
	modal.SetTitle(fmt.Sprintf(" Multiple Files: %s ", action))
	modal.SetTitleAlign(tview.AlignCenter)
	modal.SetTitleColor(TitleColorUnfocused)
	modal.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			t.handleModalClose()
			return nil
		}
		return event
	})

	pages := tview.NewPages().
		AddPage("main", t.root, true, true).
		AddPage("file-choice", modal, true, true)
	pages.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
		if modal.InRect(event.Position()) {
			return action, event
		}
		return tview.MouseConsumed, nil
	})

	t.app.SetRoot(pages, true)
	t.app.SetFocus(modal)
}

func (t *tui) showErrorModal(title, errMsg string) {
	text := fmt.Sprintf("[red]%s:[-]\n\n%s", tview.Escape(title), tview.Escape(errMsg))
	modal := tview.NewModal().
		SetText(text).
		AddButtons([]string{"Close"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			t.handleModalClose()
		})
	t.app.SetRoot(modal, true)
	t.app.SetFocus(modal)
}

func (t *tui) showSSHErrorModal(alias, errMsg string) {
	t.showErrorModal(fmt.Sprintf("SSH connection to %q failed", alias), errMsg)
}

func (t *tui) showEditTagsForm(server domain.Server) {
	form := tview.NewForm()
	form.SetBorder(true).
		SetTitle(fmt.Sprintf(" Edit Tags: %s ", server.Alias)).
		SetTitleAlign(tview.AlignCenter)

	defaultTags := strings.Join(server.Tags, ", ")
	form.AddInputField("Tags (comma):", defaultTags, 40, nil, nil)

	form.AddButton("Save", func() {
		text := strings.TrimSpace(form.GetFormItem(0).(*tview.InputField).GetText())
		var tags []string

		for _, part := range strings.Split(text, ",") {
			if s := strings.TrimSpace(part); s != "" {
				tags = append(tags, s)
			}
		}

		newServer := server
		newServer.Tags = tags
		err := t.serverService.UpdateServer(server, newServer)
		if err != nil {
			var ambig *domain.ErrAmbiguousHost
			if errors.As(err, &ambig) {
				t.showFileChoiceModal(ambig.Alias, ambig.Candidates, "Update", func(chosen string) {
					orig := server
					orig.SourceFile = chosen
					nu := newServer
					nu.SourceFile = chosen
					if err := t.serverService.UpdateServer(orig, nu); err != nil {
						t.showStatusTempColor("Tags update failed: "+err.Error(), "#FF6B6B")
						return
					}
					t.refreshServerList()
					t.showStatusTemp("Tags updated")
				})
				return
			}
			t.showStatusTempColor("Tags update failed: "+err.Error(), "#FF6B6B")
			return
		}
		t.refreshServerList()
		t.returnToMain()
		t.showStatusTemp("Tags updated")
	})
	form.AddButton("Cancel", func() { t.returnToMain() })
	form.SetCancelFunc(func() { t.returnToMain() })

	t.app.SetRoot(form, true)
	toFocus := form
	t.app.SetFocus(toFocus)
}

func (t *tui) handlePortForward() {
	if server, ok := t.serverList.GetSelectedServer(); ok {
		if server.IsWildcardServer() {
			t.showErrorModal("Port Forwarding Warning", "Cannot port forward with a wildcard pattern block")
			return
		}
		t.showPortForwardForm(server)
	}
}

func (t *tui) showPortForwardForm(server domain.Server) *tview.Form {
	modal := NewPortForwardModal(t.app, server, t.settings)

	modal.OnStart(func(fType, port, host, hostPort, bindAddr string, onlyForward bool, args []string) {
		alias := server.Alias
		if onlyForward {
			t.returnToMain()
			t.showStatusTemp("Starting port forward…")
			go func() {
				pid, err := t.serverService.StartForward(alias, args)
				t.app.QueueUpdateDraw(func() {
					if err != nil {
						t.showStatusTempColor("Forward failed: "+err.Error(), "#FF6B6B")
					} else {
						t.refreshServerList()
						t.showStatusTemp(fmt.Sprintf("Port forwarding started (pid %d)", pid))
					}
				})
			}()
			return
		}

		var sshErr error
		t.app.Suspend(func() {
			if err := t.serverService.SSHWithArgs(alias, args); err != nil {
				sshErr = err
				t.logger.Errorw("ssh session error", "alias", alias, "error", err)
			}
		})
		if t.exitOnDisconnect {
			t.app.Stop()
			return
		}
		t.app.Sync()
		t.returnToMain()
		if sshErr != nil {
			t.showSSHErrorModal(alias, sshErr.Error())
		}
	})

	modal.OnCopied(func(cmd string) {
		t.showStatusTemp("Copied command: " + cmd)
	})

	modal.OnCancel(func() {
		t.returnToMain()
	})

	modal.OnStatusTemp(func(msg string, color ...string) {
		c := ""
		if len(color) > 0 {
			c = color[0]
		}
		if c != "" {
			t.showStatusTempColor(msg, c)
		} else {
			t.showStatusTemp(msg)
		}
	})

	t.app.SetRoot(modal, true)
	t.app.SetFocus(modal.form)
	return modal.form
}

// =============================================================================
// UI State Management (hide UI elements)
// =============================================================================

// blurSearchBar moves focus back to the server list without changing layout.
func (t *tui) blurSearchBar() {
	if t.app != nil && t.serverList != nil {
		t.app.SetFocus(t.serverList)
		t.updateFocusBorders()
	}
}

// =============================================================================
// Internal Operations (perform actual work)
// =============================================================================

func (t *tui) refreshServerList() {
	query := ""
	if t.searchBar != nil {
		query = t.searchBar.InputField.GetText()
	}
	filtered, _ := t.serverService.ListServers(query)
	if strings.TrimSpace(query) == "" {
		sortServersForUI(filtered, t.sortMode)
	}
	t.serverList.UpdateServers(t.filterServersForDisplay(filtered))
	if t.activeList != nil {
		active, _ := t.serverService.ListActiveSessions(query)
		t.activeList.UpdateServers(active)
	}
}

// handleKillActiveSessions terminates active SSH sessions for the selected server.
func (t *tui) handleKillActiveSessions() {
	if t.activeList == nil {
		return
	}
	if server, ok := t.activeList.GetSelectedServer(); ok {
		go func(selected domain.Server) {
			count, err := t.serverService.KillActiveSessions(selected)
			t.app.QueueUpdateDraw(func() {
				if err != nil {
					t.showStatusTempColor("Failed to terminate SSH sessions: "+err.Error(), "#FF6B6B")
				} else {
					t.showStatusTemp(fmt.Sprintf("Terminated %d SSH session(s) for %s", count, selected.Alias))
				}
				t.refreshServerList()
			})
		}(server)
	}
}

func (t *tui) returnToMain() {
	t.app.SetRoot(t.root, true)
	if t.serverList != nil {
		t.app.SetFocus(t.serverList)
	}
	t.updateFocusBorders()
}

// showStatusTemp displays a temporary message in the status bar (default green) and then restores the default text.
func (t *tui) showStatusTemp(msg string) {
	if t.statusBar == nil {
		return
	}
	t.showStatusTempColor(msg, "#A0FFA0")
}

func (t *tui) defaultStatusText() string {
	t.autoPingMu.Lock()
	defer t.autoPingMu.Unlock()
	return t.defaultStatusTextLocked()
}

func (t *tui) defaultStatusTextLocked() string {
	base := StatusText(t.readonly)
	if t.autoPingEnabled {
		return base + fmt.Sprintf(" • [dodgerblue::b][WATCH %ds][-]", t.autoPingSecondsRemaining)
	}
	return base
}

func (t *tui) defaultStatusTextWithCountdown(rem int) string {
	base := StatusText(t.readonly)
	return base + fmt.Sprintf(" • [dodgerblue::b][WATCH %ds][-]", rem)
}

// showStatusTempColor displays a temporary colored message in the status bar and restores default text after 2s.
func (t *tui) showStatusTempColor(msg string, color string) {
	if t.statusBar == nil {
		return
	}
	t.isShowingTempStatus = true
	t.statusBar.SetText("[" + color + "]" + msg + "[-]")
	time.AfterFunc(2*time.Second, func() {
		if t.app != nil {
			t.app.QueueUpdateDraw(func() {
				t.isShowingTempStatus = false
				if t.statusBar != nil {
					t.statusBar.SetText(t.defaultStatusText())
				}
			})
		}
	})
}

func (t *tui) startAutoPing() {
	t.autoPingMu.Lock()
	defer t.autoPingMu.Unlock()
	t.startAutoPingLocked()
}

func (t *tui) stopAutoPing() {
	t.autoPingMu.Lock()
	defer t.autoPingMu.Unlock()
	t.stopAutoPingLocked()
}

func (t *tui) startAutoPingLocked() {
	if t.autoPingStop != nil {
		return
	}
	t.autoPingEnabled = true
	if t.autoPingInterval <= 0 {
		t.autoPingInterval = 60 * time.Second
	}
	t.autoPingSecondsRemaining = int(t.autoPingInterval.Seconds())
	stopCh := make(chan struct{})
	t.autoPingStop = stopCh

	if t.statusBar != nil {
		t.statusBar.SetText(t.defaultStatusTextLocked())
	}

	go t.executeBackgroundPingSweep()

	intervalSec := int(t.autoPingInterval.Seconds())
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-stopCh:
				return
			case <-ticker.C:
				t.autoPingMu.Lock()
				if !t.autoPingEnabled {
					t.autoPingMu.Unlock()
					return
				}
				t.autoPingSecondsRemaining--
				trigger := false
				if t.autoPingSecondsRemaining <= 0 {
					t.autoPingSecondsRemaining = intervalSec
					trigger = true
				}
				rem := t.autoPingSecondsRemaining
				t.autoPingMu.Unlock()

				if trigger {
					go t.executeBackgroundPingSweep()
				}

				if t.app != nil {
					t.app.QueueUpdateDraw(func() {
						t.autoPingMu.Lock()
						enabled := t.autoPingEnabled
						t.autoPingMu.Unlock()
						if enabled && t.statusBar != nil && !t.isShowingTempStatus {
							t.statusBar.SetText(t.defaultStatusTextWithCountdown(rem))
						}
					})
				}
			}
		}
	}()
}

func (t *tui) stopAutoPingLocked() {
	t.autoPingEnabled = false
	if t.autoPingStop != nil {
		close(t.autoPingStop)
		t.autoPingStop = nil
	}
	if t.statusBar != nil {
		t.statusBar.SetText(t.defaultStatusTextLocked())
	}
}

func (t *tui) handleToggleAutoPing() {
	t.autoPingMu.Lock()
	defer t.autoPingMu.Unlock()

	if t.autoPingEnabled {
		t.stopAutoPingLocked()
		if t.settings != nil {
			_ = t.settings.SaveAutoPing(false, int(t.autoPingInterval.Seconds()))
		}
		t.showStatusTemp("Ping Watch Mode: OFF")
	} else {
		t.startAutoPingLocked()
		if t.settings != nil {
			_ = t.settings.SaveAutoPing(true, int(t.autoPingInterval.Seconds()))
		}
		t.showStatusTemp(fmt.Sprintf("Ping Watch Mode: ON (every %ds)", int(t.autoPingInterval.Seconds())))
	}
}

func (t *tui) executeBackgroundPingSweep() {
	if t.serverList == nil || t.serverService == nil {
		return
	}
	servers := t.serverList.GetServers()
	if len(servers) == 0 {
		return
	}

	for _, server := range servers {
		if server.IsWildcardServer() {
			continue
		}
		go func(srv domain.Server) {
			up, dur, err := t.serverService.Ping(srv)
			if t.app != nil {
				t.app.QueueUpdateDraw(func() {
					if t.pingStatuses == nil {
						t.pingStatuses = make(map[string]domain.Server)
					}
					ps := srv
					if err != nil || !up {
						ps.PingStatus = StatusDown
						ps.PingLatency = 0
					} else {
						ps.PingStatus = StatusUp
						ps.PingLatency = dur
					}
					t.pingStatuses[srv.Alias] = ps
					t.serverService.UpdateServerPing(srv.Alias, ps.PingStatus, ps.PingLatency)
					t.updateServerListWithPingStatus()
				})
			}
		}(server)
	}
}

// Stop any active port forwarding for the selected server.
func (t *tui) handleStopForwarding() {
	if server, ok := t.serverList.GetSelectedServer(); ok {
		alias := server.Alias
		go func() {
			err := t.serverService.StopForwarding(alias)
			t.app.QueueUpdateDraw(func() {
				if err != nil {
					t.showStatusTempColor("Failed to stop forwarding: "+err.Error(), "#FF6B6B")
				} else {
					t.showStatusTemp("Stopped forwarding for " + alias)
				}
				t.refreshServerList()
			})
		}()
	}
}

func (t *tui) getUniqueGroups() []string {
	servers, err := t.serverService.ListServers("")
	if err != nil {
		return nil
	}
	uniqueGroups := make(map[string]bool)
	var groups []string
	for _, s := range servers {
		if s.Group != "" && !uniqueGroups[s.Group] {
			uniqueGroups[s.Group] = true
			groups = append(groups, s.Group)
		}
	}
	sort.Strings(groups)
	return groups
}

func (t *tui) handleGroupAction(groupName string, action string) {
	if action == "menu" {
		t.showGroupContextMenu(groupName)
	} else if action == "tmux-all" {
		t.handleConnectGroupTmux(groupName)
	}
}

func (t *tui) showGroupContextMenu(groupName string) {
	menu := tview.NewModal().
		SetText(fmt.Sprintf("Group Actions: %s", groupName)).
		AddButtons([]string{"Connect to All (tmux)", "Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Connect to All (tmux)" {
				t.handleConnectGroupTmux(groupName)
			}
			t.handleModalClose()
		})
	t.app.SetRoot(menu, true)
}

func (t *tui) handleConnectGroupTmux(groupName string) {
	if _, err := exec.LookPath("tmux"); err != nil {
		t.showStatusTempColor("tmux is not installed or not found in PATH", "#FF6B6B")
		return
	}

	servers, _ := t.serverService.ListServers("")
	var groupServers []domain.Server

	for _, s := range servers {
		if s.Group == groupName || strings.HasPrefix(s.Group, groupName+"/") {
			groupServers = append(groupServers, s)
		}
	}

	if len(groupServers) == 0 {
		t.showStatusTempColor("No servers in group "+groupName, "#FF6B6B")
		return
	}

	sort.Slice(groupServers, func(i, j int) bool {
		return groupServers[i].Alias < groupServers[j].Alias
	})

	sessionName := fmt.Sprintf("neossh-%s-%d", strings.ReplaceAll(groupName, "/", "-"), time.Now().Unix())
	fullCmd := buildTmuxCommand(sessionName, groupServers)

	t.app.Suspend(func() {
		//nolint:gosec // fullCmd is safely built from quoted session and server parameters for tmux
		cmd := exec.Command("sh", "-c", fullCmd)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Printf("Error launching tmux: %v\n", err)
			fmt.Println("Press Enter to continue...")
			var dummy string
			_, _ = fmt.Scanln(&dummy)
		}
	})
}

func buildTmuxCommand(sessionName string, groupServers []domain.Server) string {
	quote := func(s string) string {
		return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
	}

	var cmdParts []string
	sshCmd0 := BuildSSHCommand(groupServers[0])

	cmdParts = append(cmdParts,
		fmt.Sprintf("tmux new-session -d -s %s %s", quote(sessionName), quote(sshCmd0)),
		fmt.Sprintf("tmux select-pane -t %s -T %s", quote(sessionName), quote(groupServers[0].Alias)),
		fmt.Sprintf("tmux set-option -t %s pane-border-status top", quote(sessionName)),
		fmt.Sprintf("tmux set-option -t %s pane-border-format %s", quote(sessionName), quote(" #{pane_title} ")),
		fmt.Sprintf("tmux set-option -t %s mouse on", quote(sessionName)),
	)

	for i := 1; i < len(groupServers); i++ {
		sshCmdI := BuildSSHCommand(groupServers[i])
		cmdParts = append(cmdParts,
			fmt.Sprintf("tmux split-window -t %s %s", quote(sessionName), quote(sshCmdI)),
			fmt.Sprintf("tmux select-pane -t %s -T %s", quote(sessionName), quote(groupServers[i].Alias)),
		)
	}

	cmdParts = append(cmdParts,
		fmt.Sprintf("tmux select-layout -t %s tiled", quote(sessionName)),
		fmt.Sprintf("tmux attach-session -t %s", quote(sessionName)),
	)

	return strings.Join(cmdParts, " ; ")
}

func (t *tui) updateDetailsForSelection() {
	if server, ok := t.serverList.GetSelectedServer(); ok {
		t.details.UpdateServer(server)
	}
}

func (t *tui) handleGitSSHSetup() {
	if t.gitService == nil {
		t.showStatusTemp("Git service not available")
		return
	}

	setup := NewGitSSHSetup(t.app, t.gitService, t.serverRepo).
		OnDone(func() {
			t.app.SetRoot(t.root, true)
			t.app.SetFocus(t.serverList)
			t.updateDetailsForSelection()
		}).
		OnCancel(func() {
			t.app.SetRoot(t.root, true)
			t.app.SetFocus(t.serverList)
		})

	if err := setup.Show(); err != nil {
		modal := tview.NewModal().
			SetText(fmt.Sprintf("Cannot configure Git SSH key:\n\n%v", err)).
			AddButtons([]string{"OK"}).
			SetDoneFunc(func(_ int, _ string) {
				t.app.SetRoot(t.root, true)
				t.app.SetFocus(t.serverList)
			})
		t.app.SetRoot(modal, true)
	}
}

func (t *tui) handleEditServerKeyComment() {
	if t.readonly {
		t.showReadonlyModal()
		return
	}
	if t.gitService == nil {
		t.showStatusTemp("Git service not available")
		return
	}

	server, ok := t.serverList.GetSelectedServer()
	if !ok {
		t.showStatusTemp("No server selected")
		return
	}

	sshKey := t.getSSHKeyForServer(server)
	if sshKey == nil {
		t.showStatusTemp("Selected server has no SSH key configured")
		return
	}

	if sshKey.Path == "" {
		t.showStatusTemp("Cannot edit comment: key path is unknown")
		return
	}

	wasLoaded := sshKey.LoadedInAgent

	editor := NewEditKeyComment(t.app).
		SetKey(sshKey.Name, sshKey.Path, sshKey.Comment).
		OnSave(func(newComment string) {
			if sshKey.IsEncrypted {
				t.app.Suspend(func() {
					if err := t.gitService.UpdateKeyComment(sshKey.Path, newComment); err != nil {
						fmt.Printf("\nFailed to update comment: %v\n", err)
						fmt.Print("Press Enter to continue...")
						var dummy string
						_, _ = fmt.Scanln(&dummy)
					}
				})
			} else {
				if err := t.gitService.UpdateKeyComment(sshKey.Path, newComment); err != nil {
					t.showStatusTemp(fmt.Sprintf("Failed to update comment: %v", err))
					return
				}
			}

			if wasLoaded && sshKey.PublicKeyLine != "" {
				_ = t.gitService.UnloadKeyFromAgent(sshKey.PublicKeyLine)
				t.showStatusTemp("Key comment updated! Press 'l' to reload into ssh-agent.")
			} else {
				t.showStatusTemp("SSH key comment updated successfully")
			}

			t.updateDetailsForSelection()
		}).
		OnCancel(func() {
			t.app.SetRoot(t.root, true)
			t.app.SetFocus(t.serverList)
		})

	if err := editor.Show(); err != nil {
		modal := tview.NewModal().
			SetText(fmt.Sprintf("Cannot edit SSH key comment:\n\n%v", err)).
			AddButtons([]string{"OK"}).
			SetDoneFunc(func(_ int, _ string) {
				t.app.SetRoot(t.root, true)
				t.app.SetFocus(t.serverList)
			})
		t.app.SetRoot(modal, true)
	}
}

func (t *tui) handleLoadServerKeyToAgent() {
	if t.readonly {
		t.showReadonlyModal()
		return
	}
	if t.gitService == nil {
		t.showStatusTemp("Git service not available")
		return
	}

	server, ok := t.serverList.GetSelectedServer()
	if !ok {
		t.showStatusTemp("No server selected")
		return
	}

	sshKey := t.getSSHKeyForServer(server)
	if sshKey == nil {
		t.showStatusTemp("No SSH key configured for this server")
		return
	}

	if sshKey.IsEncrypted {
		t.app.Suspend(func() {
			if err := t.gitService.LoadKeyToAgent(sshKey.Path); err != nil {
				fmt.Printf("\nFailed to load key to ssh-agent: %v\n", err)
				fmt.Print("Press Enter to continue...")
				var dummy string
				_, _ = fmt.Scanln(&dummy)
			}
		})
	} else {
		if err := t.gitService.LoadKeyToAgent(sshKey.Path); err != nil {
			t.showStatusTemp(fmt.Sprintf("Failed to load key: %v", err))
			return
		}
	}

	t.showStatusTemp(fmt.Sprintf("Key %s loaded into ssh-agent", sshKey.Name))
	t.updateDetailsForSelection()
}

func (t *tui) handleUnloadServerKeyFromAgent() {
	if t.readonly {
		t.showReadonlyModal()
		return
	}
	if t.gitService == nil {
		t.showStatusTemp("Git service not available")
		return
	}

	server, ok := t.serverList.GetSelectedServer()
	if !ok {
		t.showStatusTemp("No server selected")
		return
	}

	sshKey := t.getSSHKeyForServer(server)
	if sshKey == nil {
		t.showStatusTemp("No SSH key configured for this server")
		return
	}

	if sshKey.PublicKeyLine == "" {
		t.showStatusTemp("Cannot unload key: public key line not available")
		return
	}

	if err := t.gitService.UnloadKeyFromAgent(sshKey.PublicKeyLine); err != nil {
		t.showStatusTemp(fmt.Sprintf("Failed to unload key: %v", err))
		return
	}

	t.showStatusTemp(fmt.Sprintf("Key %s unloaded from ssh-agent", sshKey.Name))
	t.updateDetailsForSelection()
}

func (t *tui) getSSHKeyForServer(server domain.Server) *domain.SSHKey {
	if t.gitService == nil || t.serverRepo == nil || len(server.IdentityFiles) == 0 {
		return nil
	}

	identityFile := server.IdentityFiles[0]
	if strings.HasPrefix(identityFile, "~/") {
		if homeDir, err := os.UserHomeDir(); err == nil {
			identityFile = filepath.Join(homeDir, identityFile[2:])
		}
	} else if identityFile == "~" {
		if homeDir, err := os.UserHomeDir(); err == nil {
			identityFile = homeDir
		}
	}

	allKeys, err := t.gitService.ListAllSSHKeys(t.serverRepo)
	if err != nil {
		return nil
	}

	for _, key := range allKeys {
		if key.Path == identityFile {
			return &key
		}
	}

	return nil
}
