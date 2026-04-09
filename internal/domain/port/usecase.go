package port

import (
	"context"
	"nomad-residence-be/internal/domain/entity"
	"nomad-residence-be/internal/domain/filter"
	"time"

	"github.com/shopspring/decimal"
)

type BookingUsecase interface {
	CreateBooking(ctx context.Context, roomID uint, checkinStr, checkoutStr string, numGuests int, guestName, guestPhone string, guestEmail, guestNote *string) (*entity.Booking, error)
	LookupBooking(ctx context.Context, code, phone string) (*entity.Booking, error)
	ListBookings(ctx context.Context, f filter.BookingFilter) ([]entity.Booking, int64, error)
	GetBookingByID(ctx context.Context, id uint) (*entity.Booking, error)
	ConfirmBooking(ctx context.Context, id uint, adminNote string) (*entity.Booking, error)
	CancelBooking(ctx context.Context, id uint, reason string) (*entity.Booking, error)
	UpdateBookingStatus(ctx context.Context, id uint, status string, adminNote, cancelReason *string) (*entity.Booking, error)
	CreateManualBooking(ctx context.Context, roomID uint, checkinStr, checkoutStr string, guestName, guestPhone string, guestEmail, guestNote *string, numGuests int, source, adminNote string) (*entity.Booking, error)
}

type RoomUsecase interface {
	ListRooms(ctx context.Context, filter filter.RoomFilter) ([]entity.Room, int64, error)
	GetRoomByID(ctx context.Context, id uint) (*entity.Room, error)
	GetRoomBySlug(ctx context.Context, slug string) (*entity.Room, error)
	CreateRoom(ctx context.Context, room *entity.Room) error
	UpdateRoom(ctx context.Context, room *entity.Room) error
	DeleteRoom(ctx context.Context, id uint) error
	AddImages(ctx context.Context, roomID uint, images []entity.RoomImage) error
	DeleteImage(ctx context.Context, imageID uint) error
	SetPrimaryImage(ctx context.Context, roomID, imageID uint) error
	ReorderImages(ctx context.Context, roomID uint, orders []entity.ImageOrder) error
	ReplaceAmenities(ctx context.Context, roomID uint, amenities []entity.Amenity) error
	CheckAvailability(ctx context.Context, roomID uint, checkin, checkout time.Time) (bool, error)
	GetRoomCalendar(ctx context.Context, roomID uint, from, to time.Time) ([]entity.CalendarDay, error)
	GetRoomDetailWithPrice(ctx context.Context, roomID uint, checkin, checkout *time.Time) (*entity.Room, *entity.PriceBreakdown, error)
	ListAvailableRooms(ctx context.Context, f filter.RoomFilter, checkin, checkout *time.Time) ([]entity.Room, int64, error)
	GetBlockedDates(ctx context.Context, roomID uint, from, to time.Time) ([]entity.BlockedDate, error)
	BlockDates(ctx context.Context, roomID uint, dates []time.Time, reason *string) error
	UnblockDate(ctx context.Context, roomID uint, date time.Time) error
	UnblockDateRange(ctx context.Context, roomID uint, from, to time.Time) error
}

type PricingUsecase interface {
	CalculatePrice(ctx context.Context, roomID uint, checkin, checkout time.Time) (*entity.PriceBreakdown, error)
	CalculatePriceForRoom(ctx context.Context, room *entity.Room, checkin, checkout time.Time) (*entity.PriceBreakdown, error)
	GetRulesByRoomID(ctx context.Context, roomID uint) ([]entity.PricingRule, error)
	GetRuleByID(ctx context.Context, id uint) (*entity.PricingRule, error)
	CreateRule(ctx context.Context, rule *entity.PricingRule) error
	UpdateRule(ctx context.Context, rule *entity.PricingRule) error
	DeleteRule(ctx context.Context, id uint) error
}

type PaymentUsecase interface {
	CreatePayment(ctx context.Context, bookingID uint, amount decimal.Decimal, method string) (*entity.Payment, error)
	GetPaymentByID(ctx context.Context, id uint) (*entity.Payment, error)
	GetPaymentsByBookingID(ctx context.Context, bookingID uint) ([]entity.Payment, error)
	GetPaymentsByBookingCode(ctx context.Context, bookingCode string) ([]entity.Payment, error)
	HandleVietQRWebhook(ctx context.Context, transactionID string, amount decimal.Decimal, content string) error
	HandleQRCallback(ctx context.Context, transactionID string, amount decimal.Decimal, content string) error
	UpdatePaymentStatus(ctx context.Context, paymentID uint, status string, adminNote *string) (*entity.Payment, error)
	ConfirmAllPendingPayments(ctx context.Context, bookingID uint, adminNote string) error
}

type IcalUsecase interface {
	GetLinksByRoomID(ctx context.Context, roomID uint) ([]entity.IcalLink, error)
	GetLinkByID(ctx context.Context, id uint) (*entity.IcalLink, error)
	CreateLink(ctx context.Context, link *entity.IcalLink) error
	UpdateLink(ctx context.Context, link *entity.IcalLink) error
	DeleteLink(ctx context.Context, id uint) error
	SyncAllIcalLinks(ctx context.Context) []entity.IcalSyncResult
	SyncSingleLinkByID(ctx context.Context, linkID uint) (*entity.IcalSyncResult, error)
	ExportRoomIcal(ctx context.Context, roomID uint) (string, error)
}
