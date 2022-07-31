package component

import (
	"komainu/interactions/response"
	"komainu/storage"
	"log"
	"strings"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
)

type Handler struct {
	Code HandlerFunction
}

type HandlerFunction func(
	state *state.State,
	kvs storage.KeyValueStore,
	event *gateway.InteractionCreateEvent,
	interaction discord.ComponentInteraction,
) api.InteractionResponse

var registrations = map[string]Handler{}

// Register sets what function should handle interactions on the given message.
func Register(identifier string, handler Handler) {
	registrations[identifier] = handler
}

// AddHandler adds the component interaction handler to the given state
func AddHandler(state *state.State, kvs storage.KeyValueStore) {
	state.AddHandler(func(e *gateway.InteractionCreateEvent) {
		if interaction, ok := e.Data.(discord.ComponentInteraction); ok {
			target := strings.SplitN(string(interaction.ID()), "/", 2)[0]

			if handler, ok := registrations[target]; ok {
				resp := handler.Code(state, kvs, e, interaction)
				if err := state.RespondInteraction(e.ID, e.Token, resp); err != nil {
					log.Printf("[%s] Failed to send component interaction response: %s", e.GuildID, err)
				}
			} else {
				log.Printf("[%s] Got a %q component interaction, but there is no registered handler!", e.GuildID, target)
				if err := state.RespondInteraction(e.ID, e.Token, response.Ephemeral("Something odd happened. It has been logged.")); err != nil {
					log.Printf("[%s] ...and there was an error informing the user: %s", e.GuildID, err)
				}
			}
		}
	})
}
