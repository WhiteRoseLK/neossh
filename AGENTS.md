# Agent Guidelines for neossh

This file provides rules and conventions for AI assistants (Antigravity, Claude Code, Copilot, Cursor, etc.) and developers working in this repository. `CLAUDE.md` only imports this file: edit the rules here.

## 🚨 Mandatory Documentation Rule: NEVER FORGET THE README

Whenever you implement a new feature (`feat`), CLI flag, configuration option, UI shortcut/keybinding, or behavior change:
1. **Always update `README.md` in the exact same branch and Pull Request**.
2. Specific sections to keep in sync:
   - **`## ⚡ What's New & Fixed vs. lazyssh?`**: Add a concise row to the table or summary.
   - **`## 💻 Command Line Usage`**: If a new CLI flag is added or modified, document its flag, shorthand, description, and default value.
   - **`## ⌨️ Keybindings`**: If a keybinding or shortcut is added/changed/disabled, update the keybinding table and notes.
   - **`### Credits & Acknowledgments`**: Add the author of the issue or feature to the community contributor list if applicable.

## 🧪 Testing & Code Quality

1. **Zero Lint Tolerance**: All changes must satisfy `golangci-lint run` with 0 warnings or errors. Format all Go files with `gofmt -w <file>`.
2. **Race Detection**: Always run `go test -v -race ./...` before committing to verify there are no data races.
3. **Conventional Commits**: Every commit and PR title MUST strictly adhere to Conventional Commits:
   - Allowed scopes: `ui`, `cli`, `config`, `parser`, `core`, `deps`, `build`, `release`, `main`.
   - Examples: `feat(cli): ...`, `fix(config): ...`, `docs: ...`, `ci: ...`.

## 🌿 PR & Workflow Process

1. Work on dedicated feature/bugfix branches.
2. Link the issue in the PR description (`Closes #<id>`).
3. Squash and merge with `gh pr merge <id> --squash --delete-branch`, only once CI is green. Never use `--admin`: it bypasses the required checks.
4. Never merge the Release Please PR (`chore(main): release …`): releasing is the maintainer's decision.
