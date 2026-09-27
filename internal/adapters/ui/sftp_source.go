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

package ui

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/WhiteRoseLK/neossh/internal/core/domain"
	"github.com/pkg/sftp"
)

// ErrFileExists is returned when attempting to transfer a file that already exists at destination.
var ErrFileExists = errors.New("file already exists at destination")

// FileEntry represents a file or directory entry in a file manager pane.
type FileEntry struct {
	Name    string
	Size    int64
	Mode    os.FileMode
	ModTime time.Time
	IsDir   bool
}

// FileSource abstracts local or remote (SFTP) file systems for dual-pane operations.
type FileSource interface {
	Title() string
	CurrentPath() string
	SetPath(newPath string) error
	List() ([]FileEntry, error)
	Open(name string) (io.ReadCloser, error)
	Create(name string) (io.WriteCloser, error)
	Stat(name string) (FileEntry, error)
	Mkdir(name string) error
	Remove(name string) error
	Close() error
}

// =============================================================================
// LocalFileSource
// =============================================================================

// LocalFileSource implements FileSource for the local host filesystem.
type LocalFileSource struct {
	currentDir string
}

// NewLocalFileSource creates a new LocalFileSource starting at initialDir or user home.
func NewLocalFileSource(initialDir ...string) (*LocalFileSource, error) {
	dir := ""
	if len(initialDir) > 0 && initialDir[0] != "" {
		dir = initialDir[0]
	}
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			home = "."
		}
		dir = home
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		abs = dir
	}
	return &LocalFileSource{currentDir: abs}, nil
}

func (l *LocalFileSource) Title() string {
	return "Local: " + l.currentDir
}

func (l *LocalFileSource) CurrentPath() string {
	return l.currentDir
}

func (l *LocalFileSource) SetPath(newPath string) error {
	p := newPath
	if strings.HasPrefix(p, "~") {
		if home, err := os.UserHomeDir(); err == nil {
			p = filepath.Join(home, strings.TrimPrefix(p, "~"))
		}
	} else if !filepath.IsAbs(p) {
		p = filepath.Join(l.currentDir, p)
	}
	clean := filepath.Clean(p)
	fi, err := os.Stat(clean)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return fmt.Errorf("%s is not a directory", clean)
	}
	l.currentDir = clean
	return nil
}

func (l *LocalFileSource) List() ([]FileEntry, error) {
	entries, err := os.ReadDir(l.currentDir)
	if err != nil {
		return nil, err
	}

	res := make([]FileEntry, 0, len(entries))
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		res = append(res, FileEntry{
			Name:    entry.Name(),
			Size:    info.Size(),
			Mode:    info.Mode(),
			ModTime: info.ModTime(),
			IsDir:   entry.IsDir(),
		})
	}
	return res, nil
}

func (l *LocalFileSource) Open(name string) (io.ReadCloser, error) {
	target := filepath.Join(l.currentDir, name)
	//nolint:gosec // G304: path within user-selected file manager directory
	return os.Open(target)
}

func (l *LocalFileSource) Create(name string) (io.WriteCloser, error) {
	target := filepath.Join(l.currentDir, name)
	//nolint:gosec // G304: path within user-selected file manager directory
	return os.Create(target)
}

func (l *LocalFileSource) Stat(name string) (FileEntry, error) {
	target := filepath.Join(l.currentDir, name)
	fi, err := os.Stat(target)
	if err != nil {
		return FileEntry{}, err
	}
	return FileEntry{
		Name:    fi.Name(),
		Size:    fi.Size(),
		Mode:    fi.Mode(),
		ModTime: fi.ModTime(),
		IsDir:   fi.IsDir(),
	}, nil
}

func (l *LocalFileSource) Mkdir(name string) error {
	target := filepath.Join(l.currentDir, name)
	return os.MkdirAll(target, 0o750)
}

func (l *LocalFileSource) Remove(name string) error {
	target := filepath.Join(l.currentDir, name)
	return os.RemoveAll(target)
}

