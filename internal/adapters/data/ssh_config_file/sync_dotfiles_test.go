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

func TestListServers_ReadsSyncDotfilesFromConfigComments(t *testing.T) {
	fs := newMemFS(t)
	defer fs.cleanup()

	main := "/home/u/.ssh/config"
	content := `# Host with sync-dotfiles comments
Host srv1 # sync-dotfiles: true
    HostName srv1.internal
    User dev

Host srv2
    # dotfiles-sync: yes
    HostName srv2.internal
    User dev

Host srv3
    # dotfiles: on
    HostName srv3.internal
    User dev

Host srv4 # sync-dotfiles: 1
    HostName srv4.internal

Host srv5 # sync-dotfiles: false
    HostName srv5.internal

Host srv6
    HostName srv6.internal
`
	fs.write(main, content)
	tmpMeta := filepath.Join(t.TempDir(), "metadata.json")
	r := newRepoForFS(t, fs, tmpMeta)

	servers, err := r.ListServers("")
	if err != nil {
		t.Fatalf("ListServers failed: %v", err)
	}

	byAlias := make(map[string]domain.Server)
	for _, s := range servers {
		byAlias[s.Alias] = s
	}

	for _, alias := range []string{"srv1", "srv2", "srv3", "srv4"} {
		if s, ok := byAlias[alias]; !ok || !s.SyncDotfilesOnConnect {
			t.Errorf("expected %s to have SyncDotfilesOnConnect=true, got %v", alias, s.SyncDotfilesOnConnect)
		}
	}

	for _, alias := range []string{"srv5", "srv6"} {
		if s, ok := byAlias[alias]; !ok || s.SyncDotfilesOnConnect {
			t.Errorf("expected %s to have SyncDotfilesOnConnect=false, got %v", alias, s.SyncDotfilesOnConnect)
		}
	}
}

func TestMetadataManager_PersistsSyncDotfiles(t *testing.T) {
	metaFile := filepath.Join(t.TempDir(), "metadata.json")
	mgr := newMetadataManager(metaFile, nil)

	s := domain.Server{
		Alias:                 "srv-test",
		SyncDotfilesOnConnect: true,
	}

	if err := mgr.updateServer(s, "srv-test"); err != nil {
		t.Fatalf("updateServer failed: %v", err)
	}

	all, err := mgr.loadAll()
	if err != nil {
		t.Fatalf("loadAll failed: %v", err)
	}

	meta, ok := all["srv-test"]
	if !ok || !meta.SyncDotfilesOnConnect {
		t.Errorf("expected meta.SyncDotfilesOnConnect=true, got %v", meta.SyncDotfilesOnConnect)
	}

	mgr2 := newMetadataManager(metaFile, nil)
	all2, err := mgr2.loadAll()
	if err != nil {
		t.Fatalf("second loadAll failed: %v", err)
	}

	if !all2["srv-test"].SyncDotfilesOnConnect {
		t.Errorf("expected persisted SyncDotfilesOnConnect to be true")
	}
}
