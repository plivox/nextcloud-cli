# nextcloud-cli

[![CI](https://github.com/plivox/nextcloud-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/plivox/nextcloud-cli/actions/workflows/ci.yml)

A command-line client for Nextcloud **Files** and **[Deck](https://apps.nextcloud.com/apps/deck)**, which also runs as an **[MCP](https://modelcontextprotocol.io) server** for AI assistants.

## Installation

**Homebrew** (macOS and Linux):

```sh
brew install plivox/tap/nextcloud-cli
```

**Install script** (macOS and Linux, see [`install.sh`](install.sh)):

```sh
curl -fsSL https://raw.githubusercontent.com/plivox/nextcloud-cli/main/install.sh | sh
```

**Windows**: download `nextcloud-cli_windows_amd64.zip` from the [releases](https://github.com/plivox/nextcloud-cli/releases/latest) and put `nextcloud.exe` in your `PATH`.

**With Go** (1.27+):

```sh
go install github.com/plivox/nextcloud-cli/cmd/nextcloud@latest
```

> On macOS, a binary downloaded with a browser is blocked by Gatekeeper because it isn't signed. Run `xattr -d com.apple.quarantine nextcloud` once, or use the install script above.

## Authentication

**Browser login (recommended).** `nextcloud login <server>` opens your browser (2FA and SSO work) and stores a dedicated app password in the system keyring, or in `~/.config/nextcloud-cli/credentials.json` (mode 600) when no keyring is available. `nextcloud logout` revokes it.

**Environment variables.** For scripts and CI, set them in the environment or in a `.env` file (see [`.env.example`](.env.example)). They take precedence over the keyring.

```sh
NEXTCLOUD_CLI_BASE_URL=https://cloud.example.com
NEXTCLOUD_CLI_USERNAME=alice
NEXTCLOUD_CLI_PASSWORD=app-password
```

## Usage

### Files

| Command                      | Description                          |
| ---------------------------- | ------------------------------------ |
| `files ls [path]`            | List a folder                        |
| `files get <remote> [local]` | Download a file                      |
| `files put <local> [remote]` | Upload a file                        |
| `files mkdir <path>`         | Create a folder and its parents      |
| `files rm [-r] <path>`       | Delete a file, or a folder with `-r` |

### Deck

| Command                                                     | Description                                 |
| ----------------------------------------------------------- | ------------------------------------------- |
| `deck board list` · `deck board <id>`                       | List boards, show one                       |
| `deck board create\|update\|delete`                         | Manage boards                               |
| `deck stack list --board <id>`                              | List stacks (columns)                       |
| `deck stack create\|update\|delete --board <id>`            | Manage stacks                               |
| `deck card list --board <id>`                               | List cards, with filters and export (below) |
| `deck card <id>`                                            | Show a card                                 |
| `deck card create\|update\|move\|delete`                    | Manage cards                                |
| `deck card assign\|unassign <id> <user>`                    | Assign users (`me` works)                   |
| `deck card label\|unlabel <id> <label>`                     | Add or remove labels, by ID or title        |
| `deck card comment <id> <text>` · `deck card comments <id>` | Add or list comments                        |
| `deck attachment list <card>`                               | List a card's attachments                   |

Stacks and labels can be given by ID or title.

**Filter cards and pick the columns** (comma-separated values are OR'ed):

```sh
nextcloud deck card list --board 40 \
  --status in-progress,overdue \
  --stack "In progress" \
  --assignee me \
  --column id,title,assignees
```

**Export a board** as JSON, with every attachment:

```sh
nextcloud deck card list --board 40 --json --download ./backup
```

**Create a card**, then move it and mark it as done:

```sh
nextcloud deck card create --board 40 --stack "To do" "Write docs" --due 2026-11-01
nextcloud deck card move 123 --stack "Done"
nextcloud deck card update 123 --done
```

Run `nextcloud <command> --help` for all options.

### MCP server

`nextcloud mcp` serves the same features to AI assistants over stdio, with the CLI's credentials. With Claude Code:

```sh
claude mcp add nextcloud -- nextcloud mcp                # read-only
claude mcp add nextcloud -- nextcloud mcp --allow-write  # also create, change and delete
```

|       | Default (read-only)         | With `--allow-write`                                 |
| ----- | --------------------------- | ---------------------------------------------------- |
| Deck  | `deck_list_*`, `deck_get_*` | Create, update, delete, move, assign, label, comment |
| Files | `files_list`, `files_read`  | `files_write`, `files_mkdir`, `files_delete`         |

## Development

Requires Go 1.27+.

```sh
git clone https://github.com/plivox/nextcloud-cli && cd nextcloud-cli
make run ARGS="deck board list"   # run from source
```

| Command      | Description                                          |
| ------------ | ---------------------------------------------------- |
| `make build` | Build `./nextcloud`                                  |
| `make test`  | Run the tests                                        |
| `make lint`  | Run golangci-lint (with wsl)                         |
| `make fmt`   | Fix blank lines (wsl) and wrap lines at 80 (golines) |
| `make vet`   | Run `go vet`                                         |

CI runs lint, tests and builds for Linux, macOS and Windows on every push and pull request.

Code layout: `cmd/nextcloud` (entry point), `internal/cli` (commands), `internal/mcp` (MCP server), `internal/api` (HTTP client, with `deck` and `files`), `internal/auth` (credentials and login).

### Releasing

1. Make sure `main` is green in [CI](https://github.com/plivox/nextcloud-cli/actions/workflows/ci.yml).
2. Tag the version ([semver](https://semver.org)) and push the tag:

   ```sh
   git tag -a v0.2.0 -m "v0.2.0"
   git push origin v0.2.0
   ```

3. The [Release workflow](https://github.com/plivox/nextcloud-cli/actions/workflows/release.yml) runs [GoReleaser](https://goreleaser.com), which:
   - builds the binaries for Linux, macOS and Windows (amd64, arm64);
   - publishes the GitHub release with archives, checksums and a changelog from the commits;
   - updates the `nextcloud-cli` cask in [plivox/homebrew-tap](https://github.com/plivox/homebrew-tap).

To try the release locally without publishing: `go run github.com/goreleaser/goreleaser/v2@latest release --snapshot --clean` (output in `dist/`).

The cask is pushed with the `HOMEBREW_TAP_GITHUB_TOKEN` secret, a fine-grained token with *Contents: read and write* on `plivox/homebrew-tap`. Renew it before it expires, or the release fails at the Homebrew step.
