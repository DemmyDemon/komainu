package commands

import (
	"komainu/storage"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
)

// CommandAccess processes a command to list access entries.
func CommandAccess(state *state.State, sniper storage.KeyValueStore, event *gateway.InteractionCreateEvent, command *discord.CommandInteraction) api.InteractionResponse {
	return ResponseMessage("Not implemented")
}

// SubCommandAccessGrant processes a command to grant access.
func SubCommandAccessGrant() api.InteractionResponse {
	return ResponseMessage("Not implemented")
}

// SubCommandAccessRevoke processes a command to revoke access.
func SubCommandAccessRevoke() api.InteractionResponse {
	return ResponseMessage("Not implemented")
}

// CommandRevoke processes a command to revoke access.
func SubCommandAccessList() api.InteractionResponse {
	return ResponseMessage("Not implemented")
}
