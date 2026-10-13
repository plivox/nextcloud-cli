package filescmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rmRecursiveFlag bool

var rmCmd = &cobra.Command{
	Use:   "rm <path>",
	Short: "Delete a file, or a folder with -r",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := cleanPath(args[0])
		if target == "/" {
			return fmt.Errorf("refusing to delete the root folder")
		}

		client, err := newClient()
		if err != nil {
			return err
		}

		info, err := client.Stat(target)
		if err != nil {
			return err
		}

		if info.IsDir() && !rmRecursiveFlag {
			return fmt.Errorf("%s is a folder (use -r to delete it)", target)
		}

		if err := client.Remove(target); err != nil {
			return err
		}

		fmt.Fprintf(os.Stdout, "Deleted %s\n", target)

		return nil
	},
}

func init() {
	rmCmd.Flags().BoolVarP(
		&rmRecursiveFlag,
		"recursive",
		"r",
		false,
		"Delete a folder and its content",
	)

	Cmd.AddCommand(rmCmd)
}
