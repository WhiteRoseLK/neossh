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
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/WhiteRoseLK/neossh/internal/core/domain"
	"go.uber.org/zap"
)

type bundleMockRepo struct {
	mockServerRepository
	configFiles []string
	metaFile    string
}

func (m *bundleMockRepo) GetConfigFiles() ([]string, error) {
	return m.configFiles, nil
}

func (m *bundleMockRepo) GetMetadataFile() string {
	return m.metaFile
}

func (m *bundleMockRepo) ListServers(query string) ([]domain.Server, error) {
	return m.servers, nil
}

func TestBundleService_ExportAndImport(t *testing.T) {
	tempDir := t.TempDir()

	sshDir := filepath.Join(tempDir, "source_ssh")
	confD := filepath.Join(sshDir, "conf.d")
	neosshDir := filepath.Join(tempDir, "source_neossh")
	if err := os.MkdirAll(confD, 0o700); err != nil {
		t.Fatalf("failed to create conf.d: %v", err)
	}
	if err := os.MkdirAll(neosshDir, 0o700); err != nil {
		t.Fatalf("failed to create neosshDir: %v", err)
	}

	mainConfigPath := filepath.Join(sshDir, "config")
	includeConfigPath := filepath.Join(confD, "work.conf")

	mainConfigContent := `Include conf.d/*.conf

Host web-prod
    HostName 192.168.1.50
    User admin
    Port 22
    IdentityFile ~/.ssh/id_ed25519
    # password: secretpassword123
`
	includeConfigContent := `Host db-internal
    HostName 10.0.0.5
    User postgres
    IdentityFile ~/.ssh/id_rsa
`
	if err := os.WriteFile(mainConfigPath, []byte(mainConfigContent), 0o600); err != nil {
		t.Fatalf("failed to write main config: %v", err)
	}
	if err := os.WriteFile(includeConfigPath, []byte(includeConfigContent), 0o600); err != nil {
		t.Fatalf("failed to write include config: %v", err)
	}

	metaContent := `{
  "settings": {
    "theme": "nord",
    "default_identity_key": "~/.ssh/id_ed25519"
  },
  "servers": {
    "web-prod": {
      "tags": ["production", "aws"],
      "file": "/Users/alice/.ssh/config",
      "pre_connect_command": "vault-login --token=secret-token-xyz"
    }
  }
}`
	metaPath := filepath.Join(neosshDir, "metadata.json")
	if err := os.WriteFile(metaPath, []byte(metaContent), 0o600); err != nil {
		t.Fatalf("failed to write metadata.json: %v", err)
	}

	settingsContent := `{
  "theme": "nord",
  "default_identity_key": "~/.ssh/id_ed25519",
  "auto_ping_enabled": true
}`
	settingsPath := filepath.Join(neosshDir, "settings.json")
	if err := os.WriteFile(settingsPath, []byte(settingsContent), 0o600); err != nil {
		t.Fatalf("failed to write settings.json: %v", err)
	}

	repo := &bundleMockRepo{
		configFiles: []string{mainConfigPath, includeConfigPath},
		metaFile:    metaPath,
		mockServerRepository: mockServerRepository{
			servers: []domain.Server{
				{Alias: "web-prod", Host: "192.168.1.50"},
				{Alias: "db-internal", Host: "10.0.0.5"},
			},
		},
	}

	logger := zap.NewNop().Sugar()
	svc := NewBundleService(repo, logger, "v1.2.0")

	bundlePath := filepath.Join(tempDir, "my-export.tar.gz")

	// 1. Export without sanitize
	summary, err := svc.Export(domain.ExportOptions{
		SSHConfigFile: mainConfigPath,
		NeosshDir:     neosshDir,
		OutputPath:    bundlePath,
		Sanitize:      false,
	})
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}
	if summary.Manifest.ServerCount != 2 {
		t.Errorf("expected 2 servers, got %d", summary.Manifest.ServerCount)
	}
	if len(summary.Manifest.Files) != 4 {
		t.Errorf("expected 4 files (main, include, metadata, settings), got %d", len(summary.Manifest.Files))
	}
	if summary.Manifest.Sanitized {
		t.Errorf("expected Sanitized to be false")
	}

	// 2. Verify bundle
	verifySummary, err := svc.Verify(bundlePath)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}
	if verifySummary.Manifest.ServerCount != 2 {
		t.Errorf("expected 2 servers in verify, got %d", verifySummary.Manifest.ServerCount)
	}
	if len(verifySummary.Warnings) > 0 {
		t.Errorf("unexpected warnings in verify: %v", verifySummary.Warnings)
	}

	// 3. Dry-Run Import
	targetSSHDir := filepath.Join(tempDir, "target_ssh")
	targetNeosshDir := filepath.Join(tempDir, "target_neossh")

	dryRunSummary, err := svc.Import(domain.ImportOptions{
		BundlePath:      bundlePath,
		TargetSSHDir:    targetSSHDir,
		TargetNeosshDir: targetNeosshDir,
		DryRun:          true,
	})
	if err != nil {
		t.Fatalf("DryRun Import failed: %v", err)
	}
	if len(dryRunSummary.RestoredFiles) != 4 {
		t.Errorf("expected 4 dry-run restored files, got %d", len(dryRunSummary.RestoredFiles))
	}
	// Verify target directories were NOT created during dry run
	if _, err := os.Stat(targetSSHDir); !os.IsNotExist(err) {
		t.Errorf("targetSSHDir should not exist after dry run")
	}

	// 4. Real Import into new target directories
	importSummary, err := svc.Import(domain.ImportOptions{
		BundlePath:      bundlePath,
		TargetSSHDir:    targetSSHDir,
		TargetNeosshDir: targetNeosshDir,
		DryRun:          false,
		CreateBackup:    true,
	})
	if err != nil {
		t.Fatalf("Real Import failed: %v", err)
	}
	if len(importSummary.RestoredFiles) != 4 {
		t.Errorf("expected 4 restored files, got %d", len(importSummary.RestoredFiles))
	}

	// Verify imported file contents
	restoredMainConfig, err := os.ReadFile(filepath.Join(targetSSHDir, "config"))
	if err != nil {
		t.Fatalf("failed to read restored main config: %v", err)
	}
	if !strings.Contains(string(restoredMainConfig), "Host web-prod") {
		t.Errorf("restored config missing web-prod host")
	}

	restoredIncludeConfig, err := os.ReadFile(filepath.Join(targetSSHDir, "conf.d", "work.conf"))
	if err != nil {
		t.Fatalf("failed to read restored include config: %v", err)
	}
	if !strings.Contains(string(restoredIncludeConfig), "Host db-internal") {
		t.Errorf("restored include config missing db-internal host")
	}

	restoredMeta, err := os.ReadFile(filepath.Join(targetNeosshDir, "metadata.json"))
	if err != nil {
		t.Fatalf("failed to read restored metadata: %v", err)
	}
	if !strings.Contains(string(restoredMeta), "production") {
		t.Errorf("restored metadata missing tags")
	}

	// 5. Re-importing over existing files with backup creation
	reimportSummary, err := svc.Import(domain.ImportOptions{
		BundlePath:      bundlePath,
		TargetSSHDir:    targetSSHDir,
		TargetNeosshDir: targetNeosshDir,
		DryRun:          false,
		CreateBackup:    true,
	})
	if err != nil {
		t.Fatalf("Re-import failed: %v", err)
	}
	if len(reimportSummary.BackupFiles) != 4 {
		t.Errorf("expected 4 backup files generated on re-import, got %d", len(reimportSummary.BackupFiles))
	}
}

