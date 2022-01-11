package commands

import (
	"fmt"
	"komainu/storage"
	"log"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
)

// CommandVote processes a command to start a vote
func CommandVote(state *state.State, kvs storage.KeyValueStore, event *gateway.InteractionCreateEvent, command *discord.CommandInteraction) CommandResponse {
	if command.Options == nil || len(command.Options) < 4 {
		log.Printf("[%s] /vote command structure is somehow nil or not the correct number of elements. Wat.\n", event.GuildID)
		return CommandResponse{ResponseEphemeral("Yeah, no, that didn't work."), nil}
	}

	now := time.Now().Unix()
	hours, err := command.Options[0].FloatValue()
	if err != nil {
		log.Printf("[%s] /vote command structure is somehow weird. Could not get the Float value of the hours option.\n", event.GuildID)
		return CommandResponse{ResponseEphemeral("Wait, what? How many hours? Try again."), nil}
	}
	future := now + int64(hours*float64(3600))

	vote := storage.Vote{
		StartTime: now,
		EndTime:   future,
		GuildID:   event.GuildID,
		MessageID: discord.NullMessageID, // This is added in the MessageID callback later.
		ChannelID: discord.NullChannelID, // This one, too!
		Question:  command.Options[1].String(),
		Options:   map[string]string{},
		Votes:     map[discord.UserID]string{},
	}
	options := command.Options[2:len(command.Options)]
	for idx, val := range options {
		label := val.String()
		if len(label) > 80 {
			return CommandResponse{ResponseEphemeral("Sorry, the options can't be longer than 80 characters!"), nil}
		}
		vote.Options[fmt.Sprintf("voteOption%d", idx)] = val.String()
	}
	buttons := makeButtons(options)

	return CommandResponse{
		api.InteractionResponse{
			Type: api.MessageInteractionWithSource,
			Data: &api.InteractionResponseData{
				Content:    option.NewNullableString(vote.String()),
				Components: discord.ComponentsPtr(buttons),
			},
		}, func(message *discord.Message) {
			vote.MessageID = message.ID
			vote.ChannelID = message.ChannelID
			vote.Store(kvs)
		},
	}
}

func makeButtons(options []discord.CommandInteractionOption) discord.Component {
	buttons := make([]discord.InteractiveComponent, len(options))
	for idx, option := range options {
		buttons[idx] = &discord.ButtonComponent{
			Style:    discord.PrimaryButtonStyle(),
			CustomID: discord.ComponentID(fmt.Sprintf("voteOption%d", idx)),
			Label:    option.String(),
		}
	}
	row := discord.ActionRowComponent(buttons)

	return &row
}
