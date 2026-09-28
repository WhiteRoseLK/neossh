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
	"strings"
	"testing"

	"github.com/WhiteRoseLK/neossh/internal/core/domain"
	"github.com/rivo/tview"
)

func TestSSHAgentStatusBarBadge(t *testing.T) {
	app := &tui{
		app:       tview.NewApplication(),
		statusBar: tview.NewTextView().SetDynamicColors(true),
	}

	// 1. Available with keys
	app.agentStatus = domain.SSHAgentStatus{
		Available: true,
		Type:      domain.AgentTypeOpenSSH,
		KeyCount:  3,
	}
	badge := app.renderAgentStatusBadge()
	if !strings.Contains(badge, "3 key(s)") || !strings.Contains(badge, "OpenSSH") {
		t.Errorf("badge mismatch for 3 keys: %q", badge)
	}

	statusText := app.defaultStatusTextLocked()
	if !strings.Contains(statusText, "3 key(s)") {
		t.Errorf("expected defaultStatusTextLocked to include badge, got %q", statusText)
	}

	// 2. Available with 0 keys
	app.agentStatus = domain.SSHAgentStatus{
		Available: true,
		Type:      domain.AgentType1Password,
		KeyCount:  0,
	}
	badge = app.renderAgentStatusBadge()
	if !strings.Contains(badge, "0 keys") || !strings.Contains(badge, "1Password") {
		t.Errorf("badge mismatch for 0 keys: %q", badge)
	}

	// 3. Inactive
	app.agentStatus = domain.SSHAgentStatus{
		Available: false,
	}
	badge = app.renderAgentStatusBadge()
	if !strings.Contains(badge, "inactive") {
		t.Errorf("badge mismatch for inactive: %q", badge)
	}
}
