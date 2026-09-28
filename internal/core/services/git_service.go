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
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/WhiteRoseLK/neossh/internal/core/domain"
	"github.com/WhiteRoseLK/neossh/internal/core/ports"
	"go.uber.org/zap"
)

const (
	// ScopeLocal represents repository-level git configuration.
	ScopeLocal = "local"
	// ScopeGlobal represents user-level git configuration.
	ScopeGlobal = "global"
	// ScopeBoth represents both repository and user-level git configuration.
	ScopeBoth = "both"

	// SSH key type constants
	keyTypeRSA     = "rsa"
	keyTypeEd25519 = "ed25519"
	keyTypeECDSA   = "ecdsa"
	keyTypeDSA     = "dsa"

	// SSH file constants
	fileKnownHosts     = "known_hosts"
	fileConfig         = "config"
	fileAuthorizedKeys = "authorized_keys"
)

type gitService struct {
	logger           *zap.SugaredLogger
	serverRepository ports.ServerRepository
}

// NewGitService creates a new instance of GitService.
func NewGitService(logger *zap.SugaredLogger) ports.GitService {
	return &gitService{
		logger: logger,
	}
}

// SetServerRepository sets the server repository for the git service.
func (gs *gitService) SetServerRepository(repo ports.ServerRepository) {
	gs.serverRepository = repo
}

// IsGitRepository checks if the given path is inside a git repository.
func (gs *gitService) IsGitRepository(path string) bool {
	// #nosec G204 -- controlled arguments
	cmd := exec.Command("git", "-C", path, "rev-parse", "--git-dir")
	return cmd.Run() == nil
}

// GetGitRootPath returns the root path of the git repository.
func (gs *gitService) GetGitRootPath(path string) (string, error) {
	// #nosec G204 -- controlled arguments
	cmd := exec.Command("git", "-C", path, "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("not a git repository or git command failed: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// GetPushRemoteURL returns the push remote name and URL for the git repository.
func (gs *gitService) GetPushRemoteURL(repoPath string) (remoteName, remoteURL string, err error) {
	// #nosec G204 -- controlled arguments
	cmd := exec.Command("git", "-C", repoPath, "remote", "-v")
	output, err := cmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to get git remotes: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var otherRemoteName, otherPushURL string

	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}
		name := parts[0]
		url := parts[1]
		operation := strings.TrimPrefix(strings.TrimSuffix(parts[2], ")"), "(")

		if operation == "push" {
			if name == "origin" {
				return name, url, nil
			} else if otherRemoteName == "" {
				otherRemoteName = name
				otherPushURL = url
			}
		}
	}

	if otherRemoteName != "" {
		return otherRemoteName, otherPushURL, nil
	}

	return "", "", fmt.Errorf("no push remote found")
}

// ListSSHKeys lists all SSH private keys from the SSH config and the specified SSH directory.
func (gs *gitService) ListSSHKeys(sshDir string, serverRepo ports.ServerRepository) ([]domain.SSHKey, error) {
	keyMap := make(map[string]*domain.SSHKey)

	// Add keys from SSH config
	gs.addKeysFromSSHConfig(serverRepo, keyMap)

	// Also scan sshDir or ~/.ssh/
	if sshDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %w", err)
		}
		sshDir = filepath.Join(homeDir, ".ssh")
	}

	if err := gs.addKeysFromDirectory(sshDir, keyMap); err != nil && len(keyMap) == 0 {
		return nil, err
	}

	gs.markKeysInAgent(keyMap)

	keys := make([]domain.SSHKey, 0, len(keyMap))
	for _, key := range keyMap {
		keys = append(keys, *key)
	}

	sort.Slice(keys, func(i, j int) bool {
		if keys[i].LoadedInAgent != keys[j].LoadedInAgent {
			return keys[i].LoadedInAgent
		}
		return keys[i].ModTime.After(keys[j].ModTime)
	})

	return keys, nil
}

