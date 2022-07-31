package bot

import (
	"context"
	_ "komainu/interactions" // To make all the interactions init()
	"komainu/interactions/autocomplete"
	"komainu/interactions/command"
	"komainu/interactions/component"
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

	// TODO: This is very out of place here. We need a interactions/message package, and for the seen interaction file to register there.
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

	// TODO: This is very out of place here. We need a interactions/delete package, I guess.
	state.AddHandler(func(e *gateway.MessageDeleteEvent) {
		if e.GuildID == discord.NullGuildID {
			return
		}
		_, err := kvs.Delete(e.GuildID, "votes", e.ID)
		if err != nil {
			log.Printf("[%s] Encountered an error removing vote from KVS after message deletion: %s\n", e.GuildID, err)
		}
	})

	// TODO: interactions/guildcreate package and register there?
	// I mean, this'll be moot once everyone is running this version anyway.
	state.AddHandler(func(e *gateway.GuildCreateEvent) {
		command.ClearObsoleteCommands(state, e.ID)
	})

	command.AddHandler(state, kvs)
	autocomplete.AddHandler(state, kvs)
	modal.AddHandler(state, kvs)
	component.AddHandler(state, kvs)

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

	// TODO: Move command registration here, no need to register *per guild*.
	// That was originally supposed to be "some commands only work in the guild I use for testing" stuff,
	// but that turned out to be a bad idea, and now I'm stuck with per-guild registration. Dumb.
	// (For now, this is a stub function that does nothing but list some rubbish)
	if err := command.RegisterCommands(state); err != nil {
		log.Fatalf("Error during command registration: %s", err)
	}

	// TODO: Maybe move these to init() in the relevant packages?
	go storage.StartClosingExpiredVotes(state, kvs)
	go storage.StartRevokingActiveRole(state, kvs)

	return state
}
