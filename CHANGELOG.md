# Changelog

All notable changes to this project will be documented in this file.
This project adheres to [Semantic Versioning](https://semver.org/) and is automated with [Release Please](https://github.com/googleapis/release-please).

## [2.3.0](https://github.com/WhiteRoseLK/neossh/compare/v2.2.0...v2.3.0) (2026-09-30)


### Features

* 1-click known_hosts conflict resolution and manager ([#135](https://github.com/WhiteRoseLK/neossh/issues/135)) ([535e860](https://github.com/WhiteRoseLK/neossh/commit/535e8607b171999a66a2b3ef9cd386182032174d))
* add command snippets library and quick runner modal ([#139](https://github.com/WhiteRoseLK/neossh/issues/139)) ([71f0a91](https://github.com/WhiteRoseLK/neossh/commit/71f0a91434216a2fade48e361e1ca9a727e6e84f))
* **cli:** first-launch onboarding wizard and --setup flag for companion tools ([#147](https://github.com/WhiteRoseLK/neossh/issues/147)) ([c6d56ef](https://github.com/WhiteRoseLK/neossh/commit/c6d56ef8c22f250960ad320c8d350d856edbb821))
* **core:** dotfiles synchronization to remote servers via chezmoi ([#146](https://github.com/WhiteRoseLK/neossh/issues/146)) ([05c07b2](https://github.com/WhiteRoseLK/neossh/commit/05c07b2c53a20eda2197f8741148b01bbb618032))
* live ssh-agent integration and status monitoring ([#137](https://github.com/WhiteRoseLK/neossh/issues/137)) ([786a42c](https://github.com/WhiteRoseLK/neossh/commit/786a42c8f23a36a3b4e4d5ecd77ef93df07fb71f))
* multi-server selection and bulk operations ([#138](https://github.com/WhiteRoseLK/neossh/issues/138)) ([38cfa96](https://github.com/WhiteRoseLK/neossh/commit/38cfa969b8b5d77fed39b4250ac693a0f6e72d2f))
* multi-session engine with overview dashboard and tabbed views ([#141](https://github.com/WhiteRoseLK/neossh/issues/141)) ([e5c78c2](https://github.com/WhiteRoseLK/neossh/commit/e5c78c2493a0f38a5f76f6bad2a93d1cc1f5526c))


### Documentation

* document multi-server selection and bulk operations ([aad529c](https://github.com/WhiteRoseLK/neossh/commit/aad529cdf80a3beeba3cbf2ca25e0e1611ea1f37))
* fix table formatting and add companion guides to wiki ([#148](https://github.com/WhiteRoseLK/neossh/issues/148)) ([e0c4cd7](https://github.com/WhiteRoseLK/neossh/commit/e0c4cd7bdfd998abdc3bd4a2239cb1b72dbcdf69))


### Maintenance & Dependencies

* disable blank issues to enforce structured templates ([#142](https://github.com/WhiteRoseLK/neossh/issues/142)) ([65b927a](https://github.com/WhiteRoseLK/neossh/commit/65b927ae1a5a6ceb973d3697d63d0517d74e136c))

## [2.2.0](https://github.com/WhiteRoseLK/neossh/compare/v2.1.0...v2.2.0) (2026-09-28)


### Features

* enhance ssh error diagnostics, remote disconnections, and custom keybindings ([#129](https://github.com/WhiteRoseLK/neossh/issues/129)) ([3625c69](https://github.com/WhiteRoseLK/neossh/commit/3625c699badecdab3b5a3560009f9bba70447811))


### Documentation

* add mkdocs documentation site and community standards ([#121](https://github.com/WhiteRoseLK/neossh/issues/121)) ([78f8739](https://github.com/WhiteRoseLK/neossh/commit/78f8739746f9448b3e5f32d1da95ac43d4376120))
* correct gollum link syntax to DisplayText|PageName ([#126](https://github.com/WhiteRoseLK/neossh/issues/126)) ([2fcef3c](https://github.com/WhiteRoseLK/neossh/commit/2fcef3c22900b1e594080e6984f67239d95ea2c6))
* decouple documentation from github pages and whiterose domain ([#123](https://github.com/WhiteRoseLK/neossh/issues/123)) ([13db18d](https://github.com/WhiteRoseLK/neossh/commit/13db18d8ba272565416dda06abeb5aa66ab890d0))
* fix malformed tags and formatting across all wiki pages ([#127](https://github.com/WhiteRoseLK/neossh/issues/127)) ([75f7d1f](https://github.com/WhiteRoseLK/neossh/commit/75f7d1f55a2568f8d1b54080f3f5d40a21026624))
* migrate documentation to github wiki ([#124](https://github.com/WhiteRoseLK/neossh/issues/124)) ([478a2ad](https://github.com/WhiteRoseLK/neossh/commit/478a2adeec95092d0893c92156643acc69c4dbfe))
* remove photo of original creator in readme ([#128](https://github.com/WhiteRoseLK/neossh/issues/128)) ([eadc2fc](https://github.com/WhiteRoseLK/neossh/commit/eadc2fca05b8895271e14f8cb474dc318956fa17))
* simplify readme to focus on essentials and link to wiki ([#125](https://github.com/WhiteRoseLK/neossh/issues/125)) ([89900bd](https://github.com/WhiteRoseLK/neossh/commit/89900bded431b25c0964530b680eed30599a1350))


### Continuous Integration & Tooling

* **deps:** bump actions/github-script from 7 to 9 ([f6c6d1d](https://github.com/WhiteRoseLK/neossh/commit/f6c6d1dddcc93a46d6bfb56e1486b6422cd8de66))
* **deps:** bump actions/setup-go from 5 to 7 ([c9689cc](https://github.com/WhiteRoseLK/neossh/commit/c9689cccd2da8a93498d9ea617639c814a98264b))
* **deps:** bump goreleaser/goreleaser-action from 5 to 7 ([#118](https://github.com/WhiteRoseLK/neossh/issues/118)) ([3fdd605](https://github.com/WhiteRoseLK/neossh/commit/3fdd605d5329f4de4102de6f4f69cd2a3c8f821c))

## [2.1.0](https://github.com/WhiteRoseLK/neossh/compare/v2.0.0...v2.1.0) (2026-09-27)


### Features

* **cli:** backup and export SSH configuration and metadata bundle ([#116](https://github.com/WhiteRoseLK/neossh/issues/116)) ([7bc5ce4](https://github.com/WhiteRoseLK/neossh/commit/7bc5ce42b3b36a3e607476fbc10dfef08a2693ba))
* **cli:** provide shell autocompletion for bash, zsh, and fish with host alias completion ([#105](https://github.com/WhiteRoseLK/neossh/issues/105)) ([d0cebf0](https://github.com/WhiteRoseLK/neossh/commit/d0cebf005f4f89044c51a40ad518b657c5347854))
* **cli:** quick launch SFTP session or external file transfer for selected server ([#117](https://github.com/WhiteRoseLK/neossh/issues/117)) ([0755734](https://github.com/WhiteRoseLK/neossh/commit/0755734a990cbe10c097d679843954745671a539))
* **core:** SSH port forwarding and tunnel assistant (local, remote, and SOCKS5) ([#112](https://github.com/WhiteRoseLK/neossh/issues/112)) ([6bfa79e](https://github.com/WhiteRoseLK/neossh/commit/6bfa79eed0766e2bfcbb32cc6d0424739bb50cb7))
* renew SSH certificates on demand via an external command ([#115](https://github.com/WhiteRoseLK/neossh/issues/115)) ([e355ff8](https://github.com/WhiteRoseLK/neossh/commit/e355ff8586ab04d75508b6e8d431b6831cdeccc0))
* show SSH certificate status ([#114](https://github.com/WhiteRoseLK/neossh/issues/114)) ([e5c60b9](https://github.com/WhiteRoseLK/neossh/commit/e5c60b940a25ee22ea9e7591b916b425df5a0cc7))
* **ui:** built-in dual-pane SFTP file manager (WinSCP / FileZilla style) ([#113](https://github.com/WhiteRoseLK/neossh/issues/113)) ([c32c37b](https://github.com/WhiteRoseLK/neossh/commit/c32c37bccb8bd817edb35069720ade800141f1a2))
* **ui:** display SSH key type badges and FIDO2/security-key indicator in server details ([#107](https://github.com/WhiteRoseLK/neossh/issues/107)) ([25cd763](https://github.com/WhiteRoseLK/neossh/commit/25cd763fca82bb588bc88a20597f34029888e8ea))
* **ui:** periodic background ping watch mode with configurable interval ([#106](https://github.com/WhiteRoseLK/neossh/issues/106)) ([1c4ff24](https://github.com/WhiteRoseLK/neossh/commit/1c4ff24cbb732e74f6aafcf35363c4a8847c368e))
* **ui:** support advanced search filters in search bar ([#104](https://github.com/WhiteRoseLK/neossh/issues/104)) ([7e6b667](https://github.com/WhiteRoseLK/neossh/commit/7e6b667121aa0c8174709d77fb3239b413381e3a))


### Bug Fixes

* **release:** standardize release workflow on Google Release Please ([#91](https://github.com/WhiteRoseLK/neossh/issues/91)) ([bac6974](https://github.com/WhiteRoseLK/neossh/commit/bac69748681d0f1fba3f7e4e66c923b7fa867406))
* **ui:** preserve case-sensitive shortcut keys in normalizeGlobalHotkey ([#108](https://github.com/WhiteRoseLK/neossh/issues/108)) ([a8699f9](https://github.com/WhiteRoseLK/neossh/commit/a8699f96f0f2a71b62a5d9514c33c9abad33344e))


### Continuous Integration & Tooling

* **build:** sync needs-triage label with issue priority ([#103](https://github.com/WhiteRoseLK/neossh/issues/103)) ([9b02df8](https://github.com/WhiteRoseLK/neossh/commit/9b02df8869b83e6ff4d4d6dac53e74067b95e87c))
* **release:** add distribution targets check and dry-run on release PRs ([#93](https://github.com/WhiteRoseLK/neossh/issues/93)) ([51cd6ba](https://github.com/WhiteRoseLK/neossh/commit/51cd6baae662c08d422318dd72dd51bf98301c5a))

## [2.0.0](https://github.com/WhiteRoseLK/neossh/compare/v1.0.0...v2.0.0) (2026-09-22)


### ⚠ BREAKING CHANGES

* **core:** preload and manage server state in-memory ([#90](https://github.com/WhiteRoseLK/neossh/issues/90))

### Features

* **cli:** add --ssh-config-readonly flag to prevent modifying config ([#52](https://github.com/WhiteRoseLK/neossh/issues/52)) ([06fdce1](https://github.com/WhiteRoseLK/neossh/commit/06fdce1654f4ecd247b38d6c28d73d68d9e5e8be))
* **cli:** add option to exit after an SSH session finishes ([#50](https://github.com/WhiteRoseLK/neossh/issues/50)) ([50e48be](https://github.com/WhiteRoseLK/neossh/commit/50e48be2093b3137217afe99069043773bc9afca))
* **cli:** allow filtering or connecting directly via command line switch ([#56](https://github.com/WhiteRoseLK/neossh/issues/56)) ([04ad838](https://github.com/WhiteRoseLK/neossh/commit/04ad838e6e5c88353dc8f8153d8c9cc4dd998d60))
* **cli:** scp command generator and clipboard helper ([#79](https://github.com/WhiteRoseLK/neossh/issues/79)) ([4bee8f9](https://github.com/WhiteRoseLK/neossh/commit/4bee8f9b95fe036b7a6d30c3cc11845851d63581))
* **config:** add Git SSH key configuration and profile switcher ([#64](https://github.com/WhiteRoseLK/neossh/issues/64)) ([5f55122](https://github.com/WhiteRoseLK/neossh/commit/5f55122c9dca6414ca2aa1e920777e61e3dcd19b))
* **config:** configurable default identity SSH key for new servers ([#83](https://github.com/WhiteRoseLK/neossh/issues/83)) ([8232760](https://github.com/WhiteRoseLK/neossh/commit/823276005038724be1a96b6af8436b8694e07d99))
* **config:** import hosts from ~/.ssh/known_hosts ([#58](https://github.com/WhiteRoseLK/neossh/issues/58)) ([b431f28](https://github.com/WhiteRoseLK/neossh/commit/b431f28ea84c4cbfacdbd45fdfda468129f7c00f))
* **config:** save server tags directly as SSH config comments ([#57](https://github.com/WhiteRoseLK/neossh/issues/57)) ([b98da74](https://github.com/WhiteRoseLK/neossh/commit/b98da747262efcdf3a3fc7543163a3875f58ea5a))
* **config:** support multiple aliases on a single Host line ([#51](https://github.com/WhiteRoseLK/neossh/issues/51)) ([5da3b08](https://github.com/WhiteRoseLK/neossh/commit/5da3b08e49938d44ddcbc60748a91228a071c82f))
* **config:** support wildcard pattern blocks ([#53](https://github.com/WhiteRoseLK/neossh/issues/53)) ([c49b412](https://github.com/WhiteRoseLK/neossh/commit/c49b41295349b24290f81ceb015a5e64118edc2d))
* **core:** integrate native OS keyring and encrypted vault for credentials ([#88](https://github.com/WhiteRoseLK/neossh/issues/88)) ([c39fc5e](https://github.com/WhiteRoseLK/neossh/commit/c39fc5e5036136b8f0d9a45a71af765eb42bb8c6))
* **core:** pre-connect command hooks and pre-SSH script execution ([#80](https://github.com/WhiteRoseLK/neossh/issues/80)) ([5a6d739](https://github.com/WhiteRoseLK/neossh/commit/5a6d739cf5edfd02042d9f922d4268189f50e793))
* **core:** sshfs remote mount integration ([#86](https://github.com/WhiteRoseLK/neossh/issues/86)) ([7b9f1c9](https://github.com/WhiteRoseLK/neossh/commit/7b9f1c911e19df0eea42ba0a421ed34e64e143aa))
* **core:** support password authentication via sshpass or credential store ([#87](https://github.com/WhiteRoseLK/neossh/issues/87)) ([7403734](https://github.com/WhiteRoseLK/neossh/commit/7403734f01f240825625ff920bcc9fcc1144cafc))
* **release:** provide Windows Scoop package manifest ([#85](https://github.com/WhiteRoseLK/neossh/issues/85)) ([8910d27](https://github.com/WhiteRoseLK/neossh/commit/8910d27b6b658992d7e72554c37191f75362720e))
* **ui:** active background SSH sessions panel and process controls ([#62](https://github.com/WhiteRoseLK/neossh/issues/62)) ([205365b](https://github.com/WhiteRoseLK/neossh/commit/205365b48aa583fb4388a47ef74c855dcbba34f7))
* **ui:** add ability to duplicate / clone existing server entry ([#40](https://github.com/WhiteRoseLK/neossh/issues/40)) ([a8ce2a1](https://github.com/WhiteRoseLK/neossh/commit/a8ce2a1c452e5ec04c151544368e3a62a338fca9))
* **ui:** add shortcut K to install SSH key via ssh-copy-id ([f68484e](https://github.com/WhiteRoseLK/neossh/commit/f68484ea3307c81fb16476a4bbd284c3e73abc9d))
* **ui:** allow hiding specific hosts from the UI list ([#60](https://github.com/WhiteRoseLK/neossh/issues/60)) ([e40c51d](https://github.com/WhiteRoseLK/neossh/commit/e40c51d3499d61605b061da068499d305a00424f))
* **ui:** automatically set terminal and tab title on SSH connection ([#78](https://github.com/WhiteRoseLK/neossh/issues/78)) ([25381ed](https://github.com/WhiteRoseLK/neossh/commit/25381edb6c346210c1243b748f70dbca13fb4456))
* **ui:** display SSH error messages in UI when connection fails ([#47](https://github.com/WhiteRoseLK/neossh/issues/47)) ([62c34fb](https://github.com/WhiteRoseLK/neossh/commit/62c34fbabff4ebcde738d9ed3721b0f9b9714eb9))
* **ui:** implement Ping All with visual status indicators ([#36](https://github.com/WhiteRoseLK/neossh/issues/36)) ([2c5fa4d](https://github.com/WhiteRoseLK/neossh/commit/2c5fa4df2d7cd566122eaedc8fda50c1ca342b2f))
* **ui:** improve keyboard navigation, focus borders, and confirmation dialogs ([#59](https://github.com/WhiteRoseLK/neossh/issues/59)) ([f4a1923](https://github.com/WhiteRoseLK/neossh/commit/f4a192303accad882835365abc8bcb45dab29a8b))
* **ui:** internationalization framework and localization support ([#65](https://github.com/WhiteRoseLK/neossh/issues/65)) ([28b9fac](https://github.com/WhiteRoseLK/neossh/commit/28b9face6670675fd1008c7328a02f25069a94b7))
* **ui:** paste SSH command from clipboard to create a server entry ([#39](https://github.com/WhiteRoseLK/neossh/issues/39)) ([1fe4be5](https://github.com/WhiteRoseLK/neossh/commit/1fe4be546ce565619831f4a814b3b1154af169ef))
* **ui:** server folders, nested grouping, and tmux session integration ([#63](https://github.com/WhiteRoseLK/neossh/issues/63)) ([b683e1e](https://github.com/WhiteRoseLK/neossh/commit/b683e1eb3ee980d4077a077478a13dab7bb6a65d))
* **ui:** support Dark, Light, and System color themes with runtime toggle ([#61](https://github.com/WhiteRoseLK/neossh/issues/61)) ([5b49771](https://github.com/WhiteRoseLK/neossh/commit/5b49771cbfc7643ede88e005a17b30787872f091))


### Bug Fixes

* **cli:** format defer closures to satisfy gofmt and errcheck ([b7b75b6](https://github.com/WhiteRoseLK/neossh/commit/b7b75b6e82c40f8eb518372557dcfb31402d5826))
* **config:** eliminate extra blank lines during edit operations ([#42](https://github.com/WhiteRoseLK/neossh/issues/42)) ([2945a77](https://github.com/WhiteRoseLK/neossh/commit/2945a77bc77252e4d6d8438d8dd420c6fac63614))
* **config:** skip directory matches in Include globs ([#82](https://github.com/WhiteRoseLK/neossh/issues/82)) ([da90951](https://github.com/WhiteRoseLK/neossh/commit/da9095156a67b89dc35305455f2588958d2b25c3))
* **config:** use relative tilde paths for IdentityFile and UserKnownHostsFile ([#44](https://github.com/WhiteRoseLK/neossh/issues/44)) ([885667f](https://github.com/WhiteRoseLK/neossh/commit/885667f1af0e5fab3400ed7709bd7f5dc0a18b8b))
* **core:** handle remote disconnects cleanly and prevent terminal hangs ([#41](https://github.com/WhiteRoseLK/neossh/issues/41)) ([9fb3516](https://github.com/WhiteRoseLK/neossh/commit/9fb3516910adf55a8912fd7f5bd271537373e970))
* **parser:** strip surrounding double quotes from Host alias entries ([#76](https://github.com/WhiteRoseLK/neossh/issues/76)) ([64f9b1e](https://github.com/WhiteRoseLK/neossh/commit/64f9b1e4b90fd1a4acf79a8789f4ed3f45a6ca12))
* **ui:** allow usernames starting with a digit in server form validation ([#77](https://github.com/WhiteRoseLK/neossh/issues/77)) ([54edffc](https://github.com/WhiteRoseLK/neossh/commit/54edffc362826c4b24d97bb4a413c805ac2d5148))


### Documentation

* add Arch Linux AUR installation instructions ([#84](https://github.com/WhiteRoseLK/neossh/issues/84)) ([40b01a9](https://github.com/WhiteRoseLK/neossh/commit/40b01a994394293a129c589364a1b954e6b71726))
* enforce mandatory README updates in agent guidelines and PR template ([#55](https://github.com/WhiteRoseLK/neossh/issues/55)) ([97be399](https://github.com/WhiteRoseLK/neossh/commit/97be399d212feda70e4251b11fd291700908acf3))
* fix README table formatting and update screenshots with neossh branding ([#81](https://github.com/WhiteRoseLK/neossh/issues/81)) ([8dd0120](https://github.com/WhiteRoseLK/neossh/commit/8dd0120c471a40b46bfe991964ce23d93182a75d))
* highlight new features and bug fixes vs lazyssh in README ([#38](https://github.com/WhiteRoseLK/neossh/issues/38)) ([28be199](https://github.com/WhiteRoseLK/neossh/commit/28be19981e399283fff985643e878615deae5da0))
* standardize issue templates and fix GitHub issue forms ([#45](https://github.com/WhiteRoseLK/neossh/issues/45)) ([98be4b5](https://github.com/WhiteRoseLK/neossh/commit/98be4b54bb5712480ed61926094a19a0330889fe))
* update README with CLI options, new features, and keybindings ([#54](https://github.com/WhiteRoseLK/neossh/issues/54)) ([b7a9a4b](https://github.com/WhiteRoseLK/neossh/commit/b7a9a4bb5c270c2a917b83f1192431d835e6b7f1))


### Continuous Integration & Tooling

* allow main scope in semantic-prs for release-please PRs ([5ed1864](https://github.com/WhiteRoseLK/neossh/commit/5ed1864f8f1952b371849c8c55c0acc601ff75ca))
* auto-label new issues with needs-triage ([#48](https://github.com/WhiteRoseLK/neossh/issues/48)) ([705a960](https://github.com/WhiteRoseLK/neossh/commit/705a9603600c602b48ec0180fcec7e117390b252))
* configure changelog-sections in release-please-config.json ([35c8b04](https://github.com/WhiteRoseLK/neossh/commit/35c8b04605650707b5db3157f3ded8ffd715b624))
* configure Google Release Please, CI workflows, and developer guidelines ([ce285f6](https://github.com/WhiteRoseLK/neossh/commit/ce285f64c47f881b8b30a419c0cb3a5dc799d832))
* **deps:** bump actions/checkout from 4 to 7 ([#24](https://github.com/WhiteRoseLK/neossh/issues/24)) ([54acde0](https://github.com/WhiteRoseLK/neossh/commit/54acde084abfdc3c2999d11eb66972dd94c88ed0))
* **deps:** bump actions/upload-artifact from 4 to 7 ([#21](https://github.com/WhiteRoseLK/neossh/issues/21)) ([b422f35](https://github.com/WhiteRoseLK/neossh/commit/b422f351ca7c9558985ec493248385e9608c422f))
* **deps:** bump amannn/action-semantic-pull-request from 5.5.3 to 6.1.1 ([#22](https://github.com/WhiteRoseLK/neossh/issues/22)) ([2fbe517](https://github.com/WhiteRoseLK/neossh/commit/2fbe51791e26cc79098c095250e6e71c68fd11f0))
* **deps:** bump googleapis/release-please-action from 4 to 5 ([#27](https://github.com/WhiteRoseLK/neossh/issues/27)) ([da2521d](https://github.com/WhiteRoseLK/neossh/commit/da2521dce0a1d09052ac54bbd7383ef66662ba0f))
* migrate to golangci-lint v2 and golangci-lint-action v9 ([#37](https://github.com/WhiteRoseLK/neossh/issues/37)) ([8a01b03](https://github.com/WhiteRoseLK/neossh/commit/8a01b03d0539cd632de3776f347597653cf10f1a))
* **release:** use RELEASE_PLEASE_TOKEN to trigger PR checks automatically ([#43](https://github.com/WhiteRoseLK/neossh/issues/43)) ([6c8ca5d](https://github.com/WhiteRoseLK/neossh/commit/6c8ca5d01412c090990951a747f0b502211420b2))


### Maintenance & Dependencies

* **deps:** bump github.com/gdamore/tcell/v2 from 2.9.0 to 2.13.10 ([#30](https://github.com/WhiteRoseLK/neossh/issues/30)) ([de24bbb](https://github.com/WhiteRoseLK/neossh/commit/de24bbba4adc80957127cf358de1ece25e011d0d))
* **deps:** bump github.com/mattn/go-runewidth from 0.0.16 to 0.0.30 ([#25](https://github.com/WhiteRoseLK/neossh/issues/25)) ([a06a3e5](https://github.com/WhiteRoseLK/neossh/commit/a06a3e559be2669333ad8b3fda63a7923cb5695c))
* **deps:** bump github.com/rivo/tview from 0.0.0-20250625164341-a4a78f1e05cb to 0.42.0 ([#26](https://github.com/WhiteRoseLK/neossh/issues/26)) ([1a00ec6](https://github.com/WhiteRoseLK/neossh/commit/1a00ec65ff4eed86fbae43573b33dc2003d8a840))
* **deps:** bump github.com/spf13/cobra from 1.9.1 to 1.10.2 ([#28](https://github.com/WhiteRoseLK/neossh/issues/28)) ([e5a8c5c](https://github.com/WhiteRoseLK/neossh/commit/e5a8c5c69dadc9bca35ac7c0ccb5b30bdead0085))
* **deps:** bump go.uber.org/zap from 1.27.0 to 1.28.0 ([#29](https://github.com/WhiteRoseLK/neossh/issues/29)) ([2a0fa10](https://github.com/WhiteRoseLK/neossh/commit/2a0fa109b56bc54a19a435064305f514a90fe682))


### Code Refactoring

* **core:** preload and manage server state in-memory ([#90](https://github.com/WhiteRoseLK/neossh/issues/90)) ([64118ec](https://github.com/WhiteRoseLK/neossh/commit/64118ec132656db5d5f516d788562766ba11fbfb))

## [1.0.0](https://github.com/WhiteRoseLK/neossh/releases/tag/v1.0.0) (2026-09-20)

### Features
* Initial release of **neossh**, an actively maintained fork and continuation of lazyssh.
* TUI header branding updated to `neossh`.
* Migration path supported from legacy `~/.lazyssh` to `~/.neossh`.
* Pre-compiled multi-platform binaries (Linux, macOS, Windows).
* Homebrew installation available via `whiteroselk/tap/neossh`.
