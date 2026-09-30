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
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/WhiteRoseLK/neossh/internal/core/domain"
	"go.uber.org/zap"
)

func helperChezmoiFactory(scenario string) func(...string) *exec.Cmd {
	return func(args ...string) *exec.Cmd {
		cs := []string{"-test.run=TestHelperProcess", "--", scenario}
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
		return cmd
	}
}

func helperSSHExecFactory(scenario string) func(string, string, bool) *exec.Cmd {
	return func(alias, command string, interactive bool) *exec.Cmd {
		cs := []string{"-test.run=TestHelperProcess", "--", scenario, alias}
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
		return cmd
	}
}

func TestServerService_IsChezmoiAvailable(t *testing.T) {
	logger := zap.NewNop().Sugar()
	repo := &mockServerRepository{}

	// Case 1: LookPath returns error
	svcMissing := NewServerService(logger, repo, WithLookPath(func(file string) (string, error) {
		if file == "chezmoi" {
			return "", errors.New("not found")
		}
		return "/usr/bin/" + file, nil
	}))
	if svcMissing.IsChezmoiAvailable() {
		t.Errorf("expected IsChezmoiAvailable to return false when chezmoi is not in PATH")
	}

	// Case 2: LookPath returns path
	svcAvailable := NewServerService(logger, repo, WithLookPath(func(file string) (string, error) {
		if file == "chezmoi" {
			return "/usr/local/bin/chezmoi", nil
		}
		return "/usr/bin/" + file, nil
	}))
	if !svcAvailable.IsChezmoiAvailable() {
		t.Errorf("expected IsChezmoiAvailable to return true when chezmoi is in PATH")
	}
}

func TestServerService_SyncDotfiles_ReadOnly(t *testing.T) {
	logger := zap.NewNop().Sugar()
	repo := &mockServerRepository{}
	svc := NewServerService(logger, repo, WithReadOnly(true))

	err := svc.SyncDotfiles("myhost")
	if !errors.Is(err, ErrReadOnly) {
		t.Errorf("expected ErrReadOnly, got %v", err)
	}
}

func TestServerService_SyncDotfiles_Wildcard(t *testing.T) {
	logger := zap.NewNop().Sugar()
	repo := &mockServerRepository{}
	svc := NewServerService(logger, repo)

	err := svc.SyncDotfiles("*.example.com")
	if err == nil || !strings.Contains(err.Error(), "wildcard") {
		t.Errorf("expected wildcard pattern error, got %v", err)
	}
}

func TestServerService_SyncDotfiles_MissingChezmoi(t *testing.T) {
	logger := zap.NewNop().Sugar()
	repo := &mockServerRepository{}
	svc := NewServerService(logger, repo, WithLookPath(func(file string) (string, error) {
		return "", errors.New("not found")
	}))

	err := svc.SyncDotfiles("myhost")
	if err == nil || !strings.Contains(err.Error(), "chezmoi not found") {
		t.Errorf("expected 'chezmoi not found' error, got %v", err)
	}
}

func TestServerService_SyncDotfiles_Success(t *testing.T) {
	logger := zap.NewNop().Sugar()
	repo := &mockServerRepository{}

	svc := NewServerService(logger, repo,
		WithLookPath(func(file string) (string, error) {
			return "/usr/bin/" + file, nil
		}),
		WithChezmoiCommand(helperChezmoiFactory("chezmoi-archive")),
	).(*serverService)

	svc.newSSHExecCommand = helperSSHExecFactory("tar-extract")

	err := svc.SyncDotfiles("myhost")
	if err != nil {
		t.Fatalf("expected successful dotfiles sync, got: %v", err)
	}
}

func TestServerService_SyncDotfiles_ChezmoiError(t *testing.T) {
	logger := zap.NewNop().Sugar()
	repo := &mockServerRepository{}

	svc := NewServerService(logger, repo,
		WithLookPath(func(file string) (string, error) {
			return "/usr/bin/" + file, nil
		}),
		WithChezmoiCommand(helperChezmoiFactory("chezmoi-fail")),
	).(*serverService)

	svc.newSSHExecCommand = helperSSHExecFactory("tar-extract")

	err := svc.SyncDotfiles("myhost")
	if err == nil || !strings.Contains(err.Error(), "chezmoi archive error") {
		t.Fatalf("expected chezmoi archive error, got: %v", err)
	}
}

