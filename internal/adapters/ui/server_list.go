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

	"github.com/WhiteRoseLK/neossh/internal/core/domain"
	"github.com/WhiteRoseLK/neossh/internal/i18n"
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
	"github.com/rivo/tview"
)

type ServerList struct {
	*tview.List
	servers                []domain.Server
	displayedItems         []*domain.Server
	displayedHeaders       []string
	collapsedGroups        map[string]bool
	multiSelectEnabled     bool
	selectedAliases        map[string]bool
	currentWidth           int
	onSelection            func(domain.Server)
	onSelectionChange      func(domain.Server)
	onMultiSelectionChange func(count int)
	onReturnToSearch       func()
	onTab                  func()
	onBacktab              func()
	onGroupAction          func(groupName string, action string)
}

func NewServerList() *ServerList {
	list := &ServerList{
		List:               tview.NewList(),
		collapsedGroups:    make(map[string]bool),
		selectedAliases:    make(map[string]bool),
		multiSelectEnabled: true,
	}
	list.build()
	return list
}

func (sl *ServerList) build() {
	sl.List.ShowSecondaryText(false)
	sl.List.SetBorder(true).
		SetTitle(i18n.T("app.title_servers")).
		SetTitleAlign(tview.AlignCenter).
		SetBorderColor(CurrentTheme.BorderColorUnfocused).
		SetTitleColor(CurrentTheme.TitleColorUnfocused)
	sl.List.
		SetSelectedBackgroundColor(CurrentTheme.SelectedBackground).
		SetSelectedTextColor(CurrentTheme.SelectedText).
		SetHighlightFullLine(true)

	sl.List.SetChangedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		if index >= 0 && index < len(sl.displayedItems) && sl.onSelectionChange != nil {
			if item := sl.displayedItems[index]; item != nil {
				sl.onSelectionChange(*item)
			}
		}
	})

	sl.List.SetInputCapture(sl.handleKeyInput)
}

func (sl *ServerList) handleKeyInput(event *tcell.EventKey) *tcell.EventKey {
	//nolint:exhaustive // We only handle specific keys and pass through others
	switch event.Key() {
	case tcell.KeyCtrlA:
		if sl.multiSelectEnabled {
			sl.ToggleSelectAll()
			return nil
		}
	case tcell.KeyTab:
		if sl.onTab != nil {
			sl.onTab()
			return nil
		}
	case tcell.KeyBacktab:
		if sl.onBacktab != nil {
			sl.onBacktab()
			return nil
		}
	case tcell.KeyESC:
		if sl.multiSelectEnabled && len(sl.selectedAliases) > 0 {
			sl.ClearSelection()
			return nil
		}
		if sl.onReturnToSearch != nil {
			sl.onReturnToSearch()
		}
		return nil
	case tcell.KeyLeft, tcell.KeyRight, tcell.KeyBackspace, tcell.KeyBackspace2:
		if sl.onReturnToSearch != nil {
			sl.onReturnToSearch()
		}
		return nil
	case tcell.KeyDown:
		return sl.selectNext()
	case tcell.KeyUp:
		return sl.selectPrev()
	case tcell.KeyEnter, tcell.KeyRune:
		return sl.handleActionKey(event)
	}
	return event
}

