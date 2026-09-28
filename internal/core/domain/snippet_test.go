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
	"reflect"
	"testing"
)

func TestExtractPlaceholders(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		expected []string
	}{
		{
			name:     "no placeholders",
			command:  "df -h && free -m",
			expected: nil,
		},
		{
			name:     "double curly placeholders",
			command:  "tail -n {{lines}} {{path}}",
			expected: []string{"lines", "path"},
		},
		{
			name:     "angle bracket placeholders",
			command:  "systemctl restart <service>",
			expected: []string{"service"},
		},
		{
			name:     "mixed and duplicate placeholders",
			command:  "ping -c {{count}} <host> && ping -c {{count}} <gateway>",
			expected: []string{"count", "host", "gateway"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractPlaceholders(tt.command)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("ExtractPlaceholders(%q) = %v, want %v", tt.command, got, tt.expected)
			}
		})
	}
}

func TestInterpolateSnippet(t *testing.T) {
	cmd := "tail -n {{lines}} <logfile> --pid={{pid}}"
	params := map[string]string{
		"lines":   "50",
		"logfile": "/var/log/syslog",
		"pid":     "1234",
	}

	expected := "tail -n 50 /var/log/syslog --pid=1234"
	got := InterpolateSnippet(cmd, params)
	if got != expected {
		t.Errorf("InterpolateSnippet() = %q, want %q", got, expected)
	}
}

func TestDefaultSnippets(t *testing.T) {
	defaults := DefaultSnippets()
	if len(defaults) == 0 {
		t.Fatalf("expected non-empty default snippets")
	}

	for _, s := range defaults {
		if s.ID == "" {
			t.Errorf("snippet has empty ID: %+v", s)
		}
		if s.Name == "" {
			t.Errorf("snippet %s has empty Name", s.ID)
		}
		if s.Command == "" {
			t.Errorf("snippet %s has empty Command", s.ID)
		}
	}
}
