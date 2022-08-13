package edit

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
	event *gateway.MessageUpdateEvent,
)

var editHandlers = []Handler{}

// Register makes the thing do the thing when the Edit happens
func Register(handler Handler) {
	editHandlers = append(editHandlers, handler)
}

// Add the edit handlers (Update, technically) to the given state
func AddHandler(state *state.State, kvs storage.KeyValueStore) {
	// Yes, this is abstraction just to keep the interface uniform with the Other Stuff.
	if state.PreHandler == nil {
		state.PreHandler = handler.New()
	}
	state.PreHandler.AddSyncHandler(func(event *gateway.MessageUpdateEvent) {
		for _, handler := range editHandlers {
			handler.Code(state, kvs, event)
		}
	})
}
