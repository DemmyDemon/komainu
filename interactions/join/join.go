package join

import (
	"komainu/storage"

	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
)

type Handler struct {
	Code HandlerFunction
}

type HandlerFunction func(
	state *state.State,
	kvs storage.KeyValueStore,
	event *gateway.GuildMemberAddEvent,
)

var joinhandlers = []Handler{}

// Register makes the Code turn over when someone is arriving
func Register(handler Handler) {
	joinhandlers = append(joinhandlers, handler)
}

// Add the join handler to the given state
// This is mostly just pointless abstraction for uniformity across events.
func AddHandler(state *state.State, kvs storage.KeyValueStore) {
	state.AddHandler(func(event *gateway.GuildMemberAddEvent) {
		for _, handler := range joinhandlers {
			handler.Code(state, kvs, event)
		}
	})
}
