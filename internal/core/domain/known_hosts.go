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
	"regexp"
	"strconv"
	"strings"
)

// KnownHostRecord represents an entry in ~/.ssh/known_hosts.
type KnownHostRecord struct {
	LineNumber  int    `json:"line_number"`
	HostPattern string `json:"host_pattern"`
	KeyType     string `json:"key_type"`
	Fingerprint string `json:"fingerprint"`
	Comment     string `json:"comment,omitempty"`
	IsHashed    bool   `json:"is_hashed"`
	RawLine     string `json:"raw_line"`
}

// HostKeyMismatchDetails contains parsed diagnostics from OpenSSH host key verification failure.
type HostKeyMismatchDetails struct {
	OffendingFile       string `json:"offending_file"`
	OffendingLine       int    `json:"offending_line"`
	OffendingKeyType    string `json:"offending_key_type"`
	TargetHost          string `json:"target_host"`
	TargetPort          int    `json:"target_port"`
	RemoteFingerprint   string `json:"remote_fingerprint"`
	RemoteKeyType       string `json:"remote_key_type"`
	ExistingFingerprint string `json:"existing_fingerprint"`
}

var (
	reOffendingKey = regexp.MustCompile(`(?i)Offending (?:[a-zA-Z0-9_-]+ )?key in ([^:\r\n]+):(\d+)`)
	reRemoteFP     = regexp.MustCompile(`(?i)fingerprint for the ([a-zA-Z0-9_-]+) key sent by the remote host is\s+(SHA256:[a-zA-Z0-9+/=]+)`)
	reHostChanged  = regexp.MustCompile(`(?i)Host key for ([^ ]+) has changed`)
)

// ParseHostKeyMismatch extracts diagnostic details from an SSH host key verification failure output.
func ParseHostKeyMismatch(errMsg string, srv *Server, alias string) *HostKeyMismatchDetails {
	lower := strings.ToLower(errMsg)
	if !strings.Contains(lower, "host key verification failed") &&
		!strings.Contains(lower, "remote host identification has changed") {
		return nil
	}

	details := &HostKeyMismatchDetails{
		TargetPort: 22,
	}

	// 1. Offending file & line
	if matches := reOffendingKey.FindStringSubmatch(errMsg); len(matches) == 3 {
		details.OffendingFile = strings.TrimSpace(matches[1])
		if line, err := strconv.Atoi(matches[2]); err == nil {
			details.OffendingLine = line
		}
	}

	// 2. Remote key fingerprint
	if matches := reRemoteFP.FindStringSubmatch(errMsg); len(matches) == 3 {
		details.RemoteKeyType = strings.TrimSpace(matches[1])
		details.RemoteFingerprint = strings.TrimSpace(matches[2])
	}

	// 3. Target host
	if matches := reHostChanged.FindStringSubmatch(errMsg); len(matches) == 2 {
		rawHost := strings.TrimSpace(matches[1])
		if strings.HasPrefix(rawHost, "[") && strings.Contains(rawHost, "]:") {
			parts := strings.Split(rawHost, "]:")
			details.TargetHost = strings.TrimPrefix(parts[0], "[")
			if p, err := strconv.Atoi(parts[1]); err == nil {
				details.TargetPort = p
			}
		} else {
			details.TargetHost = rawHost
		}
	}

	// 4. Enrich from server if available
	if srv != nil {
		if details.TargetHost == "" {
			if srv.Host != "" {
				details.TargetHost = srv.Host
			} else {
				details.TargetHost = srv.Alias
			}
		}
		if srv.Port > 0 && details.TargetPort == 22 {
			details.TargetPort = srv.Port
		}
	} else if details.TargetHost == "" && alias != "" {
		details.TargetHost = alias
	}

	return details
}
