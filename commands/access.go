package commands

import (
	"fmt"
	"komainu/storage"
	"komainu/utility"
	"log"
	"strings"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
)

// CommandAccess processes a command to list access entries.
func CommandAccess(state *state.State, kvs storage.KeyValueStore, event *gateway.InteractionCreateEvent, command *discord.CommandInteraction) CommandResponse {
	if command.Options == nil || len(command.Options) != 1 {
		log.Printf("[%s] /access command structure is somehow nil or not a single element. Wat.\n", event.GuildID)
		return CommandResponse{ResponseMessage("I'm sorry, what? Something very weird happened."), nil}
	}
	switch command.Options[0].Name {
	case "grant":
		return CommandResponse{SubCommandAccessGrant(kvs, event.GuildID, command.Options[0].Options), nil}
	case "revoke":
		return CommandResponse{SubCommandAccessRevoke(kvs, event.GuildID, command.Options[0].Options), nil}
	case "list":
		return CommandResponse{SubCommandAccessList(kvs, event.GuildID), nil}
	default:
		return CommandResponse{ResponseMessage("Unknown subcommand! Clearly *someone* dropped the ball!"), nil}
	}
}

// SubCommandAccessGrant processes a sub command to grant access.
func SubCommandAccessGrant(kvs storage.KeyValueStore, guildID discord.GuildID, options []discord.CommandInteractionOption) api.InteractionResponse {
	if options == nil || len(options) != 2 {
		log.Printf("[%s] /access grant command structure is somehow nil or not two elements. Wat.\n", guildID)
		return ResponseEphemeral("Invalid command structure.")
	}

	commandGroup := strings.ToLower(options[0].String())
	if !utility.ContainsString(commandGroups, commandGroup) {
		return ResponseEphemeral(fmt.Sprintf("Sorry, `%s` is not a valid command group.", commandGroup))
	}

	value, err := options[1].SnowflakeValue()
	if err != nil {
		log.Printf("[%s] /access grant failed to obtain snowflake from first argument (%v): %s\n", guildID, options[1], err)
		return ResponseEphemeral("An error occured, and has been logged.")
	}
	roleID := discord.RoleID(value)

	granted := []discord.RoleID{}
	found, err := kvs.GetObject(guildID, "access", commandGroup, &granted)
	if err != nil {
		log.Printf("[%s] /access grant failed to obtain access list from KVS: %s\n", guildID, err)
		return ResponseEphemeral("An error occured, and has been logged.")
	}
	if !found || !utility.ContainsRole(granted, roleID) {
		granted = append(granted, roleID)
		err := kvs.Set(guildID, "access", commandGroup, granted)
		if err != nil {
			log.Printf("[%s] /access grant failed to store updated access list in KVS: %s\n", guildID, err)
			return ResponseEphemeral("An error occured, and has been logged.")
		}
	}
	return ResponseMessageNoMention(fmt.Sprintf("<@&%s> now has access to the `%s` command group\n", roleID, commandGroup))
}

// SubCommandAccessRevoke processes a sub command to revoke access.
func SubCommandAccessRevoke(kvs storage.KeyValueStore, guildID discord.GuildID, options []discord.CommandInteractionOption) api.InteractionResponse {
	if options == nil || len(options) != 2 {
		log.Printf("[%s] /access revoke command structure is somehow nil or not two elements. Wat.\n", guildID)
		return ResponseEphemeral("Invalid command structure.")
	}

	commandGroup := strings.ToLower(options[0].String())
	if !utility.ContainsString(commandGroups, commandGroup) {
		return ResponseEphemeral(fmt.Sprintf("Sorry, `%s` is not a valid command group.", commandGroup))
	}

	value, err := options[1].SnowflakeValue()
	if err != nil {
		log.Printf("[%s] /access revoke failed to obtain snowflake from first argument (%v): %s\n", guildID, options[1], err)
		return ResponseEphemeral("An error occured, and has been logged.")
	}
	roleID := discord.RoleID(value)

	granted := []discord.RoleID{}
	found, err := kvs.GetObject(guildID, "access", commandGroup, &granted)
	if err != nil {
		log.Printf("[%s] /access revoke failed to obtain access list from KVS: %s\n", guildID, err)
		return ResponseEphemeral("An error occured, and has been logged.")
	}
	if found && utility.ContainsRole(granted, roleID) {

		for idx, item := range granted {
			if item == roleID {
				granted[idx] = granted[len(granted)-1] // Copy last element to index idx.
				granted = granted[:len(granted)-1]     // Truncate slice.
				break
			}
		}

		err := kvs.Set(guildID, "access", commandGroup, granted)
		if err != nil {
			log.Printf("[%s] /access revoke failed to store updated access list in KVS: %s\n", guildID, err)
			return ResponseEphemeral("An error occured, and has been logged.")
		}
	}
	return ResponseMessageNoMention(fmt.Sprintf("<@&%s> is denied access to the `%s` command group\n", roleID, commandGroup))
}

// SubCommandAccessList processes a sub command to list who has access to what.
func SubCommandAccessList(kvs storage.KeyValueStore, guildID discord.GuildID) api.InteractionResponse {
	var sb strings.Builder
	fmt.Fprintln(&sb, "Current access is:")
	for _, group := range commandGroups {
		granted := []discord.RoleID{}
		found, err := kvs.GetObject(guildID, "access", group, &granted)
		if err != nil {
			log.Printf("[%s] /access list failed to obtain access list from KVS: %s\n", guildID, err)
			return ResponseEphemeral("An error occured, and has been logged.")
		}
		fmt.Fprintf(&sb, "`%s`:", group)
		if !found || len(granted) == 0 {
			fmt.Fprintf(&sb, " Administrators only")
		} else {
			for _, role := range granted {
				fmt.Fprintf(&sb, " <@&%s>", role)
			}
		}
		fmt.Fprint(&sb, "\n")
	}
	return ResponseMessageNoMention(sb.String())
}
