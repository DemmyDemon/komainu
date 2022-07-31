package interactions

import (
	"bytes"
	"fmt"
	"komainu/interactions/command"
	"komainu/interactions/message"
	"komainu/interactions/response"
	"komainu/storage"
	"log"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
)

func init() {
	command.Register("seen", command.Handler{
		Description: "Check when someone was last around",
		Code:        CommandSeen,
		Options: []discord.CommandOption{
			&discord.UserOption{
				OptionName:  "user",
				Description: "The user to look up",
				Required:    true,
			},
		},
	})
	command.Register("neverseen", command.Handler{
		Description: "Get a list of people that the bot has never seen say anything!",
		Code:        CommandNeverSeen,
		Options:     []discord.CommandOption{},
	})
	command.Register("inactive", command.Handler{
		Description: "Get a list of inactive people",
		Code:        CommandInactive,
		Options: []discord.CommandOption{
			&discord.IntegerOption{
				OptionName:  "days",
				Description: "How many days of quiet makes someone inactive?",
				Required:    true,
			},
		},
	})
	command.Register("activerole", command.Handler{
		Description: "Set what role is granted and revoked for active/inactive users, and under what conditions.",
		Code:        CommandActiveRole,
		Options: []discord.CommandOption{
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
	})
	command.Register("seeeveryone", command.Handler{
		Description: "Ruin the /seen system by marking everyone here as seen right now.",
		Code:        CommandSeeEveryone,
		Options:     []discord.CommandOption{},
	})
	message.Register(message.Handler{Code: MessageSeen})
}

func MessageSeen(state *state.State, kvs storage.KeyValueStore, event *gateway.MessageCreateEvent) {
	if event.GuildID == 0 {
		return // It's either a private message, or an ephemeral-response command. Doesn't count.
	}

	if err := storage.See(kvs, event.GuildID, event.Author.ID); err != nil {
		log.Printf("[%s] Error seeing %s in %s: %s\n", event.GuildID, event.Author.ID, event.ChannelID, err)
	} else {
		log.Printf("[%s] <@%s> seen in <#%s>\n", event.GuildID, event.Author.ID, event.ChannelID)
		if err := storage.MaybeGiveActiveRole(kvs, state, event.GuildID, event.Member); err != nil {
			log.Printf("[%s] Failed to give active role to %s: %s\n", event.GuildID, event.Author.ID, err)
		}
	}
}

// CommandSeen processes a command to look up when a user was last seen.
func CommandSeen(state *state.State, kvs storage.KeyValueStore, event *gateway.InteractionCreateEvent, cmd *discord.CommandInteraction) command.Response {
	if cmd.Options != nil && len(cmd.Options) > 0 {
		option, err := cmd.Options[0].SnowflakeValue()
		if err != nil {
			log.Printf("[%s] Failed to get snowflake value for /seen: %s\n", event.GuildID, err)
			return command.Response{Response: response.Ephemeral("An error occured, and has been logged."), Callback: nil}
		}
		if me, err := state.Me(); err != nil {
			log.Printf("[%s] Failed to look up myself to see if I match /seen: %s\n", event.GuildID, err)
			return command.Response{Response: response.Ephemeral("An error occured, and has been logged."), Callback: nil}
		} else if me.ID == discord.UserID(option) {
			return command.Response{Response: response.Ephemeral("I'm right here, buddy!"), Callback: nil}
		}

		found, timestamp, err := storage.LastSeen(kvs, event.GuildID, discord.UserID(option))
		if err != nil {
			log.Printf("[%s] Failed to get %s from Key/Value Store for /seen lookup: %s\n", event.GuildID, option, err)
			return command.Response{Response: response.Ephemeral("An error occured, and has been logged."), Callback: nil}
		}
		if !found {
			return command.Response{Response: response.MessageNoMention(fmt.Sprintf("Sorry, I've never seen <@%s> say anything at all!", option)), Callback: nil}
		}
		return command.Response{Response: response.MessageNoMention(fmt.Sprintf("I last saw <@%s> <t:%d:R>", option, timestamp)), Callback: nil}
	}
	return command.Response{Response: response.Ephemeral("No user given?!"), Callback: nil}
}

// CommandInactive processes a command to list who has not been active in a given timeframe.
func CommandInactive(state *state.State, kvs storage.KeyValueStore, event *gateway.InteractionCreateEvent, cmd *discord.CommandInteraction) command.Response {
	days := int64(30)
	if cmd.Options != nil && len(cmd.Options) > 0 {
		d, err := cmd.Options[0].IntValue()
		if err != nil {
			log.Printf("[%s] Failed to get int value for /inactive: %s", event.GuildID, err)
			return command.Response{Response: response.Ephemeral("An error occured, and has been logged."), Callback: nil}
		}
		if d <= 0 {
			return command.Response{Response: response.Ephemeral(fmt.Sprintf("Everyone. Everyone has been inactive for at least %d days.", d)), Callback: nil}
		}
		days = d
	}
	atLeast := time.Now().Unix() - (24 * 3600 * days)
	members, err := state.Session.Members(event.GuildID, 0)
	if err != nil {
		log.Printf("[%s] Failed to get member list for /inactive lookup: %s", event.GuildID, err)
		return command.Response{Response: response.Ephemeral("An error occured, and has been logged."), Callback: nil}
	}

	var bt bytes.Buffer
	never := 0
	inactiveCount := 0

	now := time.Now()

	for _, member := range members {

		if member.User.Bot {
			continue
		}

		seen, when, err := storage.LastSeen(kvs, event.GuildID, member.User.ID)
		if err != nil {
			log.Printf("[%s] Failed to get a storage.LastSeen for %s: %s", event.GuildID, member.User.ID, err)
			return command.Response{Response: response.Ephemeral("An error occured, and has been logged."), Callback: nil}
		} else if !seen {
			never++
			joinTime := member.Joined.Format("2006-01-02")
			if now.Sub(member.Joined.Time()).Hours() < 24 {
				joinTime = "very recently"
			}
			if member.Nick != "" {
				fmt.Fprintf(&bt, "<%s#%s> (%s) never, joined %s\n", member.User.Username, member.User.Discriminator, member.Nick, joinTime)
			} else {
				fmt.Fprintf(&bt, "<%s#%s> never, joined %s\n", member.User.Username, member.User.Discriminator, joinTime)
			}
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

	if inactiveCount+never > 0 {
		return command.Response{Response: response.MessageAttachFile(
			message,
			fmt.Sprintf("inactive_report_%s.txt", time.Now().Format("2006-01-02")),
			&bt,
		), Callback: nil}
	} else {
		return command.Response{Response: response.Message(message), Callback: nil}
	}
}

// CommandNeverSeen processes a command to list everyone that has never been seen by the bot.
func CommandNeverSeen(state *state.State, kvs storage.KeyValueStore, event *gateway.InteractionCreateEvent, cmd *discord.CommandInteraction) command.Response {
	members, err := state.Session.Members(event.GuildID, 0)
	if err != nil {
		log.Printf("[%s] Failed to get member list for /neverseen lookup: %s", event.GuildID, err)
		return command.Response{Response: response.Ephemeral("An error occured, and has been logged."), Callback: nil}
	}
	count := 0

	var bt bytes.Buffer
	for _, member := range members {

		if member.User.Bot {
			continue
		}

		seen, _, err := storage.LastSeen(kvs, event.GuildID, member.User.ID)
		if err != nil {
			log.Printf("[%s] Failed to get a storage.LastSeen for %s: %s", event.GuildID, member.User.ID, err)
			return command.Response{Response: response.Ephemeral("An error occured, and has been logged."), Callback: nil}
		}
		if !seen {
			count++
			if member.Nick != "" {
				fmt.Fprintf(&bt, "%s#%s (%s) joined %s\n", member.User.Username, member.User.Discriminator, member.Nick, member.Joined.Format("2006-01-02"))
			} else {
				fmt.Fprintf(&bt, "%s#%s joined %s\n", member.User.Username, member.User.Discriminator, member.Joined.Format("2006-01-02"))
			}
		}
	}

	if count > 0 {
		return command.Response{Response: response.MessageAttachFile(
			fmt.Sprintf("%d users have never been seen by me.", count),
			fmt.Sprintf("never_seen_report_%s.txt", time.Now().Format("2006-01-02")),
			&bt,
		), Callback: nil}
	} else {
		return command.Response{Response: response.Message("Everyone seems to have at least said at least *something!*"), Callback: nil}
	}
}

// CommandActiveRole processes a command to set an automatic "active" role and revoke it after a certain amount of days.
func CommandActiveRole(state *state.State, kvs storage.KeyValueStore, event *gateway.InteractionCreateEvent, cmd *discord.CommandInteraction) command.Response {
	if cmd.Options == nil || len(cmd.Options) != 2 {
		log.Printf("[%s] /activerole has a weird number of arguments\n", event.GuildID)
		return command.Response{Response: response.Ephemeral("Wait, what? Something odd happened, and was logged."), Callback: nil}
	}

	// Processing these in reverse order because of the Special Meaning of days == 0
	days, err := cmd.Options[1].IntValue()
	if err != nil {
		log.Printf("[%s] Error encountered trying to turn days argument into an actual int64 in /activerole: %s\n", event.GuildID, err)
		return command.Response{Response: response.Ephemeral("That's very odd. I've logged that it didn't go according to plan."), Callback: nil}
	}

	if days == 0 {
		if _, err := kvs.Delete(event.GuildID, "activerole", "days"); err != nil {
			log.Printf("[%s] Tried to disable activerole, but %s", event.GuildID, err)
		}
		if _, err := kvs.Delete(event.GuildID, "activerole", "role"); err != nil {
			log.Printf("[%s] Tried to disable activerole, however %s", event.GuildID, err)
		}
		return command.Response{Response: response.Message("So noted. Feature disabled."), Callback: nil}
	}

	if err := kvs.Set(event.GuildID, "activerole", "days", days); err != nil {
		log.Printf("[%s] Error storing the days for /activerole: %s\n", event.GuildID, err)
		return command.Response{Response: response.Ephemeral("There is something strange in this neighbourhood. I've logged it for the Bug Busters to investigate later."), Callback: nil}
	}

	snowflake, err := cmd.Options[0].SnowflakeValue()
	if err != nil {
		log.Printf("[%s] Error encountered trying to turn role argument into an actual snowflake in /activerole: %s\n", event.GuildID, err)
		return command.Response{Response: response.Ephemeral("That's very odd. I've logged that it didn't go as planned."), Callback: nil}
	}
	roleID := discord.RoleID(snowflake)
	if err := kvs.Set(event.GuildID, "activerole", "role", roleID); err != nil {
		log.Printf("[%s] Error storing the role for /activerole: %s\n", event.GuildID, err)
		return command.Response{Response: response.Ephemeral("There is something strange in this neighbourhood. I've logged it for the Bug Busters to look at later."), Callback: nil}
	}

	return command.Response{Response: response.MessageNoMention(fmt.Sprintf("Okay, will revoke <@&%d> after %d days, and grant it to anyone that says anything.", roleID, days)), Callback: nil}

}

// CommandSeeEveryone processes a command to mark eeeeveryone in the guild as "seen" right now.
func CommandSeeEveryone(state *state.State, kvs storage.KeyValueStore, event *gateway.InteractionCreateEvent, cmd *discord.CommandInteraction) command.Response {
	members, err := state.Session.Members(event.GuildID, 0)
	if err != nil {
		log.Printf("[%s] Failed to get member list for /SeeEveryone: %s", event.GuildID, err)
		return command.Response{Response: response.Ephemeral("An error occured, and has been logged."), Callback: nil}
	}
	for _, member := range members {
		err := storage.See(kvs, event.GuildID, member.User.ID)
		if err != nil {
			log.Printf("[%s] Failed to See member during seeing spree: %s", event.GuildID, err)
			return command.Response{Response: response.Message("Okay, something weird happened partway through that. It was logged."), Callback: nil}
		}
	}
	return command.Response{Response: response.Message("Eeeeeveryone was marked as being seen just now."), Callback: nil}
}
