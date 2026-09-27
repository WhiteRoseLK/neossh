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
	"regexp"
	"strings"

	"github.com/kevinburke/ssh_config"
)

var (
	// tagCommentRegex extracts tag strings from comments.
	// Matches patterns such as:
	//   # tags: prod, database
	//   # tag: prod, database
	//   # tags:prod,database
	//   # tags = prod, database
	//   # Added by neossh # tags: prod, database
	//   # [tags: prod, database]
	tagCommentRegex = regexp.MustCompile(`(?i)(?:^|[\s#;|,(\[])tags?\s*[:=]\s*([^#;\r\n]+)`)

	// tagPatternRegex matches the tag prefix and value within a comment for replacement/removal.
	tagPatternRegex = regexp.MustCompile(`(?i)(?:^|[\s#;|,(\[])(tags?\s*[:=]\s*[^#;\r\n]+)`)
)

// extractTagsFromComment parses a comment string and returns any tag labels found.
func extractTagsFromComment(comment string) []string {
	matches := tagCommentRegex.FindAllStringSubmatch(comment, -1)
	if len(matches) == 0 {
		return nil
	}

	var tags []string
	seen := make(map[string]bool)
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		raw := match[1]
		raw = strings.Trim(raw, " \t\r\n)]}")
		var items []string
		if strings.Contains(raw, ",") {
			items = strings.Split(raw, ",")
		} else {
			items = strings.Fields(raw)
		}
		for _, item := range items {
			item = strings.Trim(item, " \"'\t\r\n")
			if item != "" && !seen[item] {
				seen[item] = true
				tags = append(tags, item)
			}
		}
	}
	return tags
}

// extractHostTags extracts all tags found across a host's EOLComment and child nodes.
func extractHostTags(host *ssh_config.Host) []string {
	var allTags []string
	seen := make(map[string]bool)

	addTags := func(tags []string) {
		for _, t := range tags {
			if t != "" && !seen[t] {
				seen[t] = true
				allTags = append(allTags, t)
			}
		}
	}

	if host.EOLComment != "" {
		addTags(extractTagsFromComment(host.EOLComment))
	}

	for _, node := range host.Nodes {
		switch n := node.(type) {
		case *ssh_config.Empty:
			if n.Comment != "" {
				addTags(extractTagsFromComment(n.Comment))
			}
		case *ssh_config.KV:
			if n.Comment != "" {
				addTags(extractTagsFromComment(n.Comment))
			}
		}
	}

	return allTags
}

// formatTagsComment formats tags into standard comment string format.
func formatTagsComment(tags []string) string {
	if len(tags) == 0 {
		return ""
	}
	return "tags: " + strings.Join(tags, ", ")
}

// replaceTagsInComment replaces the existing tags definition within comment with newTags.
func replaceTagsInComment(comment string, newTags []string) string {
	formatted := formatTagsComment(newTags)
	match := tagPatternRegex.FindStringSubmatchIndex(comment)
	if len(match) >= 4 {
		// Replace the tag portion
		subStart := match[2]
		subEnd := match[3]

		matched := comment[subStart:subEnd]
		trailingSpaces := ""
		trimmed := strings.TrimRight(matched, " \t")
		if len(trimmed) < len(matched) {
			trailingSpaces = matched[len(trimmed):]
		}

		updated := comment[:subStart] + formatted + trailingSpaces + comment[subEnd:]

		// Remove any subsequent tag occurrences if there were multiple
		for {
			m := tagPatternRegex.FindStringSubmatchIndex(updated[subStart+len(formatted):])
			if len(m) < 4 {
				break
			}
			offset := subStart + len(formatted)
			mStart := offset + m[0]
			mEnd := offset + m[1]
			updated = strings.TrimRight(updated[:mStart], " #;,\t") + updated[mEnd:]
		}
		return updated
	}

	if strings.TrimSpace(comment) == "" {
		return " " + formatted
	}
	return comment + " # " + formatted
}

// removeTagsFromComment removes all tag definitions from a comment string,
// cleaning up leftover delimiters and whitespace.
func removeTagsFromComment(comment string) string {
	result := comment
	for {
		m := tagPatternRegex.FindStringSubmatchIndex(result)
		if len(m) < 2 {
			break
		}
		left := strings.TrimRight(result[:m[0]], " #;,\t")
		right := strings.TrimLeft(result[m[1]:], " #;,\t")
		switch {
		case left == "" && right == "":
			return ""
		case left != "" && right != "":
			result = left + " # " + right
		case left != "":
			result = left
		default:
			result = " " + right
		}
	}
	if strings.Trim(result, " #\t\r\n") == "" {
		return ""
	}
	return result
}

