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
	"crypto/ed25519"
	"crypto/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/WhiteRoseLK/neossh/internal/core/domain"
	"github.com/rivo/tview"
	"golang.org/x/crypto/ssh"
)

func createTestSSHCertificate(t *testing.T, validAfter, validBefore uint64, keyID string, principals []string) []byte {
	t.Helper()

	_, caPriv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate CA key: %v", err)
	}
	caSigner, err := ssh.NewSignerFromKey(caPriv)
	if err != nil {
		t.Fatalf("failed to create CA signer: %v", err)
	}

	userPub, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate user key: %v", err)
	}
	sshUserPub, err := ssh.NewPublicKey(userPub)
	if err != nil {
		t.Fatalf("failed to convert user pubkey: %v", err)
	}

	cert := &ssh.Certificate{
		Key:             sshUserPub,
		Serial:          12345,
		CertType:        ssh.UserCert,
		KeyId:           keyID,
		ValidPrincipals: principals,
		ValidAfter:      validAfter,
		ValidBefore:     validBefore,
	}

	if err := cert.SignCert(rand.Reader, caSigner); err != nil {
		t.Fatalf("failed to sign test certificate: %v", err)
	}

	return ssh.MarshalAuthorizedKey(cert)
}

func TestServerDetails_SSH_Certificate_Display(t *testing.T) {
	tempDir := t.TempDir()
	now := time.Now()

	// 1. Valid certificate
	certFile := filepath.Join(tempDir, "valid-cert.pub")
	after := uint64(now.Add(-2 * time.Hour).Unix())
	before := uint64(now.Add(24 * time.Hour).Unix())
	certBytes := createTestSSHCertificate(t, after, before, "developer@prod", []string{"ubuntu", "dev"})
	if err := os.WriteFile(certFile, certBytes, 0o644); err != nil {
		t.Fatal(err)
	}

	server := domain.Server{
		Alias:           "cert-prod",
		Host:            "10.0.0.1",
		User:            "ubuntu",
		Port:            22,
		CertificateFile: certFile,
	}

	sd := NewServerDetails()
	sd.UpdateServer(server)

	detailsText := sd.TextView.GetText(true)

	if !strings.Contains(detailsText, "SSH Certificate Details") {
		t.Errorf("expected 'SSH Certificate Details' section in details view, got:\n%s", detailsText)
	}
	if !strings.Contains(detailsText, "Valid") {
		t.Errorf("expected 'Valid' badge in details view, got:\n%s", detailsText)
	}
	if !strings.Contains(detailsText, "developer@prod") {
		t.Errorf("expected Key ID 'developer@prod', got:\n%s", detailsText)
	}
	if !strings.Contains(detailsText, "Explicit CertificateFile") {
		t.Errorf("expected explicit source in details view, got:\n%s", detailsText)
	}

	// 2. Expired certificate (implicit discovery)
	keyFile := filepath.Join(tempDir, "id_expired")
	implicitCertFile := filepath.Join(tempDir, "id_expired-cert.pub")
	if err := os.WriteFile(keyFile, []byte("fake-key"), 0o600); err != nil {
		t.Fatal(err)
	}
	expAfter := uint64(now.Add(-10 * time.Hour).Unix())
	expBefore := uint64(now.Add(-1 * time.Hour).Unix())
	expCertBytes := createTestSSHCertificate(t, expAfter, expBefore, "expired-key-id", []string{"root"})
	if err := os.WriteFile(implicitCertFile, expCertBytes, 0o644); err != nil {
		t.Fatal(err)
	}

	implicitServer := domain.Server{
		Alias:         "implicit-host",
		Host:          "10.0.0.2",
		User:          "root",
		IdentityFiles: []string{keyFile},
	}

	sd.UpdateServer(implicitServer)
	implicitText := sd.TextView.GetText(true)

	if !strings.Contains(implicitText, "Implicit (<IdentityFile>-cert.pub)") {
		t.Errorf("expected implicit certificate source in details view, got:\n%s", implicitText)
	}
	if !strings.Contains(implicitText, "Expired") {
		t.Errorf("expected 'Expired' badge in details view, got:\n%s", implicitText)
	}
}

func TestBuildSSHCommand_WithCertificateFile(t *testing.T) {
	server := domain.Server{
		Alias:           "prod-server",
		Host:            "1.2.3.4",
		User:            "ubuntu",
		Port:            22,
		CertificateFile: "/path/to/my-cert.pub",
	}

	cmd := BuildSSHCommand(server)
	if !strings.Contains(cmd, "-o CertificateFile=/path/to/my-cert.pub") {
		t.Errorf("expected BuildSSHCommand to include CertificateFile, got: %s", cmd)
	}
}

func TestParseSSHCommand_WithCertificateFile(t *testing.T) {
	cmd := "ssh -o CertificateFile=~/.ssh/user-cert.pub ubuntu@1.2.3.4"
	server, err := ParseSSHCommand(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if server.CertificateFile != "~/.ssh/user-cert.pub" {
		t.Errorf("expected CertificateFile '~/.ssh/user-cert.pub', got %q", server.CertificateFile)
	}
}

func TestShowSSHErrorModal_WithExpiredCertificate(t *testing.T) {
	tempDir := t.TempDir()
	now := time.Now()

	expiredCertFile := filepath.Join(tempDir, "expired-cert.pub")
	expAfter := uint64(now.Add(-5 * time.Hour).Unix())
	expBefore := uint64(now.Add(-1 * time.Hour).Unix())
	expCertBytes := createTestSSHCertificate(t, expAfter, expBefore, "expired-user", []string{"root"})
	if err := os.WriteFile(expiredCertFile, expCertBytes, 0o644); err != nil {
		t.Fatal(err)
	}

	server := domain.Server{
		Alias:           "expired-host",
		Host:            "10.0.0.99",
		User:            "root",
		CertificateFile: expiredCertFile,
	}

	app := tview.NewApplication()
	serverList := NewServerList()
	serverList.UpdateServers([]domain.Server{server})

	ui := &tui{
		app:        app,
		serverList: serverList,
	}

	foundServer, ok := ui.findServerByAlias("expired-host")
	if !ok {
		t.Fatalf("expected to find server 'expired-host'")
	}
	cert := domain.InspectServerCertificate(foundServer)
	if cert == nil {
		t.Fatalf("expected certificate to be discovered")
	}
	if cert.Status != domain.CertStatusExpired {
		t.Errorf("expected expired certificate status, got %v", cert.Status)
	}

	// Also invoke showSSHErrorModal to ensure no nil pointer or panic
	ui.showSSHErrorModal("expired-host", "Permission denied (publickey).")
}