func TestBundleService_ExportSanitized(t *testing.T) {
	tempDir := t.TempDir()

	sshDir := filepath.Join(tempDir, "ssh")
	neosshDir := filepath.Join(tempDir, "neossh")
	_ = os.MkdirAll(sshDir, 0o700)
	_ = os.MkdirAll(neosshDir, 0o700)

	mainConfigPath := filepath.Join(sshDir, "config")
	mainContent := `Host secret-box
    HostName 10.1.2.3
    User developer
    IdentityFile ~/.ssh/super_secret_rsa
    # password: MySuperSecretPassword!
    # token: ghp_123456789
    PreConnectCommand vault login --token=my-secret-token %h
`
	if err := os.WriteFile(mainConfigPath, []byte(mainContent), 0o600); err != nil {
		t.Fatalf("failed to write main config: %v", err)
	}

	metaPath := filepath.Join(neosshDir, "metadata.json")
	metaContent := `{
  "settings": {
    "theme": "catppuccin",
    "default_identity_key": "/Users/developer/.ssh/id_rsa"
  },
  "servers": {
    "secret-box": {
      "file": "/Users/developer/.ssh/config",
      "pre_connect_command": "vpn connect --token=xyz"
    }
  }
}`
	if err := os.WriteFile(metaPath, []byte(metaContent), 0o600); err != nil {
		t.Fatalf("failed to write metadata: %v", err)
	}

	settingsPath := filepath.Join(neosshDir, "settings.json")
	settingsContent := `{
  "default_identity_key": "/Users/developer/.ssh/id_rsa",
  "theme": "dracula"
}`
	if err := os.WriteFile(settingsPath, []byte(settingsContent), 0o600); err != nil {
		t.Fatalf("failed to write settings: %v", err)
	}

	repo := &bundleMockRepo{
		configFiles: []string{mainConfigPath},
		metaFile:    metaPath,
	}

	svc := NewBundleService(repo, zap.NewNop().Sugar(), "v1.2.0")
	sanitizedBundle := filepath.Join(tempDir, "sanitized-bundle.tar.gz")

	summary, err := svc.Export(domain.ExportOptions{
		SSHConfigFile: mainConfigPath,
		NeosshDir:     neosshDir,
		OutputPath:    sanitizedBundle,
		Sanitize:      true,
	})
	if err != nil {
		t.Fatalf("Export with sanitize failed: %v", err)
	}
	if !summary.Manifest.Sanitized {
		t.Errorf("expected manifest.Sanitized to be true")
	}
	if summary.Manifest.Hostname != "sanitized-host" {
		t.Errorf("expected hostname 'sanitized-host', got %q", summary.Manifest.Hostname)
	}

	// Restore to inspect sanitized content
	restoreDir := filepath.Join(tempDir, "restored_sanitized")
	targetSSH := filepath.Join(restoreDir, "ssh")
	targetNeossh := filepath.Join(restoreDir, "neossh")

	_, err = svc.Import(domain.ImportOptions{
		BundlePath:      sanitizedBundle,
		TargetSSHDir:    targetSSH,
		TargetNeosshDir: targetNeossh,
	})
	if err != nil {
		t.Fatalf("Import of sanitized bundle failed: %v", err)
	}

	// 1. Verify SSH config sanitation
	cfgBytes, err := os.ReadFile(filepath.Join(targetSSH, "config"))
	if err != nil {
		t.Fatalf("read sanitized config failed: %v", err)
	}
	cfgStr := string(cfgBytes)

	if strings.Contains(cfgStr, "super_secret_rsa") {
		t.Errorf("sanitized config still contains private key path: %s", cfgStr)
	}
	if !strings.Contains(cfgStr, "# IdentityFile [sanitized]") {
		t.Errorf("sanitized config missing '# IdentityFile [sanitized]': %s", cfgStr)
	}
	if strings.Contains(cfgStr, "MySuperSecretPassword!") {
		t.Errorf("sanitized config still contains password comment: %s", cfgStr)
	}
	if strings.Contains(cfgStr, "ghp_123456789") {
		t.Errorf("sanitized config still contains token comment: %s", cfgStr)
	}
	if strings.Contains(cfgStr, "my-secret-token") {
		t.Errorf("sanitized config still contains command token: %s", cfgStr)
	}
	if !strings.Contains(cfgStr, "token=[sanitized]") {
		t.Errorf("sanitized command token placeholder missing: %s", cfgStr)
	}

	// 2. Verify metadata sanitation
	metaBytes, _ := os.ReadFile(filepath.Join(targetNeossh, "metadata.json"))
	var parsedMeta map[string]interface{}
	_ = json.Unmarshal(metaBytes, &parsedMeta)
	metaSettings := parsedMeta["settings"].(map[string]interface{})
	if metaSettings["default_identity_key"] != "" {
		t.Errorf("expected default_identity_key to be cleared, got: %v", metaSettings["default_identity_key"])
	}

	servers := parsedMeta["servers"].(map[string]interface{})
	box := servers["secret-box"].(map[string]interface{})
	if box["file"] != "~/.ssh/config" {
		t.Errorf("expected sanitized file path '~/.ssh/config', got %v", box["file"])
	}

	// 3. Verify settings.json sanitation
	setBytes, _ := os.ReadFile(filepath.Join(targetNeossh, "settings.json"))
	var parsedSettings map[string]interface{}
	_ = json.Unmarshal(setBytes, &parsedSettings)
	if parsedSettings["default_identity_key"] != "" {
		t.Errorf("expected settings default_identity_key to be cleared, got: %v", parsedSettings["default_identity_key"])
	}
}

