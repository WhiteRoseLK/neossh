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

package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/WhiteRoseLK/neossh/internal/adapters/data/ssh_config_file"
	"github.com/WhiteRoseLK/neossh/internal/adapters/security"
	"github.com/WhiteRoseLK/neossh/internal/adapters/ui"
	"github.com/WhiteRoseLK/neossh/internal/core/domain"
	"github.com/WhiteRoseLK/neossh/internal/core/ports"
	"github.com/WhiteRoseLK/neossh/internal/core/services"
	"github.com/WhiteRoseLK/neossh/internal/i18n"
	"github.com/WhiteRoseLK/neossh/internal/logger"
	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	version   = "develop"
	gitCommit = "unknown"

	sshConfigFile     string
	exitOnDisconnect  bool
	sshConfigReadonly bool
	filterQuery       string
	connectDirectly   bool
	importKnownHosts  bool
	knownHostsFile    string
	showHidden        bool
	themeFlag         string
	gitSSHFlag        string
	langFlag          string
	scpFlag           string
	sshfsFlag         string
	tunnelFlag        string
	preConnectFlag    string
	defaultKeyFlag    string
	passwordFlag      string
	pingWatchFlag     bool
	pingIntervalFlag  int

	rootCmd = newRootCmd()
)

type rootOptions struct {
	isReadonly      bool
	filter          string
	isConnect       bool
	isImportKH      bool
	theme           string
	lang            string
	defKey          string
	password        string
	isPingWatch     bool
	isPingWatchSet  bool
	pingIntervalSec int
}

