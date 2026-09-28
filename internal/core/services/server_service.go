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
	"io"
	"net"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"unicode"

	"github.com/WhiteRoseLK/neossh/internal/core/domain"
	"github.com/WhiteRoseLK/neossh/internal/core/ports"
	"go.uber.org/zap"
)

// ErrReadOnly is returned when an operation modifying SSH configuration is attempted in read-only mode.
var ErrReadOnly = errors.New("readonly mode: SSH configuration modifications are disabled")

type serverService struct {
	serverRepository ports.ServerRepository
	credentialStore  ports.CredentialStore
	logger           *zap.SugaredLogger
	readonly         bool

	serversMu     sync.RWMutex
	serversLoaded bool
	servers       []domain.Server

	fwMu     sync.Mutex
	forwards map[string][]*os.Process

	newSSHCommand         func(alias string) *exec.Cmd
	newSSHCommandWithArgs func(alias string, extraArgs []string) *exec.Cmd
	newSFTPCommand        func(alias string, args []string) *exec.Cmd
	newFileManagerCommand func(tool string, args []string) *exec.Cmd
	newHookCommand        func(cmdStr string) *exec.Cmd
	newSSHPassCommand     func(sshpassPath, pwd string, sshCmd *exec.Cmd) *exec.Cmd
	lookPath              func(file string) (string, error)
}

// ServerServiceOption allows configuring a serverService instance.
type ServerServiceOption func(*serverService)

// WithReadOnly sets whether the service operates in read-only mode.
func WithReadOnly(ro bool) ServerServiceOption {
	return func(s *serverService) {
		s.readonly = ro
	}
}

// WithCredentialStore sets the credential store for secure password management.
func WithCredentialStore(cs ports.CredentialStore) ServerServiceOption {
	return func(s *serverService) {
		s.credentialStore = cs
	}
}

// WithLookPath sets the LookPath function for finding executables (useful in tests).
func WithLookPath(lp func(string) (string, error)) ServerServiceOption {
	return func(s *serverService) {
		s.lookPath = lp
	}
}

