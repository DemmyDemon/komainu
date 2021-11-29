package storage

import (
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
)

// See saves the given user as being seen in the given guild.
func See(sniper KeyValueStore, guildID discord.GuildID, userID discord.UserID) error {
	return sniper.Set(guildID, "seen", userID, time.Now().Unix())
}

// LastSeen checks to see when the given user was seen in the given guild.
func LastSeen(sniper KeyValueStore, guildID discord.GuildID, userID discord.UserID) (bool, int64, error) {
	return sniper.GetInt64(guildID, "seen", userID)
}