func (sl *ServerList) handleActionKey(event *tcell.EventKey) *tcell.EventKey {
	isSpace := event.Key() == tcell.KeyRune && event.Rune() == ' '
	isEnter := event.Key() == tcell.KeyEnter
	isMenu := event.Key() == tcell.KeyRune && event.Rune() == 'm'
	isStar := event.Key() == tcell.KeyRune && event.Rune() == '*'

	idx := sl.List.GetCurrentItem()
	if idx < 0 || idx >= len(sl.displayedHeaders) {
		return event
	}

	groupName := sl.displayedHeaders[idx]
	if groupName != "" {
		if isSpace || isEnter {
			sl.collapsedGroups[groupName] = !sl.collapsedGroups[groupName]
			sl.UpdateServers(sl.servers)
			for i, h := range sl.displayedHeaders {
				if h == groupName {
					sl.List.SetCurrentItem(i)
					break
				}
			}
			return nil
		}
		if isMenu {
			sl.showGroupContextMenu(groupName)
			return nil
		}
		if isStar && sl.multiSelectEnabled {
			sl.ToggleSelectGroup(groupName)
			return nil
		}
	} else if idx < len(sl.displayedItems) && sl.displayedItems[idx] != nil {
		if isSpace && sl.multiSelectEnabled {
			sl.ToggleServerSelection(sl.displayedItems[idx].Alias)
			return nil
		}
		if isStar && sl.multiSelectEnabled {
			if sl.displayedItems[idx].Group != "" {
				sl.ToggleSelectGroup(sl.displayedItems[idx].Group)
			} else {
				sl.ToggleSelectAll()
			}
			return nil
		}
	}
	return event
}

func (sl *ServerList) hasAnyGroups(servers []domain.Server) bool {
	for _, s := range servers {
		if s.Group != "" {
			return true
		}
	}
	return false
}

func (sl *ServerList) addGroupHeader(fullPath, name string, depth int) {
	isCollapsed := sl.collapsedGroups[fullPath]
	icon := "[-]"
	if isCollapsed {
		icon = "[+]"
	}
	indent := strings.Repeat("  ", depth)
	badge := ""
	if sl.multiSelectEnabled && len(sl.selectedAliases) > 0 {
		selectedInGroup := 0
		totalInGroup := 0
		for _, s := range sl.serversInGroup(fullPath) {
			totalInGroup++
			if sl.selectedAliases[s.Alias] {
				selectedInGroup++
			}
		}
		if selectedInGroup > 0 {
			badge = fmt.Sprintf(" [green::b](%d/%d selected)[-]", selectedInGroup, totalInGroup)
		}
	}
	sl.List.AddItem(fmt.Sprintf("%s[yellow::b]%s %s[-]%s", indent, icon, name, badge), "", 0, nil)
	sl.displayedItems = append(sl.displayedItems, nil)
	sl.displayedHeaders = append(sl.displayedHeaders, fullPath)
}

func (sl *ServerList) isParentGroupCollapsed(fullPath string) bool {
	lastSlash := strings.LastIndex(fullPath, "/")
	if lastSlash == -1 {
		return false
	}
	parentPath := fullPath[:lastSlash]
	return sl.collapsedGroups[parentPath]
}

func (sl *ServerList) processServerGroupHeaders(group string, lastParts []string) (bool, []string) {
	currentGroup := group
	if currentGroup == "" {
		currentGroup = "Ungrouped"
	}
	parts := strings.Split(currentGroup, "/")

	commonLen := 0
	for j := 0; j < len(lastParts) && j < len(parts); j++ {
		if lastParts[j] == parts[j] {
			commonLen++
		} else {
			break
		}
	}

	fullPath := ""
	serverVisible := true
	for j, part := range parts {
		if j > 0 {
			fullPath += "/"
		}
		fullPath += part

		if sl.isParentGroupCollapsed(fullPath) {
			serverVisible = false
			break
		}

		if j >= commonLen {
			sl.addGroupHeader(fullPath, part, j)
		}

		if sl.collapsedGroups[fullPath] {
			serverVisible = false
		}
	}

	return serverVisible, parts
}

func (sl *ServerList) cleanStaleSelections(servers []domain.Server) {
	if len(sl.selectedAliases) == 0 {
		return
	}
	existing := make(map[string]bool, len(servers))
	for _, s := range servers {
		existing[s.Alias] = true
	}
	for alias := range sl.selectedAliases {
		if !existing[alias] {
			delete(sl.selectedAliases, alias)
		}
	}
}

