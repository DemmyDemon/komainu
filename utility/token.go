package utility

import (
	"log"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
)

type Token struct {
	Interval int
	count    int
	ticker   *time.Ticker
}

// Increment increments the count of the token, and starts a ticker to count it back down, as needed.
func (token *Token) Increment() {
	token.count++
	if token.ticker == nil {
		interval := 10
		if token.Interval > 0 {
			interval = token.Interval
		}
		token.ticker = time.NewTicker(time.Duration(interval) * time.Second)
		go func() {
			for { // ever
				<-token.ticker.C // Just wait for it to tick, we don't care what it returns.
				if token.count > 0 {
					log.Printf("Count now %d\n", token.count)
					token.count--
				} else {
					token.ticker.Stop()
					token.ticker = nil
					break
				}
			}
		}()
	}
}

// GetCount returns the current count on the token.
func (token *Token) GetCount() int {
	return token.count
}

type TokenBin struct {
	Max      int
	Interval int
	tokens   map[discord.Snowflake]map[discord.Snowflake]*Token
}

// ensureTokenExists makes sure the pile and key exists, and has a Token on it with the correct Interval
func (tb *TokenBin) ensureTokenExists(pile discord.Snowflake, key discord.Snowflake) {
	if tb.tokens == nil {
		tb.tokens = map[discord.Snowflake]map[discord.Snowflake]*Token{}
	}
	if _, ok := tb.tokens[pile]; !ok {
		tb.tokens[pile] = map[discord.Snowflake]*Token{}
	}
	if _, ok := tb.tokens[pile][key]; !ok {
		tb.tokens[pile][key] = &Token{Interval: tb.Interval}
	}
}

// Allocate will return a true value if the token has not reached it's maximum value, false if it has.
func (tb *TokenBin) Allocate(pile discord.Snowflake, key discord.Snowflake) bool {
	tb.ensureTokenExists(pile, key)
	if token, ok := tb.tokens[pile][key]; ok {
		if token.GetCount() < tb.Max {
			token.Increment()
			return true
		}
	}
	return false
}
