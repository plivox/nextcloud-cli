package mcp

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/plivox/nextcloud-cli/internal/api/deck"
	"github.com/plivox/nextcloud-cli/internal/api/files"
	"github.com/plivox/nextcloud-cli/internal/auth"
)

type Options struct {
	AllowWrite bool
	Version    string
}

type server struct {
	deck  *deck.Client
	files *files.Client
}

func Run(ctx context.Context, opts Options) error {
	c, err := auth.NewClient()
	if err != nil {
		return err
	}

	srv := &server{deck: deck.New(c), files: files.New(c)}

	s := mcp.NewServer(&mcp.Implementation{
		Name:    "nextcloud-cli",
		Title:   "Nextcloud",
		Version: opts.Version,
	}, nil)

	srv.addDeckReadTools(s)
	srv.addFilesReadTools(s)

	if opts.AllowWrite {
		srv.addDeckWriteTools(s)
		srv.addFilesWriteTools(s)
	}

	return s.Run(ctx, &mcp.StdioTransport{})
}

func addTool[In any](
	s *mcp.Server,
	name, description string,
	annotations *mcp.ToolAnnotations,
	handler func(context.Context, In) (any, error),
) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        name,
		Description: description,
		Annotations: annotations,
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in In) (*mcp.CallToolResult, any, error) {
		out, err := handler(ctx, in)

		return nil, out, err
	})
}

var (
	falseValue = false
	trueValue  = true

	readOnly = &mcp.ToolAnnotations{
		ReadOnlyHint:  true,
		OpenWorldHint: &falseValue,
	}
	additive = &mcp.ToolAnnotations{
		DestructiveHint: &falseValue,
		OpenWorldHint:   &falseValue,
	}
	destructive = &mcp.ToolAnnotations{
		DestructiveHint: &trueValue,
		OpenWorldHint:   &falseValue,
	}
)
