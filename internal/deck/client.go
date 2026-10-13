package deck

import "github.com/plivox/nextcloud-cli/internal/nextcloud"

const (
	deckAPIv10 = "/index.php/apps/deck/api/v1.0"
	deckAPIv11 = "/index.php/apps/deck/api/v1.1"
)

type Client struct {
	*nextcloud.Client
}

type pathParams = nextcloud.PathParams

func NewClient() (*Client, error) {
	nc, err := nextcloud.NewClient()
	if err != nil {
		return nil, err
	}

	return &Client{nc}, nil
}
