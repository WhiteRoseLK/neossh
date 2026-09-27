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
	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// PortForwardModal is an interactive modal dialog for SSH port forwarding and tunnel assistance.
type PortForwardModal struct {
	*tview.Flex
	app      *tview.Application
	server   domain.Server
	settings *settingsManager

	form        *tview.Form
	infoText    *tview.TextView
	previewText *tview.TextView

	typeChoices []string
	modeChoices []string

	currentTypeIdx int
	currentModeIdx int
	portVal        string
	hostVal        string
	hostPortVal    string
	bindAddrVal    string
	useAlias       bool

	savedProfiles      []TunnelProfile
	selectedProfileIdx int
	profileNameVal     string

	typeDropDown     *tview.DropDown
	portField        *tview.InputField
	hostField        *tview.InputField
	hostPortField    *tview.InputField
	bindAddrField    *tview.InputField
	modeDropDown     *tview.DropDown
	profileDropDown  *tview.DropDown
	profileNameField *tview.InputField
	useAliasCheckbox *tview.Checkbox

	onStart      func(fType, port, host, hostPort, bindAddr string, onlyForward bool, args []string)
	onCopied     func(cmd string)
	onCancel     func()
	onStatusTemp func(msg string, color ...string)
}

// NewPortForwardModal creates a new Port Forwarding & Tunnel Assistant modal dialog.
func NewPortForwardModal(app *tview.Application, server domain.Server, settings *settingsManager) *PortForwardModal {
	m := &PortForwardModal{
		Flex:        tview.NewFlex().SetDirection(tview.FlexRow),
		app:         app,
		server:      server,
		settings:    settings,
		form:        tview.NewForm(),
		infoText:    tview.NewTextView(),
		previewText: tview.NewTextView(),
		typeChoices: []string{ForwardTypeLocal, ForwardTypeRemote, ForwardTypeDynamic},
		modeChoices: []string{ForwardModeOnlyForward, ForwardModeForwardSSH},
		hostVal:     "localhost",
		useAlias:    true,
	}

	m.infoText.
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft).
		SetBorderPadding(0, 0, 1, 1)

	m.previewText.
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft).
		SetBorderPadding(0, 0, 1, 1)

	m.form.SetBorderPadding(0, 0, 1, 1)

	m.AddItem(m.infoText, 3, 0, false).
		AddItem(m.form, 0, 1, true).
		AddItem(m.previewText, 4, 0, false)

	titleFormat := i18n.T("forward.title")
	if titleFormat == "forward.title" {
		titleFormat = " SSH Port Forwarding & Tunnel Assistant: %s "
	}
	title := fmt.Sprintf(titleFormat, server.Alias)
	m.SetBorder(true).
		SetTitle(title).
		SetTitleAlign(tview.AlignLeft)

	m.buildForm()
	return m
}

