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
	"unicode"

	"github.com/gdamore/tcell/v2"
	"go.uber.org/zap"

	"github.com/WhiteRoseLK/neossh/internal/core/domain"
	"github.com/WhiteRoseLK/neossh/internal/core/ports"
	"github.com/WhiteRoseLK/neossh/internal/core/services"
	"github.com/WhiteRoseLK/neossh/internal/i18n"
	"github.com/rivo/tview"
)

type App interface {
	Run() error
}

// Config holds configuration options for the TUI application.
type Config struct {
	ExitOnDisconnect   bool
	ReadOnly           bool
	InitialFilter      string
	ShowHidden         bool
	Theme              string
	Language           string
	DefaultIdentityKey string
	AutoPing           bool
	AutoPingSet        bool
	AutoPingInterval   int
	ServerRepo         ports.ServerRepository
	GitService         ports.GitService
}

type tui struct {
	logger *zap.SugaredLogger

	version  string
	commit   string
	readonly bool

	app           *tview.Application
	serverService ports.ServerService
	serverRepo    ports.ServerRepository
	gitService    ports.GitService
	settings      *settingsManager

	header     *AppHeader
	searchBar  *SearchBar
	serverList *ServerList
	activeList *ServerList
	details    *ServerDetails
	statusBar  *tview.TextView

	root    *tview.Flex
	left    *tview.Flex
	content *tview.Flex

	sortMode                 SortMode
	themeWatcher             *ThemeWatcher
	pingStatuses             map[string]domain.Server
	autoPingEnabled          bool
	autoPingInterval         time.Duration
	autoPingSecondsRemaining int
	autoPingStop             chan struct{}
	autoPingMu               sync.Mutex
	autoPingFlagProvided     bool
	isShowingTempStatus      bool
	exitOnDisconnect         bool
	initialFilter            string
	showHidden               bool
	themeFlag                string
	language                 string
	defaultIdentityKey       string
	customKeybindings        map[rune]rune
	agentStatus              domain.SSHAgentStatus
	agentStatusMu            sync.RWMutex
}

func NewTUI(logger *zap.SugaredLogger, ss ports.ServerService, version, commit string, cfg ...Config) App {
	var exitOnDisconnect bool
	var readonly bool
	var initialFilter string
	var showHidden bool
	var themeFlag string
	var language string
	var defaultIdentityKey string
	var autoPing bool
	var autoPingSet bool
	var autoPingInterval int
	var serverRepo ports.ServerRepository
	var gitService ports.GitService
	if len(cfg) > 0 {
		exitOnDisconnect = cfg[0].ExitOnDisconnect
		readonly = cfg[0].ReadOnly
		initialFilter = cfg[0].InitialFilter
		showHidden = cfg[0].ShowHidden
		themeFlag = cfg[0].Theme
		language = cfg[0].Language
		defaultIdentityKey = cfg[0].DefaultIdentityKey
		autoPing = cfg[0].AutoPing
		autoPingSet = cfg[0].AutoPingSet
		autoPingInterval = cfg[0].AutoPingInterval
		serverRepo = cfg[0].ServerRepo
		gitService = cfg[0].GitService
	}

	interval := time.Duration(autoPingInterval) * time.Second
	if interval <= 0 {
		interval = 60 * time.Second
	}

	t := &tui{
		logger:                   logger,
		app:                      tview.NewApplication(),
		serverService:            ss,
		serverRepo:               serverRepo,
		gitService:               gitService,
		version:                  version,
		commit:                   commit,
		readonly:                 readonly,
		settings:                 newSettingsManager(logger),
		pingStatuses:             make(map[string]domain.Server),
		autoPingEnabled:          autoPing,
		autoPingFlagProvided:     autoPingSet,
		autoPingInterval:         interval,
		autoPingSecondsRemaining: int(interval.Seconds()),
		exitOnDisconnect:         exitOnDisconnect,
		initialFilter:            initialFilter,
		showHidden:               showHidden,
		themeFlag:                themeFlag,
		language:                 language,
		defaultIdentityKey:       defaultIdentityKey,
	}
	t.initCustomKeybindings()
	t.refreshAgentStatus()
	return t
}

