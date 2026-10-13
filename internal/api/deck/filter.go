package deck

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

var cardStatuses = map[string]string{
	"in-progress": "In progress",
	"overdue":     "Overdue",
	"done":        "Done",
	"archived":    "Archived",
}

type CardFilter struct {
	Statuses  []string
	Stacks    []string
	Assignees []string
}

func StatusKeys() []string {
	keys := make([]string, 0, len(cardStatuses))
	for k := range cardStatuses {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
}

func (c *Client) ListCards(boardID int, filter CardFilter) ([]Stack, error) {
	statuses := make(map[string]bool, len(filter.Statuses))
	for _, key := range filter.Statuses {
		status, ok := cardStatuses[strings.ToLower(key)]
		if !ok {
			return nil, fmt.Errorf(
				"unknown status: %q (available: %s)",
				key,
				strings.Join(StatusKeys(), ", "),
			)
		}

		statuses[status] = true
	}

	stacks, err := c.ListStacks(boardID)
	if err != nil {
		return nil, err
	}

	if statuses["Archived"] {
		archived, err := c.ListArchivedStacks(boardID)
		if err != nil {
			return nil, err
		}

		stacks = mergeStacks(stacks, archived)
	}

	stacks, err = filterStacks(stacks, filter.Stacks)
	if err != nil {
		return nil, err
	}

	if len(statuses) > 0 {
		stacks = filterCards(stacks, func(card Card) bool {
			return statuses[card.Status()]
		})
	}

	if len(filter.Assignees) > 0 {
		stacks = filterCards(stacks, func(card Card) bool {
			return assignedTo(card, filter.Assignees, c.Username)
		})
	}

	return stacks, nil
}

func ResolveStack(stacks []Stack, ref string) (Stack, error) {
	for _, s := range stacks {
		if strconv.Itoa(s.ID) == ref || strings.EqualFold(s.Title, ref) {
			return s, nil
		}
	}

	titles := make([]string, len(stacks))
	for i, s := range stacks {
		titles[i] = fmt.Sprintf("[%d] %s", s.ID, s.Title)
	}

	return Stack{}, fmt.Errorf(
		"unknown stack: %q (available: %s)",
		ref,
		strings.Join(titles, ", "),
	)
}

func mergeStacks(stacks, extra []Stack) []Stack {
	index := make(map[int]int, len(stacks))
	for i, s := range stacks {
		index[s.ID] = i
	}

	for _, s := range extra {
		if i, ok := index[s.ID]; ok {
			stacks[i].Cards = append(stacks[i].Cards, s.Cards...)
		} else {
			index[s.ID] = len(stacks)
			stacks = append(stacks, s)
		}
	}

	return stacks
}

func filterStacks(stacks []Stack, refs []string) ([]Stack, error) {
	if len(refs) == 0 {
		return stacks, nil
	}

	kept := make([]Stack, 0, len(refs))

	for _, ref := range refs {
		s, err := ResolveStack(stacks, ref)
		if err != nil {
			return nil, err
		}

		kept = append(kept, s)
	}

	return kept, nil
}

func filterCards(stacks []Stack, keep func(Card) bool) []Stack {
	for i, s := range stacks {
		var cards []Card

		for _, card := range s.Cards {
			if keep(card) {
				cards = append(cards, card)
			}
		}

		stacks[i].Cards = cards
	}

	return stacks
}

func assignedTo(card Card, filters []string, me string) bool {
	for _, f := range filters {
		f = strings.ToLower(f)
		if f == "none" && len(card.AssignedUsers) == 0 {
			return true
		}

		if f == "me" {
			f = strings.ToLower(me)
		}

		for _, a := range card.AssignedUsers {
			if strings.ToLower(a.Participant.UID) == f ||
				strings.Contains(
					strings.ToLower(a.Participant.DisplayName),
					f,
				) {
				return true
			}
		}
	}

	return false
}

func ResolveLabel(labels []Label, ref string) (Label, error) {
	for _, l := range labels {
		if strconv.Itoa(l.ID) == ref || strings.EqualFold(l.Title, ref) {
			return l, nil
		}
	}

	titles := make([]string, len(labels))
	for i, l := range labels {
		titles[i] = fmt.Sprintf("[%d] %s", l.ID, l.Title)
	}

	return Label{}, fmt.Errorf(
		"unknown label: %q (available: %s)",
		ref,
		strings.Join(titles, ", "),
	)
}

func ParseDueDate(s string) (*time.Time, error) {
	for _, layout := range []string{time.RFC3339, "2006-01-02 15:04", "2006-01-02"} {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return &t, nil
		}
	}

	return nil, fmt.Errorf(
		"invalid due date %q (use YYYY-MM-DD, \"YYYY-MM-DD HH:MM\" or RFC 3339)",
		s,
	)
}
