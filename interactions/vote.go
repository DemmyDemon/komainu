package interactions

import (
	"fmt"
	"komainu/interactions/command"
	"komainu/interactions/response"
	"komainu/storage"
	"log"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
)

func init() {
	command.Register("vote", commandVoteObject)
}

var commandVoteObject = command.Handler{
	Description: "Initiate a vote",
	Code:        CommandVote,
	Options: []discord.CommandOption{
		&discord.NumberOption{
			OptionName:  "length",
			Description: "The number of hours the vote should run.",
			Required:    true,
			Min:         option.NewFloat(0),
			Max:         option.NewFloat(336), // 336 hours is two weeks.
		},
		&discord.StringOption{
			OptionName:  "question",
			Description: "The question being asked. Works best as a yes/no question.",
			Required:    true,
		},
		&discord.StringOption{
			OptionName:  "first",
			Description: "The first vote option description (80 char max)",
			Required:    true,
		},
		&discord.StringOption{
			OptionName:  "second",
			Description: "The second vote option description (80 char max)",
			Required:    true,
		},
		&discord.StringOption{
			OptionName:  "third",
			Description: "The third vote option description (80 char max)",
			Required:    false,
		},
		&discord.StringOption{
			OptionName:  "fourth",
			Description: "The fourth vote option description (80 char max)",
			Required:    false,
		},
	},
}

// CommandVote processes a command to start a vote
func CommandVote(state *state.State, kvs storage.KeyValueStore, event *gateway.InteractionCreateEvent, cmd *discord.CommandInteraction) command.Response {
	if cmd.Options == nil || len(cmd.Options) < 4 {
		log.Printf("[%s] /vote command structure is somehow nil or not the correct number of elements. Wat.\n", event.GuildID)
		return command.Response{Response: response.Ephemeral("Yeah, no, that didn't work."), Callback: nil}
	}

	now := time.Now().Unix()
	hours, err := cmd.Options[0].FloatValue()
	if err != nil {
		log.Printf("[%s] /vote command structure is somehow weird. Could not get the Float value of the hours option.\n", event.GuildID)
		return command.Response{Response: response.Ephemeral("Wait, what? How many hours? Try again."), Callback: nil}
	}
	future := now + int64(hours*float64(3600))

	vote := storage.Vote{
		StartTime: now,
		EndTime:   future,
		GuildID:   event.GuildID,
		MessageID: discord.NullMessageID, // This is added in the MessageID callback later.
		ChannelID: discord.NullChannelID, // This one, too!
		Question:  cmd.Options[1].String(),
		Options:   map[string]string{},
		Votes:     map[discord.UserID]string{},
	}
	options := cmd.Options[2:len(cmd.Options)]
	for idx, val := range options {
		label := val.String()
		if len(label) > 80 {
			return command.Response{Response: response.Ephemeral("Sorry, the options can't be longer than 80 characters!"), Callback: nil}
		}
		vote.Options[fmt.Sprintf("voteOption%d", idx)] = val.String()
	}
	buttons := makeButtons(options)

	return command.Response{
		Response: api.InteractionResponse{
			Type: api.MessageInteractionWithSource,
			Data: &api.InteractionResponseData{
				Content:    option.NewNullableString(vote.String()),
				Components: discord.ComponentsPtr(buttons),
			},
		}, Callback: func(message *discord.Message) {
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
