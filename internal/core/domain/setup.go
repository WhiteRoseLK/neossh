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

// CompanionTool represents an external companion tool that enhances neossh capabilities.
type CompanionTool struct {
	ID          string // Unique identifier: "chezmoi", "yazi", "ssh-copy-id"
	Name        string // Display name
	Description string // Feature description
	Shortcut    string // Associated shortcut: "D", "F", "1-click"
	BinaryName  string // Binary checked in PATH
	Installed   bool   // Whether currently available in PATH
	Selected    bool   // Whether marked for installation
}

// PackageManagerType represents supported system package managers.
type PackageManagerType string

const (
	PackageManagerBrew   PackageManagerType = "brew"
	PackageManagerApt    PackageManagerType = "apt"
	PackageManagerDnf    PackageManagerType = "dnf"
	PackageManagerPacman PackageManagerType = "pacman"
	PackageManagerZypper PackageManagerType = "zypper"
	PackageManagerApk    PackageManagerType = "apk"
	PackageManagerWinget PackageManagerType = "winget"
	PackageManagerScoop  PackageManagerType = "scoop"
	PackageManagerChoco  PackageManagerType = "choco"
	PackageManagerNone   PackageManagerType = ""
)

// PackageManagerInfo contains details about the detected host package manager.
type PackageManagerInfo struct {
	Type        PackageManagerType
	Executable  string
	DisplayName string
}