func (m *PortForwardModal) buildForm() {
	m.form.Clear(true)

	// Server info header
	target := m.server.Host
	if m.server.User != "" {
		target = fmt.Sprintf("%s@%s", m.server.User, m.server.Host)
	}
	if m.server.Port != 0 && m.server.Port != 22 {
		target = fmt.Sprintf("%s:%d", target, m.server.Port)
	}
	m.infoText.SetText(fmt.Sprintf("[yellow::b]%s[-::-] (%s)\n[gray]Configure port forwarding (-L, -R, -D), manage favorite tunnel profiles, and copy commands.[-]",
		m.server.Alias, target))

	// Item 0: Type dropdown
	m.typeDropDown = tview.NewDropDown()
	m.typeDropDown.SetLabel("Type").
		SetOptions(m.typeChoices, func(_ string, index int) {
			m.currentTypeIdx = index
			isDynamic := m.typeChoices[m.currentTypeIdx] == ForwardTypeDynamic
			if isDynamic {
				if m.hostField != nil {
					m.hostField.SetText("").SetDisabled(true)
				}
				if m.hostPortField != nil {
					m.hostPortField.SetText("").SetDisabled(true)
				}
			} else {
				if m.hostField != nil {
					m.hostField.SetDisabled(false)
					if m.hostField.GetText() == "" {
						m.hostField.SetText("localhost")
						m.hostVal = "localhost"
					}
				}
				if m.hostPortField != nil {
					m.hostPortField.SetDisabled(false)
				}
			}
			m.updatePreview()
		})
	m.typeDropDown.SetCurrentOption(m.currentTypeIdx)
	m.form.AddFormItem(m.typeDropDown)

	// Item 1: Port input
	m.portField = tview.NewInputField().
		SetLabel("Port").
		SetText(m.portVal).
		SetFieldWidth(8).
		SetChangedFunc(func(text string) {
			m.portVal = strings.TrimSpace(text)
			m.updatePreview()
		})
	m.form.AddFormItem(m.portField)

	// Item 2: Destination Host input
	m.hostField = tview.NewInputField().
		SetLabel("Host").
		SetText(m.hostVal).
		SetFieldWidth(35).
		SetChangedFunc(func(text string) {
			m.hostVal = strings.TrimSpace(text)
			m.updatePreview()
		})
	m.form.AddFormItem(m.hostField)

	// Item 3: Destination Host Port input
	m.hostPortField = tview.NewInputField().
		SetLabel("Host Port").
		SetText(m.hostPortVal).
		SetFieldWidth(8).
		SetChangedFunc(func(text string) {
			m.hostPortVal = strings.TrimSpace(text)
			m.updatePreview()
		})
	m.form.AddFormItem(m.hostPortField)

	// Item 4: Bind Address input
	m.bindAddrField = tview.NewInputField().
		SetLabel("Bind Address (optional)").
		SetText(m.bindAddrVal).
		SetFieldWidth(35).
		SetChangedFunc(func(text string) {
			m.bindAddrVal = strings.TrimSpace(text)
			m.updatePreview()
		})
	m.form.AddFormItem(m.bindAddrField)

	// Item 5: Mode dropdown
	m.modeDropDown = tview.NewDropDown()
	m.modeDropDown.SetLabel("Mode").
		SetOptions(m.modeChoices, func(_ string, index int) {
			m.currentModeIdx = index
			m.updatePreview()
		})
	m.modeDropDown.SetCurrentOption(m.currentModeIdx)
	m.form.AddFormItem(m.modeDropDown)

	// Item 6: Saved Profiles dropdown
	m.profileDropDown = tview.NewDropDown().SetLabel("Favorite Profile")
	m.form.AddFormItem(m.profileDropDown)
	m.reloadProfiles("")

	// Item 7: Profile Name input (for saving)
	m.profileNameField = tview.NewInputField().
		SetLabel("Profile Name (to save)").
		SetText(m.profileNameVal).
		SetFieldWidth(25).
		SetChangedFunc(func(text string) {
			m.profileNameVal = strings.TrimSpace(text)
		})
	m.form.AddFormItem(m.profileNameField)

	// Item 8: Use SSH Alias checkbox
	m.useAliasCheckbox = tview.NewCheckbox().
		SetLabel("Use SSH Alias").
		SetChecked(m.useAlias).
		SetChangedFunc(func(checked bool) {
			m.useAlias = checked
			m.updatePreview()
		})
	m.form.AddFormItem(m.useAliasCheckbox)

	// Buttons
	btnStart := i18n.T("forward.btn_start")
	if btnStart == "forward.btn_start" {
		btnStart = "Start"
	}
	m.form.AddButton(btnStart, func() {
		m.startForward()
	})

	btnCopy := i18n.T("forward.btn_copy")
	if btnCopy == "forward.btn_copy" {
		btnCopy = "Copy Command"
	}
	m.form.AddButton(btnCopy, func() {
		m.copyAndClose()
	})

	btnSave := i18n.T("forward.btn_save")
	if btnSave == "forward.btn_save" {
		btnSave = "Save Profile"
	}
	m.form.AddButton(btnSave, func() {
		m.saveCurrentProfile()
	})

	btnDelete := i18n.T("forward.btn_delete")
	if btnDelete == "forward.btn_delete" {
		btnDelete = "Delete Profile"
	}
	m.form.AddButton(btnDelete, func() {
		m.deleteSelectedProfile()
	})

	btnCancel := i18n.T("forward.btn_cancel")
	if btnCancel == "forward.btn_cancel" {
		btnCancel = "Cancel"
	}
	m.form.AddButton(btnCancel, func() {
		m.cancel()
	})

	m.form.SetCancelFunc(func() {
		m.cancel()
	})

	m.form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		//nolint:exhaustive // Only escape and ctrl keys handled
		switch event.Key() {
		case tcell.KeyEscape:
			m.cancel()
			return nil
		case tcell.KeyCtrlS:
			m.copyAndClose()
			return nil
		default:
			return event
		}
	})

	m.updatePreview()
}

