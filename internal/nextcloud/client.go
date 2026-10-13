package nextcloud

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/go-resty/resty/v2"
)

type Client struct {
	BaseURL  string
	Username string
	r        *resty.Client
}

type PathParams map[string]int

func NewClientFromEnv() (*Client, error) {
	baseURL := os.Getenv("NEXTCLOUD_BASE_URL")
	username := os.Getenv("NEXTCLOUD_USERNAME")
	password := os.Getenv("NEXTCLOUD_PASSWORD")

	if baseURL == "" || username == "" || password == "" {
		return nil, fmt.Errorf(
			"NEXTCLOUD_BASE_URL, NEXTCLOUD_USERNAME and NEXTCLOUD_PASSWORD must be set (via .env or the environment)",
		)
	}

	r := resty.New().
		SetBaseURL(baseURL).
		SetBasicAuth(username, password)

	return &Client{
		BaseURL:  baseURL,
		Username: username,
		r:        r,
	}, nil
}

func (c *Client) R() *resty.Request {
	return c.r.R()
}

func (c *Client) OCSRequest(params PathParams) *resty.Request {
	req := c.R().SetHeader("OCS-APIRequest", "true")
	for k, v := range params {
		req.SetPathParam(k, strconv.Itoa(v))
	}

	return req
}

func (c *Client) GetJSON(
	path string,
	params PathParams,
	out interface{},
) error {
	resp, err := c.OCSRequest(params).
		SetHeader("Accept", "application/json").
		SetResult(out).
		Get(path)
	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf(
			"request to %s failed: %s",
			resp.Request.URL,
			resp.Status(),
		)
	}

	return nil
}
