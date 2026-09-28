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
	"testing"

	"github.com/WhiteRoseLK/neossh/internal/core/domain"
	"github.com/rivo/tview"
)

type mockKnownHostsService struct {
	mockReadOnlyService
	records     []domain.KnownHostRecord
	removedLine int
}

func (m *mockKnownHostsService) ListKnownHostRecords(string) ([]domain.KnownHostRecord, error) {
	return m.records, nil
}

func (m *mockKnownHostsService) RemoveKnownHostByLine(_ string, line int) (string, error) {
	m.removedLine = line
	var updated []domain.KnownHostRecord
	for _, r := range m.records {
		if r.LineNumber != line {
			updated = append(updated, r)
		}
	}
	m.records = updated
	return "known_hosts.old", nil
}

func TestKnownHostsModal(t *testing.T) {
	app := tview.NewApplication()
	records := []domain.KnownHostRecord{
		{
			LineNumber:  1,
			HostPattern: "github.com",
			KeyType:     "ssh-ed25519",
			Fingerprint: "SHA256:abcd1234",
			Comment:     "GitHub key",
		},
		{
			LineNumber:  2,
			HostPattern: "192.168.1.100",
			KeyType:     "ecdsa-sha2-nistp256",
			Fingerprint: "SHA256:efgh5678",
			Comment:     "Local node",
		},
		{
			LineNumber:  3,
			HostPattern: "|1|abcd...",
			KeyType:     "ssh-rsa",
			Fingerprint: "SHA256:ijkl9012",
			IsHashed:    true,
		},
	}

	mockSvc := &mockKnownHostsService{
		records: records,
	}

	var statusMsg string
	onStatus := func(msg, color string) {
		statusMsg = msg
	}
	var closed bool
	onClose := func() {
		closed = true
	}

	modal := NewKnownHostsModal(app, mockSvc, onClose, onStatus)
	if modal == nil {
		t.Fatal("expected non-nil modal")
	}

	// Verify table populated with 3 entries (+ 1 header row = 4 rows)
	if rows := modal.table.GetRowCount(); rows != 4 {
		t.Errorf("expected 4 rows, got %d", rows)
	}

	// Test filtering
	modal.filterInput.SetText("github")
	if len(modal.filtered) != 1 {
		t.Errorf("expected 1 filtered item, got %d", len(modal.filtered))
	}
	if modal.filtered[0].HostPattern != "github.com" {
		t.Errorf("expected github.com, got %q", modal.filtered[0].HostPattern)
	}

	// Clear filter
	modal.filterInput.SetText("")
	if len(modal.filtered) != 3 {
		t.Errorf("expected 3 items after clearing filter, got %d", len(modal.filtered))
	}

	// Test close callback
	modal.onClose()
	if !closed {
		t.Error("expected onClose to be called")
	}
	_ = statusMsg
}
