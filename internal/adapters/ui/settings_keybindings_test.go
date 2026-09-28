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

func TestSettingsManager_Keybindings(t *testing.T) {
	tempDir := t.TempDir()
	mgr := &settingsManager{
		filePath: filepath.Join(tempDir, "settings.json"),
		logger:   zap.NewNop().Sugar(),
	}

	// 1. Initial load should return empty map
	kb, err := mgr.LoadKeybindings()
	if err != nil {
		t.Fatalf("expected nil error on empty load, got %v", err)
	}
	if len(kb) != 0 {
		t.Fatalf("expected empty keybindings map, got %v", kb)
	}

	// 2. Save custom keybindings
	custom := map[string]string{
		"add":   "n",
		"clone": "c",
		"copy":  "y",
		"quit":  "x",
	}
	if err := mgr.SaveKeybindings(custom); err != nil {
		t.Fatalf("failed to save keybindings: %v", err)
	}

	// 3. Reload and verify
	loaded, err := mgr.LoadKeybindings()
	if err != nil {
		t.Fatalf("failed to reload keybindings: %v", err)
	}
	if len(loaded) != len(custom) {
		t.Fatalf("expected %d entries, got %d", len(custom), len(loaded))
	}
	for k, want := range custom {
		if got := loaded[k]; got != want {
			t.Errorf("key %q = %q, want %q", k, got, want)
		}
	}
}
