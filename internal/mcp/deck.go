package mcp

import (
	"context"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/plivox/nextcloud-cli/internal/api/deck"
)

type cardSummary struct {
	ID        int      `json:"id"`
	Title     string   `json:"title"`
	StackID   int      `json:"stackId"`
	Stack     string   `json:"stack"`
	Status    string   `json:"status"`
	Assignees []string `json:"assignees,omitempty"`
	Labels    []string `json:"labels,omitempty"`
	DueDate   *string  `json:"dueDate,omitempty"`
	Done      *string  `json:"done,omitempty"`
}

func summarize(s deck.Stack, c deck.Card) cardSummary {
	sum := cardSummary{
		ID:      c.ID,
		Title:   c.Title,
		StackID: s.ID,
		Stack:   s.Title,
		Status:  c.Status(),
		DueDate: c.DueDate,
		Done:    c.Done,
	}

	for _, a := range c.AssignedUsers {
		sum.Assignees = append(sum.Assignees, a.Participant.UID)
	}

	for _, l := range c.Labels {
		sum.Labels = append(sum.Labels, l.Title)
	}

	return sum
}

type boardIDInput struct {
	BoardID int `json:"board_id" jsonschema:"ID of the board"`
}

type cardInput struct {
	CardID  int `json:"card_id"            jsonschema:"ID of the card"`
	BoardID int `json:"board_id,omitempty" jsonschema:"ID of the card's board, if known (faster lookup)"`
}

func (srv *server) findCard(
	in cardInput,
) (deck.Card, deck.Board, deck.Stack, error) {
	var boardID *int
	if in.BoardID != 0 {
		boardID = &in.BoardID
	}

	return srv.deck.FindCard(in.CardID, boardID)
}

func (srv *server) addDeckReadTools(s *mcp.Server) {
	addTool(
		s,
		"deck_list_boards",
		"List the Deck boards the user can access.",
		readOnly,
		func(_ context.Context, _ struct{}) (any, error) {
			boards, err := srv.deck.ListBoards()

			return map[string]any{"boards": boards}, err
		},
	)

	addTool(
		s,
		"deck_get_board",
		"Get a Deck board with its stacks (columns) and labels.",
		readOnly,
		func(_ context.Context, in boardIDInput) (any, error) {
			board, err := srv.deck.GetBoard(in.BoardID)
			if err != nil {
				return nil, err
			}

			stacks, err := srv.deck.ListStacks(in.BoardID)
			if err != nil {
				return nil, err
			}

			type stackSummary struct {
				ID    int    `json:"id"`
				Title string `json:"title"`
				Order int    `json:"order"`
				Cards int    `json:"cards"`
			}

			summaries := make([]stackSummary, len(stacks))
			for i, st := range stacks {
				summaries[i] = stackSummary{
					st.ID,
					st.Title,
					st.Order,
					len(st.Cards),
				}
			}

			return map[string]any{"board": board, "stacks": summaries}, nil
		},
	)

	type listCardsInput struct {
		BoardID   int      `json:"board_id"            jsonschema:"ID of the board"`
		Statuses  []string `json:"statuses,omitempty"  jsonschema:"Only cards with one of these statuses: in-progress, overdue, done, archived"`
		Stacks    []string `json:"stacks,omitempty"    jsonschema:"Only cards in these stacks, by ID or title"`
		Assignees []string `json:"assignees,omitempty" jsonschema:"Only cards assigned to one of these users (uid or part of the display name); 'me' for the current user, 'none' for unassigned cards"`
	}

	addTool(
		s,
		"deck_list_cards",
		"List the cards of a Deck board, optionally filtered by status, stack and assignee. Archived cards are only included when filtering on the 'archived' status.",
		readOnly,
		func(_ context.Context, in listCardsInput) (any, error) {
			stacks, err := srv.deck.ListCards(in.BoardID, deck.CardFilter{
				Statuses:  in.Statuses,
				Stacks:    in.Stacks,
				Assignees: in.Assignees,
			})
			if err != nil {
				return nil, err
			}

			cards := []cardSummary{}

			for _, st := range stacks {
				for _, c := range st.Cards {
					cards = append(cards, summarize(st, c))
				}
			}

			return map[string]any{"cards": cards}, nil
		},
	)

	addTool(
		s,
		"deck_get_card",
		"Get a Deck card's details, including its description and comments.",
		readOnly,
		func(_ context.Context, in cardInput) (any, error) {
			card, board, stack, err := srv.findCard(in)
			if err != nil {
				return nil, err
			}

			comments, err := srv.deck.ListComments(card.ID)
			if err != nil {
				return nil, err
			}

			return map[string]any{
				"card":   card,
				"status": card.Status(),
				"board": map[string]any{
					"id":    board.ID,
					"title": board.Title,
				},
				"stack": map[string]any{
					"id":    stack.ID,
					"title": stack.Title,
				},
				"comments": comments,
			}, nil
		},
	)
}

