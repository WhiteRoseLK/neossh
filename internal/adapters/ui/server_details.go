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
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/WhiteRoseLK/neossh/internal/core/domain"
	"github.com/WhiteRoseLK/neossh/internal/core/ports"
	"github.com/WhiteRoseLK/neossh/internal/i18n"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type ServerDetails struct {
	*tview.TextView
	readonly        bool
	gitService      ports.GitService
	serverRepo      ports.ServerRepository
	agentStatusFunc func() domain.SSHAgentStatus
	onTab           func()
	onBacktab       func()
	onEscape        func()
}

func NewServerDetails(readonly ...bool) *ServerDetails {
	ro := false
	if len(readonly) > 0 {
		ro = readonly[0]
	}
	details := &ServerDetails{
		TextView: tview.NewTextView(),
		readonly: ro,
	}
	details.build()
	return details
}

// SetGitService configures git service and server repository for SSH key details.
func (sd *ServerDetails) SetGitService(gs ports.GitService, sr ports.ServerRepository) *ServerDetails {
	sd.gitService = gs
	sd.serverRepo = sr
	return sd
}

// SetAgentStatusFunc configures a function to query live SSH agent status.
func (sd *ServerDetails) SetAgentStatusFunc(fn func() domain.SSHAgentStatus) *ServerDetails {
	sd.agentStatusFunc = fn
	return sd
}

func (sd *ServerDetails) build() {
	sd.TextView.SetDynamicColors(true).
		SetWrap(true).
		SetBorder(true).
		SetTitle(i18n.T("app.title_details")).
		SetTitleAlign(tview.AlignCenter).
		SetBorderColor(CurrentTheme.BorderColorUnfocused).
		SetTitleColor(CurrentTheme.TitleColorUnfocused)

	sd.TextView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		//nolint:exhaustive // We only handle navigation keys and pass through others
		switch event.Key() {
		case tcell.KeyTab:
			if sd.onTab != nil {
				sd.onTab()
				return nil
			}
		case tcell.KeyBacktab:
			if sd.onBacktab != nil {
				sd.onBacktab()
				return nil
			}
		case tcell.KeyEscape:
			if sd.onEscape != nil {
				sd.onEscape()
				return nil
			}
		}
		return event
	})
}

// renderTagChips builds colored tag chips for details view.
func renderTagChips(tags []string) string {
	if len(tags) == 0 {
		return ""
	}
	chips := make([]string, 0, len(tags))
	for _, t := range tags {
		chips = append(chips, fmt.Sprintf("[%s:%s] %s [-:-:-]",
			CurrentTheme.TagChipText, CurrentTheme.TagChipBg, t))
	}
	return strings.Join(chips, " ")
}

