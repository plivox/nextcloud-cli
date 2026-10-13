package filescmd

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get <remote> [local]",
	Short: "Download a file (to the current directory by default)",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		remote := cleanPath(args[0])

		local := path.Base(remote)
		if len(args) == 2 {
			local = args[1]
		}

		if info, err := os.Stat(local); err == nil && info.IsDir() {
			local = filepath.Join(local, path.Base(remote))
		}

		client, err := newClient()
		if err != nil {
			return err
		}

		info, err := client.Stat(remote)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return fmt.Errorf("%s is a folder", remote)
		}

		stream, err := client.ReadStream(remote)
		if err != nil {
			return err
		}
		defer stream.Close()

		out, err := os.Create(local)
		if err != nil {
			return err
		}
		defer out.Close()

		bar := progressbar.NewOptions64(info.Size(),
			progressbar.OptionSetDescription(path.Base(remote)),
			progressbar.OptionSetWriter(os.Stderr),
			progressbar.OptionShowBytes(true),
			progressbar.OptionOnCompletion(func() { fmt.Fprintln(os.Stderr) }),
		)

		if _, err := io.Copy(io.MultiWriter(out, bar), stream); err != nil {
			return err
		}

		fmt.Fprintf(os.Stdout, "Downloaded %s to %s\n", remote, local)

		return nil
	},
}

func init() {
	Cmd.AddCommand(getCmd)
}