func (gs *gitService) addKeysFromSSHConfig(serverRepo ports.ServerRepository, keyMap map[string]*domain.SSHKey) {
	if serverRepo == nil {
		return
	}

	servers, err := serverRepo.ListServers("")
	if err != nil {
		return
	}

	for _, server := range servers {
		for _, identityFile := range server.IdentityFiles {
			gs.addKeyFromPath(identityFile, keyMap)
		}
	}
}

func (gs *gitService) addKeyFromPath(identityFile string, keyMap map[string]*domain.SSHKey) {
	if identityFile == "" {
		return
	}

	if strings.HasPrefix(identityFile, "~/") {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			identityFile = filepath.Join(homeDir, identityFile[2:])
		}
	}

	if _, exists := keyMap[identityFile]; exists {
		return
	}

	info, err := os.Stat(identityFile)
	if err != nil {
		return
	}

	isPrivateKey, isEncrypted := gs.isSSHPrivateKey(identityFile)
	if !isPrivateKey {
		return
	}

	key := &domain.SSHKey{
		Name:         filepath.Base(identityFile),
		Path:         identityFile,
		HasPublicKey: false,
		ModTime:      info.ModTime(),
		IsEncrypted:  isEncrypted,
		Source:       "config",
	}

	pubKeyPath := identityFile + ".pub"
	if _, err := os.Stat(pubKeyPath); err == nil {
		key.HasPublicKey = true
	}

	keyMap[identityFile] = key
}