// OnStart registers the callback when Start is pressed.
func (m *PortForwardModal) OnStart(fn func(fType, port, host, hostPort, bindAddr string, onlyForward bool, args []string)) *PortForwardModal {
	m.onStart = fn
	return m
}

// OnCopied registers the callback when Copy Command is pressed.
func (m *PortForwardModal) OnCopied(fn func(cmd string)) *PortForwardModal {
	m.onCopied = fn
	return m
}

// OnCancel registers the callback when Cancel is pressed.
func (m *PortForwardModal) OnCancel(fn func()) *PortForwardModal {
	m.onCancel = fn
	return m
}

// OnStatusTemp registers the callback for displaying temporary status messages.
func (m *PortForwardModal) OnStatusTemp(fn func(msg string, color ...string)) *PortForwardModal {
	m.onStatusTemp = fn
	return m
}

// Form returns the underlying tview.Form.
func (m *PortForwardModal) Form() *tview.Form {
	return m.form
}

func (m *PortForwardModal) notifyStatus(msg string, color ...string) {
	if m.onStatusTemp != nil {
		m.onStatusTemp(msg, color...)
	}
}

func (m *PortForwardModal) updatePreview() {
	cmd := m.currentCommand()
	previewLabel := i18n.T("forward.preview")
	if previewLabel == "forward.preview" {
		previewLabel = "Generated Command:"
	}
	m.previewText.SetText(fmt.Sprintf("[%s::b]%s[-::-]\n[green::b]%s[-::-]",
		CurrentTheme.HintKey, previewLabel, cmd))
}

func (m *PortForwardModal) currentCommand() string {
	ft := m.typeChoices[m.currentTypeIdx]
	onlyForward := m.modeChoices[m.currentModeIdx] == ForwardModeOnlyForward
	return BuildForwardCommand(m.server, ft, m.portVal, m.hostVal, m.hostPortVal, m.bindAddrVal, onlyForward, m.useAlias)
}

func (m *PortForwardModal) reloadProfiles(selectName string) {
	if m.settings == nil {
		m.profileDropDown.SetOptions([]string{"<Custom / New Profile>"}, nil)
		m.profileDropDown.SetCurrentOption(0)
		return
	}

	profiles, _ := m.settings.LoadTunnelProfiles(m.server.Alias)
	m.savedProfiles = profiles

	options := []string{"<Custom / New Profile>"}
	selectedIdx := 0
	for i, p := range profiles {
		label := fmt.Sprintf("%s (%s %s)", p.Name, p.Type, p.Port)
		options = append(options, label)
		if selectName != "" && strings.EqualFold(p.Name, selectName) {
			selectedIdx = i + 1
		}
	}

	m.profileDropDown.SetOptions(options, func(_ string, index int) {
		m.onProfileSelected(index)
	})
	m.selectedProfileIdx = selectedIdx
	m.profileDropDown.SetCurrentOption(selectedIdx)
}

func (m *PortForwardModal) onProfileSelected(index int) {
	m.selectedProfileIdx = index
	if index <= 0 || index-1 >= len(m.savedProfiles) {
		return
	}

	p := m.savedProfiles[index-1]
	m.profileNameVal = p.Name
	if m.profileNameField != nil {
		m.profileNameField.SetText(p.Name)
	}

	typeIdx := 0
	switch p.Type {
	case ForwardTypeRemote:
		typeIdx = 1
	case ForwardTypeDynamic:
		typeIdx = 2
	default:
		typeIdx = 0
	}
	m.currentTypeIdx = typeIdx
	if m.typeDropDown != nil {
		m.typeDropDown.SetCurrentOption(typeIdx)
	}

	m.portVal = p.Port
	if m.portField != nil {
		m.portField.SetText(p.Port)
	}

	m.hostVal = p.Host
	if m.hostField != nil {
		m.hostField.SetText(p.Host)
	}

	m.hostPortVal = p.HostPort
	if m.hostPortField != nil {
		m.hostPortField.SetText(p.HostPort)
	}

	m.bindAddrVal = p.BindAddress
	if m.bindAddrField != nil {
		m.bindAddrField.SetText(p.BindAddress)
	}

	modeIdx := 0
	if p.Mode == ForwardModeForwardSSH {
		modeIdx = 1
	}
	m.currentModeIdx = modeIdx
	if m.modeDropDown != nil {
		m.modeDropDown.SetCurrentOption(modeIdx)
	}

	m.updatePreview()
}

