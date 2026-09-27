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

package ports

import "github.com/WhiteRoseLK/neossh/internal/core/domain"

// BundleService handles backup, export, verification, and import of SSH configurations and neossh metadata.
type BundleService interface {
	// Export creates a tar.gz bundle of SSH configs and neossh metadata with optional sanitation.
	Export(opts domain.ExportOptions) (*domain.BundleSummary, error)

	// Verify inspects and validates a bundle archive without extracting files.
	Verify(bundlePath string) (*domain.BundleSummary, error)

	// Import restores configuration files and metadata from a bundle into target directories.
	Import(opts domain.ImportOptions) (*domain.BundleSummary, error)
}