// NewServerService creates a new instance of serverService.
func NewServerService(logger *zap.SugaredLogger, sr ports.ServerRepository, opts ...ServerServiceOption) ports.ServerService {
	s := &serverService{
		logger:           logger,
		serverRepository: sr,
		lookPath:         exec.LookPath,
		newSSHPassCommand: func(sshpassPath, pwd string, sshCmd *exec.Cmd) *exec.Cmd {
			args := append([]string{"-e", sshCmd.Path}, sshCmd.Args[1:]...)
			//nolint:gosec // G204: intentional execution of sshpass with ssh arguments
			cmd := exec.Command(sshpassPath, args...)
			env := sshCmd.Env
			if len(env) == 0 {
				env = os.Environ()
			}
			env = append(env, "SSHPASS="+pwd)
			cmd.Env = env
			return cmd
		},
		newSSHCommand: func(alias string) *exec.Cmd {
			//nolint:gosec // G204: intentional SSH command
			return exec.Command("ssh", "-F", sr.GetConfigFile(), alias)
		},
		newSSHCommandWithArgs: func(alias string, extraArgs []string) *exec.Cmd {
			args := append([]string{}, extraArgs...)
			args = append(args, "-F", sr.GetConfigFile(), alias)
			//nolint:gosec // G204: intentional SSH command
			return exec.Command("ssh", args...)
		},
		newSFTPCommand: func(alias string, extraArgs []string) *exec.Cmd {
			args := make([]string, 0, 2+len(extraArgs)+1)
			args = append(args, "-F", sr.GetConfigFile())
			args = append(args, extraArgs...)
			args = append(args, alias)
			//nolint:gosec // G204: intentional SFTP command
			return exec.Command("sftp", args...)
		},
		newHookCommand: func(cmdStr string) *exec.Cmd {
			if runtime.GOOS == "windows" {
				//nolint:gosec // G204: intentional user pre-connect command hook
				return exec.Command("cmd.exe", "/c", cmdStr)
			}
			//nolint:gosec // G204: intentional user pre-connect command hook
			return exec.Command("sh", "-c", cmdStr)
		},
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// ReloadServers reloads all servers from the underlying repository into memory.
func (s *serverService) ReloadServers() error {
	s.serversMu.Lock()
	defer s.serversMu.Unlock()
	return s.loadFromRepoLocked()
}

// UpdateServerPing updates the ping status and latency of a server in-memory.
func (s *serverService) UpdateServerPing(alias string, status string, latency time.Duration) {
	s.serversMu.Lock()
	defer s.serversMu.Unlock()

	for i := range s.servers {
		if s.servers[i].Alias == alias || hasAlias(s.servers[i].Aliases, alias) {
			s.servers[i].PingStatus = status
			s.servers[i].PingLatency = latency
			return
		}
	}
}

func hasAlias(aliases []string, target string) bool {
	for _, a := range aliases {
		if strings.EqualFold(a, target) {
			return true
		}
	}
	return false
}

func (s *serverService) loadFromRepoLocked() error {
	if s.serverRepository == nil {
		s.servers = nil
		s.serversLoaded = true
		return nil
	}
	rawServers, err := s.serverRepository.ListServers("")
	if err != nil {
		s.logger.Errorw("failed to load servers from repository", "error", err)
		return err
	}
	s.enrichServersWithCredentials(rawServers)

	// Preserve in-memory ping statuses if previously loaded
	if len(s.servers) > 0 {
		pingMap := make(map[string]struct {
			status  string
			latency time.Duration
		}, len(s.servers))
		for _, srv := range s.servers {
			if srv.PingStatus != "" {
				pingMap[srv.Alias] = struct {
					status  string
					latency time.Duration
				}{status: srv.PingStatus, latency: srv.PingLatency}
			}
		}
		for i := range rawServers {
			if p, ok := pingMap[rawServers[i].Alias]; ok {
				rawServers[i].PingStatus = p.status
				rawServers[i].PingLatency = p.latency
			}
		}
	}

	s.servers = rawServers
	s.serversLoaded = true
	return nil
}

func (s *serverService) ensureLoaded() error {
	s.serversMu.RLock()
	if s.serversLoaded {
		s.serversMu.RUnlock()
		return nil
	}
	s.serversMu.RUnlock()

	s.serversMu.Lock()
	defer s.serversMu.Unlock()
	if s.serversLoaded {
		return nil
	}
	return s.loadFromRepoLocked()
}

// ListServers returns servers. With empty query, keep pinned-first default ordering.
// With non-empty query, perform fuzzy subsequence matching and rank by relevance.
func (s *serverService) ListServers(query string) ([]domain.Server, error) {
	if err := s.ensureLoaded(); err != nil {
		return nil, err
	}

	s.serversMu.RLock()
	defer s.serversMu.RUnlock()

	q := strings.TrimSpace(query)
	if q == "" {
		if s.servers == nil {
			return nil, nil
		}
		out := make([]domain.Server, len(s.servers))
		copy(out, s.servers)
		sort.SliceStable(out, func(i, j int) bool {
			pi := !out[i].PinnedAt.IsZero()
			pj := !out[j].PinnedAt.IsZero()
			if pi != pj {
				return pi
			}
			if pi && pj {
				return out[i].PinnedAt.After(out[j].PinnedAt)
			}
			ai := strings.ToLower(out[i].Alias)
			aj := strings.ToLower(out[j].Alias)
			if ai != aj {
				return ai < aj
			}
			return out[i].Alias < out[j].Alias
		})
		return out, nil
	}

	filter := parseSearchQuery(q)
	type scored struct {
		srv   domain.Server
		score int
	}
	results := make([]scored, 0, len(s.servers))
	for _, srv := range s.servers {
		score := filter.Score(srv)
		if score > 0 {
			results = append(results, scored{srv: srv, score: score})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].score != results[j].score {
			return results[i].score > results[j].score
		}
		pi := !results[i].srv.PinnedAt.IsZero()
		pj := !results[j].srv.PinnedAt.IsZero()
		if pi != pj {
			return pi
		}
		if pi && pj && !results[i].srv.PinnedAt.Equal(results[j].srv.PinnedAt) {
			return results[i].srv.PinnedAt.After(results[j].srv.PinnedAt)
		}
		return strings.ToLower(results[i].srv.Alias) < strings.ToLower(results[j].srv.Alias)
	})

	out := make([]domain.Server, len(results))
	for i, r := range results {
		out[i] = r.srv
	}
	return out, nil
}

func (s *serverService) enrichServersWithCredentials(servers []domain.Server) {
	if s.credentialStore == nil {
		return
	}
	for i := range servers {
		if pwd, err := s.credentialStore.GetPassword(servers[i].Alias); err == nil && pwd != "" {
			servers[i].Password = pwd
		}
	}
}

// fuzzyScore computes a VS Code–like fuzzy subsequence score for q against s.
// Returns 0 if q is not a subsequence of s (case-insensitive).
func fuzzyScore(q, s string) int {
	if q == "" || s == "" {
		return 0
	}

	// Preprocess to run subsequence on lowercased forms, but keep original for case bonuses.
	ql := []rune(strings.ToLower(q))
	sl := []rune(strings.ToLower(s))
	sr := []rune(s)

	positions := make([]int, 0, len(ql))
	si := 0
	for qi := 0; qi < len(ql); qi++ {
		found := -1
		for ; si < len(sl); si++ {
			if sl[si] == ql[qi] {
				found = si
				positions = append(positions, si)
				si++
				break
			}
		}
		if found == -1 {
			return 0 // not a subsequence
		}
	}

	if len(positions) == 0 {
		return 0
	}

	score := 0

	// Base: +1 per matched char
	score += len(positions)

	// Early start bonus
	startIdx := positions[0]
	if startIdx < 20 {
		score += (20 - startIdx)
	}

	// Adjacency bonus and gap penalty
	totalGapPenalty := 0
	for i := 1; i < len(positions); i++ {
		if positions[i] == positions[i-1]+1 {
			score += 5
		} else {
			gap := positions[i] - positions[i-1] - 1
			if gap > 0 {
				totalGapPenalty += gap
			}
		}
	}
	if totalGapPenalty > 15 {
		totalGapPenalty = 15
	}
	score -= totalGapPenalty

	// Word boundary bonus and case-sensitive bonus
	for idx, pos := range positions {
		var prev rune
		if pos > 0 {
			prev = sr[pos-1]
		}
		curr := sr[pos]
		if isWordBoundary(prev, curr, pos) {
			if pos == 0 {
				score += 8
			} else {
				score += 6
			}
		}
		// Case bonus if rune matches exactly (same case) at this position.
		if idx < len([]rune(q)) {
			if []rune(q)[idx] == curr {
				score += 1
			}
		}
	}

	return score
}

func isWordBoundary(prev, curr rune, idx int) bool {
	if idx == 0 {
		return true
	}
	if prev == '-' || prev == '_' || prev == '.' || prev == '/' || unicode.IsSpace(prev) {
		return true
	}
	// camelCase boundary: previous is lower and current is upper
	if unicode.IsLower(prev) && unicode.IsUpper(curr) {
		return true
	}
	return false
}

// validateServer performs core validation of server fields.
func validateServer(srv domain.Server) error {
	if strings.TrimSpace(srv.Alias) == "" {
		return fmt.Errorf("alias is required")
	}
	if ok, _ := regexp.MatchString(`^[A-Za-z0-9_.-]+$`, srv.Alias); !ok {
		return fmt.Errorf("alias may contain letters, digits, dot, dash, underscore")
	}
	if strings.TrimSpace(srv.Host) == "" {
		return fmt.Errorf("Host/IP is required")
	}
	if ip := net.ParseIP(srv.Host); ip == nil {
		if strings.Contains(srv.Host, " ") {
			return fmt.Errorf("host must not contain spaces")
		}
		if ok, _ := regexp.MatchString(`^[A-Za-z0-9.-]+$`, srv.Host); !ok {
			return fmt.Errorf("host contains invalid characters")
		}
		if strings.HasPrefix(srv.Host, ".") || strings.HasSuffix(srv.Host, ".") {
			return fmt.Errorf("host must not start or end with a dot")
		}
		for _, lbl := range strings.Split(srv.Host, ".") {
			if lbl == "" {
				return fmt.Errorf("host must not contain empty labels")
			}
			if strings.HasPrefix(lbl, "-") || strings.HasSuffix(lbl, "-") {
				return fmt.Errorf("hostname labels must not start or end with a hyphen")
			}
		}
	}
	if srv.Port != 0 && (srv.Port < 1 || srv.Port > 65535) {
		return fmt.Errorf("port must be a number between 1 and 65535")
	}
	return nil
}

// UpdateServer updates an existing server with new details.
func (s *serverService) UpdateServer(server domain.Server, newServer domain.Server) error {
	if s.readonly {
		return ErrReadOnly
	}
	if err := validateServer(newServer); err != nil {
		s.logger.Warnw("validation failed on update", "error", err, "server", newServer)
		return err
	}
	err := s.serverRepository.UpdateServer(server, newServer)
	if err != nil {
		s.logger.Errorw("failed to update server", "error", err, "server", server)
		return err
	}
	if s.credentialStore != nil {
		if newServer.Password != "" {
			_ = s.credentialStore.SetPassword(newServer.Alias, newServer.Password)
		} else if server.Password != "" {
			_ = s.credentialStore.DeletePassword(server.Alias)
		}
		if newServer.Alias != server.Alias && server.Password != "" {
			_ = s.credentialStore.DeletePassword(server.Alias)
		}
	}
	s.serversMu.Lock()
	_ = s.loadFromRepoLocked()
	s.serversMu.Unlock()
	return nil
}

// AddServer adds a new server to the repository.
func (s *serverService) AddServer(server domain.Server) error {
	if s.readonly {
		return ErrReadOnly
	}
	if err := validateServer(server); err != nil {
		s.logger.Warnw("validation failed on add", "error", err, "server", server)
		return err
	}
	err := s.serverRepository.AddServer(server)
	if err != nil {
		s.logger.Errorw("failed to add server", "error", err, "server", server)
		return err
	}
	if s.credentialStore != nil && server.Password != "" {
		_ = s.credentialStore.SetPassword(server.Alias, server.Password)
	}
	s.serversMu.Lock()
	_ = s.loadFromRepoLocked()
	s.serversMu.Unlock()
	return nil
}

// DeleteServer removes a server from the repository.
func (s *serverService) DeleteServer(server domain.Server) error {
	if s.readonly {
		return ErrReadOnly
	}
	err := s.serverRepository.DeleteServer(server)
	if err != nil {
		s.logger.Errorw("failed to delete server", "error", err, "server", server)
		return err
	}
	if s.credentialStore != nil {
		_ = s.credentialStore.DeletePassword(server.Alias)
	}
	s.serversMu.Lock()
	_ = s.loadFromRepoLocked()
	s.serversMu.Unlock()
	return nil
}

// SetPinned sets or clears a pin timestamp for the server alias.
func (s *serverService) SetPinned(alias string, pinned bool) error {
	err := s.serverRepository.SetPinned(alias, pinned)
	if err != nil {
		s.logger.Errorw("failed to set pin state", "error", err, "alias", alias, "pinned", pinned)
		return err
	}
	s.serversMu.Lock()
	_ = s.loadFromRepoLocked()
	s.serversMu.Unlock()
	return nil
}

// SetHidden sets or clears hidden status for the server alias.
func (s *serverService) SetHidden(alias string, hidden bool) error {
	if s.readonly {
		return ErrReadOnly
	}
	err := s.serverRepository.SetHidden(alias, hidden)
	if err != nil {
		s.logger.Errorw("failed to set hidden state", "error", err, "alias", alias, "hidden", hidden)
		return err
	}
	s.serversMu.Lock()
	_ = s.loadFromRepoLocked()
	s.serversMu.Unlock()
	return nil
}

var terminalTitleWriter io.Writer = os.Stdout

// SetTerminalTitle sets the terminal emulator window/tab title using standard OSC 0 sequence.
func SetTerminalTitle(title string) {
	if terminalTitleWriter != nil {
		_, _ = fmt.Fprintf(terminalTitleWriter, "\033]0;%s\007", title)
	}
}

// RestoreTerminalTitle restores the terminal emulator window/tab title to default.
func RestoreTerminalTitle() {
	if terminalTitleWriter != nil {
		_, _ = fmt.Fprint(terminalTitleWriter, "\033]0;\007")
	}
}

func (s *serverService) formatTerminalTitle(alias string) string {
	title := alias
	if servers, err := s.ListServers(alias); err == nil {
		for _, srv := range servers {
			if strings.EqualFold(srv.Alias, alias) {
				if srv.Host != "" && !strings.EqualFold(srv.Host, alias) {
					title = fmt.Sprintf("%s (%s)", alias, srv.Host)
				}
				break
			}
		}
	}
	return title
}

// interpolateHookCommand substitutes OpenSSH-style expansion tokens (%h, %p, %r, %n, %%) in the hook template.
func interpolateHookCommand(template string, s domain.Server) string {
	portStr := strconv.Itoa(s.Port)
	if s.Port == 0 {
		portStr = "22"
	}
	r := strings.NewReplacer(
		"%h", s.Host,
		"%p", portStr,
		"%r", s.User,
		"%n", s.Alias,
		"%%", "%",
	)
	return r.Replace(template)
}

func (s *serverService) runPreConnectHook(alias string) error {
	servers, err := s.ListServers("")
	if err != nil {
		s.logger.Warnw("failed to list servers for pre-connect hook lookup", "error", err)
	}

	var target *domain.Server
	for i := range servers {
		if strings.EqualFold(servers[i].Alias, alias) {
			target = &servers[i]
			break
		}
	}

	var hooks []string

	// 1. Global hook from environment variable
	if envHook := strings.TrimSpace(os.Getenv("NEOSSH_PRE_CONNECT_HOOK")); envHook != "" {
		hooks = append(hooks, envHook)
	}

	// 2. Global hook from repository settings
	if globalHook, err := s.serverRepository.GetPreConnectCommand(); err == nil && strings.TrimSpace(globalHook) != "" {
		gh := strings.TrimSpace(globalHook)
		if !slices.Contains(hooks, gh) {
			hooks = append(hooks, gh)
		}
	}

	// 3. Server-specific hook
	if target != nil && strings.TrimSpace(target.PreConnectCommand) != "" {
		sh := strings.TrimSpace(target.PreConnectCommand)
		if !slices.Contains(hooks, sh) {
			hooks = append(hooks, sh)
		}
	}

	if len(hooks) == 0 {
		return nil
	}

	srv := domain.Server{Alias: alias, Host: alias, Port: 22}
	if target != nil {
		srv = *target
		if srv.Port == 0 {
			srv.Port = 22
		}
	}

	for _, hookTemplate := range hooks {
		cmdStr := interpolateHookCommand(hookTemplate, srv)
		s.logger.Infow("executing pre-connect hook", "alias", alias, "command", cmdStr)

		hookCmdFactory := s.newHookCommand
		if hookCmdFactory == nil {
			hookCmdFactory = func(c string) *exec.Cmd {
				if runtime.GOOS == "windows" {
					//nolint:gosec // G204: intentional user pre-connect command hook
					return exec.Command("cmd.exe", "/c", c)
				}
				//nolint:gosec // G204: intentional user pre-connect command hook
				return exec.Command("sh", "-c", c)
			}
		}

		cmd := hookCmdFactory(cmdStr)
		if cmd == nil {
			return fmt.Errorf("pre-connect hook factory returned nil")
		}

		cmd.Env = append(cmd.Environ(),
			"NEOSSH_ALIAS="+srv.Alias,
			"NEOSSH_HOST="+srv.Host,
			"NEOSSH_USER="+srv.User,
			"NEOSSH_PORT="+strconv.Itoa(srv.Port),
			"NEOSSH_CONFIG="+s.serverRepository.GetConfigFile(),
		)

		stderrBuf := newLimitedBuffer(2048)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = io.MultiWriter(os.Stderr, stderrBuf)

		if err := cmd.Run(); err != nil {
			s.logger.Errorw("pre-connect hook failed", "alias", alias, "command", cmdStr, "error", err)
			msg := strings.TrimSpace(stderrBuf.String())
			if msg != "" {
				return fmt.Errorf("pre-connect hook failed (%s): %s", cmdStr, msg)
			}
			return fmt.Errorf("pre-connect hook failed (%s): %w", cmdStr, err)
		}
	}

	return nil
}

func (s *serverService) ensureValidCertificate(alias string) error {
	servers, err := s.ListServers("")
	if err != nil {
		s.logger.Warnw("failed to list servers for certificate renewal check", "error", err)
	}

	var target *domain.Server
	for i := range servers {
		if strings.EqualFold(servers[i].Alias, alias) {
			target = &servers[i]
			break
		}
	}

	if target == nil {
		return nil
	}

	certCmd := strings.TrimSpace(target.CertificateCommand)
	if certCmd == "" {
		return nil
	}

	cert := domain.InspectServerCertificate(*target)

	// If certificate is valid and not expiring soon, no renewal needed!
	if cert != nil && cert.FileExists && cert.Status == domain.CertStatusValid {
		s.logger.Infow("SSH certificate is valid, skipping renewal", "alias", alias, "remaining", cert.TimeRemaining)
		return nil
	}

	// Certificate is missing, expired, or expiring soon. Execute renewal command.
	s.logger.Infow("SSH certificate needs acquisition/renewal, executing certificate command",
		"alias", alias, "status", func() string {
			if cert == nil {
				return "missing"
			}
			return string(cert.Status)
		}())

	srv := *target
	if srv.Port == 0 {
		srv.Port = 22
	}
	cmdStr := interpolateHookCommand(certCmd, srv)

	hookCmdFactory := s.newHookCommand
	if hookCmdFactory == nil {
		hookCmdFactory = func(c string) *exec.Cmd {
			if runtime.GOOS == "windows" {
				//nolint:gosec // G204: intentional user certificate command hook
				return exec.Command("cmd.exe", "/c", c)
			}
			//nolint:gosec // G204: intentional user certificate command hook
			return exec.Command("sh", "-c", c)
		}
	}

	cmd := hookCmdFactory(cmdStr)
	if cmd == nil {
		return fmt.Errorf("certificate command factory returned nil")
	}

	cmd.Env = append(cmd.Environ(),
		"NEOSSH_ALIAS="+srv.Alias,
		"NEOSSH_HOST="+srv.Host,
		"NEOSSH_USER="+srv.User,
		"NEOSSH_PORT="+strconv.Itoa(srv.Port),
		"NEOSSH_CONFIG="+s.serverRepository.GetConfigFile(),
	)

	stderrBuf := newLimitedBuffer(2048)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = io.MultiWriter(os.Stderr, stderrBuf)

	if err := cmd.Run(); err != nil {
		s.logger.Errorw("certificate renewal command failed", "alias", alias, "command", cmdStr, "error", err)
		msg := strings.TrimSpace(stderrBuf.String())
		if msg != "" {
			return fmt.Errorf("certificate command failed (%s): %s", cmdStr, msg)
		}
		return fmt.Errorf("certificate command failed (%s): %w", cmdStr, err)
	}

	// Verify that a valid certificate is now present on disk
	newCert := domain.InspectServerCertificate(*target)
	if newCert == nil || !newCert.FileExists {
		certPath, _ := domain.ResolveServerCertificatePath(*target)
		if certPath == "" {
			switch {
			case target.CertificateFile != "":
				certPath = target.CertificateFile
			case len(target.IdentityFiles) > 0:
				certPath = target.IdentityFiles[0] + "-cert.pub"
			default:
				certPath = "certificate file (unspecified CertificateFile or IdentityFile-cert.pub)"
			}
		}
		return fmt.Errorf("certificate command succeeded but certificate %s was not found on disk", certPath)
	}

	if newCert.Status == domain.CertStatusExpired {
		return fmt.Errorf("certificate command succeeded but certificate at %q is expired", newCert.Path)
	}
	if newCert.Status == domain.CertStatusNotYetValid {
		return fmt.Errorf("certificate command succeeded but certificate at %q is not yet valid", newCert.Path)
	}
	if newCert.Status != domain.CertStatusValid && newCert.Status != domain.CertStatusExpiringSoon {
		errMsg := newCert.KeyID
		if errMsg == "" {
			errMsg = "not a valid SSH certificate"
		}
		return fmt.Errorf("certificate command succeeded but certificate at %q is invalid: %s", newCert.Path, errMsg)
	}

	s.logger.Infow("SSH certificate successfully verified after renewal", "alias", alias, "status", newCert.Status)
	return nil
}

func (s *serverService) getPasswordForServer(alias string) string {
	if pwd := os.Getenv("NEOSSH_PASSWORD"); pwd != "" {
		return pwd
	}
	if pwd := os.Getenv("SSHPASS"); pwd != "" {
		return pwd
	}
	if s.credentialStore != nil {
		if pwd, err := s.credentialStore.GetPassword(alias); err == nil && pwd != "" {
			return pwd
		}
	}
	servers, err := s.ListServers("")
	if err != nil {
		return ""
	}
	for i := range servers {
		if strings.EqualFold(servers[i].Alias, alias) {
			return servers[i].Password
		}
	}
	return ""
}

func (s *serverService) wrapWithSSHPass(alias string, cmd *exec.Cmd) (*exec.Cmd, error) {
	pwd := s.getPasswordForServer(alias)
	if pwd == "" {
		return cmd, nil
	}

	lookPath := s.lookPath
	if lookPath == nil {
		lookPath = exec.LookPath
	}
	sshpassPath, err := lookPath("sshpass")
	if err != nil {
		return nil, fmt.Errorf(
			"password authentication configured for %q, but 'sshpass' is not installed in PATH: please install sshpass",
			alias,
		)
	}

	if s.newSSHPassCommand == nil {
		return cmd, nil
	}

	wrapped := s.newSSHPassCommand(sshpassPath, pwd, cmd)
	if wrapped == nil {
		return nil, fmt.Errorf("sshpass command creation failed")
	}
	return wrapped, nil
}

// SSH starts an interactive SSH session to the given alias using the system's ssh client.
func (s *serverService) SSH(alias string) error {
	s.logger.Infow("ssh start", "alias", alias)
	if strings.ContainsAny(alias, "*?") {
		return fmt.Errorf("cannot initiate direct SSH connection to a wildcard pattern block")
	}

	if err := s.ensureValidCertificate(alias); err != nil {
		return err
	}

	if err := s.runPreConnectHook(alias); err != nil {
		return err
	}

	title := s.formatTerminalTitle(alias)
	SetTerminalTitle(title)
	defer RestoreTerminalTitle()

	cmdFactory := s.newSSHCommand
	if cmdFactory == nil {
		cmdFactory = func(a string) *exec.Cmd {
			//nolint:gosec // G204: intentional SSH command
			return exec.Command("ssh", "-F", s.serverRepository.GetConfigFile(), a)
		}
	}
	cmd := cmdFactory(alias)
	if cmd == nil {
		err := fmt.Errorf("ssh command factory returned nil")
		s.logger.Errorw("ssh command creation failed", "alias", alias, "error", err)
		return err
	}
	cmd, err := s.wrapWithSSHPass(alias, cmd)
	if err != nil {
		s.logger.Errorw("sshpass wrap failed", "alias", alias, "error", err)
		return err
	}
	stderrBuf := newLimitedBuffer(2048)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = io.MultiWriter(os.Stderr, stderrBuf)
	if err := cmd.Run(); err != nil {
		if isRemoteDisconnectError(err, stderrBuf.String()) {
			s.logger.Infow("ssh session ended by remote", "alias", alias)
		} else {
			s.logger.Errorw("ssh command failed", "alias", alias, "error", err)
			msg := strings.TrimSpace(stderrBuf.String())
			if msg != "" {
				return fmt.Errorf("%s", msg)
			}
			return err
		}
	}

	if err := s.serverRepository.RecordSSH(alias); err != nil {
		s.logger.Errorw("failed to record ssh metadata", "alias", alias, "error", err)
	} else {
		s.serversMu.Lock()
		for i := range s.servers {
			if strings.EqualFold(s.servers[i].Alias, alias) {
				s.servers[i].LastSeen = time.Now()
				s.servers[i].SSHCount++
				break
			}
		}
		s.serversMu.Unlock()
	}

	s.logger.Infow("ssh end", "alias", alias)
	return nil
}

// SSHWithArgs runs system ssh with provided extra args (e.g., -L/-R/-D) for the given alias.
func (s *serverService) SSHWithArgs(alias string, extraArgs []string) error {
	s.logger.Infow("ssh start (with args)", "alias", alias, "args", extraArgs)
	if strings.ContainsAny(alias, "*?") {
		return fmt.Errorf("cannot initiate direct SSH connection to a wildcard pattern block")
	}

	if err := s.ensureValidCertificate(alias); err != nil {
		return err
	}

	if err := s.runPreConnectHook(alias); err != nil {
		return err
	}

	title := s.formatTerminalTitle(alias)
	SetTerminalTitle(title)
	defer RestoreTerminalTitle()

	cmdFactory := s.newSSHCommandWithArgs
	if cmdFactory == nil {
		cmdFactory = func(a string, extra []string) *exec.Cmd {
			args := append([]string{}, extra...)
			args = append(args, "-F", s.serverRepository.GetConfigFile(), a)
			//nolint:gosec // G204: intentional SSH command
			return exec.Command("ssh", args...)
		}
	}
	cmd := cmdFactory(alias, extraArgs)
	if cmd == nil {
		err := fmt.Errorf("ssh command factory returned nil")
		s.logger.Errorw("ssh command creation failed", "alias", alias, "error", err)
		return err
	}
	cmd, err := s.wrapWithSSHPass(alias, cmd)
	if err != nil {
		s.logger.Errorw("sshpass wrap failed", "alias", alias, "error", err)
		return err
	}
	stderrBuf := newLimitedBuffer(2048)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = io.MultiWriter(os.Stderr, stderrBuf)
	if err := cmd.Run(); err != nil {
		if isRemoteDisconnectError(err, stderrBuf.String()) {
			s.logger.Infow("ssh session ended by remote", "alias", alias)
		} else {
			s.logger.Errorw("ssh (with args) failed", "alias", alias, "error", err)
			msg := strings.TrimSpace(stderrBuf.String())
			if msg != "" {
				return fmt.Errorf("%s", msg)
			}
			return err
		}
	}
	if err := s.serverRepository.RecordSSH(alias); err != nil {
		s.logger.Errorw("failed to record ssh metadata", "alias", alias, "error", err)
	} else {
		s.serversMu.Lock()
		for i := range s.servers {
			if strings.EqualFold(s.servers[i].Alias, alias) {
				s.servers[i].LastSeen = time.Now()
				s.servers[i].SSHCount++
				break
			}
		}
		s.serversMu.Unlock()
	}
	s.logger.Infow("ssh end (with args)", "alias", alias)
	return nil
}

// SFTP starts an interactive SFTP session for the given alias.
func (s *serverService) SFTP(alias string) error {
	return s.LaunchFileManager(alias, "")
}

// LaunchFileManager launches either the standard OpenSSH SFTP client or a configured
// external file manager (yazi, ranger, filezilla, cyberduck, etc.) for the given server alias.
func (s *serverService) LaunchFileManager(alias string, customTool string) error {
	s.logger.Infow("file manager launch start", "alias", alias, "customTool", customTool)
	if strings.ContainsAny(alias, "*?") {
		return fmt.Errorf("cannot initiate file transfer to a wildcard pattern block")
	}

	servers, err := s.ListServers("")
	if err != nil {
		return fmt.Errorf("failed to load servers: %w", err)
	}

	var target *domain.Server
	for i := range servers {
		if strings.EqualFold(servers[i].Alias, alias) || hasAlias(servers[i].Aliases, alias) {
			target = &servers[i]
			break
		}
	}

	if target == nil {
		return fmt.Errorf("server with alias %q not found", alias)
	}

	if target.IsWildcard || target.IsWildcardServer() {
		return fmt.Errorf("cannot initiate file transfer to a wildcard pattern block")
	}

	if err := s.ensureValidCertificate(alias); err != nil {
		return err
	}

	if err := s.runPreConnectHook(alias); err != nil {
		return err
	}

	tool := strings.TrimSpace(customTool)
	if tool == "" {
		if configuredTool, err := s.serverRepository.GetFileManager(); err == nil && strings.TrimSpace(configuredTool) != "" {
			tool = strings.TrimSpace(configuredTool)
		}
	}
	if tool == "" {
		tool = "sftp"
	}

	host := target.Host
	user := target.User
	port := target.Port
	if port <= 0 {
		port = 22
	}

	sftpURL := buildSFTPURL(user, host, port)
	fishURL := buildFishURL(user, host, port)
	configFile := s.serverRepository.GetConfigFile()

	cmd, isGUI := s.buildFileManagerCmd(tool, *target, sftpURL, fishURL, configFile)

	if strings.EqualFold(tool, "sftp") || tool == "" {
		cmd, err = s.wrapWithSSHPass(alias, cmd)
		if err != nil {
			s.logger.Errorw("sshpass wrap failed for sftp", "alias", alias, "error", err)
			return err
		}
	}

	if isGUI {
		if err := cmd.Start(); err != nil {
			s.logger.Errorw("failed to start GUI file manager", "tool", tool, "alias", alias, "error", err)
			return fmt.Errorf("failed to launch %s: %w", tool, err)
		}
		s.logger.Infow("GUI file manager launched", "tool", tool, "alias", alias)
		_ = s.serverRepository.RecordSSH(alias)
		return nil
	}

	title := fmt.Sprintf("%s (SFTP)", s.formatTerminalTitle(alias))
	SetTerminalTitle(title)
	defer RestoreTerminalTitle()

	stderrBuf := newLimitedBuffer(2048)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = io.MultiWriter(os.Stderr, stderrBuf)

	if err := cmd.Run(); err != nil {
		if isRemoteDisconnectError(err, stderrBuf.String()) {
			s.logger.Infow("sftp session ended by remote", "alias", alias)
		} else {
			s.logger.Errorw("file manager command failed", "alias", alias, "tool", tool, "error", err)
			msg := strings.TrimSpace(stderrBuf.String())
			if msg != "" {
				return fmt.Errorf("%s", msg)
			}
			return err
		}
	}

	if err := s.serverRepository.RecordSSH(alias); err != nil {
		s.logger.Errorw("failed to record ssh metadata after sftp", "alias", alias, "error", err)
	}

	s.logger.Infow("file manager session ended", "alias", alias, "tool", tool)
	return nil
}

func buildSFTPURL(user, host string, port int) string {
	if port == 22 || port <= 0 {
		if user != "" {
			return fmt.Sprintf("sftp://%s@%s/", user, host)
		}
		return fmt.Sprintf("sftp://%s/", host)
	}
	if user != "" {
		return fmt.Sprintf("sftp://%s@%s:%d/", user, host, port)
	}
	return fmt.Sprintf("sftp://%s:%d/", host, port)
}

func buildFishURL(user, host string, port int) string {
	if port == 22 || port <= 0 {
		if user != "" {
			return fmt.Sprintf("fish://%s@%s/", user, host)
		}
		return fmt.Sprintf("fish://%s/", host)
	}
	if user != "" {
		return fmt.Sprintf("fish://%s@%s:%d/", user, host, port)
	}
	return fmt.Sprintf("fish://%s:%d/", host, port)
}

func (s *serverService) buildFileManagerCmd(
	tool string, target domain.Server, sftpURL, fishURL, configFile string,
) (*exec.Cmd, bool) {
	alias := target.Alias
	normalized := strings.ToLower(strings.TrimSpace(tool))

	if s.newFileManagerCommand != nil {
		cmd := s.newFileManagerCommand(tool, []string{alias, sftpURL, fishURL})
		return cmd, isGUIFileManager(tool)
	}

	switch normalized {
	case "sftp", "":
		if s.newSFTPCommand != nil {
			cmd := s.newSFTPCommand(alias, nil)
			return cmd, false
		}
		//nolint:gosec // G204: intentional SFTP command
		return exec.Command("sftp", "-F", configFile, alias), false

	case "yazi":
		//nolint:gosec // G204: intentional user file manager
		return exec.Command("yazi", sftpURL), false

	case "ranger":
		//nolint:gosec // G204: intentional user file manager
		return exec.Command("ranger", sftpURL), false

	case "filezilla":
		//nolint:gosec // G204: intentional user file manager
		return exec.Command("filezilla", sftpURL), true

	case "cyberduck":
		if runtime.GOOS == "darwin" {
			//nolint:gosec // G204: intentional open command on macOS
			return exec.Command("open", "-a", "Cyberduck", sftpURL), true
		}
		//nolint:gosec // G204: intentional user file manager
		return exec.Command("cyberduck", sftpURL), true

	case "nautilus":
		//nolint:gosec // G204: intentional user file manager
		return exec.Command("nautilus", sftpURL), true

	case "dolphin":
		//nolint:gosec // G204: intentional user file manager
		return exec.Command("dolphin", fishURL), true

	default:
		cmdStr := tool
		cmdStr = strings.ReplaceAll(cmdStr, "%url", sftpURL)
		cmdStr = strings.ReplaceAll(cmdStr, "%sftp_url", sftpURL)
		cmdStr = strings.ReplaceAll(cmdStr, "%sftp", sftpURL)
		cmdStr = strings.ReplaceAll(cmdStr, "%fish_url", fishURL)
		cmdStr = strings.ReplaceAll(cmdStr, "%fish", fishURL)
		cmdStr = strings.ReplaceAll(cmdStr, "%a", alias)
		cmdStr = strings.ReplaceAll(cmdStr, "%n", alias)
		cmdStr = strings.ReplaceAll(cmdStr, "%h", target.Host)
		cmdStr = strings.ReplaceAll(cmdStr, "%u", target.User)
		cmdStr = strings.ReplaceAll(cmdStr, "%r", target.User)
		cmdStr = strings.ReplaceAll(cmdStr, "%p", strconv.Itoa(target.Port))
		cmdStr = strings.ReplaceAll(cmdStr, "%c", configFile)

		isGUI := isGUIFileManager(tool)
		if runtime.GOOS == "windows" {
			//nolint:gosec // G204: intentional custom file manager command
			return exec.Command("cmd.exe", "/c", cmdStr), isGUI
		}
		//nolint:gosec // G204: intentional custom file manager command
		return exec.Command("sh", "-c", cmdStr), isGUI
	}
}

func isGUIFileManager(tool string) bool {
	lower := strings.ToLower(strings.TrimSpace(tool))
	switch lower {
	case "filezilla", "cyberduck", "nautilus", "dolphin", "thunar", "nemo", "pcmanfm", "open", "xdg-open":
		return true
	default:
		return false
	}
}

func isRemoteDisconnectError(err error, stderr string) bool {
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		return false
	}

	lower := strings.ToLower(stderr)
	disconnectSignals := []string{
		"connection closed by remote host",
		"connection closed by foreign host",
		"closed by remote host",
		"closed by foreign host",
		"remote host closed the connection",
		"connection closed",
		"connection reset by peer",
		"connection to ",
		"kex_exchange_identification",
		"packet_write_wait",
		"broken pipe",
		"software caused connection abort",
	}

	for _, signal := range disconnectSignals {
		if strings.Contains(lower, signal) {
			return true
		}
	}

	return false
}