// updateHostTags updates or removes tags on an SSH config host entry.
func (r *Repository) updateHostTags(host *ssh_config.Host, newTags []string) {
	// 1. Check if an Empty node in host.Nodes contains tags (block comment)
	emptyIdx := -1
	for i, node := range host.Nodes {
		if empty, ok := node.(*ssh_config.Empty); ok {
			if len(extractTagsFromComment(empty.Comment)) > 0 {
				emptyIdx = i
				break
			}
		}
	}

	eolHasTags := len(extractTagsFromComment(host.EOLComment)) > 0

	if emptyIdx != -1 {
		if len(newTags) > 0 {
			empty := host.Nodes[emptyIdx].(*ssh_config.Empty)
			empty.Comment = replaceTagsInComment(empty.Comment, newTags)
		} else {
			host.Nodes = append(host.Nodes[:emptyIdx], host.Nodes[emptyIdx+1:]...)
		}
		if eolHasTags {
			host.EOLComment = removeTagsFromComment(host.EOLComment)
			if strings.TrimSpace(host.EOLComment) == "" {
				host.EOLComment = ""
				host.SpaceBeforeComment = ""
			}
		}
		return
	}

	// 2. Check if a KV node contains tags in its comment
	kvIdx := -1
	for i, node := range host.Nodes {
		if kv, ok := node.(*ssh_config.KV); ok {
			if len(extractTagsFromComment(kv.Comment)) > 0 {
				kvIdx = i
				break
			}
		}
	}

	if kvIdx != -1 {
		kv := host.Nodes[kvIdx].(*ssh_config.KV)
		if len(newTags) > 0 {
			kv.Comment = replaceTagsInComment(kv.Comment, newTags)
		} else {
			kv.Comment = removeTagsFromComment(kv.Comment)
			if strings.TrimSpace(kv.Comment) == "" {
				kv.Comment = ""
				kv.SpaceAfterValue = ""
			}
		}
		return
	}

	// 3. Handle EOLComment on the Host line
	if eolHasTags {
		if len(newTags) > 0 {
			host.EOLComment = replaceTagsInComment(host.EOLComment, newTags)
		} else {
			host.EOLComment = removeTagsFromComment(host.EOLComment)
			if strings.TrimSpace(host.EOLComment) == "" {
				host.EOLComment = ""
				host.SpaceBeforeComment = ""
			}
		}
		return
	}

	// 4. No tags currently exist on the host
	if len(newTags) > 0 {
		if host.EOLComment == "" || strings.TrimSpace(host.EOLComment) == "Added by neossh" {
			host.EOLComment = " " + formatTagsComment(newTags)
			if host.SpaceBeforeComment == "" {
				host.SpaceBeforeComment = strings.Repeat(" ", 4)
			}
		} else {
			host.EOLComment = host.EOLComment + " # " + formatTagsComment(newTags)
		}
	}
}

var preConnectCommentRegex = regexp.MustCompile(`(?i)(?:^|[\s#;|,(\[])(?:pre-connect|pre-ssh|hook)\s*[:=]\s*([^#;\r\n]+)`)

func extractPreConnectFromComment(comment string) string {
	match := preConnectCommentRegex.FindStringSubmatch(comment)
	if len(match) >= 2 {
		return strings.Trim(match[1], " \t\r\n)]}\"'")
	}
	return ""
}

func extractHostPreConnectCommand(host *ssh_config.Host) string {
	if cmd := extractPreConnectFromComment(host.EOLComment); cmd != "" {
		return cmd
	}
	for _, node := range host.Nodes {
		switch n := node.(type) {
		case *ssh_config.Empty:
			if cmd := extractPreConnectFromComment(n.Comment); cmd != "" {
				return cmd
			}
		case *ssh_config.KV:
			if cmd := extractPreConnectFromComment(n.Comment); cmd != "" {
				return cmd
			}
		}
	}
	return ""
}

var certificateCommandCommentRegex = regexp.MustCompile(`(?i)(?:^|[\s#;|,(\[])(?:certificate-command|cert-command|cert-renew)\s*[:=]\s*([^#;\r\n]+)`)

func extractCertificateCommandFromComment(comment string) string {
	match := certificateCommandCommentRegex.FindStringSubmatch(comment)
	if len(match) >= 2 {
		return strings.Trim(match[1], " \t\r\n)]}\"'")
	}
	return ""
}

func extractHostCertificateCommand(host *ssh_config.Host) string {
	if cmd := extractCertificateCommandFromComment(host.EOLComment); cmd != "" {
		return cmd
	}
	for _, node := range host.Nodes {
		switch n := node.(type) {
		case *ssh_config.Empty:
			if cmd := extractCertificateCommandFromComment(n.Comment); cmd != "" {
				return cmd
			}
		case *ssh_config.KV:
			if cmd := extractCertificateCommandFromComment(n.Comment); cmd != "" {
				return cmd
			}
		}
	}
	return ""
}
