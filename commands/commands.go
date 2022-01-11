package commands

import (
	"komainu/storage"
	"komainu/utility"
	"log"
	"sort"
	"strings"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
)

type CommandFunction func(
	state *state.State,
	kvs storage.KeyValueStore,
	event *gateway.InteractionCreateEvent,
	command *discord.CommandInteraction,
) CommandResponse

type CommandResponse struct {
	InteractionResponse api.InteractionResponse
	Callback            func(message *discord.Message)
}

type Command struct {
	group       string
	description string
	code        CommandFunction
	options     []discord.CommandOption
}

var commands = map[string]Command{
	"access": {"access", "Grant, revoke and list command group access", CommandAccess, []discord.CommandOption{
		&discord.SubcommandOption{
			OptionName:  "grant",
			Description: "Grant a role access to something",
			Options: []discord.CommandOptionValue{
				&discord.StringOption{
					OptionName:  "group",
					Description: "The command group to grant access to",
					Required:    true,
				},
				&discord.RoleOption{
					OptionName:  "role",
					Description: "The role that gets this access",
					Required:    true,
				},
			},
		},
		&discord.SubcommandOption{
			OptionName:  "revoke",
			Description: "Revoke access to something from a role",
			Options: []discord.CommandOptionValue{
				&discord.StringOption{
					OptionName:  "group",
					Description: "The command group to revoke access from",
					Required:    true,
				},
				&discord.RoleOption{
					OptionName:  "role",
					Description: "The role that loses this access",
					Required:    true,
				},
			},
		},
		&discord.SubcommandOption{
			OptionName:  "list",
			Description: "List what roles have access to what command groups",
			Options:     []discord.CommandOptionValue{},
		},
	}},

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

	"faq": {"faquser", "Look up a FAQ topic", CommandFaq, []discord.CommandOption{
		&discord.StringOption{
			OptionName:  "topic",
			Description: "The name of the topic you wish to recall",
			Required:    true,
		},
	}},
	"faqset": {"faqadmin", "Manage FAQ topics", CommandFaqSet, []discord.CommandOption{
		&discord.SubcommandOption{
			OptionName:  "add",
			Description: "Add a topic to the FAQ",
			Options: []discord.CommandOptionValue{
				&discord.StringOption{
					OptionName:  "topic",
					Description: "The word used to recall this item later",
					Required:    true,
				},
				&discord.StringOption{
					OptionName:  "content",
					Description: "What you want the topic to contain",
					Required:    true,
				},
			},
		},
		&discord.SubcommandOption{
			OptionName:  "remove",
			Description: "Remove a topic from the FAQ",
			Options: []discord.CommandOptionValue{
				&discord.StringOption{
					OptionName:  "topic",
					Description: "What do you want to permanently obliterate from the FAQ?",
					Required:    true,
				},
			},
		},
		&discord.SubcommandOption{
			OptionName:  "list",
			Description: "List the known topics in the FAQ",
			Options:     []discord.CommandOptionValue{},
		},
	}},

	"vote": {"vote", "Initiate a vote", CommandVote, []discord.CommandOption{
		&discord.NumberOption{
			OptionName:  "length",
			Description: "The number of hours the vote should run.",
			Required:    true,
			Min:         option.NewFloat(0),
			Max:         option.NewFloat(336), // 336 hours is two weeks.
		},
		&discord.StringOption{
			OptionName:  "question",
			Description: "The question being asked. Works best as a yes/no question.",
			Required:    true,
		},
		&discord.StringOption{
			OptionName:  "first",
			Description: "The first vote option description (80 char max)",
			Required:    true,
		},
		&discord.StringOption{
			OptionName:  "second",
			Description: "The second vote option description (80 char max)",
			Required:    true,
		},
		&discord.StringOption{
			OptionName:  "third",
			Description: "The third vote option description (80 char max)",
			Required:    false,
		},
		&discord.StringOption{
			OptionName:  "fourth",
			Description: "The fourth vote option description (80 char max)",
			Required:    false,
		},
	}},
}

// *Another* global, to avoid an initalization cycle :-/
// Value is defined when AddCommandHandler is called, to avoid it being cyclc.
var commandGroups []string

// The Token Bins. 5 and 10 are arbituary numbers, and it decrements at 10 second intervals.
var userTokenBin = &utility.TokenBin{Max: 5, Interval: 10}
var channelTokenBin = &utility.TokenBin{Max: 10, Interval: 10}

// GetCommandGroups returns all the command groups.
func GetCommandGroups() []string {
	keys := make(map[string]bool)
	groups := []string{}
	for _, data := range commands {
		if _, value := keys[data.group]; !value {
			keys[data.group] = true
			groups = append(groups, data.group)
		}
	}
	sort.Strings(groups)
	return groups
}

