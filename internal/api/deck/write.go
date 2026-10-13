package deck

import (
	"net/http"
	"time"
)

const (
	boardPath = deckAPIv10 + "/boards/{boardId}"
	stackPath = boardPath + "/stacks/{stackId}"
	cardPath  = stackPath + "/cards/{cardId}"
)

type BoardUpdate struct {
	Title    *string
	Color    *string
	Archived *bool
}

type StackUpdate struct {
	Title *string
	Order *int
}

type NewCard struct {
	Title       string
	Description string
	DueDate     *time.Time
}

type CardUpdate struct {
	Title        *string
	Description  *string
	DueDate      *time.Time
	ClearDueDate bool
	Done         *bool
	Archived     *bool
}

func (c *Client) CreateBoard(title, color string) (Board, error) {
	var board Board

	err := c.SendJSON(http.MethodPost, deckAPIv10+"/boards", nil,
		map[string]interface{}{"title": title, "color": color}, &board)

	return board, err
}

func (c *Client) UpdateBoard(boardID int, u BoardUpdate) (Board, error) {
	board, err := c.GetBoard(boardID)
	if err != nil {
		return Board{}, err
	}

	if u.Title != nil {
		board.Title = *u.Title
	}

	if u.Color != nil {
		board.Color = *u.Color
	}

	if u.Archived != nil {
		board.Archived = *u.Archived
	}

	var updated Board

	err = c.SendJSON(http.MethodPut, boardPath,
		pathParams{"boardId": boardID},
		map[string]interface{}{
			"title":    board.Title,
			"color":    board.Color,
			"archived": board.Archived,
		}, &updated)

	return updated, err
}

func (c *Client) DeleteBoard(boardID int) error {
	return c.SendJSON(http.MethodDelete, boardPath,
		pathParams{"boardId": boardID}, nil, nil)
}

func (c *Client) CreateStack(
	boardID int,
	title string,
	order int,
) (Stack, error) {
	var stack Stack

	err := c.SendJSON(http.MethodPost, boardPath+"/stacks",
		pathParams{"boardId": boardID},
		map[string]interface{}{"title": title, "order": order}, &stack)

	return stack, err
}

func (c *Client) UpdateStack(
	boardID, stackID int,
	u StackUpdate,
) (Stack, error) {
	stack, err := c.GetStack(boardID, stackID)
	if err != nil {
		return Stack{}, err
	}

	if u.Title != nil {
		stack.Title = *u.Title
	}

	if u.Order != nil {
		stack.Order = *u.Order
	}

	var updated Stack

	err = c.SendJSON(http.MethodPut, stackPath,
		pathParams{"boardId": boardID, "stackId": stackID},
		map[string]interface{}{"title": stack.Title, "order": stack.Order},
		&updated)

	return updated, err
}

func (c *Client) DeleteStack(boardID, stackID int) error {
	return c.SendJSON(http.MethodDelete, stackPath,
		pathParams{"boardId": boardID, "stackId": stackID}, nil, nil)
}

func (c *Client) GetCard(boardID, stackID, cardID int) (Card, error) {
	var card Card

	err := c.GetJSON(cardPath,
		pathParams{"boardId": boardID, "stackId": stackID, "cardId": cardID},
		&card)

	return card, err
}

func (c *Client) CreateCard(boardID, stackID int, n NewCard) (Card, error) {
	body := map[string]interface{}{
		"title":       n.Title,
		"type":        "plain",
		"order":       999,
		"description": n.Description,
	}
	if n.DueDate != nil {
		body["duedate"] = n.DueDate.UTC().Format(time.RFC3339)
	}

	var card Card

	err := c.SendJSON(http.MethodPost, stackPath+"/cards",
		pathParams{"boardId": boardID, "stackId": stackID}, body, &card)

	return card, err
}

func (c *Client) UpdateCard(
	boardID, stackID, cardID int,
	u CardUpdate,
) (Card, error) {
	card, err := c.GetCard(boardID, stackID, cardID)
	if err != nil {
		return Card{}, err
	}

	if card.Type == "" {
		card.Type = "plain"
	}

	if u.Title != nil {
		card.Title = *u.Title
	}

	if u.Description != nil {
		card.Description = *u.Description
	}

	if u.DueDate != nil {
		due := u.DueDate.UTC().Format(time.RFC3339)
		card.DueDate = &due
	}

	if u.ClearDueDate {
		card.DueDate = nil
	}

	if u.Done != nil {
		card.Done = nil

		if *u.Done {
			now := time.Now().UTC().Format(time.RFC3339)
			card.Done = &now
		}
	}

	if u.Archived != nil {
		card.Archived = *u.Archived
	}

	var updated Card

	err = c.SendJSON(http.MethodPut, cardPath,
		pathParams{"boardId": boardID, "stackId": stackID, "cardId": cardID},
		map[string]interface{}{
			"title":       card.Title,
			"description": card.Description,
			"type":        card.Type,
			"owner":       card.Owner.UID,
			"order":       card.Order,
			"duedate":     card.DueDate,
			"archived":    card.Archived,
			"done":        card.Done,
		}, &updated)

	return updated, err
}

func (c *Client) DeleteCard(boardID, stackID, cardID int) error {
	return c.SendJSON(http.MethodDelete, cardPath,
		pathParams{"boardId": boardID, "stackId": stackID, "cardId": cardID},
		nil, nil)
}

func (c *Client) MoveCard(
	boardID, stackID, cardID, toStackID, order int,
) error {
	return c.SendJSON(http.MethodPut, cardPath+"/reorder",
		pathParams{"boardId": boardID, "stackId": toStackID, "cardId": cardID},
		map[string]interface{}{"stackId": toStackID, "order": order}, nil)
}

func (c *Client) cardAction(
	boardID, stackID, cardID int,
	action string,
	body map[string]interface{},
) error {
	return c.SendJSON(http.MethodPut, cardPath+"/"+action,
		pathParams{"boardId": boardID, "stackId": stackID, "cardId": cardID},
		body, nil)
}

func (c *Client) AssignUser(boardID, stackID, cardID int, userID string) error {
	return c.cardAction(boardID, stackID, cardID, "assignUser",
		map[string]interface{}{"userId": userID})
}

func (c *Client) UnassignUser(
	boardID, stackID, cardID int,
	userID string,
) error {
	return c.cardAction(boardID, stackID, cardID, "unassignUser",
		map[string]interface{}{"userId": userID})
}

func (c *Client) AssignLabel(boardID, stackID, cardID, labelID int) error {
	return c.cardAction(boardID, stackID, cardID, "assignLabel",
		map[string]interface{}{"labelId": labelID})
}

func (c *Client) RemoveLabel(boardID, stackID, cardID, labelID int) error {
	return c.cardAction(boardID, stackID, cardID, "removeLabel",
		map[string]interface{}{"labelId": labelID})
}
