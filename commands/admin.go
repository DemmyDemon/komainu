package commands

import (
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
)

func CommandGrant(state *state.State, event *gateway.InteractionCreateEvent, command *discord.CommandInteraction) api.InteractionResponse {
	return ResponseMessage("Not implemented")
}
func CommandRevoke(state *state.State, event *gateway.InteractionCreateEvent, command *discord.CommandInteraction) api.InteractionResponse {
	return ResponseMessage("Not implemented")
}
func CommandAccess(state *state.State, event *gateway.InteractionCreateEvent, command *discord.CommandInteraction) api.InteractionResponse {
	return ResponseMessage("Not implemented")
}
