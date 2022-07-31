package command

import (
	"komainu/interactions/response"
	"komainu/storage"
	"komainu/utility"
	"log"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
)

type Command func(
	state *state.State,
	kvs storage.KeyValueStore,
	event *gateway.InteractionCreateEvent,
	command *discord.CommandInteraction,
) Response

type Response struct {
	Response api.InteractionResponse
	Callback func(message *discord.Message)
}

// IsEphemeral checks if the contained InteractionResponse is only shown to the user initiating the interaction.
func (cr *Response) IsEphemeral() bool {
	return cr.Response.Data.Flags&api.EphemeralResponse != 0
}

// Length returns the number of runes in the content string. This is not the same as the number of bytes!
func (cr *Response) Length() int {
	if cr.Response.Data.Content == nil {
		return 0
	}
	runes := []rune(cr.Response.Data.Content.Val)
	return len(runes)
}

type Handler struct {
	Description string
	Code        Command
	Options     []discord.CommandOption
}

// commands holds the Commands to be registered with each joined guild.
var commands = map[string]Handler{}

// The Token Bins. 5 and 10 are arbituary numbers, and it decrements at 10 second intervals.
var userTokenBin = &utility.TokenBin{Max: 5, Interval: 10}
var channelTokenBin = &utility.TokenBin{Max: 10, Interval: 10}

func Register(name string, command Handler) {
	commands[name] = command
}

// AddHandler adds handler for commands, but also the GuildCreate event for command registration.
func AddHandler(state *state.State, kvs storage.KeyValueStore) {
	state.AddHandler(func(e *gateway.InteractionCreateEvent) {
		if interaction, ok := e.Data.(*discord.CommandInteraction); ok {
			if e.GuildID == discord.NullGuildID || e.Member == nil { // Command issued in private
				state.RespondInteraction(e.ID, e.Token, response.Ephemeral("I'm sorry, I do not respond to commands in private."))
				return
			}
			if !userTokenBin.Allocate(discord.Snowflake(e.GuildID), discord.Snowflake(e.Member.User.ID)) {
				if err := state.RespondInteraction(e.ID, e.Token, response.Ephemeral("You are using too many commands too quickly. Calm down.")); err != nil {
					log.Println("An error occured posting throttle warning emphemral response (user):", err)
				}
				return
			}
			if !channelTokenBin.Allocate(discord.Snowflake(e.GuildID), discord.Snowflake(e.ChannelID)) {
				if err := state.RespondInteraction(e.ID, e.Token, response.Ephemeral("Too many commands being processed in this channel right now. Please wait.")); err != nil {
					log.Println("An error occured posting throttle warning emphemral response (channel):", err)
				}
				return
			}

			if val, ok := commands[interaction.Name]; ok {
				resp := val.Code(state, kvs, e, interaction)

				if resp.Length() > 1500 {
					if resp.IsEphemeral() {
						resp.Response.Data.Content = option.NewNullableString(resp.Response.Data.Content.Val)
					}
				}

				if err := state.RespondInteraction(e.ID, e.Token, resp.Response); err != nil {
					log.Printf("[%s] Failed to send command interaction response: %s", e.GuildID, err)
				}
				if resp.Callback != nil {
					message, err := state.InteractionResponse(e.AppID, e.Token)
					if err != nil {
						log.Printf("Error %s getting message reference for %s command callback\n", err, interaction.Name)
						return
					}
					if message != nil && message.ID != discord.NullMessageID {
						resp.Callback(message)
					}
				}
			}
		}
	})
}

// RegisterCommands chews up the commands registered for the bot and actually registers them with Discord.
func RegisterCommands(state *state.State) error {
	app, err := state.CurrentApplication()
	if err != nil {
		return err
	}
	adminOnly := discord.NewPermissions(0)
	bulkCommands := []api.CreateCommandData{}
	for name, data := range commands {
		bulkCommands = append(bulkCommands, api.CreateCommandData{
			Name:                     name,
			Description:              data.Description,
			Options:                  data.Options,
			DefaultMemberPermissions: adminOnly,
		})
	}
	registered, err := state.BulkOverwriteCommands(app.ID, bulkCommands)
	if err != nil {
		return err
	}
	log.Printf("%d commands successfully registered", len(registered))
	return nil
}

// ClearObsoleteCommands removes the old, obsolete, per-guild-registered commands.
func ClearObsoleteCommands(state *state.State, guildID discord.GuildID) {
	app, err := state.CurrentApplication()
	if err != nil {
		log.Println("Failed to clear obsolete commands: Could not determine app ID:", err)
		return
	}

	currentCommands, err := state.GuildCommands(app.ID, guildID)
	if err != nil {
		log.Printf("[%s] Failed to clear obsolete commands: Could not determine current guild commands:%s\n", guildID, err)
		return
	}
	for _, command := range currentCommands {
		if command.AppID == app.ID {
			if err := state.DeleteGuildCommand(app.ID, guildID, command.ID); err != nil {
				log.Printf("[%s] Tried to remove obsolete Guild command /%s, but %s\n", guildID, command.Name, err)
			} else {
				log.Printf("[%s] Successfully removed obsolete Guild command /%s\n", guildID, command.Name)
			}
		}
	}
}
