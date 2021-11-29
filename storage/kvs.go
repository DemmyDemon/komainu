package storage

import "github.com/diamondburned/arikawa/v3/discord"

type KeyValueStore interface {
	Open(string) error
	Close() error
	Get(guild discord.GuildID, collection string, key interface{}) (bool, []byte, error)
	Set(guild discord.GuildID, collection string, key interface{}, rawValue interface{}) error
	GetObject(guild discord.GuildID, collection string, key interface{}, target interface{}) (bool, error)
	GetString(guild discord.GuildID, collection string, key interface{}) (bool, string, error)
	GetInt64(guild discord.GuildID, collection string, key interface{}) (bool, int64, error)
	GetUint64(guild discord.GuildID, collection string, key interface{}) (bool, uint64, error)
	Delete(guild discord.GuildID, collection string, key interface{}) (bool, error)
}