// HasAccess checks if the given user has access to the given command group in the given guild.
func HasAccess(kvs storage.KeyValueStore, state *state.State, guildID discord.GuildID, channelID discord.ChannelID, member *discord.Member, group string) bool {
	if member == nil {
		return false
	}

	// First we check the KVS
	granted := []discord.RoleID{}
	found, err := kvs.GetObject(guildID, "access", group, &granted)
	if err != nil {
		log.Printf("[%s] HasAccess check failed to obtain access list from KVS: %s\n", guildID, err)
	} else if found {
		if utility.RoleInCommon(granted, member.RoleIDs) {
			return true
		}
	}

	// Then we check if this is The Owner Themself
	if guild, err := state.Guild(guildID); err != nil {
		log.Printf("Could not look up guild %s for access check: %s\n", guildID, err)
		return false // Better safe than sorry!
	} else if guild.OwnerID == member.User.ID {
		return true // Owner always has access to everything.
	}

	// Lastly, we check if they're an administrator
	if permissions, err := state.Permissions(channelID, member.User.ID); err != nil {
		log.Printf("Could not look up permissions for %s in channel %s for access check: %s\n", member.User.ID, channelID, err)
		return false // Better safe than sorry!
	} else if permissions.Has(discord.PermissionAdministrator) {
		return true // Administrators get access to everyting
	}

	return false // If all else fails, they're not authorized.
}

func AddDeleteHandler(state *state.State, kvs storage.KeyValueStore) {
	state.AddHandler(func(e *gateway.MessageDeleteEvent) {
		if e.GuildID == discord.NullGuildID {
			return
		}
		_, err := kvs.Delete(e.GuildID, "votes", e.ID)
		if err != nil {
			log.Printf("[%s] Encountered an error removing vote from KVS after message deletion: %s\n", e.GuildID, err)
		}
	})
}

// AddCommandHandler, surprisingly, adds the command handler.
func AddCommandHandler(state *state.State, kvs storage.KeyValueStore) {
	commandGroups = GetCommandGroups()
	state.AddHandler(func(e *gateway.InteractionCreateEvent) {

		switch interaction := e.Data.(type) {
		case *discord.CommandInteraction:

			if !userTokenBin.Allocate(discord.Snowflake(e.GuildID), discord.Snowflake(e.Member.User.ID)) {
				if err := state.RespondInteraction(e.ID, e.Token, ResponseEphemeral("You are using too many commands too quickly. Calm down.")); err != nil {
					log.Println("An error occured posting throttle warning emphemral response (user):", err)
				}
				return
			}
			if !channelTokenBin.Allocate(discord.Snowflake(e.GuildID), discord.Snowflake(e.ChannelID)) {
				if err := state.RespondInteraction(e.ID, e.Token, ResponseEphemeral("Too many commands being processed in this channel right now. Please wait.")); err != nil {
					log.Println("An error occured posting throttle warning emphemral response (channel):", err)
				}
				return
			}

			if val, ok := commands[interaction.Name]; ok {
				if !HasAccess(kvs, state, e.GuildID, e.ChannelID, e.Member, val.group) {
					if err := state.RespondInteraction(e.ID, e.Token, ResponseEphemeral("Sorry, access was denied.")); err != nil {
						log.Println("An error occured posting access denied response:", err)
					}
					return
				}

				response := val.code(state, kvs, e, interaction)

				if err := state.RespondInteraction(e.ID, e.Token, response.InteractionResponse); err != nil {
					log.Println("Failed to send interaction response:", err)
				}
				if response.Callback != nil {
					message, err := state.InteractionResponse(e.AppID, e.Token)
					if err != nil {
						log.Printf("Error %s getting message reference for %s command callback\n", err, interaction.Name)
						return
					}
					if message != nil && message.ID != discord.NullMessageID {
						response.Callback(message)
					}
				}
			}
		case discord.ComponentInteraction:
			isVote, response, err := storage.HandleInteractionAsVote(state, kvs, e, interaction)
			if err != nil {
				log.Printf("[%s] error while trying to handle an interaction as a vote: %s\n", e.GuildID, err)
				return
			}
			if isVote {
				if response != "" {
					if err := state.RespondInteraction(e.ID, e.Token, ResponseEphemeral(response)); err != nil {
						log.Printf("[%s] Failed to send component interaction ephemeral response: %s\n", e.GuildID, err)
					}
				}
				return
			}
		default:
			log.Printf("Unhandled interaction type %T\n", e.Data)
			return
		}
	})
}

// RegisterCommands registers the command in the given guild, clearing out any obsolete commands.
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

// ResponseEphemeral generates an emphemeral response message from the strings given.
func ResponseEphemeral(message ...string) api.InteractionResponse {
	return api.InteractionResponse{
		Type: api.MessageInteractionWithSource,
		Data: &api.InteractionResponseData{
			Content: option.NewNullableString(strings.Join(message, " ")),
			Flags:   api.EphemeralResponse,
		},
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