type limitedBuffer struct {
	buf   bytes.Buffer
	limit int
}

func newLimitedBuffer(limit int) *limitedBuffer {
	return &limitedBuffer{limit: limit}
}

func (b *limitedBuffer) Write(p []byte) (int, error) {
	if b == nil || b.limit <= 0 {
		return len(p), nil
	}

	remaining := b.limit - b.buf.Len()
	if remaining <= 0 {
		return len(p), nil
	}

	toWrite := p
	if len(p) > remaining {
		toWrite = p[:remaining]
	}

	if _, err := b.buf.Write(toWrite); err != nil {
		return 0, err
	}
	return len(p), nil
}

func (b *limitedBuffer) String() string {
	if b == nil {
		return ""
	}
	return b.buf.String()
}

// CopySSHKey installs public SSH keys to the remote host using ssh-copy-id.
func (s *serverService) CopySSHKey(alias string) error {
	if s.readonly {
		return ErrReadOnly
	}
	s.logger.Infow("ssh-copy-id start", "alias", alias)
	if strings.ContainsAny(alias, "*?") {
		return fmt.Errorf("cannot install SSH key to a wildcard pattern block")
	}

	if _, err := exec.LookPath("ssh-copy-id"); err != nil {
		s.logger.Errorw("ssh-copy-id missing", "error", err)
		return fmt.Errorf("ssh-copy-id not found; please install OpenSSH (e.g. brew install openssh)")
	}

	var args []string
	if cfg := s.serverRepository.GetConfigFile(); cfg != "" {
		args = append(args, "-F", cfg)
	}
	args = append(args, alias)

	//nolint:gosec // G204: intentional ssh-copy-id command execution
	cmd := exec.Command("ssh-copy-id", args...)
	stderrBuf := newLimitedBuffer(2048)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = io.MultiWriter(os.Stderr, stderrBuf)

	if err := cmd.Run(); err != nil {
		s.logger.Errorw("ssh-copy-id command failed", "alias", alias, "error", err)
		msg := strings.TrimSpace(stderrBuf.String())
		if msg != "" {
			return fmt.Errorf("%s", msg)
		}
		return err
	}

	s.logger.Infow("ssh-copy-id end", "alias", alias)
	return nil
}

