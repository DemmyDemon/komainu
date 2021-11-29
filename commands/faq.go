package commands

import (
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
)

// CommandFaq processes a command to retrieve a FAQ item.
func CommandFaq(state *state.State, event *gateway.InteractionCreateEvent, command *discord.CommandInteraction) api.InteractionResponse {
	return ResponseMessage("Not implemented")
}

// CommandFaqOn processes a command to store a FAQ item.
func CommandFaqOn(state *state.State, event *gateway.InteractionCreateEvent, command *discord.CommandInteraction) api.InteractionResponse {
	return ResponseMessage("Not implemented")
}

// CommandFaqOff processes a command to remove a FAQ item.
func CommandFaqOff(state *state.State, event *gateway.InteractionCreateEvent, command *discord.CommandInteraction) api.InteractionResponse {
	return ResponseMessage("Not implemented")
}

// CommandFaqList processes a command to list all FAQ items.
func CommandFaqList(state *state.State, event *gateway.InteractionCreateEvent, command *discord.CommandInteraction) api.InteractionResponse {
	return ResponseMessage("Not implemented")
}
