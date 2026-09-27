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
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/WhiteRoseLK/neossh/internal/core/domain"
	"go.uber.org/zap"
)

func TestServerService_LaunchFileManager_StandardSFTP(t *testing.T) {
	repo := &mockServerRepository{
		servers: []domain.Server{
			{
				Alias: "srv-prod",
				Host:  "192.168.1.100",
				User:  "admin",
				Port:  2222,
			},
		},
	}

	sftpExecuted := false
	var passedArgs []string

	svc := &serverService{
		logger:           zap.NewNop().Sugar(),
		serverRepository: repo,
		newSFTPCommand: func(alias string, args []string) *exec.Cmd {
			sftpExecuted = true
			passedArgs = args
			cs := []string{"-test.run=TestHelperProcess", "--", "success", alias}
			cmd := exec.Command(os.Args[0], cs...)
			cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
			return cmd
		},
	}

	err := svc.SFTP("srv-prod")
	if err != nil {
		t.Fatalf("expected SFTP to succeed, got: %v", err)
	}
	if !sftpExecuted {
		t.Errorf("expected sftp command to be executed")
	}
	_ = passedArgs
	if repo.recordCalls != 1 || repo.lastAlias != "srv-prod" {
		t.Errorf("expected RecordSSH to be called for srv-prod, got %d calls for %q", repo.recordCalls, repo.lastAlias)
	}
}

func TestServerService_LaunchFileManager_WildcardError(t *testing.T) {
	repo := &mockServerRepository{
		servers: []domain.Server{
			{
				Alias:      "*.staging",
				Host:       "10.0.0.1",
				IsWildcard: true,
			},
		},
	}

	svc := &serverService{
		logger:           zap.NewNop().Sugar(),
		serverRepository: repo,
	}

	err := svc.LaunchFileManager("*.staging", "")
	if err == nil {
		t.Fatal("expected error launching file manager for wildcard host")
	}
	if !strings.Contains(err.Error(), "wildcard") {
		t.Errorf("expected error to mention wildcard, got: %v", err)
	}
}

func TestServerService_LaunchFileManager_NotFound(t *testing.T) {
	repo := &mockServerRepository{
		servers: []domain.Server{
			{Alias: "srv1", Host: "10.0.0.1"},
		},
	}

	svc := &serverService{
		logger:           zap.NewNop().Sugar(),
		serverRepository: repo,
	}

	err := svc.LaunchFileManager("nonexistent", "")
	if err == nil {
		t.Fatal("expected error for nonexistent server")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected error to mention 'not found', got: %v", err)
	}
}

func TestServerService_LaunchFileManager_ConfiguredTools(t *testing.T) {
	repo := &mockServerRepository{
		servers: []domain.Server{
			{
				Alias: "remote-box",
				Host:  "myserver.com",
				User:  "deploy",
				Port:  2202,
			},
		},
	}

	tests := []struct {
		name          string
		tool          string
		expectedCmd   string
		expectedArgs  []string
		expectedIsGUI bool
	}{
		{
			name:          "yazi terminal file manager",
			tool:          "yazi",
			expectedCmd:   "yazi",
			expectedArgs:  []string{"sftp://deploy@myserver.com:2202/"},
			expectedIsGUI: false,
		},
		{
			name:          "ranger terminal file manager",
			tool:          "ranger",
			expectedCmd:   "ranger",
			expectedArgs:  []string{"sftp://deploy@myserver.com:2202/"},
			expectedIsGUI: false,
		},
		{
			name:          "filezilla GUI client",
			tool:          "filezilla",
			expectedCmd:   "filezilla",
			expectedArgs:  []string{"sftp://deploy@myserver.com:2202/"},
			expectedIsGUI: true,
		},
		{
			name:          "dolphin KDE file manager with fish url",
			tool:          "dolphin",
			expectedCmd:   "dolphin",
			expectedArgs:  []string{"fish://deploy@myserver.com:2202/"},
			expectedIsGUI: true,
		},
		{
			name:          "nautilus GNOME files with sftp url",
			tool:          "nautilus",
			expectedCmd:   "nautilus",
			expectedArgs:  []string{"sftp://deploy@myserver.com:2202/"},
			expectedIsGUI: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &serverService{
				logger:           zap.NewNop().Sugar(),
				serverRepository: repo,
			}

			sftpURL := buildSFTPURL("deploy", "myserver.com", 2202)
			fishURL := buildFishURL("deploy", "myserver.com", 2202)

			cmd, isGUI := svc.buildFileManagerCmd(tt.tool, repo.servers[0], sftpURL, fishURL, repo.GetConfigFile())

			if isGUI != tt.expectedIsGUI {
				t.Errorf("expected isGUI=%v, got %v", tt.expectedIsGUI, isGUI)
			}

			if !strings.HasSuffix(cmd.Path, tt.expectedCmd) && filepath.Base(cmd.Path) != tt.expectedCmd {
				t.Errorf("expected command %q, got %q", tt.expectedCmd, cmd.Path)
			}

			if len(tt.expectedArgs) > 0 {
				found := false
				for _, arg := range cmd.Args {
					if arg == tt.expectedArgs[0] {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected arg %q in command args %v", tt.expectedArgs[0], cmd.Args)
				}
			}
		})
	}
}

