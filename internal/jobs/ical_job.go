package jobs

import (
	"context"
	"log/slog"
	"nomad-residence-be/internal/usecase"
	"time"
)

type IcalJob struct {
	icalUsecase *usecase.IcalUsecase
	logger      *slog.Logger
}

func NewIcalJob(icalUsecase *usecase.IcalUsecase, logger *slog.Logger) *IcalJob {
	return &IcalJob{icalUsecase: icalUsecase, logger: logger}
}

func (j *IcalJob) Run() {
	batchID := shortID()
	j.logger.Debug("Starting Ical Job", slog.String("batch_id", batchID))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	results := j.icalUsecase.SyncAllIcalLinks(ctx)

	success, failed := 0, 0
	for _, res := range results {
		if res.Success {
			success++
		} else {
			failed++
		}
	}

	j.logger.Info("Finished Ical Job",
		slog.String("batch_id", batchID),
		slog.Int("success", success),
		slog.Int("failed", failed),
	)
}
