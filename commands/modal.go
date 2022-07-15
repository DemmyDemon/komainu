//go:build modaltest

package commands

import (
	"komainu/storage"
	"log"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
)

func init() {
	registerCommandObject("modal", CommandModalObject)
	registerModalHandlerObject("modal", ModalTestModalHandlerObject)
}

var CommandModalObject Command = Command{
	group:       "faq",
	description: "Testing some Modal stuff",
	code:        CommandModal,
	options: []discord.CommandOption{
		&discord.StringOption{
			OptionName:  "title",
			Description: "The title of the modal, or whatever",
			Required:    false,
		},
		&discord.StringOption{
			OptionName:  "body",
			Description: "The body of the modal, or whatever",
			Required:    false,
		},
	},
}

var ModalTestModalHandlerObject ModalHandler = ModalHandler{
	code: func(state *state.State, kvs storage.KeyValueStore, event *gateway.InteractionCreateEvent, interaction *discord.ModalInteraction) CommandResponse {
		log.Printf("Modal handler got:  %#v", interaction)
		foo := DecodeModalResponse(interaction.Components)
		log.Printf("%#v", foo)
		return CommandResponse{
			ResponseEphemeral("It got here!"),
			nil,
		}
	},
}

func CommandModal(state *state.State, kvs storage.KeyValueStore, event *gateway.InteractionCreateEvent, command *discord.CommandInteraction) CommandResponse {
	//modalStuff := foo(command)
	title := ""
	body := ""
	if command.Options != nil {
		if len(command.Options) >= 1 {
			title = command.Options[0].String()
		}
		if len(command.Options) >= 2 {
			body = command.Options[1].String()
		}
	}

	return CommandResponse{
		ResponseModal(
			event.SenderID(),
			event.GuildID, "modal", "Modal testing",
			discord.TextInputComponent{
				CustomID: "title",
				Label:    "Title label",
				Required: true,
				Value:    option.NewNullableString(title),
				Style:    discord.TextInputShortStyle,
			},
			discord.TextInputComponent{
				CustomID: "body",
				Label:    "Body label",
				Required: true,
				Value:    option.NewNullableString(body),
				Style:    discord.TextInputParagraphStyle,
			},
		), nil,
	}
}