func parseRootOptions(cmd *cobra.Command, args []string) rootOptions {
	isReadonly := sshConfigReadonly
	if ro, err := cmd.Flags().GetBool("readonly"); err == nil && ro {
		isReadonly = true
	}
	if ro, err := cmd.Flags().GetBool("ssh-config-readonly"); err == nil && ro {
		isReadonly = true
	}

	filter := filterQuery
	if len(args) > 0 {
		filter = args[0]
	}
	if f, err := cmd.Flags().GetString("filter"); err == nil && f != "" {
		filter = f
	}

	isConnect := connectDirectly
	if c, err := cmd.Flags().GetBool("connect"); err == nil && c {
		isConnect = true
	}

	isImportKH := importKnownHosts
	if ikh, err := cmd.Flags().GetBool("import-known-hosts"); err == nil && ikh {
		isImportKH = true
	}
	if kh, err := cmd.Flags().GetString("known-hosts"); err == nil && kh != "" {
		knownHostsFile = kh
	}

	theme := themeFlag
	if t, err := cmd.Flags().GetString("theme"); err == nil && t != "" {
		theme = t
	}

	lang := langFlag
	if l, err := cmd.Flags().GetString("lang"); err == nil && l != "" {
		lang = l
	}

	defKey := defaultKeyFlag
	if dk, err := cmd.Flags().GetString("default-key"); err == nil && cmd.Flags().Changed("default-key") {
		defKey = dk
	}

	password := passwordFlag
	if p, err := cmd.Flags().GetString("password"); err == nil && p != "" {
		password = p
	}

	isPingWatch := pingWatchFlag
	isPingWatchSet := cmd.Flags().Changed("ping-watch")
	if pw, err := cmd.Flags().GetBool("ping-watch"); err == nil && pw {
		isPingWatch = true
	}

	pingIntervalSec := pingIntervalFlag
	if pi, err := cmd.Flags().GetInt("ping-interval"); err == nil && pi > 0 {
		pingIntervalSec = pi
	}

	return rootOptions{
		isReadonly:      isReadonly,
		filter:          filter,
		isConnect:       isConnect,
		isImportKH:      isImportKH,
		theme:           theme,
		lang:            lang,
		defKey:          defKey,
		password:        password,
		isPingWatch:     isPingWatch,
		isPingWatchSet:  isPingWatchSet,
		pingIntervalSec: pingIntervalSec,
	}
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   ui.AppName + " [filter]",
		Short: "NeoSSH server picker TUI",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := parseRootOptions(cmd, args)
			if opts.password != "" {
				_ = os.Setenv("NEOSSH_PASSWORD", opts.password)
			}

			log, err := logger.New("NEOSSH")
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			//nolint:errcheck // log.Sync may return an error which is safe to ignore here
			defer log.Sync()

			home, err := os.UserHomeDir()
			if err != nil {
				log.Errorw("failed to get user home directory", "error", err)
				os.Exit(1)
			}

			resolvedConfig, cleanup, err := resolveSSHConfigFile(home, sshConfigFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error resolving config file: %v\n", err)
				os.Exit(1)
			}
			defer cleanup()

			metaDataFile, err := ensureMetadataFile(home)
			if err != nil {
				log.Errorw("failed to setup metadata file", "error", err)
				os.Exit(1)
			}

			serverRepo := ssh_config_file.NewRepository(log, resolvedConfig, metaDataFile)
			credStore := security.NewCredentialStore(log, filepath.Dir(metaDataFile))
			serverService := services.NewServerService(
				log,
				serverRepo,
				services.WithReadOnly(opts.isReadonly),
				services.WithCredentialStore(credStore),
			)
			gitService := services.NewGitService(log)
			gitService.SetServerRepository(serverRepo)

			if handled, err := handleGitSSHFlag(cmd, gitService, opts.isReadonly); handled {
				return err
			}

			if handled, err := handleDefaultKeyFlag(cmd, serverService, opts.isReadonly, opts.defKey); handled {
				return err
			}

			if scpFlag != "" {
				return handleSCPFlag(serverService, scpFlag)
			}

			if sshfsFlag != "" {
				return handleSSHFSFlag(serverService, sshfsFlag)
			}

			if tunnelFlag != "" {
				return handleTunnelFlag(serverService, tunnelFlag)
			}

			if preConnectFlag != "" {
				_ = os.Setenv("NEOSSH_PRE_CONNECT_HOOK", preConnectFlag)
			}

			if opts.isImportKH {
				return handleImportKnownHostsFlag(serverService, opts.isReadonly, home, knownHostsFile)
			}

			if opts.isConnect {
				connected, err := handleDirectConnect(opts.filter, serverService)
				if err != nil {
					return err
				}
				if connected {
					return nil
				}
				// If multiple matches without exact match, launch interactive TUI pre-filtered so user can choose
			}

			tui := ui.NewTUI(log, serverService, version, gitCommit, ui.Config{
				ExitOnDisconnect:   exitOnDisconnect,
				ReadOnly:           opts.isReadonly,
				InitialFilter:      opts.filter,
				ShowHidden:         showHidden,
				Theme:              opts.theme,
				Language:           opts.lang,
				DefaultIdentityKey: opts.defKey,
				AutoPing:           opts.isPingWatch,
				AutoPingSet:        opts.isPingWatchSet,
				AutoPingInterval:   opts.pingIntervalSec,
				ServerRepo:         serverRepo,
				GitService:         gitService,
			})

			return tui.Run()
		},
	}

	cmd.PersistentFlags().StringVar(
		&sshConfigFile, "sshconfig", "", "path to ssh config file (default: ~/.ssh/config)",
	)
	cmd.PersistentFlags().BoolVarP(
		&exitOnDisconnect, "exit-on-disconnect", "x", false, "exit neossh after SSH session finishes",
	)
	cmd.PersistentFlags().BoolVar(
		&exitOnDisconnect, "auto-exit", false, "exit neossh after SSH session finishes",
	)
	cmd.PersistentFlags().BoolVar(
		&sshConfigReadonly, "ssh-config-readonly", false, "run in read-only mode (prevent modifying ~/.ssh/config)",
	)
	cmd.PersistentFlags().BoolVarP(
		&sshConfigReadonly, "readonly", "r", false, "run in read-only mode (alias for --ssh-config-readonly)",
	)
	cmd.PersistentFlags().StringVarP(
		&filterQuery, "filter", "f", "", "pre-filter server list by alias, hostname, or tag",
	)
	cmd.PersistentFlags().BoolVarP(
		&connectDirectly, "connect", "c", false, "connect directly to matching server without launching full TUI picker",
	)
	cmd.PersistentFlags().BoolVar(
		&importKnownHosts, "import-known-hosts", false, "import hosts from ~/.ssh/known_hosts into SSH config",
	)
	cmd.PersistentFlags().StringVar(
		&knownHostsFile, "known-hosts", "", "path to known_hosts file (default: ~/.ssh/known_hosts)",
	)
	cmd.PersistentFlags().StringVar(
		&gitSSHFlag, "git-ssh", "", "configure Git SSH key for current repo (or globally) or view current setting",
	)
	cmd.PersistentFlags().BoolVarP(
		&showHidden, "show-hidden", "H", false, "display hidden servers in UI list",
	)
	cmd.PersistentFlags().StringVarP(
		&themeFlag, "theme", "t", "", "set color theme: dark, light, or system",
	)
	cmd.PersistentFlags().StringVarP(
		&langFlag, "lang", "l", "", "set interface language: en, fr, zh-CN (or via NEOSSH_LANG)",
	)
	cmd.PersistentFlags().StringVar(
		&scpFlag, "scp", "", "generate SCP command templates for server alias (e.g. --scp myserver)",
	)
	cmd.PersistentFlags().StringVar(
		&sshfsFlag, "sshfs", "", "generate SSHFS remote mount command for server alias (e.g. --sshfs myserver)",
	)
	cmd.PersistentFlags().StringVar(
		&tunnelFlag, "tunnel", "", "generate SSH port forwarding / tunnel commands for server alias (e.g. --tunnel myserver)",
	)
	cmd.PersistentFlags().StringVar(
		&tunnelFlag, "forward", "", "generate SSH port forwarding / tunnel commands for server alias (alias for --tunnel)",
	)
	cmd.PersistentFlags().StringVar(
		&preConnectFlag, "pre-connect", "", "run local hook command before SSH connect (supports %h, %p, %r, %n)",
	)
	cmd.PersistentFlags().StringVar(
		&defaultKeyFlag, "default-key", "", "get or set default SSH identity key for new servers",
	)
	cmd.PersistentFlags().StringVarP(
		&passwordFlag, "password", "P", "", "password for automated sshpass authentication",
	)
	cmd.PersistentFlags().BoolVar(
		&pingWatchFlag, "ping-watch", false, "enable periodic background ping watch mode",
	)
	cmd.PersistentFlags().IntVar(
		&pingIntervalFlag, "ping-interval", 0,
		"interval in seconds for periodic background ping watch mode (default: 60)",
	)

	cmd.ValidArgsFunction = func(
		cmd *cobra.Command, args []string, toComplete string,
	) ([]string, cobra.ShellCompDirective) {
		if len(args) > 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return getSSHHostAliasesForCompletion(cmd, toComplete), cobra.ShellCompDirectiveNoFileComp
	}

	_ = cmd.RegisterFlagCompletionFunc("theme", func(
		_ *cobra.Command, _ []string, _ string,
	) ([]string, cobra.ShellCompDirective) {
		return ui.GetThemeNames(), cobra.ShellCompDirectiveNoFileComp
	})
	_ = cmd.RegisterFlagCompletionFunc("lang", func(
		_ *cobra.Command, _ []string, _ string,
	) ([]string, cobra.ShellCompDirective) {
		return i18n.SupportedLanguages(), cobra.ShellCompDirectiveNoFileComp
	})
	_ = cmd.RegisterFlagCompletionFunc("sshconfig", func(
		_ *cobra.Command, _ []string, _ string,
	) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveDefault
	})
	_ = cmd.RegisterFlagCompletionFunc("known-hosts", func(
		_ *cobra.Command, _ []string, _ string,
	) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveDefault
	})
	_ = cmd.RegisterFlagCompletionFunc("filter", func(
		cmd *cobra.Command, _ []string, toComplete string,
	) ([]string, cobra.ShellCompDirective) {
		return getSSHHostAliasesForCompletion(cmd, toComplete), cobra.ShellCompDirectiveNoFileComp
	})
	_ = cmd.RegisterFlagCompletionFunc("scp", func(
		cmd *cobra.Command, _ []string, toComplete string,
	) ([]string, cobra.ShellCompDirective) {
		return getSSHHostAliasesForCompletion(cmd, toComplete), cobra.ShellCompDirectiveNoFileComp
	})
	_ = cmd.RegisterFlagCompletionFunc("sshfs", func(
		cmd *cobra.Command, _ []string, toComplete string,
	) ([]string, cobra.ShellCompDirective) {
		return getSSHHostAliasesForCompletion(cmd, toComplete), cobra.ShellCompDirectiveNoFileComp
	})
	_ = cmd.RegisterFlagCompletionFunc("tunnel", func(
		cmd *cobra.Command, _ []string, toComplete string,
	) ([]string, cobra.ShellCompDirective) {
		return getSSHHostAliasesForCompletion(cmd, toComplete), cobra.ShellCompDirectiveNoFileComp
	})
	_ = cmd.RegisterFlagCompletionFunc("forward", func(
		cmd *cobra.Command, _ []string, toComplete string,
	) ([]string, cobra.ShellCompDirective) {
		return getSSHHostAliasesForCompletion(cmd, toComplete), cobra.ShellCompDirectiveNoFileComp
	})

	cmd.AddCommand(newCompletionCmd())

	cmd.SilenceUsage = true
	return cmd
}

func newCompletionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion script",
		Long: fmt.Sprintf(`To load completions:

Bash:

  $ source <(%[1]s completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ %[1]s completion bash > /etc/bash_completion.d/%[1]s
  # macOS:
  $ %[1]s completion bash > $(brew --prefix)/etc/bash_completion.d/%[1]s

Zsh:

  # If shell completion is not already enabled in your environment,
  # you will need to enable it. You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ %[1]s completion zsh > "${fpath[1]}/_%[1]s"

  # You will need to start a new shell for this setup to take effect.

Fish:

  $ %[1]s completion fish | source

  # To load completions for each session, execute once:
  $ %[1]s completion fish > ~/.config/fish/completions/%[1]s.fish

PowerShell:

  PS> %[1]s completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> %[1]s completion powershell > %[1]s.ps1
  # and source this file from your PowerShell profile.
`, ui.AppName),
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletion(cmd.OutOrStdout())
			case "zsh":
				return cmd.Root().GenZshCompletion(cmd.OutOrStdout())
			case "fish":
				return cmd.Root().GenFishCompletion(cmd.OutOrStdout(), true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletionWithDesc(cmd.OutOrStdout())
			default:
				return fmt.Errorf("unsupported shell type: %s", args[0])
			}
		},
	}
}

