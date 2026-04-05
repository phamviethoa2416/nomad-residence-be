package usecase

import (
	"context"
	"log/slog"
	"nomad-residence-be/internal/domain/entity"
	"nomad-residence-be/internal/domain/filter"
	"nomad-residence-be/internal/repository"
	apperrors "nomad-residence-be/pkg/errors"
	"nomad-residence-be/pkg/utils"
	"time"
)

type RoomUsecase struct {
	roomRepo        repository.RoomRepository
	blockedDateRepo repository.BlockedDateRepository
	bookingRepo     repository.BookingRepository
	pricingUsecase  *PricingUsecase
	logger          *slog.Logger
}

func NewRoomUsecase(
	roomRepo repository.RoomRepository,
	blockedDateRepo repository.BlockedDateRepository,
	bookingRepo repository.BookingRepository,
	pricingUsecase *PricingUsecase,
	logger *slog.Logger,
) *RoomUsecase {
	return &RoomUsecase{
		roomRepo:        roomRepo,
		blockedDateRepo: blockedDateRepo,
		bookingRepo:     bookingRepo,
		pricingUsecase:  pricingUsecase,
		logger:          logger,
	}
}

func (uc *RoomUsecase) ListRooms(ctx context.Context, filter filter.RoomFilter) ([]entity.Room, int64, error) {
	return uc.roomRepo.FindAll(ctx, filter)
}

func (uc *RoomUsecase) GetRoomByID(ctx context.Context, id uint) (*entity.Room, error) {
	room, err := uc.roomRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return nil, apperrors.ErrRoomNotFound
	}
	return room, nil
}

func (uc *RoomUsecase) GetRoomBySlug(ctx context.Context, slug string) (*entity.Room, error) {
	room, err := uc.roomRepo.FindBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if room == nil {
		return nil, apperrors.ErrRoomNotFound
	}
	return room, nil
}

func (uc *RoomUsecase) CreateRoom(ctx context.Context, room *entity.Room) error {
	return uc.roomRepo.Create(ctx, room)
}

func (uc *RoomUsecase) UpdateRoom(ctx context.Context, room *entity.Room) error {
	return uc.roomRepo.Update(ctx, room)
}

func (uc *RoomUsecase) DeleteRoom(ctx context.Context, id uint) error {
	existing, err := uc.roomRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return apperrors.ErrRoomNotFound
	}
	return uc.roomRepo.Delete(ctx, id)
}

func (uc *RoomUsecase) AddImages(ctx context.Context, roomID uint, images []entity.RoomImage) error {
	room, err := uc.roomRepo.FindByID(ctx, roomID)
	if err != nil {
		return err
	}
	if room == nil {
		return apperrors.ErrRoomNotFound
	}

	for i := range images {
		images[i].RoomID = roomID
	}
	return uc.roomRepo.AddImages(ctx, images)
}

func (uc *RoomUsecase) DeleteImage(ctx context.Context, imageID uint) error {
	return uc.roomRepo.DeleteImage(ctx, imageID)
}

func (uc *RoomUsecase) SetPrimaryImage(ctx context.Context, roomID, imageID uint) error {
	if err := uc.roomRepo.ResetPrimaryImages(ctx, roomID); err != nil {
		return err
	}
	return uc.roomRepo.UpdateImageSortOrder(ctx, imageID, 0)
}

func (uc *RoomUsecase) ReorderImages(ctx context.Context, roomID uint, orders []struct {
	ID        uint
	SortOrder int
}) error {
	for _, o := range orders {
		if err := uc.roomRepo.UpdateImageSortOrder(ctx, o.ID, o.SortOrder); err != nil {
			return err
		}
	}
	return nil
}

func (uc *RoomUsecase) ReplaceAmenities(ctx context.Context, roomID uint, amenities []entity.RoomAmenity) error {
	room, err := uc.roomRepo.FindByID(ctx, roomID)
	if err != nil {
		return err
	}
	if room == nil {
		return apperrors.ErrRoomNotFound
	}
	return uc.roomRepo.ReplaceAmenities(ctx, roomID, amenities)
}

func (uc *RoomUsecase) CheckAvailability(
	ctx context.Context,
	roomID uint,
	checkin, checkout time.Time,
) (bool, error) {
	return uc.bookingRepo.IsAvailable(ctx, roomID, checkin, checkout, nil)
}

type CalendarDay struct {
	Date      string `json:"date"`
	Available bool   `json:"available"`
	Source    string `json:"source,omitempty"`
}

