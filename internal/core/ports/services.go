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

package ports

import (
	"time"

	"github.com/WhiteRoseLK/neossh/internal/core/domain"
)

type ServerService interface {
	ListServers(query string) ([]domain.Server, error)
	UpdateServer(server domain.Server, newServer domain.Server) error
	AddServer(server domain.Server) error
	DeleteServer(server domain.Server) error
	SetPinned(alias string, pinned bool) error
	SetHidden(alias string, hidden bool) error
	SSH(alias string) error
	SSHWithArgs(alias string, extraArgs []string) error
	SFTP(alias string) error
	LaunchFileManager(alias string, customTool string) error
	CopySSHKey(alias string) error
	StartForward(alias string, extraArgs []string) (int, error)
	StopForwarding(alias string) error
	IsForwarding(alias string) bool
	Ping(server domain.Server) (bool, time.Duration, error)
	DiscoverKnownHosts(knownHostsPath string) ([]domain.Server, domain.ImportResult, error)
	ImportKnownHosts(knownHostsPath string) (domain.ImportResult, error)
	ListKnownHostRecords(knownHostsPath string) ([]domain.KnownHostRecord, error)
	RemoveKnownHost(knownHostsPath, host string, port int) (string, int, error)
	RemoveKnownHostByLine(knownHostsPath string, lineNumber int) (string, error)
	ScanAndAddKnownHost(knownHostsPath, host string, port int) (*domain.KnownHostRecord, error)
	GetTheme() (string, error)
	SaveTheme(theme string) error
	ListActiveSessions(query string) ([]domain.Server, error)
	KillActiveSessions(server domain.Server) (int, error)
	ResolveConfigServer(server domain.Server) (domain.Server, bool, error)
	GetDefaultIdentityKey() (string, error)
	SaveDefaultIdentityKey(key string) error
	ReloadServers() error
	UpdateServerPing(alias string, status string, latency time.Duration)
	GetSSHAgentStatus() domain.SSHAgentStatus
	GetSnippets() ([]domain.Snippet, error)
	SaveSnippet(snippet domain.Snippet) error
	DeleteSnippet(id string) error
	ExecuteRemoteCommand(alias string, command string) (string, error)
	RunInteractiveRemoteCommand(alias string, command string) error
	SyncDotfiles(alias string) error
	IsChezmoiAvailable() bool
}

// GitService provides Git and SSH key management operations.
type GitService interface {
	IsGitRepository(path string) bool
	GetGitRootPath(path string) (string, error)
	GetPushRemoteURL(repoPath string) (remoteName, remoteURL string, err error)
	ListSSHKeys(sshDir string, serverRepo ServerRepository) ([]domain.SSHKey, error)
	ConfigureGitSSHKey(repoPath string, keyPath string, scope string) error
	GetCurrentGitSSHConfig(repoPath string) (string, error)
	ClearGitSSHConfig(repoPath string, scope string) error
	GetLoadedAgentKeys() ([]string, error)
	GetSSHAgentStatus() domain.SSHAgentStatus

	// SSH Key management
	ListAllSSHKeys(serverRepo ServerRepository) ([]domain.SSHKey, error)
	ListSSHKeysFromConfig(serverRepo ServerRepository) ([]domain.SSHKey, error)
	ListSSHKeysFromAgent() ([]domain.SSHKey, error)
	LoadKeyToAgent(keyPath string) error
	UnloadKeyFromAgent(publicKeyLine string) error
	UpdateKeyComment(keyPath, comment string) error

	// Server repository resolution
	SetServerRepository(repo ServerRepository)
}

// CompanionService provides companion tools detection, installation, and onboarding lifecycle management.
type CompanionService interface {
	DetectPackageManager() domain.PackageManagerInfo
	CheckCompanionTools() []domain.CompanionTool
	InstallCompanionTools(toolIDs []string, onProgress func(tool domain.CompanionTool, status string, err error)) error
	IsFirstRun() bool
	MarkFirstRunCompleted() error
}
