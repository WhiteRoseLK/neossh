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
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/WhiteRoseLK/neossh/internal/core/domain"
	"github.com/WhiteRoseLK/neossh/internal/core/ports"
	"go.uber.org/zap"
)

const (
	CurrentBundleVersion = "1.0.0"
	ManifestFileName     = "manifest.json"
	MaxArchiveFileSize   = 50 * 1024 * 1024 // 50MB limit per file to prevent zip-bomb
)

var (
	reIdentityDirective  = regexp.MustCompile(`(?i)^\s*IdentityFile\s+(.*)$`)
	reSensitiveComment   = regexp.MustCompile(`(?i)#.*(password|token|secret|api[_-]?key|private_key)\s*[:=].*`)
	reSecretTokenInCmd   = regexp.MustCompile(`(?i)(token|password|secret|key)=([^\s"']+)`)
	reHomePrefixUsername = regexp.MustCompile(`^(/Users/[^/]+|/home/[^/]+)(/.*)?$`)
)

type bundleService struct {
	serverRepo ports.ServerRepository
	logger     *zap.SugaredLogger
	version    string
}

// NewBundleService creates a new BundleService instance.
func NewBundleService(serverRepo ports.ServerRepository, logger *zap.SugaredLogger, version string) ports.BundleService {
	if version == "" {
		version = "develop"
	}
	return &bundleService{
		serverRepo: serverRepo,
		logger:     logger,
		version:    version,
	}
}

