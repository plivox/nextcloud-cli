package api

import (
	"encoding/base64"
	"strings"

	"github.com/studio-b12/gowebdav"
)

func (c *Client) DAV(root string) *gowebdav.Client {
	dav := gowebdav.NewAuthClient(
		strings.TrimRight(c.BaseURL, "/")+root,
		gowebdav.NewEmptyAuth(),
	)

	token := base64.StdEncoding.EncodeToString(
		[]byte(c.loginName + ":" + c.password),
	)
	dav.SetHeader("Authorization", "Basic "+token)
	dav.SetHeader("User-Agent", UserAgent)

	return dav
}
