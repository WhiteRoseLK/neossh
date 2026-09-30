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
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/WhiteRoseLK/neossh/internal/core/domain"
	"github.com/WhiteRoseLK/neossh/internal/core/ports"
	"github.com/kevinburke/ssh_config"
	"go.uber.org/zap"
)

// Repository implements ServerRepository interface for SSH config file operations.
type Repository struct {
	configPath      string
	fileSystem      FileSystem
	metadataManager *metadataManager
	snippetManager  *SnippetManager
	logger          *zap.SugaredLogger
}

// NewRepository creates a new SSH config repository.
func NewRepository(logger *zap.SugaredLogger, configPath, metaDataPath string) ports.ServerRepository {
	snippetPath := filepath.Join(filepath.Dir(metaDataPath), "snippets.json")
	if custom := os.Getenv("NEOSSH_SNIPPETS_FILE"); custom != "" {
		snippetPath = custom
	}
	return &Repository{
		logger:          logger,
		configPath:      configPath,
		fileSystem:      DefaultFileSystem{},
		metadataManager: newMetadataManager(metaDataPath, logger),
		snippetManager:  NewSnippetManager(snippetPath),
	}
}

// NewRepositoryWithFS creates a new SSH config repository with a custom filesystem.
func NewRepositoryWithFS(logger *zap.SugaredLogger, configPath string, metaDataPath string, fs FileSystem) ports.ServerRepository {
	snippetPath := filepath.Join(filepath.Dir(metaDataPath), "snippets.json")
	if custom := os.Getenv("NEOSSH_SNIPPETS_FILE"); custom != "" {
		snippetPath = custom
	}
	return &Repository{
		logger:          logger,
		configPath:      configPath,
		fileSystem:      fs,
		metadataManager: newMetadataManager(metaDataPath, logger),
		snippetManager:  NewSnippetManager(snippetPath),
	}
}

// ListServers returns all servers matching the query pattern.
// Empty query returns all servers.
func (r *Repository) ListServers(query string) ([]domain.Server, error) {
	lc, err := r.loadConfig()
	if err != nil {
		return nil, err
	}

	servers := r.toDomainServer(lc)
	metadata, err := r.metadataManager.loadAll()
	if err != nil {
		r.logger.Warnf("Failed to load metadata: %v", err)
		metadata = make(map[string]ServerMetadata)
	}
	servers = r.mergeMetadata(servers, metadata)
	if query == "" {
		return servers, nil
	}

	return r.filterServers(servers, query), nil
}

// AddServer adds a new server to the SSH config. If server.SourceFile is set
// and matches a loaded file, the new host is written there; otherwise it
// goes into the main config file.
func (r *Repository) AddServer(server domain.Server) error {
	lc, err := r.loadConfig()
	if err != nil {
		return err
	}

	if r.serverExists(lc, server.Alias) {
		return fmt.Errorf("server with alias '%s' already exists", server.Alias)
	}
	for _, alias := range server.Aliases {
		if alias != server.Alias && r.serverExists(lc, alias) {
			return fmt.Errorf("server with alias '%s' already exists", alias)
		}
	}

	target := lc.findFile(server.SourceFile)
	if target == nil {
		// Default: main file.
		target = &lc.files[0]
	}

	host := r.createHostFromServer(server)
	target.cfg.Hosts = append(target.cfg.Hosts, host)

	if err := r.saveFiles(lc, []string{target.path}); err != nil {
		r.logger.Warnf("Failed to save config while adding new server: %v", err)
		return fmt.Errorf("failed to save config: %w", err)
	}
	return r.metadataManager.updateServer(server, server.Alias)
}

