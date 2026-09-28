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
	"fmt"

	"github.com/WhiteRoseLK/neossh/internal/i18n"
	"github.com/rivo/tview"
)

func DefaultStatusText() string {
	if i18n.GetLanguage() != i18n.LangEn {
		return i18n.T("statusbar.normal")
	}
	k := CurrentTheme.HintKey
	return fmt.Sprintf(
		"[%s]Tab[-] Panels  • [%s]↑↓[-] Navigate  • [%s]Space[-] Select  • [%s]Enter[-] SSH  • [%s]F[-] SFTP  • [%s]f[-] Forward  • [%s]x[-] Stop Forward  • "+
			"[%s]c[-] Copy SSH  • [%s]o[-] SCP  • [%s]M[-] SSHFS  • [%s]v[-] Paste SSH  • [%s]y[-] Clone  • "+
			"[%s]h[-] Copy Host  • [%s]m[-] Hide  • [%s]H[-] Toggle Hidden  • [%s]a[-] Add  • [%s]e[-] Edit  • "+
			"[%s]g/G/W[-] Ping (All/Watch)  • [%s]K[-] Install Key  • [%s]d[-] Delete  • [%s]p[-] Pin/Unpin  • "+
			"[%s]P[-] Git SSH  • [%s]C[-] Comment  • [%s]X[-] Snippets  • [%s]T[-] Theme  • [%s]i[-] Import  • [%s]/[-] Search  • [%s]q[-] Quit",
		k, k, k, k, k, k, k, k, k, k, k, k, k, k, k, k, k, k, k, k, k, k, k, k, k, k, k, k)
}

func ReadonlyStatusText() string {
	if i18n.GetLanguage() != i18n.LangEn {
		return i18n.T("statusbar.readonly")
	}
	k := CurrentTheme.HintKey
	return fmt.Sprintf(
		"[%s]Tab[-] Panels  • [%s]↑↓[-] Navigate  • [%s]Space[-] Select  • [%s]Enter[-] SSH  • [%s]F[-] SFTP  • [%s]f[-] Forward  • [%s]x[-] Stop Forward  • "+
			"[%s]c[-] Copy SSH  • [%s]o[-] SCP  • [%s]M[-] SSHFS  • [%s]h[-] Copy Host  • [%s]H[-] Toggle Hidden  • "+
			"[%s]g/G/W[-] Ping (All/Watch)  • [%s]p[-] Pin/Unpin  • [%s]P[-] Git SSH  • [%s]X[-] Snippets  • [%s]T[-] Theme  • [%s]/[-] Search  • "+
			"[%s]q[-] Quit  • [red::b][READONLY][-]",
		k, k, k, k, k, k, k, k, k, k, k, k, k, k, k, k, k, k, k)
}

func StatusText(readonly bool) string {
	if readonly {
		return ReadonlyStatusText()
	}
	return DefaultStatusText()
}

func NewStatusBar(readonly ...bool) *tview.TextView {
	ro := false
	if len(readonly) > 0 {
		ro = readonly[0]
	}
	status := tview.NewTextView().SetDynamicColors(true)
	status.SetBackgroundColor(CurrentTheme.StatusBarBackground)
	status.SetTextAlign(tview.AlignCenter)
	status.SetText(StatusText(ro))
	return status
}
