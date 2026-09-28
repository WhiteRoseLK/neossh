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
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/WhiteRoseLK/neossh/internal/core/domain"
)

func TestFormatSSHErrorMessage_Diagnostics(t *testing.T) {
	// Create an expired test certificate
	tempDir := t.TempDir()
	expiredCertPath := filepath.Join(tempDir, "id_ed25519-cert.pub")
	nowSec := uint64(time.Now().Unix())
	validAfter := nowSec - 7200
	validBefore := nowSec - 3600
	expiredCertBytes := createTestSSHCertificate(t, validAfter, validBefore, "test-expired-key", []string{"root"})
	if err := os.WriteFile(expiredCertPath, expiredCertBytes, 0o600); err != nil {
		t.Fatalf("failed to write expired test cert: %v", err)
	}

	tests := []struct {
		name          string
		alias         string
		rawErr        string
		server        *domain.Server
		wantTitle     string
		wantInMessage string
	}{
		{
			name:          "Host key mismatch",
			alias:         "srv1",
			rawErr:        "Host key verification failed.",
			server:        nil,
			wantTitle:     `SSH connection to "srv1" failed`,
			wantInMessage: "Host Key Mismatch",
		},
		{
			name:          "Permission denied",
			alias:         "prod-web",
			rawErr:        "Permission denied (publickey,gssapi-keyex,gssapi-with-mic,password).",
			server:        nil,
			wantTitle:     `SSH connection to "prod-web" failed`,
			wantInMessage: "Authentication Refused",
		},
		{
			name:          "Connection refused",
			alias:         "db-node",
			rawErr:        "ssh: connect to host 192.168.1.100 port 22: Connection refused",
			server:        nil,
			wantTitle:     `SSH connection to "db-node" failed`,
			wantInMessage: "Connection Refused",
		},
		{
			name:          "Connection timed out",
			alias:         "vpn-remote",
			rawErr:        "ssh: connect to host 10.8.0.5 port 2222: Connection timed out",
			server:        nil,
			wantTitle:     `SSH connection to "vpn-remote" failed`,
			wantInMessage: "Connection Timed Out",
		},
		{
			name:          "DNS resolution failure",
			alias:         "bad-domain",
			rawErr:        "ssh: Could not resolve hostname unknown.internal: nodename nor servname provided",
			server:        nil,
			wantTitle:     `SSH connection to "bad-domain" failed`,
			wantInMessage: "DNS Resolution Error",
		},
		{
			name:          "ProxyJump failure",
			alias:         "jump-host",
			rawErr:        "kex_exchange_identification: Connection closed by remote host\r\nssh: ProxyCommand / ProxyJump failed",
			server:        nil,
			wantTitle:     `SSH connection to "jump-host" failed`,
			wantInMessage: "Proxy / Bastion Error",
		},
		{
			name:          "Non-zero exit code without stderr",
			alias:         "srv-exit",
			rawErr:        "exit status 255",
			server:        nil,
			wantTitle:     `SSH connection to "srv-exit" failed`,
			wantInMessage: "Non-Zero Exit Code",
		},
		{
			name:          "Empty error message fallback",
			alias:         "srv-empty",
			rawErr:        "",
			server:        nil,
			wantTitle:     `SSH connection to "srv-empty" failed`,
			wantInMessage: "Unknown error",
		},
		{
			name:   "Expired certificate detection",
			alias:  "cert-server",
			rawErr: "Permission denied (publickey).",
			server: &domain.Server{
				Alias:           "cert-server",
				CertificateFile: expiredCertPath,
			},
			wantTitle:     `SSH connection to "cert-server" failed`,
			wantInMessage: "SSH Certificate Expired",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			title, msg := formatSSHErrorMessage(tt.server, tt.alias, tt.rawErr)
			if title != tt.wantTitle {
				t.Errorf("title = %q, want %q", title, tt.wantTitle)
			}
			if !strings.Contains(msg, tt.wantInMessage) {
				t.Errorf("message does not contain %q. Got:\n%s", tt.wantInMessage, msg)
			}
		})
	}
}
