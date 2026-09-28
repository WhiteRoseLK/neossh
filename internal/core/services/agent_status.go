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
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/WhiteRoseLK/neossh/internal/core/domain"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// QuerySSHAgentStatus inspects the active SSH agent via UNIX domain socket or ssh-add fallback.
func QuerySSHAgentStatus() domain.SSHAgentStatus {
	sock := os.Getenv("SSH_AUTH_SOCK")
	if sock == "" {
		if runtime.GOOS == "windows" {
			sock = `\\.\pipe\openssh-ssh-agent`
		} else {
			return domain.SSHAgentStatus{
				Available: false,
				Type:      domain.AgentTypeGeneric,
				Error:     "SSH_AUTH_SOCK is not set",
			}
		}
	}

	agentType := domain.DetectAgentType(sock)

	// On unix-like platforms, connect via unix socket
	if !strings.HasPrefix(sock, `\\.\pipe\`) {
		// #nosec G704 -- sock is local UNIX socket path for SSH agent
		conn, err := net.DialTimeout("unix", sock, 1*time.Second)
		if err == nil {
			defer func() {
				_ = conn.Close()
			}()

			agClient := agent.NewClient(conn)
			keys, err := agClient.List()
			if err == nil {
				var records []domain.AgentKeyRecord
				for _, k := range keys {
					fp := ""
					if pubKey, err := ssh.ParsePublicKey(k.Blob); err == nil {
						fp = ssh.FingerprintSHA256(pubKey)
					} else if pubKey, err := ssh.ParsePublicKey(k.Marshal()); err == nil {
						fp = ssh.FingerprintSHA256(pubKey)
					}
					records = append(records, domain.AgentKeyRecord{
						Format:      k.Format,
						Fingerprint: fp,
						Comment:     k.Comment,
					})
				}

				return domain.SSHAgentStatus{
					Available:  true,
					SocketPath: sock,
					Type:       agentType,
					KeyCount:   len(records),
					Keys:       records,
				}
			}
		}
	}

	// Subprocess fallback (ssh-add -l)
	return queryAgentViaSubprocess(sock, agentType)
}

func queryAgentViaSubprocess(sock string, agentType domain.AgentType) domain.SSHAgentStatus {
	cmd := exec.Command("ssh-add", "-l")
	output, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			if exitErr.ExitCode() == 1 {
				// Exit code 1 means agent is alive but has no identities
				return domain.SSHAgentStatus{
					Available:  true,
					SocketPath: sock,
					Type:       agentType,
					KeyCount:   0,
					Keys:       []domain.AgentKeyRecord{},
				}
			}
		}
		return domain.SSHAgentStatus{
			Available:  false,
			SocketPath: sock,
			Type:       agentType,
			Error:      "Failed to connect to ssh-agent",
		}
	}

	var records []domain.AgentKeyRecord
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		fields := strings.Fields(trimmed)
		if len(fields) >= 2 {
			fp := fields[1]
			comment := ""
			format := ""
			if len(fields) >= 3 {
				comment = fields[2]
			}
			if len(fields) >= 4 {
				format = strings.Trim(fields[len(fields)-1], "()")
			}
			records = append(records, domain.AgentKeyRecord{
				Format:      format,
				Fingerprint: fp,
				Comment:     comment,
			})
		}
	}

	return domain.SSHAgentStatus{
		Available:  true,
		SocketPath: sock,
		Type:       agentType,
		KeyCount:   len(records),
		Keys:       records,
	}
}
