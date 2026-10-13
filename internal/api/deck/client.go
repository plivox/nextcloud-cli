package deck

import "github.com/plivox/nextcloud-cli/internal/api"

const (
	deckAPIv10 = "/index.php/apps/deck/api/v1.0"
	deckAPIv11 = "/index.php/apps/deck/api/v1.1"
)

type Client struct {
	*api.Client
}

type pathParams = api.PathParams

func New(c *api.Client) *Client {
	return &Client{c}
}
