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
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/WhiteRoseLK/neossh/internal/core/domain"
)

// SnippetManager handles thread-safe persistence of snippet catalogs to disk.
type SnippetManager struct {
	mu       sync.RWMutex
	filePath string
	cached   []domain.Snippet
	loaded   bool
}

// ResolveSnippetsFilePath returns the appropriate path for snippets.json.
// Honors NEOSSH_SNIPPETS_FILE, XDG_CONFIG_HOME, and standard ~/.config/neossh.
func ResolveSnippetsFilePath() string {
	if custom := os.Getenv("NEOSSH_SNIPPETS_FILE"); custom != "" {
		return filepath.Clean(custom)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}

	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "neossh", "snippets.json")
	}

	dotConfig := filepath.Join(home, ".config", "neossh", "snippets.json")
	dotNeossh := filepath.Join(home, ".neossh", "snippets.json")

	// If dotConfig already exists, use it
	if _, err := os.Stat(dotConfig); err == nil {
		return dotConfig
	}
	// If dotNeossh already exists, use it
	if _, err := os.Stat(dotNeossh); err == nil {
		return dotNeossh
	}

	// Default to ~/.config/neossh/snippets.json
	return dotConfig
}

// NewSnippetManager creates a new SnippetManager instance for the given file path.
func NewSnippetManager(filePath string) *SnippetManager {
	if filePath == "" {
		filePath = ResolveSnippetsFilePath()
	}
	return &SnippetManager{
		filePath: filepath.Clean(filePath),
	}
}

// GetSnippets loads and returns all snippets from the catalog.
// If the file does not exist, it seeds it with standard default snippets.
func (m *SnippetManager) GetSnippets() ([]domain.Snippet, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.loaded {
		cp := make([]domain.Snippet, len(m.cached))
		copy(cp, m.cached)
		return cp, nil
	}

	// Read file from disk
	cleanPath := filepath.Clean(m.filePath)
	data, err := os.ReadFile(cleanPath) // #nosec G304: reading user snippets config file
	if err != nil {
		if os.IsNotExist(err) {
			defaults := domain.DefaultSnippets()
			m.cached = defaults
			m.loaded = true
			_ = m.saveLocked(defaults)
			cp := make([]domain.Snippet, len(defaults))
			copy(cp, defaults)
			return cp, nil
		}
		return nil, fmt.Errorf("read snippets file %s: %w", m.filePath, err)
	}

	var snippets []domain.Snippet
	if err := json.Unmarshal(data, &snippets); err != nil {
		return nil, fmt.Errorf("parse snippets json: %w", err)
	}

	if len(snippets) == 0 {
		snippets = domain.DefaultSnippets()
		_ = m.saveLocked(snippets)
	}

	m.cached = snippets
	m.loaded = true
	cp := make([]domain.Snippet, len(snippets))
	copy(cp, snippets)
	return cp, nil
}

// SaveSnippet adds or updates a snippet in the catalog.
func (m *SnippetManager) SaveSnippet(s domain.Snippet) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.loaded {
		m.mu.Unlock()
		_, _ = m.GetSnippets()
		m.mu.Lock()
	}

	if s.ID == "" {
		sanitizedName := strings.ToLower(strings.ReplaceAll(s.Name, " ", "-"))
		s.ID = fmt.Sprintf("%s-%d", sanitizedName, time.Now().UnixNano()%100000)
	}

	found := false
	for i, existing := range m.cached {
		if existing.ID == s.ID {
			m.cached[i] = s
			found = true
			break
		}
	}
	if !found {
		m.cached = append(m.cached, s)
	}

	return m.saveLocked(m.cached)
}

// DeleteSnippet removes a snippet by ID.
func (m *SnippetManager) DeleteSnippet(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.loaded {
		m.mu.Unlock()
		_, _ = m.GetSnippets()
		m.mu.Lock()
	}

	var updated []domain.Snippet
	for _, s := range m.cached {
		if s.ID != id {
			updated = append(updated, s)
		}
	}
	m.cached = updated
	return m.saveLocked(m.cached)
}

func (m *SnippetManager) saveLocked(snippets []domain.Snippet) error {
	dir := filepath.Dir(m.filePath)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("create snippet dir: %w", err)
	}

	data, err := json.MarshalIndent(snippets, "", "  ")
	if err != nil {
		return fmt.Errorf("serialize snippets: %w", err)
	}

	tmpFile, err := os.CreateTemp(dir, "snippets-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp snippet file: %w", err)
	}
	tmpName := tmpFile.Name()

	if _, err := tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		_ = os.Remove(tmpName)
		return fmt.Errorf("write temp snippet file: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("close temp snippet file: %w", err)
	}

	if err := os.Chmod(tmpName, 0o600); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("chmod snippet file: %w", err)
	}

	if err := os.Rename(tmpName, m.filePath); err != nil {
		_ = os.Remove(tmpName)
		return fmt.Errorf("atomic rename snippet file: %w", err)
	}

	return nil
}
