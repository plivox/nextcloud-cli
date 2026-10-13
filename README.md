# nextcloud-cli

A command-line client for Nextcloud. It currently supports [Deck](https://apps.nextcloud.com/apps/deck) (read-only).

## Installation

```sh
go install github.com/plivox/nextcloud-cli/cmd/nextcloud@latest
```

Or from a clone:

```sh
make build   # produces ./nextcloud
```

## Configuration

Set these variables in your environment, or in a `.env` file in the current directory (see `.env.example`):

```sh
NEXTCLOUD_BASE_URL=https://cloud.example.com
NEXTCLOUD_USERNAME=your_username
NEXTCLOUD_PASSWORD=your_password
```

Using an [app password](https://docs.nextcloud.com/server/latest/user_manual/en/session_management.html#managing-devices) rather than your main password is recommended.

## Usage

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
