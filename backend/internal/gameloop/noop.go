package gameloop

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
)

type NoopProcessor struct{}

func (NoopProcessor) Process(ctx context.Context, gameID uuid.UUID, cmd Command) ProcessResult {
	slog.Debug("game loop noop", "game_id", gameID, "cmd", cmd.Type)
	return ProcessResult{}
}
