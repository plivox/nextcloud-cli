package deckcmd

import (
	"github.com/plivox/nextcloud-cli/internal/api/deck"
	"github.com/plivox/nextcloud-cli/internal/auth"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "deck",
	Short: "Nextcloud Deck boards, stacks, cards and attachments",
}

func newClient() (*deck.Client, error) {
	c, err := auth.NewClient()
	if err != nil {
		return nil, err
	}

	return deck.New(c), nil
}
