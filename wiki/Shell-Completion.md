#  Shell Autocompletion

`neossh` includes rich shell autocompletion for **Bash**, **Zsh**, **Fish**, and **PowerShell**, complete with dynamic server alias suggestions directly parsed from your SSH configuration.

---

## Capabilities

- **Dynamic Server Alias Completion**: Press <kbd>tab</kbd> after `neossh`, `neossh -c`, `neossh --scp`, `neossh --sshfs`, `neossh --sftp`, or `neossh --tunnel` to view matching host aliases along with their `user@host:port` descriptions.
- **Flag & Option Completion**: Flags like `--theme` suggest available themes (`dark`, `light`, `system`), `--lang` suggests supported locales (`en`, `fr`, `zh-CN`), and `--sshconfig` / `--known-hosts` trigger path completions.

---

## Installation Instructions

### Bash

    **Current Session:**
    ```bash
    source <(neossh completion bash)
    ```

    **Permanent Installation (Linux):**
    ```bash
    neossh completion bash | sudo tee /etc/bash_completion.d/neossh > /dev/null
    ```

    **Permanent Installation (macOS via Homebrew):**
    ```bash
    neossh completion bash > $(brew --prefix)/etc/bash_completion.d/neossh
    ```

### Zsh

    Ensure completion is enabled in your `~/.zshrc`:
    ```bash
    echo "autoload -U compinit; compinit" >> ~/.zshrc
    ```

    Generate and save completions to your Zsh completion directory:
    ```bash
    # Create directory if necessary
    mkdir -p ~/.zsh/completion
    echo 'fpath=(~/.zsh/completion $fpath)' >> ~/.zshrc

    neossh completion zsh > ~/.zsh/completion/_neossh
    exec zsh
    ```

    Or install into Homebrew site-functions (macOS):
    ```bash
    neossh completion zsh > $(brew --prefix)/share/zsh/site-functions/_neossh
    ```

### Fish

    **Current Session:**
    ```fish
    neossh completion fish | source
    ```

    **Permanent Installation:**
    ```fish
    mkdir -p ~/.config/fish/completions
    neossh completion fish > ~/.config/fish/completions/neossh.fish
    ```

### PowerShell

    **Current Session:**
    ```powershell
    neossh completion powershell | Out-String | Invoke-Expression
    ```

    **Permanent Installation:**
    ```powershell
    # Save completion script
    neossh completion powershell > "$HOME\Documents\PowerShell\neossh.ps1"

    # Add to your PowerShell profile:
    Add-Content $PROFILE "`n. `"$HOME\Documents\PowerShell\neossh.ps1`""
    ```

---

## Verifying Completion

After restarting your shell or sourcing your profile, type:

```bash
neossh -c <TAB>
```

Your shell should present the list of available server aliases parsed from your `~/.ssh/config`.
