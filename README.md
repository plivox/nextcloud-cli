# nextcloud-cli

A command-line client for Nextcloud. It currently supports Files and [Deck](https://apps.nextcloud.com/apps/deck), and can run as an [MCP](https://modelcontextprotocol.io) server for AI assistants.

## Installation

```sh
go install github.com/plivox/nextcloud-cli/cmd/nextcloud@latest
```

Or from a clone:

```sh
make build   # produces ./nextcloud
```

## Authentication

### Browser login (recommended)

```sh
nextcloud login cloud.example.com
```

This opens your browser to log in (2FA and SSO work), then stores a dedicated app password in the system keyring (macOS Keychain, Windows Credential Manager, or Secret Service on Linux). If no keyring is available (e.g. a headless Linux server), it falls back to `~/.config/nextcloud-cli/credentials.json`, readable only by you (mode 600). Your main password is never seen by the CLI.

`nextcloud logout` revokes that app password on the server and removes it from this machine.

### Environment variables

For scripts and CI, set these variables in your environment, or in a `.env` file in the current directory (see `.env.example`). They take precedence over the keyring.

```sh
NEXTCLOUD_CLI_BASE_URL=https://cloud.example.com
NEXTCLOUD_CLI_USERNAME=your_username
NEXTCLOUD_CLI_PASSWORD=your_password
```

Using an [app password](https://docs.nextcloud.com/server/latest/user_manual/en/session_management.html#managing-devices) rather than your main password is recommended.

## Usage

### Files

```sh
nextcloud files ls [path]
nextcloud files get <remote> [local]   # to the current directory by default
nextcloud files put <local> [remote]   # to the root folder by default
nextcloud files mkdir <path>
nextcloud files rm <path>              # -r for a folder
```

### Deck

```sh
nextcloud deck board list
nextcloud deck board <id>
nextcloud deck stack list --board <id>
nextcloud deck card <id>
nextcloud deck card list --board <id>
nextcloud deck attachment list <card_id>
```

`card list` can filter and export:

```sh
# Filter by status, stack and assignee (comma-separated values are OR'ed)
nextcloud deck card list --board 40 --status in-progress,overdue --stack "In progress" --assignee me

# Pick the table columns
nextcloud deck card list --board 40 --column id,title,assignees

# JSON output, and download all attachments (plus the JSON) into ./backup
nextcloud deck card list --board 40 --json --download ./backup
```

Creating and updating:

```sh
nextcloud deck board create "Roadmap" [--color ff0000]
nextcloud deck board update <id> [--title ...] [--color ...] [--archived]
nextcloud deck stack create --board <id> "To do"
nextcloud deck card create --board <id> --stack "To do" "Write docs" [--description ...] [--due 2026-11-01]
nextcloud deck card update <id> [--title ...] [--description ...] [--due ...|none] [--done] [--archived]
nextcloud deck card move <id> --stack "Done"
nextcloud deck card assign <id> me
nextcloud deck card label <id> "Bug"
nextcloud deck card comment <id> "Looks good"
nextcloud deck card delete <id>
```

Run `nextcloud <command> --help` for all options.

## MCP server

`nextcloud mcp` runs a [Model Context Protocol](https://modelcontextprotocol.io) server on stdio, using the same credentials as the CLI (run `nextcloud login` first). For example with Claude Code:

```sh
claude mcp add nextcloud -- nextcloud mcp
```

Read-only tools are always available: `deck_list_boards`, `deck_get_board`, `deck_list_cards`, `deck_get_card`, `files_list`, `files_read`.

Tools that create, change or delete boards, stacks, cards, comments and files are only exposed with `--allow-write`:

```sh
claude mcp add nextcloud -- nextcloud mcp --allow-write
```

## Development

```sh
make fmt    # wsl fixes + golines (80 columns)
make lint   # golangci-lint
make vet
```
