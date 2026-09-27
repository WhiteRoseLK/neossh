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

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBundleCommands_EndToEnd(t *testing.T) {
	tempDir := t.TempDir()

	// Setup fake home directory
	fakeHome := filepath.Join(tempDir, "home")
	sshDir := filepath.Join(fakeHome, ".ssh")
	neosshDir := filepath.Join(fakeHome, ".neossh")
	if err := os.MkdirAll(sshDir, 0o700); err != nil {
		t.Fatalf("failed to create ssh dir: %v", err)
	}
	if err := os.MkdirAll(neosshDir, 0o700); err != nil {
		t.Fatalf("failed to create neossh dir: %v", err)
	}

	cfgFile := filepath.Join(sshDir, "config")
	cfgContent := `Host srv-test
    HostName 10.0.0.1
    User testuser
    IdentityFile ~/.ssh/id_test
    # password: MySecretPassword
`
	if err := os.WriteFile(cfgFile, []byte(cfgContent), 0o600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	metaFile := filepath.Join(neosshDir, "metadata.json")
	metaContent := `{"servers":{"srv-test":{"tags":["tag1"]}}}`
	if err := os.WriteFile(metaFile, []byte(metaContent), 0o600); err != nil {
		t.Fatalf("failed to write metadata: %v", err)
	}

	t.Setenv("HOME", fakeHome)

	bundleFile := filepath.Join(tempDir, "test-bundle.tar.gz")

	// 1. Test 'export --sanitize'
	exportCmd := newExportCmd()
	var exportOut bytes.Buffer
	exportCmd.SetOut(&exportOut)
	exportCmd.SetArgs([]string{"--output", bundleFile, "--sanitize", "--sshconfig", cfgFile})

	if err := exportCmd.Execute(); err != nil {
		t.Fatalf("export command failed: %v", err)
	}
	outStr := exportOut.String()
	if !strings.Contains(outStr, "Successfully exported configuration bundle") {
		t.Errorf("expected success message in export output, got: %s", outStr)
	}
	if !strings.Contains(outStr, "Sanitation: YES") {
		t.Errorf("expected Sanitation: YES in export output, got: %s", outStr)
	}

	// 2. Test 'verify'
	verifyCmd := newVerifyCmd()
	var verifyOut bytes.Buffer
	verifyCmd.SetOut(&verifyOut)
	verifyCmd.SetArgs([]string{bundleFile})

	if err := verifyCmd.Execute(); err != nil {
		t.Fatalf("verify command failed: %v", err)
	}
	verifyStr := verifyOut.String()
	if !strings.Contains(verifyStr, "Bundle is valid") {
		t.Errorf("expected 'Bundle is valid' in verify output, got: %s", verifyStr)
	}
	if !strings.Contains(verifyStr, "Sanitized: true") {
		t.Errorf("expected 'Sanitized: true' in verify output, got: %s", verifyStr)
	}

	// 3. Test 'import --dry-run'
	targetSSH := filepath.Join(tempDir, "target_home", ".ssh")
	targetNeossh := filepath.Join(tempDir, "target_home", ".neossh")

	importCmd := newImportCmd()
	var dryRunOut bytes.Buffer
	importCmd.SetOut(&dryRunOut)
	importCmd.SetArgs([]string{bundleFile, "--dry-run", "--sshdir", targetSSH, "--neosshdir", targetNeossh})

	if err := importCmd.Execute(); err != nil {
		t.Fatalf("import dry-run failed: %v", err)
	}
	dryStr := dryRunOut.String()
	if !strings.Contains(dryStr, "Dry-run verification completed") {
		t.Errorf("expected dry-run message, got: %s", dryStr)
	}
	if _, err := os.Stat(targetSSH); !os.IsNotExist(err) {
		t.Errorf("expected targetSSH not to exist after dry-run")
	}

	// 4. Test real 'import'
	realImportCmd := newImportCmd()
	var importOut bytes.Buffer
	realImportCmd.SetOut(&importOut)
	realImportCmd.SetArgs([]string{bundleFile, "--sshdir", targetSSH, "--neosshdir", targetNeossh})

	if err := realImportCmd.Execute(); err != nil {
		t.Fatalf("real import failed: %v", err)
	}
	realImportStr := importOut.String()
	if !strings.Contains(realImportStr, "Configuration bundle successfully restored") {
		t.Errorf("expected success message, got: %s", realImportStr)
	}

	// Verify restored file exists and was sanitized
	restoredCfg, err := os.ReadFile(filepath.Join(targetSSH, "config"))
	if err != nil {
		t.Fatalf("failed to read restored config: %v", err)
	}
	restoredStr := string(restoredCfg)
	if strings.Contains(restoredStr, "id_test") {
		t.Errorf("restored config should have been sanitized, found id_test: %s", restoredStr)
	}
	if !strings.Contains(restoredStr, "# IdentityFile [sanitized]") {
		t.Errorf("expected '# IdentityFile [sanitized]', got: %s", restoredStr)
	}
	if strings.Contains(restoredStr, "MySecretPassword") {
		t.Errorf("restored config still contains password: %s", restoredStr)
	}

	// 5. Test root command invocation with 'backup' alias
	root := newRootCmd()
	var rootOut bytes.Buffer
	root.SetOut(&rootOut)
	backupFile := filepath.Join(tempDir, "root-backup.tar.gz")
	root.SetArgs([]string{"backup", "-o", backupFile, "--sshconfig", cfgFile})
	if err := root.Execute(); err != nil {
		t.Fatalf("root command 'backup' failed: %v", err)
	}
	if _, err := os.Stat(backupFile); err != nil {
		t.Errorf("expected backup file to exist, got: %v", err)
	}
}
