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

| Command | |
|---|---|
| `files ls [path]` | List a folder |
| `files get <remote> [local]` | Download a file |
| `files put <local> [remote]` | Upload a file |
| `files mkdir <path>` | Create a folder and its parents |
| `files rm [-r] <path>` | Delete a file, or a folder with `-r` |

### Deck

| Command | |
|---|---|
| `deck board list` · `deck board <id>` | List boards, show one |
| `deck board create\|update\|delete` | Manage boards |
| `deck stack list --board <id>` | List stacks (columns) |
| `deck stack create\|update\|delete --board <id>` | Manage stacks |
| `deck card list --board <id>` | List cards, with filters and export (below) |
| `deck card <id>` | Show a card |
| `deck card create\|update\|move\|delete` | Manage cards |
| `deck card assign\|unassign <id> <user>` | Assign users (`me` works) |
| `deck card label\|unlabel <id> <label>` | Add or remove labels, by ID or title |
| `deck card comment <id> <text>` · `deck card comments <id>` | Add or list comments |
| `deck attachment list <card>` | List a card's attachments |

Stacks and labels can be given by ID or title. Examples:

```sh
# Filter (comma-separated values are OR'ed) and pick columns
nextcloud deck card list --board 40 --status in-progress,overdue --stack "In progress" \
  --assignee me --column id,title,assignees

# Export the board as JSON and download every attachment
nextcloud deck card list --board 40 --json --download ./backup

# Create a card, then move it and mark it as done
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

| Mode | Tools |
|---|---|
| Always | `deck_list_boards`, `deck_get_board`, `deck_list_cards`, `deck_get_card`, `files_list`, `files_read` |
| `--allow-write` | Create, update and delete boards, stacks, cards and files; move cards; assign users; labels; comments |

## Development

```sh
make build   # ./nextcloud
make fmt     # wsl + golines (80 columns)
make lint    # golangci-lint
```