// StartForward starts ssh port forwarding in the background and tracks the process.
func (s *serverService) StartForward(alias string, extraArgs []string) (int, error) {
	s.fwMu.Lock()
	if s.forwards == nil {
		s.forwards = make(map[string][]*os.Process)
	}
	s.fwMu.Unlock()

	extraArgs = append(extraArgs, "-N", alias)

	// #nosec G204
	cmd := exec.Command("ssh", extraArgs...)

	// Detach from TTY: discard stdio
	devNull, err := os.OpenFile(os.DevNull, os.O_RDWR, 0)
	if err != nil {
		return 0, fmt.Errorf("failed to open devnull: %w", err)
	}
	defer func() {
		if devNull != nil {
			_ = devNull.Close()
		}
	}()

	cmd.Stdin = devNull
	cmd.Stdout = devNull
	cmd.Stderr = devNull
	// Set SysProcAttr in an OS-specific way (see sysprocattr_* files)
	sysProcAttr := &syscall.SysProcAttr{}
	setDetach(sysProcAttr)
	cmd.SysProcAttr = sysProcAttr

	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("failed to start ssh: %w", err)
	}

	proc := cmd.Process
	if proc == nil {
		return 0, fmt.Errorf("process is nil after start")
	}
	pid := proc.Pid

	// Track process
	s.fwMu.Lock()
	s.forwards[alias] = append(s.forwards[alias], proc)
	s.fwMu.Unlock()

	// Cleanup on exit
	go func(a string, c *exec.Cmd, dn *os.File) {
		_ = c.Wait()
		_ = dn.Close()

		s.fwMu.Lock()
		defer s.fwMu.Unlock()

		procs := s.forwards[a]
		if len(procs) == 0 {
			return
		}

		filtered := make([]*os.Process, 0, len(procs))
		for _, p := range procs {
			if p != nil && p.Pid != pid {
				filtered = append(filtered, p)
			}
		}

		if len(filtered) == 0 {
			delete(s.forwards, a)
		} else {
			s.forwards[a] = filtered
		}
	}(alias, cmd, devNull)

	devNull = nil // Prevent defer from closing it

	return pid, nil
}

