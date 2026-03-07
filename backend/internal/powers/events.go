package powers

import (
	"context"
	"sync"

	"github.com/google/uuid"
)

type EventKind int

const (
	EventKindGemsGained EventKind = iota + 1
	EventKindStealAttempt
	EventKindTurnStart
)

type Event struct {
	Kind EventKind

	// Runner is set by the emitter so handlers can perform game state updates (e.g. teamwork doubling gems).
	Runner PowerRunner

	// GemsGained
	GemsGameID   uuid.UUID
	GemsPlayerID uuid.UUID
	GemsAmount   int
	GemsIsBonus  bool

	// StealAttempt
	StealGameID   uuid.UUID
	StealerID     uuid.UUID
	TargetID      uuid.UUID
	StealResource string
	Cancelled     bool

	// TurnStart
	TurnStartGameID   uuid.UUID
	TurnStartPlayerID uuid.UUID
}

type EventHandler func(ctx context.Context, ev *Event)

type eventBus struct {
	mu       sync.RWMutex
	handlers map[EventKind][]EventHandler
}

func newEventBus() *eventBus {
	return &eventBus{handlers: make(map[EventKind][]EventHandler)}
}

func (b *eventBus) Register(kind EventKind, h EventHandler) {
	b.mu.Lock()
	b.handlers[kind] = append(b.handlers[kind], h)
	b.mu.Unlock()
}

func (b *eventBus) Emit(ctx context.Context, ev *Event) {
	b.mu.RLock()
	list := b.handlers[ev.Kind]
	if len(list) > 0 {
		list = append([]EventHandler(nil), list...)
	}
	b.mu.RUnlock()
	for _, h := range list {
		h(ctx, ev)
	}
}

var defaultBus = newEventBus()

func RegisterEventHandler(kind EventKind, h EventHandler) {
	defaultBus.Register(kind, h)
}

func EmitEvent(ctx context.Context, ev *Event) {
	defaultBus.Emit(ctx, ev)
}
