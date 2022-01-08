package commands

import (
	"fmt"
	"komainu/storage"
	"komainu/utility"
	"log"
	"sort"
	"strings"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
)

// CommandFaq processes a command to retrieve a FAQ item.
func CommandFaq(state *state.State, kvs storage.KeyValueStore, event *gateway.InteractionCreateEvent, command *discord.CommandInteraction) CommandResponse {
	if command.Options == nil || len(command.Options) != 1 {
		log.Printf("[%s] /faq command structure is somehow nil or not a single element. Wat.\n", event.GuildID)
		return CommandResponse{ResponseEphemeral("Invalid command structure."), nil}
	}
	topic := strings.ToLower(command.Options[0].String())
	exists, value, err := kvs.GetString(event.GuildID, "faq", topic)
	if err != nil {
		log.Printf("[%s] /faq failed to GetString the topic %s: %s", event.GuildID, topic, err)
		return CommandResponse{ResponseEphemeral("An error occured, and has been logged."), nil}
	}
	if !exists {
		return CommandResponse{ResponseEphemeral(fmt.Sprintf("Sorry, I've never heard of %s", topic)), nil}
	}
	return CommandResponse{ResponseMessageNoMention(value), nil}
}

// CommandFaqSet processes commands to faff about in the topics list
func CommandFaqSet(state *state.State, kvs storage.KeyValueStore, event *gateway.InteractionCreateEvent, command *discord.CommandInteraction) CommandResponse {
	if command.Options == nil || len(command.Options) != 1 {
		log.Printf("[%s] /faqset command structure is somehow nil or not a single element. Wat.\n", event.GuildID)
		return CommandResponse{ResponseMessage("I'm sorry, what? Something very weird happened."), nil}
	}
	switch command.Options[0].Name {
	case "list":
		return CommandResponse{SubCommandFaqList(kvs, event.GuildID), nil}
	case "add":
		return CommandResponse{SubCommandFaqAdd(kvs, event.GuildID, command.Options[0].Options), nil}
	case "remove":
		return CommandResponse{SubCommandFaqRemove(kvs, event.GuildID, command.Options[0].Options), nil}
	default:
		return CommandResponse{ResponseEphemeral("Unknown subcommand! Clearly *someone* dropped the ball!"), nil}
	}
}

// SubCommandFaqAdd processes a subcommand to store a FAQ item.
func SubCommandFaqAdd(kvs storage.KeyValueStore, guildID discord.GuildID, options []discord.CommandInteractionOption) api.InteractionResponse {
	if options == nil || len(options) != 2 {
		log.Printf("[%s] /faqset add command structure is somehow nil or not two elements. Wat.\n", guildID)
		return ResponseEphemeral("Invalid command structure.")
	}
	key := strings.ToLower(options[0].String())
	value := options[1].String()
	err := kvs.Set(guildID, "faq", key, value)
	if err != nil {
		log.Printf("[%s] /faqset add storage failed: %s", guildID, err)
		return ResponseEphemeral("An error occured, and has been logged.")
	}

	faqList := []string{}
	_, err = kvs.GetObject(guildID, "faq", "LIST", &faqList)
	if err != nil {
		log.Printf("[%s] /faqset add storage is fine, but failed to get the LIST: %s", guildID, err)
		return ResponseMessageNoMention(fmt.Sprintf("Kinda learned %s: %s", key, value))
	}

	found := false
	for _, candidate := range faqList {
		if candidate == key {
			found = true
			break
		}
	}

	if !found {
		faqList = append(faqList, key)
		err = kvs.Set(guildID, "faq", "LIST", faqList)
		if err != nil {
			log.Printf("[%s] /faqset add storage is fine, but failed to store the LIST: %s", guildID, err)
			return ResponseMessageNoMention(fmt.Sprintf("Sort of learned %s: %s", key, value))
		}
	}

	return ResponseMessageNoMention(fmt.Sprintf("Learned %s: %s", key, value))
}

// SubCommandFaqRemove processes a command to remove a FAQ item.
func SubCommandFaqRemove(kvs storage.KeyValueStore, guildID discord.GuildID, options []discord.CommandInteractionOption) api.InteractionResponse {
	if options == nil || len(options) != 1 {
		log.Printf("[%s] /faqset remove command structure is somehow nil or not one element. Wat.\n", guildID)
		return ResponseEphemeral("Invalid command structure.")
	}
	topic := strings.ToLower(options[0].String())
	//topic := command.Options[0].String()
	exists, value, err := kvs.GetString(guildID, "faq", topic)
	if err != nil {
		log.Printf("[%s] /faqset remove failed to GetString the topic %s: %s", guildID, topic, err)
		return ResponseEphemeral("An error occured, and has been logged.")
	}
	if !exists {
		return ResponseEphemeral(fmt.Sprintf("Sorry, I've never heard of %s", topic))
	}
	removed, err := kvs.Delete(guildID, "faq", topic)
	if err != nil {
		log.Printf("[%s] /faqset remove failed to Delete the topic %s: %s", guildID, topic, err)
		return ResponseEphemeral("An error occured, and has been logged.")
	}
	if !removed {
		// Is it even possible to get here?
		return ResponseEphemeral(fmt.Sprintf("Sorry, I've never heard of %s", topic))
	}

	faqList := []string{}
	_, err = kvs.GetObject(guildID, "faq", "LIST", &faqList)
	if err != nil {
		log.Printf("[%s] /faqset remove storage is fine, but failed to get the LIST: %s", guildID, err)
	}
	for idx, item := range faqList {
		if item == topic {
			faqList[idx] = faqList[len(faqList)-1] // Copy last element to index idx.
			faqList = faqList[:len(faqList)-1]     // Truncate slice.
			break
		}
	}
	err = kvs.Set(guildID, "faq", "LIST", faqList)
	if err != nil {
		log.Printf("[%s] /faqset remove storage is fine, but failed to save the LIST: %s", guildID, err)
		return ResponseMessageNoMention(fmt.Sprintf("Kinda forgot %s: %s", topic, value))
	}

	return ResponseMessageNoMention(fmt.Sprintf("Forgot %s: %s", topic, value))
}

// SubCommandFaqList processes a subcommand to list all FAQ items.
func SubCommandFaqList(kvs storage.KeyValueStore, guildID discord.GuildID) api.InteractionResponse {
	faqList := []string{}
	exists, err := kvs.GetObject(guildID, "faq", "LIST", &faqList)
	if err != nil {
		log.Printf("[%s] /faqset list failed to get the LIST: %s", guildID, err)
		return ResponseMessage("An error occured, and has been logged.")
	}
	if exists && len(faqList) > 0 {
		sort.Strings(faqList)
		var sb strings.Builder
		fmt.Fprintln(&sb, "**Here are the topics I know:**")
		for _, topic := range faqList {
			fmt.Fprintf(&sb, "- %s\n", utility.UcFirst(topic))
		}
		return ResponseMessageNoMention(sb.String())
	}
	return ResponseEphemeral("I'm sad to say, there are no known topics. Maybe I just kinda sort of learned it, and didn't put anything in the list?")
}
