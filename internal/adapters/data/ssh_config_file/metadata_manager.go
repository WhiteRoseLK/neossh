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
	"time"

	"github.com/WhiteRoseLK/neossh/internal/core/domain"
	"go.uber.org/zap"
)

// Settings contains application-level settings stored in the metadata file.
type Settings struct {
	Theme              string `json:"theme,omitempty"`
	PreConnectCommand  string `json:"pre_connect_command,omitempty"`
	DefaultIdentityKey string `json:"default_identity_key,omitempty"`
	FileManager        string `json:"file_manager,omitempty"`
}

type ServerMetadata struct {
	Tags                  []string `json:"tags,omitempty"`
	Group                 string   `json:"group,omitempty"`
	LastSeen              string   `json:"last_seen,omitempty"`
	PinnedAt              string   `json:"pinned_at,omitempty"`
	Hidden                bool     `json:"hidden,omitempty"`
	SSHCount              int      `json:"ssh_count,omitempty"`
	PreConnectCommand     string   `json:"pre_connect_command,omitempty"`
	CertificateCommand    string   `json:"certificate_command,omitempty"`
	SyncDotfilesOnConnect bool     `json:"sync_dotfiles_on_connect,omitempty"`
	// File is the absolute path of the SSH config file neossh should
	// write to when editing or deleting this host. Populated lazily on
	// the first successful write and used to suppress the ambiguity
	// prompt on subsequent edits.
	File string `json:"file,omitempty"`
}

// MetadataFile is the top-level structure of the metadata JSON file.
type MetadataFile struct {
	Settings Settings                  `json:"settings,omitempty"`
	Servers  map[string]ServerMetadata `json:"servers,omitempty"`
}

type metadataManager struct {
	filePath string
	logger   *zap.SugaredLogger
}

func newMetadataManager(filePath string, logger *zap.SugaredLogger) *metadataManager {
	return &metadataManager{filePath: filePath, logger: logger}
}

func (m *metadataManager) loadFile() (*MetadataFile, error) {
	result := &MetadataFile{
		Servers: make(map[string]ServerMetadata),
	}

	if _, err := os.Stat(m.filePath); os.IsNotExist(err) {
		return result, nil
	}

	data, err := os.ReadFile(m.filePath)
	if err != nil {
		return nil, fmt.Errorf("read metadata '%s': %w", m.filePath, err)
	}

	if len(data) == 0 {
		return result, nil
	}

	if err := json.Unmarshal(data, result); err != nil {
		return nil, fmt.Errorf("parse metadata JSON '%s': %w", m.filePath, err)
	}

	if len(result.Servers) == 0 {
		var oldFormat map[string]ServerMetadata
		if err := json.Unmarshal(data, &oldFormat); err == nil && len(oldFormat) > 0 {
			for _, v := range oldFormat {
				if len(v.Tags) > 0 || v.LastSeen != "" || v.PinnedAt != "" || v.Hidden || v.SSHCount > 0 || v.File != "" {
					result.Servers = oldFormat
					break
				}
			}
		}
	}

	if result.Servers == nil {
		result.Servers = make(map[string]ServerMetadata)
	}

	return result, nil
}

func (m *metadataManager) saveFile(file *MetadataFile) error {
	if err := m.ensureDirectory(); err != nil {
		if m.logger != nil {
			m.logger.Errorw("failed to ensure metadata directory", "path", m.filePath, "error", err)
		}
		return fmt.Errorf("ensure metadata directory for '%s': %w", m.filePath, err)
	}

	data, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		if m.logger != nil {
			m.logger.Errorw("failed to marshal metadata", "path", m.filePath, "error", err)
		}
		return fmt.Errorf("marshal metadata for '%s': %w", m.filePath, err)
	}

	if err := os.WriteFile(m.filePath, data, 0o600); err != nil {
		if m.logger != nil {
			m.logger.Errorw("failed to write metadata file", "path", m.filePath, "error", err)
		}
		return fmt.Errorf("write metadata '%s': %w", m.filePath, err)
	}
	return nil
}

func (m *metadataManager) loadAll() (map[string]ServerMetadata, error) {
	file, err := m.loadFile()
	if err != nil {
		return nil, err
	}
	return file.Servers, nil
}

func (m *metadataManager) saveAll(metadata map[string]ServerMetadata) error {
	file, err := m.loadFile()
	if err != nil {
		file = &MetadataFile{}
	}
	file.Servers = metadata
	return m.saveFile(file)
}

