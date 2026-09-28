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
	"strings"
)

// Snippet represents a reusable shell command snippet in the library catalog.
type Snippet struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Command     string   `json:"command"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

var (
	doubleCurlyRegex  = regexp.MustCompile(`\{\{([a-zA-Z0-9_\-]+)\}\}`)
	angleBracketRegex = regexp.MustCompile(`<([a-zA-Z0-9_\-]+)>`)
)

// ExtractPlaceholders extracts all parameter placeholder names from a command string.
// Supports both {{name}} and <name> placeholder syntax.
func ExtractPlaceholders(command string) []string {
	var result []string
	seen := make(map[string]bool)

	// Match {{placeholder}}
	matches := doubleCurlyRegex.FindAllStringSubmatch(command, -1)
	for _, m := range matches {
		if len(m) > 1 && !seen[m[1]] {
			seen[m[1]] = true
			result = append(result, m[1])
		}
	}

	// Match <placeholder>
	matchesAngle := angleBracketRegex.FindAllStringSubmatch(command, -1)
	for _, m := range matchesAngle {
		if len(m) > 1 && !seen[m[1]] {
			seen[m[1]] = true
			result = append(result, m[1])
		}
	}

	return result
}

// InterpolateSnippet replaces placeholders in a command string with provided values.
// Supports {{param}} and <param> syntaxes.
func InterpolateSnippet(command string, params map[string]string) string {
	res := command
	for k, v := range params {
		res = strings.ReplaceAll(res, "{{"+k+"}}", v)
		res = strings.ReplaceAll(res, "<"+k+">", v)
	}
	return res
}

// DefaultSnippets returns standard built-in command snippets for common system inspection tasks.
func DefaultSnippets() []Snippet {
	return []Snippet{
		{
			ID:          "sys-uptime",
			Name:        "System Info & Uptime",
			Command:     "uname -a && uptime",
			Description: "Displays OS kernel release, uptime, and load averages.",
			Tags:        []string{"system", "quick"},
		},
		{
			ID:          "disk-usage",
			Name:        "Disk Usage (Human Readable)",
			Command:     "df -h",
			Description: "Displays filesystem disk space usage in human-readable units.",
			Tags:        []string{"storage", "system"},
		},
		{
			ID:          "memory-usage",
			Name:        "Memory Usage",
			Command:     "free -h 2>/dev/null || vm_stat",
			Description: "Displays total and available physical memory.",
			Tags:        []string{"memory", "performance"},
		},
		{
			ID:          "top-cpu",
			Name:        "Top 10 CPU Processes",
			Command:     "ps aux --sort=-%cpu | head -n 11",
			Description: "Lists the top 10 processes consuming the most CPU time.",
			Tags:        []string{"performance", "processes"},
		},
		{
			ID:          "listening-ports",
			Name:        "Listening Ports",
			Command:     "ss -tuln 2>/dev/null || netstat -tuln",
			Description: "Inspects all TCP and UDP listening ports and sockets.",
			Tags:        []string{"network", "security"},
		},
		{
			ID:          "docker-ps",
			Name:        "Docker Containers",
			Command:     "docker ps -a",
			Description: "Lists all running and stopped Docker containers.",
			Tags:        []string{"containers", "docker"},
		},
		{
			ID:          "tail-journal",
			Name:        "Tail System Journal",
			Command:     "journalctl -n 50 -f",
			Description: "Follows systemd journal logs live (last 50 entries).",
			Tags:        []string{"logs", "systemd"},
		},
		{
			ID:          "tail-file",
			Name:        "Tail Log File",
			Command:     "tail -n 100 -f {{logfile}}",
			Description: "Follows a specific log file with customizable path placeholder.",
			Tags:        []string{"logs"},
		},
	}
}
