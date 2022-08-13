package delete

import (
	"komainu/storage"

	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/diamondburned/arikawa/v3/utils/handler"
)

type Handler struct {
	Code HandlerFunction
}

type HandlerFunction func(
	state *state.State,
	kvs storage.KeyValueStore,
	event *gateway.MessageDeleteEvent,
)

var deleteHandlers = []Handler{}

// Register makes the Code go brrr when a message dies
func Register(handler Handler) {
	deleteHandlers = append(deleteHandlers, handler)
}

// Add the deletion handler to the given state
// TODO: Figure out a way to make this more than just pointless abstraction
func AddHandler(state *state.State, kvs storage.KeyValueStore) {
	if state.PreHandler == nil {
		state.PreHandler = handler.New()
	}
	state.PreHandler.AddSyncHandler(func(event *gateway.MessageDeleteEvent) {
		for _, handler := range deleteHandlers {
			handler.Code(state, kvs, event)
		}
	})
}