// getSSHKeyForServer attempts to fetch SSH key details for the server's identity file.
func (sd *ServerDetails) getSSHKeyForServer(server domain.Server) *domain.SSHKey {
	if sd.gitService == nil || sd.serverRepo == nil {
		return nil
	}
	if len(server.IdentityFiles) == 0 {
		return nil
	}
	identityFile := expandIdentityFile(server.IdentityFiles[0])

	allKeys, err := sd.gitService.ListAllSSHKeys(sd.serverRepo)
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

// expandIdentityFile resolves ~ prefixes in identity file paths.
func expandIdentityFile(identityFile string) string {
	if strings.HasPrefix(identityFile, "~/") {
		if homeDir, err := os.UserHomeDir(); err == nil {
			return filepath.Join(homeDir, identityFile[2:])
		}
	} else if identityFile == "~" {
		if homeDir, err := os.UserHomeDir(); err == nil {
			return homeDir
		}
	}
	return identityFile
}

// renderKeyWithBadges renders the Key: line in basic info with type badges and file status.
func (sd *ServerDetails) renderKeyWithBadges(server domain.Server) string {
	if len(server.IdentityFiles) == 0 {
		return "[dim]-[-]"
	}

	keyPath := strings.Join(server.IdentityFiles, ", ")
	sshKey := sd.getSSHKeyForServer(server)
	if sshKey == nil {
		// No key info available — check if file exists on disk
		expandedPath := expandIdentityFile(server.IdentityFiles[0])
		if _, err := os.Stat(expandedPath); err != nil {
			return fmt.Sprintf("[white]%s[-] [red][⚠ missing][-]", keyPath)
		}
		return fmt.Sprintf("[white]%s[-]", keyPath)
	}

	// Build badges
	var badges []string

	// Key type badge
	typeLabel := strings.ToUpper(sshKey.Type)
	if sshKey.Size > 0 {
		typeLabel = fmt.Sprintf("%s-%d", typeLabel, sshKey.Size)
	}
	badges = append(badges, fmt.Sprintf("[black:#61AFEF] %s [-:-:-]", typeLabel))

	// FIDO2 badge
	if sshKey.IsFIDO2 {
		badges = append(badges, "[black:#E06C75] FIDO2 [-:-:-]")
	}

	// File status badge
	if !sshKey.FileExists {
		badges = append(badges, "[red][⚠ missing][-]")
	}

	return fmt.Sprintf("[white]%s[-] %s", keyPath, strings.Join(badges, " "))
}

// formatKeyTypeBadge returns a colored badge string for the SSH key type in the details section.
func (sd *ServerDetails) formatKeyTypeBadge(sshKey *domain.SSHKey) string {
	typeLabel := strings.ToUpper(sshKey.Type)
	if sshKey.Size > 0 {
		typeLabel = fmt.Sprintf("%s-%d", typeLabel, sshKey.Size)
	}

	badge := fmt.Sprintf("[black:#61AFEF] %s [-:-:-]", typeLabel)
	if sshKey.IsFIDO2 {
		badge += " [black:#E06C75] FIDO2/Security Key [-:-:-]"
	}
	return badge
}

func (sd *ServerDetails) renderCertBasic(cert *domain.SSHCertificate) string {
	if cert == nil {
		return ""
	}
	badge, desc := domain.FormatCertStatusBadge(cert)
	if desc != "" {
		return fmt.Sprintf("%s (%s)", badge, desc)
	}
	return badge
}

func (sd *ServerDetails) renderCertDetailsSection(cert *domain.SSHCertificate, certCmd string) string {
	if cert == nil && certCmd == "" {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("\n[::b]SSH Certificate Details[-]\n")
	if cert != nil {
		fmt.Fprintf(&sb, "  Path: [white]%s[-]\n", cert.Path)
		if cert.IsImplicit {
			sb.WriteString("  Source: [cyan]Implicit (<IdentityFile>-cert.pub)[-]\n")
		} else {
			sb.WriteString("  Source: [cyan]Explicit CertificateFile[-]\n")
		}

		badge, desc := domain.FormatCertStatusBadge(cert)
		if desc != "" {
			fmt.Fprintf(&sb, "  Status: %s (%s)\n", badge, desc)
		} else {
			fmt.Fprintf(&sb, "  Status: %s\n", badge)
		}

		if cert.FileExists {
			if !cert.ValidBefore.IsZero() {
				fmt.Fprintf(&sb, "  Valid To: [white]%s[-]\n", cert.ValidBefore.Format("2006-01-02 15:04:05"))
			}
			if !cert.ValidAfter.IsZero() {
				fmt.Fprintf(&sb, "  Valid From: [white]%s[-]\n", cert.ValidAfter.Format("2006-01-02 15:04:05"))
			}
			if cert.Lifetime > 0 && cert.Lifetime < time.Duration(1<<60) {
				fmt.Fprintf(&sb, "  Lifetime: [white]%s[-]\n", domain.FormatDuration(cert.Lifetime))
			}
			if cert.KeyID != "" {
				fmt.Fprintf(&sb, "  Key ID: [white]%s[-]\n", cert.KeyID)
			}
			if len(cert.Principals) > 0 {
				fmt.Fprintf(&sb, "  Principals: [white]%s[-]\n", strings.Join(cert.Principals, ", "))
			}
			if cert.Serial > 0 {
				fmt.Fprintf(&sb, "  Serial: [white]%d[-]\n", cert.Serial)
			}
		}
	} else {
		sb.WriteString("  Status: [yellow]⚠ Missing certificate on disk[-]\n")
	}

	if certCmd != "" {
		fmt.Fprintf(&sb, "  Renew Command: [white]%s[-]\n", certCmd)
	}

	return sb.String()
}

func (sd *ServerDetails) UpdateServer(server domain.Server) {
	lastSeen := server.LastSeen.Format("2006-01-02 15:04:05")
	if server.LastSeen.IsZero() {
		lastSeen = i18n.T("details.never")
	}
	serverKey := sd.renderKeyWithBadges(server)
	cert := domain.InspectServerCertificate(server)
	certLine := ""
	if cert != nil {
		certLine = fmt.Sprintf("  Cert: %s\n", sd.renderCertBasic(cert))
	}

	pinnedStr := "true"
	if server.PinnedAt.IsZero() {
		pinnedStr = "false"
	}
	tagsText := renderTagChips(server.Tags)
	if server.IsWildcardServer() {
		wildcardChip := "[black:#E5C07B] wildcard [-:-:-]"
		if tagsText != "" {
			tagsText = wildcardChip + " " + tagsText
		} else {
			tagsText = wildcardChip
		}
	}

	// Basic information
	aliasText := strings.Join(server.Aliases, ", ")
	if aliasText == "" {
		aliasText = server.Alias
	}
	if server.IsWildcardServer() {
		aliasText = fmt.Sprintf("%s [#E5C07B][wildcard][-]", aliasText)
	}

	userText := server.User

	hostText := server.Host

	portText := fmt.Sprintf("%d", server.Port)
	if server.Port == 0 {
		portText = ""
	}

	hiddenStr := "No"
	if server.Hidden {
		hiddenStr = "Yes"
	}

	groupText := server.Group
	if groupText == "" {
		groupText = "-"
	}

	text := fmt.Sprintf(
		"[::b]%s[-]\n\n[::b]%s[-]\n  Host: [white]%s[-]\n  User: [white]%s[-]\n"+
			"  Port: [white]%s[-]\n  Key:  %s\n%s  Group: [white]%s[-]\n"+
			"  Tags: %s\n  Pinned: [white]%s[-]\n  Hidden: [white]%s[-]\n"+
			"  Last SSH: %s\n  SSH Count: [white]%d[-]\n",
		aliasText, i18n.T("details.label.basic"), hostText, userText, portText,
		serverKey, certLine, groupText, tagsText, pinnedStr, hiddenStr,
		lastSeen, server.SSHCount)

	// Add SSH Key Details section if key is configured
	if sshKey := sd.getSSHKeyForServer(server); sshKey != nil {
		text += fmt.Sprintf("\n[::b]%s[-]\n", i18n.T("details.label.sshkey"))
		text += fmt.Sprintf("  Path: [white]%s[-]\n", sshKey.Path)

		// Type badge with FIDO2 indicator
		typeBadge := sd.formatKeyTypeBadge(sshKey)
		text += fmt.Sprintf("  Type: %s\n", typeBadge)

		if sshKey.Size > 0 {
			text += fmt.Sprintf("  Size: [white]%d bits[-]\n", sshKey.Size)
		}

		// File existence indicator
		if sshKey.FileExists {
			text += "  File: [green]✓ Found on disk[-]\n"
		} else {
			text += "  File: [red]⚠ Missing on disk[-]\n"
		}

		if sshKey.Fingerprint != "" {
			text += fmt.Sprintf("  Fingerprint: [white]%s[-]\n", sshKey.Fingerprint)
		}
		if sshKey.Comment != "" {
			text += fmt.Sprintf("  Comment: [white]%s[-]\n", sshKey.Comment)
		}
		loadedInAgent := sshKey.LoadedInAgent
		agentName := "ssh-agent"
		if sd.agentStatusFunc != nil {
			st := sd.agentStatusFunc()
			if st.Available {
				agentName = string(st.Type)
				if st.HasKey(sshKey.Fingerprint, sshKey.Comment, sshKey.Path) {
					loadedInAgent = true
				}
			}
		}

		if loadedInAgent {
			text += fmt.Sprintf("  Agent: [green]✓ Loaded in %s[-]\n", agentName)
		} else {
			text += fmt.Sprintf("  Agent: [dim]○ Not loaded in %s (press 'l' to load)[-]\n", agentName)
		}
		if sshKey.IsEncrypted {
			text += "  Status: [yellow]🔒 Encrypted (passphrase)[-]\n"
		} else {
			text += "  Status: [dim]🔓 Unencrypted[-]\n"
		}
	}

	text += sd.renderCertDetailsSection(cert, server.CertificateCommand)

	// Advanced settings section (only show non-empty fields)
	// Organized by logical grouping for better readability
	type fieldEntry struct {
		name  string
		value string
	}

	type fieldGroup struct {
		name   string
		fields []fieldEntry
	}

	// Create field groups for better organization and future extensibility
	groups := []fieldGroup{
		{
			name: "Connection & Proxy",
			fields: []fieldEntry{
				{"ProxyJump", server.ProxyJump},
				{"ProxyCommand", server.ProxyCommand},
				{"RemoteCommand", server.RemoteCommand},
				{"RequestTTY", server.RequestTTY},
				{"SessionType", server.SessionType},
				{"ConnectTimeout", server.ConnectTimeout},
				{"ConnectionAttempts", server.ConnectionAttempts},
				{"BindAddress", server.BindAddress},
				{"BindInterface", server.BindInterface},
				{"AddressFamily", server.AddressFamily},
				{"ExitOnForwardFailure", server.ExitOnForwardFailure},
				{"IPQoS", server.IPQoS},
				{"CanonicalizeHostname", server.CanonicalizeHostname},
				{"CanonicalDomains", server.CanonicalDomains},
				{"CanonicalizeFallbackLocal", server.CanonicalizeFallbackLocal},
				{"CanonicalizeMaxDots", server.CanonicalizeMaxDots},
				{"CanonicalizePermittedCNAMEs", server.CanonicalizePermittedCNAMEs},
				{"ServerAliveInterval", server.ServerAliveInterval},
				{"ServerAliveCountMax", server.ServerAliveCountMax},
				{"Compression", server.Compression},
				{"TCPKeepAlive", server.TCPKeepAlive},
				{"BatchMode", server.BatchMode},
				{"ControlMaster", server.ControlMaster},
				{"ControlPath", server.ControlPath},
				{"ControlPersist", server.ControlPersist},
			},
		},
		{
			name: "Authentication",
			fields: []fieldEntry{
				{"PubkeyAuthentication", server.PubkeyAuthentication},
				{"CertificateFile", server.CertificateFile},
				{"CertificateCommand", server.CertificateCommand},
				{"PubkeyAcceptedAlgorithms", server.PubkeyAcceptedAlgorithms},
				{"HostbasedAcceptedAlgorithms", server.HostbasedAcceptedAlgorithms},
				{"Password (sshpass)", maskedPassword(server.Password)},
				{"PasswordAuthentication", server.PasswordAuthentication},
				{"PreferredAuthentications", server.PreferredAuthentications},
				{"IdentitiesOnly", server.IdentitiesOnly},
				{"AddKeysToAgent", server.AddKeysToAgent},
				{"IdentityAgent", server.IdentityAgent},
				{"KbdInteractiveAuthentication", server.KbdInteractiveAuthentication},
				{"NumberOfPasswordPrompts", server.NumberOfPasswordPrompts},
			},
		},
		{
			name: "Forwarding",
			fields: []fieldEntry{
				{"ForwardAgent", server.ForwardAgent},
				{"ForwardX11", server.ForwardX11},
				{"ForwardX11Trusted", server.ForwardX11Trusted},
				{"LocalForward", strings.Join(server.LocalForward, ", ")},
				{"RemoteForward", strings.Join(server.RemoteForward, ", ")},
				{"DynamicForward", strings.Join(server.DynamicForward, ", ")},
				{"ClearAllForwardings", server.ClearAllForwardings},
				{"GatewayPorts", server.GatewayPorts},
			},
		},
		{
			name: "Security & Cryptography",
			fields: []fieldEntry{
				{"StrictHostKeyChecking", server.StrictHostKeyChecking},
				{"CheckHostIP", server.CheckHostIP},
				{"FingerprintHash", server.FingerprintHash},
				{"UserKnownHostsFile", server.UserKnownHostsFile},
				{"HostKeyAlgorithms", server.HostKeyAlgorithms},
				{"Ciphers", server.Ciphers},
				{"MACs", server.MACs},
				{"KexAlgorithms", server.KexAlgorithms},
				{"VerifyHostKeyDNS", server.VerifyHostKeyDNS},
				{"UpdateHostKeys", server.UpdateHostKeys},
				{"HashKnownHosts", server.HashKnownHosts},
				{"VisualHostKey", server.VisualHostKey},
			},
		},
		{
			name: "Environment & Execution",
			fields: []fieldEntry{
				{"PreConnectCommand", server.PreConnectCommand},
				{"LocalCommand", server.LocalCommand},
				{"PermitLocalCommand", server.PermitLocalCommand},
				{"EscapeChar", server.EscapeChar},
				{"SendEnv", strings.Join(server.SendEnv, ", ")},
				{"SetEnv", strings.Join(server.SetEnv, ", ")},
			},
		},
		{
			name: "Debugging",
			fields: []fieldEntry{
				{"LogLevel", server.LogLevel},
			},
		},
	}

	// Build advanced settings text without group labels for cleaner display
	hasAdvanced := false
	advancedText := fmt.Sprintf("\n[::b]%s[-]\n", i18n.T("details.label.advanced"))

	for _, group := range groups {
		for _, field := range group.fields {
			if field.value != "" {
				hasAdvanced = true
				advancedText += fmt.Sprintf("  %s: [white]%s[-]\n", field.name, field.value)
			}
		}
	}

	if hasAdvanced {
		text += advancedText
	}

	// Commands list
	cmdHeader := i18n.T("details.label.commands")
	if sd.readonly {
		text += fmt.Sprintf(
			"\n[::b]%s[-]\n  Enter: SSH connect\n  f: Port forward\n  x: Stop forwarding\n"+
				"  c: Copy SSH command\n  h: Copy Host\n  g: Ping server\n  G: Ping all servers\n"+
				"  W: Ping watch mode\n  P: Git SSH profile\n  r: Refresh list\n  p: Pin/Unpin\n"+
				"  [#888888]Modifications disabled (readonly mode)[-]",
			cmdHeader,
		)
	} else {
		text += fmt.Sprintf(
			"\n[::b]%s[-]\n  Enter: SSH connect\n  f: Port forward\n  x: Stop forwarding\n"+
				"  c: Copy SSH command\n  v: Paste SSH command\n  y: Clone server\n  h: Copy Host\n"+
				"  g: Ping server\n  G: Ping all servers\n  W: Ping watch mode\n  P: Git SSH profile\n"+
				"  C: Edit Key Comment\n  l/u: Load/Unload agent key\n  K: Install SSH Key\n"+
				"  r: Refresh list\n  a: Add new server\n  e: Edit entry\n  t: Edit tags\n"+
				"  d: Delete entry\n  p: Pin/Unpin",
			cmdHeader,
		)
	}

	sd.TextView.SetText(text)
}

func (sd *ServerDetails) ShowEmpty() {
	sd.TextView.SetText(i18n.T("details.no_match"))
}

func (sd *ServerDetails) OnTab(fn func()) *ServerDetails {
	sd.onTab = fn
	return sd
}

func (sd *ServerDetails) OnBacktab(fn func()) *ServerDetails {
	sd.onBacktab = fn
	return sd
}

func (sd *ServerDetails) OnEscape(fn func()) *ServerDetails {
	sd.onEscape = fn
	return sd
}

func maskedPassword(p string) string {
	if p == "" {
		return ""
	}
	return "••••••••"
}
