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

package ssh_config_file

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/WhiteRoseLK/neossh/internal/core/domain"
)

func TestSnippetManager_SeedDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "sub", "snippets.json")

	mgr := NewSnippetManager(filePath)
	snippets, err := mgr.GetSnippets()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(snippets) == 0 {
		t.Fatalf("expected seeded default snippets, got 0")
	}

	// File should exist on disk
	if _, err := os.Stat(filePath); err != nil {
		t.Errorf("expected snippets file to be created on disk: %v", err)
	}

	// Reload from new manager instance to verify disk persistence
	mgr2 := NewSnippetManager(filePath)
	reloaded, err := mgr2.GetSnippets()
	if err != nil {
		t.Fatalf("unexpected error reloading: %v", err)
	}
	if len(reloaded) != len(snippets) {
		t.Errorf("expected %d snippets, got %d", len(snippets), len(reloaded))
	}
}

func TestSnippetManager_CRUD(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "snippets.json")

	mgr := NewSnippetManager(filePath)
	initial, _ := mgr.GetSnippets()

	// 1. Add snippet
	newSnippet := domain.Snippet{
		ID:          "custom-1",
		Name:        "Custom Ping",
		Command:     "ping -c 3 {{host}}",
		Description: "Pings a host 3 times",
		Tags:        []string{"network"},
	}
	if err := mgr.SaveSnippet(newSnippet); err != nil {
		t.Fatalf("failed to save snippet: %v", err)
	}

	list, _ := mgr.GetSnippets()
	if len(list) != len(initial)+1 {
		t.Fatalf("expected count %d, got %d", len(initial)+1, len(list))
	}

	// 2. Update snippet
	newSnippet.Command = "ping -c 5 {{host}}"
	if err := mgr.SaveSnippet(newSnippet); err != nil {
		t.Fatalf("failed to update snippet: %v", err)
	}
	list, _ = mgr.GetSnippets()
	found := false
	for _, s := range list {
		if s.ID == "custom-1" {
			found = true
			if s.Command != "ping -c 5 {{host}}" {
				t.Errorf("command not updated, got %q", s.Command)
			}
			break
		}
	}
	if !found {
		t.Errorf("custom-1 not found in list")
	}

	// 3. Delete snippet
	if err := mgr.DeleteSnippet("custom-1"); err != nil {
		t.Fatalf("failed to delete snippet: %v", err)
	}
	list, _ = mgr.GetSnippets()
	for _, s := range list {
		if s.ID == "custom-1" {
			t.Errorf("custom-1 still present after delete")
		}
	}
}

func TestResolveSnippetsFilePath(t *testing.T) {
	// 1. Custom env var
	t.Setenv("NEOSSH_SNIPPETS_FILE", "/tmp/custom-snippets.json")
	p := ResolveSnippetsFilePath()
	if p != "/tmp/custom-snippets.json" {
		t.Errorf("expected custom env path, got %q", p)
	}

	// 2. XDG_CONFIG_HOME
	t.Setenv("NEOSSH_SNIPPETS_FILE", "")
	t.Setenv("XDG_CONFIG_HOME", "/custom/xdg")
	p2 := ResolveSnippetsFilePath()
	if p2 != "/custom/xdg/neossh/snippets.json" {
		t.Errorf("expected xdg path, got %q", p2)
	}
}
