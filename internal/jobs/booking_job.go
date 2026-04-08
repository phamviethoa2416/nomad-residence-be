package jobs

import (
	"context"
	"fmt"
	"log/slog"
	"nomad-residence-be/internal/domain/port"
	"time"
)

type BookingJob struct {
	bookingRepo port.BookingRepository
	logger      *slog.Logger
}

func NewBookingJob(bookingRepo port.BookingRepository, logger *slog.Logger) *BookingJob {
	return &BookingJob{bookingRepo: bookingRepo, logger: logger}
}

func (j *BookingJob) Run() {
	batchID := shortID()
	j.logger.Debug("Starting Booking Job", slog.String("batch_id", batchID))

	j.cancelExpiredBookings(batchID)
	j.markCompletedBookings(batchID)

	j.logger.Debug("Finished Booking Job", slog.String("batch_id", batchID))
}

func (j *BookingJob) cancelExpiredBookings(batchID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	count, err := j.bookingRepo.CancelExpiredPending(ctx)
	if err != nil {
		j.logger.Error("Failed to cancel expired bookings",
			slog.String("batchID", batchID),
			slog.Any("error", err),
		)
		return
	}

	if count > 0 {
		j.logger.Info(" Đã hủy đơn hết hạn",
			slog.String("batch_id", batchID),
			slog.Int64("count", count),
		)
	}
}

func (j *BookingJob) markCompletedBookings(batchID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	count, err := j.bookingRepo.MarkCompletedPastCheckout(ctx)
	if err != nil {
		j.logger.Error("Failed to mark completed bookings",
			slog.String("batch_id", batchID),
			slog.Any("error", err),
		)
		return
	}

	if count > 0 {
		j.logger.Info("[CronJob] Đã chuyển đơn sang hoàn tất",
			slog.String("batch_id", batchID),
			slog.Int64("count", count),
		)
	}
}

func shortID() string {
	return fmt.Sprintf("%08x", time.Now().UnixNano()&0xFFFFFFFF)
}