func (m *PortForwardModal) startForward() {
	if err := validatePort(m.portVal); err != nil {
		m.notifyStatus("Invalid port: "+err.Error(), "#FF6B6B")
		return
	}
	if m.bindAddrVal != "" {
		if err := validateBindAddress(m.bindAddrVal); err != nil {
			m.notifyStatus("Invalid bind address: "+err.Error(), "#FF6B6B")
			return
		}
	}

	ft := m.typeChoices[m.currentTypeIdx]
	if ft != ForwardTypeDynamic {
		if err := validateHost(m.hostVal); err != nil {
			m.notifyStatus("Invalid host: "+err.Error(), "#FF6B6B")
			return
		}
		if err := validatePort(m.hostPortVal); err != nil {
			m.notifyStatus("Invalid host port: "+err.Error(), "#FF6B6B")
			return
		}
	}

	onlyForward := m.modeChoices[m.currentModeIdx] == ForwardModeOnlyForward
	args := BuildForwardArgs(ft, m.portVal, m.hostVal, m.hostPortVal, m.bindAddrVal)

	if m.onStart != nil {
		m.onStart(ft, m.portVal, m.hostVal, m.hostPortVal, m.bindAddrVal, onlyForward, args)
	}
}

func (m *PortForwardModal) copyAndClose() {
	cmd := m.currentCommand()
	_ = clipboard.WriteAll(cmd)
	m.notifyStatus("Copied to clipboard: "+cmd, "#51CF66")
	if m.onCopied != nil {
		m.onCopied(cmd)
	}
}

func (m *PortForwardModal) saveCurrentProfile() {
	name := strings.TrimSpace(m.profileNameVal)
	if name == "" {
		m.notifyStatus("Profile name cannot be empty", "#FF6B6B")
		return
	}
	if err := validatePort(m.portVal); err != nil {
		m.notifyStatus("Invalid port: "+err.Error(), "#FF6B6B")
		return
	}
	ft := m.typeChoices[m.currentTypeIdx]
	if ft != ForwardTypeDynamic {
		if err := validateHost(m.hostVal); err != nil {
			m.notifyStatus("Invalid host: "+err.Error(), "#FF6B6B")
			return
		}
		if err := validatePort(m.hostPortVal); err != nil {
			m.notifyStatus("Invalid host port: "+err.Error(), "#FF6B6B")
			return
		}
	}
	if m.bindAddrVal != "" {
		if err := validateBindAddress(m.bindAddrVal); err != nil {
			m.notifyStatus("Invalid bind address: "+err.Error(), "#FF6B6B")
			return
		}
	}

	profile := TunnelProfile{
		Name:        name,
		Type:        ft,
		Port:        m.portVal,
		Host:        m.hostVal,
		HostPort:    m.hostPortVal,
		BindAddress: m.bindAddrVal,
		Mode:        m.modeChoices[m.currentModeIdx],
	}

	if m.settings != nil {
		if err := m.settings.SaveTunnelProfile(m.server.Alias, profile); err != nil {
			m.notifyStatus("Failed to save profile: "+err.Error(), "#FF6B6B")
			return
		}
	}

	m.reloadProfiles(name)
	m.notifyStatus(fmt.Sprintf("Profile '%s' saved", name), "#51CF66")
}

func (m *PortForwardModal) deleteSelectedProfile() {
	if m.selectedProfileIdx <= 0 || m.selectedProfileIdx-1 >= len(m.savedProfiles) {
		m.notifyStatus("Select a saved profile to delete", "#FF6B6B")
		return
	}

	profile := m.savedProfiles[m.selectedProfileIdx-1]
	if m.settings != nil {
		if err := m.settings.DeleteTunnelProfile(m.server.Alias, profile.Name); err != nil {
			m.notifyStatus("Failed to delete profile: "+err.Error(), "#FF6B6B")
			return
		}
	}

	m.profileNameVal = ""
	if m.profileNameField != nil {
		m.profileNameField.SetText("")
	}
	m.reloadProfiles("")
	m.notifyStatus(fmt.Sprintf("Profile '%s' deleted", profile.Name), "#51CF66")
}

func (m *PortForwardModal) cancel() {
	if m.onCancel != nil {
		m.onCancel()
	}
}

// Show renders and focuses the PortForwardModal.
func (m *PortForwardModal) Show() error {
	m.app.SetRoot(m, true)
	m.app.SetFocus(m.form)
	return nil
}