func getSSHHostAliasesForCompletion(cmd *cobra.Command, toComplete string) []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}

	customConfig, _ := cmd.Flags().GetString("sshconfig")
	resolvedConfig, cleanup, err := resolveSSHConfigFile(home, customConfig)
	if err != nil {
		return nil
	}
	defer cleanup()

	silentLog := zap.NewNop().Sugar()
	repo := ssh_config_file.NewRepository(silentLog, resolvedConfig, "")
	servers, err := repo.ListServers("")
	if err != nil {
		return nil
	}

	var completions []string
	seen := make(map[string]bool)

	for _, s := range servers {
		if s.IsWildcard || s.IsWildcardServer() {
			continue
		}

		var descParts []string
		if s.User != "" && s.Host != "" {
			descParts = append(descParts, fmt.Sprintf("%s@%s", s.User, s.Host))
		} else if s.Host != "" {
			descParts = append(descParts, s.Host)
		}
		if s.Port != 0 && s.Port != 22 {
			if len(descParts) > 0 {
				descParts[0] = fmt.Sprintf("%s:%d", descParts[0], s.Port)
			} else {
				descParts = append(descParts, fmt.Sprintf(":%d", s.Port))
			}
		}
		desc := strings.Join(descParts, " ")

		allAliases := append([]string{s.Alias}, s.Aliases...)
		for _, alias := range allAliases {
			if alias == "" || seen[alias] {
				continue
			}
			seen[alias] = true

			if toComplete != "" && !strings.HasPrefix(strings.ToLower(alias), strings.ToLower(toComplete)) {
				continue
			}

			if desc != "" {
				completions = append(completions, fmt.Sprintf("%s\t%s", alias, desc))
			} else {
				completions = append(completions, alias)
			}
		}
	}
	return completions
}

