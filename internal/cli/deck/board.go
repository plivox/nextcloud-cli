package deckcmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

var boardCmd = &cobra.Command{
	Use:   "board [id]",
	Short: "Show a board's metadata",
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

		board, err := client.GetBoard(boardID)
		if err != nil {
			return err
		}

		fmt.Fprintf(os.Stdout, "ID:       %d\n", board.ID)
		fmt.Fprintf(os.Stdout, "Title:    %s\n", board.Title)
		fmt.Fprintf(
			os.Stdout,
			"Owner:    %s (%s)\n",
			board.Owner.DisplayName,
			board.Owner.UID,
		)
		fmt.Fprintf(os.Stdout, "Color:    #%s\n", board.Color)
		fmt.Fprintf(os.Stdout, "Archived: %t\n", board.Archived)

		return nil
	},
}

var boardListCmd = &cobra.Command{
	Use:   "list",
	Short: "List boards",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}

		boards, err := client.ListBoards()
		if err != nil {
			return err
		}

		for _, b := range boards {
			fmt.Fprintf(os.Stdout, "[%d] %s\n", b.ID, b.Title)
		}

		return nil
	},
}

func init() {
	boardCmd.AddCommand(boardListCmd)
	Cmd.AddCommand(boardCmd)
}
