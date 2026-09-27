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
	"time"
)

// BundleFileCategory represents the logical type of file stored in the bundle.
type BundleFileCategory string

const (
	BundleCategorySSHConfig  BundleFileCategory = "ssh_config"
	BundleCategorySSHInclude BundleFileCategory = "ssh_include"
	BundleCategoryMetadata   BundleFileCategory = "metadata"
	BundleCategorySettings   BundleFileCategory = "settings"
)

// BundleFileEntry represents a single archived file inside the bundle manifest.
type BundleFileEntry struct {
	ArchivePath    string             `json:"archive_path"`            // Relative path inside the tar archive (e.g. "ssh/config", "ssh/conf.d/hosts")
	Category       BundleFileCategory `json:"category"`                // ssh_config, ssh_include, metadata, settings
	RelativePath   string             `json:"relative_path,omitempty"` // Portable path relative to ~/.ssh or ~/.neossh
	OriginalPath   string             `json:"original_path,omitempty"` // Original path on exporting machine
	SizeBytes      int64              `json:"size_bytes"`
	FileMode       uint32             `json:"file_mode"`
	SHA256Checksum string             `json:"sha256,omitempty"`
}

// BundleManifest describes the contents and metadata of an exported configuration bundle.
type BundleManifest struct {
	Version       string            `json:"version"`            // Manifest schema version
	NeosshVersion string            `json:"neossh_version"`     // Version of neossh that created the bundle
	CreatedAt     time.Time         `json:"created_at"`         // Export timestamp
	Hostname      string            `json:"hostname,omitempty"` // Source machine hostname
	Sanitized     bool              `json:"sanitized"`          // True if IdentityFiles and secrets were stripped
	ServerCount   int               `json:"server_count"`       // Total number of servers included
	Files         []BundleFileEntry `json:"files"`              // List of archived configuration files
}

// ExportOptions specifies parameters for exporting a configuration bundle.
type ExportOptions struct {
	SSHConfigFile string // Path to main SSH config (e.g. ~/.ssh/config)
	NeosshDir     string // Path to neossh directory containing metadata.json & settings.json (e.g. ~/.neossh)
	OutputPath    string // Destination tar.gz file path (empty = auto-generate timestamped path)
	Sanitize      bool   // Strip IdentityFiles, passwords, and sensitive comments
}

// ImportOptions specifies parameters for restoring a configuration bundle.
type ImportOptions struct {
	BundlePath      string // Path to source .tar.gz bundle
	TargetSSHDir    string // Target directory for SSH configs (defaults to ~/.ssh)
	TargetNeosshDir string // Target directory for neossh configs (defaults to ~/.neossh)
	DryRun          bool   // Validate and preview without writing to disk
	Overwrite       bool   // Overwrite existing files directly without prompting
	CreateBackup    bool   // Create timestamped .bak backups of existing files before overwriting
}

// BundleSummary provides information about an export or import operation.
type BundleSummary struct {
	Manifest      BundleManifest `json:"manifest"`
	OutputPath    string         `json:"output_path,omitempty"`
	RestoredFiles []string       `json:"restored_files,omitempty"`
	BackupFiles   []string       `json:"backup_files,omitempty"`
	Warnings      []string       `json:"warnings,omitempty"`
}