func (t *tui) refreshAgentStatus() {
	if t.serverService == nil {
		return
	}
	st := t.serverService.GetSSHAgentStatus()
	t.agentStatusMu.Lock()
	t.agentStatus = st
	t.agentStatusMu.Unlock()
}

func (t *tui) getAgentStatus() domain.SSHAgentStatus {
	t.agentStatusMu.RLock()
	defer t.agentStatusMu.RUnlock()
	return t.agentStatus
}

func (t *tui) getDefaultIdentityKey() string {
	if t.defaultIdentityKey != "" {
		return t.defaultIdentityKey
	}
	if t.serverService != nil {
		if key, err := t.serverService.GetDefaultIdentityKey(); err == nil && key != "" {
			return key
		}
	}
	if t.settings != nil {
		if key, err := t.settings.LoadDefaultIdentityKey(); err == nil && key != "" {
			return key
		}
	}
	return ""
}

func (t *tui) ShowHidden() bool {
	return t.showHidden
}

func (t *tui) InitialFilter() string {
	return t.initialFilter
}

func (t *tui) IsReadOnly() bool {
	return t.readonly
}

func (t *tui) SetReadOnly(ro bool) {
	t.readonly = ro
	if t.header != nil {
		t.header.readonly = ro
	}
	if t.details != nil {
		t.details.readonly = ro
	}
	if t.statusBar != nil {
		t.statusBar.SetText(StatusText(ro))
	}
	t.updateListTitle()
}

func (t *tui) ExitOnDisconnect() bool {
	return t.exitOnDisconnect
}

func (t *tui) Run() error {
	defer func() {
		if r := recover(); r != nil {
			t.logger.Errorw("panic recovered", "error", r)
		}
	}()
	services.SetTerminalTitle("neossh")
	defer services.RestoreTerminalTitle()
	defer t.stopAutoPing()
	t.app.EnableMouse(true)
	t.initializeI18n()
	t.initializeTheme()
	t.initializeThemeWatcher()
	t.buildComponents()
	t.loadPreferences()
	t.buildLayout()
	t.bindEvents()
	t.loadInitialData()
	t.app.SetRoot(t.root, true)
	t.logger.Infow("starting TUI application", "version", t.version, "commit", t.commit)
	if err := t.app.Run(); err != nil {
		t.logger.Errorw("application run error", "error", err)
		return err
	}
	t.stopThemeWatcher()
	return nil
}

func (t *tui) initializeI18n() {
	i18n.Init(t.language)
}

const (
	BorderColorUnfocused = tcell.Color238
	BorderColorFocused   = tcell.ColorDodgerBlue
	TitleColorUnfocused  = tcell.Color250
	TitleColorFocused    = tcell.ColorWhite
)

func configureCursor(screen tcell.Screen) {
	if screen != nil {
		screen.SetCursorStyle(tcell.CursorStyleBlinkingBlock)
	}
}

func (t *tui) initializeTheme() {
	switch {
	case t.themeFlag != "":
		SetTheme(t.themeFlag)
	case t.settings != nil:
		if th, err := t.settings.LoadTheme(); err == nil && th != "" {
			SetTheme(th)
		} else if t.serverService != nil {
			if th, err := t.serverService.GetTheme(); err == nil && th != "" {
				SetTheme(th)
			}
		}
	case t.serverService != nil:
		if th, err := t.serverService.GetTheme(); err == nil && th != "" {
			SetTheme(th)
		}
	}
	ApplyTheme()
}