func (srv *server) addDeckWriteTools(s *mcp.Server) {
	type createBoardInput struct {
		Title string `json:"title"           jsonschema:"Title of the board"`
		Color string `json:"color,omitempty" jsonschema:"Hexadecimal color without #, e.g. 0082c9"`
	}

	addTool(s, "deck_create_board", "Create a Deck board.", additive,
		func(_ context.Context, in createBoardInput) (any, error) {
			if in.Color == "" {
				in.Color = "0082c9"
			}

			return srv.deck.CreateBoard(in.Title, in.Color)
		})

	type updateBoardInput struct {
		BoardID  int     `json:"board_id"           jsonschema:"ID of the board"`
		Title    *string `json:"title,omitempty"    jsonschema:"New title"`
		Color    *string `json:"color,omitempty"    jsonschema:"New hexadecimal color without #"`
		Archived *bool   `json:"archived,omitempty" jsonschema:"Archive or unarchive the board"`
	}

	addTool(
		s,
		"deck_update_board",
		"Update a Deck board's title, color or archived state.",
		additive,
		func(_ context.Context, in updateBoardInput) (any, error) {
			return srv.deck.UpdateBoard(in.BoardID, deck.BoardUpdate{
				Title: in.Title, Color: in.Color, Archived: in.Archived,
			})
		},
	)

	addTool(
		s,
		"deck_delete_board",
		"Delete a Deck board and everything in it.",
		destructive,
		func(_ context.Context, in boardIDInput) (any, error) {
			return map[string]any{
				"deleted": in.BoardID,
			}, srv.deck.DeleteBoard(
				in.BoardID,
			)
		},
	)

	type createStackInput struct {
		BoardID int    `json:"board_id"        jsonschema:"ID of the board"`
		Title   string `json:"title"           jsonschema:"Title of the stack (column)"`
		Order   *int   `json:"order,omitempty" jsonschema:"Position among the stacks"`
	}

	addTool(
		s,
		"deck_create_stack",
		"Create a stack (column) in a Deck board.",
		additive,
		func(_ context.Context, in createStackInput) (any, error) {
			order := 999
			if in.Order != nil {
				order = *in.Order
			}

			return srv.deck.CreateStack(in.BoardID, in.Title, order)
		},
	)

	type updateStackInput struct {
		BoardID int     `json:"board_id"        jsonschema:"ID of the board"`
		StackID int     `json:"stack_id"        jsonschema:"ID of the stack"`
		Title   *string `json:"title,omitempty" jsonschema:"New title"`
		Order   *int    `json:"order,omitempty" jsonschema:"New position among the stacks"`
	}

	addTool(
		s,
		"deck_update_stack",
		"Rename or reorder a stack (column).",
		additive,
		func(_ context.Context, in updateStackInput) (any, error) {
			return srv.deck.UpdateStack(
				in.BoardID,
				in.StackID,
				deck.StackUpdate{
					Title: in.Title, Order: in.Order,
				},
			)
		},
	)

	type stackInput struct {
		BoardID int `json:"board_id" jsonschema:"ID of the board"`
		StackID int `json:"stack_id" jsonschema:"ID of the stack"`
	}

	addTool(
		s,
		"deck_delete_stack",
		"Delete a stack (column) and its cards.",
		destructive,
		func(_ context.Context, in stackInput) (any, error) {
			return map[string]any{
				"deleted": in.StackID,
			}, srv.deck.DeleteStack(
				in.BoardID,
				in.StackID,
			)
		},
	)

	type createCardInput struct {
		BoardID     int    `json:"board_id"              jsonschema:"ID of the board"`
		Stack       string `json:"stack"                 jsonschema:"Stack (column) ID or title"`
		Title       string `json:"title"                 jsonschema:"Title of the card"`
		Description string `json:"description,omitempty" jsonschema:"Markdown description"`
		Due         string `json:"due,omitempty"         jsonschema:"Due date: YYYY-MM-DD, 'YYYY-MM-DD HH:MM' or RFC 3339"`
	}

	addTool(s, "deck_create_card", "Create a card in a Deck stack.", additive,
		func(_ context.Context, in createCardInput) (any, error) {
			stacks, err := srv.deck.ListStacks(in.BoardID)
			if err != nil {
				return nil, err
			}

			stack, err := deck.ResolveStack(stacks, in.Stack)
			if err != nil {
				return nil, err
			}

			n := deck.NewCard{Title: in.Title, Description: in.Description}
			if in.Due != "" {
				if n.DueDate, err = deck.ParseDueDate(in.Due); err != nil {
					return nil, err
				}
			}

			return srv.deck.CreateCard(in.BoardID, stack.ID, n)
		})

	type updateCardInput struct {
		CardID      int     `json:"card_id"               jsonschema:"ID of the card"`
		BoardID     int     `json:"board_id,omitempty"    jsonschema:"ID of the card's board, if known"`
		Title       *string `json:"title,omitempty"       jsonschema:"New title"`
		Description *string `json:"description,omitempty" jsonschema:"New Markdown description (replaces the current one)"`
		Due         *string `json:"due,omitempty"         jsonschema:"New due date (YYYY-MM-DD, ...), or 'none' to remove it"`
		Done        *bool   `json:"done,omitempty"        jsonschema:"Mark the card as done (true) or not done (false)"`
		Archived    *bool   `json:"archived,omitempty"    jsonschema:"Archive (true) or unarchive (false) the card"`
	}

	addTool(s, "deck_update_card",
		"Update a Deck card. Only the given fields change.", additive,
		func(_ context.Context, in updateCardInput) (any, error) {
			card, board, stack, err := srv.findCard(
				cardInput{in.CardID, in.BoardID},
			)
			if err != nil {
				return nil, err
			}

			u := deck.CardUpdate{
				Title: in.Title, Description: in.Description,
				Done: in.Done, Archived: in.Archived,
			}

			if in.Due != nil {
				if *in.Due == "" || strings.EqualFold(*in.Due, "none") {
					u.ClearDueDate = true
				} else if u.DueDate, err = deck.ParseDueDate(*in.Due); err != nil {
					return nil, err
				}
			}

			return srv.deck.UpdateCard(board.ID, stack.ID, card.ID, u)
		})

	type moveCardInput struct {
		CardID  int    `json:"card_id"            jsonschema:"ID of the card"`
		BoardID int    `json:"board_id,omitempty" jsonschema:"ID of the card's board, if known"`
		Stack   string `json:"stack"              jsonschema:"Target stack (column) ID or title"`
		Order   int    `json:"order,omitempty"    jsonschema:"Position in the target stack (0 = top)"`
	}

	addTool(
		s,
		"deck_move_card",
		"Move a Deck card to another stack (column) or position.",
		additive,
		func(_ context.Context, in moveCardInput) (any, error) {
			card, board, stack, err := srv.findCard(
				cardInput{in.CardID, in.BoardID},
			)
			if err != nil {
				return nil, err
			}

			stacks, err := srv.deck.ListStacks(board.ID)
			if err != nil {
				return nil, err
			}

			target, err := deck.ResolveStack(stacks, in.Stack)
			if err != nil {
				return nil, err
			}

			err = srv.deck.MoveCard(
				board.ID,
				stack.ID,
				card.ID,
				target.ID,
				in.Order,
			)

			return map[string]any{"card": card.ID, "stack": target.Title}, err
		},
	)

	addTool(s, "deck_delete_card", "Delete a Deck card.", destructive,
		func(_ context.Context, in cardInput) (any, error) {
			card, board, stack, err := srv.findCard(in)
			if err != nil {
				return nil, err
			}

			return map[string]any{
				"deleted": card.ID,
			}, srv.deck.DeleteCard(
				board.ID,
				stack.ID,
				card.ID,
			)
		})

	type userInput struct {
		CardID  int    `json:"card_id"            jsonschema:"ID of the card"`
		BoardID int    `json:"board_id,omitempty" jsonschema:"ID of the card's board, if known"`
		User    string `json:"user"               jsonschema:"User ID (uid); 'me' for the current user"`
	}

	userTool := func(name, description string, action func(*deck.Client, int, int, int, string) error) {
		addTool(s, name, description, additive,
			func(_ context.Context, in userInput) (any, error) {
				card, board, stack, err := srv.findCard(
					cardInput{in.CardID, in.BoardID},
				)
				if err != nil {
					return nil, err
				}

				if strings.EqualFold(in.User, "me") {
					in.User = srv.deck.Username
				}

				return map[string]any{"card": card.ID, "user": in.User},
					action(srv.deck, board.ID, stack.ID, card.ID, in.User)
			})
	}

	userTool(
		"deck_assign_user",
		"Assign a user to a Deck card.",
		(*deck.Client).AssignUser,
	)
	userTool(
		"deck_unassign_user",
		"Unassign a user from a Deck card.",
		(*deck.Client).UnassignUser,
	)

	type labelInput struct {
		CardID  int    `json:"card_id"            jsonschema:"ID of the card"`
		BoardID int    `json:"board_id,omitempty" jsonschema:"ID of the card's board, if known"`
		Label   string `json:"label"              jsonschema:"Label ID or title (see deck_get_board)"`
	}

	labelTool := func(name, description string, action func(*deck.Client, int, int, int, int) error) {
		addTool(s, name, description, additive,
			func(_ context.Context, in labelInput) (any, error) {
				card, board, stack, err := srv.findCard(
					cardInput{in.CardID, in.BoardID},
				)
				if err != nil {
					return nil, err
				}

				full, err := srv.deck.GetBoard(board.ID)
				if err != nil {
					return nil, err
				}

				label, err := deck.ResolveLabel(full.Labels, in.Label)
				if err != nil {
					return nil, err
				}

				return map[string]any{"card": card.ID, "label": label.Title},
					action(srv.deck, board.ID, stack.ID, card.ID, label.ID)
			})
	}

	labelTool(
		"deck_add_label",
		"Add a label to a Deck card.",
		(*deck.Client).AssignLabel,
	)
	labelTool(
		"deck_remove_label",
		"Remove a label from a Deck card.",
		(*deck.Client).RemoveLabel,
	)

	type commentInput struct {
		CardID  int    `json:"card_id" jsonschema:"ID of the card"`
		Message string `json:"message" jsonschema:"Comment text (max 1000 characters); @uid mentions a user"`
	}

	addTool(s, "deck_add_comment", "Add a comment to a Deck card.", additive,
		func(_ context.Context, in commentInput) (any, error) {
			return srv.deck.AddComment(in.CardID, in.Message)
		})
}
