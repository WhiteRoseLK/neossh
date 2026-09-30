# Installation

`neossh` is available on macOS, Linux, and Windows through package managers, pre-compiled binaries, Go, or building from source.

---

## Option 1: Homebrew (macOS & Linux) — Official Tap

Install `neossh` using the official Homebrew tap:

```bash
brew tap WhiteRoseLK/tap
brew trust WhiteRoseLK/tap
brew install neossh
```

*(If you previously had `lazyssh` installed, Homebrew will seamlessly prompt to replace it while preserving your server configs and metadata).*

> [!NOTE]
> **Why an external tap instead of `brew install neossh` directly?**  
> Homebrew Core requires new packages to meet a community adoption threshold (typically 50–75 GitHub stars) before being accepted into the central registry.
> 
> Once the project meets Homebrew's notoriety criteria, we will submit a formula to `homebrew/core`.  
> ⭐ **[Star the repository](https://github.com/WhiteRoseLK/neossh)** to help us reach the threshold for Homebrew Core inclusion!

---

## Option 2: Arch Linux (AUR)

`neossh` is available in the Arch User Repository ([AUR/neossh](https://aur.archlinux.org/packages/neossh)):

```bash
# Using yay:
yay -S neossh

# Using paru:
paru -S neossh
```

---

## Option 3: Pre-compiled Binaries (Direct Download)

Ready-to-run binaries are available for **macOS**, **Linux**, and **Windows** on the [Releases page](https://github.com/WhiteRoseLK/neossh/releases/latest).

### macOS & Linux:

```bash
# Automatically downloads and extracts the latest binary for your OS and architecture:
OS="$(uname -s)"
ARCH="$(uname -m)"
[ "$ARCH" = "x86_64" ] && ARCH="x86_64" || ARCH="arm64"

curl -sL "https://github.com/WhiteRoseLK/neossh/releases/latest/download/neossh_${OS}_${ARCH}.tar.gz" | tar -xz

# Move binary to PATH:
sudo mv neossh /usr/local/bin/
```

### Windows:

**Via Scoop (Recommended):**
```powershell
# Install directly via repository manifest:
scoop install https://raw.githubusercontent.com/WhiteRoseLK/neossh/main/scoop/neossh.json

# Or add the official scoop bucket:
scoop bucket add neossh https://github.com/WhiteRoseLK/scoop-bucket
scoop install neossh
```

**Manual Zip Download:**
1. Download the `.zip` archive from the [Releases page](https://github.com/WhiteRoseLK/neossh/releases/latest).
2. Extract `neossh.exe` to a folder in your `PATH` (e.g. `C:\Windows\System32` or a dedicated tools directory).

---

## Option 4: Go Install

If you have Go installed:

```bash
go install github.com/WhiteRoseLK/neossh/cmd@latest
```

*(Make sure `$GOPATH/bin` or `~/go/bin` is in your `$PATH`).*

---

## Option 5: Build from Source

**Prerequisites**: [Go](https://go.dev/) 1.22+ and `git` (and optionally `make`).

```bash
# 1. Clone the repository
git clone https://github.com/WhiteRoseLK/neossh.git
cd neossh

# 2. Build using Make
make build

# 3. Install the binary into your PATH
sudo cp bin/neossh /usr/local/bin/
```

*Or build manually without Make:*

```bash
go build -ldflags "-X main.version=v1.0.0" -o neossh ./cmd/main.go
sudo mv neossh /usr/local/bin/
```

---

## Next Steps

- Run `neossh` or `neossh --setup` to launch the interactive companion tools setup wizard (`chezmoi`, `yazi`, `ssh-copy-id`).
- Check out the **[[Quick-Start]]** guide to get productive in 2 minutes!
