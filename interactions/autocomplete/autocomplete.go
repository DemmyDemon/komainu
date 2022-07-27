package autocomplete

import (
	"komainu/storage"
	"log"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/json"
)

type Handler struct {
	Code HandlerFunction
}

type HandlerFunction func(
	state *state.State,
	kvs storage.KeyValueStore,
	event *gateway.InteractionCreateEvent,
	interaction *discord.AutocompleteInteraction,
) api.AutocompleteChoices

var autocompleters = map[string]Handler{}

func Register(name string, handler Handler) {
	autocompleters[name] = handler
}

func GetAutocompleteValue(interaction *discord.AutocompleteInteraction) (bool, json.Raw) {
	for _, option := range interaction.Options {
		if option.Focused {
			return true, option.Value
		}
	}
	return false, json.Raw{}
}

// AddHandler adds the autocomplete handler to the given state
func AddHandler(state *state.State, kvs storage.KeyValueStore) {
	state.AddHandler(func(e *gateway.InteractionCreateEvent) {
		if interaction, ok := e.Data.(*discord.AutocompleteInteraction); ok {
			if val, ok := autocompleters[interaction.Name]; ok {
				response := val.Code(state, kvs, e, interaction)
				state.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
					Type: api.AutocompleteResult,
					Data: &api.InteractionResponseData{
						Choices: response,
					},
				})
			} else {
				state.RespondInteraction(e.ID, e.Token, api.InteractionResponse{
					Type: api.AutocompleteResult,
					Data: &api.InteractionResponseData{},
				})
				log.Printf("[%s] Unknown autocomplete target %q used", e.GuildID, interaction.Name)
			}
		}
	})
}
