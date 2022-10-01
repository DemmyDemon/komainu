//go:build ignore

package interactions

import (
	"komainu/storage"
	"log"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
)

// DeleteRole will delete the role selection settings if the role selction message is removed
func DeleteRole(state *state.State, kvs storage.KeyValueStore, e *gateway.MessageDeleteEvent) {
	if e.GuildID == discord.NullGuildID {
		return
	}
	err := kvs.Delete(e.GuildID, "roleselect", e.ID)
	if err != nil {
		log.Printf("[%s] Encountered an error removing role select from KVS after message deletion: %s\n", e.GuildID, err)
	}
}

/*
func ComponentRoleSelect(state *state.State, kvs storage.KeyValueStore, e *gateway.InteractionCreateEvent, interaction discord.ComponentInteraction) api.InteractionResponse {
	isRole, resp, err := storage.HandleInteractionAsRoleSelect(state, kvs, e, interaction)
	if err != nil {

	}
}
*/