func (sl *ServerList) addServerItem(s domain.Server, idx int, maxAliasWidth, listWidth int, indent string) {
	var primary, secondary string
	if sl.multiSelectEnabled {
		primary, secondary = formatServerLine(s, maxAliasWidth, listWidth, sl.selectedAliases[s.Alias])
	} else {
		primary, secondary = formatServerLine(s, maxAliasWidth, listWidth)
	}
	if indent != "" {
		primary = indent + primary
	}
	sl.List.AddItem(primary, secondary, 0, func() {
		if sl.onSelection != nil {
			sl.onSelection(sl.servers[idx])
		}
	})
	sl.displayedItems = append(sl.displayedItems, &sl.servers[idx])
	sl.displayedHeaders = append(sl.displayedHeaders, "")
}

func (sl *ServerList) UpdateServers(servers []domain.Server) {
	currentAlias := ""
	if idx := sl.List.GetCurrentItem(); idx >= 0 && idx < len(sl.displayedItems) {
		if sl.displayedItems[idx] != nil {
			currentAlias = sl.displayedItems[idx].Alias
		}
	}

	sl.servers = servers
	sl.List.Clear()
	sl.displayedItems = make([]*domain.Server, 0, len(servers))
	sl.displayedHeaders = make([]string, 0, len(servers))

	sl.cleanStaleSelections(servers)

	maxAliasWidth := 0
	for _, s := range servers {
		width := runewidth.StringWidth(s.Alias)
		if width > maxAliasWidth {
			maxAliasWidth = width
		}
	}

	_, _, listWidth, _ := sl.List.GetInnerRect() //nolint:dogsled // only width is needed
	sl.currentWidth = listWidth

	inPinnedSection := false
	lastGroupParts := []string{}
	hasGroups := sl.hasAnyGroups(servers)

	for i := range servers {
		s := servers[i]
		isPinned := !s.PinnedAt.IsZero()

		if isPinned {
			if !inPinnedSection {
				inPinnedSection = true
				if hasGroups {
					sl.addGroupHeader("Pinned", "Pinned", 0)
				}
				lastGroupParts = []string{}
			}
			if hasGroups && sl.collapsedGroups["Pinned"] {
				continue
			}
		} else {
			if inPinnedSection {
				inPinnedSection = false
				lastGroupParts = []string{}
			}

			if hasGroups {
				visible, parts := sl.processServerGroupHeaders(s.Group, lastGroupParts)
				lastGroupParts = parts
				if !visible {
					continue
				}

				indent := strings.Repeat("  ", len(parts)+1)
				sl.addServerItem(s, i, maxAliasWidth, listWidth, indent)
				continue
			}
		}

		indent := ""
		if hasGroups && isPinned {
			indent = "  "
		}
		sl.addServerItem(s, i, maxAliasWidth, listWidth, indent)
	}

	if sl.List.GetItemCount() > 0 {
		restoreIdx := -1
		if currentAlias != "" {
			for i, item := range sl.displayedItems {
				if item != nil && item.Alias == currentAlias {
					restoreIdx = i
					break
				}
			}
		}
		if restoreIdx == -1 {
			for i, item := range sl.displayedItems {
				if item != nil {
					restoreIdx = i
					break
				}
			}
			if restoreIdx == -1 {
				restoreIdx = 0
			}
		}
		sl.List.SetCurrentItem(restoreIdx)
		if sl.onSelectionChange != nil && sl.displayedItems[restoreIdx] != nil {
			sl.onSelectionChange(*sl.displayedItems[restoreIdx])
		}
	}
}

func (sl *ServerList) RefreshDisplay() {
	_, _, width, _ := sl.List.GetInnerRect() //nolint:dogsled // only width is needed
	if width != sl.currentWidth && width > 0 {
		sl.currentWidth = width
		currentIdx := sl.List.GetCurrentItem()
		sl.UpdateServers(sl.servers)
		if currentIdx >= 0 && currentIdx < sl.List.GetItemCount() {
			sl.List.SetCurrentItem(currentIdx)
		}
	}
}