func (l *LocalFileSource) Close() error {
	return nil
}

// =============================================================================
// SFTPFileSource
// =============================================================================

// SFTPFileSource implements FileSource over an OpenSSH SFTP pipe connection.
type SFTPFileSource struct {
	alias      string
	client     *sftp.Client
	cmd        *exec.Cmd
	currentDir string
}

// NewSFTPFileSource creates a new SFTPFileSource connected to the given server.
func NewSFTPFileSource(server domain.Server) (*SFTPFileSource, error) {
	target := server.Alias
	if target == "" {
		target = server.Host
		if server.User != "" {
			target = fmt.Sprintf("%s@%s", server.User, target)
		}
	}

	var args []string
	if server.Port > 0 && server.Port != 22 {
		args = append(args, "-p", fmt.Sprintf("%d", server.Port))
	}
	for _, kf := range server.IdentityFiles {
		if kf != "" {
			args = append(args, "-i", kf)
		}
	}
	if server.ProxyJump != "" {
		args = append(args, "-J", server.ProxyJump)
	}
	args = append(args, target, "-s", "sftp")

	// #nosec G204
	cmd := exec.Command("ssh", args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to open stdin pipe: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		_ = stdin.Close()
		return nil, fmt.Errorf("failed to open stdout pipe: %w", err)
	}

	stderrBuf := &bytes.Buffer{}
	cmd.Stderr = stderrBuf

	if err := cmd.Start(); err != nil {
		_ = stdin.Close()
		_ = stdout.Close()
		return nil, fmt.Errorf("failed to start ssh sftp process: %w", err)
	}

	client, err := sftp.NewClientPipe(stdout, stdin)
	if err != nil {
		_ = stdin.Close()
		_ = stdout.Close()
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		_ = cmd.Wait()
		msg := strings.TrimSpace(stderrBuf.String())
		if msg != "" {
			return nil, fmt.Errorf("%s: %w", msg, err)
		}
		return nil, fmt.Errorf("failed to initialize sftp client: %w", err)
	}

	// Determine starting remote directory
	currentDir := "."
	if pwd, err := client.Getwd(); err == nil && pwd != "" {
		currentDir = pwd
	}

	return &SFTPFileSource{
		alias:      server.Alias,
		client:     client,
		cmd:        cmd,
		currentDir: currentDir,
	}, nil
}

// NewSFTPFileSourceFromClient creates an SFTPFileSource wrapping an existing sftp.Client (useful for testing).
func NewSFTPFileSourceFromClient(alias string, client *sftp.Client, cmd *exec.Cmd) *SFTPFileSource {
	currentDir := "."
	if pwd, err := client.Getwd(); err == nil && pwd != "" {
		currentDir = pwd
	}
	return &SFTPFileSource{
		alias:      alias,
		client:     client,
		cmd:        cmd,
		currentDir: currentDir,
	}
}

func (s *SFTPFileSource) Title() string {
	name := s.alias
	if name == "" {
		name = "Remote"
	}
	return fmt.Sprintf("Remote (%s): %s", name, s.currentDir)
}

func (s *SFTPFileSource) CurrentPath() string {
	return s.currentDir
}

func (s *SFTPFileSource) SetPath(newPath string) error {
	p := newPath
	if !path.IsAbs(p) {
		p = path.Join(s.currentDir, p)
	}
	clean := path.Clean(p)
	fi, err := s.client.Stat(clean)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return fmt.Errorf("%s is not a directory", clean)
	}
	s.currentDir = clean
	return nil
}

func (s *SFTPFileSource) List() ([]FileEntry, error) {
	entries, err := s.client.ReadDir(s.currentDir)
	if err != nil {
		return nil, err
	}

	res := make([]FileEntry, 0, len(entries))
	for _, fi := range entries {
		res = append(res, FileEntry{
			Name:    fi.Name(),
			Size:    fi.Size(),
			Mode:    fi.Mode(),
			ModTime: fi.ModTime(),
			IsDir:   fi.IsDir(),
		})
	}
	return res, nil
}

