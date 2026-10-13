package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-resty/resty/v2"
)

type Client struct {
	BaseURL   string
	Username  string
	loginName string
	password  string
	r         *resty.Client
}

type PathParams map[string]int

const UserAgent = "nextcloud-cli"

func NewClient(server, loginName, password, userID string) *Client {
	r := resty.New().
		SetBaseURL(server).
		SetBasicAuth(loginName, password).
		SetHeader("User-Agent", UserAgent)

	return &Client{
		BaseURL:   server,
		Username:  userID,
		loginName: loginName,
		password:  password,
		r:         r,
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
	return c.SendJSON(http.MethodGet, path, params, nil, out)
}

func (c *Client) SendJSON(
	method, path string,
	params PathParams,
	body, out interface{},
) error {
	req := c.OCSRequest(params).SetHeader("Accept", "application/json")
	if body != nil {
		req.SetBody(body)
	}

	if out != nil {
		req.SetResult(out)
	}

	resp, err := req.Execute(method, path)
	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		msg := resp.Status()

		var apiErr struct {
			Message string `json:"message"`
		}
		if json.Unmarshal(resp.Body(), &apiErr) == nil && apiErr.Message != "" {
			msg += ": " + apiErr.Message
		}

		return fmt.Errorf("%s %s failed: %s", method, resp.Request.URL, msg)
	}

	return nil
}