// UpdateServer updates an existing server in the SSH config. The host is
// mutated in whichever file currently defines it (preferring server.SourceFile
// when the alias is defined in multiple files).
func (r *Repository) UpdateServer(server domain.Server, newServer domain.Server) error {
	lc, err := r.loadConfig()
	if err != nil {
		return err
	}

	matches := r.findHostMatches(lc, server.Alias)
	if len(matches) == 0 {
		for _, a := range server.Aliases {
			matches = r.findHostMatches(lc, a)
			if len(matches) > 0 {
				break
			}
		}
	}
	if len(matches) == 0 {
		return fmt.Errorf("server with alias '%s' not found", server.Alias)
	}
	if len(matches) > 1 && !preferenceResolves(matches, server.SourceFile) {
		return &domain.ErrAmbiguousHost{Alias: server.Alias, Candidates: matchPaths(matches)}
	}

	picked := pickWritableMatch(matches, server.SourceFile)
	host := picked.host

	if server.Alias != newServer.Alias {
		newMatches := r.findHostMatches(lc, newServer.Alias)
		for _, m := range newMatches {
			if m.host != host {
				return fmt.Errorf("server with alias '%s' already exists", newServer.Alias)
			}
		}
	}
	for _, a := range newServer.Aliases {
		if a == server.Alias || a == newServer.Alias {
			continue
		}
		newMatches := r.findHostMatches(lc, a)
		for _, m := range newMatches {
			if m.host != host {
				return fmt.Errorf("server with alias '%s' already exists", a)
			}
		}
	}

	if len(newServer.Aliases) > 0 {
		aliases := make([]string, 0, len(newServer.Aliases)+1)
		seen := make(map[string]bool)
		cleanPrimary := strings.Trim(strings.TrimSpace(newServer.Alias), "\"'")
		if cleanPrimary != "" {
			aliases = append(aliases, cleanPrimary)
			seen[cleanPrimary] = true
		}
		for _, a := range newServer.Aliases {
			cleanA := strings.Trim(strings.TrimSpace(a), "\"'")
			if cleanA != "" && !seen[cleanA] {
				aliases = append(aliases, cleanA)
				seen[cleanA] = true
			}
		}
		newPatterns := make([]*ssh_config.Pattern, 0, len(aliases))
		for _, a := range aliases {
			newPatterns = append(newPatterns, &ssh_config.Pattern{Str: a})
		}
		host.Patterns = newPatterns
	} else if server.Alias != newServer.Alias {
		cleanOld := strings.Trim(strings.TrimSpace(server.Alias), "\"'")
		cleanNew := strings.Trim(strings.TrimSpace(newServer.Alias), "\"'")
		newPatterns := make([]*ssh_config.Pattern, 0, len(host.Patterns))
		for _, pattern := range host.Patterns {
			if strings.Trim(strings.TrimSpace(pattern.Str), "\"'") == cleanOld {
				newPatterns = append(newPatterns, &ssh_config.Pattern{Str: cleanNew})
			} else {
				newPatterns = append(newPatterns, pattern)
			}
		}
		host.Patterns = newPatterns
	}

	r.updateHostNodes(host, server, newServer)

	if err := r.saveFiles(lc, []string{picked.path}); err != nil {
		r.logger.Warnf("Failed to save config while updating server: %v", err)
		return fmt.Errorf("failed to save config: %w", err)
	}
	if err := r.metadataManager.updateServer(newServer, server.Alias); err != nil {
		return err
	}
	return r.metadataManager.setFile(newServer.Alias, picked.path)
}

// DeleteServer removes a server from the SSH config (from whichever file
// currently defines it; preferring server.SourceFile on ambiguity).
func (r *Repository) DeleteServer(server domain.Server) error {
	lc, err := r.loadConfig()
	if err != nil {
		return err
	}

	matches := r.findHostMatches(lc, server.Alias)
	if len(matches) == 0 {
		for _, a := range server.Aliases {
			matches = r.findHostMatches(lc, a)
			if len(matches) > 0 {
				break
			}
		}
	}
	if len(matches) == 0 {
		return fmt.Errorf("server with alias '%s' not found", server.Alias)
	}
	if len(matches) > 1 && !preferenceResolves(matches, server.SourceFile) {
		return &domain.ErrAmbiguousHost{Alias: server.Alias, Candidates: matchPaths(matches)}
	}

	picked := pickWritableMatch(matches, server.SourceFile)
	lookupAlias := server.Alias
	if !r.hostContainsPattern(picked.host, lookupAlias) {
		for _, a := range server.Aliases {
			if r.hostContainsPattern(picked.host, a) {
				lookupAlias = a
				break
			}
		}
	}
	picked.cfg.Hosts = r.removeHostByAlias(picked.cfg.Hosts, lookupAlias)

	if err := r.saveFiles(lc, []string{picked.path}); err != nil {
		r.logger.Warnf("Failed to save config while deleting server: %v", err)
		return fmt.Errorf("failed to save config: %w", err)
	}
	return r.metadataManager.deleteServer(server.Alias)
}

// SetPinned sets or unsets the pinned status of a server.
func (r *Repository) SetPinned(alias string, pinned bool) error {
	return r.metadataManager.setPinned(alias, pinned)
}

// SetHidden sets or unsets the hidden status of a server.
func (r *Repository) SetHidden(alias string, hidden bool) error {
	return r.metadataManager.setHidden(alias, hidden)
}

// RecordSSH increments the SSH access count and updates the last seen timestamp for a server.
func (r *Repository) RecordSSH(alias string) error {
	return r.metadataManager.recordSSH(alias)
}

