package interactions

import (
	"fmt"
	"komainu/interactions/autocomplete"
	"komainu/interactions/command"
	"komainu/interactions/modal"
	"komainu/interactions/response"
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

func init() {
	command.Register("faq", commandFaqObject)
	command.Register("faqset", commandFaqSetObject)
	modal.Register("faqadd", modal.Handler{Code: FAQAddModalHandler})
	autocomplete.Register("faq", autocomplete.Handler{Code: FaqAutocomplete})
}

var commandFaqObject = command.Handler{
	Description: "Look up a FAQ topic",
	Code:        CommandFaq,
	Options: []discord.CommandOption{
		&discord.StringOption{
			OptionName:   "topic",
			Description:  "The name of the topic you wish to recall",
			Required:     true,
			Autocomplete: true,
		},
	},
}

var commandFaqSetObject = command.Handler{
	Description: "Manage FAQ topics",
	Code:        CommandFaqSet,
	Options: []discord.CommandOption{
		&discord.SubcommandOption{
			OptionName:  "add",
			Description: "Add a topic to the FAQ",
			Options: []discord.CommandOptionValue{
				&discord.StringOption{
					OptionName:  "topic",
					Description: "The word used to recall this item later",
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
	},
}

// CommandFaq processes a command to retrieve a FAQ item.
func CommandFaq(state *state.State, kvs storage.KeyValueStore, event *gateway.InteractionCreateEvent, cmd *discord.CommandInteraction) command.Response {
	if cmd.Options == nil || len(cmd.Options) != 1 {
		log.Printf("[%s] /faq command structure is somehow nil or not a single element. Wat.\n", event.GuildID)
		return command.Response{Response: response.Ephemeral("Invalid command structure."), Callback: nil}
	}
	topic := strings.ToLower(cmd.Options[0].String())
	value := ""
	exists, err := kvs.Get(event.GuildID, "faq", topic, &value)
	if err != nil {
		log.Printf("[%s] /faq failed to GetString the topic %s: %s", event.GuildID, topic, err)
		return command.Response{Response: response.Ephemeral("An error occured, and has been logged."), Callback: nil}
	}
	if !exists {
		return command.Response{Response: response.Ephemeral(fmt.Sprintf("Sorry, I've never heard of %s", topic)), Callback: nil}
	}
	return command.Response{Response: response.MessageNoMention(value), Callback: nil}
}

// CommandFaqSet processes commands to faff about in the topics list
func CommandFaqSet(state *state.State, kvs storage.KeyValueStore, event *gateway.InteractionCreateEvent, cmd *discord.CommandInteraction) command.Response {
	if cmd.Options == nil || len(cmd.Options) != 1 {
		log.Printf("[%s] /faqset command structure is somehow nil or not a single element. Wat.\n", event.GuildID)
		return command.Response{Response: response.Message("I'm sorry, what? Something very weird happened."), Callback: nil}
	}
	switch cmd.Options[0].Name {
	case "list":
		return command.Response{Response: SubCommandFaqList(kvs, event.GuildID), Callback: nil}
	case "add":
		return command.Response{Response: SubCommandFaqAdd(kvs, event.GuildID, event.SenderID(), cmd.Options[0].Options), Callback: nil}
	case "remove":
		return command.Response{Response: SubCommandFaqRemove(kvs, event.GuildID, cmd.Options[0].Options), Callback: nil}
	default:
		return command.Response{Response: response.Ephemeral("Unknown subcommand! Clearly *someone* dropped the ball!"), Callback: nil}
	}
}

// SubCommandFaqAdd processes a subcommand to store a FAQ item.
func SubCommandFaqAdd(kvs storage.KeyValueStore, guildID discord.GuildID, userID discord.UserID, options []discord.CommandInteractionOption) api.InteractionResponse {
	if options == nil || len(options) != 1 {
		log.Printf("[%s] /faqset add command structure is somehow not exactly one element. Wat.\n", guildID)
		return response.Ephemeral("Invalid command structure.")
	}
	key := strings.ToLower(options[0].String())
	value := ""
	_, err := kvs.Get(guildID, "faq", key, &value)
	if err != nil {
		log.Printf("[%s] /faqset add storage lookup failed: %s", guildID, err)
		return response.Ephemeral("An error occured, and has been logged.")
	}
	addOrUpdate := "Add FAQ item"
	if value != "" {
		addOrUpdate = "Update FAQ item"
	}

	return modal.Respond(
		userID, guildID, "faqadd", addOrUpdate,
		discord.TextInputComponent{
			CustomID:     discord.ComponentID(key),
			Label:        key,
			Value:        option.NewNullableString(value),
			Style:        discord.TextInputParagraphStyle,
			LengthLimits: [2]int{1, 1500},
		},
	)
}

// SubCommandFaqRemove processes a command to remove a FAQ item.
func SubCommandFaqRemove(kvs storage.KeyValueStore, guildID discord.GuildID, options []discord.CommandInteractionOption) api.InteractionResponse {
	if options == nil || len(options) != 1 {
		log.Printf("[%s] /faqset remove command structure is somehow nil or not one element. Wat.\n", guildID)
		return response.Ephemeral("Invalid command structure.")
	}
	topic := strings.ToLower(options[0].String())
	value := ""
	exists, err := kvs.Get(guildID, "faq", topic, &value)
	if err != nil {
		log.Printf("[%s] /faqset remove failed to GetString the topic %s: %s", guildID, topic, err)
		return response.Ephemeral("An error occured, and has been logged.")
	}
	if !exists {
		return response.Ephemeral(fmt.Sprintf("Sorry, I've never heard of %s", topic))
	}
	err = kvs.Delete(guildID, "faq", topic)
	if err != nil {
		log.Printf("[%s] /faqset remove failed to Delete the topic %s: %s", guildID, topic, err)
		return response.Ephemeral("An error occured, and has been logged.")
	}
	return response.MessageNoMention(fmt.Sprintf("Forgot %s: %s", topic, value))
}

// SubCommandFaqList processes a subcommand to list all FAQ items.
func SubCommandFaqList(kvs storage.KeyValueStore, guildID discord.GuildID) api.InteractionResponse {
	faqList, err := kvs.Keys(guildID, "faq")
	if err != nil {
		log.Printf("[%s] /faqset list failed to get the list: %s", guildID, err)
		return response.Message("An error occured, and has been logged.")
	}
	if len(faqList) > 0 {
		sort.Strings(faqList)
		var sb strings.Builder
		fmt.Fprintln(&sb, "**Here are the topics I know:**")
		for _, topic := range faqList {
			fmt.Fprintf(&sb, "- %s\n", utility.UcFirst(topic))
		}
		return response.MessageNoMention(sb.String())
	}
	return response.Ephemeral("I'm sad to say, there are no known topics.")
}

func FAQAddModalHandler(state *state.State, kvs storage.KeyValueStore, event *gateway.InteractionCreateEvent, interaction *discord.ModalInteraction) command.Response {
	data := modal.DecodeModalResponse(interaction.Components)
	for key, value := range data {
		err := kvs.Set(event.GuildID, "faq", key, value)
		if err != nil {
			log.Printf("[%s] Error storing FAQ item %q: %s", event.GuildID, key, err)
			return command.Response{Response: response.Ephemeral("There was an error saving that, but it has been logged!"), Callback: nil}
		}
		// Early return because we only expect one, but ranging over the one is the simplest code. *shrug*
		return command.Response{Response: response.MessageNoMention(fmt.Sprintf("Neat! I learned all about %q", key)), Callback: nil}
	}
	log.Printf("[%s] There was no data when trying to sote FAQ data?!  %#v", event.GuildID, interaction.Components)
	return command.Response{Response: response.Ephemeral("There was a weird problem, but don't worry! It has been logged for review."), Callback: nil}
}

func FaqAutocomplete(state *state.State, kvs storage.KeyValueStore, event *gateway.InteractionCreateEvent, interaction *discord.AutocompleteInteraction) api.AutocompleteChoices {
	choices := api.AutocompleteStringChoices{}
	found, value := autocomplete.GetAutocompleteValue(interaction)
	if !found {
		log.Printf("[%s] Could not determine autocomplete value from %#v", event.GuildID, interaction)
		return choices // Still empty at this point
	}

	typed := strings.ToLower(value.String())
	typed = strings.ReplaceAll(typed, "\"", "") // Because the value is quoted, for some damn reason.

	keys, err := kvs.Keys(event.GuildID, "faq")
	if err != nil {
		log.Printf("[%s] Error looking up FAQ keys: %s", event.GuildID, err)
		return choices // Still empty at this point.
	}

	for _, key := range keys {
		if strings.HasPrefix(key, typed) {
			choices = append(choices, discord.StringChoice{Name: utility.UcFirst(key), Value: key})
		}
	}

	return choices
}
