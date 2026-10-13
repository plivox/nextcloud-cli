package deck

type Owner struct {
	UID         string `json:"uid"`
	DisplayName string `json:"displayname"`
}

type Board struct {
	ID       int     `json:"id"`
	Title    string  `json:"title"`
	Owner    Owner   `json:"owner"`
	Color    string  `json:"color"`
	Archived bool    `json:"archived"`
	Labels   []Label `json:"labels,omitempty"`
}

func (c *Client) ListBoards() ([]Board, error) {
	var boards []Board

	if err := c.GetJSON(deckAPIv10+"/boards", nil, &boards); err != nil {
		return nil, err
	}

	return boards, nil
}

func (c *Client) GetBoard(boardID int) (Board, error) {
	var board Board

	if err := c.GetJSON(
		deckAPIv10+"/boards/{boardId}",
		pathParams{"boardId": boardID},
		&board,
	); err != nil {
		return Board{}, err
	}

	return board, nil
}
