package leave

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
	event *gateway.GuildMemberRemoveEvent,
)

var leavehandlers = []Handler{}

// Register makes the Code spin when someone wanders off
func Register(handler Handler) {
	leavehandlers = append(leavehandlers, handler)
}

// Add the join handler to the given state
// This is mostly just pointless abstraction for uniformity across events.
func AddHandler(state *state.State, kvs storage.KeyValueStore) {
	state.AddHandler(func(event *gateway.GuildMemberRemoveEvent) {
		for _, handler := range leavehandlers {
			handler.Code(state, kvs, event)
		}
	})
}
