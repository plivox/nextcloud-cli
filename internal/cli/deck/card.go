package deckcmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/plivox/nextcloud-cli/internal/api/deck"

	"github.com/rodaine/table"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

var cardViewBoardFlag int
var cardListBoardFlag int
var cardListJSONFlag bool
var cardListAttachmentsFlag bool
var cardListDownloadFlag string
var cardListColumnsFlag []string
var cardListStatusFlag []string
var cardListStackFlag []string
var cardListAssigneeFlag []string

var defaultCardListColumns = []string{
	"id",
	"title",
	"stack",
	"status",
	"attachments",
}

type cardColumn struct {
	header string
	value  func(s deck.Stack, c deck.Card, withAttachments bool) interface{}
}

var cardColumns = map[string]cardColumn{
	"id": {
		"ID",
		func(s deck.Stack, c deck.Card, _ bool) interface{} { return c.ID },
	},
	"title": {
		"Title",
		func(s deck.Stack, c deck.Card, _ bool) interface{} { return c.Title },
	},
	"stack": {
		"Stack",
		func(s deck.Stack, c deck.Card, _ bool) interface{} { return s.Title },
	},
	"status": {
		"Status",
		func(s deck.Stack, c deck.Card, _ bool) interface{} { return c.Status() },
	},
	"assignees": {
		"Assignee(s)",
		func(s deck.Stack, c deck.Card, _ bool) interface{} { return assigneesOf(c) },
	},
	"attachments": {
		"Attachments",
		func(s deck.Stack, c deck.Card, withAttachments bool) interface{} {
			return attachmentsSummary(c, withAttachments)
		},
	},
}

func assigneesOf(c deck.Card) string {
	if len(c.AssignedUsers) == 0 {
		return ""
	}

	names := make([]string, len(c.AssignedUsers))
	for i, a := range c.AssignedUsers {
		names[i] = a.Participant.DisplayName
	}

	return strings.Join(names, ", ")
}

func attachmentsSummary(c deck.Card, withDetail bool) string {
	if withDetail && len(c.Attachments) > 0 {
		names := make([]string, len(c.Attachments))
		for i, a := range c.Attachments {
			names[i] = a.Data
		}

		return fmt.Sprintf(
			"%d (%s)",
			c.AttachmentCount,
			strings.Join(names, ", "),
		)
	}

	return fmt.Sprintf("%d", c.AttachmentCount)
}

func availableCardColumnKeys() []string {
	keys := make([]string, 0, len(cardColumns))
	for k := range cardColumns {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
}

var cardCmd = &cobra.Command{
	Use:   "card <id>",
	Short: "Show a card's details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cardID, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid card id: %s", args[0])
		}

		client, err := newClient()
		if err != nil {
			return err
		}

		var boardID *int

		if cardViewBoardFlag != 0 {
			boardID = &cardViewBoardFlag
		}

		card, board, stack, err := client.FindCard(cardID, boardID)
		if err != nil {
			return err
		}

		fmt.Fprintf(os.Stdout, "ID:        %d\n", card.ID)
		fmt.Fprintf(os.Stdout, "Title:     %s\n", card.Title)
		fmt.Fprintf(os.Stdout, "Board:     [%d] %s\n", board.ID, board.Title)
		fmt.Fprintf(os.Stdout, "Stack:     %s\n", stack.Title)
		fmt.Fprintf(os.Stdout, "Archived:  %t\n", card.Archived)

		if card.DueDate != nil {
			fmt.Fprintf(os.Stdout, "Due date:  %s\n", *card.DueDate)
		}

		if card.Done != nil {
			fmt.Fprintf(os.Stdout, "Done on:   %s\n", *card.Done)
		}

		if len(card.Labels) > 0 {
			titles := make([]string, len(card.Labels))
			for i, l := range card.Labels {
				titles[i] = l.Title
			}

			fmt.Fprintf(
				os.Stdout,
				"Labels:    %s\n",
				strings.Join(titles, ", "),
			)
		}

		if len(card.AssignedUsers) > 0 {
			names := make([]string, len(card.AssignedUsers))
			for i, a := range card.AssignedUsers {
				names[i] = a.Participant.DisplayName
			}

			fmt.Fprintf(os.Stdout, "Assigned:  %s\n", strings.Join(names, ", "))
		}

		if card.Description != "" {
			fmt.Fprintf(os.Stdout, "\nDescription:\n%s\n", card.Description)
		}

		return nil
	},
}