func resolveSSHConfigFile(home, customPath string) (string, func(), error) {
	if customPath == "" {
		return filepath.Join(home, ".ssh", "config"), func() {}, nil
	}

	stat, err := os.Stat(customPath)
	if err != nil {
		return "", nil, err
	}

	if stat.Mode()&os.ModeType != 0 {
		f, err := os.CreateTemp("", "tmpfile-")
		if err != nil {
			return "", nil, err
		}

		cleanup := func() {
			_ = f.Close()
			_ = os.Remove(f.Name())
		}

		fd, err := os.Open(customPath) //nolint:gosec // G304: path comes from user flag, intentional
		if err != nil {
			cleanup()
			return "", nil, err
		}
		defer func() {
			_ = fd.Close()
		}()

		content, err := io.ReadAll(fd)
		if err != nil {
			cleanup()
			return "", nil, err
		}
		if _, err := f.WriteString(string(content)); err != nil {
			cleanup()
			return "", nil, err
		}

		return f.Name(), cleanup, nil
	}

	return customPath, func() {}, nil
}

func ensureMetadataFile(home string) (string, error) {
	configDir := filepath.Join(home, ".neossh")
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		configDir = filepath.Join(xdgConfig, "neossh")
	}
	if err := os.MkdirAll(configDir, 0o750); err != nil {
		return "", err
	}
	metaDataFile := filepath.Join(configDir, "metadata.json")

	// Migrate metadata from legacy lazyssh if neossh metadata doesn't exist yet
	if _, err := os.Stat(metaDataFile); os.IsNotExist(err) {
		legacyFile := filepath.Join(home, ".lazyssh", "metadata.json")
		if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
			legacyFile = filepath.Join(xdgConfig, "lazyssh", "metadata.json")
		}
		//nolint:gosec // G304: path constructed from user home directory
		if data, err := os.ReadFile(legacyFile); err == nil {
			_ = os.WriteFile(metaDataFile, data, 0o600)
		}
	}

	return metaDataFile, nil
}

func handleDirectConnect(filter string, serverService ports.ServerService) (bool, error) {
	if strings.TrimSpace(filter) == "" {
		return false, fmt.Errorf("--connect requires a server alias or filter argument (e.g. neossh -c <alias>)")
	}

	matches, err := serverService.ListServers(filter)
	if err != nil {
		return false, fmt.Errorf("failed to query servers: %w", err)
	}

	var targetServer *domain.Server
	for i := range matches {
		if strings.EqualFold(matches[i].Alias, filter) || slices.ContainsFunc(matches[i].Aliases, func(a string) bool {
			return strings.EqualFold(a, filter)
		}) {
			targetServer = &matches[i]
			break
		}
	}
	if targetServer == nil && len(matches) == 1 {
		targetServer = &matches[0]
	}

	if targetServer != nil {
		if targetServer.IsWildcardServer() {
			return false, fmt.Errorf("cannot initiate direct SSH connection to wildcard pattern block '%s'", targetServer.Alias)
		}
		return true, serverService.SSH(targetServer.Alias)
	}

	if len(matches) == 0 {
		return false, fmt.Errorf("no server matching '%s' found", filter)
	}

	return false, nil
}

