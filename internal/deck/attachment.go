package deck

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
)

type Attachment struct {
	ID           int                    `json:"id"`
	CardID       int                    `json:"cardId"`
	Type         string                 `json:"type"`
	Data         string                 `json:"data"`
	CreatedAt    int64                  `json:"createdAt"`
	CreatedBy    string                 `json:"createdBy"`
	LastModified int64                  `json:"lastModified"`
	DeletedAt    int64                  `json:"deletedAt"`
	ExtendedData map[string]interface{} `json:"extendedData"`
}

func (c *Client) ListAttachments(
	boardID, stackID, cardID int,
) ([]Attachment, error) {
	var attachments []Attachment

	params := pathParams{
		"boardId": boardID,
		"stackId": stackID,
		"cardId":  cardID,
	}
	if err := c.GetJSON(
		deckAPIv11+"/boards/{boardId}/stacks/{stackId}/cards/{cardId}/attachments",
		params,
		&attachments,
	); err != nil {
		return nil, err
	}

	return attachments, nil
}

func (c *Client) DownloadAttachment(
	boardID, stackID, cardID int,
	a Attachment,
) ([]byte, error) {
	if a.Type == "file" {
		return c.downloadAttachmentViaWebDAV(a)
	}

	return c.downloadAttachmentViaDeckAPI(boardID, stackID, cardID, a.ID)
}

func (c *Client) downloadAttachmentViaWebDAV(a Attachment) ([]byte, error) {
	filePath, _ := a.ExtendedData["path"].(string)
	if filePath == "" {
		return nil, fmt.Errorf(
			"attachment %d: Nextcloud Files path not found (extendedData.path)",
			a.ID,
		)
	}

	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, "remote.php/dav/files", c.Username, filePath)

	resp, err := c.R().Get(u.String())
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf(
			"WebDAV download of attachment %d (%s) failed: %s",
			a.ID,
			filePath,
			resp.Status(),
		)
	}

	return resp.Body(), nil
}

func (c *Client) downloadAttachmentViaDeckAPI(
	boardID, stackID, cardID, attachmentID int,
) ([]byte, error) {
	params := pathParams{
		"boardId":      boardID,
		"stackId":      stackID,
		"cardId":       cardID,
		"attachmentId": attachmentID,
	}

	resp, err := c.OCSRequest(params).
		Get(
			deckAPIv10 + "/boards/{boardId}/stacks/{stackId}/cards/{cardId}/attachments/{attachmentId}",
		)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf(
			"download of attachment %d failed: %s",
			attachmentID,
			resp.Status(),
		)
	}

	return resp.Body(), nil
}
