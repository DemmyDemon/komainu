package interactions

import (
	"fmt"
	"komainu/interactions/command"
	"komainu/interactions/delete"
	"komainu/interactions/response"
	"komainu/storage"
	"log"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
)

const (
	deleteLogCollection = "deletelog"
	deleteLogKey        = "channel"
)

func init() {
	command.Register("deletelog", commandDeletelogObject)
	delete.Register(deleteLogHandler)
}

var commandDeletelogObject = command.Handler{
	Description: "Designate what channel to log deleted messages in",
	Code:        CommandDeletelog,
	Options: []discord.CommandOption{
		&discord.ChannelOption{
			OptionName:  "channel",
			Description: "Where to log deletions, blank to disable",
			Required:    false,
		},
	},
}

var deleteLogHandler = delete.Handler{
	Code: DeleteLogging,
}

func CommandDeletelog(state *state.State, kvs storage.KeyValueStore, event *gateway.InteractionCreateEvent, cmd *discord.CommandInteraction) command.Response {
	if cmd.Options == nil || len(cmd.Options) != 1 {
		log.Printf("[%s] <@%s> disabled delete log functionality", event.GuildID, event.SenderID())
		_, err := kvs.Delete(event.GuildID, deleteLogCollection, deleteLogKey)
		if err != nil {
			log.Printf("[%s] Failed to remove Delete Log Channel setting: %s", event.GuildID, err)
			return command.Response{Response: response.Ephemeral("Sorry, there was a hickup disabling the delete log functionality. The error was logged.")}
		}
		return command.Response{Response: response.Message("Okay, I will not log deletions.")}
	}
	channelSnowflake, err := cmd.Options[0].SnowflakeValue()
	if err != nil {
		log.Printf("[%s] Delete Log setting failed to get snowflake:  %s", event.GuildID, err)
		return command.Response{Response: response.Ephemeral("There was an issue setting the delete log channel. It has been logged.")}
	}
	channelId := discord.ChannelID(channelSnowflake)
	deleteLogChannel, err := state.Channel(channelId)
	if err != nil {
		log.Printf("[%s] Delete Log setting failed to get channel object: %s", event.GuildID, err)
		return command.Response{Response: response.Ephemeral("There was a problem setting the delete log channel. It has been logged.")}
	}

	kvs.Set(event.GuildID, deleteLogCollection, deleteLogKey, channelId)
	log.Printf("[%s] <@%s> set delete logging to <#%s>", event.GuildID, event.SenderID(), channelId)

	return command.Response{Response: response.Message(fmt.Sprintf("<#%s> is now the delete log channel", deleteLogChannel.ID))}
}

func DeleteLogging(state *state.State, kvs storage.KeyValueStore, event *gateway.MessageDeleteEvent) {
	deleteLogChannelID := discord.NullChannelID
	exist, err := kvs.GetObject(event.GuildID, deleteLogCollection, deleteLogKey, &deleteLogChannelID)
	if err != nil {
		log.Printf("[%s] Message deleted, but error looking up delete log channel ID: %s", event.GuildID, err)
		return
	}
	if !exist {
		return
	}
	message, err := state.Message(event.ChannelID, event.ID)
	if err != nil {
		_, sendErr := state.SendMessage(deleteLogChannelID, fmt.Sprintf("Unknown message %s in <#%s> was deleted. Originally posted <t:%d>", event.ID, event.ChannelID, event.ID.Time().Unix()))
		if sendErr != nil {
			log.Printf("[%s] UNKNOWN message deleted, error logging to delete log channel: %s", event.GuildID, err)
		}
		return
	}

	metaMessage := fmt.Sprintf("<@%s> had their message in <#%s> deleted. Originally posted <t:%d>", message.Author.ID, message.ChannelID, message.Timestamp.Time().Unix())

	color := discord.Color(0xFF0000)

	origialContent := message.Content
	if origialContent == "" {
		origialContent = "No actual message, maybe it was just an image?"
		color = discord.Color(0xFFFF00)
	}

	payload := []discord.Embed{
		{
			Type:        discord.NormalEmbed,
			Description: origialContent,
			Color:       color,
		},
	}

	if _, err := state.SendMessageComplex(deleteLogChannelID, api.SendMessageData{
		Content: metaMessage,
		Embeds:  payload,
		AllowedMentions: &api.AllowedMentions{
			Parse: []api.AllowedMentionType{},
		},
	}); err != nil {
		log.Printf("[%s] Message deleted, but error logging it: %s", event.GuildID, err)
	}
}
