package storage

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
)

// Vote describes a vote attached to a Discord message.
type Vote struct {
	StartTime int64
	EndTime   int64
	GuildID   discord.GuildID
	ChannelID discord.ChannelID
	MessageID discord.MessageID
	Question  string
	Order     []string
	Options   map[string]string
	Votes     map[discord.UserID]string
}

// Store saves the vote struct to kvs
func (vote *Vote) Store(kvs KeyValueStore) error {
	return kvs.Set(vote.GuildID, "votes", vote.MessageID, vote)
}

// Tally returns the current vote tally along with a slice containing the different vote options for easy sorting.
func (vote *Vote) Tally() (tally map[string]int, keys []string) {
	tally = map[string]int{}
	for _, key := range vote.Order {
		label := vote.Options[key]
		tally[label] = 0
		keys = append(keys, label)
	}
	for _, opt := range vote.Votes {
		optionLabel := "!!UNKNOWN OPTION!!"
		if label, ok := vote.Options[opt]; ok {
			optionLabel = label
		}
		if _, ok := tally[optionLabel]; ok {
			tally[optionLabel] = tally[optionLabel] + 1
		} else {
			tally[optionLabel] = 1
		}
	}
	return tally, keys
}

// String returns the vote as a string, which means formatting it as suitable as a Discord message.
func (vote *Vote) String() (voteText string) {
	var sb strings.Builder
	now := time.Now().Unix()
	tally, keys := vote.Tally()

	if vote.EndTime <= now {
		fmt.Fprintf(&sb, "%s\n\nVoting closed <t:%d:R>.\n\n", vote.Question, vote.EndTime)
		// If voting has ended, we rank them by score.
		sort.SliceStable(keys, func(i int, j int) bool {
			return tally[keys[i]] > tally[keys[j]]
		})
	} else {
		fmt.Fprintf(&sb, "%s\n\nCloses <t:%d:R>.\n\n", vote.Question, vote.EndTime)
	}

	for _, opt := range keys {
		count := tally[opt]
		plural := "s"
		if count == 1 {
			plural = ""
		}
		fmt.Fprintf(&sb, "**%s** (%d vote%s)\n", opt, count, plural)
	}
	return sb.String()
}

// GetVote gets a specific vote for the given guild and message. Returns a boolean to let you know if the vote exists, that Vote object if it does and any error that occured fetching it.
func GetVote(kvs KeyValueStore, guildID discord.GuildID, messageID discord.MessageID) (exist bool, vote *Vote, err error) {
	exist, err = kvs.Get(guildID, "votes", messageID, &vote)
	return exist, vote, err
}

// CloseExpiredVotes iterates over all the known votes in the connected guilds, and closes the ended ones.
func CloseExpiredVotes(state *state.State, kvs KeyValueStore) error {
	guilds, err := state.Guilds()
	if err != nil {
		return fmt.Errorf("closing expired votes could not fetch current guilds: %w", err)
	}
	now := time.Now().Unix()
	for _, guild := range guilds {
		keys, err := kvs.Keys(guild.ID, "votes")
		if err != nil {
			return fmt.Errorf("closing expired votes could not get keys for guild: %w", err)
		}
		for _, key := range keys {
			vote := Vote{}
			exist, err := kvs.Get(guild.ID, "votes", key, &vote)
			if err != nil {
				return fmt.Errorf("closing expired votes could not obtain vote object: %w", err)
			}
			if exist {
				if vote.ChannelID == discord.NullChannelID || vote.ChannelID == 0 {
					log.Printf("[%s] Closing expired votes encountered vote with no channel ID -- PURGING", guild.ID)
					err := kvs.Delete(guild.ID, "votes", key)
					if err != nil {
						log.Printf("[%s] Error purging invalid vote: %s\n", guild.ID, err)
					}
					continue
				}
				if vote.EndTime <= now {
					_, err := state.EditMessageComplex(vote.ChannelID, vote.MessageID, api.EditMessageData{
						Content:    option.NewNullableString(vote.String()),
						Components: &discord.ContainerComponents{},
					})
					if err != nil {
						return fmt.Errorf("closing expired votes could not update vote message: %w", err)
					}
					err = kvs.Delete(guild.ID, "votes", key)
					if err != nil {
						return fmt.Errorf("encoutered an error removing expired vote: %w", err)
					}
				}
			}
		}
	}
	return nil
}

// StartClosingExpiredVotes starts a ticker and, once a minute, calls CloseExpiredVotes.
// Intended to be called as a goroutine.
func StartClosingExpiredVotes(state *state.State, kvs KeyValueStore) {
	ticker := time.NewTicker(1 * time.Minute)
	for {
		<-ticker.C
		if err := CloseExpiredVotes(state, kvs); err != nil {
			log.Printf("Error encountered closing expired votes: %s", err)
		}
	}
}
