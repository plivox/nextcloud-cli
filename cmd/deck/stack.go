package deckcmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/plivox/nextcloud-cli/internal/deck"

	"github.com/spf13/cobra"
)

var stackViewBoardFlag int
var stackListBoardFlag int

var stackCmd = &cobra.Command{
	Use:   "stack <id>",
	Short: "Show a stack's (column's) details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		stackID, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid stack id: %s", args[0])
		}

		client, err := deck.NewClientFromEnv()
		if err != nil {
			return err
		}

		stack, err := client.GetStack(stackViewBoardFlag, stackID)
		if err != nil {
			return err
		}

		fmt.Fprintf(os.Stdout, "ID:    %d\n", stack.ID)
		fmt.Fprintf(os.Stdout, "Title: %s\n", stack.Title)
		fmt.Fprintf(os.Stdout, "Board: %d\n", stack.BoardID)
		fmt.Fprintf(os.Stdout, "Order: %d\n", stack.Order)
		fmt.Fprintf(os.Stdout, "Cards: %d\n", len(stack.Cards))

		return nil
	},
}

var stackListCmd = &cobra.Command{
	Use:   "list",
	Short: "List a board's stacks (columns)",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := deck.NewClientFromEnv()
		if err != nil {
			return err
		}

		stacks, err := client.ListStacks(stackListBoardFlag)
		if err != nil {
			return err
		}

		for _, s := range stacks {
			fmt.Fprintf(
				os.Stdout,
				"[%d] %s (order: %d, %d card(s))\n",
				s.ID,
				s.Title,
				s.Order,
				len(s.Cards),
			)
		}

		return nil
	},
}

func init() {
	stackCmd.Flags().IntVar(
		&stackViewBoardFlag,
		"board",
		0,
		"Board ID (required)",
	)
	_ = stackCmd.MarkFlagRequired("board")

	stackListCmd.Flags().IntVar(
		&stackListBoardFlag,
		"board",
		0,
		"Board ID (required)",
	)
	_ = stackListCmd.MarkFlagRequired("board")

	stackCmd.AddCommand(stackListCmd)
	Cmd.AddCommand(stackCmd)
}