func (s *SFTPFileSource) Open(name string) (io.ReadCloser, error) {
	target := path.Join(s.currentDir, name)
	return s.client.Open(target)
}

func (s *SFTPFileSource) Create(name string) (io.WriteCloser, error) {
	target := path.Join(s.currentDir, name)
	return s.client.Create(target)
}

func (s *SFTPFileSource) Stat(name string) (FileEntry, error) {
	target := path.Join(s.currentDir, name)
	fi, err := s.client.Stat(target)
	if err != nil {
		return FileEntry{}, err
	}
	return FileEntry{
		Name:    fi.Name(),
		Size:    fi.Size(),
		Mode:    fi.Mode(),
		ModTime: fi.ModTime(),
		IsDir:   fi.IsDir(),
	}, nil
}

func (s *SFTPFileSource) Mkdir(name string) error {
	target := path.Join(s.currentDir, name)
	return s.client.MkdirAll(target)
}

func (s *SFTPFileSource) Remove(name string) error {
	target := path.Join(s.currentDir, name)
	return s.client.Remove(target)
}

func (s *SFTPFileSource) Close() error {
	var err error
	if s.client != nil {
		err = s.client.Close()
	}
	if s.cmd != nil && s.cmd.Process != nil {
		_ = s.cmd.Process.Kill()
		_ = s.cmd.Wait()
	}
	return err
}

// =============================================================================
// Transfer Helpers
// =============================================================================

// FormatFileSize returns a human-readable representation of a byte size.
func FormatFileSize(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

// TransferFile streams a file from src to dst. If destination file exists and overwrite is false, ErrFileExists is returned.
func TransferFile(src, dst FileSource, entry FileEntry, overwrite bool, progressFn func(copied, total int64)) error {
	if entry.IsDir {
		return transferDirRecursive(src, dst, entry.Name, overwrite, progressFn)
	}

	if !overwrite {
		if _, err := dst.Stat(entry.Name); err == nil {
			return ErrFileExists
		}
	}

	reader, err := src.Open(entry.Name)
	if err != nil {
		return fmt.Errorf("failed to open source file %q: %w", entry.Name, err)
	}
	defer func() { _ = reader.Close() }()

	writer, err := dst.Create(entry.Name)
	if err != nil {
		return fmt.Errorf("failed to create destination file %q: %w", entry.Name, err)
	}
	defer func() { _ = writer.Close() }()

	buf := make([]byte, 32*1024)
	var copied int64
	for {
		nr, rErr := reader.Read(buf)
		if nr > 0 {
			nw, wErr := writer.Write(buf[0:nr])
			if nw < 0 || nr < nw {
				nw = 0
				if wErr == nil {
					wErr = errors.New("invalid write result")
				}
			}
			copied += int64(nw)
			if progressFn != nil {
				progressFn(copied, entry.Size)
			}
			if wErr != nil {
				return wErr
			}
			if nr != nw {
				return io.ErrShortWrite
			}
		}
		if rErr != nil {
			if errors.Is(rErr, io.EOF) {
				break
			}
			return rErr
		}
	}

	return nil
}

func transferDirRecursive(src, dst FileSource, dirName string, overwrite bool, progressFn func(copied, total int64)) error {
	_ = dst.Mkdir(dirName)

	prevSrc := src.CurrentPath()
	prevDst := dst.CurrentPath()
	defer func() {
		_ = src.SetPath(prevSrc)
		_ = dst.SetPath(prevDst)
	}()

	if err := src.SetPath(dirName); err != nil {
		return err
	}
	if err := dst.SetPath(dirName); err != nil {
		return err
	}

	entries, err := src.List()
	if err != nil {
		return err
	}

	for _, e := range entries {
		if e.Name == "." || e.Name == ".." {
			continue
		}
		if err := TransferFile(src, dst, e, overwrite, progressFn); err != nil {
			return err
		}
	}

	return nil
}
