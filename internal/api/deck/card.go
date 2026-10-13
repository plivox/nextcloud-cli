package deck

import (
	"encoding/json"
	"fmt"
	"time"
)

type Label struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Color string `json:"color"`
}

type Participant struct {
	UID         string `json:"uid"`
	DisplayName string `json:"displayname"`
}

func (p *Participant) UnmarshalJSON(data []byte) error {
	var uid string
	if json.Unmarshal(data, &uid) == nil {
		*p = Participant{UID: uid, DisplayName: uid}

		return nil
	}

	type plain Participant

	return json.Unmarshal(data, (*plain)(p))
}

type Assignment struct {
	ID          int         `json:"id"`
	Participant Participant `json:"participant"`
}

type Card struct {
	ID              int          `json:"id"`
	StackID         int          `json:"stackId"`
	Title           string       `json:"title"`
	Type            string       `json:"type"`
	Order           int          `json:"order"`
	Owner           Participant  `json:"owner"`
	Description     string       `json:"description"`
	DueDate         *string      `json:"duedate"`
	Done            *string      `json:"done"`
	Archived        bool         `json:"archived"`
	Labels          []Label      `json:"labels"`
	AssignedUsers   []Assignment `json:"assignedUsers"`
	AttachmentCount int          `json:"attachmentCount"`
	Attachments     []Attachment `json:"attachments,omitempty"`
}

func (c Card) Status() string {
	if c.Archived {
		return "Archived"
	}

	if c.Done != nil {
		return "Done"
	}

	if c.DueDate != nil {
		if due, err := time.Parse(
			time.RFC3339,
			*c.DueDate,
		); err == nil &&
			due.Before(time.Now()) {
			return "Overdue"
		}
	}

	return "In progress"
}

func (c *Client) FindCard(
	cardID int,
	boardID *int,
) (Card, Board, Stack, error) {
	search := func(board Board) (Card, Stack, bool, error) {
		for _, list := range []func(int) ([]Stack, error){
			c.ListStacks,
			c.ListArchivedStacks,
		} {
			stacks, err := list(board.ID)
			if err != nil {
				return Card{}, Stack{}, false, err
			}

			for _, stack := range stacks {
				for _, card := range stack.Cards {
					if card.ID == cardID {
						return card, stack, true, nil
					}
				}
			}
		}

		return Card{}, Stack{}, false, nil
	}

	if boardID != nil {
		board, err := c.GetBoard(*boardID)
		if err != nil {
			return Card{}, Board{}, Stack{}, err
		}

		card, stack, found, err := search(board)
		if err != nil {
			return Card{}, Board{}, Stack{}, err
		}

		if !found {
			return Card{}, Board{}, Stack{}, fmt.Errorf(
				"card %d not found in board %d",
				cardID,
				*boardID,
			)
		}

		return card, board, stack, nil
	}

	boards, err := c.ListBoards()
	if err != nil {
		return Card{}, Board{}, Stack{}, err
	}

	for _, board := range boards {
		card, stack, found, err := search(board)
		if err != nil {
			return Card{}, Board{}, Stack{}, err
		}

		if found {
			return card, board, stack, nil
		}
	}

	return Card{}, Board{}, Stack{}, fmt.Errorf(
		"card %d not found in any accessible board",
		cardID,
	)
}
