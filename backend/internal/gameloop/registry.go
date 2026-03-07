package gameloop

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
)

type CommandType string

const (
	CmdPick     CommandType = "pick"
	CmdDiscard  CommandType = "discard"
	CmdDeclare  CommandType = "declare"
	CmdPower    CommandType = "power"
	CmdSkipPower CommandType = "skip_power"
	CmdAttack   CommandType = "attack"
	CmdDemask   CommandType = "demask"
	CmdAccuse   CommandType = "accuse"
	CmdChallengeWindowClosed CommandType = "challenge_window_closed"
)

type Command struct {
	Type     CommandType
	PlayerID uuid.UUID
	Payload  json.RawMessage
}

type DelayedCommand struct {
	After   time.Duration
	Command Command
}

type ProcessResult struct {
	Err      error
	Schedule []DelayedCommand
}

type Processor interface {
	Process(ctx context.Context, gameID uuid.UUID, cmd Command) ProcessResult
}

type Registry struct {
	mu     sync.RWMutex
	loops  map[uuid.UUID]chan<- Command
	done   map[uuid.UUID]chan struct{}
	proc   Processor
}

func NewRegistry(proc Processor) *Registry {
	return &Registry{
		loops: make(map[uuid.UUID]chan<- Command),
		done:  make(map[uuid.UUID]chan struct{}),
		proc:  proc,
	}
}

func (r *Registry) Start(ctx context.Context, gameID uuid.UUID) {
	r.mu.Lock()
	if _, ok := r.loops[gameID]; ok {
		r.mu.Unlock()
		return
	}
	ch := make(chan Command, 64)
	done := make(chan struct{})
	r.loops[gameID] = ch
	r.done[gameID] = done
	r.mu.Unlock()

	go func() {
		defer func() {
			r.mu.Lock()
			delete(r.loops, gameID)
			close(done)
			delete(r.done, gameID)
			r.mu.Unlock()
		}()
		for {
			select {
			case <-ctx.Done():
				return
			case cmd, ok := <-ch:
				if !ok {
					return
				}
				res := r.proc.Process(ctx, gameID, cmd)
				if res.Err != nil {
					slog.Error("game loop process", "game_id", gameID, "cmd", cmd.Type, "err", res.Err)
				}
				for _, d := range res.Schedule {
					dc := d
					go func() {
						time.Sleep(dc.After)
						r.Send(ctx, gameID, dc.Command)
					}()
				}
			}
		}
	}()
}

func (r *Registry) Send(ctx context.Context, gameID uuid.UUID, cmd Command) bool {
	r.mu.RLock()
	ch, ok := r.loops[gameID]
	r.mu.RUnlock()
	if !ok {
		return false
	}
	select {
	case ch <- cmd:
		return true
	case <-ctx.Done():
		return false
	}
}

func (r *Registry) Stop(gameID uuid.UUID) {
	r.mu.Lock()
	ch, ok := r.loops[gameID]
	if !ok {
		r.mu.Unlock()
		return
	}
	delete(r.loops, gameID)
	done := r.done[gameID]
	r.mu.Unlock()
	close(ch)
	<-done
}
