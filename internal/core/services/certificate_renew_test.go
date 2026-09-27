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
	"crypto/ed25519"
	"crypto/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/WhiteRoseLK/neossh/internal/core/domain"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
)

func createTestSSHCertificate(t *testing.T, validAfter, validBefore uint64, keyID string) []byte {
	t.Helper()

	caPub, caPriv, err := ed25519.GenerateKey(rand.Reader)
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
		Serial:          1001,
		CertType:        ssh.UserCert,
		KeyId:           keyID,
		ValidPrincipals: []string{"dev", "ubuntu"},
		ValidAfter:      validAfter,
		ValidBefore:     validBefore,
		Permissions: ssh.Permissions{
			Extensions: map[string]string{
				"permit-pty": "",
			},
		},
	}

	if err := cert.SignCert(rand.Reader, caSigner); err != nil {
		t.Fatalf("failed to sign test certificate: %v", err)
	}
	_ = caPub

	return ssh.MarshalAuthorizedKey(cert)
}

func TestServerService_EnsureValidCertificate(t *testing.T) {
	tempDir := t.TempDir()

	now := time.Now()
	validCertBytes := createTestSSHCertificate(t,
		uint64(now.Add(-1*time.Hour).Unix()),
		uint64(now.Add(24*time.Hour).Unix()),
		"valid-key-id",
	)
	expiredCertBytes := createTestSSHCertificate(t,
		uint64(now.Add(-48*time.Hour).Unix()),
		uint64(now.Add(-2*time.Hour).Unix()),
		"expired-key-id",
	)

	validSourcePath := filepath.Join(tempDir, "source-valid-cert.pub")
	if err := os.WriteFile(validSourcePath, validCertBytes, 0o600); err != nil {
		t.Fatalf("failed to write source valid cert: %v", err)
	}

	expiredSourcePath := filepath.Join(tempDir, "source-expired-cert.pub")
	if err := os.WriteFile(expiredSourcePath, expiredCertBytes, 0o600); err != nil {
		t.Fatalf("failed to write source expired cert: %v", err)
	}

	t.Run("no certificate command configured", func(t *testing.T) {
		repo := &mockServerRepository{
			servers: []domain.Server{
				{Alias: "srv-nocmd", Host: "10.0.0.1"},
			},
		}

		hookRun := false
		svc := &serverService{
			logger:           zap.NewNop().Sugar(),
			serverRepository: repo,
			newHookCommand: func(cmdStr string) *exec.Cmd {
				hookRun = true
				cs := []string{"-test.run=TestHelperProcess", "--", "hook-success"}
				cmd := exec.Command(os.Args[0], cs...)
				cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
				return cmd
			},
		}

		err := svc.ensureValidCertificate("srv-nocmd")
		if err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
		if hookRun {
			t.Errorf("hook should not run when CertificateCommand is empty")
		}
	})

	t.Run("certificate already valid skips command execution", func(t *testing.T) {
		certPath := filepath.Join(tempDir, "srv1-cert.pub")
		if err := os.WriteFile(certPath, validCertBytes, 0o600); err != nil {
			t.Fatalf("failed to write cert: %v", err)
		}

		repo := &mockServerRepository{
			servers: []domain.Server{
				{
					Alias:              "srv1",
					Host:               "10.0.0.2",
					CertificateFile:    certPath,
					CertificateCommand: "step ssh login %u@%h",
				},
			},
		}

		hookRun := false
		svc := &serverService{
			logger:           zap.NewNop().Sugar(),
			serverRepository: repo,
			newHookCommand: func(cmdStr string) *exec.Cmd {
				hookRun = true
				cs := []string{"-test.run=TestHelperProcess", "--", "hook-success"}
				cmd := exec.Command(os.Args[0], cs...)
				cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
				return cmd
			},
		}

		err := svc.ensureValidCertificate("srv1")
		if err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
		if hookRun {
			t.Errorf("expected certificate command NOT to run when certificate is valid")
		}
	})

	t.Run("missing certificate triggers command and succeeds upon valid cert creation", func(t *testing.T) {
		certPath := filepath.Join(tempDir, "srv-missing-cert.pub")

		repo := &mockServerRepository{
			servers: []domain.Server{
				{
					Alias:              "srv-missing",
					Host:               "10.0.0.3",
					User:               "root",
					Port:               2222,
					CertificateFile:    certPath,
					CertificateCommand: "vault login %h",
				},
			},
		}

		hookRun := false
		var passedCmd string
		svc := &serverService{
			logger:           zap.NewNop().Sugar(),
			serverRepository: repo,
			newHookCommand: func(cmdStr string) *exec.Cmd {
				hookRun = true
				passedCmd = cmdStr
				cs := []string{"-test.run=TestHelperProcess", "--", "copy-file", validSourcePath, certPath}
				cmd := exec.Command(os.Args[0], cs...)
				cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
				return cmd
			},
		}

		err := svc.ensureValidCertificate("srv-missing")
		if err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
		if !hookRun {
			t.Errorf("expected certificate command to run when certificate is missing")
		}
		if passedCmd != "vault login 10.0.0.3" {
			t.Errorf("expected command interpolation %q, got %q", "vault login 10.0.0.3", passedCmd)
		}

		// Ensure file was created
		if _, err := os.Stat(certPath); err != nil {
			t.Errorf("expected certificate file to exist on disk after renewal: %v", err)
		}
	})

	t.Run("expired certificate triggers command and updates to valid", func(t *testing.T) {
		certPath := filepath.Join(tempDir, "srv-expired-cert.pub")
		if err := os.WriteFile(certPath, expiredCertBytes, 0o600); err != nil {
			t.Fatalf("failed to write initial expired cert: %v", err)
		}

		repo := &mockServerRepository{
			servers: []domain.Server{
				{
					Alias:              "srv-expired",
					Host:               "10.0.0.4",
					CertificateFile:    certPath,
					CertificateCommand: "step ssh login",
				},
			},
		}

		hookRun := false
		svc := &serverService{
			logger:           zap.NewNop().Sugar(),
			serverRepository: repo,
			newHookCommand: func(cmdStr string) *exec.Cmd {
				hookRun = true
				cs := []string{"-test.run=TestHelperProcess", "--", "copy-file", validSourcePath, certPath}
				cmd := exec.Command(os.Args[0], cs...)
				cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
				return cmd
			},
		}

		err := svc.ensureValidCertificate("srv-expired")
		if err != nil {
			t.Fatalf("expected nil error after successful renewal, got: %v", err)
		}
		if !hookRun {
			t.Errorf("expected certificate command to run when certificate was expired")
		}
	})

	t.Run("command failure returns clear error with stderr", func(t *testing.T) {
		certPath := filepath.Join(tempDir, "srv-fail-cert.pub")

		repo := &mockServerRepository{
			servers: []domain.Server{
				{
					Alias:              "srv-fail",
					Host:               "10.0.0.5",
					CertificateFile:    certPath,
					CertificateCommand: "step ssh login %h",
				},
			},
		}

		svc := &serverService{
			logger:           zap.NewNop().Sugar(),
			serverRepository: repo,
			newHookCommand: func(cmdStr string) *exec.Cmd {
				cs := []string{"-test.run=TestHelperProcess", "--", "cert-fail"}
				cmd := exec.Command(os.Args[0], cs...)
				cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
				return cmd
			},
		}

		err := svc.ensureValidCertificate("srv-fail")
		if err == nil {
			t.Fatalf("expected error when certificate command fails")
		}
		if !strings.Contains(err.Error(), "certificate command failed") {
			t.Errorf("expected error to mention 'certificate command failed', got: %v", err)
		}
		if !strings.Contains(err.Error(), "step-ca: token expired or unauthorized") {
			t.Errorf("expected error to include stderr message, got: %v", err)
		}
	})

	t.Run("command succeeds but certificate file not found on disk", func(t *testing.T) {
		certPath := filepath.Join(tempDir, "srv-missing-after-cert.pub")

		repo := &mockServerRepository{
			servers: []domain.Server{
				{
					Alias:              "srv-missing-after",
					Host:               "10.0.0.6",
					CertificateFile:    certPath,
					CertificateCommand: "dummy-command",
				},
			},
		}

		svc := &serverService{
			logger:           zap.NewNop().Sugar(),
			serverRepository: repo,
			newHookCommand: func(cmdStr string) *exec.Cmd {
				// Succeed without creating the cert file
				cs := []string{"-test.run=TestHelperProcess", "--", "hook-success"}
				cmd := exec.Command(os.Args[0], cs...)
				cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
				return cmd
			},
		}

		err := svc.ensureValidCertificate("srv-missing-after")
		if err == nil {
			t.Fatalf("expected error when cert file was not produced")
		}
		if !strings.Contains(err.Error(), "was not found on disk") {
			t.Errorf("expected error to mention 'was not found on disk', got: %v", err)
		}
	})

	t.Run("command succeeds but produced certificate is still expired", func(t *testing.T) {
		certPath := filepath.Join(tempDir, "srv-still-expired-cert.pub")

		repo := &mockServerRepository{
			servers: []domain.Server{
				{
					Alias:              "srv-still-expired",
					Host:               "10.0.0.7",
					CertificateFile:    certPath,
					CertificateCommand: "renew-command",
				},
			},
		}

		svc := &serverService{
			logger:           zap.NewNop().Sugar(),
			serverRepository: repo,
			newHookCommand: func(cmdStr string) *exec.Cmd {
				cs := []string{"-test.run=TestHelperProcess", "--", "copy-file", expiredSourcePath, certPath}
				cmd := exec.Command(os.Args[0], cs...)
				cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
				return cmd
			},
		}

		err := svc.ensureValidCertificate("srv-still-expired")
		if err == nil {
			t.Fatalf("expected error when produced cert is expired")
		}
		if !strings.Contains(err.Error(), "is expired") {
			t.Errorf("expected error to mention 'is expired', got: %v", err)
		}
	})

	t.Run("command succeeds but produced file is not a valid certificate", func(t *testing.T) {
		certPath := filepath.Join(tempDir, "srv-invalid-cert.pub")
		invalidContentPath := filepath.Join(tempDir, "invalid.pub")
		if err := os.WriteFile(invalidContentPath, []byte("this is not an ssh certificate"), 0o600); err != nil {
			t.Fatalf("failed to write invalid cert: %v", err)
		}

		repo := &mockServerRepository{
			servers: []domain.Server{
				{
					Alias:              "srv-invalid",
					Host:               "10.0.0.8",
					CertificateFile:    certPath,
					CertificateCommand: "renew-command",
				},
			},
		}

		svc := &serverService{
			logger:           zap.NewNop().Sugar(),
			serverRepository: repo,
			newHookCommand: func(cmdStr string) *exec.Cmd {
				cs := []string{"-test.run=TestHelperProcess", "--", "copy-file", invalidContentPath, certPath}
				cmd := exec.Command(os.Args[0], cs...)
				cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
				return cmd
			},
		}

		err := svc.ensureValidCertificate("srv-invalid")
		if err == nil {
			t.Fatalf("expected error when produced cert is not a valid certificate")
		}
		if !strings.Contains(err.Error(), "is invalid") {
			t.Errorf("expected error to mention 'is invalid', got: %v", err)
		}
	})
}

