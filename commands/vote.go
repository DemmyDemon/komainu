package commands

import (
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
		StartTime:    now,
		EndTime:      future,
		GuildID:      event.GuildID,
		MessageID:    discord.NullMessageID, // This is added in the MessageID callback later.
		Question:     command.Options[1].String(),
		Option1:      command.Options[2].String(),
		Option2:      command.Options[3].String(),
		Option1Votes: []discord.UserID{},
		Option2Votes: []discord.UserID{},
		Option3Votes: []discord.UserID{},
		Option4Votes: []discord.UserID{},
	}

	if len(command.Options) > 4 {
		vote.Option3 = command.Options[4].String()
	}
	if len(command.Options) == 6 {
		vote.Option4 = command.Options[5].String()
	}
	buttons := makeButtons(command.Options[2:len(command.Options)])

	return CommandResponse{
		api.InteractionResponse{
			Type: api.MessageInteractionWithSource,
			Data: &api.InteractionResponseData{
				Content:    option.NewNullableString(vote.Question),
				Components: discord.ComponentsPtr(buttons),
			},
		}, func(mi discord.MessageID) {
			vote.MessageID = mi
			// vote.Store(kvs)
		},
	}
}

var optionCustomIDs = []string{"first", "second", "third", "fourth", "broken"}

func makeButtons(options []discord.CommandInteractionOption) discord.Component {
	buttons := make([]discord.InteractiveComponent, len(options))
	for idx, option := range options {
		buttons[idx] = &discord.ButtonComponent{
			Style:    discord.PrimaryButtonStyle(),
			CustomID: discord.ComponentID(optionCustomIDs[idx]),
			Label:    option.String(),
		}
	}
	row := discord.ActionRowComponent(buttons)

	return &row
}