func TestServerService_SyncDotfiles_RemoteTarError(t *testing.T) {
	logger := zap.NewNop().Sugar()
	repo := &mockServerRepository{}

	svc := NewServerService(logger, repo,
		WithLookPath(func(file string) (string, error) {
			return "/usr/bin/" + file, nil
		}),
		WithChezmoiCommand(helperChezmoiFactory("chezmoi-archive")),
	).(*serverService)

	svc.newSSHExecCommand = helperSSHExecFactory("tar-fail")

	err := svc.SyncDotfiles("myhost")
	if err == nil || !strings.Contains(err.Error(), "remote extraction error") {
		t.Fatalf("expected remote extraction error, got: %v", err)
	}
}

func TestServerService_MaybeSyncDotfilesOnConnect(t *testing.T) {
	logger := zap.NewNop().Sugar()
	repo := &mockServerRepository{
		servers: []domain.Server{
			{Alias: "srv-synced", Host: "1.2.3.4", SyncDotfilesOnConnect: true},
			{Alias: "srv-plain", Host: "1.2.3.5", SyncDotfilesOnConnect: false},
		},
	}

	syncCalled := false
	svc := NewServerService(logger, repo,
		WithLookPath(func(file string) (string, error) {
			return "/usr/bin/" + file, nil
		}),
		WithChezmoiCommand(func(args ...string) *exec.Cmd {
			syncCalled = true
			return helperChezmoiFactory("chezmoi-archive")(args...)
		}),
	).(*serverService)

	svc.newSSHExecCommand = helperSSHExecFactory("tar-extract")
	if err := svc.ReloadServers(); err != nil {
		t.Fatalf("failed to reload servers: %v", err)
	}

	// 1. Plain server: should not trigger sync
	svc.maybeSyncDotfilesOnConnect("srv-plain")
	if syncCalled {
		t.Errorf("expected no sync on connect for srv-plain")
	}

	// 2. Synced server: should trigger sync
	svc.maybeSyncDotfilesOnConnect("srv-synced")
	if !syncCalled {
		t.Errorf("expected sync to be called for srv-synced")
	}
}

func TestServerService_SyncDotfiles_WrappedWithPassword(t *testing.T) {
	logger := zap.NewNop().Sugar()
	repo := &mockServerRepository{}
	credStore := &mockCredentialStore{
		passwords: map[string]string{
			"auth-srv": "secretpass",
		},
	}

	svc := NewServerService(logger, repo,
		WithCredentialStore(credStore),
		WithLookPath(func(file string) (string, error) {
			if file == "sshpass" {
				return "/usr/bin/sshpass", nil
			}
			return "/usr/bin/" + file, nil
		}),
		WithChezmoiCommand(helperChezmoiFactory("chezmoi-archive")),
	).(*serverService)

	wrapped := false
	svc.newSSHPassCommand = func(sshpassPath, pwd string, sshCmd *exec.Cmd) *exec.Cmd {
		if pwd == "secretpass" {
			wrapped = true
		}
		cs := []string{"-test.run=TestHelperProcess", "--", "tar-extract"}
		cmd := exec.Command(os.Args[0], cs...)
		cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
		cmd.Stdin = os.Stdin
		cmd.Stdout = io.Discard
		cmd.Stderr = io.Discard
		return cmd
	}

	svc.newSSHExecCommand = func(alias, command string, interactive bool) *exec.Cmd {
		return exec.Command("ssh", alias, command)
	}

	err := svc.SyncDotfiles("auth-srv")
	if err != nil {
		t.Fatalf("expected successful sync with sshpass: %v", err)
	}
	if !wrapped {
		t.Errorf("expected ssh command to be wrapped with sshpass password")
	}
}
