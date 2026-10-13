package nextcloud

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-resty/resty/v2"
)

type Client struct {
	BaseURL  string
	Username string
	r        *resty.Client
}

type PathParams map[string]int

func NewClient() (*Client, error) {
	creds, err := LoadCredentials()
	if err != nil {
		return nil, err
	}

	return NewClientFromCredentials(creds), nil
}

func NewClientFromCredentials(creds Credentials) *Client {
	r := resty.New().
		SetBaseURL(creds.Server).
		SetBasicAuth(creds.LoginName, creds.AppPassword).
		SetHeader("User-Agent", userAgent)

	return &Client{
		BaseURL:  creds.Server,
		Username: creds.UserID,
		r:        r,
	}
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
