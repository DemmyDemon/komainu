package storage

import "github.com/diamondburned/arikawa/v3/discord"

type KeyValueStore interface {
	Open(string) error
	Close() error
	Set(guild discord.GuildID, collection string, key interface{}, rawValue interface{}) (err error)
	Get(guild discord.GuildID, collection string, key interface{}) (exist bool, value []byte, err error)
	GetObject(guild discord.GuildID, collection string, key interface{}, target interface{}) (exist bool, err error)
	GetString(guild discord.GuildID, collection string, key interface{}) (exist bool, value string, err error)
	GetInt64(guild discord.GuildID, collection string, key interface{}) (exist bool, value int64, err error)
	GetUint64(guild discord.GuildID, collection string, key interface{}) (exist bool, value uint64, err error)
	Delete(guild discord.GuildID, collection string, key interface{}) (wasDeleted bool, err error)
	Keys(guild discord.GuildID, collection string) (keys []string, err error)
}