// StopForwarding kills all active forward processes for the alias.
func (s *serverService) StopForwarding(alias string) error {
	s.fwMu.Lock()
	procs := s.forwards[alias]
	delete(s.forwards, alias)
	s.fwMu.Unlock()

	if len(procs) == 0 {
		return nil
	}

	var errs []error
	for _, p := range procs {
		if p != nil {
			if err := p.Signal(syscall.SIGTERM); err != nil {
				// If SIGTERM fails, try SIGKILL
				if killErr := p.Kill(); killErr != nil {
					errs = append(errs, fmt.Errorf("failed to kill pid %d: %w", p.Pid, killErr))
				}
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors stopping forwards: %v", errs)
	}
	return nil
}

// IsForwarding reports whether there is at least one active forward for alias.
func (s *serverService) IsForwarding(alias string) bool {
	s.fwMu.Lock()
	defer s.fwMu.Unlock()
	return len(s.forwards[alias]) > 0
}

// Ping checks if the server is reachable on its SSH port.
func (s *serverService) Ping(server domain.Server) (bool, time.Duration, error) {
	if server.IsWildcardServer() {
		return false, 0, fmt.Errorf("cannot ping a wildcard pattern block")
	}
	start := time.Now()

	host, port, ok := resolveSSHDestination(server.Alias)
	if !ok {

		host = strings.TrimSpace(server.Host)
		if host == "" {
			host = server.Alias
		}
		if server.Port > 0 {
			port = server.Port
		} else {
			port = 22
		}
	}
	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))

	dialer := net.Dialer{Timeout: 3 * time.Second}
	conn, err := dialer.Dial("tcp", addr)
	if err != nil {
		return false, time.Since(start), err
	}
	_ = conn.Close()
	return true, time.Since(start), nil
}

