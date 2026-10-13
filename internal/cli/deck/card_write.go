package deckcmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/plivox/nextcloud-cli/internal/api/deck"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
)

var (
	cardWriteBoardFlag  int
	cardStackFlag       string
	cardTitleFlag       string
	cardDescriptionFlag string
	cardDueFlag         string
	cardDoneFlag        bool
	cardArchivedFlag    bool
	cardOrderFlag       int
)

type cardRef struct {
	client *deck.Client
	card   deck.Card
	board  deck.Board
	stack  deck.Stack
}

func findCard(arg string) (cardRef, error) {
	cardID, err := strconv.Atoi(arg)
	if err != nil {
		return cardRef{}, fmt.Errorf("invalid card id: %s", arg)
	}

	client, err := newClient()
	if err != nil {
		return cardRef{}, err
	}

	var boardID *int
	if cardWriteBoardFlag != 0 {
		boardID = &cardWriteBoardFlag
	}

	card, board, stack, err := client.FindCard(cardID, boardID)
	if err != nil {
		return cardRef{}, err
	}

	return cardRef{client, card, board, stack}, nil
}

func resolveStack(
	client *deck.Client,
	boardID int,
	ref string,
) (deck.Stack, error) {
	stacks, err := client.ListStacks(boardID)
	if err != nil {
		return deck.Stack{}, err
	}

	return deck.ResolveStack(stacks, ref)
}

var cardCreateCmd = &cobra.Command{
	Use:   "create <title>",
	Short: "Create a card",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}

		stack, err := resolveStack(client, cardWriteBoardFlag, cardStackFlag)
		if err != nil {
			return err
		}

		n := deck.NewCard{Title: args[0], Description: cardDescriptionFlag}
		if cardDueFlag != "" {
			if n.DueDate, err = deck.ParseDueDate(cardDueFlag); err != nil {
				return err
			}
		}

		card, err := client.CreateCard(cardWriteBoardFlag, stack.ID, n)
		if err != nil {
			return err
		}

		fmt.Fprintf(
			os.Stdout,
			"Created card [%d] %s in %s\n",
			card.ID,
			card.Title,
			stack.Title,
		)

		return nil
	},
}

var cardUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a card's title, description, due date, done or archived state",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var u deck.CardUpdate

		flags := cmd.Flags()
		if flags.Changed("title") {
			u.Title = &cardTitleFlag
		}

		if flags.Changed("description") {
			u.Description = &cardDescriptionFlag
		}

		if flags.Changed("due") {
			if cardDueFlag == "" || strings.EqualFold(cardDueFlag, "none") {
				u.ClearDueDate = true
			} else {
				due, err := deck.ParseDueDate(cardDueFlag)
				if err != nil {
					return err
				}

				u.DueDate = due
			}
		}

		if flags.Changed("done") {
			u.Done = &cardDoneFlag
		}

		if flags.Changed("archived") {
			u.Archived = &cardArchivedFlag
		}

		ref, err := findCard(args[0])
		if err != nil {
			return err
		}

		card, err := ref.client.UpdateCard(
			ref.board.ID,
			ref.stack.ID,
			ref.card.ID,
			u,
		)
		if err != nil {
			return err
		}

		fmt.Fprintf(os.Stdout, "Updated card [%d] %s\n", card.ID, card.Title)

		return nil
	},
}

var cardMoveCmd = &cobra.Command{
	Use:   "move <id>",
	Short: "Move a card to another stack, or to another position",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ref, err := findCard(args[0])
		if err != nil {
			return err
		}

		target := ref.stack
		if cardStackFlag != "" {
			if target, err = resolveStack(
				ref.client,
				ref.board.ID,
				cardStackFlag,
			); err != nil {
				return err
			}
		}

		if err := ref.client.MoveCard(
			ref.board.ID,
			ref.stack.ID,
			ref.card.ID,
			target.ID,
			cardOrderFlag,
		); err != nil {
			return err
		}

		fmt.Fprintf(
			os.Stdout,
			"Moved card [%d] %s to %s\n",
			ref.card.ID,
			ref.card.Title,
			target.Title,
		)

		return nil
	},
}

var cardDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a card (it can be restored from the Deck web interface)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ref, err := findCard(args[0])
		if err != nil {
			return err
		}

		if err := ref.client.DeleteCard(
			ref.board.ID,
			ref.stack.ID,
			ref.card.ID,
		); err != nil {
			return err
		}

		fmt.Fprintf(
			os.Stdout,
			"Deleted card [%d] %s\n",
			ref.card.ID,
			ref.card.Title,
		)

		return nil
	},
}

func userCommand(
	use, short, done string,
	action func(*deck.Client, int, int, int, string) error,
) *cobra.Command {
	return &cobra.Command{
		Use:   use + " <id> <user>...",
		Short: short,
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref, err := findCard(args[0])
			if err != nil {
				return err
			}

			for _, user := range args[1:] {
				if strings.EqualFold(user, "me") {
					user = ref.client.Username
				}

				if err := action(
					ref.client,
					ref.board.ID,
					ref.stack.ID,
					ref.card.ID,
					user,
				); err != nil {
					return err
				}

				fmt.Fprintf(
					os.Stdout,
					"%s %s on card [%d]\n",
					done,
					user,
					ref.card.ID,
				)
			}

			return nil
		},
	}
}

