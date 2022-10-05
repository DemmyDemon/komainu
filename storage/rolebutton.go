package storage

import (
	"github.com/diamondburned/arikawa/v3/discord"
)

type RoleButton struct {
	RoleID    discord.RoleID
	GuildID   discord.GuildID
	MessageID discord.MessageID
}

func (rs *RoleButton) Store(kvs KeyValueStore, messageID discord.MessageID) error {
	rs.MessageID = messageID
	return kvs.Set(rs.GuildID, "rolebutton", messageID, rs)
}

func GetRoleForButton(kvs KeyValueStore, guildID discord.GuildID, messageID discord.MessageID) (exist bool, role discord.RoleID, err error) {
	button := RoleButton{}
	exist, err = kvs.Get(guildID, "rolebutton", messageID, &button)
	if err == nil && exist {
		role = button.RoleID
	}
	return
}
