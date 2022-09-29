package storage

import (
	"github.com/diamondburned/arikawa/v3/discord"
)

type KeyValueStore interface {
	Close() error
	Set(guild discord.GuildID, collection string, key any, rawValue any) (err error)
	Get(guild discord.GuildID, collection string, key any, out any) (exist bool, err error)
	Delete(guild discord.GuildID, collection string, key any) (err error)
	Keys(guild discord.GuildID, collection string) (keys []string, err error)
}