// Export creates a tar.gz bundle of SSH configs and neossh metadata with optional sanitation.
func (s *bundleService) Export(opts domain.ExportOptions) (*domain.BundleSummary, error) {
	sshConfigPath, neosshDir := s.resolveExportPaths(opts)

	outputPath := strings.TrimSpace(opts.OutputPath)
	if outputPath == "" {
		ts := time.Now().Format("20060102-150405")
		outputPath = fmt.Sprintf("neossh-bundle-%s.tar.gz", ts)
	}
	outputPath = filepath.Clean(domain.ExpandTilde(outputPath))

	// Ensure destination directory exists
	if dir := filepath.Dir(outputPath); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return nil, fmt.Errorf("create output directory %q: %w", dir, err)
		}
	}

	// Collect config files
	configFiles := s.gatherConfigFiles(sshConfigPath)

	var warnings []string
	var fileEntries []domain.BundleFileEntry
	archiveBuffers := make(map[string][]byte)

	sshBaseDir := filepath.Dir(sshConfigPath)

	// Process SSH config files
	for i, cfgPath := range configFiles {
		cleaned := filepath.Clean(cfgPath)
		data, readErr := os.ReadFile(cleaned) // #nosec G304: reading resolved ssh config file
		if readErr != nil {
			if os.IsNotExist(readErr) && i > 0 {
				warnings = append(warnings, fmt.Sprintf("Included file not found, skipping: %s", cfgPath))
				continue
			}
			return nil, fmt.Errorf("read config file %q: %w", cfgPath, readErr)
		}

		category := domain.BundleCategorySSHInclude
		relPath := ""
		archivePath := ""

		if i == 0 {
			category = domain.BundleCategorySSHConfig
			relPath = "config"
			archivePath = "ssh/config"
		} else {
			if rel, err := filepath.Rel(sshBaseDir, cleaned); err == nil && !strings.HasPrefix(rel, "..") {
				relPath = rel
				archivePath = "ssh/" + filepath.ToSlash(rel)
			} else {
				base := filepath.Base(cleaned)
				archivePath = fmt.Sprintf("ssh/includes/%s", base)
				relPath = fmt.Sprintf("includes/%s", base)
				warnings = append(warnings, fmt.Sprintf("Included file %q is outside SSH dir, packaged as %q", cfgPath, archivePath))
			}
		}

		if opts.Sanitize {
			data = sanitizeSSHConfigFile(data)
		}

		checksum := sha256.Sum256(data)
		fileEntries = append(fileEntries, domain.BundleFileEntry{
			ArchivePath:    archivePath,
			Category:       category,
			RelativePath:   relPath,
			OriginalPath:   sanitizePathString(cleaned, opts.Sanitize),
			SizeBytes:      int64(len(data)),
			FileMode:       0o600,
			SHA256Checksum: hex.EncodeToString(checksum[:]),
		})
		archiveBuffers[archivePath] = data
	}

	// Process metadata.json
	metaPath := filepath.Join(neosshDir, "metadata.json")
	if s.serverRepo != nil {
		if customMeta := s.serverRepo.GetMetadataFile(); customMeta != "" {
			metaPath = customMeta
		}
	}
	if data, err := os.ReadFile(filepath.Clean(metaPath)); err == nil { // #nosec G304: user metadata file read
		if opts.Sanitize {
			data = sanitizeMetadataJSON(data)
		}
		checksum := sha256.Sum256(data)
		archivePath := "neossh/metadata.json"
		fileEntries = append(fileEntries, domain.BundleFileEntry{
			ArchivePath:    archivePath,
			Category:       domain.BundleCategoryMetadata,
			RelativePath:   "metadata.json",
			OriginalPath:   sanitizePathString(metaPath, opts.Sanitize),
			SizeBytes:      int64(len(data)),
			FileMode:       0o600,
			SHA256Checksum: hex.EncodeToString(checksum[:]),
		})
		archiveBuffers[archivePath] = data
	}

	// Process settings.json
	settingsPath := filepath.Join(neosshDir, "settings.json")
	if data, err := os.ReadFile(filepath.Clean(settingsPath)); err == nil { // #nosec G304: user settings file read
		if opts.Sanitize {
			data = sanitizeSettingsJSON(data)
		}
		checksum := sha256.Sum256(data)
		archivePath := "neossh/settings.json"
		fileEntries = append(fileEntries, domain.BundleFileEntry{
			ArchivePath:    archivePath,
			Category:       domain.BundleCategorySettings,
			RelativePath:   "settings.json",
			OriginalPath:   sanitizePathString(settingsPath, opts.Sanitize),
			SizeBytes:      int64(len(data)),
			FileMode:       0o600,
			SHA256Checksum: hex.EncodeToString(checksum[:]),
		})
		archiveBuffers[archivePath] = data
	}

	// Count servers
	serverCount := s.countServers()

	hostname, _ := os.Hostname()
	if opts.Sanitize {
		hostname = "sanitized-host"
	}

	manifest := domain.BundleManifest{
		Version:       CurrentBundleVersion,
		NeosshVersion: s.version,
		CreatedAt:     time.Now().UTC(),
		Hostname:      hostname,
		Sanitized:     opts.Sanitize,
		ServerCount:   serverCount,
		Files:         fileEntries,
	}

	manifestBytes, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal bundle manifest: %w", err)
	}

	// Create output tar.gz file
	outFile, err := os.OpenFile(filepath.Clean(outputPath), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600) // #nosec G304: writing user bundle
	if err != nil {
		return nil, fmt.Errorf("create bundle file %q: %w", outputPath, err)
	}
	defer func() {
		_ = outFile.Close()
	}()

	gzipWriter := gzip.NewWriter(outFile)
	defer func() {
		_ = gzipWriter.Close()
	}()

	tarWriter := tar.NewWriter(gzipWriter)
	defer func() {
		_ = tarWriter.Close()
	}()

	// Write manifest.json as first entry
	if err := writeTarEntry(tarWriter, ManifestFileName, manifestBytes, 0o600); err != nil {
		return nil, fmt.Errorf("write manifest to archive: %w", err)
	}

	// Write all files into archive
	for _, entry := range fileEntries {
		data := archiveBuffers[entry.ArchivePath]
		if err := writeTarEntry(tarWriter, entry.ArchivePath, data, entry.FileMode); err != nil {
			return nil, fmt.Errorf("write %q to archive: %w", entry.ArchivePath, err)
		}
	}

	s.logger.Infow("exported neossh bundle successfully",
		"outputPath", outputPath,
		"files", len(fileEntries),
		"servers", serverCount,
		"sanitized", opts.Sanitize,
	)

	return &domain.BundleSummary{
		Manifest:   manifest,
		OutputPath: outputPath,
		Warnings:   warnings,
	}, nil
}

// Verify inspects and validates a bundle archive without extracting files.
func (s *bundleService) Verify(bundlePath string) (*domain.BundleSummary, error) {
	manifest, archiveFiles, warnings, err := s.readAndValidateBundle(bundlePath)
	if err != nil {
		return nil, err
	}

	_ = archiveFiles
	return &domain.BundleSummary{
		Manifest:   *manifest,
		OutputPath: bundlePath,
		Warnings:   warnings,
	}, nil
}