func (sl *ServerList) GetSelectedServer() (domain.Server, bool) {
	idx := sl.List.GetCurrentItem()
	if idx >= 0 && idx < len(sl.displayedItems) {
		item := sl.displayedItems[idx]
		if item != nil {
			return *item, true
		}
	}
	return domain.Server{}, false
}

// GetServers returns a shallow copy of the current server list held in memory.
func (sl *ServerList) GetServers() []domain.Server {
	if sl.servers == nil {
		return nil
	}
	cp := make([]domain.Server, len(sl.servers))
	copy(cp, sl.servers)
	return cp
}

// UpdatePingStatuses updates the ping status and latency for the current servers in memory
// and refreshes the display without reloading from disk.
func (sl *ServerList) UpdatePingStatuses(statuses map[string]domain.Server) {
	for i := range sl.servers {
		if ps, ok := statuses[sl.servers[i].Alias]; ok {
			sl.servers[i].PingStatus = ps.PingStatus
			sl.servers[i].PingLatency = ps.PingLatency
		}
	}
	sl.UpdateServers(sl.servers)
}

func (sl *ServerList) OnSelection(fn func(server domain.Server)) *ServerList {
	sl.onSelection = fn
	return sl
}

func (sl *ServerList) OnSelectionChange(fn func(server domain.Server)) *ServerList {
	sl.onSelectionChange = fn
	return sl
}

func (sl *ServerList) OnReturnToSearch(fn func()) *ServerList {
	sl.onReturnToSearch = fn
	return sl
}

func (sl *ServerList) OnTab(fn func()) *ServerList {
	sl.onTab = fn
	return sl
}

func (sl *ServerList) OnBacktab(fn func()) *ServerList {
	sl.onBacktab = fn
	return sl
}

func (sl *ServerList) OnGroupAction(fn func(groupName string, action string)) *ServerList {
	sl.onGroupAction = fn
	return sl
}

func (sl *ServerList) showGroupContextMenu(groupName string) {
	if sl.onGroupAction != nil {
		sl.onGroupAction(groupName, "menu")
	}
}

func (sl *ServerList) selectNext() *tcell.EventKey {
	current := sl.List.GetCurrentItem()
	count := sl.List.GetItemCount()
	if count == 0 {
		return nil
	}
	if current < count-1 {
		sl.List.SetCurrentItem(current + 1)
	}
	return nil
}

func (sl *ServerList) selectPrev() *tcell.EventKey {
	current := sl.List.GetCurrentItem()
	count := sl.List.GetItemCount()
	if count == 0 {
		return nil
	}
	if current > 0 {
		sl.List.SetCurrentItem(current - 1)
	}
	return nil
}

func (sl *ServerList) SetMultiSelectEnabled(enabled bool) *ServerList {
	sl.multiSelectEnabled = enabled
	return sl
}

func (sl *ServerList) IsMultiSelectEnabled() bool {
	return sl.multiSelectEnabled
}

func (sl *ServerList) OnMultiSelectionChange(fn func(count int)) *ServerList {
	sl.onMultiSelectionChange = fn
	return sl
}

func (sl *ServerList) IsServerSelected(alias string) bool {
	return sl.selectedAliases != nil && sl.selectedAliases[alias]
}

func (sl *ServerList) SetServerSelected(alias string, selected bool) {
	if sl.selectedAliases == nil {
		sl.selectedAliases = make(map[string]bool)
	}
	if selected {
		sl.selectedAliases[alias] = true
	} else {
		delete(sl.selectedAliases, alias)
	}
	sl.UpdateServers(sl.servers)
	if sl.onMultiSelectionChange != nil {
		sl.onMultiSelectionChange(len(sl.selectedAliases))
	}
}

