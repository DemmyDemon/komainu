package storage

import (
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
)

func See(guildID discord.GuildID, userID discord.UserID) error {
	return Sniper().Set(guildID, "seen", userID, time.Now().Unix())
}

func LastSeen(guildID discord.GuildID, userID discord.UserID) (bool, int64, error) {
	return Sniper().GetInt64(guildID, "seen", userID)
}
