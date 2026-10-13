package filescmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var mkdirCmd = &cobra.Command{
	Use:   "mkdir <path>",
	Short: "Create a folder, including missing parent folders",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := cleanPath(args[0])

		client, err := newClient()
		if err != nil {
			return err
		}

		if err := client.MkdirAll(target, 0o755); err != nil {
			return err
		}

		fmt.Fprintf(os.Stdout, "Created %s\n", target)

		return nil
	},
}

func init() {
	Cmd.AddCommand(mkdirCmd)
}