func (uc *RoomUsecase) GetRoomCalendar(
	ctx context.Context,
	roomID uint,
	from, to time.Time,
) ([]CalendarDay, error) {
	from = utils.TruncateToDate(from)
	to = utils.TruncateToDate(to)

	blockedDates, err := uc.blockedDateRepo.FindByRoomAndRange(ctx, roomID, from, to)
	if err != nil {
		return nil, err
	}

	blockedMap := make(map[string]string, len(blockedDates))
	for _, bd := range blockedDates {
		key := utils.FormatDate(bd.Date)
		blockedMap[key] = string(bd.Source)
	}

	dates := utils.GetDateRange(from, to)
	calendar := make([]CalendarDay, 0, len(dates))
	for _, date := range dates {
		key := utils.FormatDate(date)
		source, blocked := blockedMap[key]
		calendar = append(calendar, CalendarDay{
			Date:      key,
			Available: !blocked,
			Source:    source,
		})
	}

	return calendar, nil
}

func (uc *RoomUsecase) GetRoomDetailWithPrice(
	ctx context.Context,
	roomID uint,
	checkin, checkout *time.Time,
) (*entity.Room, *PriceBreakdown, error) {
	room, err := uc.roomRepo.FindByID(ctx, roomID)
	if err != nil {
		return nil, nil, err
	}
	if room == nil {
		return nil, nil, apperrors.ErrRoomNotFound
	}

	if checkin == nil || checkout == nil {
		return room, nil, nil
	}

	breakdown, err := uc.pricingUsecase.CalculatePriceForRoom(ctx, room, *checkin, *checkout)
	if err != nil {
		uc.logger.Warn("price calculation failed, returning room without pricing",
			slog.Uint64("room_id", uint64(roomID)),
			slog.Any("error", err),
		)
		return room, nil, nil
	}

	return room, breakdown, nil
}

// ListAvailableRooms returns rooms that are available for the given date range,
// along with their pricing, filtered by the provided criteria.
func (uc *RoomUsecase) ListAvailableRooms(
	ctx context.Context,
	f filter.RoomFilter,
	checkin, checkout *time.Time,
) ([]entity.Room, int64, error) {
	rooms, total, err := uc.roomRepo.FindAll(ctx, f)
	if err != nil {
		return nil, 0, err
	}

	if checkin == nil || checkout == nil {
		return rooms, total, nil
	}

	unavailableIDs, err := uc.blockedDateRepo.GetUnavailableRoomIDs(ctx, *checkin, *checkout)
	if err != nil {
		uc.logger.Error("failed to get unavailable room IDs", slog.Any("error", err))
		return rooms, total, nil
	}

	blockedSet := make(map[uint]bool, len(unavailableIDs))
	for _, id := range unavailableIDs {
		blockedSet[id] = true
	}

	available := make([]entity.Room, 0, len(rooms))
	for _, room := range rooms {
		if !blockedSet[room.ID] {
			available = append(available, room)
		}
	}

	return available, int64(len(available)), nil
}

// --- Blocked Dates Management ---

func (uc *RoomUsecase) GetBlockedDates(
	ctx context.Context,
	roomID uint,
	from, to time.Time,
) ([]entity.BlockedDate, error) {
	return uc.blockedDateRepo.FindByRoomAndRange(ctx, roomID, from, to)
}

func (uc *RoomUsecase) BlockDates(ctx context.Context, roomID uint, dates []time.Time, reason *string) error {
	room, err := uc.roomRepo.FindByID(ctx, roomID)
	if err != nil {
		return err
	}
	if room == nil {
		return apperrors.ErrRoomNotFound
	}

	blocked := make([]entity.BlockedDate, 0, len(dates))
	for _, d := range dates {
		blocked = append(blocked, entity.BlockedDate{
			RoomID: roomID,
			Date:   utils.TruncateToDate(d),
			Source: entity.BlockManual,
			Reason: reason,
		})
	}

	return uc.blockedDateRepo.BulkCreate(ctx, blocked)
}

func (uc *RoomUsecase) UnblockDate(ctx context.Context, roomID uint, date time.Time) error {
	return uc.blockedDateRepo.DeleteByRoomAndDate(ctx, roomID, utils.TruncateToDate(date))
}

func (uc *RoomUsecase) UnblockDateRange(ctx context.Context, roomID uint, from, to time.Time) error {
	return uc.blockedDateRepo.DeleteByRoomAndRange(ctx, roomID, from, to)
}
