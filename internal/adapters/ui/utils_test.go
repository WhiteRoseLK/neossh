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
	"testing"

	"github.com/WhiteRoseLK/neossh/internal/core/domain"
	"github.com/rivo/tview"
)

func TestBuildSSHCommand_PortForwarding(t *testing.T) {
	tests := []struct {
		name     string
		server   domain.Server
		expected []string // expected parts in the command
	}{
		{
			name: "local forward",
			server: domain.Server{
				Alias:        "test",
				Host:         "example.com",
				User:         "user",
				LocalForward: []string{"8080:localhost:80", "3306:db.internal:3306"},
			},
			expected: []string{"ssh", "-L", "8080:localhost:80", "-L", "3306:db.internal:3306", "user@example.com"},
		},
		{
			name: "remote forward",
			server: domain.Server{
				Alias:         "test",
				Host:          "example.com",
				User:          "user",
				RemoteForward: []string{"8080:localhost:3000", "*:80:localhost:8080"},
			},
			expected: []string{"ssh", "-R", "8080:localhost:3000", "-R", "*:80:localhost:8080", "user@example.com"},
		},
		{
			name: "dynamic forward",
			server: domain.Server{
				Alias:          "test",
				Host:           "example.com",
				User:           "user",
				DynamicForward: []string{"1080", "localhost:1081"},
			},
			expected: []string{"ssh", "-D", "1080", "-D", "localhost:1081", "user@example.com"},
		},
		{
			name: "all forward types",
			server: domain.Server{
				Alias:          "test",
				Host:           "example.com",
				User:           "user",
				LocalForward:   []string{"8080:localhost:80"},
				RemoteForward:  []string{"9090:localhost:9090"},
				DynamicForward: []string{"1080"},
			},
			expected: []string{"ssh", "-L", "8080:localhost:80", "-R", "9090:localhost:9090", "-D", "1080", "user@example.com"},
		},
		{
			name: "forward with bind address",
			server: domain.Server{
				Alias:        "test",
				Host:         "example.com",
				User:         "user",
				LocalForward: []string{"127.0.0.1:8080:localhost:80", "*:3000:localhost:3000"},
			},
			expected: []string{"ssh", "-L", "127.0.0.1:8080:localhost:80", "-L", "*:3000:localhost:3000", "user@example.com"},
		},
		{
			name: "forward with additional options",
			server: domain.Server{
				Alias:                "test",
				Host:                 "example.com",
				User:                 "user",
				LocalForward:         []string{"8080:localhost:80"},
				ExitOnForwardFailure: "yes",
				GatewayPorts:         "clientspecified",
			},
			expected: []string{"ssh", "-L", "8080:localhost:80", "-o", "ExitOnForwardFailure=yes", "-o", "GatewayPorts=clientspecified", "user@example.com"},
		},
		{
			name: "clear all forwardings",
			server: domain.Server{
				Alias:               "test",
				Host:                "example.com",
				User:                "user",
				LocalForward:        []string{"8080:localhost:80"},
				ClearAllForwardings: "yes",
			},
			expected: []string{"ssh", "-L", "8080:localhost:80", "-o", "ClearAllForwardings=yes", "user@example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildSSHCommand(tt.server)

			// Check that all expected parts are in the result
			for _, part := range tt.expected {
				if !strings.Contains(result, part) {
					t.Errorf("BuildSSHCommand() missing expected part %q in result: %q", part, result)
				}
			}

			// Additional check: ensure the command contains "ssh" command
			// Now it includes alias comment, so check for "\nssh " or just "ssh " at the beginning
			if !strings.Contains(result, "\nssh ") && !strings.HasPrefix(result, "ssh ") {
				t.Errorf("BuildSSHCommand() should contain 'ssh ' command, got: %q", result)
			}
		})
	}
}