// resolveSSHDestination uses `ssh -G <alias>` to extract HostName and Port from the user's SSH config.
// Returns host, port, ok where ok=false if resolution failed.
func resolveSSHDestination(alias string) (string, int, bool) {
	alias = strings.TrimSpace(alias)
	if alias == "" {
		return "", 0, false
	}
	//nolint:gosec // G204: alias is passed as an SSH config host argument to resolve destination
	cmd := exec.Command("ssh", "-G", alias)
	out, err := cmd.Output()
	if err != nil {
		return "", 0, false
	}
	host := ""
	port := 0
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "hostname ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				host = parts[1]
			}
		}
		if strings.HasPrefix(line, "port ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				if p, err := strconv.Atoi(parts[1]); err == nil {
					port = p
				}
			}
		}
	}
	if host == "" {
		host = alias
	}
	if port == 0 {
		port = 22
	}
	return host, port, true
}

// DiscoverKnownHosts discovers hosts from known_hosts without modifying the configuration.
func (s *serverService) DiscoverKnownHosts(knownHostsPath string) ([]domain.Server, domain.ImportResult, error) {
	return s.serverRepository.DiscoverKnownHosts(knownHostsPath)
}

// ImportKnownHosts imports unconfigured hosts from known_hosts into the SSH configuration.
func (s *serverService) ImportKnownHosts(knownHostsPath string) (domain.ImportResult, error) {
	if s.readonly {
		return domain.ImportResult{}, ErrReadOnly
	}
	res, err := s.serverRepository.ImportKnownHosts(knownHostsPath)
	if err == nil && res.Imported > 0 {
		s.serversMu.Lock()
		_ = s.loadFromRepoLocked()
		s.serversMu.Unlock()
	}
	return res, err
}

