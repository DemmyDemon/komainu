package commands

import (
	"komainu/storage"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/json"
)

type AutocompleteHandler struct {
	code AutocompleteFunction
}

type AutocompleteFunction func(
	state *state.State,
	kvs storage.KeyValueStore,
	event *gateway.InteractionCreateEvent,
	interaction *discord.AutocompleteInteraction,
) api.AutocompleteChoices

var autocompleters = map[string]AutocompleteHandler{}

func registerAutocompleteHandlerObject(name string, handler AutocompleteHandler) {
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
