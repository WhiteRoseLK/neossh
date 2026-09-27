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
	"path/filepath"
	"strings"
	"testing"

	"github.com/WhiteRoseLK/neossh/internal/core/domain"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"go.uber.org/zap"
)

func TestBuildForwardSpec(t *testing.T) {
	tests := []struct {
		name     string
		fType    string
		port     string
		host     string
		hostPort string
		bindAddr string
		expected string
	}{
		{
			name:     "Local forward standard",
			fType:    ForwardTypeLocal,
			port:     "5432",
			host:     "localhost",
			hostPort: "5432",
			expected: "5432:localhost:5432",
		},
		{
			name:     "Local forward with bind address",
			fType:    ForwardTypeLocal,
			port:     "8080",
			host:     "remote.internal",
			hostPort: "80",
			bindAddr: "127.0.0.1",
			expected: "127.0.0.1:8080:remote.internal:80",
		},
		{
			name:     "Remote forward",
			fType:    ForwardTypeRemote,
			port:     "9000",
			host:     "localhost",
			hostPort: "3000",
			expected: "9000:localhost:3000",
		},
		{
			name:     "Dynamic SOCKS5 without bind address",
			fType:    ForwardTypeDynamic,
			port:     "1080",
			expected: "1080",
		},
		{
			name:     "Dynamic SOCKS5 with bind address",
			fType:    ForwardTypeDynamic,
			port:     "1080",
			bindAddr: "0.0.0.0",
			expected: "0.0.0.0:1080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildForwardSpec(tt.fType, tt.port, tt.host, tt.hostPort, tt.bindAddr)
			if got != tt.expected {
				t.Errorf("BuildForwardSpec() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestBuildForwardArgs(t *testing.T) {
	localArgs := BuildForwardArgs(ForwardTypeLocal, "5432", "localhost", "5432", "")
	if len(localArgs) != 2 || localArgs[0] != "-L" || localArgs[1] != "5432:localhost:5432" {
		t.Errorf("unexpected local args: %v", localArgs)
	}

	remoteArgs := BuildForwardArgs(ForwardTypeRemote, "8080", "127.0.0.1", "80", "")
	if len(remoteArgs) != 2 || remoteArgs[0] != "-R" || remoteArgs[1] != "8080:127.0.0.1:80" {
		t.Errorf("unexpected remote args: %v", remoteArgs)
	}

	dynArgs := BuildForwardArgs(ForwardTypeDynamic, "1080", "", "", "127.0.0.1")
	if len(dynArgs) != 2 || dynArgs[0] != "-D" || dynArgs[1] != "127.0.0.1:1080" {
		t.Errorf("unexpected dynamic args: %v", dynArgs)
	}
}

func TestBuildForwardCommand(t *testing.T) {
	srv := domain.Server{
		Alias:         "prod-db",
		Host:          "192.168.1.10",
		User:          "admin",
		Port:          2222,
		IdentityFiles: []string{"~/.ssh/id_ed25519"},
	}

	t.Run("with alias only forward", func(t *testing.T) {
		cmd := BuildForwardCommand(srv, ForwardTypeLocal, "5432", "localhost", "5432", "", true, true)
		expected := "ssh -N -L 5432:localhost:5432 prod-db"
		if cmd != expected {
			t.Errorf("got %q, want %q", cmd, expected)
		}
	})

	t.Run("with alias interactive SSH forward", func(t *testing.T) {
		cmd := BuildForwardCommand(srv, ForwardTypeRemote, "8080", "localhost", "80", "", false, true)
		expected := "ssh -R 8080:localhost:80 prod-db"
		if cmd != expected {
			t.Errorf("got %q, want %q", cmd, expected)
		}
	})

	t.Run("without alias", func(t *testing.T) {
		cmd := BuildForwardCommand(srv, ForwardTypeDynamic, "1080", "", "", "", true, false)
		if !strings.Contains(cmd, "-N -D 1080") {
			t.Errorf("expected -N -D 1080, got %q", cmd)
		}
		if !strings.Contains(cmd, "admin@192.168.1.10") {
			t.Errorf("expected user@host target, got %q", cmd)
		}
		if !strings.Contains(cmd, "-p 2222") {
			t.Errorf("expected -p 2222, got %q", cmd)
		}
	})

	t.Run("with empty port placeholders", func(t *testing.T) {
		cmd := BuildForwardCommand(srv, ForwardTypeLocal, "", "", "", "", true, true)
		expected := "ssh -N -L <port>:localhost:<hostport> prod-db"
		if cmd != expected {
			t.Errorf("got %q, want %q", cmd, expected)
		}
	})
}

func TestPortForwardModal_InteractionAndProfiles(t *testing.T) {
	app := tview.NewApplication()
	tmpDir := t.TempDir()
	logger := zap.NewNop().Sugar()
	sm := &settingsManager{
		filePath: filepath.Join(tmpDir, "settings.json"),
		logger:   logger,
	}

	srv := domain.Server{
		Alias: "bastion",
		Host:  "10.0.0.1",
		User:  "ubuntu",
		Port:  22,
	}

	// Pre-seed a tunnel profile
	_ = sm.SaveTunnelProfile("bastion", TunnelProfile{
		Name:     "K8s API",
		Type:     ForwardTypeLocal,
		Port:     "6443",
		Host:     "127.0.0.1",
		HostPort: "6443",
		Mode:     ForwardModeOnlyForward,
	})

	modal := NewPortForwardModal(app, srv, sm)

	var lastStatus string
	modal.OnStatusTemp(func(msg string, color ...string) {
		lastStatus = msg
	})

	var copiedCmd string
	modal.OnCopied(func(cmd string) {
		copiedCmd = cmd
	})

	var startedType, startedPort string
	var startedOnlyForward bool
	modal.OnStart(func(fType, port, host, hostPort, bindAddr string, onlyForward bool, args []string) {
		startedType = fType
		startedPort = port
		startedOnlyForward = onlyForward
	})

	// 1. Initial preview check
	preview := modal.previewText.GetText(true)
	if !strings.Contains(preview, "ssh") {
		t.Fatalf("expected preview to contain ssh, got %q", preview)
	}

	// 2. Select the saved profile "K8s API"
	modal.onProfileSelected(1)
	if modal.portVal != "6443" {
		t.Errorf("expected port 6443 after selecting profile, got %q", modal.portVal)
	}
	if modal.hostVal != "127.0.0.1" {
		t.Errorf("expected host 127.0.0.1, got %q", modal.hostVal)
	}

	// 3. Save a new profile "Web Dashboard"
	modal.profileNameVal = "Web Dashboard"
	modal.portVal = "3000"
	modal.hostVal = "localhost"
	modal.hostPortVal = "3000"
	modal.saveCurrentProfile()

	if !strings.Contains(lastStatus, "saved") {
		t.Errorf("expected status 'saved', got %q", lastStatus)
	}

	saved, err := sm.LoadTunnelProfiles("bastion")
	if err != nil || len(saved) != 2 {
		t.Fatalf("expected 2 saved profiles, got %d (err: %v)", len(saved), err)
	}

	// 4. Test Copy Command
	modal.copyAndClose()
	if !strings.Contains(copiedCmd, "3000:localhost:3000") {
		t.Errorf("expected copied command to contain 3000:localhost:3000, got %q", copiedCmd)
	}

	// 5. Test Start Forwarding
	modal.startForward()
	if startedPort != "3000" || startedType != ForwardTypeLocal || !startedOnlyForward {
		t.Errorf("unexpected start forward params: type=%s, port=%s, onlyForward=%v", startedType, startedPort, startedOnlyForward)
	}

	// 6. Test Delete Profile
	modal.selectedProfileIdx = 2 // select second profile
	modal.deleteSelectedProfile()
	if !strings.Contains(lastStatus, "deleted") {
		t.Errorf("expected status 'deleted', got %q", lastStatus)
	}

	saved, _ = sm.LoadTunnelProfiles("bastion")
	if len(saved) != 1 {
		t.Fatalf("expected 1 saved profile remaining, got %d", len(saved))
	}

	// 7. Test Cancel
	canceled := false
	modal.OnCancel(func() {
		canceled = true
	})
	modal.cancel()
	if !canceled {
		t.Errorf("expected cancel callback to be invoked")
	}

	// 8. Test Dynamic SOCKS5 mode disabling host fields
	modal.currentTypeIdx = 2 // Dynamic
	modal.typeDropDown.SetCurrentOption(2)
	if modal.hostField.GetText() != "" {
		t.Errorf("expected hostField to be cleared in Dynamic mode")
	}
	if modal.hostPortField.GetText() != "" {
		t.Errorf("expected hostPortField to be cleared in Dynamic mode")
	}

	// 9. Test validation failure on invalid port
	modal.portVal = "999999"
	modal.startForward()
	if !strings.Contains(lastStatus, "Invalid port") {
		t.Errorf("expected invalid port status, got %q", lastStatus)
	}
}

func TestPortForwardModal_KeyCapture(t *testing.T) {
	app := tview.NewApplication()
	srv := domain.Server{Alias: "srv1", Host: "1.1.1.1"}
	modal := NewPortForwardModal(app, srv, nil)

	canceled := false
	modal.OnCancel(func() {
		canceled = true
	})

	copied := false
	modal.OnCopied(func(cmd string) {
		copied = true
	})

	// Test Escape -> Cancel
	handler := modal.form.InputHandler()
	escEvent := tcell.NewEventKey(tcell.KeyEscape, 0, tcell.ModNone)
	handler(escEvent, nil)
	if !canceled {
		t.Errorf("expected escape to cancel")
	}

	// Test Ctrl+S -> Copy
	ctrlSEvent := tcell.NewEventKey(tcell.KeyCtrlS, 0, tcell.ModNone)
	handler(ctrlSEvent, nil)
	if !copied {
		t.Errorf("expected Ctrl+S to copy command")
	}
}
