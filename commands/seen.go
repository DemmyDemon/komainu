package commands

import (
	"bytes"
	"fmt"
	"komainu/storage"
	"log"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
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

var CommandActiveRoleObject Command = Command{
	group:       "seen",
	description: "Set what role is granted and revoked for active/inactive users, and under what conditions.",
	code:        CommandActiveRole,
	options: []discord.CommandOption{
		&discord.RoleOption{
			OptionName:  "role",
			Description: "The role to giveth and taketh away.",
			Required:    true,
		},
		&discord.NumberOption{
			OptionName:  "days",
			Description: "How many days someone needs to be inactive to lose the role. Set to zero to disable this function.",
			Required:    true,
			Min:         option.NewFloat(0),
			Max:         option.NewFloat(365),
		},
	},
}

var CommandSeeEveryoneObject Command = Command{
	group:       "seen",
	description: "Ruin the /seen system by marking everyone here as seen right now.",
	code:        CommandSeeEveryone,
	options:     []discord.CommandOption{},
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
			return CommandResponse{ResponseEphemeral(fmt.Sprintf("Everyone. Everyone has been inactive for at least %d days.", d)), nil}
		}
		days = d
	}
	atLeast := time.Now().Unix() - (24 * 3600 * days)
	members, err := state.Session.Members(event.GuildID, 0)
	if err != nil {
		log.Printf("[%s] Failed to get member list for /inactive lookup: %s", event.GuildID, err)
		return CommandResponse{ResponseEphemeral("An error occured, and has been logged."), nil}
	}

	var bt bytes.Buffer
	never := 0
	inactiveCount := 0

	now := time.Now()

	for _, member := range members {
		seen, when, err := storage.LastSeen(kvs, event.GuildID, member.User.ID)
		if err != nil {
			log.Printf("[%s] Failed to get a storage.LastSeen for %s: %s", event.GuildID, member.User.ID, err)
			return CommandResponse{ResponseEphemeral("An error occured, and has been logged."), nil}
		} else if !seen {
			never++
		} else if when <= atLeast {
			then := time.Unix(when, 0)
			timeDiff := now.Sub(then)
			if member.Nick != "" {
				fmt.Fprintf(&bt, "<%s#%s> (%s) %d days\n", member.User.Username, member.User.Discriminator, member.Nick, int(timeDiff.Hours()/24))
			} else {
				fmt.Fprintf(&bt, "<%s#%s> %d days\n", member.User.Username, member.User.Discriminator, int(timeDiff.Hours()/24))
			}
			inactiveCount++
		}
	}

	message := fmt.Sprintf("%d inactive in the last %d days, out of %d members.", inactiveCount+never, days, len(members))
	if never > 0 {
		message += fmt.Sprintf(" (Including %d that I have never seen say anything!)", never)
	}
	message += "\n"

	return CommandResponse{ResponseMessageAttachText(
		message,
		fmt.Sprintf("inactive_report_%s.txt", time.Now().Format("2006-01-02")),
		&bt,
	), nil}
}

// CommandNeverSeen processes a command to list everyone that has never been seen by the bot.
func CommandNeverSeen(state *state.State, kvs storage.KeyValueStore, event *gateway.InteractionCreateEvent, command *discord.CommandInteraction) CommandResponse {
	members, err := state.Session.Members(event.GuildID, 0)
	if err != nil {
		log.Printf("[%s] Failed to get member list for /neverseen lookup: %s", event.GuildID, err)
		return CommandResponse{ResponseEphemeral("An error occured, and has been logged."), nil}
	}
	count := 0

	var bt bytes.Buffer
	for _, member := range members {
		seen, _, err := storage.LastSeen(kvs, event.GuildID, member.User.ID)
		if err != nil {
			log.Printf("[%s] Failed to get a storage.LastSeen for %s: %s", event.GuildID, member.User.ID, err)
			return CommandResponse{ResponseEphemeral("An error occured, and has been logged."), nil}
		}
		if !seen {
			count++
			if member.Nick != "" {
				fmt.Fprintf(&bt, "<%s#%s> (%s)\n", member.User.Username, member.User.Discriminator, member.Nick)
			} else {
				fmt.Fprintf(&bt, "<%s#%s>\n", member.User.Username, member.User.Discriminator)
			}
		}
	}

	if count > 0 {
		return CommandResponse{ResponseMessageAttachText(
			fmt.Sprintf("%d users have never been seen by me.", count),
			fmt.Sprintf("never_seen_report_%s.txt", time.Now().Format("2006-01-02")),
			&bt,
		), nil}
	} else {
		return CommandResponse{ResponseMessage("Everyone seems to have at least said at least *something!*"), nil}
	}
}

