package interactions

import (
	"fmt"
	"komainu/interactions/command"
	"komainu/interactions/join"
	"komainu/interactions/leave"
	"komainu/interactions/response"
	"komainu/storage"
	"log"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
)

const (
	trafficLogCollection = "trafficlog"
	trafficLogKey        = "channel"
)

func init() {
	command.Register("trafficlog", commandTrafficLogObject)
	join.Register(join.Handler{Code: joinLogging})
	leave.Register(leave.Handler{Code: leaveLogging})
}

var commandTrafficLogObject = command.Handler{
	Description: "Designate what channel to log joining and leaving in",
	Code:        CommandTrafficLog,
	Options: []discord.CommandOption{
		&discord.ChannelOption{
			OptionName:  "channel",
			Description: "Where to log when someone joins or leaves. Blank to disable.",
			Required:    false,
		},
	},
}

func CommandTrafficLog(state *state.State, kvs storage.KeyValueStore, event *gateway.InteractionCreateEvent, cmd *discord.CommandInteraction) command.Response {
	if cmd.Options == nil || len(cmd.Options) != 1 {
		log.Printf("[%s] <@%s> disabled traffic log functionality", event.GuildID, event.SenderID())
		err := kvs.Delete(event.GuildID, trafficLogCollection, trafficLogKey)
		if err != nil {
			log.Printf("[%s] Failed to remove Traffic Log Channel setting: %s", event.GuildID, err)
			return command.Response{Response: response.Ephemeral("Sorry, there was a hickup disabling the traffic log functionality. The error was logged.")}
		}
		return command.Response{Response: response.Message("Okay, I will not log join and leave traffic.")}
	}
	channelSnowflake, err := cmd.Options[0].SnowflakeValue()
	if err != nil {
		log.Printf("[%s] Traffic  Log setting failed to get snowflake:  %s", event.GuildID, err)
		return command.Response{Response: response.Ephemeral("There was an issue setting the traffic log channel. It has been logged.")}
	}
	channelId := discord.ChannelID(channelSnowflake)
	trafficLogChannel, err := state.Channel(channelId)
	if err != nil {
		log.Printf("[%s] Traffic Log setting failed to get channel object: %s", event.GuildID, err)
		return command.Response{Response: response.Ephemeral("There was a problem setting the traffic log channel. It has been logged.")}
	}

	kvs.Set(event.GuildID, trafficLogCollection, trafficLogKey, channelId)
	log.Printf("[%s] <@%s> set Traffic logging to <#%s>", event.GuildID, event.SenderID(), channelId)

	return command.Response{Response: response.Message(fmt.Sprintf("<#%s> is now the traffic log channel", trafficLogChannel.ID))}
}

func getTrafficLogChannel(kvs storage.KeyValueStore, guildID discord.GuildID) (exist bool, channel discord.ChannelID) {
	channel = discord.NullChannelID
	exist, err := kvs.Get(guildID, trafficLogCollection, trafficLogKey, &channel)
	if err != nil {
		log.Printf("[%s] Failed to obtain traffic logging channel: %s", guildID, err)
		exist = false // Just making sure, mm-kay?
	}
	return
}

func joinLogging(state *state.State, kvs storage.KeyValueStore, event *gateway.GuildMemberAddEvent) {
	exist, channelID := getTrafficLogChannel(kvs, event.GuildID)
	if !exist {
		return
	}
	_, err := state.SendMessageComplex(channelID, api.SendMessageData{
		Content: fmt.Sprintf("%s (%s#%s) has joined the server", event.Member.User.Mention(), event.User.Username, event.User.Discriminator),
		AllowedMentions: &api.AllowedMentions{
			Parse: []api.AllowedMentionType{},
		},
	})
	if err != nil {
		log.Printf("[%s] There was a join event, but there was an error logging it: %s", event.GuildID, err)
	}
}

func leaveLogging(state *state.State, kvs storage.KeyValueStore, event *gateway.GuildMemberRemoveEvent) {
	exist, channelID := getTrafficLogChannel(kvs, event.GuildID)
	if !exist {
		return
	}
	_, err := state.SendMessageComplex(channelID, api.SendMessageData{
		Content: fmt.Sprintf("%s (%s#%s) has left the server", event.User.ID.Mention(), event.User.Username, event.User.Discriminator),
		AllowedMentions: &api.AllowedMentions{
			Parse: []api.AllowedMentionType{},
		},
	})
	if err != nil {
		log.Printf("[%s] There was a leave event, but there was an error logging it: %s", event.GuildID, err)
	}
}
