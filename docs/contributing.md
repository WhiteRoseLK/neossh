---
title: Contributing
---

# :material-handshake: Contributing

Contributions are welcome! Feel free to open an [Issue](https://github.com/WhiteRoseLK/neossh/issues) or submit a Pull Request.

## Development Setup

### Prerequisites

- [Go](https://go.dev/) 1.22+
- `git`
- `make` (optional)

### Build

```bash
git clone https://github.com/WhiteRoseLK/neossh.git
cd neossh
make build
```

### Run Tests

```bash
go test -v -race ./...
```

### Lint

```bash
golangci-lint run
```

## Commit Convention

This project uses [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add new feature
fix: resolve bug
docs: update documentation
chore: maintenance task
refactor: code restructuring
test: add or update tests
```

## Credits & Acknowledgments

- **[Adembc](https://github.com/Adembc)**: Original author and creator of [lazyssh](https://github.com/Adembc/lazyssh). Without his architectural work, neossh would not exist.
- **Community contributors**: Full credit to all contributors from the upstream repository whose ideas and pull requests made this project possible.

## License

This project is licensed under the [Apache-2.0 License](https://github.com/WhiteRoseLK/neossh/blob/main/LICENSE).
