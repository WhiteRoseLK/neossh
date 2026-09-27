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

package ssh_config_file

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/WhiteRoseLK/neossh/internal/core/domain"
)

func TestConvertCLIForwardToConfigFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "basic local forward",
			input:    "8080:localhost:80",
			expected: "8080 localhost:80",
		},
		{
			name:     "local forward with bind address",
			input:    "127.0.0.1:8080:localhost:80",
			expected: "127.0.0.1:8080 localhost:80",
		},
		{
			name:     "local forward with wildcard bind",
			input:    "*:8080:localhost:80",
			expected: "*:8080 localhost:80",
		},
		{
			name:     "remote forward",
			input:    "8080:localhost:3000",
			expected: "8080 localhost:3000",
		},
		{
			name:     "remote forward with bind address",
			input:    "0.0.0.0:80:localhost:8080",
			expected: "0.0.0.0:80 localhost:8080",
		},
		{
			name:     "forward with IPv6 address",
			input:    "8080:[2001:db8::1]:80",
			expected: "8080 [2001:db8::1]:80",
		},
		{
			name:     "forward with domain",
			input:    "3306:db.example.com:3306",
			expected: "3306 db.example.com:3306",
		},
		{
			name:     "invalid format - only one colon",
			input:    "8080:localhost",
			expected: "8080:localhost", // returned as-is
		},
		{
			name:     "invalid format - no colons",
			input:    "8080",
			expected: "8080", // returned as-is
		},
	}

	r := &Repository{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := r.convertCLIForwardToConfigFormat(tt.input)
			if result != tt.expected {
				t.Errorf("convertCLIForwardToConfigFormat(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestConvertConfigForwardToCLIFormat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "basic local forward",
			input:    "8080 localhost:80",
			expected: "8080:localhost:80",
		},
		{
			name:     "local forward with bind address",
			input:    "127.0.0.1:8080 localhost:80",
			expected: "127.0.0.1:8080:localhost:80",
		},
		{
			name:     "local forward with wildcard bind",
			input:    "*:8080 localhost:80",
			expected: "*:8080:localhost:80",
		},
		{
			name:     "remote forward",
			input:    "8080 localhost:3000",
			expected: "8080:localhost:3000",
		},
		{
			name:     "remote forward with bind address",
			input:    "0.0.0.0:80 localhost:8080",
			expected: "0.0.0.0:80:localhost:8080",
		},
		{
			name:     "forward with IPv6 address",
			input:    "8080 [2001:db8::1]:80",
			expected: "8080:[2001:db8::1]:80",
		},
		{
			name:     "forward with domain",
			input:    "3306 db.example.com:3306",
			expected: "3306:db.example.com:3306",
		},
		{
			name:     "already in CLI format",
			input:    "8080:localhost:80",
			expected: "8080:localhost:80", // returned as-is
		},
		{
			name:     "no space separator",
			input:    "8080",
			expected: "8080", // returned as-is
		},
	}

	r := &Repository{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := r.convertConfigForwardToCLIFormat(tt.input)
			if result != tt.expected {
				t.Errorf("convertConfigForwardToCLIFormat(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestListServers_WildcardPatternBlocks(t *testing.T) {
	fs := newMemFS(t)
	defer fs.cleanup()

	main := "/home/u/.ssh/config"
	content := `# Global comment
Host *.corp
    User corpuser
    Port 2222
    IdentityFile ~/.ssh/corp_id

Host *
    ServerAliveInterval 60
    ServerAliveCountMax 3

Host *.internal.example.com
    User devops

Host bastion
    HostName bastion.example.com
    User admin
`
	fs.write(main, content)
	tmpMeta := filepath.Join(t.TempDir(), "metadata.json")
	r := newRepoForFS(t, fs, tmpMeta)

	servers, err := r.ListServers("")
	if err != nil {
		t.Fatalf("ListServers failed: %v", err)
	}

	if len(servers) != 4 {
		t.Fatalf("expected 4 servers, got %d", len(servers))
	}

	byAlias := make(map[string]domain.Server)
	for _, s := range servers {
		byAlias[s.Alias] = s
	}

	// Verify *.corp
	corp, ok := byAlias["*.corp"]
	if !ok {
		t.Fatalf("missing *.corp server")
	}
	if !corp.IsWildcard {
		t.Errorf("expected *.corp IsWildcard=true")
	}
	if !corp.IsWildcardServer() {
		t.Errorf("expected *.corp IsWildcardServer()=true")
	}
	if corp.User != "corpuser" {
		t.Errorf("expected user 'corpuser', got %q", corp.User)
	}
	if corp.Port != 2222 {
		t.Errorf("expected port 2222, got %d", corp.Port)
	}
	if len(corp.IdentityFiles) != 1 || corp.IdentityFiles[0] != "~/.ssh/corp_id" {
		t.Errorf("expected IdentityFile '~/.ssh/corp_id', got %v", corp.IdentityFiles)
	}

	// Verify *
	all, ok := byAlias["*"]
	if !ok {
		t.Fatalf("missing * server")
	}
	if !all.IsWildcard {
		t.Errorf("expected * IsWildcard=true")
	}
	if !all.IsWildcardServer() {
		t.Errorf("expected * IsWildcardServer()=true")
	}
	if all.ServerAliveInterval != "60" {
		t.Errorf("expected ServerAliveInterval '60', got %q", all.ServerAliveInterval)
	}

	// Verify *.internal.example.com
	internal, ok := byAlias["*.internal.example.com"]
	if !ok {
		t.Fatalf("missing *.internal.example.com server")
	}
	if !internal.IsWildcard {
		t.Errorf("expected *.internal.example.com IsWildcard=true")
	}
	if internal.User != "devops" {
		t.Errorf("expected user 'devops', got %q", internal.User)
	}

	// Verify bastion
	bastion, ok := byAlias["bastion"]
	if !ok {
		t.Fatalf("missing bastion server")
	}
	if bastion.IsWildcard {
		t.Errorf("expected bastion IsWildcard=false")
	}
	if bastion.IsWildcardServer() {
		t.Errorf("expected bastion IsWildcardServer()=false")
	}
	if bastion.Host != "bastion.example.com" {
		t.Errorf("expected host 'bastion.example.com', got %q", bastion.Host)
	}
}

func TestUpdateServer_PreservesWildcards(t *testing.T) {
	fs := newMemFS(t)
	defer fs.cleanup()

	main := "/home/u/.ssh/config"
	content := `Host *.corp
    User corpuser
    Port 2222

Host bastion
    HostName bastion.example.com
    User admin
    Port 22

Host *
    ServerAliveInterval 60
`
	fs.write(main, content)
	tmpMeta := filepath.Join(t.TempDir(), "metadata.json")
	r := newRepoForFS(t, fs, tmpMeta)

	oldBastion := domain.Server{
		Alias: "bastion",
		Host:  "bastion.example.com",
		User:  "admin",
		Port:  22,
	}
	newBastion := oldBastion
	newBastion.User = "superadmin"
	newBastion.Port = 2200

	if err := r.UpdateServer(oldBastion, newBastion); err != nil {
		t.Fatalf("UpdateServer failed: %v", err)
	}

	saved := fs.read(main)
	if !strings.Contains(saved, "Host *.corp") {
		t.Errorf("expected 'Host *.corp' to be preserved in:\n%s", saved)
	}
	if !strings.Contains(saved, "User corpuser") {
		t.Errorf("expected 'User corpuser' to be preserved in:\n%s", saved)
	}
	if !strings.Contains(saved, "Host *") {
		t.Errorf("expected 'Host *' to be preserved in:\n%s", saved)
	}
	if !strings.Contains(saved, "ServerAliveInterval 60") {
		t.Errorf("expected 'ServerAliveInterval 60' to be preserved in:\n%s", saved)
	}
	if !strings.Contains(saved, "User superadmin") {
		t.Errorf("expected 'User superadmin' in:\n%s", saved)
	}
}

func TestAddAndDeleteServer_PreservesWildcards(t *testing.T) {
	fs := newMemFS(t)
	defer fs.cleanup()

	main := "/home/u/.ssh/config"
	content := `Host *.corp
    User corpuser
    Port 2222

Host bastion
    HostName bastion.example.com
    User admin
`
	fs.write(main, content)
	tmpMeta := filepath.Join(t.TempDir(), "metadata.json")
	r := newRepoForFS(t, fs, tmpMeta)

	// Add server
	newSrv := domain.Server{
		Alias: "web-01",
		Host:  "web.example.com",
		User:  "ubuntu",
		Port:  22,
	}
	if err := r.AddServer(newSrv); err != nil {
		t.Fatalf("AddServer failed: %v", err)
	}

	saved := fs.read(main)
	if !strings.Contains(saved, "Host *.corp") || !strings.Contains(saved, "User corpuser") {
		t.Fatalf("wildcard block corrupted after AddServer:\n%s", saved)
	}

	// Delete bastion
	bastion := domain.Server{Alias: "bastion"}
	if err := r.DeleteServer(bastion); err != nil {
		t.Fatalf("DeleteServer failed: %v", err)
	}

	savedAfterDelete := fs.read(main)
	if !strings.Contains(savedAfterDelete, "Host *.corp") || !strings.Contains(savedAfterDelete, "User corpuser") {
		t.Fatalf("wildcard block corrupted after DeleteServer:\n%s", savedAfterDelete)
	}
	if strings.Contains(savedAfterDelete, "bastion.example.com") {
		t.Errorf("bastion was not deleted:\n%s", savedAfterDelete)
	}
}

func TestListServers_QuotedHostAliases(t *testing.T) {
	fs := newMemFS(t)
	defer fs.cleanup()

	main := "/home/u/.ssh/config"
	content := `Host "TEST-Testing" 'staging-box'
    HostName 172.16.1.1
    User my-user
    Port 22
    IdentityFile ~/.ssh/my_file.pem

Host "single-quoted-server"
    HostName 10.0.0.1
    User dev
`
	fs.write(main, content)
	tmpMeta := filepath.Join(t.TempDir(), "metadata.json")
	r := newRepoForFS(t, fs, tmpMeta)

	servers, err := r.ListServers("")
	if err != nil {
		t.Fatalf("ListServers failed: %v", err)
	}

	if len(servers) != 2 {
		t.Fatalf("expected 2 servers, got %d", len(servers))
	}

	s1 := servers[0]
	if s1.Alias != "TEST-Testing" {
		t.Errorf("expected clean primary alias 'TEST-Testing', got %q", s1.Alias)
	}
	if len(s1.Aliases) != 2 || s1.Aliases[0] != "TEST-Testing" || s1.Aliases[1] != "staging-box" {
		t.Errorf("unexpected aliases: %+v", s1.Aliases)
	}
	if s1.Host != "172.16.1.1" {
		t.Errorf("unexpected Host: %q", s1.Host)
	}

	s2 := servers[1]
	if s2.Alias != "single-quoted-server" {
		t.Errorf("expected clean alias 'single-quoted-server', got %q", s2.Alias)
	}

	// Verify update works properly on quoted host
	s1Updated := s1
	s1Updated.Host = "172.16.1.2"
	s1Updated.User = "updated-user"
	if err := r.UpdateServer(s1, s1Updated); err != nil {
		t.Fatalf("UpdateServer on originally quoted host failed: %v", err)
	}

	saved := fs.read(main)
	if !strings.Contains(saved, "172.16.1.2") || !strings.Contains(saved, "updated-user") {
		t.Errorf("expected updated values in saved config:\n%s", saved)
	}
}

func TestCertificateFile_CRUD(t *testing.T) {
	fs := newMemFS(t)
	defer fs.cleanup()

	main := "/home/u/.ssh/config"
	fs.write(main, "")
	tmpMeta := filepath.Join(t.TempDir(), "metadata.json")
	r := newRepoForFS(t, fs, tmpMeta)

	server := domain.Server{
		Alias:           "cert-host",
		Host:            "192.168.1.100",
		User:            "root",
		CertificateFile: "~/.ssh/id_ed25519-cert.pub",
		SourceFile:      main,
	}

	if err := r.AddServer(server); err != nil {
		t.Fatalf("AddServer failed: %v", err)
	}

	servers, err := r.ListServers("")
	if err != nil {
		t.Fatalf("ListServers failed: %v", err)
	}
	if len(servers) != 1 {
		t.Fatalf("expected 1 server, got %d", len(servers))
	}
	if servers[0].CertificateFile != "~/.ssh/id_ed25519-cert.pub" {
		t.Errorf("expected CertificateFile '~/.ssh/id_ed25519-cert.pub', got %q", servers[0].CertificateFile)
	}

	saved := fs.read(main)
	if !strings.Contains(saved, "CertificateFile ~/.ssh/id_ed25519-cert.pub") {
		t.Errorf("expected config to contain CertificateFile, got:\n%s", saved)
	}

	// Update CertificateFile
	updated := servers[0]
	updated.CertificateFile = "~/.ssh/custom-cert.pub"
	if err := r.UpdateServer(servers[0], updated); err != nil {
		t.Fatalf("UpdateServer failed: %v", err)
	}

	servers, err = r.ListServers("")
	if err != nil {
		t.Fatalf("ListServers after update failed: %v", err)
	}
	if servers[0].CertificateFile != "~/.ssh/custom-cert.pub" {
		t.Errorf("expected updated CertificateFile '~/.ssh/custom-cert.pub', got %q", servers[0].CertificateFile)
	}
}
