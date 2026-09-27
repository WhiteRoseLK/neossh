---
title: Installation
---

# :material-download: Installation

## Homebrew (macOS & Linux)

Install neossh using the official Homebrew tap:

```bash
brew tap WhiteRoseLK/tap
brew trust WhiteRoseLK/tap
brew install neossh
```

!!! note "Why an external tap?"
    Homebrew Core requires new packages to meet a community adoption threshold (typically 50–75 GitHub stars) before being accepted into the central registry. Once neossh reaches this threshold, it will be submitted to `homebrew/core`.

    :star: **[Star the repository](https://github.com/WhiteRoseLK/neossh)** to help reach the threshold!

If you previously had `lazyssh` installed, Homebrew will seamlessly prompt to replace it while preserving your server configs and metadata.

---

## Arch Linux (AUR)

neossh is available in the Arch User Repository ([AUR/neossh](https://aur.archlinux.org/packages/neossh)):

```bash
# Using yay:
yay -S neossh

# Using paru:
paru -S neossh
```

---

## Pre-compiled Binaries

Ready-to-run binaries are available for **macOS**, **Linux**, and **Windows** on the [Releases page](https://github.com/WhiteRoseLK/neossh/releases/latest).

=== "macOS & Linux"

    ```bash
    # Automatically downloads and extracts the latest binary:
    OS="$(uname -s)"
    ARCH="$(uname -m)"
    [ "$ARCH" = "x86_64" ] && ARCH="x86_64" || ARCH="arm64"

    curl -sL "https://github.com/WhiteRoseLK/neossh/releases/latest/download/neossh_${OS}_${ARCH}.tar.gz" | tar -xz

    # Move binary to PATH:
    sudo mv neossh /usr/local/bin/
    ```

=== "Windows (Scoop)"

    ```powershell
    # Install directly via repository manifest:
    scoop install https://raw.githubusercontent.com/WhiteRoseLK/neossh/main/scoop/neossh.json

    # Or add the official scoop bucket:
    scoop bucket add neossh https://github.com/WhiteRoseLK/scoop-bucket
    scoop install neossh
    ```

=== "Windows (Manual)"

    1. Download the `.zip` archive from the [Releases page](https://github.com/WhiteRoseLK/neossh/releases/latest).
    2. Extract `neossh.exe` to a folder in your `PATH` (e.g. `C:\Windows\System32` or a dedicated tools directory).

---

## Go Install

If you have Go installed:

```bash
go install github.com/WhiteRoseLK/neossh/cmd@latest
```

!!! tip
    Make sure `$GOPATH/bin` or `~/go/bin` is in your `$PATH`.

---

## Build from Source

**Prerequisites**: [Go](https://go.dev/) 1.22+ and `git` (and optionally `make`).

=== "With Make"

    ```bash
    git clone https://github.com/WhiteRoseLK/neossh.git
    cd neossh
    make build
    sudo cp bin/neossh /usr/local/bin/
    ```

=== "Manual"

    ```bash
    git clone https://github.com/WhiteRoseLK/neossh.git
    cd neossh
    go build -ldflags "-X main.version=v1.0.0" -o neossh ./cmd/main.go
    sudo mv neossh /usr/local/bin/
    ```

---

## Next Steps

Once installed, head to the [Quick Start](quickstart.md) guide to get up and running in minutes.
