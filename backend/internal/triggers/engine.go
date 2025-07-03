package triggers

import (
	"no-heroes-no-lies/internal/models"
)

// Event carries runtime data for any in-game occurrence.
type Event struct {
	Type     string                 // e.g. "steal_attempt"
	ActorID  string                 // player who caused it
	TargetID string                 // optional target
	Amount   int                    // generic numeric field (gems, coins…)
	Meta     map[string]interface{} // free-form extra data
	Cancel   bool                   // set by handler to veto default action
}

// Handler signature.
type Handler func(
	session *models.GameSession,
	ev *Event,
)

// registry maps event type → slice of handlers.
var registry = map[string][]Handler{}

// Register a handler for a given event type.
func Register(event string, fn Handler) {
	registry[event] = append(registry[event], fn)
}

// Dispatch fires an event; handlers may mutate ev or set Cancel=true.
// Returns true if event was cancelled by any handler.
func Dispatch(
	session *models.GameSession,
	ev *Event,
) bool {
	for _, h := range registry[ev.Type] {
		h(session, ev)
		if ev.Cancel {
			return true
		}
	}
	return false
}
