package deckcmd

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

var attachmentListBoardFlag int

var attachmentCmd = &cobra.Command{
	Use:   "attachment",
	Short: "Manage a card's attachments",
}

var attachmentListCmd = &cobra.Command{
	Use:   "list <card_id>",
	Short: "List a card's attachments",
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

		if attachmentListBoardFlag != 0 {
			boardID = &attachmentListBoardFlag
		}

		_, board, stack, err := client.FindCard(cardID, boardID)
		if err != nil {
			return err
		}

		attachments, err := client.ListAttachments(board.ID, stack.ID, cardID)
		if err != nil {
			return err
		}

		if len(attachments) == 0 {
			fmt.Fprintln(os.Stdout, "No attachments.")
			return nil
		}

		for _, a := range attachments {
			size := ""
			if fs, ok := a.ExtendedData["filesize"]; ok {
				size = fmt.Sprintf(" (%.0f bytes)", toFloat(fs))
			}

			created := time.Unix(a.CreatedAt, 0).Format("2006-01-02 15:04")
			fmt.Fprintf(
				os.Stdout,
				"[%d] %s%s — %s, added by %s on %s\n",
				a.ID,
				a.Data,
				size,
				a.Type,
				a.CreatedBy,
				created,
			)
		}

		return nil
	},
}

func toFloat(v interface{}) float64 {
	switch n := v.(type) {
	case float64:
		return n
	default:
		return 0
	}
}

func init() {
	attachmentListCmd.Flags().IntVar(
		&attachmentListBoardFlag,
		"board",
		0,
		"Restrict the card search to this board (optional)",
	)

	attachmentCmd.AddCommand(attachmentListCmd)
	Cmd.AddCommand(attachmentCmd)
}
