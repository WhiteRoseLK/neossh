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
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestBuildCustomKeyMap(t *testing.T) {
	raw := map[string]string{
		"add_server":   "n",
		"clone_server": "c",
		"copy_command": "y",
		"quit":         "x",
		"x":            "q", // direct rune to rune
	}

	keyMap := buildCustomKeyMap(raw)

	// 'n' and 'N' should map to 'a' (add)
	if got := keyMap['n']; got != 'a' {
		t.Errorf("keyMap['n'] = %q, want 'a'", got)
	}
	if got := keyMap['N']; got != 'a' {
		t.Errorf("keyMap['N'] = %q, want 'a'", got)
	}

	// 'c' should map to 'y' (clone)
	if got := keyMap['c']; got != 'y' {
		t.Errorf("keyMap['c'] = %q, want 'y'", got)
	}

	// 'y' should map to 'c' (copy)
	if got := keyMap['y']; got != 'c' {
		t.Errorf("keyMap['y'] = %q, want 'c'", got)
	}

	// 'x' should map to 'q' (quit)
	if got := keyMap['x']; got != 'q' {
		t.Errorf("keyMap['x'] = %q, want 'q'", got)
	}
}

func TestResolveCommandKey(t *testing.T) {
	ui := &tui{
		customKeybindings: map[rune]rune{
			'n': 'a',
			'N': 'a',
			'c': 'y',
		},
	}

	// Test overridden key 'n' -> 'a'
	evN := tcell.NewEventKey(tcell.KeyRune, 'n', tcell.ModNone)
	if got := ui.resolveCommandKey(evN); got != 'a' {
		t.Errorf("resolveCommandKey('n') = %q, want 'a'", got)
	}

	// Test default key 'e' -> 'e' (untouched)
	evE := tcell.NewEventKey(tcell.KeyRune, 'e', tcell.ModNone)
	if got := ui.resolveCommandKey(evE); got != 'e' {
		t.Errorf("resolveCommandKey('e') = %q, want 'e'", got)
	}

	// Test default key with capital normalization 'Q' -> 'q'
	evQ := tcell.NewEventKey(tcell.KeyRune, 'Q', tcell.ModNone)
	if got := ui.resolveCommandKey(evQ); got != 'q' {
		t.Errorf("resolveCommandKey('Q') = %q, want 'q'", got)
	}
}
