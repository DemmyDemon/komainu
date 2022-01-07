package storage

import (
	"log"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
)

// Vote describes a vote attached to a Discord message.
type Vote struct {
	StartTime    int64
	EndTime      int64
	GuildID      discord.GuildID
	MessageID    discord.MessageID
	Question     string
	Option1      string
	Option2      string
	Option3      string
	Option4      string
	Option1Votes []discord.UserID
	Option2Votes []discord.UserID
	Option3Votes []discord.UserID
	Option4Votes []discord.UserID
}

func (vote *Vote) Store(kvs KeyValueStore) error {
	collection, err := GetVoteCollection(kvs, vote.GuildID)
	if err != nil {
		return err
	}
	collection.Votes[vote.MessageID] = *vote
	return collection.Store(kvs)
}

// VoteCollection holds all the Vote structs for a specific Guild.
type VoteCollection struct {
	GuildID discord.GuildID
	Votes   map[discord.MessageID]Vote
}

func (collection *VoteCollection) TrimExpired(kvs KeyValueStore) ([]Vote, error) {
	now := time.Now().Unix()
	trimmed := []Vote{}
	for key, vote := range collection.Votes {
		if vote.EndTime < now {
			trimmed = append(trimmed, vote)
			delete(collection.Votes, key)
		}
	}
	return trimmed, collection.Store(kvs)
}

func (collection *VoteCollection) Store(kvs KeyValueStore) error {
	return kvs.Set(collection.GuildID, "votes", "collection", collection)
}

func (collection *VoteCollection) DeleteVote(kvs KeyValueStore, messageID discord.MessageID) error {
	if _, ok := collection.Votes[messageID]; ok {
		delete(collection.Votes, messageID)
		return collection.Store(kvs)
	}
	return nil
}

// GetVote gets a specific vote for the given guild and message. Returns that vote (or an empty one) and any error that occured fetching it.
// The suggested way to test if a vote is valid, is to check if the MessageID property of the vote matches the one used to retrieve it.
func GetVote(kvs KeyValueStore, guildID discord.GuildID, messageID discord.MessageID) (Vote, error) {
	collection, err := GetVoteCollection(kvs, guildID)
	if err != nil {
		return Vote{}, err
	}
	if vote, ok := collection.Votes[messageID]; ok {
		return vote, nil
	}
	return Vote{}, nil
}

func GetVoteCollection(kvs KeyValueStore, guildID discord.GuildID) (VoteCollection, error) {

	collection := VoteCollection{
		GuildID: guildID,
		Votes:   make(map[discord.MessageID]Vote),
	}
	found, err := kvs.GetObject(guildID, "votes", "collection", &collection)
	if err != nil {
		log.Printf("[%s] Error fetching votes collection: %s", guildID, err)
		return collection, err
	}
	if !found {
		return collection, nil
	}

	return collection, nil

}
