package cli

import (
	"github.com/plivox/nextcloud-cli/internal/mcp"
	"github.com/spf13/cobra"
)

var mcpAllowWriteFlag bool

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Run an MCP server on stdio, for AI assistants",
	Long: `Run a Model Context Protocol server on stdin/stdout.

Read-only tools are always available. Tools that create, change or delete
Deck boards, stacks, cards and files are only exposed with --allow-write.

Example (Claude Code):
  claude mcp add nextcloud -- nextcloud mcp --allow-write`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return mcp.Run(cmd.Context(), mcp.Options{
			AllowWrite: mcpAllowWriteFlag,
			Version:    version(),
		})
	},
}

func init() {
	mcpCmd.Flags().BoolVar(
		&mcpAllowWriteFlag,
		"allow-write",
		false,
		"Expose tools that modify Deck and Files",
	)

	rootCmd.AddCommand(mcpCmd)
}