func (gs *gitService) addKeysFromDirectory(sshDir string, keyMap map[string]*domain.SSHKey) error {
	entries, err := os.ReadDir(sshDir)
	if err != nil {
		if len(keyMap) > 0 {
			return nil
		}
		return fmt.Errorf("failed to read SSH directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || gs.shouldSkipFile(entry.Name()) {
			continue
		}

		keyPath := filepath.Join(sshDir, entry.Name())
		if _, exists := keyMap[keyPath]; exists {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		isPrivateKey, isEncrypted := gs.isSSHPrivateKey(keyPath)
		if !isPrivateKey {
			continue
		}

		key := &domain.SSHKey{
			Name:         entry.Name(),
			Path:         keyPath,
			HasPublicKey: false,
			ModTime:      info.ModTime(),
			IsEncrypted:  isEncrypted,
			Source:       "filesystem",
		}

		pubKeyPath := keyPath + ".pub"
		if _, err := os.Stat(pubKeyPath); err == nil {
			key.HasPublicKey = true
		}

		keyMap[keyPath] = key
	}

	return nil
}

func (gs *gitService) shouldSkipFile(name string) bool {
	return name == fileKnownHosts || name == fileConfig || name == fileAuthorizedKeys ||
		strings.HasSuffix(name, ".pub") || strings.HasPrefix(name, ".")
}

func (gs *gitService) markKeysInAgent(keyMap map[string]*domain.SSHKey) {
	agentKeys, err := gs.GetLoadedAgentKeys()
	if err != nil || len(agentKeys) == 0 {
		return
	}

	for keyPath, key := range keyMap {
		for _, agentLine := range agentKeys {
			if strings.Contains(agentLine, keyPath) || strings.Contains(agentLine, key.Name) {
				key.LoadedInAgent = true
				break
			}
		}
	}
}

// isSSHPrivateKey checks if a file is an SSH private key and if it's encrypted.
func (gs *gitService) isSSHPrivateKey(path string) (bool, bool) {
	// #nosec G304 -- reading SSH key paths from ~/.ssh or ssh config
	file, err := os.Open(path)
	if err != nil {
		return false, false
	}
	defer func() {
		_ = file.Close()
	}()

	scanner := bufio.NewScanner(file)
	isPrivateKey := false
	isEncrypted := false

	lineCount := 0
	for scanner.Scan() && lineCount < 15 {
		line := scanner.Text()
		lineCount++

		if strings.Contains(line, "BEGIN") && strings.Contains(line, "PRIVATE KEY") {
			isPrivateKey = true
		}
		if strings.Contains(line, "ENCRYPTED") || strings.Contains(line, "Proc-Type: 4,ENCRYPTED") || strings.Contains(line, "YmNyeXB0") {
			isEncrypted = true
		}
	}

	return isPrivateKey, isEncrypted
}

// ConfigureGitSSHKey configures Git to use the specified SSH key via core.sshCommand.
func (gs *gitService) ConfigureGitSSHKey(repoPath string, keyPath string, scope string) error {
	if scope != ScopeLocal && scope != ScopeGlobal {
		return fmt.Errorf("invalid scope: %s (must be 'local' or 'global')", scope)
	}

	if _, err := os.Stat(keyPath); err != nil {
		return fmt.Errorf("SSH key not found at %s: %w", keyPath, err)
	}

	sshCommand := fmt.Sprintf("ssh -i %s -o IdentitiesOnly=yes", keyPath)

	var cmd *exec.Cmd
	if scope == ScopeLocal {
		// #nosec G204 -- arguments are controlled
		cmd = exec.Command("git", "-C", repoPath, "config", "--local", "core.sshCommand", sshCommand)
	} else {
		// #nosec G204 -- arguments are controlled
		cmd = exec.Command("git", "config", "--global", "core.sshCommand", sshCommand)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to configure Git SSH key: %w (output: %s)", err, string(output))
	}

	gs.logger.Infof("Successfully configured Git to use SSH key: %s (scope: %s)", keyPath, scope)
	return nil
}

// GetCurrentGitSSHConfig retrieves the current Git SSH configuration.
func (gs *gitService) GetCurrentGitSSHConfig(repoPath string) (string, error) {
	if repoPath != "" {
		// #nosec G204 -- controlled arguments
		cmd := exec.Command("git", "-C", repoPath, "config", "--local", "core.sshCommand")
		if output, err := cmd.Output(); err == nil && len(output) > 0 {
			return strings.TrimSpace(string(output)), nil
		}
	}

	cmd := exec.Command("git", "config", "--global", "core.sshCommand")
	if output, err := cmd.Output(); err == nil && len(output) > 0 {
		return strings.TrimSpace(string(output)), nil
	}

	if envSSH := os.Getenv("GIT_SSH_COMMAND"); envSSH != "" {
		return envSSH + " (from environment)", nil
	}

	return "", nil
}

// ClearGitSSHConfig removes the Git SSH configuration.
func (gs *gitService) ClearGitSSHConfig(repoPath string, scope string) error {
	if scope != ScopeLocal && scope != ScopeGlobal && scope != ScopeBoth {
		return fmt.Errorf("invalid scope: %s (must be 'local', 'global', or 'both')", scope)
	}

	var errs []string

	if (scope == ScopeLocal || scope == ScopeBoth) && repoPath != "" {
		// #nosec G204 -- controlled arguments
		cmd := exec.Command("git", "-C", repoPath, "config", "--local", "--unset", "core.sshCommand")
		if output, err := cmd.CombinedOutput(); err != nil {
			var exitErr *exec.ExitError
			if !errors.As(err, &exitErr) || exitErr.ExitCode() != 5 {
				errs = append(errs, fmt.Sprintf("local: %v (%s)", err, strings.TrimSpace(string(output))))
			}
		}
	}

	if scope == ScopeGlobal || scope == ScopeBoth {
		cmd := exec.Command("git", "config", "--global", "--unset", "core.sshCommand")
		if output, err := cmd.CombinedOutput(); err != nil {
			var exitErr *exec.ExitError
			if !errors.As(err, &exitErr) || exitErr.ExitCode() != 5 {
				errs = append(errs, fmt.Sprintf("global: %v (%s)", err, strings.TrimSpace(string(output))))
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to clear Git SSH config: %s", strings.Join(errs, "; "))
	}

	return nil
}

// GetLoadedAgentKeys returns a list of SSH key fingerprint lines currently in ssh-agent.
func (gs *gitService) GetLoadedAgentKeys() ([]string, error) {
	cmd := exec.Command("ssh-add", "-l")
	output, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to list ssh-agent keys: %w", err)
	}

	var fingerprints []string
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			fingerprints = append(fingerprints, strings.TrimSpace(line))
		}
	}

	return fingerprints, nil
}

