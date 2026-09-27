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
	"testing"

	"github.com/WhiteRoseLK/neossh/internal/core/domain"
)

func TestListServers_ReadsCertificateCommandFromConfigComments(t *testing.T) {
	fs := newMemFS(t)
	defer fs.cleanup()

	main := "/home/u/.ssh/config"
	content := `# Host with cert-command comment
Host step-srv # cert-command: step ssh login %u@%h
    HostName step.internal
    User dev

Host vault-srv
    # certificate-command: vault write -field=signed_key ssh/sign/user
    HostName vault.internal
    User ubuntu
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

	byAlias := make(map[string]domain.Server)
	for _, s := range servers {
		byAlias[s.Alias] = s
	}

	if s, ok := byAlias["step-srv"]; !ok || s.CertificateCommand != "step ssh login %u@%h" {
		t.Errorf("step-srv CertificateCommand = %q, want 'step ssh login %%u@%%h'", s.CertificateCommand)
	}

	if s, ok := byAlias["vault-srv"]; !ok || s.CertificateCommand != "vault write -field=signed_key ssh/sign/user" {
		t.Errorf("vault-srv CertificateCommand = %q, want 'vault write -field=signed_key ssh/sign/user'", s.CertificateCommand)
	}
}