// Import restores configuration files and metadata from a bundle into target directories.
func (s *bundleService) Import(opts domain.ImportOptions) (*domain.BundleSummary, error) {
	manifest, archiveFiles, warnings, err := s.readAndValidateBundle(opts.BundlePath)
	if err != nil {
		return nil, err
	}

	targetSSHDir, targetNeosshDir := s.resolveImportPaths(opts)

	var restoredFiles []string
	var backupFiles []string

	if opts.DryRun {
		for _, entry := range manifest.Files {
			dest := s.resolveDestinationPath(entry, targetSSHDir, targetNeosshDir)
			restoredFiles = append(restoredFiles, dest)
			if _, statErr := os.Stat(dest); statErr == nil {
				backupFiles = append(backupFiles, dest+".bak.<timestamp>")
			}
		}
		return &domain.BundleSummary{
			Manifest:      *manifest,
			OutputPath:    opts.BundlePath,
			RestoredFiles: restoredFiles,
			BackupFiles:   backupFiles,
			Warnings:      warnings,
		}, nil
	}

	// Ensure destination directories exist with secure permissions
	if err := os.MkdirAll(targetSSHDir, 0o700); err != nil {
		return nil, fmt.Errorf("create SSH directory %q: %w", targetSSHDir, err)
	}
	if err := os.MkdirAll(targetNeosshDir, 0o700); err != nil {
		return nil, fmt.Errorf("create neossh directory %q: %w", targetNeosshDir, err)
	}

	ts := time.Now().Format("20060102-150405")

	for _, entry := range manifest.Files {
		data, exists := archiveFiles[entry.ArchivePath]
		if !exists {
			warnings = append(warnings, fmt.Sprintf("Missing file %q declared in manifest", entry.ArchivePath))
			continue
		}

		dest := s.resolveDestinationPath(entry, targetSSHDir, targetNeosshDir)
		dest = filepath.Clean(dest)

		// Create parent directory
		parentDir := filepath.Dir(dest)
		if err := os.MkdirAll(parentDir, 0o700); err != nil {
			return nil, fmt.Errorf("create directory %q: %w", parentDir, err)
		}

		// Handle backup of existing file
		if _, statErr := os.Stat(dest); statErr == nil {
			if opts.CreateBackup {
				bakPath := fmt.Sprintf("%s.bak.%s", dest, ts)
				if err := copyDiskFile(dest, bakPath); err != nil {
					return nil, fmt.Errorf("backup existing file %q to %q: %w", dest, bakPath, err)
				}
				backupFiles = append(backupFiles, bakPath)
			}
		}

		mode := os.FileMode(entry.FileMode)
		if mode == 0 {
			mode = 0o600
		}

		if err := os.WriteFile(dest, data, mode); err != nil { // #nosec G304: writing restored config
			return nil, fmt.Errorf("write restored file %q: %w", dest, err)
		}
		restoredFiles = append(restoredFiles, dest)
	}

	// Reload server repository if available
	if s.serverRepo != nil {
		_, _ = s.serverRepo.ListServers("")
	}

	s.logger.Infow("imported neossh bundle successfully",
		"bundlePath", opts.BundlePath,
		"restoredFiles", len(restoredFiles),
		"backupFiles", len(backupFiles),
	)

	return &domain.BundleSummary{
		Manifest:      *manifest,
		OutputPath:    opts.BundlePath,
		RestoredFiles: restoredFiles,
		BackupFiles:   backupFiles,
		Warnings:      warnings,
	}, nil
}

func (s *bundleService) resolveExportPaths(opts domain.ExportOptions) (sshConfigPath string, neosshDir string) {
	home, homeErr := os.UserHomeDir()
	if homeErr != nil {
		home = "."
	}

	sshConfigPath = strings.TrimSpace(opts.SSHConfigFile)
	if sshConfigPath == "" {
		if s.serverRepo != nil {
			sshConfigPath = s.serverRepo.GetConfigFile()
		}
	}
	if sshConfigPath == "" {
		sshConfigPath = filepath.Join(home, ".ssh", "config")
	}
	sshConfigPath = filepath.Clean(domain.ExpandTilde(sshConfigPath))

	neosshDir = strings.TrimSpace(opts.NeosshDir)
	if neosshDir == "" {
		neosshDir = filepath.Join(home, ".neossh")
		if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
			neosshDir = filepath.Join(xdg, "neossh")
		}
	}
	neosshDir = filepath.Clean(domain.ExpandTilde(neosshDir))

	return sshConfigPath, neosshDir
}