type agentKeyInfo struct {
	fingerprint string
	comment     string
}

// ListAllSSHKeys returns all SSH keys from config identity files and ssh-agent.
func (gs *gitService) ListAllSSHKeys(serverRepo ports.ServerRepository) ([]domain.SSHKey, error) {
	return gs.ListSSHKeysFromConfig(serverRepo)
}

// ListSSHKeysFromConfig returns SSH keys from config files and ~/.ssh directory.
func (gs *gitService) ListSSHKeysFromConfig(serverRepo ports.ServerRepository) ([]domain.SSHKey, error) {
	keysMap := make(map[string]*domain.SSHKey)
	agentKeys := gs.getAgentKeyMap()

	if serverRepo != nil {
		servers, _ := serverRepo.ListServers("")
		for _, server := range servers {
			for _, identityFile := range server.IdentityFiles {
				expandedPath := identityFile
				if strings.HasPrefix(identityFile, "~/") {
					if home, err := os.UserHomeDir(); err == nil {
						expandedPath = filepath.Join(home, identityFile[2:])
					}
				}

				if _, exists := keysMap[expandedPath]; !exists {
					if key := gs.parseKeyFile(expandedPath, agentKeys); key != nil {
						key.Source = "config"
						keysMap[expandedPath] = key
					}
				}
			}
		}
	}

	if home, err := os.UserHomeDir(); err == nil {
		sshDir := filepath.Join(home, ".ssh")
		if entries, err := os.ReadDir(sshDir); err == nil {
			for _, entry := range entries {
				if entry.IsDir() || gs.shouldSkipFile(entry.Name()) {
					continue
				}

				fullPath := filepath.Join(sshDir, entry.Name())
				if _, exists := keysMap[fullPath]; !exists {
					if key := gs.parseKeyFile(fullPath, agentKeys); key != nil {
						key.Source = "filesystem"
						keysMap[fullPath] = key
					}
				}
			}
		}
	}

	keys := make([]domain.SSHKey, 0, len(keysMap))
	for _, key := range keysMap {
		keys = append(keys, *key)
	}

	sort.Slice(keys, func(i, j int) bool {
		if keys[i].LoadedInAgent != keys[j].LoadedInAgent {
			return keys[i].LoadedInAgent
		}
		return keys[i].Name < keys[j].Name
	})

	return keys, nil
}

func (gs *gitService) getAgentKeyMap() map[string]agentKeyInfo {
	agentKeys := make(map[string]agentKeyInfo)
	cmd := exec.Command("ssh-add", "-l")
	if output, err := cmd.Output(); err == nil {
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		for _, line := range lines {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				info := agentKeyInfo{
					fingerprint: parts[1],
				}
				if len(parts) >= 3 {
					endIdx := len(parts)
					last := parts[len(parts)-1]
					if strings.HasPrefix(last, "(") && strings.HasSuffix(last, ")") {
						endIdx = len(parts) - 1
					}
					if endIdx > 2 {
						info.comment = strings.Join(parts[2:endIdx], " ")
					}
				}
				agentKeys[parts[1]] = info
			}
		}
	}
	return agentKeys
}

func (gs *gitService) parseKeyFile(path string, agentKeys map[string]agentKeyInfo) *domain.SSHKey {
	info, err := os.Stat(path)
	if err != nil {
		return nil
	}

	// #nosec G304 -- validated SSH path
	content, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "PRIVATE KEY") {
		return nil
	}

	key := &domain.SSHKey{
		Path:       path,
		Name:       filepath.Base(path),
		ModTime:    info.ModTime(),
		Source:     "config",
		FileExists: true,
	}

	gs.detectKeyTypeAndEncryption(key, contentStr)
	gs.populatePublicKeyInfo(key, agentKeys)

	return key
}