func TestBuildSSHCommand_CompleteCommand(t *testing.T) {
	server := domain.Server{
		Alias:          "myserver",
		Host:           "example.com",
		User:           "admin",
		Port:           2222,
		LocalForward:   []string{"8080:localhost:80", "3306:db.internal:3306"},
		RemoteForward:  []string{"9090:localhost:9090"},
		DynamicForward: []string{"1080"},
		IdentityFiles:  []string{"~/.ssh/id_rsa"},
	}

	result := BuildSSHCommand(server)

	// Check command structure
	if !strings.HasPrefix(result, "# neossh-alias:myserver") {
		t.Errorf("Command should start with alias comment, got: %q", result)
	}

	if !strings.Contains(result, "\nssh ") {
		t.Errorf("Command should contain 'ssh ' after alias comment, got: %q", result)
	}

	// Check port
	if !strings.Contains(result, "-p 2222") {
		t.Errorf("Command should contain port flag '-p 2222', got: %q", result)
	}

	// Check identity file
	if !strings.Contains(result, "-i ~/.ssh/id_rsa") {
		t.Errorf("Command should contain identity file flag, got: %q", result)
	}

	// Check all forwards
	expectedForwards := []string{
		"-L 8080:localhost:80",
		"-L 3306:db.internal:3306",
		"-R 9090:localhost:9090",
		"-D 1080",
	}

	for _, forward := range expectedForwards {
		if !strings.Contains(result, forward) {
			t.Errorf("Command should contain forward %q, got: %q", forward, result)
		}
	}

	// Check user@host
	if !strings.Contains(result, "admin@example.com") {
		t.Errorf("Command should contain 'admin@example.com', got: %q", result)
	}
}

func TestFormatServerLine_PingIndicators(t *testing.T) {
	srvUp := domain.Server{
		Alias:       "prod-srv",
		Host:        "192.168.1.10",
		PingStatus:  StatusUp,
		PingLatency: 45 * 1000 * 1000, // 45ms
	}

	primary, _ := formatServerLine(srvUp, 10, 100)
	if !strings.Contains(primary, "● 45ms") {
		t.Errorf("expected latency badge '● 45ms' in primary line, got: %q", primary)
	}

	srvDown := domain.Server{
		Alias:      "down-srv",
		Host:       "192.168.1.20",
		PingStatus: StatusDown,
	}

	primaryDown, _ := formatServerLine(srvDown, 10, 100)
	if !strings.Contains(primaryDown, "● DOWN") {
		t.Errorf("expected '● DOWN' in primary line, got: %q", primaryDown)
	}

	srvChecking := domain.Server{
		Alias:      "check-srv",
		Host:       "192.168.1.30",
		PingStatus: StatusChecking,
	}

	primaryCheck, _ := formatServerLine(srvChecking, 10, 100)
	if !strings.Contains(primaryCheck, "● ... ") {
		t.Errorf("expected '● ... ' in primary line, got: %q", primaryCheck)
	}
}

func TestStripSimpleColors(t *testing.T) {
	colored := "[white::b]hello[-] [#AAAAAA]world[-]"
	stripped := stripSimpleColors(colored)
	if stripped != "hello world" {
		t.Errorf("expected 'hello world', got: %q", stripped)
	}
}

func TestTaggedStringWidth_Checkbox(t *testing.T) {
	checked := fmt.Sprintf("[%s::b]%s[-] ", "green", tview.Escape("[✓]"))
	unchecked := fmt.Sprintf("[%s]%s[-] ", "gray", tview.Escape("[ ]"))
	t.Logf("checked: %q width: %d, stripped: %q len: %d", checked, tview.TaggedStringWidth(checked), stripSimpleColors(checked), len(stripSimpleColors(checked)))
	t.Logf("unchecked: %q width: %d, stripped: %q len: %d", unchecked, tview.TaggedStringWidth(unchecked), stripSimpleColors(unchecked), len(stripSimpleColors(unchecked)))
	if tview.TaggedStringWidth(checked) != 4 {
		t.Errorf("expected checked width 4, got %d", tview.TaggedStringWidth(checked))
	}
	if tview.TaggedStringWidth(unchecked) != 4 {
		t.Errorf("expected unchecked width 4, got %d", tview.TaggedStringWidth(unchecked))
	}
}
