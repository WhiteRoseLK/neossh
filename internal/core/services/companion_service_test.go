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
	"errors"
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/WhiteRoseLK/neossh/internal/core/domain"
	"go.uber.org/zap"
)

type mockFirstRunRepo struct {
	mockServerRepository
	firstRunCompleted bool
	saveErr           error
	getErr            error
}

func (m *mockFirstRunRepo) GetFirstRunCompleted() (bool, error) {
	return m.firstRunCompleted, m.getErr
}

func (m *mockFirstRunRepo) SaveFirstRunCompleted(c bool) error {
	m.firstRunCompleted = c
	return m.saveErr
}

func TestCompanionService_DetectPackageManager_Darwin(t *testing.T) {
	log := zap.NewNop().Sugar()

	// With brew present
	svc := NewCompanionService(log, nil,
		WithCompanionOS("darwin"),
		WithCompanionLookPath(func(bin string) (string, error) {
			if bin == "brew" || bin == "/opt/homebrew/bin/brew" {
				return "/opt/homebrew/bin/brew", nil
			}
			return "", os.ErrNotExist
		}),
	)
	pm := svc.DetectPackageManager()
	if pm.Type != domain.PackageManagerBrew {
		t.Errorf("expected brew, got %v", pm.Type)
	}

	// Without brew
	svcNone := NewCompanionService(log, nil,
		WithCompanionOS("darwin"),
		WithCompanionLookPath(func(bin string) (string, error) {
			return "", os.ErrNotExist
		}),
	)
	pmNone := svcNone.DetectPackageManager()
	if pmNone.Type != domain.PackageManagerNone {
		t.Errorf("expected none, got %v", pmNone.Type)
	}
}

func TestCompanionService_DetectPackageManager_Linux(t *testing.T) {
	log := zap.NewNop().Sugar()

	tests := []struct {
		availableBin string
		expected     domain.PackageManagerType
	}{
		{"apt-get", domain.PackageManagerApt},
		{"dnf", domain.PackageManagerDnf},
		{"pacman", domain.PackageManagerPacman},
		{"zypper", domain.PackageManagerZypper},
		{"apk", domain.PackageManagerApk},
	}

	for _, tc := range tests {
		t.Run(string(tc.expected), func(t *testing.T) {
			svc := NewCompanionService(log, nil,
				WithCompanionOS("linux"),
				WithCompanionLookPath(func(bin string) (string, error) {
					if bin == tc.availableBin {
						return "/usr/bin/" + bin, nil
					}
					return "", os.ErrNotExist
				}),
			)
			pm := svc.DetectPackageManager()
			if pm.Type != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, pm.Type)
			}
		})
	}
}

func TestCompanionService_DetectPackageManager_Windows(t *testing.T) {
	log := zap.NewNop().Sugar()

	tests := []struct {
		availableBin string
		expected     domain.PackageManagerType
	}{
		{"winget", domain.PackageManagerWinget},
		{"scoop", domain.PackageManagerScoop},
		{"choco", domain.PackageManagerChoco},
	}

	for _, tc := range tests {
		t.Run(string(tc.expected), func(t *testing.T) {
			svc := NewCompanionService(log, nil,
				WithCompanionOS("windows"),
				WithCompanionLookPath(func(bin string) (string, error) {
					if bin == tc.availableBin {
						return "C:\\Windows\\System32\\" + bin + ".exe", nil
					}
					return "", os.ErrNotExist
				}),
			)
			pm := svc.DetectPackageManager()
			if pm.Type != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, pm.Type)
			}
		})
	}
}

func TestCompanionService_CheckCompanionTools(t *testing.T) {
	log := zap.NewNop().Sugar()

	// chezmoi and yazi installed, ssh-copy-id missing
	svc := NewCompanionService(log, nil,
		WithCompanionLookPath(func(bin string) (string, error) {
			if bin == "chezmoi" || bin == "yazi" {
				return "/usr/local/bin/" + bin, nil
			}
			return "", os.ErrNotExist
		}),
	)

	tools := svc.CheckCompanionTools()
	if len(tools) != 3 {
		t.Fatalf("expected 3 tools, got %d", len(tools))
	}

	byID := make(map[string]domain.CompanionTool)
	for _, tool := range tools {
		byID[tool.ID] = tool
	}

	if !byID["chezmoi"].Installed || byID["chezmoi"].Selected {
		t.Errorf("chezmoi should be installed and not selected")
	}
	if !byID["yazi"].Installed || byID["yazi"].Selected {
		t.Errorf("yazi should be installed and not selected")
	}
	if byID["ssh-copy-id"].Installed || !byID["ssh-copy-id"].Selected {
		t.Errorf("ssh-copy-id should be missing and selected")
	}
}

