package commands

import (
	"log"
	"strings"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
)

type CommandFunction func(
	state *state.State,
	event *gateway.InteractionCreateEvent,
	command *discord.CommandInteraction,
) api.InteractionResponse

type Command struct {
	group       string
	description string
	code        CommandFunction
	options     []discord.CommandOption
}

var commands = map[string]Command{
	// "grant":  {"admin", "Grant access to a command gruop to a role", CommandGrant, []discord.CommandOption{}},
	// "revoke": {"admin", "Revoke access to a command group for a role", CommandRevoke, []discord.CommandOption{}},
	// "access": {"admin", "See what roles have access to what command group", CommandAccess, []discord.CommandOption{}},

	"seen": {"seen", "Check when someone was last around", CommandSeen, []discord.CommandOption{
		&discord.UserOption{
			OptionName:  "user",
			Description: "The user to look up",
			Required:    true,
		},
	}},
	"inactive": {"seen", "Get a list of inactive people", CommandInactive, []discord.CommandOption{
		&discord.IntegerOption{
			OptionName:  "days",
			Description: "How many days of quiet makes someone inactive?",
			Required:    true,
		},
	}},

	// "faq":     {"faquser", "Look up a FAQ topic", CommandFaq, []discord.CommandOption{}},
	// "faqon":   {"faq", "Add a topic to the FAQ", CommandFaqOn, []discord.CommandOption{}},
	// "faqoff":  {"faq", "Remove a topic from the FAQ", CommandFaqOff, []discord.CommandOption{}},
	// "faqlist": {"faq", "List the FAQ topics", CommandFaqList, []discord.CommandOption{}},

	// "vote": {"vote", "Initiate a vote", CommandVote, []discord.CommandOption{}},
}

func HasAccess(state *state.State, guildID discord.GuildID, channelID discord.ChannelID, member *discord.Member, group string) bool {
	if member == nil {
		return false
	}

	// TODO: Check member.RoleIDs against the roles stored under group string in Sniper

	if guild, err := state.Guild(guildID); err != nil {
		log.Printf("Could not look up guild %s for access check: %s\n", guildID, err)
		return false // Better safe than sorry!
	} else if guild.OwnerID == member.User.ID {
		return true // Owner always has access to everything.
	}

	if permissions, err := state.Permissions(channelID, member.User.ID); err != nil {
		log.Printf("Could not look up permissions for %s in channel %s for access check: %s\n", member.User.ID, channelID, err)
		return false // Better safe than sorry!
	} else if permissions.Has(discord.PermissionAdministrator) {
		return true // Administrators get access to everyting
	}

	return false // If all else fails, they're not authorized.
}

func AddCommandHandler(state *state.State) {
	state.AddHandler(func(e *gateway.InteractionCreateEvent) {
		command, ok := e.Data.(*discord.CommandInteraction)
		if !ok {
			return
		}
		if val, ok := commands[command.Name]; ok {
			if !HasAccess(state, e.GuildID, e.ChannelID, e.Member, val.group) {
				if err := state.RespondInteraction(e.ID, e.Token, ResponseMessage("Sorry, access was denied.")); err != nil {
					log.Println("An error occured posting access denied response:", err)
				}
				return
			}

			response := val.code(state, e, command)

			if err := state.RespondInteraction(e.ID, e.Token, response); err != nil {
				log.Println("Failed to send interaction resposne:", err)
			}
		}
	})
}

func RegisterCommands(state *state.State, guildID discord.GuildID) {
	app, err := state.CurrentApplication()
	if err != nil {
		log.Println("Failed to register commands: Could not determine app ID:", err)
		return
	}

	currentCommands, err := state.GuildCommands(app.ID, guildID)
	if err != nil {
		log.Printf("[%s] Failed to register commands: Could not determine current guild commands:%s\n", guildID, err)
		return
	}
	for _, command := range currentCommands {
		if command.AppID == app.ID {
			if _, ok := commands[command.Name]; !ok {
				if err := state.DeleteGuildCommand(app.ID, guildID, command.ID); err != nil {
					log.Printf("[%s] Tried to remove obsolete command /%s, but %s\n", guildID, command.Name, err)
				}
			}
		}
	}

	for name, data := range commands {
		_, err := state.CreateGuildCommand(app.ID, guildID, api.CreateCommandData{
			Name:        name,
			Description: data.description,
			Options:     data.options,
		})
		if err != nil {
			log.Printf("[%s] Failed to create guild command /%s: %s\n", guildID, name, err)
		} else {
			log.Printf("[%s] Registered command /%s", guildID, name)
		}
	}
}

// ResponseMessage generates an InteractionResponse from the strings given.
func ResponseMessage(message ...string) api.InteractionResponse {
	return api.InteractionResponse{
		Type: api.MessageInteractionWithSource,
		Data: &api.InteractionResponseData{
			Content: option.NewNullableString(strings.Join(message, " ")),
		},
	}
}

// ResponseMessageNoMention generates an InteractionResponse from the strings given, and suppresses any mentions this might cause.
func ResponseMessageNoMention(message ...string) api.InteractionResponse {
	return api.InteractionResponse{
		Type: api.MessageInteractionWithSource,
		Data: &api.InteractionResponseData{
			Content: option.NewNullableString(strings.Join(message, " ")),
			AllowedMentions: &api.AllowedMentions{
				Parse: []api.AllowedMentionType{},
			},
		},
	}
}