// GetTheme returns the current theme name from settings.
func (s *serverService) GetTheme() (string, error) {
	return s.serverRepository.GetTheme()
}

// SaveTheme saves the theme name to settings.
func (s *serverService) SaveTheme(theme string) error {
	return s.serverRepository.SaveTheme(theme)
}

// GetDefaultIdentityKey returns the default identity SSH key from environment or repository settings.
func (s *serverService) GetDefaultIdentityKey() (string, error) {
	if envKey := os.Getenv("NEOSSH_DEFAULT_KEY"); envKey != "" {
		return envKey, nil
	}
	if envKey := os.Getenv("NEOSSH_DEFAULT_IDENTITY_KEY"); envKey != "" {
		return envKey, nil
	}
	return s.serverRepository.GetDefaultIdentityKey()
}

// SaveDefaultIdentityKey saves the default identity SSH key to repository settings.
func (s *serverService) SaveDefaultIdentityKey(key string) error {
	return s.serverRepository.SaveDefaultIdentityKey(key)
}

const unknownLabel = "unknown"

var psCommand = func() *exec.Cmd {
	return exec.Command("ps", "-ax", "-o", "pid=", "-o", "comm=", "-o", "args=")
}

var killPIDFunc = func(pid int) error {
	proc, findErr := os.FindProcess(pid)
	if findErr != nil {
		return findErr
	}
	return proc.Signal(syscall.SIGTERM)
}

type activeSSHSession struct {
	alias          string
	host           string
	user           string
	port           int
	identityFiles  []string
	localForward   []string
	remoteForward  []string
	dynamicForward []string
	pid            int
}