func (s *bundleService) resolveImportPaths(opts domain.ImportOptions) (targetSSHDir string, targetNeosshDir string) {
	home, homeErr := os.UserHomeDir()
	if homeErr != nil {
		home = "."
	}

	targetSSHDir = strings.TrimSpace(opts.TargetSSHDir)
	if targetSSHDir == "" {
		targetSSHDir = filepath.Join(home, ".ssh")
	}
	targetSSHDir = filepath.Clean(domain.ExpandTilde(targetSSHDir))

	targetNeosshDir = strings.TrimSpace(opts.TargetNeosshDir)
	if targetNeosshDir == "" {
		targetNeosshDir = filepath.Join(home, ".neossh")
		if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
			targetNeosshDir = filepath.Join(xdg, "neossh")
		}
	}
	targetNeosshDir = filepath.Clean(domain.ExpandTilde(targetNeosshDir))

	return targetSSHDir, targetNeosshDir
}

func (s *bundleService) gatherConfigFiles(mainConfig string) []string {
	if s.serverRepo != nil {
		files, err := s.serverRepo.GetConfigFiles()
		if err == nil && len(files) > 0 {
			return files
		}
	}

	// Fallback if repository did not provide list
	if _, err := os.Stat(mainConfig); os.IsNotExist(err) {
		return []string{mainConfig}
	}
	return []string{mainConfig}
}

func (s *bundleService) countServers() int {
	if s.serverRepo == nil {
		return 0
	}
	servers, err := s.serverRepo.ListServers("")
	if err != nil {
		return 0
	}
	return len(servers)
}

func (s *bundleService) readAndValidateBundle(bundlePath string) (*domain.BundleManifest, map[string][]byte, []string, error) {
	cleanPath := filepath.Clean(domain.ExpandTilde(bundlePath))
	file, err := os.Open(cleanPath) // #nosec G304: user input bundle path
	if err != nil {
		return nil, nil, nil, fmt.Errorf("open bundle %q: %w", bundlePath, err)
	}
	defer func() {
		_ = file.Close()
	}()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("invalid gzip archive %q: %w", bundlePath, err)
	}
	defer func() {
		_ = gzipReader.Close()
	}()

	tarReader := tar.NewReader(gzipReader)
	archiveFiles := make(map[string][]byte)
	var manifestBytes []byte
	var warnings []string

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, nil, fmt.Errorf("corrupt tar archive: %w", err)
		}

		if header.Typeflag != tar.TypeReg {
			continue
		}

		// Security: prevent Zip-Slip / path traversal
		cleanName, secErr := sanitizeTarEntryName(header.Name)
		if secErr != nil {
			return nil, nil, nil, secErr
		}

		if header.Size > MaxArchiveFileSize {
			return nil, nil, nil, fmt.Errorf("file %q exceeds maximum permitted size of %d bytes", header.Name, MaxArchiveFileSize)
		}

		data := make([]byte, header.Size)
		if _, err := io.ReadFull(tarReader, data); err != nil {
			return nil, nil, nil, fmt.Errorf("read entry %q: %w", header.Name, err)
		}

		if cleanName == ManifestFileName {
			manifestBytes = data
		} else {
			archiveFiles[cleanName] = data
		}
	}

	if len(manifestBytes) == 0 {
		return nil, nil, nil, fmt.Errorf("bundle %q is missing %s", bundlePath, ManifestFileName)
	}

	var manifest domain.BundleManifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return nil, nil, nil, fmt.Errorf("invalid bundle manifest JSON: %w", err)
	}

	// Verify checksums of entries
	for _, entry := range manifest.Files {
		data, exists := archiveFiles[entry.ArchivePath]
		if !exists {
			warnings = append(warnings, fmt.Sprintf("Entry %q in manifest is missing in archive", entry.ArchivePath))
			continue
		}

		if entry.SHA256Checksum != "" {
			calc := sha256.Sum256(data)
			hexCalc := hex.EncodeToString(calc[:])
			if hexCalc != entry.SHA256Checksum {
				warnings = append(warnings, fmt.Sprintf("Checksum mismatch for %q: expected %s, got %s", entry.ArchivePath, entry.SHA256Checksum, hexCalc))
			}
		}
	}

	return &manifest, archiveFiles, warnings, nil
}

func (s *bundleService) resolveDestinationPath(entry domain.BundleFileEntry, targetSSHDir, targetNeosshDir string) string {
	switch entry.Category {
	case domain.BundleCategorySSHConfig:
		return filepath.Join(targetSSHDir, "config")
	case domain.BundleCategorySSHInclude:
		if entry.RelativePath != "" {
			return filepath.Join(targetSSHDir, entry.RelativePath)
		}
		return filepath.Join(targetSSHDir, "includes", filepath.Base(entry.ArchivePath))
	case domain.BundleCategoryMetadata:
		return filepath.Join(targetNeosshDir, "metadata.json")
	case domain.BundleCategorySettings:
		return filepath.Join(targetNeosshDir, "settings.json")
	default:
		return filepath.Join(targetNeosshDir, filepath.Base(entry.ArchivePath))
	}
}

