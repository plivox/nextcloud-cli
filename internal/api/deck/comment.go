package deck

import "net/http"

const commentsPath = "/ocs/v2.php/apps/deck/api/v1.0/cards/{cardId}/comments"

type Comment struct {
	ID               int    `json:"id"`
	CardID           int    `json:"objectId"`
	Message          string `json:"message"`
	ActorID          string `json:"actorId"`
	ActorDisplayName string `json:"actorDisplayName"`
	CreationDateTime string `json:"creationDateTime"`
}

func (c *Client) ListComments(cardID int) ([]Comment, error) {
	var resp struct {
		OCS struct {
			Data []Comment `json:"data"`
		} `json:"ocs"`
	}

	err := c.GetJSON(
		commentsPath+"?limit=100",
		pathParams{"cardId": cardID},
		&resp,
	)

	return resp.OCS.Data, err
}

func (c *Client) AddComment(cardID int, message string) (Comment, error) {
	var resp struct {
		OCS struct {
			Data Comment `json:"data"`
		} `json:"ocs"`
	}

	err := c.SendJSON(http.MethodPost, commentsPath,
		pathParams{"cardId": cardID},
		map[string]interface{}{"message": message}, &resp)

	return resp.OCS.Data, err
}