func (gs *gitService) detectKeyTypeAndEncryption(key *domain.SSHKey, contentStr string) {
	switch {
	case strings.Contains(contentStr, "RSA PRIVATE KEY"):
		key.Type = keyTypeRSA
		key.IsEncrypted = strings.Contains(contentStr, "ENCRYPTED")
	case strings.Contains(contentStr, "OPENSSH PRIVATE KEY"):
		lowerName := strings.ToLower(key.Name)
		switch {
		case strings.Contains(key.Path, "ed25519") || strings.Contains(lowerName, "ed25519"):
			key.Type = keyTypeEd25519
		case strings.Contains(key.Path, "ecdsa") || strings.Contains(lowerName, "ecdsa"):
			key.Type = keyTypeECDSA
		default:
			key.Type = keyTypeRSA
		}
		key.IsEncrypted = strings.Contains(contentStr, "Proc-Type: 4,ENCRYPTED") || strings.Contains(contentStr, "YmNyeXB0")
	case strings.Contains(contentStr, "DSA PRIVATE KEY"):
		key.Type = keyTypeDSA
		key.IsEncrypted = strings.Contains(contentStr, "ENCRYPTED")
	case strings.Contains(contentStr, "EC PRIVATE KEY"):
		key.Type = keyTypeECDSA
		key.IsEncrypted = strings.Contains(contentStr, "ENCRYPTED")
	default:
		key.Type = keyTypeRSA
	}
}

func (gs *gitService) populatePublicKeyInfo(key *domain.SSHKey, agentKeys map[string]agentKeyInfo) {
	pubPath := key.Path + ".pub"
	if _, err := os.Stat(pubPath); err != nil {
		return
	}

	key.HasPublicKey = true

	// #nosec G304 -- validated public key path
	if pubContent, err := os.ReadFile(pubPath); err == nil {
		pubKeyLine := strings.TrimSpace(string(pubContent))
		parts := strings.Fields(pubKeyLine)
		if len(parts) >= 2 {
			// Detect key type and FIDO2 from authoritative public key header
			gs.detectKeyTypeFromPubHeader(key, parts[0])
		}
		if len(parts) >= 3 {
			key.Comment = strings.Join(parts[2:], " ")
		}
		key.PublicKeyLine = pubKeyLine
	}

	// #nosec G204 -- validated public key path
	cmd := exec.Command("ssh-keygen", "-lf", pubPath)
	if output, err := cmd.Output(); err == nil {
		parts := strings.Fields(string(output))
		if len(parts) >= 2 {
			key.Size, _ = strconv.Atoi(parts[0])
			key.Fingerprint = parts[1]

			for fp, info := range agentKeys {
				if fp == key.Fingerprint || (key.Comment != "" && info.comment == key.Comment) {
					key.LoadedInAgent = true
					if key.Comment == "" {
						key.Comment = info.comment
					}
					break
				}
			}
		}
	}
}

// detectKeyTypeFromPubHeader sets the key type and FIDO2 flag from the public key algorithm header.
// This is authoritative — it overrides any filename-based heuristic.
func (gs *gitService) detectKeyTypeFromPubHeader(key *domain.SSHKey, header string) {
	switch header {
	case "sk-ssh-ed25519@openssh.com":
		key.Type = keyTypeEd25519
		key.IsFIDO2 = true
	case "sk-ecdsa-sha2-nistp256@openssh.com":
		key.Type = keyTypeECDSA
		key.IsFIDO2 = true
	case "ssh-ed25519":
		key.Type = keyTypeEd25519
	case "ssh-rsa":
		key.Type = keyTypeRSA
	case "ecdsa-sha2-nistp256", "ecdsa-sha2-nistp384", "ecdsa-sha2-nistp521":
		key.Type = keyTypeECDSA
	case "ssh-dss":
		key.Type = keyTypeDSA
	}
}