var cardListCmd = &cobra.Command{
	Use:   "list",
	Short: "List a board's cards, grouped by stack",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}

		stacks, err := client.ListCards(cardListBoardFlag, deck.CardFilter{
			Statuses:  cardListStatusFlag,
			Stacks:    cardListStackFlag,
			Assignees: cardListAssigneeFlag,
		})
		if err != nil {
			return err
		}

		needAttachments := cardListAttachmentsFlag || cardListDownloadFlag != ""

		if needAttachments {
			for si, s := range stacks {
				for ci, c := range s.Cards {
					attachments, err := client.ListAttachments(
						cardListBoardFlag,
						s.ID,
						c.ID,
					)
					if err != nil {
						return err
					}

					stacks[si].Cards[ci].Attachments = attachments
				}
			}
		}

		if cardListDownloadFlag != "" {
			total := 0

			for _, s := range stacks {
				for _, c := range s.Cards {
					total += len(c.Attachments)
				}
			}

			if total == 0 {
				fmt.Fprintln(os.Stderr, "No attachments to download.")
			} else {
				bar := progressbar.NewOptions(
					total,
					progressbar.OptionSetDescription("Downloading attachments"),
					progressbar.OptionSetWriter(os.Stderr),
					progressbar.OptionShowCount(),
					progressbar.OptionOnCompletion(
						func() { fmt.Fprintln(os.Stderr) },
					),
				)

				for _, s := range stacks {
					for _, c := range s.Cards {
						for _, a := range c.Attachments {
							dir := filepath.Join(
								cardListDownloadFlag,
								"card",
								strconv.Itoa(c.ID),
								"attachments",
							)
							if err := os.MkdirAll(dir, 0o755); err != nil {
								return err
							}

							content, err := client.DownloadAttachment(
								cardListBoardFlag,
								s.ID,
								c.ID,
								a,
							)
							if err != nil {
								return err
							}

							dest := filepath.Join(dir, filepath.Base(a.Data))
							if err := os.WriteFile(
								dest,
								content,
								0o644,
							); err != nil {
								return err
							}

							_ = bar.Add(1)
						}
					}
				}
			}
		}

		if cardListJSONFlag {
			if cardListDownloadFlag != "" {
				if err := os.MkdirAll(cardListDownloadFlag, 0o755); err != nil {
					return err
				}

				jsonPath := filepath.Join(
					cardListDownloadFlag,
					fmt.Sprintf("board-%d.json", cardListBoardFlag),
				)

				content, err := json.MarshalIndent(stacks, "", "  ")
				if err != nil {
					return err
				}

				if err := os.WriteFile(jsonPath, content, 0o644); err != nil {
					return err
				}

				fmt.Fprintf(os.Stderr, "JSON saved to %s\n", jsonPath)
			}

			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")

			return enc.Encode(stacks)
		}

		columns := cardListColumnsFlag
		if len(columns) == 0 {
			columns = defaultCardListColumns
		}

		defs := make([]cardColumn, len(columns))

		headers := make([]interface{}, len(columns))
		for i, key := range columns {
			def, ok := cardColumns[strings.ToLower(key)]
			if !ok {
				return fmt.Errorf(
					"unknown column: %q (available: %s)",
					key,
					strings.Join(availableCardColumnKeys(), ", "),
				)
			}

			defs[i] = def
			headers[i] = def.header
		}

		tbl := table.New(
			headers...).WithHeaderFormatter(
			func(format string, vals ...interface{}) string {
				return strings.ToUpper(fmt.Sprintf(format, vals...))
			},
		)

		for _, s := range stacks {
			for _, c := range s.Cards {
				row := make([]interface{}, len(defs))
				for i, def := range defs {
					row[i] = def.value(s, c, cardListAttachmentsFlag)
				}

				tbl.AddRow(row...)
			}
		}

		tbl.Print()

		return nil
	},
}

func init() {
	cardCmd.Flags().IntVar(
		&cardViewBoardFlag,
		"board",
		0,
		"Restrict the search to this board (optional, otherwise searches all boards)",
	)

	cardListCmd.Flags().IntVar(
		&cardListBoardFlag,
		"board",
		0,
		"Board ID (required)",
	)
	_ = cardListCmd.MarkFlagRequired("board")
	cardListCmd.Flags().BoolVar(&cardListJSONFlag, "json", false, "JSON output")
	cardListCmd.Flags().BoolVar(
		&cardListAttachmentsFlag,
		"attachments",
		false,
		"Include attachment details for each card",
	)
	cardListCmd.Flags().StringVar(
		&cardListDownloadFlag,
		"download",
		"",
		"Download attachments (and the JSON with --json) into <dir> (use . for the current directory)",
	)
	cardListCmd.Flags().StringSliceVar(
		&cardListStatusFlag,
		"status",
		nil,
		fmt.Sprintf(
			"Only show cards with this status, comma-separated or repeated (available: %s)",
			strings.Join(deck.StatusKeys(), ", "),
		),
	)
	cardListCmd.Flags().StringSliceVar(
		&cardListStackFlag,
		"stack",
		nil,
		"Only show cards in this stack, by ID or title (case-insensitive), comma-separated or repeated",
	)
	cardListCmd.Flags().StringSliceVar(
		&cardListAssigneeFlag,
		"assignee",
		nil,
		"Only show cards assigned to this user, by uid or part of the display name (case-insensitive); \"me\" for yourself, \"none\" for unassigned cards; comma-separated or repeated",
	)
	cardListCmd.Flags().StringSliceVar(
		&cardListColumnsFlag,
		"column",
		defaultCardListColumns,
		fmt.Sprintf(
			"Table columns, comma-separated or repeated (available: %s)",
			strings.Join(availableCardColumnKeys(), ", "),
		),
	)

	cardCmd.AddCommand(cardListCmd)
	Cmd.AddCommand(cardCmd)
}
