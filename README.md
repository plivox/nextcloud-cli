# nextcloud-cli

A command-line client for Nextcloud. It currently supports Files and [Deck](https://apps.nextcloud.com/apps/deck) (read-only).

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

Run `nextcloud <command> --help` for all options.

## Development

```sh
make fmt    # wsl fixes + golines (80 columns)
make lint   # golangci-lint
make vet
```
