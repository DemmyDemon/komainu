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
func CommandFaq(state *state.State, sniper storage.KeyValueStore, event *gateway.InteractionCreateEvent, command *discord.CommandInteraction) api.InteractionResponse {
	if command.Options == nil || len(command.Options) != 1 {
		log.Printf("[%s] /faq command structure is somehow nil or not a single element. Wat.\n", event.GuildID)
		return ResponseMessage("Invalid command structure.")
	}
	topic := strings.ToLower(command.Options[0].String())
	exists, value, err := sniper.GetString(event.GuildID, "faq", topic)
	if err != nil {
		log.Printf("[%s] /faq failed to GetString the topic %s: %s", event.GuildID, topic, err)
		return ResponseMessage("An error occured, and has been logged.")
	}
	if !exists {
		return ResponseMessage(fmt.Sprintf("Sorry, I've never heard of %s", topic))
	}
	return ResponseMessage(value)
}

// CommandFaqSet processes commands to faff about in the topics list
func CommandFaqSet(state *state.State, sniper storage.KeyValueStore, event *gateway.InteractionCreateEvent, command *discord.CommandInteraction) api.InteractionResponse {
	if command.Options == nil || len(command.Options) != 1 {
		log.Printf("[%s] /faqset command structure is somehow nil or not a single element. Wat.\n", event.GuildID)
		return ResponseMessage("I'm sorry, what? Something very weird happened.")
	}
	switch command.Options[0].Name {
	case "list":
		return SubCommandFaqList(sniper, event.GuildID)
	case "add":
		return SubCommandFaqAdd(sniper, event.GuildID, command.Options[0].Options)
	case "remove":
		return SubCommandFaqRemove(sniper, event.GuildID, command.Options[0].Options)
	default:
		return ResponseMessage("Unknown subcommand! Clearly *someone* dropped the ball!")
	}
}

// SubCommandFaqAdd processes a subcommand to store a FAQ item.
func SubCommandFaqAdd(sniper storage.KeyValueStore, guildID discord.GuildID, options []discord.CommandInteractionOption) api.InteractionResponse {
	if options == nil || len(options) != 2 {
		log.Printf("[%s] /faqset add command structure is somehow nil or not two elements. Wat.\n", guildID)
		return ResponseMessage("Invalid command structure.")
	}
	key := strings.ToLower(options[0].String())
	value := options[1].String()
	err := sniper.Set(guildID, "faq", key, value)
	if err != nil {
		log.Printf("[%s] /faqset add storage failed: %s", guildID, err)
		return ResponseMessage("An error occured, and has been logged.")
	}

	faqList := []string{}
	_, err = sniper.GetObject(guildID, "faq", "LIST", &faqList)
	if err != nil {
		log.Printf("[%s] /faqset add storage is fine, but failed to get the LIST: %s", guildID, err)
		return ResponseMessage(fmt.Sprintf("Kinda learned %s: %s", key, value))
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
		err = sniper.Set(guildID, "faq", "LIST", faqList)
		if err != nil {
			log.Printf("[%s] /faqset add storage is fine, but failed to store the LIST: %s", guildID, err)
			return ResponseMessage(fmt.Sprintf("Sort of learned %s: %s", key, value))
		}
	}

	return ResponseMessage(fmt.Sprintf("Learned %s: %s", key, value))
}

// SubCommandFaqRemove processes a command to remove a FAQ item.
func SubCommandFaqRemove(sniper storage.KeyValueStore, guildID discord.GuildID, options []discord.CommandInteractionOption) api.InteractionResponse {
	if options == nil || len(options) != 1 {
		log.Printf("[%s] /faqset remove command structure is somehow nil or not one element. Wat.\n", guildID)
		return ResponseMessage("Invalid command structure.")
	}
	topic := strings.ToLower(options[0].String())
	//topic := command.Options[0].String()
	exists, value, err := sniper.GetString(guildID, "faq", topic)
	if err != nil {
		log.Printf("[%s] /faqset remove failed to GetString the topic %s: %s", guildID, topic, err)
		return ResponseMessage("An error occured, and has been logged.")
	}
	if !exists {
		return ResponseMessage(fmt.Sprintf("Sorry, I've never heard of %s", topic))
	}
	removed, err := sniper.Delete(guildID, "faq", topic)
	if err != nil {
		log.Printf("[%s] /faqset remove failed to Delete the topic %s: %s", guildID, topic, err)
		return ResponseMessage("An error occured, and has been logged.")
	}
	if !removed {
		// Is it even possible to get here?
		return ResponseMessage(fmt.Sprintf("Sorry, I've never heard of %s", topic))
	}

	faqList := []string{}
	_, err = sniper.GetObject(guildID, "faq", "LIST", &faqList)
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
	err = sniper.Set(guildID, "faq", "LIST", faqList)
	if err != nil {
		log.Printf("[%s] /faqset remove storage is fine, but failed to save the LIST: %s", guildID, err)
		return ResponseMessage(fmt.Sprintf("Kinda forgot %s: %s", topic, value))
	}

	return ResponseMessage(fmt.Sprintf("Forgot %s: %s", topic, value))
}

// SubCommandFaqList processes a subcommand to list all FAQ items.
func SubCommandFaqList(sniper storage.KeyValueStore, guildID discord.GuildID) api.InteractionResponse {
	faqList := []string{}
	exists, err := sniper.GetObject(guildID, "faq", "LIST", &faqList)
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
		return ResponseMessage(sb.String())
	}
	return ResponseMessage("I'm sad to say, there are no known topics. Maybe I just kinda sort of learned it, and didn't put anything in the list?")
}