func (m *metadataManager) GetSettings() (Settings, error) {
	file, err := m.loadFile()
	if err != nil {
		return Settings{}, err
	}
	return file.Settings, nil
}

func (m *metadataManager) SaveSettings(settings Settings) error {
	file, err := m.loadFile()
	if err != nil {
		if m.logger != nil {
			m.logger.Errorw("failed to load metadata in SaveSettings", "path", m.filePath, "error", err)
		}
		return fmt.Errorf("load metadata: %w", err)
	}
	file.Settings = settings
	return m.saveFile(file)
}

func (m *metadataManager) updateServer(server domain.Server, oldAlias string) error {
	metadata, err := m.loadAll()
	if err != nil {
		m.logger.Errorw("failed to load metadata in updateServer", "path", m.filePath, "alias", server.Alias, "old_alias", oldAlias, "error", err)
		return fmt.Errorf("load metadata: %w", err)
	}

	if oldAlias != server.Alias {
		oldMeta, ok := metadata[oldAlias]
		if ok {
			metadata[server.Alias] = oldMeta
		}
		delete(metadata, oldAlias)
	}

	existing := metadata[server.Alias]
	merged := existing

	merged.Tags = server.Tags
	merged.Group = server.Group

	if !server.LastSeen.IsZero() {
		merged.LastSeen = server.LastSeen.Format(time.RFC3339)
	}

	if !server.PinnedAt.IsZero() {
		merged.PinnedAt = server.PinnedAt.Format(time.RFC3339)
	}

	merged.Hidden = server.Hidden

	if server.SSHCount > 0 {
		merged.SSHCount = server.SSHCount
	}

	merged.PreConnectCommand = server.PreConnectCommand
	merged.CertificateCommand = server.CertificateCommand
	merged.SyncDotfilesOnConnect = server.SyncDotfilesOnConnect

	metadata[server.Alias] = merged
	return m.saveAll(metadata)
}

// setFile records the config file neossh should write to next time the
// alias is edited or deleted. Empty path clears the memory.
func (m *metadataManager) setFile(alias, path string) error {
	metadata, err := m.loadAll()
	if err != nil {
		return fmt.Errorf("load metadata: %w", err)
	}

	meta := metadata[alias]
	if meta.File == path {
		return nil
	}
	meta.File = path
	metadata[alias] = meta
	return m.saveAll(metadata)
}

func (m *metadataManager) deleteServer(alias string) error {
	metadata, err := m.loadAll()
	if err != nil {
		m.logger.Errorw("failed to load metadata in deleteServer", "path", m.filePath, "alias", alias, "error", err)
		return fmt.Errorf("load metadata: %w", err)
	}

	delete(metadata, alias)
	return m.saveAll(metadata)
}

func (m *metadataManager) setPinned(alias string, pinned bool) error {
	metadata, err := m.loadAll()
	if err != nil {
		m.logger.Errorw("failed to load metadata in setPinned", "path", m.filePath, "alias", alias, "pinned", pinned, "error", err)
		return fmt.Errorf("load metadata: %w", err)
	}

	meta := metadata[alias]
	if pinned {
		meta.PinnedAt = time.Now().Format(time.RFC3339)
	} else {
		meta.PinnedAt = ""
	}

	metadata[alias] = meta
	return m.saveAll(metadata)
}

func (m *metadataManager) setHidden(alias string, hidden bool) error {
	metadata, err := m.loadAll()
	if err != nil {
		m.logger.Errorw("failed to load metadata in setHidden", "path", m.filePath, "alias", alias, "hidden", hidden, "error", err)
		return fmt.Errorf("load metadata: %w", err)
	}

	meta := metadata[alias]
	meta.Hidden = hidden
	metadata[alias] = meta
	return m.saveAll(metadata)
}

func (m *metadataManager) recordSSH(alias string) error {
	metadata, err := m.loadAll()
	if err != nil {
		m.logger.Errorw("failed to load metadata in recordSSH", "path", m.filePath, "alias", alias, "error", err)
		return fmt.Errorf("load metadata: %w", err)
	}

	meta := metadata[alias]
	meta.LastSeen = time.Now().Format(time.RFC3339)
	meta.SSHCount++

	metadata[alias] = meta
	return m.saveAll(metadata)
}

func (m *metadataManager) ensureDirectory() error {
	dir := filepath.Dir(m.filePath)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("mkdir '%s': %w", dir, err)
	}
	return nil
}
