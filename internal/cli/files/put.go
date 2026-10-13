package filescmd

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

var putCmd = &cobra.Command{
	Use:   "put <local> [remote]",
	Short: "Upload a file (to the root folder by default)",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		local := args[0]

		info, err := os.Stat(local)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return fmt.Errorf("%s is a folder", local)
		}

		client, err := newClient()
		if err != nil {
			return err
		}

		remote := "/"
		if len(args) == 2 {
			remote = args[1]
		}

		isDir := strings.HasSuffix(remote, "/")
		remote = cleanPath(remote)

		if !isDir {
			if target, err := client.Stat(
				remote,
			); err == nil &&
				target.IsDir() {
				isDir = true
			}
		}

		if isDir {
			remote = path.Join(remote, filepath.Base(local))
		}

		in, err := os.Open(local)
		if err != nil {
			return err
		}
		defer in.Close()

		bar := progressbar.NewOptions64(info.Size(),
			progressbar.OptionSetDescription(filepath.Base(local)),
			progressbar.OptionSetWriter(os.Stderr),
			progressbar.OptionShowBytes(true),
			progressbar.OptionOnCompletion(func() { fmt.Fprintln(os.Stderr) }),
		)
		reader := progressbar.NewReader(in, bar)

		if err := client.WriteStreamWithLength(
			remote,
			&reader,
			info.Size(),
			0o644,
		); err != nil {
			return err
		}

		fmt.Fprintf(os.Stdout, "Uploaded %s to %s\n", local, remote)

		return nil
	},
}

func init() {
	Cmd.AddCommand(putCmd)
}
