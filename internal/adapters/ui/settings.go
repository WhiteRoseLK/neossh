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
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"go.uber.org/zap"
)

type settingsManager struct {
	mu       sync.RWMutex
	filePath string
	logger   *zap.SugaredLogger
}

// SettingsManager manages user settings in ~/.neossh/settings.json.
type SettingsManager = settingsManager

// NewDefaultSettingsManager creates a SettingsManager using default paths and a no-op logger.
func NewDefaultSettingsManager() *SettingsManager {
	return newSettingsManager(zap.NewNop().Sugar())
}

// TunnelProfile represents a saved SSH port forwarding and tunnel configuration profile.
type TunnelProfile struct {
	Name        string `json:"name"`
	Type        string `json:"type"`                   // "Local", "Remote", "Dynamic"
	Port        string `json:"port"`                   // port number
	Host        string `json:"host,omitempty"`         // destination host
	HostPort    string `json:"host_port,omitempty"`    // destination port
	BindAddress string `json:"bind_address,omitempty"` // bind address (e.g. 127.0.0.1, 0.0.0.0)
	Mode        string `json:"mode,omitempty"`         // "Only forward" or "Forward + SSH"
}

type uiSettings struct {
	SortMode                SortMode                   `json:"sort_mode,omitempty"`
	Theme                   string                     `json:"theme,omitempty"`
	DefaultIdentityKey      string                     `json:"default_identity_key,omitempty"`
	AutoPingEnabled         bool                       `json:"auto_ping_enabled,omitempty"`
	AutoPingIntervalSeconds int                        `json:"auto_ping_interval_seconds,omitempty"`
	TunnelProfiles          map[string][]TunnelProfile `json:"tunnel_profiles,omitempty"`
}

func newSettingsManager(logger *zap.SugaredLogger) *settingsManager {
	home, err := os.UserHomeDir()
	if err != nil {
		logger.Warnw("failed to determine home directory for settings", "error", err)
		return nil
	}

	settingsDir := filepath.Join(home, ".neossh")
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		settingsDir = filepath.Join(xdgConfig, "neossh")
	}
	settingsPath := filepath.Join(settingsDir, "settings.json")

	// Migrate settings from legacy lazyssh if neossh settings don't exist yet
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		legacyPath := filepath.Join(home, ".lazyssh", "settings.json")
		if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
			legacyPath = filepath.Join(xdgConfig, "lazyssh", "settings.json")
		}
		//nolint:gosec // G304: path constructed from user home directory
		if data, err := os.ReadFile(legacyPath); err == nil {
			_ = os.MkdirAll(settingsDir, 0o750)
			_ = os.WriteFile(settingsPath, data, 0o600)
		}
	}

	return &settingsManager{
		filePath: settingsPath,
		logger:   logger,
	}
}

