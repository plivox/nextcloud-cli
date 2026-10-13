package cmd

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	deckcmd "github.com/plivox/nextcloud-cli/cmd/deck"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:           "nextcloud",
	Short:         "Command-line client for Nextcloud",
	SilenceErrors: true,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		cmd.SilenceUsage = true
	},
}

func init() {
	rootCmd.AddCommand(deckcmd.Cmd)
}

func Execute() {
	_ = godotenv.Load()

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
