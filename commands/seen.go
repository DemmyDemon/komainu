package commands

import (
	"fmt"
	"komainu/storage"
	"log"
	"strings"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
)

var CommandSeenObject Command = Command{
	group:       "seen",
	description: "Check when someone was last around",
	code:        CommandSeen,
	options: []discord.CommandOption{
		&discord.UserOption{
			OptionName:  "user",
			Description: "The user to look up",
			Required:    true,
		},
	},
}
var CommandNeverSeenObject Command = Command{
	group:       "seen",
	description: "Get a list of people that the bot has never seen say anything!",
	code:        CommandNeverSeen,
	options:     []discord.CommandOption{},
}

var CommandInactiveObject Command = Command{
	group:       "seen",
	description: "Get a list of inactive people",
	code:        CommandInactive,
	options: []discord.CommandOption{
		&discord.IntegerOption{
			OptionName:  "days",
			Description: "How many days of quiet makes someone inactive?",
			Required:    true,
		},
	},
}

// CommandSeen processes a command to look up when a user was last seen.
func CommandSeen(state *state.State, kvs storage.KeyValueStore, event *gateway.InteractionCreateEvent, command *discord.CommandInteraction) CommandResponse {
	if command.Options != nil && len(command.Options) > 0 {
		option, err := command.Options[0].SnowflakeValue()
		if err != nil {
			log.Printf("[%s] Failed to get snowflake value for /seen: %s\n", event.GuildID, err)
			return CommandResponse{ResponseEphemeral("An error occured, and has been logged."), nil}
		}
		if me, err := state.Me(); err != nil {
			log.Printf("[%s] Failed to look up myself to see if I match /seen: %s\n", event.GuildID, err)
			return CommandResponse{ResponseEphemeral("An error occured, and has been logged."), nil}
		} else if me.ID == discord.UserID(option) {
			return CommandResponse{ResponseEphemeral("I'm right here, buddy!"), nil}
		}

		found, timestamp, err := storage.LastSeen(kvs, event.GuildID, discord.UserID(option))
		if err != nil {
			log.Printf("[%s] Failed to get %s from Key/Value Store for /seen lookup: %s\n", event.GuildID, option, err)
			return CommandResponse{ResponseEphemeral("An error occured, and has been logged."), nil}
		}
		if !found {
			return CommandResponse{ResponseMessageNoMention(fmt.Sprintf("Sorry, I've never seen <@%s> say anything at all!", option)), nil}
		}
		return CommandResponse{ResponseMessageNoMention(fmt.Sprintf("I last saw <@%s> <t:%d:R>", option, timestamp)), nil}
	}
	return CommandResponse{ResponseEphemeral("No user given?!"), nil}
}

// CommandInactive processes a command to list who has not been active in a given timeframe.
func CommandInactive(state *state.State, kvs storage.KeyValueStore, event *gateway.InteractionCreateEvent, command *discord.CommandInteraction) CommandResponse {
	days := int64(30)
	if command.Options != nil && len(command.Options) > 0 {
		d, err := command.Options[0].IntValue()
		if err != nil {
			log.Printf("[%s] Failed to get int value for /inactive: %s", event.GuildID, err)
			return CommandResponse{ResponseEphemeral("An error occured, and has been logged."), nil}
		}
		if d <= 0 {
			return CommandResponse{ResponseEphemeral(fmt.Sprintf("Everyone. Everyone has been inactive for %d days.", d)), nil}
		}
		days = d
	}
	atLeast := time.Now().Unix() - (24 * 3600 * days)
	members, err := state.Session.Members(event.GuildID, 0)
	if err != nil {
		log.Printf("[%s] Failed to get member list for /inactive lookup: %s", event.GuildID, err)
		return CommandResponse{ResponseEphemeral("An error occured, and has been logged."), nil}
	}

	never := 0
	var sb strings.Builder

	inactiveCount := 0

	for _, member := range members {
		seen, when, err := storage.LastSeen(kvs, event.GuildID, member.User.ID)
		if err != nil {
			log.Printf("[%s] Failed to get a storage.LastSeen for %s: %s", event.GuildID, member.User.ID, err)
			return CommandResponse{ResponseEphemeral("An error occured, and has been logged."), nil}
		} else if !seen {
			never++
		} else if when <= atLeast {
			fmt.Fprintf(&sb, "<@%d> <t:%d:R>\n", member.User.ID, when)
			inactiveCount++
		}
	}
	fmt.Fprintf(&sb, "%d inactive in the last %d days, out of %d members.\n", inactiveCount+never, days, len(members))
	if never > 0 {
		fmt.Fprintf(&sb, "(Including %d that I have never seen say anything!)", never)
	}
	return CommandResponse{ResponseMessageNoMention(sb.String()), nil}
}

// CommandNeverSeen processes a command to list everyone that has never been seen by the bot.
func CommandNeverSeen(state *state.State, kvs storage.KeyValueStore, event *gateway.InteractionCreateEvent, command *discord.CommandInteraction) CommandResponse {
	members, err := state.Session.Members(event.GuildID, 0)
	if err != nil {
		log.Printf("[%s] Failed to get member list for /neverseen lookup: %s", event.GuildID, err)
		return CommandResponse{ResponseEphemeral("An error occured, and has been logged."), nil}
	}
	count := 0

	var sb strings.Builder
	for _, member := range members {
		seen, _, err := storage.LastSeen(kvs, event.GuildID, member.User.ID)
		if err != nil {
			log.Printf("[%s] Failed to get a storage.LastSeen for %s: %s", event.GuildID, member.User.ID, err)
			return CommandResponse{ResponseEphemeral("An error occured, and has been logged."), nil}
		}
		if !seen {
			count++
			fmt.Fprintf(&sb, "<@%d> ", member.User.ID)
		}
	}
	if count > 0 {
		fmt.Fprintf(&sb, "\n%d users have never been seen by me.", count)
	} else {
		fmt.Fprintf(&sb, "Everyone seems to have at least said at least *something!*")
	}
	return CommandResponse{ResponseMessageNoMention(sb.String()), nil}
}