func (t *tui) initializeThemeWatcher() {
	t.themeWatcher = NewThemeWatcher(func(newTheme string) {
		if CurrentThemeMode != ThemeSystem {
			return
		}
		if newTheme == CurrentTheme.Name {
			return
		}
		if t.app != nil {
			t.app.QueueUpdateDraw(func() {
				if newTheme == ThemeLight {
					CurrentTheme = &LightTheme
				} else {
					CurrentTheme = &DarkTheme
				}
				ApplyTheme()
				t.rebuildUI()
				t.showStatusTemp("Theme: " + newTheme + " (system)")
			})
		}
	})
	if CurrentThemeMode == ThemeSystem {
		t.themeWatcher.Start()
	}
}

func (t *tui) stopThemeWatcher() {
	if t.themeWatcher != nil {
		t.themeWatcher.Stop()
	}
}

func (t *tui) buildComponents() {
	t.header = NewAppHeader(t.version, t.commit, RepoURL, t.readonly)
	t.searchBar = NewSearchBar().
		OnSearch(t.handleSearchInput).
		OnEscape(t.blurSearchBar).
		OnNavigate(t.handleSearchNavigate).
		OnTab(t.handleServerListFocus).
		OnBacktab(t.handleDetailsFocus)
	IsForwarding = t.serverService.IsForwarding

	t.serverList = NewServerList().
		OnSelectionChange(t.handleServerSelectionChange).
		OnReturnToSearch(t.handleReturnToSearch).
		OnTab(t.handleActiveListFocus).
		OnBacktab(t.handleSearchFocus).
		OnGroupAction(t.handleGroupAction)
	t.activeList = NewServerList().
		OnSelectionChange(t.handleServerSelectionChange).
		OnReturnToSearch(t.handleReturnToSearch).
		OnTab(t.handleDetailsFocus).
		OnBacktab(t.handleServerListFocus)
	t.activeList.SetTitle(i18n.T("app.title_active"))
	t.details = NewServerDetails(t.readonly).
		SetGitService(t.gitService, t.serverRepo).
		SetAgentStatusFunc(t.getAgentStatus).
		OnTab(t.handleSearchFocus).
		OnBacktab(t.handleActiveListFocus).
		OnEscape(t.handleServerListFocus)
	t.statusBar = NewStatusBar(t.readonly)
	t.statusBar.SetText(t.defaultStatusText())

	// default sort mode
	t.sortMode = SortByAliasAsc
}

func (t *tui) loadPreferences() {
	if t.settings == nil {
		return
	}

	if mode, err := t.settings.LoadSortMode(); err == nil {
		t.sortMode = mode
	} else {
		t.logger.Warnw("failed to load sort mode preference", "error", err)
	}

	if !t.autoPingFlagProvided {
		if enabled, intervalSec, err := t.settings.LoadAutoPing(); err == nil {
			t.autoPingEnabled = enabled
			if t.autoPingInterval <= 0 {
				t.autoPingInterval = time.Duration(intervalSec) * time.Second
			}
		}
	}
	if t.autoPingInterval <= 0 {
		t.autoPingInterval = 60 * time.Second
	}
	t.autoPingSecondsRemaining = int(t.autoPingInterval.Seconds())
	if t.autoPingEnabled {
		t.startAutoPing()
	}
}

func (t *tui) buildLayout() {
	t.left = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(t.searchBar, 3, 0, false).
		AddItem(t.serverList, 0, 1, true).
		AddItem(t.activeList, 8, 0, false)

	right := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(t.details, 0, 1, false)

	t.content = tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(t.left, 0, 3, true).
		AddItem(right, 0, 2, false)

	t.root = tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(t.header, 2, 0, false).
		AddItem(t.content, 0, 1, true).
		AddItem(t.statusBar, 1, 0, false)
}

func (t *tui) bindEvents() {
	t.root.SetInputCapture(t.handleGlobalKeys)
	t.app.SetBeforeDrawFunc(func(screen tcell.Screen) bool {
		configureCursor(screen)
		t.updateFocusBorders()
		if t.serverList != nil {
			t.serverList.RefreshDisplay()
		}
		if t.activeList != nil {
			t.activeList.RefreshDisplay()
		}
		return false
	})
}

