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

type mockAgentModalService struct {
	mockReadOnlyService
	status domain.SSHAgentStatus
}

func (m *mockAgentModalService) GetSSHAgentStatus() domain.SSHAgentStatus {
	return m.status
}

func TestSSHAgentModal(t *testing.T) {
	app := tview.NewApplication()
	mockSvc := &mockAgentModalService{
		status: domain.SSHAgentStatus{
			Available:  true,
			SocketPath: "/tmp/ssh-test/agent.sock",
			Type:       domain.AgentType1Password,
			KeyCount:   2,
			Keys: []domain.AgentKeyRecord{
				{
					Format:      "ssh-ed25519",
					Fingerprint: "SHA256:11111111",
					Comment:     "id_ed25519",
				},
				{
					Format:      "rsa-sha2-512",
					Fingerprint: "SHA256:22222222",
					Comment:     "work_rsa",
				},
			},
		},
	}

	var closed bool
	onClose := func() {
		closed = true
	}
	var statusMsg string
	onStatus := func(msg, color string) {
		statusMsg = msg
	}

	modal := NewSSHAgentModal(app, mockSvc, onClose, onStatus)
	if modal == nil {
		t.Fatal("expected non-nil modal")
	}

	// Verify table has 2 entries + 1 header = 3 rows
	if rows := modal.table.GetRowCount(); rows != 3 {
		t.Errorf("expected 3 rows, got %d", rows)
	}

	headerText := modal.headerText.GetText(false)
	if !strings.Contains(headerText, "Active") {
		t.Errorf("expected header text to contain 'Active', got %q", headerText)
	}
	if !strings.Contains(headerText, "1Password") {
		t.Errorf("expected header text to contain '1Password', got %q", headerText)
	}

	// Test close callback
	modal.onClose()
	if !closed {
		t.Error("expected onClose to be called")
	}
	_ = statusMsg
}
