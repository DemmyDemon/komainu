package bot

import (
	"context"
	"komainu/commands"
	"komainu/storage"
	"log"
	"os"

	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
)

// Connect connects to Discord
func Connect(cfg *storage.Configuration) *state.State {
	var token = os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatalln("No BOT_TOKEN found in environment variables.")
	}

	// seen := storage.GetSeen()

	state, err := state.New("Bot " + token)
	if err != nil {
		log.Fatalln("Failed to create Discord state:", err)
	}

	state.AddHandler(func(e *gateway.MessageCreateEvent) {
		if err := storage.See(e.GuildID, e.Author.ID); err != nil {
			log.Printf("Seen in %d: %d sent a message in %s, BUT WAS NOT RECORDED:%s\n", e.GuildID, e.Author.ID, e.ChannelID, err)
		} else {
			log.Printf("Seen in %d: %d sent a message in %s\n", e.GuildID, e.Author.ID, e.ChannelID)
		}
	})

	commands.AddCommandHandler(state)

	state.AddHandler(func(e *gateway.GuildCreateEvent) {
		commands.RegisterCommands(state, e.ID)
	})

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

	return state
}
