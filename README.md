# dotkey-cli

CLI tool for [dotkey](https://github.com/muhmunt/dotkey) — pull, push, and manage environment variables from your terminal.

[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](../LICENSE)
[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](go.mod)

## Installation

```bash
# macOS / Linux
curl -fsSL https://raw.githubusercontent.com/muhmunt/dotkey-cli/main/install.sh | sh

# With Go
go install github.com/muhmunt/dotkey-cli@latest
```

## Quick start

```bash
dotkey login
dotkey project use my-service
dotkey env use development
dotkey pull
```

## Commands

| Command | Description |
|---------|-------------|
| `dotkey login` | Authenticate via browser |
| `dotkey pull [env]` | Pull secrets → write `.env` |
| `dotkey push [env]` | Push local `.env` → server |
| `dotkey add KEY=VALUE` | Add or update a variable |
| `dotkey remove KEY` | Delete a variable |
| `dotkey diff dev prod` | Compare two environments |
| `dotkey history [env]` | Show change history |
| `dotkey rollback ID` | Restore a previous value |

See [full command reference](https://muhmunt.github.io/dotkey/cli/commands/) in the docs.

## Config

Stored at `~/.dotkey/config.yaml`. Set via `dotkey login` or manually:

```yaml
api_url: http://localhost:8080
web_url: http://localhost:3000
token: your-jwt-token
```

## CI usage

```bash
DOTKEY_TOKEN=${{ secrets.DOTKEY_TOKEN }} dotkey pull production
```

## License

[MIT](https://github.com/muhmunt/dotkey/blob/main/LICENSE)