func (t *tui) updateFocusBorders() {
	searchFocused := t.searchBar != nil && t.searchBar.HasFocus()
	listFocused := t.serverList != nil && t.serverList.HasFocus()
	activeFocused := t.activeList != nil && t.activeList.HasFocus()
	detailsFocused := t.details != nil && t.details.HasFocus()

	if t.searchBar != nil {
		if searchFocused {
			t.searchBar.SetBorderColor(CurrentTheme.BorderColorFocused).SetTitleColor(CurrentTheme.TitleColorFocused)
		} else {
			t.searchBar.SetBorderColor(CurrentTheme.BorderColorUnfocused).SetTitleColor(CurrentTheme.TitleColorUnfocused)
		}
	}

	if t.serverList != nil {
		if listFocused {
			t.serverList.SetBorderColor(CurrentTheme.BorderColorFocused).SetTitleColor(CurrentTheme.TitleColorFocused)
		} else {
			t.serverList.SetBorderColor(CurrentTheme.BorderColorUnfocused).SetTitleColor(CurrentTheme.TitleColorUnfocused)
		}
	}

	if t.activeList != nil {
		if activeFocused {
			t.activeList.SetBorderColor(CurrentTheme.BorderColorFocused).SetTitleColor(CurrentTheme.TitleColorFocused)
		} else {
			t.activeList.SetBorderColor(CurrentTheme.BorderColorUnfocused).SetTitleColor(CurrentTheme.TitleColorUnfocused)
		}
	}

	if t.details != nil {
		if detailsFocused {
			t.details.SetBorderColor(CurrentTheme.BorderColorFocused).SetTitleColor(CurrentTheme.TitleColorFocused)
		} else {
			t.details.SetBorderColor(CurrentTheme.BorderColorUnfocused).SetTitleColor(CurrentTheme.TitleColorUnfocused)
		}
	}
}

// rebuildUI rebuilds all UI components to apply theme changes.
// It preserves the current state (search query, selection, sort mode).
func (t *tui) rebuildUI() {
	query := ""
	if t.searchBar != nil {
		query = t.searchBar.InputField.GetText()
	}
	currentIdx := 0
	if t.serverList != nil {
		currentIdx = t.serverList.GetCurrentItem()
	}
	activeIdx := 0
	if t.activeList != nil {
		activeIdx = t.activeList.GetCurrentItem()
	}

	t.buildComponents()
	t.buildLayout()
	t.bindEvents()

	if query != "" {
		t.searchBar.InputField.SetText(query)
	}
	t.loadInitialData()
	if currentIdx >= 0 && currentIdx < t.serverList.GetItemCount() {
		t.serverList.SetCurrentItem(currentIdx)
	}
	if activeIdx >= 0 && activeIdx < t.activeList.GetItemCount() {
		t.activeList.SetCurrentItem(activeIdx)
	}
	if srv, ok := t.serverList.GetSelectedServer(); ok {
		t.details.UpdateServer(srv)
	}

	t.app.SetRoot(t.root, true)
	t.app.SetFocus(t.serverList)
	t.updateFocusBorders()
}

func (t *tui) filterServersForDisplay(servers []domain.Server) []domain.Server {
	if t.showHidden {
		return servers
	}
	filtered := make([]domain.Server, 0, len(servers))
	for _, s := range servers {
		if !s.Hidden {
			filtered = append(filtered, s)
		}
	}
	return filtered
}

func (t *tui) loadInitialData() {
	query := t.initialFilter
	servers, _ := t.serverService.ListServers(query)
	if strings.TrimSpace(query) == "" {
		sortServersForUI(servers, t.sortMode)
	}
	displayServers := t.filterServersForDisplay(servers)
	t.updateListTitle()
	t.serverList.UpdateServers(displayServers)
	if t.activeList != nil {
		active, _ := t.serverService.ListActiveSessions(query)
		t.activeList.UpdateServers(active)
	}
	if query != "" {
		t.searchBar.SetText(query)
		if len(displayServers) == 0 {
			t.details.ShowEmpty()
		}
	}
}

