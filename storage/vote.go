package storage

import (
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
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
	for _, option := range vote.Votes {
		optionLabel := "!!UNKNOWN OPTION!!"
		if label, ok := vote.Options[option]; ok {
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

	for _, option := range keys {
		count := tally[option]
		fmt.Fprintf(&sb, "**%s** (%d votes)\n", option, count)
	}
	return sb.String()
}

func (vote *Vote) Buttons() *discord.ContainerComponents {
	buttons := []discord.InteractiveComponent{}
	for key, value := range vote.Options {
		button := &discord.ButtonComponent{
			Style:    discord.PrimaryButtonStyle(),
			CustomID: discord.ComponentID(key),
			Label:    value,
		}
		buttons = append(buttons, button)
	}
	row := discord.ActionRowComponent(buttons)
	return discord.ComponentsPtr(&row)
}

func (vote *Vote) Selector() *discord.ContainerComponents {
	selectable := []discord.SelectOption{}
	for key, label := range vote.Options {
		selectable = append(selectable, discord.SelectOption{
			Label: label,
			Value: key,
		})
	}
	row := discord.ActionRowComponent([]discord.InteractiveComponent{
		&discord.SelectComponent{
			Options:     selectable,
			CustomID:    "vote",
			Placeholder: "Cast your vote!",
			ValueLimits: [2]int{0, 1},
		},
	})
	return discord.ComponentsPtr(&row)
}

// GetVote gets a specific vote for the given guild and message. Returns a boolean to let you know if the vote exists, that Vote object if it does and any error that occured fetching it.
func GetVote(kvs KeyValueStore, guildID discord.GuildID, messageID discord.MessageID) (exist bool, vote *Vote, err error) {
	exist, err = kvs.Get(guildID, "votes", messageID, &vote)
	return exist, vote, err
}

// HandleInteractionAsVote determines if the given interaction is a vote button click, and acts accordingly.
func HandleInteractionAsVote(state *state.State, kvs KeyValueStore, e *gateway.InteractionCreateEvent, interaction discord.ComponentInteraction) (isVote bool, response string, err error) {
	exist, vote, err := GetVote(kvs, e.GuildID, e.Message.ID)
	if err != nil {
		return true, "Something very odd happened.", fmt.Errorf("handling interaction as vote: %w", err)
	}
	if !exist {
		return false, "", nil
	}

	now := time.Now().Unix()
	if vote.EndTime <= now {
		return true, "I'm sorry, that vote is closed!", nil
	}

	selector, ok := interaction.(*discord.SelectInteraction)

	if !ok {
		return true, "Your response was not in the right format, somehow?!", errors.New("submitted vote was not from a SelectInteraction")
	}

	if len(selector.Values) != 1 {
		return true, "You must select exactly one item", fmt.Errorf("%d values selected in vote, expected 1", len(selector.Values))
	}

	voted := selector.Values[0]

	label, ok := vote.Options[voted]
	if !ok {
		return true, "Sorry, you can't vote for that.", fmt.Errorf("vote cast for %s, which is not an option", voted)
	}

	vote.Votes[e.SenderID()] = voted
	if _, err := state.EditMessage(e.ChannelID, e.Message.ID, vote.String()); err != nil {
		return true, "There was an error registering your vote.", fmt.Errorf("handling interaction as vote: %w", err)
	}
	if err := vote.Store(kvs); err != nil {
		return true, "There was an error storing your vote.", fmt.Errorf("storing a vote: %w", err)
	}
	return true, fmt.Sprintf("Your vote for...\n%s\n...is registered.", label), nil

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
// Intednded to be called as a goroutine.
func StartClosingExpiredVotes(state *state.State, kvs KeyValueStore) {
	ticker := time.NewTicker(1 * time.Minute)
	for {
		<-ticker.C
		if err := CloseExpiredVotes(state, kvs); err != nil {
			log.Printf("Error encountered closing expired votes: %s", err)
		}
	}
}
