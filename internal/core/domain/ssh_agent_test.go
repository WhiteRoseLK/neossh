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

import (
	"testing"
)

func TestDetectAgentType(t *testing.T) {
	tests := []struct {
		socket string
		want   AgentType
	}{
		{"/Users/user/Library/Group Containers/2b62183b.com.1password/t/agent.sock", AgentType1Password},
		{"/Users/user/.gnupg/S.gpg-agent.ssh", AgentTypeGPGAgent},
		{"/private/tmp/com.apple.launchd.xxx/Listeners", AgentTypeApple},
		{`\\.\pipe\openssh-ssh-agent`, AgentTypeWindows},
		{"/tmp/ssh-abcdef/agent.1234", AgentTypeOpenSSH},
		{"", AgentTypeGeneric},
	}

	for _, tt := range tests {
		got := DetectAgentType(tt.socket)
		if got != tt.want {
			t.Errorf("DetectAgentType(%q) = %q, want %q", tt.socket, got, tt.want)
		}
	}
}

func TestSSHAgentStatus_HasKey(t *testing.T) {
	status := SSHAgentStatus{
		Available: true,
		Type:      AgentTypeOpenSSH,
		KeyCount:  2,
		Keys: []AgentKeyRecord{
			{
				Format:      "ssh-ed25519",
				Fingerprint: "SHA256:abc123xyz",
				Comment:     "/home/user/.ssh/id_ed25519",
			},
			{
				Format:      "rsa-sha2-512",
				Fingerprint: "SHA256:rsa456key",
				Comment:     "work-key",
			},
		},
	}

	// Match by fingerprint
	if !status.HasKey("SHA256:abc123xyz", "", "") {
		t.Error("expected key to match by fingerprint")
	}

	// Match by comment
	if !status.HasKey("", "work-key", "") {
		t.Error("expected key to match by comment")
	}

	// Match by path
	if !status.HasKey("", "", "/home/user/.ssh/id_ed25519") {
		t.Error("expected key to match by path")
	}

	// Non-matching
	if status.HasKey("SHA256:unknown", "unknown", "/tmp/unknown") {
		t.Error("expected key not to match")
	}

	// When unavailable
	unavailable := SSHAgentStatus{Available: false}
	if unavailable.HasKey("SHA256:abc123xyz", "", "") {
		t.Error("expected false when agent unavailable")
	}
}