func labelCommand(
	use, short, done string,
	action func(*deck.Client, int, int, int, int) error,
) *cobra.Command {
	return &cobra.Command{
		Use:   use + " <id> <label>...",
		Short: short,
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref, err := findCard(args[0])
			if err != nil {
				return err
			}

			board, err := ref.client.GetBoard(ref.board.ID)
			if err != nil {
				return err
			}

			for _, name := range args[1:] {
				label, err := deck.ResolveLabel(board.Labels, name)
				if err != nil {
					return err
				}

				if err := action(
					ref.client,
					ref.board.ID,
					ref.stack.ID,
					ref.card.ID,
					label.ID,
				); err != nil {
					return err
				}

				fmt.Fprintf(
					os.Stdout,
					"%s label %s on card [%d]\n",
					done,
					label.Title,
					ref.card.ID,
				)
			}

			return nil
		},
	}
}

var cardCommentCmd = &cobra.Command{
	Use:   "comment <id> <message>",
	Short: "Add a comment to a card",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cardID, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid card id: %s", args[0])
		}

		client, err := newClient()
		if err != nil {
			return err
		}

		comment, err := client.AddComment(cardID, args[1])
		if err != nil {
			return err
		}

		fmt.Fprintf(
			os.Stdout,
			"Added comment %d on card [%d]\n",
			comment.ID,
			cardID,
		)

		return nil
	},
}

var cardCommentsCmd = &cobra.Command{
	Use:   "comments <id>",
	Short: "List a card's comments",
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

		comments, err := client.ListComments(cardID)
		if err != nil {
			return err
		}

		if len(comments) == 0 {
			fmt.Fprintln(os.Stdout, "No comments.")

			return nil
		}

		tbl := table.New("Date", "Author", "Message")
		for _, c := range comments {
			tbl.AddRow(c.CreationDateTime, c.ActorDisplayName, c.Message)
		}

		tbl.Print()

		return nil
	},
}

func init() {
	cardCreateCmd.Flags().IntVar(
		&cardWriteBoardFlag,
		"board",
		0,
		"Board ID (required)",
	)
	_ = cardCreateCmd.MarkFlagRequired("board")
	cardCreateCmd.Flags().StringVar(
		&cardStackFlag,
		"stack",
		"",
		"Stack ID or title (required)",
	)
	_ = cardCreateCmd.MarkFlagRequired("stack")
	cardCreateCmd.Flags().StringVar(
		&cardDescriptionFlag,
		"description",
		"",
		"Markdown description",
	)
	cardCreateCmd.Flags().StringVar(
		&cardDueFlag,
		"due",
		"",
		"Due date: YYYY-MM-DD, \"YYYY-MM-DD HH:MM\" or RFC 3339",
	)

	cardUpdateCmd.Flags().StringVar(&cardTitleFlag, "title", "", "New title")
	cardUpdateCmd.Flags().StringVar(
		&cardDescriptionFlag,
		"description",
		"",
		"New Markdown description",
	)
	cardUpdateCmd.Flags().StringVar(
		&cardDueFlag,
		"due",
		"",
		"New due date (YYYY-MM-DD, ...), or \"none\" to remove it",
	)
	cardUpdateCmd.Flags().BoolVar(
		&cardDoneFlag,
		"done",
		false,
		"Mark as done (--done) or not done (--done=false)",
	)
	cardUpdateCmd.Flags().BoolVar(
		&cardArchivedFlag,
		"archived",
		false,
		"Archive (--archived) or unarchive (--archived=false)",
	)

	cardMoveCmd.Flags().StringVar(
		&cardStackFlag,
		"stack",
		"",
		"Target stack ID or title (default: the current stack)",
	)
	cardMoveCmd.Flags().IntVar(
		&cardOrderFlag,
		"order",
		0,
		"Position in the stack (0 = top)",
	)

	assignCmd := userCommand(
		"assign",
		"Assign users to a card (\"me\" for yourself)",
		"Assigned",
		(*deck.Client).AssignUser,
	)
	unassignCmd := userCommand(
		"unassign",
		"Unassign users from a card",
		"Unassigned",
		(*deck.Client).UnassignUser,
	)
	labelCmd := labelCommand(
		"label",
		"Add labels to a card, by ID or title",
		"Added",
		(*deck.Client).AssignLabel,
	)
	unlabelCmd := labelCommand(
		"unlabel",
		"Remove labels from a card",
		"Removed",
		(*deck.Client).RemoveLabel,
	)

	for _, c := range []*cobra.Command{cardUpdateCmd, cardMoveCmd, cardDeleteCmd, assignCmd, unassignCmd, labelCmd, unlabelCmd} {
		c.Flags().IntVar(
			&cardWriteBoardFlag,
			"board",
			0,
			"Board ID (optional, speeds up finding the card)",
		)
	}

	cardCmd.AddCommand(
		cardCreateCmd,
		cardUpdateCmd,
		cardMoveCmd,
		cardDeleteCmd,
		assignCmd,
		unassignCmd,
		labelCmd,
		unlabelCmd,
		cardCommentCmd,
		cardCommentsCmd,
	)
}
