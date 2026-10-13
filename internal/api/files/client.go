package files

import (
	"net/url"

	"github.com/plivox/nextcloud-cli/internal/api"
	"github.com/studio-b12/gowebdav"
)

type Client struct {
	*gowebdav.Client
}

func New(c *api.Client) *Client {
	return &Client{c.DAV("/remote.php/dav/files/" + url.PathEscape(c.Username))}
}
