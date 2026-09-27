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
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func TestLocalFileSource_NavigationAndCRUD(t *testing.T) {
	tmpDir := t.TempDir()

	// Create initial test directory structure
	subDir := filepath.Join(tmpDir, "subfolder")
	if err := os.Mkdir(subDir, 0o755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	testFile := filepath.Join(tmpDir, "hello.txt")
	if err := os.WriteFile(testFile, []byte("world"), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	source, err := NewLocalFileSource(tmpDir)
	if err != nil {
		t.Fatalf("failed to create LocalFileSource: %v", err)
	}

	// Test List()
	entries, err := source.List()
	if err != nil {
		t.Fatalf("failed to list entries: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}

	// Test Stat()
	entry, err := source.Stat("hello.txt")
	if err != nil {
		t.Fatalf("failed to stat hello.txt: %v", err)
	}
	if entry.Size != 5 || entry.IsDir {
		t.Errorf("unexpected stat for hello.txt: size=%d, isDir=%v", entry.Size, entry.IsDir)
	}

	// Test Open()
	rc, err := source.Open("hello.txt")
	if err != nil {
		t.Fatalf("failed to open hello.txt: %v", err)
	}
	data, err := io.ReadAll(rc)
	_ = rc.Close()
	if err != nil || string(data) != "world" {
		t.Fatalf("read mismatch: %s (err: %v)", string(data), err)
	}

	// Test SetPath() to child directory
	if err := source.SetPath("subfolder"); err != nil {
		t.Fatalf("failed to set path to subfolder: %v", err)
	}
	if filepath.Base(source.CurrentPath()) != "subfolder" {
		t.Errorf("expected current path to be subfolder, got %s", source.CurrentPath())
	}

	// Test Create() in child directory
	wc, err := source.Create("created.txt")
	if err != nil {
		t.Fatalf("failed to create file in child dir: %v", err)
	}
	_, _ = wc.Write([]byte("new content"))
	_ = wc.Close()

	// Test SetPath("..")
	if err := source.SetPath(".."); err != nil {
		t.Fatalf("failed to navigate up: %v", err)
	}
	if source.CurrentPath() != tmpDir {
		t.Errorf("expected current path %s, got %s", tmpDir, source.CurrentPath())
	}

	// Test Remove()
	if err := source.Remove("hello.txt"); err != nil {
		t.Fatalf("failed to remove hello.txt: %v", err)
	}
	if _, err := source.Stat("hello.txt"); err == nil {
		t.Errorf("expected error after removing hello.txt, got nil")
	}
}

func TestTransferFile_LocalToLocal(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()

	srcFile := filepath.Join(srcDir, "data.bin")
	content := []byte("streaming data transfer test content 12345")
	if err := os.WriteFile(srcFile, content, 0o644); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	src, _ := NewLocalFileSource(srcDir)
	dst, _ := NewLocalFileSource(dstDir)

	entry, err := src.Stat("data.bin")
	if err != nil {
		t.Fatalf("failed to stat source file: %v", err)
	}

	var progressCalls int
	progressFn := func(copied, total int64) {
		progressCalls++
		if total != int64(len(content)) {
			t.Errorf("expected total=%d, got %d", len(content), total)
		}
	}

	// 1. Initial transfer without existing file
	if err := TransferFile(src, dst, entry, false, progressFn); err != nil {
		t.Fatalf("TransferFile failed: %v", err)
	}
	if progressCalls == 0 {
		t.Errorf("expected progressFn to be called at least once")
	}

	// Verify destination file content
	dstContent, err := os.ReadFile(filepath.Join(dstDir, "data.bin"))
	if err != nil || !bytes.Equal(dstContent, content) {
		t.Fatalf("content mismatch in destination: %s", string(dstContent))
	}

	// 2. Transfer again without overwrite -> expect ErrFileExists
	err = TransferFile(src, dst, entry, false, nil)
	if !errors.Is(err, ErrFileExists) {
		t.Fatalf("expected ErrFileExists, got %v", err)
	}

	// 3. Transfer again with overwrite=true -> should succeed
	err = TransferFile(src, dst, entry, true, nil)
	if err != nil {
		t.Fatalf("expected transfer with overwrite to succeed, got %v", err)
	}

	// 4. Test directory recursive transfer
	folder := filepath.Join(srcDir, "myfolder")
	_ = os.Mkdir(folder, 0o755)
	_ = os.WriteFile(filepath.Join(folder, "nested.txt"), []byte("nested"), 0o644)

	dirEntry, err := src.Stat("myfolder")
	if err != nil {
		t.Fatalf("failed to stat dir: %v", err)
	}

	if err := TransferFile(src, dst, dirEntry, false, nil); err != nil {
		t.Fatalf("failed to recursively transfer dir: %v", err)
	}

	nestedContent, err := os.ReadFile(filepath.Join(dstDir, "myfolder", "nested.txt"))
	if err != nil || string(nestedContent) != "nested" {
		t.Fatalf("nested content mismatch: %s", string(nestedContent))
	}
}

func TestFormatFileSize(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{500, "500 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, tt := range tests {
		got := FormatFileSize(tt.bytes)
		if got != tt.expected {
			t.Errorf("FormatFileSize(%d) = %q, want %q", tt.bytes, got, tt.expected)
		}
	}
}

func TestSFTPManager_UI_NavigationAndSorting(t *testing.T) {
	app := tview.NewApplication()
	leftDir := t.TempDir()
	rightDir := t.TempDir()

	// Populate left dir
	_ = os.WriteFile(filepath.Join(leftDir, "b_file.txt"), []byte("hello"), 0o644)
	_ = os.WriteFile(filepath.Join(leftDir, "a_file.txt"), []byte("big large content here"), 0o644)
	_ = os.Mkdir(filepath.Join(leftDir, "z_folder"), 0o755)

	leftSrc, _ := NewLocalFileSource(leftDir)
	rightSrc, _ := NewLocalFileSource(rightDir)

	mgr := NewSFTPManager(app, leftSrc, rightSrc)

	closed := false
	mgr.OnClose(func() {
		closed = true
	})

	// Check initial active pane
	if mgr.activePane != 0 {
		t.Errorf("expected initial active pane 0, got %d", mgr.activePane)
	}

	// 1. Test Tab -> switches to right pane
	tabEvent := tcell.NewEventKey(tcell.KeyTab, 0, tcell.ModNone)
	mgr.leftTable.InputHandler()(tabEvent, nil)
	if mgr.activePane != 1 {
		t.Errorf("expected active pane 1 after Tab, got %d", mgr.activePane)
	}

	// Tab again -> switches back to left pane
	mgr.rightTable.InputHandler()(tabEvent, nil)
	if mgr.activePane != 0 {
		t.Errorf("expected active pane 0 after second Tab, got %d", mgr.activePane)
	}

	// 2. Test sorting cycle via 's'
	if mgr.sortMode != SortFilesByName {
		t.Errorf("expected SortFilesByName initially")
	}
	sEvent := tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone)
	mgr.leftTable.InputHandler()(sEvent, nil)
	if mgr.sortMode != SortFilesBySize {
		t.Errorf("expected SortFilesBySize after first 's', got %d", mgr.sortMode)
	}
	mgr.leftTable.InputHandler()(sEvent, nil)
	if mgr.sortMode != SortFilesByDate {
		t.Errorf("expected SortFilesByDate after second 's', got %d", mgr.sortMode)
	}

	// 3. Test Enter into directory
	// Row 0 is header, Row 1 is "..", Row 2 should be directory "z_folder"
	mgr.leftTable.Select(2, 0)
	enterEvent := tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone)
	mgr.leftTable.InputHandler()(enterEvent, nil)
	if filepath.Base(leftSrc.CurrentPath()) != "z_folder" {
		t.Errorf("expected path to be z_folder, got %s", leftSrc.CurrentPath())
	}

	// 4. Test Backspace -> navigate back up
	bsEvent := tcell.NewEventKey(tcell.KeyBackspace, 0, tcell.ModNone)
	mgr.leftTable.InputHandler()(bsEvent, nil)
	if leftSrc.CurrentPath() != leftDir {
		t.Errorf("expected path %s after backspace, got %s", leftDir, leftSrc.CurrentPath())
	}

	// 5. Test Upload 'u' (Left -> Right)
	// Select "a_file.txt"
	mgr.leftTable.Select(3, 0)
	uEvent := tcell.NewEventKey(tcell.KeyRune, 'u', tcell.ModNone)
	mgr.leftTable.InputHandler()(uEvent, nil)

	// Wait for transfer goroutine to complete
	time.Sleep(100 * time.Millisecond)

	// Verify file landed in right dir
	dstFiles, _ := rightSrc.List()
	if len(dstFiles) == 0 {
		t.Errorf("expected file to be uploaded to right dir")
	}

	// 6. Test Close 'q'
	qEvent := tcell.NewEventKey(tcell.KeyRune, 'q', tcell.ModNone)
	mgr.leftTable.InputHandler()(qEvent, nil)
	if !closed {
		t.Errorf("expected onClose to be called when pressing 'q'")
	}
}
