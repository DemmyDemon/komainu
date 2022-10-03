package storage

import (
	"github.com/diamondburned/arikawa/v3/discord"
)

type RoleSelector struct {
	Roles     map[discord.RoleID]bool
	GuildID   discord.GuildID
	MessageID discord.MessageID
}

func (rs *RoleSelector) Store(kvs KeyValueStore, messageID discord.MessageID) error {
	rs.MessageID = messageID
	return kvs.Set(rs.GuildID, "roleselector", messageID, rs)
}

func (rs *RoleSelector) Has(roleID discord.RoleID) bool {
	if _, ok := rs.Roles[roleID]; ok {
		return true
	}
	return false
}

func GetRoleSelector(kvs KeyValueStore, guildID discord.GuildID, messageID discord.MessageID) (exist bool, selector RoleSelector, err error) {
	exist, err = kvs.Get(guildID, "roleselector", messageID, &selector)
	return
}
