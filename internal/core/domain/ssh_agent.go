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

package domain

import "strings"

// AgentType identifies the implementation of the detected SSH agent.
type AgentType string

const (
	AgentTypeOpenSSH   AgentType = "OpenSSH"
	AgentType1Password AgentType = "1Password"
	AgentTypeGPGAgent  AgentType = "GPG-Agent"
	AgentTypeApple     AgentType = "macOS Keychain"
	AgentTypeWindows   AgentType = "Windows OpenSSH"
	AgentTypeGeneric   AgentType = "Generic Agent"
)

// AgentKeyRecord represents an active key loaded inside the SSH agent.
type AgentKeyRecord struct {
	Format      string `json:"format"`
	Fingerprint string `json:"fingerprint"`
	Comment     string `json:"comment"`
}

// SSHAgentStatus provides live telemetry on the local SSH agent.
type SSHAgentStatus struct {
	Available  bool             `json:"available"`
	SocketPath string           `json:"socket_path"`
	Type       AgentType        `json:"type"`
	KeyCount   int              `json:"key_count"`
	Keys       []AgentKeyRecord `json:"keys"`
	Error      string           `json:"error,omitempty"`
}

// DetectAgentType determines the agent implementation based on its socket or pipe path.
func DetectAgentType(socketPath string) AgentType {
	lower := strings.ToLower(socketPath)
	switch {
	case strings.Contains(lower, "1password") || strings.Contains(lower, "2b62183b"):
		return AgentType1Password
	case strings.Contains(lower, "gnupg") || strings.Contains(lower, "gpg"):
		return AgentTypeGPGAgent
	case strings.Contains(lower, "com.apple") || strings.Contains(lower, "listeners"):
		return AgentTypeApple
	case strings.Contains(lower, "openssh-ssh-agent") || strings.Contains(lower, "pipe"):
		return AgentTypeWindows
	case socketPath != "":
		return AgentTypeOpenSSH
	default:
		return AgentTypeGeneric
	}
}

// HasKey checks if a key by fingerprint, comment, or filename is loaded in the agent status.
func (s SSHAgentStatus) HasKey(fingerprint, comment, path string) bool {
	if !s.Available || len(s.Keys) == 0 {
		return false
	}
	fpLower := strings.ToLower(strings.TrimSpace(fingerprint))
	commentLower := strings.ToLower(strings.TrimSpace(comment))
	pathLower := strings.ToLower(strings.TrimSpace(path))

	for _, k := range s.Keys {
		if fpLower != "" && strings.EqualFold(strings.TrimSpace(k.Fingerprint), fpLower) {
			return true
		}
		kComment := strings.ToLower(strings.TrimSpace(k.Comment))
		if kComment == "" {
			continue
		}
		if commentLower != "" && kComment == commentLower {
			return true
		}
		if pathLower != "" && (kComment == pathLower || strings.HasSuffix(pathLower, "/"+kComment)) {
			return true
		}
	}
	return false
}
