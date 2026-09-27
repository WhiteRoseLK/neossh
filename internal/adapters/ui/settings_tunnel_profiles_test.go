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
	"testing"

	"go.uber.org/zap"
)

func TestSettingsManager_TunnelProfiles(t *testing.T) {
	logger := zap.NewNop().Sugar()

	t.Run("nil manager", func(t *testing.T) {
		var sm *settingsManager
		profiles, err := sm.LoadTunnelProfiles("srv1")
		if err == nil {
			t.Errorf("expected error from nil settingsManager, got nil")
		}
		if profiles != nil {
			t.Errorf("expected nil profiles, got %v", profiles)
		}

		if err := sm.SaveTunnelProfile("srv1", TunnelProfile{Name: "p1"}); err == nil {
			t.Errorf("expected error saving to nil settingsManager")
		}

		if err := sm.DeleteTunnelProfile("srv1", "p1"); err == nil {
			t.Errorf("expected error deleting from nil settingsManager")
		}
	})

	t.Run("non-existent file returns nil profiles", func(t *testing.T) {
		tmpDir := t.TempDir()
		sm := &settingsManager{
			filePath: filepath.Join(tmpDir, "settings.json"),
			logger:   logger,
		}

		profiles, err := sm.LoadTunnelProfiles("myserver")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if profiles != nil {
			t.Errorf("expected nil profiles, got %v", profiles)
		}
	})

	t.Run("save, load, update, and delete tunnel profiles", func(t *testing.T) {
		tmpDir := t.TempDir()
		sm := &settingsManager{
			filePath: filepath.Join(tmpDir, "settings.json"),
			logger:   logger,
		}

		p1 := TunnelProfile{
			Name:     "PostgreSQL",
			Type:     ForwardTypeLocal,
			Port:     "5432",
			Host:     "localhost",
			HostPort: "5432",
			Mode:     ForwardModeOnlyForward,
		}
		p2 := TunnelProfile{
			Name: "SOCKS Proxy",
			Type: ForwardTypeDynamic,
			Port: "1080",
			Mode: ForwardModeOnlyForward,
		}

		// Save p1 for host1
		if err := sm.SaveTunnelProfile("host1", p1); err != nil {
			t.Fatalf("failed to save p1: %v", err)
		}

		// Save p2 for host1
		if err := sm.SaveTunnelProfile("host1", p2); err != nil {
			t.Fatalf("failed to save p2: %v", err)
		}

		// Load profiles for host1
		profiles, err := sm.LoadTunnelProfiles("host1")
		if err != nil {
			t.Fatalf("failed to load profiles: %v", err)
		}
		if len(profiles) != 2 {
			t.Fatalf("expected 2 profiles, got %d", len(profiles))
		}
		if profiles[0].Name != "PostgreSQL" || profiles[1].Name != "SOCKS Proxy" {
			t.Errorf("unexpected profiles: %+v", profiles)
		}

		// Update p1 with modified port
		p1Updated := p1
		p1Updated.Port = "5433"
		if err := sm.SaveTunnelProfile("host1", p1Updated); err != nil {
			t.Fatalf("failed to update p1: %v", err)
		}

		profiles, err = sm.LoadTunnelProfiles("host1")
		if err != nil {
			t.Fatalf("failed to load profiles after update: %v", err)
		}
		if len(profiles) != 2 {
			t.Fatalf("expected 2 profiles, got %d", len(profiles))
		}
		if profiles[0].Port != "5433" {
			t.Errorf("expected port 5433, got %s", profiles[0].Port)
		}

		// Delete p1
		if err := sm.DeleteTunnelProfile("host1", "PostgreSQL"); err != nil {
			t.Fatalf("failed to delete p1: %v", err)
		}

		profiles, err = sm.LoadTunnelProfiles("host1")
		if err != nil {
			t.Fatalf("failed to load profiles after delete: %v", err)
		}
		if len(profiles) != 1 || profiles[0].Name != "SOCKS Proxy" {
			t.Fatalf("expected 1 profile 'SOCKS Proxy', got %+v", profiles)
		}

		// Delete p2
		if err := sm.DeleteTunnelProfile("host1", "SOCKS Proxy"); err != nil {
			t.Fatalf("failed to delete p2: %v", err)
		}

		profiles, err = sm.LoadTunnelProfiles("host1")
		if err != nil {
			t.Fatalf("failed to load profiles after deleting all: %v", err)
		}
		if len(profiles) != 0 {
			t.Errorf("expected 0 profiles, got %d", len(profiles))
		}
	})
}
