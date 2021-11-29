package commands

import (
	"komainu/storage"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
)

// CommandGrant processes a command to grant access.
func CommandGrant(state *state.State, sniper storage.KeyValueStore, event *gateway.InteractionCreateEvent, command *discord.CommandInteraction) api.InteractionResponse {
	return ResponseMessage("Not implemented")
}

// CommandRevoke processes a command to revoke access.
func CommandRevoke(state *state.State, sniper storage.KeyValueStore, event *gateway.InteractionCreateEvent, command *discord.CommandInteraction) api.InteractionResponse {
	return ResponseMessage("Not implemented")
}

// CommandAccess processes a command to list access entries.
func CommandAccess(state *state.State, sniper storage.KeyValueStore, event *gateway.InteractionCreateEvent, command *discord.CommandInteraction) api.InteractionResponse {
	return ResponseMessage("Not implemented")
}
