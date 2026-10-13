package api

import (
	"fmt"
	"net/http"
)

func (c *Client) CurrentUserID() (string, error) {
	var user struct {
		OCS struct {
			Data struct {
				ID string `json:"id"`
			} `json:"data"`
		} `json:"ocs"`
	}

	if err := c.GetJSON("/ocs/v2.php/cloud/user", nil, &user); err != nil {
		return "", err
	}

	return user.OCS.Data.ID, nil
}

func (c *Client) RevokeAppPassword() error {
	resp, err := c.OCSRequest(nil).Delete("/ocs/v2.php/core/apppassword")
	if err != nil {
		return err
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("revoking the app password failed: %s", resp.Status())
	}

	return nil
}
