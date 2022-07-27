package bot

import (
	"context"
	_ "komainu/interactions" // To make all the interactions init()
	"komainu/interactions/autocomplete"
	"komainu/interactions/command"
	"komainu/interactions/modal"
	"komainu/storage"
	"log"
	"os"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
)

// Connect connects to Discord
func Connect(cfg *storage.Configuration, kvs storage.KeyValueStore) *state.State {
	var token = os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatalln("No BOT_TOKEN found in environment variables.")
	}

	state := state.New("Bot " + token)

	// TODO: Break up this function!

	state.AddHandler(func(e *gateway.MessageCreateEvent) {
		if e.GuildID == 0 {
			return // It's either a private message, or an ephemeral-response command. Doesn't count.
		}

		if err := storage.See(kvs, e.GuildID, e.Author.ID); err != nil {
			log.Printf("Seen in %d: %d sent a message in %s, BUT WAS NOT RECORDED:%s\n", e.GuildID, e.Author.ID, e.ChannelID, err)
		} else {
			log.Printf("Seen in %d: %d sent a message in %s\n", e.GuildID, e.Author.ID, e.ChannelID)
			if err := storage.MaybeGiveActiveRole(kvs, state, e.GuildID, e.Member); err != nil {
				log.Printf("[%s] Failed to MaybeGiveActiveRole to %s: %s", e.GuildID, e.Author.ID, err)
			}
		}
	})

	state.AddHandler(func(e *gateway.MessageDeleteEvent) {
		if e.GuildID == discord.NullGuildID {
			return
		}
		_, err := kvs.Delete(e.GuildID, "votes", e.ID)
		if err != nil {
			log.Printf("[%s] Encountered an error removing vote from KVS after message deletion: %s\n", e.GuildID, err)
		}
	})

	command.AddHandler(state, kvs)
	autocomplete.AddHandler(state, kvs)
	modal.AddHandler(state, kvs)

	state.AddIntents(gateway.IntentGuilds |
		gateway.IntentGuildMembers |
		gateway.IntentGuildBans |
		gateway.IntentGuildEmojis |
		gateway.IntentGuildIntegrations |
		gateway.IntentGuildInvites |
		gateway.IntentGuildMessages |
		gateway.IntentGuildMessageReactions |
		gateway.IntentGuildMessageTyping |
		gateway.IntentDirectMessages |
		gateway.IntentDirectMessageReactions |
		gateway.IntentDirectMessageTyping)

	if err := state.Open(context.Background()); err != nil {
		log.Fatalln("Failed to connect to Discord:", err)
	}

	user, err := state.Me()
	if err != nil {
		log.Fatalln("Failed to get myself:", err)
	}
	log.Printf("Connected to Discord as %s#%s\n", user.Username, user.Discriminator)

	go storage.StartClosingExpiredVotes(state, kvs)
	go storage.StartRevokingActiveRole(state, kvs)

	return state
}