// ListActiveSessions returns servers representing active SSH processes.
func (s *serverService) ListActiveSessions(query string) ([]domain.Server, error) {
	activeSessions, err := s.listActiveSSHSessions()
	if err != nil {
		return nil, err
	}

	configured, err := s.ListServers("")
	if err != nil {
		return nil, err
	}

	aliasIndex := make(map[string]domain.Server, len(configured))
	hostIndex := make(map[string]domain.Server, len(configured))
	for _, server := range configured {
		aliasIndex[strings.ToLower(server.Alias)] = server
		for _, alias := range server.Aliases {
			aliasIndex[strings.ToLower(alias)] = server
		}
		if server.Host != "" {
			hostIndex[strings.ToLower(server.Host)] = server
		}
	}

	query = strings.ToLower(strings.TrimSpace(query))
	entries := make([]domain.Server, 0, len(activeSessions))
	for _, session := range activeSessions {
		entry := domain.Server{}
		if session.alias != "" {
			if server, ok := aliasIndex[strings.ToLower(session.alias)]; ok {
				entry = server
			}
		}
		if entry.Alias == "" && session.host != "" {
			if server, ok := hostIndex[strings.ToLower(session.host)]; ok {
				entry = server
			}
		}

		if entry.Alias == "" {
			if session.alias != "" {
				entry.Alias = session.alias
			} else {
				entry.Alias = unknownLabel
			}
			entry.Aliases = []string{entry.Alias}
		}

		if session.host != "" {
			entry.Host = session.host
		} else if entry.Host == "" {
			entry.Host = unknownLabel
		}
		if session.user != "" {
			entry.User = session.user
		}
		if session.port > 0 {
			entry.Port = session.port
		} else if entry.Port == 0 {
			entry.Port = 22
		}

		entry.IdentityFiles = mergeIdentityFiles(entry.IdentityFiles, session.identityFiles)
		entry.LocalForward = mergeForwardSpecs(entry.LocalForward, session.localForward)
		entry.RemoteForward = mergeForwardSpecs(entry.RemoteForward, session.remoteForward)
		entry.DynamicForward = mergeForwardSpecs(entry.DynamicForward, session.dynamicForward)
		entry.ActivePID = session.pid
		entry.LastSeen = time.Now()
		entry.Tags = append([]string{"active"}, entry.Tags...)

		if query != "" && !matchesServerQuery(entry, query) {
			continue
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

// KillActiveSessions terminates active SSH sessions matching the server.
func (s *serverService) KillActiveSessions(server domain.Server) (int, error) {
	sessions, err := s.listActiveSSHSessions()
	if err != nil {
		return 0, err
	}

	if server.ActivePID > 0 {
		for _, session := range sessions {
			if session.pid == server.ActivePID {
				if err := killPIDFunc(session.pid); err != nil {
					return 0, err
				}
				return 1, nil
			}
		}
		return 0, fmt.Errorf("active ssh session not found")
	}

	var pids []int
	for _, session := range sessions {
		if matchSessionForServer(server, session) && session.pid > 0 {
			pids = append(pids, session.pid)
		}
	}
	if len(pids) == 0 {
		return 0, fmt.Errorf("no active ssh sessions found")
	}

	var errs []error
	killed := 0
	for _, pid := range pids {
		if killErr := killPIDFunc(pid); killErr != nil {
			errs = append(errs, fmt.Errorf("pid %d: %w", pid, killErr))
			continue
		}
		killed++
	}

	if len(errs) > 0 {
		return killed, fmt.Errorf("failed to terminate sessions: %v", errs)
	}
	return killed, nil
}

// ResolveConfigServer attempts to map a server entry to a configured server.
func (s *serverService) ResolveConfigServer(server domain.Server) (domain.Server, bool, error) {
	servers, err := s.ListServers("")
	if err != nil {
		return domain.Server{}, false, err
	}
	for _, candidate := range servers {
		if strings.EqualFold(candidate.Alias, server.Alias) {
			return candidate, true, nil
		}
		for _, alias := range candidate.Aliases {
			if strings.EqualFold(alias, server.Alias) {
				return candidate, true, nil
			}
		}
		if server.Host != "" && candidate.Host != "" {
			if strings.EqualFold(candidate.Host, server.Host) {
				return candidate, true, nil
			}
		}
	}
	return domain.Server{}, false, nil
}

func mergeIdentityFiles(existing []string, incoming []string) []string {
	if len(incoming) == 0 {
		return existing
	}
	seen := make(map[string]struct{}, len(existing))
	for _, v := range existing {
		seen[v] = struct{}{}
	}
	for _, v := range incoming {
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		existing = append(existing, v)
		seen[v] = struct{}{}
	}
	return existing
}

func matchSessionForServer(server domain.Server, session activeSSHSession) bool {
	if session.alias != "" {
		if strings.EqualFold(session.alias, server.Alias) {
			return true
		}
		for _, alias := range server.Aliases {
			if strings.EqualFold(session.alias, alias) {
				return true
			}
		}
	}
	if session.host != "" && server.Host != "" {
		return strings.EqualFold(session.host, server.Host)
	}
	return false
}

func mergeForwardSpecs(existing []string, incoming []string) []string {
	if len(incoming) == 0 {
		return existing
	}
	seen := make(map[string]struct{}, len(existing))
	for _, v := range existing {
		seen[v] = struct{}{}
	}
	for _, v := range incoming {
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		existing = append(existing, v)
		seen[v] = struct{}{}
	}
	return existing
}

func matchesServerQuery(server domain.Server, query string) bool {
	q := strings.TrimSpace(query)
	if q == "" {
		return true
	}
	return parseSearchQuery(q).Matches(server)
}

func (s *serverService) listActiveSSHSessions() ([]activeSSHSession, error) {
	cmd := psCommand()
	out, err := cmd.Output()
	if err != nil {
		return nil, nil
	}

	sessions := make([]activeSSHSession, 0)
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		pid, comm, args := splitPSLine(line)
		if comm != "ssh" && !strings.HasSuffix(comm, "/ssh") {
			continue
		}
		parts := strings.Fields(args)
		if len(parts) == 0 {
			continue
		}
		session := parseSSHArgs(parts)
		if session.alias == "" {
			continue
		}
		session.pid = pid
		sessions = append(sessions, session)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return sessions, nil
}

func splitPSLine(line string) (pid int, comm string, args string) {
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return 0, "", ""
	}
	if n, err := strconv.Atoi(fields[0]); err == nil {
		pid = n
	}
	comm = fields[1]
	start := strings.Index(line, comm)
	if start == -1 {
		if len(fields) > 2 {
			args = strings.Join(fields[2:], " ")
		}
		return pid, comm, strings.TrimSpace(args)
	}
	args = strings.TrimSpace(line[start+len(comm):])
	return pid, comm, args
}

func parseSSHArgs(args []string) activeSSHSession {
	if len(args) == 0 {
		return activeSSHSession{}
	}
	state := parseSSHOptions(args)
	if state.dest == "" {
		return activeSSHSession{}
	}

	host := state.dest
	if at := strings.LastIndex(state.dest, "@"); at > -1 {
		if state.user == "" {
			state.user = state.dest[:at]
		}
		host = state.dest[at+1:]
	}
	if host == "" {
		host = unknownLabel
	}
	if state.port == 0 {
		state.port = 22
	}

	return activeSSHSession{
		alias:          state.dest,
		host:           host,
		user:           state.user,
		port:           state.port,
		identityFiles:  state.identityFiles,
		localForward:   state.localForward,
		remoteForward:  state.remoteForward,
		dynamicForward: state.dynamicForward,
	}
}

type sshParseState struct {
	user           string
	port           int
	dest           string
	identityFiles  []string
	localForward   []string
	remoteForward  []string
	dynamicForward []string
}

func parseSSHOptions(args []string) sshParseState {
	state := sshParseState{
		identityFiles:  make([]string, 0),
		localForward:   make([]string, 0),
		remoteForward:  make([]string, 0),
		dynamicForward: make([]string, 0),
	}

	start := 1
	if args[0] != "ssh" && !strings.HasSuffix(args[0], "/ssh") {
		start = 0
	}
	for i := start; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			if i+1 < len(args) {
				state.dest = args[i+1]
			}
			break
		}
		if strings.HasPrefix(arg, "-") {
			if sshOptionConsumesValue(arg) {
				val, nextIdx := sshOptionValue(arg, args, i)
				i = nextIdx
				applySSHFlag(arg, val, &state)
			}
			continue
		}
		state.dest = arg
		break
	}
	return state
}

func applySSHFlag(arg, val string, state *sshParseState) {
	switch {
	case strings.HasPrefix(arg, "-p"):
		if n, err := strconv.Atoi(val); err == nil {
			state.port = n
		}
	case strings.HasPrefix(arg, "-l"):
		if val != "" {
			state.user = val
		}
	case strings.HasPrefix(arg, "-i"):
		if val != "" {
			state.identityFiles = append(state.identityFiles, val)
		}
	case strings.HasPrefix(arg, "-L"):
		if val != "" {
			state.localForward = append(state.localForward, val)
		}
	case strings.HasPrefix(arg, "-R"):
		if val != "" {
			state.remoteForward = append(state.remoteForward, val)
		}
	case strings.HasPrefix(arg, "-D"):
		if val != "" {
			state.dynamicForward = append(state.dynamicForward, val)
		}
	case strings.HasPrefix(arg, "-o"):
		applySSHOptionValue(val, state)
	}
}

func sshOptionValue(arg string, args []string, idx int) (string, int) {
	if len(arg) > 2 {
		return arg[2:], idx
	}
	if idx+1 < len(args) {
		return args[idx+1], idx + 1
	}
	return "", idx
}

func applySSHOptionValue(val string, state *sshParseState) {
	lowerVal := strings.ToLower(val)
	switch {
	case strings.HasPrefix(lowerVal, "user="):
		state.user = val[len("user="):]
	case strings.HasPrefix(lowerVal, "port="):
		if n, err := strconv.Atoi(val[len("port="):]); err == nil {
			state.port = n
		}
	case strings.HasPrefix(lowerVal, "identityfile="):
		identity := val[len("identityfile="):]
		if identity != "" {
			state.identityFiles = append(state.identityFiles, identity)
		}
	case strings.HasPrefix(lowerVal, "localforward="):
		spec := val[len("localforward="):]
		if spec != "" {
			state.localForward = append(state.localForward, spec)
		}
	case strings.HasPrefix(lowerVal, "remoteforward="):
		spec := val[len("remoteforward="):]
		if spec != "" {
			state.remoteForward = append(state.remoteForward, spec)
		}
	case strings.HasPrefix(lowerVal, "dynamicforward="):
		spec := val[len("dynamicforward="):]
		if spec != "" {
			state.dynamicForward = append(state.dynamicForward, spec)
		}
	}
}

func sshOptionConsumesValue(opt string) bool {
	base := opt
	if len(opt) > 2 && strings.HasPrefix(opt, "-") && !strings.HasPrefix(opt, "--") {
		base = opt[:2]
	}
	switch base {
	case "-p", "-l", "-i", "-o", "-F", "-b", "-c", "-D", "-E", "-e", "-I", "-J", "-L", "-m", "-O", "-Q", "-R", "-S", "-W", "-w":
		return true
	default:
		return false
	}
}
