package mcp

import (
	"context"
	"fmt"
	"io"
	"path"
	"unicode/utf8"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	defaultReadLimit = 256 * 1024
	maxReadLimit     = 1024 * 1024
)

type pathInput struct {
	Path string `json:"path" jsonschema:"Path in the user's Nextcloud files, e.g. /Documents/notes.md"`
}

func cleanPath(p string) string {
	return path.Clean("/" + p)
}

func (srv *server) addFilesReadTools(s *mcp.Server) {
	addTool(
		s,
		"files_list",
		"List a folder in the user's Nextcloud files.",
		readOnly,
		func(_ context.Context, in pathInput) (any, error) {
			entries, err := srv.files.ReadDir(cleanPath(in.Path))
			if err != nil {
				return nil, err
			}

			type entry struct {
				Name     string `json:"name"`
				IsDir    bool   `json:"isDir"`
				Size     int64  `json:"size"`
				Modified string `json:"modified"`
			}

			out := make([]entry, len(entries))
			for i, e := range entries {
				out[i] = entry{
					e.Name(),
					e.IsDir(),
					e.Size(),
					e.ModTime().Format("2006-01-02T15:04:05Z07:00"),
				}
			}

			return map[string]any{
				"path":    cleanPath(in.Path),
				"entries": out,
			}, nil
		},
	)

	type readInput struct {
		Path     string `json:"path"                jsonschema:"Path of a text file in the user's Nextcloud files"`
		MaxBytes int    `json:"max_bytes,omitempty" jsonschema:"Maximum size to read, default 262144, at most 1048576"`
	}

	addTool(
		s,
		"files_read",
		"Read a text file from the user's Nextcloud files.",
		readOnly,
		func(_ context.Context, in readInput) (any, error) {
			limit := in.MaxBytes
			if limit <= 0 {
				limit = defaultReadLimit
			}

			limit = min(limit, maxReadLimit)

			stream, err := srv.files.ReadStream(cleanPath(in.Path))
			if err != nil {
				return nil, err
			}
			defer stream.Close()

			content, err := io.ReadAll(io.LimitReader(stream, int64(limit)+1))
			if err != nil {
				return nil, err
			}

			if len(content) > limit {
				return nil, fmt.Errorf("file is larger than %d bytes", limit)
			}

			if !utf8.Valid(content) {
				return nil, fmt.Errorf("%s is not a text file", in.Path)
			}

			return map[string]any{
				"path":    cleanPath(in.Path),
				"content": string(content),
			}, nil
		},
	)
}

func (srv *server) addFilesWriteTools(s *mcp.Server) {
	type writeInput struct {
		Path    string `json:"path"    jsonschema:"Destination path; an existing file is overwritten"`
		Content string `json:"content" jsonschema:"Text content of the file"`
	}

	addTool(
		s,
		"files_write",
		"Write a text file to the user's Nextcloud files, overwriting it if it exists.",
		destructive,
		func(_ context.Context, in writeInput) (any, error) {
			target := cleanPath(in.Path)

			return map[string]any{"written": target, "bytes": len(in.Content)},
				srv.files.Write(target, []byte(in.Content), 0o644)
		},
	)

	addTool(
		s,
		"files_mkdir",
		"Create a folder (and missing parents) in the user's Nextcloud files.",
		additive,
		func(_ context.Context, in pathInput) (any, error) {
			target := cleanPath(in.Path)

			return map[string]any{
				"created": target,
			}, srv.files.MkdirAll(
				target,
				0o755,
			)
		},
	)

	type deleteInput struct {
		Path      string `json:"path"                jsonschema:"Path of the file or folder to delete"`
		Recursive bool   `json:"recursive,omitempty" jsonschema:"Required to delete a folder"`
	}

	addTool(
		s,
		"files_delete",
		"Delete a file, or a folder with recursive=true.",
		destructive,
		func(_ context.Context, in deleteInput) (any, error) {
			target := cleanPath(in.Path)
			if target == "/" {
				return nil, fmt.Errorf("refusing to delete the root folder")
			}

			info, err := srv.files.Stat(target)
			if err != nil {
				return nil, err
			}

			if info.IsDir() && !in.Recursive {
				return nil, fmt.Errorf(
					"%s is a folder (set recursive to delete it)",
					target,
				)
			}

			return map[string]any{"deleted": target}, srv.files.Remove(target)
		},
	)
}
