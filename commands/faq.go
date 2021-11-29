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

// CommandFaqOn processes a command to store a FAQ item.
func CommandFaqOn(state *state.State, sniper storage.KeyValueStore, event *gateway.InteractionCreateEvent, command *discord.CommandInteraction) api.InteractionResponse {
	if command.Options == nil || len(command.Options) != 2 {
		log.Printf("[%s] /faqon command structure is somehow nil or not two elements. Wat.\n", event.GuildID)
		return ResponseMessage("Invalid command structure.")
	}
	key := strings.ToLower(command.Options[0].String())
	value := command.Options[1].String()
	err := sniper.Set(event.GuildID, "faq", key, value)
	if err != nil {
		log.Printf("[%s] /faqon storage failed: %s", event.GuildID, err)
		return ResponseMessage("An error occured, and has been logged.")
	}

	faqList := []string{}
	_, err = sniper.GetObject(event.GuildID, "faq", "LIST", &faqList)
	if err != nil {
		log.Printf("[%s] /faqon storage is fine, but failed to get the LIST: %s", event.GuildID, err)
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
		err = sniper.Set(event.GuildID, "faq", "LIST", faqList)
		if err != nil {
			log.Printf("[%s] /faqon storage is fine, but failed to store the LIST: %s", event.GuildID, err)
			return ResponseMessage(fmt.Sprintf("Sort of learned %s: %s", key, value))
		}
	}

	return ResponseMessage(fmt.Sprintf("Learned %s: %s", key, value))
}

// CommandFaqOff processes a command to remove a FAQ item.
func CommandFaqOff(state *state.State, sniper storage.KeyValueStore, event *gateway.InteractionCreateEvent, command *discord.CommandInteraction) api.InteractionResponse {
	if command.Options == nil || len(command.Options) != 1 {
		log.Printf("[%s] /faqoff command structure is somehow nil or not one element. Wat.\n", event.GuildID)
		return ResponseMessage("Invalid command structure.")
	}
	topic := strings.ToLower(command.Options[0].String())
	//topic := command.Options[0].String()
	exists, value, err := sniper.GetString(event.GuildID, "faq", topic)
	if err != nil {
		log.Printf("[%s] /faqoff failed to GetString the topic %s: %s", event.GuildID, topic, err)
		return ResponseMessage("An error occured, and has been logged.")
	}
	if !exists {
		return ResponseMessage(fmt.Sprintf("Sorry, I've never heard of %s", topic))
	}
	removed, err := sniper.Delete(event.GuildID, "faq", topic)
	if err != nil {
		log.Printf("[%s] /faqoff failed to Delete the topic %s: %s", event.GuildID, topic, err)
		return ResponseMessage("An error occured, and has been logged.")
	}
	if !removed {
		// Is it even possible to get here?
		return ResponseMessage(fmt.Sprintf("Sorry, I've never heard of %s", topic))
	}

	faqList := []string{}
	_, err = sniper.GetObject(event.GuildID, "faq", "LIST", &faqList)
	if err != nil {
		log.Printf("[%s] /faqoff storage is fine, but failed to get the LIST: %s", event.GuildID, err)
	}
	for idx, item := range faqList {
		if item == topic {
			faqList[idx] = faqList[len(faqList)-1] // Copy last element to index idx.
			faqList = faqList[:len(faqList)-1]     // Truncate slice.
			break
		}
	}
	err = sniper.Set(event.GuildID, "faq", "LIST", faqList)
	if err != nil {
		log.Printf("[%s] /faqoff storage is fine, but failed to save the LIST: %s", event.GuildID, err)
		return ResponseMessage(fmt.Sprintf("Kinda forgot %s: %s", topic, value))
	}

	return ResponseMessage(fmt.Sprintf("Forgot %s: %s", topic, value))
}

// CommandFaqList processes a command to list all FAQ items.
func CommandFaqList(state *state.State, sniper storage.KeyValueStore, event *gateway.InteractionCreateEvent, command *discord.CommandInteraction) api.InteractionResponse {
	faqList := []string{}
	exists, err := sniper.GetObject(event.GuildID, "faq", "LIST", &faqList)
	if err != nil {
		log.Printf("[%s] /faqlist failed to get the LIST: %s", event.GuildID, err)
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