// ListSSHKeysFromAgent returns SSH keys currently loaded in ssh-agent.
func (gs *gitService) ListSSHKeysFromAgent() ([]domain.SSHKey, error) {
	keys := make([]domain.SSHKey, 0)

	cmd := exec.Command("ssh-add", "-L")
	output, err := cmd.Output()
	if err != nil {
		return keys, nil
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		key := gs.parseAgentKeyLine(line)
		if key != nil {
			keys = append(keys, *key)
		}
	}

	return gs.mergeAgentKeysWithFilesystem(keys), nil
}

func (gs *gitService) parseAgentKeyLine(line string) *domain.SSHKey {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil
	}
	parts := strings.Fields(line)
	if len(parts) < 2 {
		return nil
	}

	key := &domain.SSHKey{
		LoadedInAgent: true,
		PublicKeyLine: line,
		Source:        "agent",
	}

	// Detect FIDO2/security-key types first, then standard types
	switch parts[0] {
	case "sk-ssh-ed25519@openssh.com":
		key.Type = keyTypeEd25519
		key.IsFIDO2 = true
	case "sk-ecdsa-sha2-nistp256@openssh.com":
		key.Type = keyTypeECDSA
		key.IsFIDO2 = true
	default:
		keyType := strings.TrimPrefix(parts[0], "ssh-")
		switch keyType {
		case "rsa":
			key.Type = keyTypeRSA
		case "ed25519":
			key.Type = keyTypeEd25519
		case "ecdsa-sha2-nistp256", "ecdsa-sha2-nistp384", "ecdsa-sha2-nistp521":
			key.Type = keyTypeECDSA
		case "dss":
			key.Type = keyTypeDSA
		default:
			key.Type = keyType
		}
	}

	if len(parts) >= 3 {
		key.Comment = strings.Join(parts[2:], " ")
		key.Name = key.Comment
	} else {
		key.Name = fmt.Sprintf("agent-key-%s", key.Type)
	}

	cmd := exec.Command("ssh-keygen", "-lf", "-")
	cmd.Stdin = strings.NewReader(line + "\n")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err == nil {
		fpFields := strings.Fields(stdout.String())
		if len(fpFields) >= 2 {
			key.Fingerprint = fpFields[1]
			if size, err := strconv.Atoi(fpFields[0]); err == nil {
				key.Size = size
			}
		}
	}

	return key
}

func (gs *gitService) mergeAgentKeysWithFilesystem(keys []domain.SSHKey) []domain.SSHKey {
	fsKeys, err := gs.ListSSHKeysFromConfig(gs.serverRepository)
	if err != nil {
		return keys
	}

	fsKeyByPubKey := make(map[string]domain.SSHKey)
	fsKeyByFingerprint := make(map[string]domain.SSHKey)
	fsKeyByName := make(map[string]domain.SSHKey)

	for _, fsKey := range fsKeys {
		if fsKey.PublicKeyLine != "" {
			fsKeyByPubKey[fsKey.PublicKeyLine] = fsKey
		}
		if fsKey.Fingerprint != "" {
			fsKeyByFingerprint[fsKey.Fingerprint] = fsKey
		}
		if fsKey.Name != "" {
			fsKeyByName[fsKey.Name] = fsKey
		}
	}

	for i := range keys {
		var matched *domain.SSHKey
		if k, ok := fsKeyByPubKey[keys[i].PublicKeyLine]; ok {
			matched = &k
		} else if keys[i].Fingerprint != "" {
			if k, ok := fsKeyByFingerprint[keys[i].Fingerprint]; ok {
				matched = &k
			}
		} else if keys[i].Name != "" {
			if k, ok := fsKeyByName[keys[i].Name]; ok {
				matched = &k
			}
		}

		if matched != nil {
			keys[i].Path = matched.Path
			if matched.Comment != "" {
				keys[i].Comment = matched.Comment
			}
			if matched.Name != "" && len(matched.Name) > len(keys[i].Name) {
				keys[i].Name = matched.Name
			}
			if keys[i].Fingerprint == "" && matched.Fingerprint != "" {
				keys[i].Fingerprint = matched.Fingerprint
			}
			if keys[i].Type == "" && matched.Type != "" {
				keys[i].Type = matched.Type
			}
			if keys[i].Size == 0 && matched.Size > 0 {
				keys[i].Size = matched.Size
			}
			if matched.HasPublicKey {
				keys[i].HasPublicKey = true
			}
			if matched.IsEncrypted {
				keys[i].IsEncrypted = true
			}
		}
	}

	return keys
}

