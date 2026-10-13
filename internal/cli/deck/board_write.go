package deckcmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/plivox/nextcloud-cli/internal/api/deck"
	"github.com/spf13/cobra"
)

var (
	boardColorFlag    string
	boardTitleFlag    string
	boardArchivedFlag bool
)

var boardCreateCmd = &cobra.Command{
	Use:   "create <title>",
	Short: "Create a board",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}

		board, err := client.CreateBoard(args[0], boardColorFlag)
		if err != nil {
			return err
		}

		fmt.Fprintf(os.Stdout, "Created board [%d] %s\n", board.ID, board.Title)

		return nil
	},
}

var boardUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a board's title, color or archived state",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		boardID, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid board id: %s", args[0])
		}

		var u deck.BoardUpdate
		if cmd.Flags().Changed("title") {
			u.Title = &boardTitleFlag
		}

		if cmd.Flags().Changed("color") {
			u.Color = &boardColorFlag
		}

		if cmd.Flags().Changed("archived") {
			u.Archived = &boardArchivedFlag
		}

		client, err := newClient()
		if err != nil {
			return err
		}

		board, err := client.UpdateBoard(boardID, u)
		if err != nil {
			return err
		}

		fmt.Fprintf(os.Stdout, "Updated board [%d] %s\n", board.ID, board.Title)

		return nil
	},
}

var boardDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a board (it can be restored from the Deck web interface)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		boardID, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid board id: %s", args[0])
		}

		client, err := newClient()
		if err != nil {
			return err
		}

		if err := client.DeleteBoard(boardID); err != nil {
			return err
		}

		fmt.Fprintf(os.Stdout, "Deleted board %d\n", boardID)

		return nil
	},
}

func init() {
	boardCreateCmd.Flags().StringVar(
		&boardColorFlag,
		"color",
		"0082c9",
		"Hexadecimal color, without #",
	)

	boardUpdateCmd.Flags().StringVar(&boardTitleFlag, "title", "", "New title")
	boardUpdateCmd.Flags().StringVar(
		&boardColorFlag,
		"color",
		"",
		"New hexadecimal color, without #",
	)
	boardUpdateCmd.Flags().BoolVar(
		&boardArchivedFlag,
		"archived",
		false,
		"Archive (--archived) or unarchive (--archived=false) the board",
	)

	boardCmd.AddCommand(boardCreateCmd, boardUpdateCmd, boardDeleteCmd)
}