func handleGitSSHFlag(cmd *cobra.Command, gitService ports.GitService, isReadonly bool) (bool, error) {
	if !cmd.Flags().Changed("git-ssh") {
		return false, nil
	}

	scope := services.ScopeLocal
	cwd, _ := os.Getwd()
	if !gitService.IsGitRepository(cwd) {
		scope = services.ScopeGlobal
	}
	if gitSSHFlag == "" {
		cfg, err := gitService.GetCurrentGitSSHConfig(cwd)
		if err != nil {
			return true, err
		}
		if cfg == "" {
			fmt.Println("No Git SSH key currently configured.")
		} else {
			fmt.Printf("Current Git SSH configuration: %s\n", cfg)
		}
		return true, nil
	}
	if isReadonly {
		return true, fmt.Errorf("cannot configure Git SSH key in read-only mode")
	}
	if err := gitService.ConfigureGitSSHKey(cwd, gitSSHFlag, scope); err != nil {
		return true, err
	}
	fmt.Printf("Successfully configured Git to use SSH key: %s (scope: %s)\n", gitSSHFlag, scope)
	return true, nil
}

func handleImportKnownHostsFlag(
	serverService ports.ServerService, isReadonly bool, home, customPath string,
) error {
	if isReadonly {
		return fmt.Errorf("cannot import known_hosts in read-only mode")
	}
	khPath := customPath
	if khPath == "" {
		khPath = filepath.Join(home, ".ssh", "known_hosts")
	}
	result, err := serverService.ImportKnownHosts(khPath)
	if err != nil {
		return fmt.Errorf("failed to import known_hosts: %w", err)
	}
	if result.Imported == 0 {
		fmt.Printf("No new hosts to import from %s (%d host(s) already configured or skipped).\n",
			khPath, result.Skipped)
	} else {
		fmt.Printf("Successfully imported %d host(s) from %s (%d already configured or skipped).\n",
			result.Imported, khPath, result.Skipped)
	}
	return nil
}

func handleSCPFlag(serverService ports.ServerService, alias string) error {
	servers, err := serverService.ListServers("")
	if err != nil {
		return fmt.Errorf("failed to list servers: %w", err)
	}

	var found *domain.Server
	for i := range servers {
		if strings.EqualFold(servers[i].Alias, alias) {
			found = &servers[i]
			break
		}
	}

	if found == nil {
		return fmt.Errorf("server alias %q not found", alias)
	}

	uploadCmd := ui.BuildSCPUploadCommand(*found, "./<local_file>", "~/<remote_file>", false, false)
	downloadCmd := ui.BuildSCPDownloadCommand(*found, "~/<remote_file>", "./<local_file>", false, false)
	aliasUploadCmd := ui.BuildSCPUploadCommand(*found, "./<local_file>", "~/<remote_file>", false, true)
	aliasDownloadCmd := ui.BuildSCPDownloadCommand(*found, "~/<remote_file>", "./<local_file>", false, true)

	fmt.Printf("SCP Command Templates for [%s]:\n\n", found.Alias)
	fmt.Printf("• Upload (Local -> Remote):\n  %s\n\n", uploadCmd)
	fmt.Printf("• Download (Remote -> Local):\n  %s\n\n", downloadCmd)
	fmt.Printf("• Quick Alias Upload:\n  %s\n\n", aliasUploadCmd)
	fmt.Printf("• Quick Alias Download:\n  %s\n", aliasDownloadCmd)

	if err := clipboard.WriteAll(uploadCmd); err == nil {
		fmt.Println("\n✓ Copied default upload command to system clipboard.")
	}
	return nil
}

func handleSSHFSFlag(serverService ports.ServerService, alias string) error {
	servers, err := serverService.ListServers("")
	if err != nil {
		return fmt.Errorf("failed to list servers: %w", err)
	}

	var found *domain.Server
	for i := range servers {
		if strings.EqualFold(servers[i].Alias, alias) {
			found = &servers[i]
			break
		}
	}

	if found == nil {
		return fmt.Errorf("server alias %q not found", alias)
	}

	mountPoint := fmt.Sprintf("~/mounts/%s", found.Alias)
	mountCmd := ui.BuildSSHFSCommand(*found, "/", mountPoint, false, true, false)
	aliasMountCmd := ui.BuildSSHFSCommand(*found, "/", mountPoint, false, true, true)
	roMountCmd := ui.BuildSSHFSCommand(*found, "/", mountPoint, true, true, false)
	unmountCmd := ui.BuildSSHFSUnmountCommand(mountPoint)

	fmt.Printf("SSHFS Remote Mount Commands for [%s]:\n\n", found.Alias)
	fmt.Printf("• Mount Remote Root (Full Config):\n  %s\n\n", mountCmd)
	fmt.Printf("• Mount via SSH Config Alias:\n  %s\n\n", aliasMountCmd)
	fmt.Printf("• Mount Read-Only:\n  %s\n\n", roMountCmd)
	fmt.Printf("• Unmount Remote Filesystem:\n  %s\n", unmountCmd)

	if err := clipboard.WriteAll(mountCmd); err == nil {
		fmt.Println("\n✓ Copied default mount command to system clipboard.")
	}
	return nil
}