// CommandActiveRole processes a command to set an automatic "active" role and revoke it after a certain amount of days.
func CommandActiveRole(state *state.State, kvs storage.KeyValueStore, event *gateway.InteractionCreateEvent, command *discord.CommandInteraction) CommandResponse {
	if command.Options == nil || len(command.Options) != 2 {
		log.Printf("[%s] /activerole has a weird number of arguments\n", event.GuildID)
		return CommandResponse{ResponseEphemeral("Wait, what? Something odd happened, and was logged."), nil}
	}

	// Processing these in reverse order because of the Special Meaning of days == 0
	days, err := command.Options[1].IntValue()
	if err != nil {
		log.Printf("[%s] Error encountered trying to turn days argument into an actual int64 in /activerole: %s\n", event.GuildID, err)
		return CommandResponse{ResponseEphemeral("That's very odd. I've logged that it didn't go according to plan."), nil}
	}

	if days == 0 {
		if _, err := kvs.Delete(event.GuildID, "activerole", "days"); err != nil {
			log.Printf("[%s] Tried to disable activerole, but %s", event.GuildID, err)
		}
		if _, err := kvs.Delete(event.GuildID, "activerole", "role"); err != nil {
			log.Printf("[%s] Tried to disable activerole, however %s", event.GuildID, err)
		}
		return CommandResponse{ResponseMessage("So noted. Feature disabled."), nil}
	}

	if err := kvs.Set(event.GuildID, "activerole", "days", days); err != nil {
		log.Printf("[%s] Error storing the days for /activerole: %s\n", event.GuildID, err)
		return CommandResponse{ResponseEphemeral("There is something strange in this neighbourhood. I've logged it for the Bug Busters to investigate later."), nil}
	}

	snowflake, err := command.Options[0].SnowflakeValue()
	if err != nil {
		log.Printf("[%s] Error encountered trying to turn role argument into an actual snowflake in /activerole: %s\n", event.GuildID, err)
		return CommandResponse{ResponseEphemeral("That's very odd. I've logged that it didn't go as planned."), nil}
	}
	roleID := discord.RoleID(snowflake)
	if err := kvs.Set(event.GuildID, "activerole", "role", roleID); err != nil {
		log.Printf("[%s] Error storing the role for /activerole: %s\n", event.GuildID, err)
		return CommandResponse{ResponseEphemeral("There is something strange in this neighbourhood. I've logged it for the Bug Busters to look at later."), nil}
	}

	return CommandResponse{ResponseMessageNoMention(fmt.Sprintf("Okay, will revoke <@&%d> after %d days, and grant it to anyone that says anything.", roleID, days)), nil}

}

// CommandSeeEveryone processes a command to mark eeeeveryone in the guild as "seen" right now.
func CommandSeeEveryone(state *state.State, kvs storage.KeyValueStore, event *gateway.InteractionCreateEvent, command *discord.CommandInteraction) CommandResponse {
	members, err := state.Session.Members(event.GuildID, 0)
	if err != nil {
		log.Printf("[%s] Failed to get member list for /SeeEveryone: %s", event.GuildID, err)
		return CommandResponse{ResponseEphemeral("An error occured, and has been logged."), nil}
	}
	for _, member := range members {
		err := storage.See(kvs, event.GuildID, member.User.ID)
		if err != nil {
			log.Printf("[%s] Failed to See member during seeing spree: %s", event.GuildID, err)
			return CommandResponse{ResponseMessage("Okay, something weird happened partway through that. It was logged."), nil}
		}
	}
	return CommandResponse{ResponseMessage("Eeeeeveryone was marked as being seen just now."), nil}
}