func TestServerService_LaunchFileManager_CustomTemplate(t *testing.T) {
	repo := &mockServerRepository{
		servers: []domain.Server{
			{
				Alias: "web1",
				Host:  "10.0.0.10",
				User:  "ubuntu",
				Port:  22,
			},
		},
	}

	svc := &serverService{
		logger:           zap.NewNop().Sugar(),
		serverRepository: repo,
	}

	sftpURL := buildSFTPURL("ubuntu", "10.0.0.10", 22)
	fishURL := buildFishURL("ubuntu", "10.0.0.10", 22)

	customTemplate := "my-transfer-tool --alias=%a --host=%h --user=%u --target=%url"
	cmd, isGUI := svc.buildFileManagerCmd(customTemplate, repo.servers[0], sftpURL, fishURL, "~/.ssh/config")
	if isGUI {
		t.Errorf("expected custom command not to be GUI by default")
	}

	cmdStr := strings.Join(cmd.Args, " ")
	if !strings.Contains(cmdStr, "--alias=web1") {
		t.Errorf("expected --alias=web1 in %s", cmdStr)
	}
	if !strings.Contains(cmdStr, "--host=10.0.0.10") {
		t.Errorf("expected --host=10.0.0.10 in %s", cmdStr)
	}
	if !strings.Contains(cmdStr, "--user=ubuntu") {
		t.Errorf("expected --user=ubuntu in %s", cmdStr)
	}
	if !strings.Contains(cmdStr, "sftp://ubuntu@10.0.0.10/") {
		t.Errorf("expected sftp URL in %s", cmdStr)
	}
}

func TestBuildSFTPAndFishURL(t *testing.T) {
	tests := []struct {
		user         string
		host         string
		port         int
		expectedSFTP string
		expectedFish string
	}{
		{
			user:         "root",
			host:         "1.2.3.4",
			port:         22,
			expectedSFTP: "sftp://root@1.2.3.4/",
			expectedFish: "fish://root@1.2.3.4/",
		},
		{
			user:         "alice",
			host:         "example.com",
			port:         2222,
			expectedSFTP: "sftp://alice@example.com:2222/",
			expectedFish: "fish://alice@example.com:2222/",
		},
		{
			user:         "",
			host:         "example.com",
			port:         22,
			expectedSFTP: "sftp://example.com/",
			expectedFish: "fish://example.com/",
		},
	}

	for _, tt := range tests {
		sftpURL := buildSFTPURL(tt.user, tt.host, tt.port)
		if sftpURL != tt.expectedSFTP {
			t.Errorf("buildSFTPURL(%s, %s, %d) = %q, expected %q", tt.user, tt.host, tt.port, sftpURL, tt.expectedSFTP)
		}
		fishURL := buildFishURL(tt.user, tt.host, tt.port)
		if fishURL != tt.expectedFish {
			t.Errorf("buildFishURL(%s, %s, %d) = %q, expected %q", tt.user, tt.host, tt.port, fishURL, tt.expectedFish)
		}
	}
}
