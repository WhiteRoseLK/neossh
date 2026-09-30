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

package services

import (
	"bytes"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/WhiteRoseLK/neossh/internal/core/domain"
	"github.com/WhiteRoseLK/neossh/internal/core/ports"
	"go.uber.org/zap"
)

type companionService struct {
	logger         *zap.SugaredLogger
	repo           ports.ServerRepository
	lookPath       func(string) (string, error)
	commandFactory func(string, ...string) *exec.Cmd
	goos           string
}

// CompanionServiceOption configures a companionService.
type CompanionServiceOption func(*companionService)

// WithCompanionCommandFactory overrides the command execution factory.
func WithCompanionCommandFactory(factory func(string, ...string) *exec.Cmd) CompanionServiceOption {
	return func(s *companionService) {
		s.commandFactory = factory
	}
}

// WithCompanionLookPath overrides executable path lookup.
func WithCompanionLookPath(lookPath func(string) (string, error)) CompanionServiceOption {
	return func(s *companionService) {
		s.lookPath = lookPath
	}
}

// WithCompanionOS overrides the detected operating system.
func WithCompanionOS(goos string) CompanionServiceOption {
	return func(s *companionService) {
		s.goos = goos
	}
}

// NewCompanionService creates a new CompanionService.
func NewCompanionService(logger *zap.SugaredLogger, repo ports.ServerRepository, opts ...CompanionServiceOption) ports.CompanionService {
	s := &companionService{
		logger:         logger,
		repo:           repo,
		lookPath:       exec.LookPath,
		commandFactory: exec.Command,
		goos:           runtime.GOOS,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// DetectPackageManager detects the available package manager on the host system.
func (s *companionService) DetectPackageManager() domain.PackageManagerInfo {
	switch s.goos {
	case "darwin":
		if path, err := s.lookPath("brew"); err == nil {
			return domain.PackageManagerInfo{
				Type:        domain.PackageManagerBrew,
				Executable:  path,
				DisplayName: "Homebrew (brew)",
			}
		}
		// Also check standard paths if not in PATH
		for _, standardPath := range []string{"/opt/homebrew/bin/brew", "/usr/local/bin/brew"} {
			if path, err := s.lookPath(standardPath); err == nil {
				return domain.PackageManagerInfo{
					Type:        domain.PackageManagerBrew,
					Executable:  path,
					DisplayName: "Homebrew (brew)",
				}
			}
		}
	case "linux":
		candidates := []struct {
			name domain.PackageManagerType
			bin  string
			disp string
		}{
			{domain.PackageManagerApt, "apt-get", "APT (apt-get)"},
			{domain.PackageManagerApt, "apt", "APT (apt)"},
			{domain.PackageManagerDnf, "dnf", "DNF (dnf)"},
			{domain.PackageManagerPacman, "pacman", "Pacman (pacman)"},
			{domain.PackageManagerZypper, "zypper", "Zypper (zypper)"},
			{domain.PackageManagerApk, "apk", "Alpine (apk)"},
		}
		for _, c := range candidates {
			if path, err := s.lookPath(c.bin); err == nil {
				return domain.PackageManagerInfo{
					Type:        c.name,
					Executable:  path,
					DisplayName: c.disp,
				}
			}
		}
	case "windows":
		candidates := []struct {
			name domain.PackageManagerType
			bin  string
			disp string
		}{
			{domain.PackageManagerWinget, "winget", "Windows Package Manager (winget)"},
			{domain.PackageManagerScoop, "scoop", "Scoop (scoop)"},
			{domain.PackageManagerChoco, "choco", "Chocolatey (choco)"},
		}
		for _, c := range candidates {
			if path, err := s.lookPath(c.bin); err == nil {
				return domain.PackageManagerInfo{
					Type:        c.name,
					Executable:  path,
					DisplayName: c.disp,
				}
			}
		}
	}

	return domain.PackageManagerInfo{
		Type:        domain.PackageManagerNone,
		Executable:  "",
		DisplayName: "None detected",
	}
}

// defaultCompanionTools returns the canonical list of companion tools.
func defaultCompanionTools() []domain.CompanionTool {
	return []domain.CompanionTool{
		{
			ID:          "chezmoi",
			Name:        "chezmoi",
			Description: "Dotfiles & shell configs sync to remote hosts with zero remote install",
			Shortcut:    "D",
			BinaryName:  "chezmoi",
		},
		{
			ID:          "yazi",
			Name:        "yazi",
			Description: "Terminal-based interactive SFTP file management and directory navigation",
			Shortcut:    "F",
			BinaryName:  "yazi",
		},
		{
			ID:          "ssh-copy-id",
			Name:        "ssh-copy-id",
			Description: "1-click OpenSSH public key deployment to remote authorized_keys",
			Shortcut:    "1-click",
			BinaryName:  "ssh-copy-id",
		},
	}
}

// CheckCompanionTools checks which tools are currently installed in PATH.
func (s *companionService) CheckCompanionTools() []domain.CompanionTool {
	tools := defaultCompanionTools()
	for i := range tools {
		if _, err := s.lookPath(tools[i].BinaryName); err == nil {
			tools[i].Installed = true
			tools[i].Selected = false
		} else {
			tools[i].Installed = false
			tools[i].Selected = true // missing tools default to selected
		}
	}
	return tools
}

type installCmdSpec struct {
	bin  string
	args []string
}

var installCommandsTable = map[domain.PackageManagerType]map[string]installCmdSpec{
	domain.PackageManagerBrew: {
		"chezmoi":     {"brew", []string{"install", "chezmoi"}},
		"yazi":        {"brew", []string{"install", "yazi"}},
		"ssh-copy-id": {"brew", []string{"install", "ssh-copy-id"}},
	},
	domain.PackageManagerApt: {
		"chezmoi":     {"apt-get", []string{"install", "-y", "chezmoi"}},
		"yazi":        {"apt-get", []string{"install", "-y", "yazi"}},
		"ssh-copy-id": {"apt-get", []string{"install", "-y", "openssh-client"}},
	},
	domain.PackageManagerDnf: {
		"chezmoi":     {"dnf", []string{"install", "-y", "chezmoi"}},
		"yazi":        {"dnf", []string{"install", "-y", "yazi"}},
		"ssh-copy-id": {"dnf", []string{"install", "-y", "openssh-clients"}},
	},
	domain.PackageManagerPacman: {
		"chezmoi":     {"pacman", []string{"-S", "--noconfirm", "chezmoi"}},
		"yazi":        {"pacman", []string{"-S", "--noconfirm", "yazi"}},
		"ssh-copy-id": {"pacman", []string{"-S", "--noconfirm", "openssh"}},
	},
	domain.PackageManagerZypper: {
		"chezmoi":     {"zypper", []string{"install", "-y", "chezmoi"}},
		"yazi":        {"zypper", []string{"install", "-y", "yazi"}},
		"ssh-copy-id": {"zypper", []string{"install", "-y", "openssh"}},
	},
	domain.PackageManagerApk: {
		"chezmoi":     {"apk", []string{"add", "chezmoi"}},
		"yazi":        {"apk", []string{"add", "yazi"}},
		"ssh-copy-id": {"apk", []string{"add", "openssh-client"}},
	},
	domain.PackageManagerWinget: {
		"chezmoi": {"winget", []string{"install", "--id", "twpayne.chezmoi", "--accept-source-agreements", "--accept-package-agreements"}},
		"yazi":    {"winget", []string{"install", "--id", "sxyazi.yazi", "--accept-source-agreements", "--accept-package-agreements"}},
	},
	domain.PackageManagerScoop: {
		"chezmoi":     {"scoop", []string{"install", "chezmoi"}},
		"yazi":        {"scoop", []string{"install", "yazi"}},
		"ssh-copy-id": {"scoop", []string{"install", "ssh-copy-id"}},
	},
	domain.PackageManagerChoco: {
		"chezmoi":     {"choco", []string{"install", "chezmoi", "-y"}},
		"yazi":        {"choco", []string{"install", "yazi", "-y"}},
		"ssh-copy-id": {"choco", []string{"install", "ssh-copy-id", "-y"}},
	},
}

// resolveInstallCommand returns the command and arguments to install the specified tool with the package manager.
func resolveInstallCommand(pm domain.PackageManagerType, toolID string) (string, []string, error) {
	if pm == domain.PackageManagerNone {
		return "", nil, fmt.Errorf("no package manager detected")
	}
	tools, ok := installCommandsTable[pm]
	if !ok {
		return "", nil, fmt.Errorf("unsupported package manager %q", pm)
	}
	spec, ok := tools[toolID]
	if !ok {
		if pm == domain.PackageManagerWinget && toolID == "ssh-copy-id" {
			return "", nil, fmt.Errorf("ssh-copy-id is included with OpenSSH for Windows")
		}
		return "", nil, fmt.Errorf("unsupported tool %q for package manager %q", toolID, pm)
	}
	return spec.bin, spec.args, nil
}

// InstallCompanionTools installs selected companion tools.
func (s *companionService) InstallCompanionTools(toolIDs []string, onProgress func(tool domain.CompanionTool, status string, err error)) error {
	pm := s.DetectPackageManager()
	if pm.Type == domain.PackageManagerNone {
		return fmt.Errorf("no supported package manager detected on host system")
	}

	toolsMap := make(map[string]domain.CompanionTool)
	for _, t := range defaultCompanionTools() {
		toolsMap[t.ID] = t
	}

	for _, toolID := range toolIDs {
		tool, ok := toolsMap[toolID]
		if !ok {
			continue
		}

		bin, args, err := resolveInstallCommand(pm.Type, toolID)
		if err != nil {
			if onProgress != nil {
				onProgress(tool, "failed", err)
			}
			return err
		}

		if onProgress != nil {
			onProgress(tool, "installing", nil)
		}

		cmd := s.commandFactory(bin, args...)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			errOutput := strings.TrimSpace(stderr.String())
			if errOutput != "" {
				err = fmt.Errorf("%w: %s", err, errOutput)
			}
			if onProgress != nil {
				onProgress(tool, "failed", err)
			}
			return fmt.Errorf("failed to install %s: %w", tool.Name, err)
		}

		tool.Installed = true
		if onProgress != nil {
			onProgress(tool, "installed", nil)
		}
	}

	return nil
}

// IsFirstRun returns true if onboarding setup has not been completed.
func (s *companionService) IsFirstRun() bool {
	if s.repo == nil {
		return false
	}
	completed, err := s.repo.GetFirstRunCompleted()
	if err == nil && completed {
		return false
	}
	return true
}

// MarkFirstRunCompleted marks the first run / onboarding as completed.
func (s *companionService) MarkFirstRunCompleted() error {
	if s.repo == nil {
		return nil
	}
	return s.repo.SaveFirstRunCompleted(true)
}
