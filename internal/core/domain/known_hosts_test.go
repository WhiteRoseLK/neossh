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

func TestParseHostKeyMismatch(t *testing.T) {
	errText := `@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
@    WARNING: REMOTE HOST IDENTIFICATION HAS CHANGED!     @
@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@
IT IS POSSIBLE THAT SOMEONE IS DOING SOMETHING NASTY!
Someone could be eavesdropping on you right now (man-in-the-middle attack)!
It is also possible that a host key has just been changed.
The fingerprint for the ED25519 key sent by the remote host is
SHA256:4t7Zg+s9abcdefghijklmnopqrstuvwxyz123456789.
Please contact your system administrator.
Add correct host key in /home/user/.ssh/known_hosts to get rid of this message.
Offending ECDSA key in /home/user/.ssh/known_hosts:42
Host key for 192.168.1.50 has changed and you have requested strict checking.
Host key verification failed.`

	srv := &Server{
		Alias: "my-node",
		Host:  "192.168.1.50",
		Port:  22,
	}

	details := ParseHostKeyMismatch(errText, srv, "my-node")
	if details == nil {
		t.Fatal("expected non-nil details")
	}

	if details.OffendingFile != "/home/user/.ssh/known_hosts" {
		t.Errorf("expected offending file /home/user/.ssh/known_hosts, got %q", details.OffendingFile)
	}
	if details.OffendingLine != 42 {
		t.Errorf("expected offending line 42, got %d", details.OffendingLine)
	}
	if details.RemoteKeyType != "ED25519" {
		t.Errorf("expected remote key type ED25519, got %q", details.RemoteKeyType)
	}
	if details.RemoteFingerprint != "SHA256:4t7Zg+s9abcdefghijklmnopqrstuvwxyz123456789" {
		t.Errorf("expected remote fingerprint, got %q", details.RemoteFingerprint)
	}
	if details.TargetHost != "192.168.1.50" {
		t.Errorf("expected target host 192.168.1.50, got %q", details.TargetHost)
	}
	if details.TargetPort != 22 {
		t.Errorf("expected target port 22, got %d", details.TargetPort)
	}

	// Test non-mismatch error
	if ParseHostKeyMismatch("Connection refused", srv, "my-node") != nil {
		t.Error("expected nil for connection refused error")
	}
}
