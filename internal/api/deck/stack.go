package deck

type Stack struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	BoardID int    `json:"boardId"`
	Order   int    `json:"order"`
	Cards   []Card `json:"cards"`
}

func (c *Client) ListStacks(boardID int) ([]Stack, error) {
	var stacks []Stack

	if err := c.GetJSON(
		deckAPIv10+"/boards/{boardId}/stacks",
		pathParams{"boardId": boardID},
		&stacks,
	); err != nil {
		return nil, err
	}

	return stacks, nil
}

func (c *Client) GetStack(boardID, stackID int) (Stack, error) {
	var stack Stack

	params := pathParams{"boardId": boardID, "stackId": stackID}
	if err := c.GetJSON(
		deckAPIv10+"/boards/{boardId}/stacks/{stackId}",
		params,
		&stack,
	); err != nil {
		return Stack{}, err
	}

	return stack, nil
}

func (c *Client) ListArchivedStacks(boardID int) ([]Stack, error) {
	var stacks []Stack

	if err := c.GetJSON(
		deckAPIv10+"/boards/{boardId}/stacks/archived",
		pathParams{"boardId": boardID},
		&stacks,
	); err != nil {
		return nil, err
	}

	return stacks, nil
}
