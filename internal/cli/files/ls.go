package filescmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/rodaine/table"
	"github.com/spf13/cobra"
)

var lsCmd = &cobra.Command{
	Use:   "ls [path]",
	Short: "List a folder",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := "/"
		if len(args) == 1 {
			target = cleanPath(args[0])
		}

		client, err := newClient()
		if err != nil {
			return err
		}

		info, err := client.Stat(target)
		if err != nil {
			return err
		}

		entries := []os.FileInfo{info}

		if info.IsDir() {
			entries, err = client.ReadDir(target)
			if err != nil {
				return err
			}
		}

		sort.Slice(entries, func(i, j int) bool {
			if entries[i].IsDir() != entries[j].IsDir() {
				return entries[i].IsDir()
			}

			return strings.ToLower(
				entries[i].Name(),
			) < strings.ToLower(
				entries[j].Name(),
			)
		})

		tbl := table.New("Name", "Size", "Modified").
			WithHeaderFormatter(
				func(format string, vals ...interface{}) string {
					return strings.ToUpper(fmt.Sprintf(format, vals...))
				},
			)

		for _, e := range entries {
			name, size := e.Name(), humanSize(e.Size())
			if e.IsDir() {
				name, size = name+"/", ""
			}

			tbl.AddRow(
				name,
				size,
				e.ModTime().Local().Format("2006-01-02 15:04"),
			)
		}

		tbl.Print()

		return nil
	},
}

func init() {
	Cmd.AddCommand(lsCmd)
}