func (m *settingsManager) LoadSortMode() (SortMode, error) {
	if m == nil {
		return SortByAliasAsc, errors.New("nil settings manager")
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	settings, err := m.loadLocked()
	if err != nil {
		return SortByAliasAsc, err
	}

	if !settings.SortMode.valid() {
		return SortByAliasAsc, nil
	}

	return settings.SortMode, nil
}

func (m *settingsManager) SaveSortMode(mode SortMode) error {
	if m == nil {
		return errors.New("nil settings manager")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	settings, err := m.loadLocked()
	if err != nil {
		return err
	}

	settings.SortMode = mode
	return m.saveLocked(settings)
}

func (m *settingsManager) LoadTheme() (string, error) {
	if m == nil {
		return ThemeDark, errors.New("nil settings manager")
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	settings, err := m.loadLocked()
	if err != nil {
		return ThemeDark, err
	}

	if settings.Theme == "" {
		return ThemeDark, nil
	}

	return settings.Theme, nil
}

func (m *settingsManager) SaveTheme(theme string) error {
	if m == nil {
		return errors.New("nil settings manager")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	settings, err := m.loadLocked()
	if err != nil {
		return err
	}

	settings.Theme = theme
	return m.saveLocked(settings)
}

func (m *settingsManager) LoadDefaultIdentityKey() (string, error) {
	if m == nil {
		return "", errors.New("nil settings manager")
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	settings, err := m.loadLocked()
	if err != nil {
		return "", err
	}

	return settings.DefaultIdentityKey, nil
}

func (m *settingsManager) SaveDefaultIdentityKey(key string) error {
	if m == nil {
		return errors.New("nil settings manager")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	settings, err := m.loadLocked()
	if err != nil {
		return err
	}

	settings.DefaultIdentityKey = key
	return m.saveLocked(settings)
}

func (m *settingsManager) LoadAutoPing() (bool, int, error) {
	if m == nil {
		return false, 60, errors.New("nil settings manager")
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	settings, err := m.loadLocked()
	if err != nil {
		return false, 60, err
	}

	interval := settings.AutoPingIntervalSeconds
	if interval <= 0 {
		interval = 60
	}

	return settings.AutoPingEnabled, interval, nil
}

func (m *settingsManager) SaveAutoPing(enabled bool, intervalSeconds int) error {
	if m == nil {
		return errors.New("nil settings manager")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	settings, err := m.loadLocked()
	if err != nil {
		return err
	}

	settings.AutoPingEnabled = enabled
	if intervalSeconds > 0 {
		settings.AutoPingIntervalSeconds = intervalSeconds
	}
	return m.saveLocked(settings)
}

// LoadTunnelProfiles loads saved favorite tunnel profiles for the given host alias.
func (m *settingsManager) LoadTunnelProfiles(alias string) ([]TunnelProfile, error) {
	if m == nil {
		return nil, errors.New("nil settings manager")
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	settings, err := m.loadLocked()
	if err != nil {
		return nil, err
	}

	if settings.TunnelProfiles == nil {
		return nil, nil
	}

	profiles := settings.TunnelProfiles[alias]
	if len(profiles) == 0 {
		return nil, nil
	}

	out := make([]TunnelProfile, len(profiles))
	copy(out, profiles)
	return out, nil
}

// SaveTunnelProfile saves or updates a favorite tunnel profile for the given host alias.
func (m *settingsManager) SaveTunnelProfile(alias string, profile TunnelProfile) error {
	if m == nil {
		return errors.New("nil settings manager")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	settings, err := m.loadLocked()
	if err != nil {
		return err
	}

	if settings.TunnelProfiles == nil {
		settings.TunnelProfiles = make(map[string][]TunnelProfile)
	}

	profiles := settings.TunnelProfiles[alias]
	found := false
	for i, p := range profiles {
		if strings.EqualFold(p.Name, profile.Name) {
			profiles[i] = profile
			found = true
			break
		}
	}
	if !found {
		profiles = append(profiles, profile)
	}
	settings.TunnelProfiles[alias] = profiles

	return m.saveLocked(settings)
}

// DeleteTunnelProfile deletes a saved tunnel profile by name for the given host alias.
func (m *settingsManager) DeleteTunnelProfile(alias string, profileName string) error {
	if m == nil {
		return errors.New("nil settings manager")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	settings, err := m.loadLocked()
	if err != nil {
		return err
	}

	if settings.TunnelProfiles == nil {
		return nil
	}

	profiles := settings.TunnelProfiles[alias]
	newProfiles := make([]TunnelProfile, 0, len(profiles))
	for _, p := range profiles {
		if !strings.EqualFold(p.Name, profileName) {
			newProfiles = append(newProfiles, p)
		}
	}

	if len(newProfiles) == 0 {
		delete(settings.TunnelProfiles, alias)
	} else {
		settings.TunnelProfiles[alias] = newProfiles
	}

	return m.saveLocked(settings)
}

func (m *settingsManager) loadLocked() (uiSettings, error) {
	var settings uiSettings

	data, err := os.ReadFile(m.filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return settings, nil
		}
		return settings, err
	}

	if len(data) == 0 {
		return settings, nil
	}

	if err := json.Unmarshal(data, &settings); err != nil {
		// Handle unmarshal errors (e.g., old string format vs new int format)
		// by returning default settings. This allows automatic migration from
		// old format to new format when SaveSortMode is called.
		m.logger.Warnw("failed to parse settings file, using defaults", "error", err, "path", m.filePath)
		return uiSettings{}, nil
	}

	return settings, nil
}

func (m *settingsManager) saveLocked(settings uiSettings) error {
	if err := os.MkdirAll(filepath.Dir(m.filePath), 0o750); err != nil {
		return err
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.filePath, data, 0o600)
}
