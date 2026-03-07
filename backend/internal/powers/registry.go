package powers

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
)

type TargetType string

const (
	TargetSelf        TargetType = "self"
	TargetOtherPlayer TargetType = "other_player"
	TargetMonster     TargetType = "monster"
	TargetSession     TargetType = "session"
)

type PowerDef struct {
	Name       string
	CostGems   int
	TargetType TargetType
	IsPassive  bool
}

type ExecutionContext struct {
	GameID       uuid.UUID
	ActorID      uuid.UUID
	TargetPlayer *uuid.UUID
	MonsterSlot  *int
	Payload      json.RawMessage
}

type Executor func(ctx context.Context, ec ExecutionContext, runner PowerRunner) error

type Registry struct {
	defs  map[string]PowerDef
	exec  map[string]Executor
}

func NewRegistry() *Registry {
	return &Registry{
		defs: make(map[string]PowerDef),
		exec: make(map[string]Executor),
	}
}

func (r *Registry) Register(name string, def PowerDef, exec Executor) {
	r.defs[name] = def
	r.exec[name] = exec
}

func (r *Registry) Get(name string) (PowerDef, bool) {
	d, ok := r.defs[name]
	return d, ok
}

func (r *Registry) Execute(ctx context.Context, name string, ec ExecutionContext, runner PowerRunner) error {
	f, ok := r.exec[name]
	if !ok {
		return nil
	}
	return f(ctx, ec, runner)
}
