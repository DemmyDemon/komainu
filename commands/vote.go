package commands

import (
	"komainu/storage"
	"log"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
)

// CommandVote processes a command to start a vote
func CommandVote(state *state.State, sniper storage.KeyValueStore, event *gateway.InteractionCreateEvent, command *discord.CommandInteraction) api.InteractionResponse {
	if command.Options == nil || len(command.Options) != 4 {
		log.Printf("[%s] /vote command structure is somehow nil or not the correct number of elements. Wat.\n", event.GuildID)
		return ResponseMessage("Yeah, no, that didn't work.")
	}
	return ResponseMessage("Sorry, this is still being worked on!")
}
