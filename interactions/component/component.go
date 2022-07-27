package component

import (
	"komainu/interactions/response"
	"komainu/storage"
	"log"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
)

// TODO: Re-d the whole "vote" thing so more than one component interaction is possible/practical

// AddHandler adds the component interaction handler to the given state
func AddHandler(state *state.State, kvs storage.KeyValueStore) {
	state.AddHandler(func(e *gateway.InteractionCreateEvent) {
		if interaction, ok := e.Data.(discord.ComponentInteraction); ok {
			isVote, resp, err := storage.HandleInteractionAsVote(state, kvs, e, interaction)
			if err != nil {
				log.Printf("[%s] error while trying to handle an interaction as a vote: %s\n", e.GuildID, err)
				return
			}
			if isVote {
				if resp != "" {
					if err := state.RespondInteraction(e.ID, e.Token, response.Ephemeral(resp)); err != nil {
						log.Printf("[%s] Failed to send component interaction ephemeral response: %s\n", e.GuildID, err)
					}
				}
				return
			}
		}
	})
}