func handleTunnelFlag(serverService ports.ServerService, alias string) error {
	servers, err := serverService.ListServers("")
	if err != nil {
		return fmt.Errorf("failed to list servers: %w", err)
	}

	var found *domain.Server
	for i := range servers {
		if strings.EqualFold(servers[i].Alias, alias) {
			found = &servers[i]
			break
		}
	}

	if found == nil {
		return fmt.Errorf("server alias %q not found", alias)
	}

	localCmd := ui.BuildForwardCommand(*found, ui.ForwardTypeLocal, "5432", "localhost", "5432", "", true, true)
	remoteCmd := ui.BuildForwardCommand(*found, ui.ForwardTypeRemote, "8080", "localhost", "80", "", true, true)
	dynCmd := ui.BuildForwardCommand(*found, ui.ForwardTypeDynamic, "1080", "", "", "", true, true)

	fmt.Printf("SSH Port Forwarding & Tunnel Commands for [%s]:\n\n", found.Alias)
	fmt.Printf("• Local Forwarding (-L local_port:target_host:target_port):\n  %s\n\n", localCmd)
	fmt.Printf("• Remote Forwarding (-R remote_port:target_host:target_port):\n  %s\n\n", remoteCmd)
	fmt.Printf("• Dynamic SOCKS5 Proxy (-D local_port):\n  %s\n\n", dynCmd)

	settingsMgr := ui.NewDefaultSettingsManager()
	if settingsMgr != nil {
		profiles, _ := settingsMgr.LoadTunnelProfiles(found.Alias)
		if len(profiles) > 0 {
			fmt.Printf("Saved Favorite Profiles for [%s]:\n", found.Alias)
			for _, p := range profiles {
				onlyFw := p.Mode != ui.ForwardModeForwardSSH
				cmdStr := ui.BuildForwardCommand(*found, p.Type, p.Port, p.Host, p.HostPort, p.BindAddress, onlyFw, true)
				fmt.Printf("• %s (%s):\n  %s\n", p.Name, p.Type, cmdStr)
			}
			fmt.Println()
		}
	}

	if err := clipboard.WriteAll(localCmd); err == nil {
		fmt.Println("✓ Copied default Local forward command to system clipboard.")
	}
	return nil
}

func handleDefaultKeyFlag(
	cmd *cobra.Command, serverService ports.ServerService, isReadonly bool, key string,
) (bool, error) {
	if !cmd.Flags().Changed("default-key") {
		return false, nil
	}

	trimmed := strings.TrimSpace(key)
	if trimmed == "" {
		current, err := serverService.GetDefaultIdentityKey()
		if err != nil {
			return true, fmt.Errorf("failed to get default identity key: %w", err)
		}
		if current == "" {
			fmt.Println("No default SSH identity key configured.")
		} else {
			fmt.Printf("Current default SSH identity key: %s\n", current)
		}
		return true, nil
	}

	if isReadonly {
		return true, fmt.Errorf("cannot configure default identity key in read-only mode")
	}

	if err := serverService.SaveDefaultIdentityKey(trimmed); err != nil {
		return true, fmt.Errorf("failed to save default identity key: %w", err)
	}

	fmt.Printf("Successfully set default SSH identity key: %s\n", trimmed)
	return true, nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
