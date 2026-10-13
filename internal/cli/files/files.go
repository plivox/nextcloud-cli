package filescmd

import (
	"fmt"
	"path"

	"github.com/plivox/nextcloud-cli/internal/api/files"
	"github.com/plivox/nextcloud-cli/internal/auth"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "files",
	Short: "Nextcloud Files: list, download, upload and delete",
}

func newClient() (*files.Client, error) {
	c, err := auth.NewClient()
	if err != nil {
		return nil, err
	}

	return files.New(c), nil
}

func cleanPath(p string) string {
	return path.Clean("/" + p)
}

func humanSize(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}

	div, exp := int64(unit), 0
	for m := n / unit; m >= unit; m /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %ciB", float64(n)/float64(div), "KMGTPE"[exp])
}