func (t *tui) updateListTitle() {
	if t.serverList != nil {
		roIndicator := ""
		if t.readonly {
			roIndicator = " [READONLY]"
		}
		hiddenIndicator := ""
		if t.showHidden {
			hiddenIndicator = " [SHOW HIDDEN]"
		}
		serversTitle := strings.TrimSpace(i18n.T("app.title_servers"))
		t.serverList.SetTitle(fmt.Sprintf(" %s%s%s — Sort: %s ", serversTitle, roIndicator, hiddenIndicator, t.sortMode.String()))
	}
}

func (t *tui) persistSortMode() {
	if t.settings == nil {
		return
	}

	if err := t.settings.SaveSortMode(t.sortMode); err != nil {
		t.logger.Warnw("failed to save sort mode preference", "error", err)
	}
}

func (t *tui) initCustomKeybindings() {
	if t.settings == nil {
		return
	}
	raw, err := t.settings.LoadKeybindings()
	if err != nil || len(raw) == 0 {
		return
	}
	t.customKeybindings = buildCustomKeyMap(raw)
}

var canonicalActionAliases = map[string]rune{
	"add":                'a',
	"add_server":         'a',
	"edit":               'e',
	"edit_server":        'e',
	"delete":             'd',
	"delete_server":      'd',
	"clone":              'y',
	"clone_server":       'y',
	"pin":                'p',
	"pin_server":         'p',
	"copy":               'c',
	"copy_command":       'c',
	"copy_host":          'h',
	"paste":              'v',
	"paste_command":      'v',
	"scp":                'o',
	"scp_modal":          'o',
	"sshfs":              'M',
	"sshfs_modal":        'M',
	"sftp":               'F',
	"launch_sftp":        'F',
	"port_forward":       'f',
	"port_forwarding":    'f',
	"sort":               's',
	"sort_toggle":        's',
	"ping":               'g',
	"ping_server":        'g',
	"ping_all":           'G',
	"theme":              'T',
	"theme_toggle":       'T',
	"tags":               't',
	"edit_tags":          't',
	"hidden":             'm',
	"toggle_hidden":      'm',
	"show_hidden":        'H',
	"toggle_show_hidden": 'H',
	"load_key":           'l',
	"load_agent_key":     'l',
	"unload_key":         'u',
	"unload_agent_key":   'u',
	"git_ssh":            'P',
	"git_profiles":       'P',
	"install_key":        'K',
	"quit":               'q',
}

func buildCustomKeyMap(raw map[string]string) map[rune]rune {
	res := make(map[rune]rune)
	for actionOrKey, targetOrKey := range raw {
		actionOrKey = strings.TrimSpace(actionOrKey)
		targetOrKey = strings.TrimSpace(targetOrKey)
		if actionOrKey == "" || targetOrKey == "" {
			continue
		}

		// Syntax 1: "add_server": "n" (action -> key)
		if actionRune, isAction := canonicalActionAliases[strings.ToLower(actionOrKey)]; isAction {
			runes := []rune(targetOrKey)
			if len(runes) == 1 {
				res[runes[0]] = actionRune
				if actionRune != 'T' && actionRune != 'H' && actionRune != 'G' && actionRune != 'P' && actionRune != 'M' && actionRune != 'F' && actionRune != 'K' {
					res[unicode.ToUpper(runes[0])] = actionRune
					res[unicode.ToLower(runes[0])] = actionRune
				}
			}
			continue
		}

		// Syntax 2: "n": "a" (key -> key)
		srcRunes := []rune(actionOrKey)
		targetRunes := []rune(targetOrKey)
		if len(srcRunes) == 1 && len(targetRunes) == 1 {
			res[srcRunes[0]] = targetRunes[0]
		}
	}
	return res
}
