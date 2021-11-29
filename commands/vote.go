package commands

import (
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
)

// CommandVote processes a command to start a vote
func CommandVote(state *state.State, event *gateway.InteractionCreateEvent, command *discord.CommandInteraction) api.InteractionResponse {
	return ResponseMessage("Not implemented")
}