// GetConfigFile gets the path to the ssh config file.
func (r *Repository) GetConfigFile() string {
	return r.configPath
}

// GetConfigFiles returns the paths of the main SSH config file and all resolved Include files.
func (r *Repository) GetConfigFiles() ([]string, error) {
	lc, err := r.loadConfig()
	if err != nil {
		return nil, err
	}
	return lc.paths(), nil
}

// GetMetadataFile gets the path to the metadata file.
func (r *Repository) GetMetadataFile() string {
	if r.metadataManager == nil {
		return ""
	}
	return r.metadataManager.filePath
}

// GetSettings returns the application settings.
func (r *Repository) GetSettings() (Settings, error) {
	return r.metadataManager.GetSettings()
}

// SaveSettings saves the application settings.
func (r *Repository) SaveSettings(settings Settings) error {
	return r.metadataManager.SaveSettings(settings)
}

// GetTheme returns the current theme name from settings.
func (r *Repository) GetTheme() (string, error) {
	settings, err := r.metadataManager.GetSettings()
	if err != nil {
		return "", err
	}
	return settings.Theme, nil
}

// SaveTheme saves the theme name to settings.
func (r *Repository) SaveTheme(theme string) error {
	settings, err := r.metadataManager.GetSettings()
	if err != nil {
		settings = Settings{}
	}
	settings.Theme = theme
	return r.metadataManager.SaveSettings(settings)
}

// GetPreConnectCommand returns the global pre-connect command from settings.
func (r *Repository) GetPreConnectCommand() (string, error) {
	settings, err := r.metadataManager.GetSettings()
	if err != nil {
		return "", err
	}
	return settings.PreConnectCommand, nil
}

// SavePreConnectCommand saves the global pre-connect command to settings.
func (r *Repository) SavePreConnectCommand(cmd string) error {
	settings, err := r.metadataManager.GetSettings()
	if err != nil {
		settings = Settings{}
	}
	settings.PreConnectCommand = cmd
	return r.metadataManager.SaveSettings(settings)
}

// GetDefaultIdentityKey returns the default identity SSH key from settings.
func (r *Repository) GetDefaultIdentityKey() (string, error) {
	settings, err := r.metadataManager.GetSettings()
	if err != nil {
		return "", err
	}
	return settings.DefaultIdentityKey, nil
}

// SaveDefaultIdentityKey saves the default identity SSH key to settings.
func (r *Repository) SaveDefaultIdentityKey(key string) error {
	settings, err := r.metadataManager.GetSettings()
	if err != nil {
		settings = Settings{}
	}
	settings.DefaultIdentityKey = key
	return r.metadataManager.SaveSettings(settings)
}

// GetFileManager returns the configured file manager command or tool name from settings.
func (r *Repository) GetFileManager() (string, error) {
	settings, err := r.metadataManager.GetSettings()
	if err != nil {
		return "", err
	}
	return settings.FileManager, nil
}

// SaveFileManager saves the file manager command or tool name to settings.
func (r *Repository) SaveFileManager(tool string) error {
	settings, err := r.metadataManager.GetSettings()
	if err != nil {
		settings = Settings{}
	}
	settings.FileManager = tool
	return r.metadataManager.SaveSettings(settings)
}

// GetFirstRunCompleted returns whether onboarding setup has been completed.
func (r *Repository) GetFirstRunCompleted() (bool, error) {
	return r.metadataManager.GetFirstRunCompleted()
}

// SaveFirstRunCompleted saves the onboarding setup completion flag.
func (r *Repository) SaveFirstRunCompleted(completed bool) error {
	return r.metadataManager.SaveFirstRunCompleted(completed)
}

func (r *Repository) GetSnippets() ([]domain.Snippet, error) {
	if r.snippetManager == nil {
		r.snippetManager = NewSnippetManager("")
	}
	return r.snippetManager.GetSnippets()
}

func (r *Repository) SaveSnippet(snippet domain.Snippet) error {
	if r.snippetManager == nil {
		r.snippetManager = NewSnippetManager("")
	}
	return r.snippetManager.SaveSnippet(snippet)
}

func (r *Repository) DeleteSnippet(id string) error {
	if r.snippetManager == nil {
		r.snippetManager = NewSnippetManager("")
	}
	return r.snippetManager.DeleteSnippet(id)
}

// LoadSettings loads application settings from the metadata file at the given path.
// This is a standalone function for use during app initialization before the repository is created.
func LoadSettings(metaDataPath string) (Settings, error) {
	mm := newMetadataManager(metaDataPath, nil)
	return mm.GetSettings()
}