func TestBundleService_ZipSlipPrevention(t *testing.T) {
	tempDir := t.TempDir()
	maliciousTar := filepath.Join(tempDir, "malicious.tar.gz")

	outFile, err := os.Create(maliciousTar)
	if err != nil {
		t.Fatalf("failed to create tar file: %v", err)
	}
	defer outFile.Close()

	gw := gzip.NewWriter(outFile)
	tw := tar.NewWriter(gw)

	// Write manifest
	manifest := domain.BundleManifest{
		Version: CurrentBundleVersion,
		Files: []domain.BundleFileEntry{
			{ArchivePath: "manifest.json"},
		},
	}
	manifestBytes, _ := json.Marshal(manifest)
	_ = writeTarEntry(tw, "manifest.json", manifestBytes, 0o600)

	// Inject illegal ZipSlip entry
	evilHeader := &tar.Header{
		Name:     "../../../../../../tmp/evil_file.txt",
		Mode:     0o600,
		Size:     10,
		Typeflag: tar.TypeReg,
	}
	_ = tw.WriteHeader(evilHeader)
	_, _ = tw.Write([]byte("malicious!"))

	_ = tw.Close()
	_ = gw.Close()

	svc := NewBundleService(nil, zap.NewNop().Sugar(), "v1.2.0")

	// Verify should reject ZipSlip path traversal
	_, err = svc.Verify(maliciousTar)
	if err == nil {
		t.Fatalf("expected Verify to reject path traversal, but succeeded")
	}
	if !strings.Contains(err.Error(), "illegal path traversal") {
		t.Errorf("expected error to mention 'illegal path traversal', got: %v", err)
	}

	// Import should also reject it
	_, err = svc.Import(domain.ImportOptions{
		BundlePath:   maliciousTar,
		TargetSSHDir: tempDir,
	})
	if err == nil {
		t.Fatalf("expected Import to reject path traversal, but succeeded")
	}
}
