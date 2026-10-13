package deckcmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/plivox/nextcloud-cli/internal/api/deck"
	"github.com/spf13/cobra"
)

var (
	stackWriteBoardFlag int
	stackTitleFlag      string
	stackOrderFlag      int
)

var stackCreateCmd = &cobra.Command{
	Use:   "create <title>",
	Short: "Create a stack (column) in a board",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}

		stack, err := client.CreateStack(
			stackWriteBoardFlag,
			args[0],
			stackOrderFlag,
		)
		if err != nil {
			return err
		}

		fmt.Fprintf(os.Stdout, "Created stack [%d] %s\n", stack.ID, stack.Title)

		return nil
	},
}

var stackUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Rename or reorder a stack",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		stackID, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid stack id: %s", args[0])
		}

		var u deck.StackUpdate
		if cmd.Flags().Changed("title") {
			u.Title = &stackTitleFlag
		}

		if cmd.Flags().Changed("order") {
			u.Order = &stackOrderFlag
		}

		client, err := newClient()
		if err != nil {
			return err
		}

		stack, err := client.UpdateStack(stackWriteBoardFlag, stackID, u)
		if err != nil {
			return err
		}

		fmt.Fprintf(os.Stdout, "Updated stack [%d] %s\n", stack.ID, stack.Title)

		return nil
	},
}

var stackDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a stack and its cards",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		stackID, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid stack id: %s", args[0])
		}

		client, err := newClient()
		if err != nil {
			return err
		}

		if err := client.DeleteStack(stackWriteBoardFlag, stackID); err != nil {
			return err
		}

		fmt.Fprintf(os.Stdout, "Deleted stack %d\n", stackID)

		return nil
	},
}

func init() {
	for _, c := range []*cobra.Command{stackCreateCmd, stackUpdateCmd, stackDeleteCmd} {
		c.Flags().IntVar(
			&stackWriteBoardFlag,
			"board",
			0,
			"Board ID (required)",
		)
		_ = c.MarkFlagRequired("board")
	}

	stackCreateCmd.Flags().IntVar(
		&stackOrderFlag,
		"order",
		999,
		"Position among the stacks",
	)
	stackUpdateCmd.Flags().StringVar(&stackTitleFlag, "title", "", "New title")
	stackUpdateCmd.Flags().IntVar(
		&stackOrderFlag,
		"order",
		0,
		"New position among the stacks",
	)

	stackCmd.AddCommand(stackCreateCmd, stackUpdateCmd, stackDeleteCmd)
}