// LoadKeyToAgent loads an SSH key into ssh-agent.
func (gs *gitService) LoadKeyToAgent(keyPath string) error {
	// #nosec G204 -- keyPath is an existing key path
	cmd := exec.Command("ssh-add", keyPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// UnloadKeyFromAgent removes an SSH key from ssh-agent using its public key line.
func (gs *gitService) UnloadKeyFromAgent(publicKeyLine string) error {
	if publicKeyLine == "" {
		return fmt.Errorf("empty public key line provided")
	}

	var stderrBuf bytes.Buffer
	cmd := exec.Command("ssh-add", "-d", "-")
	cmd.Stdin = strings.NewReader(publicKeyLine + "\n")
	cmd.Stdout = os.Stdout
	cmd.Stderr = &stderrBuf

	if err := cmd.Run(); err != nil {
		stderrMsg := stderrBuf.String()
		if stderrMsg != "" {
			return fmt.Errorf("failed to unload key: %w (%s)", err, strings.TrimSpace(stderrMsg))
		}
		return fmt.Errorf("failed to unload key: %w", err)
	}

	gs.logger.Infof("Successfully unloaded key")
	return nil
}

// UpdateKeyComment updates the comment for an SSH key using ssh-keygen.
func (gs *gitService) UpdateKeyComment(keyPath, comment string) error {
	if _, err := os.Stat(keyPath); err != nil {
		return fmt.Errorf("key file not found: %w", err)
	}

	pubKeyPath := keyPath + ".pub"
	if _, err := os.Stat(pubKeyPath); err != nil {
		return fmt.Errorf("public key file (.pub) not found: %w", err)
	}

	originalMode, err := getFileMode(keyPath)
	if err != nil {
		return fmt.Errorf("failed to get private key file permissions: %w", err)
	}
	pubOriginalMode, err := getFileMode(pubKeyPath)
	if err != nil {
		return fmt.Errorf("failed to get public key file permissions: %w", err)
	}

	privateReadonly := originalMode&0o200 == 0
	pubReadonly := pubOriginalMode&0o200 == 0

	if privateReadonly {
		if err := os.Chmod(keyPath, 0o600); err != nil {
			return fmt.Errorf("failed to make private key file writable: %w", err)
		}
	}

	if pubReadonly {
		if err := os.Chmod(pubKeyPath, 0o600); err != nil {
			if privateReadonly {
				_ = os.Chmod(keyPath, originalMode)
			}
			return fmt.Errorf("failed to make public key file writable: %w", err)
		}
	}

	// #nosec G204 -- keyPath and comment are validated
	cmd := exec.Command("ssh-keygen", "-c", "-C", comment, "-f", keyPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if privateReadonly {
			_ = os.Chmod(keyPath, originalMode)
		}
		if pubReadonly {
			_ = os.Chmod(pubKeyPath, pubOriginalMode)
		}
		return fmt.Errorf("failed to update key comment: %w", err)
	}

	if privateReadonly {
		_ = os.Chmod(keyPath, originalMode)
	}
	if pubReadonly {
		_ = os.Chmod(pubKeyPath, pubOriginalMode)
	}

	gs.logger.Infof("Successfully updated key comment for %s", keyPath)
	return nil
}

func getFileMode(path string) (os.FileMode, error) {
	info, err := os.Stat(path)
	if err != nil {
		info, err = os.Lstat(path)
		if err != nil {
			return 0, err
		}
	}
	return info.Mode(), nil
}

// GetSSHAgentStatus returns live telemetry on the active SSH agent.
func (gs *gitService) GetSSHAgentStatus() domain.SSHAgentStatus {
	return QuerySSHAgentStatus()
}