func sanitizeTarEntryName(name string) (string, error) {
	clean := filepath.Clean(name)
	clean = filepath.ToSlash(clean)
	if strings.HasPrefix(clean, "../") || clean == ".." || strings.HasPrefix(clean, "/") || filepath.IsAbs(clean) || strings.Contains(clean, ":") {
		return "", fmt.Errorf("illegal path traversal in bundle entry: %q", name)
	}
	return clean, nil
}

func writeTarEntry(tw *tar.Writer, name string, data []byte, mode uint32) error {
	header := &tar.Header{
		Name:     filepath.ToSlash(name),
		Mode:     int64(mode),
		Size:     int64(len(data)),
		ModTime:  time.Now(),
		Typeflag: tar.TypeReg,
	}
	if err := tw.WriteHeader(header); err != nil {
		return err
	}
	_, err := tw.Write(data)
	return err
}

func copyDiskFile(src, dst string) error {
	in, err := os.Open(filepath.Clean(src)) // #nosec G304: copying disk file for backup
	if err != nil {
		return err
	}
	defer func() {
		_ = in.Close()
	}()

	out, err := os.OpenFile(filepath.Clean(dst), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600) // #nosec G304: creating backup file
	if err != nil {
		return err
	}
	defer func() {
		_ = out.Close()
	}()

	_, err = io.Copy(out, in)
	return err
}

func sanitizeSSHConfigFile(data []byte) []byte {
	lines := strings.Split(string(data), "\n")
	var result []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// 1. Sanitize IdentityFile
		if matches := reIdentityDirective.FindStringSubmatch(line); len(matches) > 1 {
			indent := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
			result = append(result, indent+"# IdentityFile [sanitized]")
			continue
		}

		// 2. Sanitize sensitive comments
		if reSensitiveComment.MatchString(trimmed) {
			indent := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
			result = append(result, indent+"# [sanitized secret comment]")
			continue
		}

		// 3. Sanitize inline secrets in commands
		if strings.HasPrefix(strings.ToLower(trimmed), "preconnectcommand") ||
			strings.HasPrefix(strings.ToLower(trimmed), "certificatecommand") ||
			strings.HasPrefix(strings.ToLower(trimmed), "localcommand") {
			line = reSecretTokenInCmd.ReplaceAllString(line, "$1=[sanitized]")
		}

		result = append(result, line)
	}

	return []byte(strings.Join(result, "\n"))
}

func sanitizeMetadataJSON(data []byte) []byte {
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return data
	}

	// Sanitize settings inside metadata.json
	if settingsRaw, ok := parsed["settings"].(map[string]interface{}); ok {
		if _, hasDefKey := settingsRaw["default_identity_key"]; hasDefKey {
			settingsRaw["default_identity_key"] = ""
		}
	}

	// Sanitize servers file paths and secrets
	if serversRaw, ok := parsed["servers"].(map[string]interface{}); ok {
		for _, srvVal := range serversRaw {
			if srvMap, ok := srvVal.(map[string]interface{}); ok {
				if filePath, ok := srvMap["file"].(string); ok && filePath != "" {
					srvMap["file"] = sanitizePathString(filePath, true)
				}
				if precmd, ok := srvMap["pre_connect_command"].(string); ok && precmd != "" {
					srvMap["pre_connect_command"] = reSecretTokenInCmd.ReplaceAllString(precmd, "$1=[sanitized]")
				}
				if certcmd, ok := srvMap["certificate_command"].(string); ok && certcmd != "" {
					srvMap["certificate_command"] = reSecretTokenInCmd.ReplaceAllString(certcmd, "$1=[sanitized]")
				}
			}
		}
	}

	sanitized, err := json.MarshalIndent(parsed, "", "  ")
	if err != nil {
		return data
	}
	return sanitized
}

func sanitizeSettingsJSON(data []byte) []byte {
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return data
	}

	if _, hasDefKey := parsed["default_identity_key"]; hasDefKey {
		parsed["default_identity_key"] = ""
	}

	sanitized, err := json.MarshalIndent(parsed, "", "  ")
	if err != nil {
		return data
	}
	return sanitized
}

func sanitizePathString(p string, sanitize bool) string {
	if !sanitize {
		return p
	}
	// Replace /Users/username/ or /home/username/ with ~/
	return reHomePrefixUsername.ReplaceAllString(p, "~$2")
}
