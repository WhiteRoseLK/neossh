#  Internationalization (i18n)

`neossh` features full runtime internationalization, allowing you to use the terminal interface in your native language.

---

## Supported Languages

| Code | Language | Native Name |
|:---:|---|---|
| `en` | English | English (Default) |
| `fr` | French | Français |
| `zh-CN` | Simplified Chinese | 简体中文 |

---

## Configuring the Language

You can choose your interface language through multiple convenient mechanisms:

### CLI Flag

Set the language at startup using `--lang` or `-l`:

```bash
# Launch in French
neossh --lang fr

# Launch in Simplified Chinese
neossh -l zh-CN
```

### Environment Variable

Configure a default language in your shell profile (`~/.zshrc`, `~/.bashrc`, etc.):

```bash
export NEOSSH_LANG=fr
```

### Shell Autocompletion

Shell completions dynamically provide supported language codes:

```bash
neossh --lang <TAB>
# suggestions: en  fr  zh-CN
```

---

## Translation Fallback

`neossh` employs a resilient key-based fallback strategy. If a phrase is missing in a specific language catalog, it seamlessly falls back to standard English without crashing or displaying raw translation placeholders.

---

## Contributing New Languages

Translations are maintained as structured JSON resource bundles in `internal/adapters/ui/i18n/locales/`.

To contribute a new language or improve existing translations:

1. Fork the repository: `https://github.com/WhiteRoseLK/neossh`
2. Create or update `internal/adapters/ui/i18n/locales/<lang_code>.json` using `en.json` as a template.
3. Open a pull request following our [Contribution Guidelines](https://github.com/WhiteRoseLK/neossh/blob/main/CONTRIBUTING.md).
