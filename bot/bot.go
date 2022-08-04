package bot

import (
	"context"
	_ "komainu/interactions" // To make all the interactions init()
	"komainu/interactions/autocomplete"
	"komainu/interactions/command"
	"komainu/interactions/component"
	"komainu/interactions/delete"
	"komainu/interactions/message"
	"komainu/interactions/modal"
	"komainu/storage"
	"log"
	"os"

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

	addBoatloadOfIntents(state)

	// TODO: interactions/guildcreate package and register there?
	// I mean, this'll be moot once everyone is running this version anyway.
	state.AddHandler(func(e *gateway.GuildCreateEvent) {
		command.ClearObsoleteCommands(state, e.ID)
	})

	command.AddHandler(state, kvs)
	autocomplete.AddHandler(state, kvs)
	modal.AddHandler(state, kvs)
	component.AddHandler(state, kvs)
	message.AddHandler(state, kvs)
	delete.AddHandler(state, kvs)

	if err := state.Open(context.Background()); err != nil {
		log.Fatalln("Failed to connect to Discord:", err)
	}

	user, err := state.Me()
	if err != nil {
		log.Fatalln("Failed to get myself:", err)
	}
	log.Printf("Connected to Discord as %s#%s\n", user.Username, user.Discriminator)

	if err := command.RegisterCommands(state); err != nil {
		log.Fatalf("Error during command registration: %s", err)
	}

	// TODO: Maybe move these to init() in the relevant packages?
	go storage.StartClosingExpiredVotes(state, kvs)
	go storage.StartRevokingActiveRole(state, kvs)

	return state
}

func addBoatloadOfIntents(state *state.State) {
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
}