func (sl *ServerList) ToggleServerSelection(alias string) {
	if sl.selectedAliases == nil {
		sl.selectedAliases = make(map[string]bool)
	}
	if sl.selectedAliases[alias] {
		delete(sl.selectedAliases, alias)
	} else {
		sl.selectedAliases[alias] = true
	}
	sl.UpdateServers(sl.servers)
	if sl.onMultiSelectionChange != nil {
		sl.onMultiSelectionChange(len(sl.selectedAliases))
	}
}

func (sl *ServerList) ClearSelection() {
	sl.selectedAliases = make(map[string]bool)
	sl.UpdateServers(sl.servers)
	if sl.onMultiSelectionChange != nil {
		sl.onMultiSelectionChange(0)
	}
}

func (sl *ServerList) SelectAllDisplayed() {
	if sl.selectedAliases == nil {
		sl.selectedAliases = make(map[string]bool)
	}
	for _, item := range sl.displayedItems {
		if item != nil {
			sl.selectedAliases[item.Alias] = true
		}
	}
	sl.UpdateServers(sl.servers)
	if sl.onMultiSelectionChange != nil {
		sl.onMultiSelectionChange(len(sl.selectedAliases))
	}
}

func (sl *ServerList) DeselectAllDisplayed() {
	for _, item := range sl.displayedItems {
		if item != nil {
			delete(sl.selectedAliases, item.Alias)
		}
	}
	sl.UpdateServers(sl.servers)
	if sl.onMultiSelectionChange != nil {
		sl.onMultiSelectionChange(len(sl.selectedAliases))
	}
}

func (sl *ServerList) ToggleSelectAll() {
	allSelected := true
	displayedCount := 0
	for _, item := range sl.displayedItems {
		if item != nil {
			displayedCount++
			if !sl.selectedAliases[item.Alias] {
				allSelected = false
			}
		}
	}
	if displayedCount == 0 {
		return
	}
	if allSelected {
		sl.DeselectAllDisplayed()
	} else {
		sl.SelectAllDisplayed()
	}
}

func (sl *ServerList) serversInGroup(groupName string) []*domain.Server {
	var inGroup []*domain.Server
	for i := range sl.servers {
		s := &sl.servers[i]
		switch {
		case groupName == "Pinned":
			if !s.PinnedAt.IsZero() {
				inGroup = append(inGroup, s)
			}
		case groupName == "Ungrouped":
			if s.Group == "" && s.PinnedAt.IsZero() {
				inGroup = append(inGroup, s)
			}
		default:
			if s.Group == groupName || strings.HasPrefix(s.Group, groupName+"/") {
				inGroup = append(inGroup, s)
			}
		}
	}
	return inGroup
}

func (sl *ServerList) ToggleSelectGroup(groupName string) {
	groupServers := sl.serversInGroup(groupName)
	if len(groupServers) == 0 {
		return
	}
	allSelected := true
	for _, s := range groupServers {
		if !sl.selectedAliases[s.Alias] {
			allSelected = false
			break
		}
	}
	for _, s := range groupServers {
		if allSelected {
			delete(sl.selectedAliases, s.Alias)
		} else {
			sl.selectedAliases[s.Alias] = true
		}
	}
	sl.UpdateServers(sl.servers)
	if sl.onMultiSelectionChange != nil {
		sl.onMultiSelectionChange(len(sl.selectedAliases))
	}
}

func (sl *ServerList) GetMultiSelectedServers() []domain.Server {
	if len(sl.selectedAliases) == 0 {
		return nil
	}
	var res []domain.Server
	for _, s := range sl.servers {
		if sl.selectedAliases[s.Alias] {
			res = append(res, s)
		}
	}
	return res
}

func (sl *ServerList) GetMultiSelectionCount() int {
	return len(sl.selectedAliases)
}

func (sl *ServerList) GetTargetServers() []domain.Server {
	multi := sl.GetMultiSelectedServers()
	if len(multi) > 0 {
		return multi
	}
	if focused, ok := sl.GetSelectedServer(); ok {
		return []domain.Server{focused}
	}
	return nil
}
