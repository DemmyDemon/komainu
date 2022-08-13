//go:build ignore

// TODO: Flesh out the report function. Maybe.
package interactions

import (
	"komainu/interactions/command"
	"komainu/interactions/modal"
	"komainu/storage"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
)

func init() {
	command.Register("Report message", command.Handler{
		//Description: "Bring message to moderator attention",
		Type: discord.MessageCommand,
		Code: CommandReport,
	})
	// TODO: Write and register a handler for a modal response here.
}

func CommandReport(state *state.State, kvs storage.KeyValueStore, event *gateway.InteractionCreateEvent, cmd *discord.CommandInteraction) command.Response {
	return modal.Respond() // blah blah, take a report message.
}