func TestServerService_SSH_CertificateRenewalIntegration(t *testing.T) {
	tempDir := t.TempDir()
	now := time.Now()
	validCertBytes := createTestSSHCertificate(t,
		uint64(now.Add(-1*time.Hour).Unix()),
		uint64(now.Add(24*time.Hour).Unix()),
		"integration-key-id",
	)
	validSourcePath := filepath.Join(tempDir, "valid-src.pub")
	if err := os.WriteFile(validSourcePath, validCertBytes, 0o600); err != nil {
		t.Fatalf("failed to write valid cert: %v", err)
	}

	certPath := filepath.Join(tempDir, "srv-integ-cert.pub")

	repo := &mockServerRepository{
		servers: []domain.Server{
			{
				Alias:              "integ-srv",
				Host:               "10.0.0.9",
				CertificateFile:    certPath,
				CertificateCommand: "step ssh login",
			},
		},
	}

	certCommandRun := false
	sshCommandRun := false

	svc := &serverService{
		logger:           zap.NewNop().Sugar(),
		serverRepository: repo,
		newHookCommand: func(cmdStr string) *exec.Cmd {
			certCommandRun = true
			cs := []string{"-test.run=TestHelperProcess", "--", "copy-file", validSourcePath, certPath}
			cmd := exec.Command(os.Args[0], cs...)
			cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
			return cmd
		},
		newSSHCommand: func(alias string) *exec.Cmd {
			sshCommandRun = true
			cs := []string{"-test.run=TestHelperProcess", "--", "success", alias}
			cmd := exec.Command(os.Args[0], cs...)
			cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
			return cmd
		},
	}

	// First run: cert is missing -> cert command should run and SSH should succeed
	err := svc.SSH("integ-srv")
	if err != nil {
		t.Fatalf("expected SSH to succeed, got: %v", err)
	}
	if !certCommandRun {
		t.Errorf("expected cert command to run on first connection")
	}
	if !sshCommandRun {
		t.Errorf("expected SSH command to run after renewal")
	}

	// Second run: cert is now valid -> cert command should NOT run, SSH should run
	certCommandRun = false
	sshCommandRun = false
	err = svc.SSH("integ-srv")
	if err != nil {
		t.Fatalf("expected second SSH to succeed, got: %v", err)
	}
	if certCommandRun {
		t.Errorf("expected cert command NOT to run when cert is still valid")
	}
	if !sshCommandRun {
		t.Errorf("expected SSH command to run on second connection")
	}
}
