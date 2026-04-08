package port

import (
	"context"
	"nomad-residence-be/internal/domain/entity"
)

type NotificationService interface {
	SendBookingConfirmationEmail(ctx context.Context, booking *entity.Booking) error
	SendBookingCancellationEmail(ctx context.Context, booking *entity.Booking) error
	NotifyAdminBookingConfirmed(ctx context.Context, booking *entity.Booking) error
}