func TestCompanionService_InstallCompanionTools_Success(t *testing.T) {
	log := zap.NewNop().Sugar()

	var executedCmds []string
	factory := func(name string, args ...string) *exec.Cmd {
		executedCmds = append(executedCmds, fmt.Sprintf("%s %v", name, args))
		cs := []string{"-test.run=TestHelperProcess", "--", "success"}
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
		return cmd
	}

	svc := NewCompanionService(log, nil,
		WithCompanionOS("darwin"),
		WithCompanionLookPath(func(bin string) (string, error) {
			if bin == "brew" {
				return "/usr/local/bin/brew", nil
			}
			return "", os.ErrNotExist
		}),
		WithCompanionCommandFactory(factory),
	)

	var progressStatus []string
	err := svc.InstallCompanionTools([]string{"chezmoi", "yazi"}, func(tool domain.CompanionTool, status string, err error) {
		progressStatus = append(progressStatus, fmt.Sprintf("%s:%s", tool.ID, status))
	})

	if err != nil {
		t.Fatalf("InstallCompanionTools returned error: %v", err)
	}

	if len(executedCmds) != 2 {
		t.Errorf("expected 2 commands, got %d: %v", len(executedCmds), executedCmds)
	}

	expectedProgress := []string{"chezmoi:installing", "chezmoi:installed", "yazi:installing", "yazi:installed"}
	if len(progressStatus) != len(expectedProgress) {
		t.Errorf("expected progress %v, got %v", expectedProgress, progressStatus)
	}
}

func TestCompanionService_InstallCompanionTools_Failure(t *testing.T) {
	log := zap.NewNop().Sugar()

	// Missing package manager
	svcNoPM := NewCompanionService(log, nil,
		WithCompanionOS("darwin"),
		WithCompanionLookPath(func(bin string) (string, error) {
			return "", os.ErrNotExist
		}),
	)

	err := svcNoPM.InstallCompanionTools([]string{"chezmoi"}, nil)
	if err == nil {
		t.Errorf("expected error when no package manager detected")
	}

	// Command failure
	failFactory := func(name string, args ...string) *exec.Cmd {
		cs := []string{"-test.run=TestHelperProcess", "--", "exit-1"}
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
		return cmd
	}

	svcFails := NewCompanionService(log, nil,
		WithCompanionOS("darwin"),
		WithCompanionLookPath(func(bin string) (string, error) {
			if bin == "brew" {
				return "/usr/local/bin/brew", nil
			}
			return "", os.ErrNotExist
		}),
		WithCompanionCommandFactory(failFactory),
	)

	var failedStatus bool
	errFail := svcFails.InstallCompanionTools([]string{"chezmoi"}, func(tool domain.CompanionTool, status string, err error) {
		if status == "failed" {
			failedStatus = true
		}
	})

	if errFail == nil {
		t.Errorf("expected error when command fails")
	}
	if !failedStatus {
		t.Errorf("expected failed status callback")
	}
}

func TestCompanionService_FirstRunLifecycle(t *testing.T) {
	log := zap.NewNop().Sugar()
	repo := &mockFirstRunRepo{firstRunCompleted: false}

	svc := NewCompanionService(log, repo)
	if !svc.IsFirstRun() {
		t.Errorf("expected IsFirstRun() to be true")
	}

	if err := svc.MarkFirstRunCompleted(); err != nil {
		t.Fatalf("MarkFirstRunCompleted failed: %v", err)
	}

	if svc.IsFirstRun() {
		t.Errorf("expected IsFirstRun() to be false after completion")
	}

	// Error path
	repoErr := &mockFirstRunRepo{saveErr: errors.New("disk full")}
	svcErr := NewCompanionService(log, repoErr)
	if err := svcErr.MarkFirstRunCompleted(); err == nil {
		t.Errorf("expected error when repo fails")
	}
}
