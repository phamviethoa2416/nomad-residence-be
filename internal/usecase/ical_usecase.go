package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"nomad-residence-be/internal/domain/entity"
	"nomad-residence-be/internal/domain/port"
	"nomad-residence-be/internal/infrastructure/notification"
	"nomad-residence-be/pkg/errors"
	"nomad-residence-be/pkg/utils"
	"time"

	"github.com/arran4/golang-ical"
)

type IcalUsecase struct {
	icalLinkRepo port.IcalLinkRepository
	blockedRepo  port.BlockedDateRepository
	bookingRepo  port.BookingRepository
	roomRepo     port.RoomRepository
	notif        *notification.Service
	fetcher      port.IcalFetcher
	tokenSecret  string
	appURL       string
	logger       *slog.Logger
}

func NewIcalUsecase(
	icalLinkRepo port.IcalLinkRepository,
	blockedRepo port.BlockedDateRepository,
	bookingRepo port.BookingRepository,
	roomRepo port.RoomRepository,
	notif *notification.Service,
	fetcher port.IcalFetcher,
	tokenSecret string,
	appURL string,
	logger *slog.Logger,
) *IcalUsecase {
	return &IcalUsecase{
		icalLinkRepo: icalLinkRepo,
		blockedRepo:  blockedRepo,
		bookingRepo:  bookingRepo,
		roomRepo:     roomRepo,
		notif:        notif,
		fetcher:      fetcher,
		tokenSecret:  tokenSecret,
		appURL:       appURL,
		logger:       logger,
	}
}

func (u *IcalUsecase) ExportRoomIcal(ctx context.Context, roomID uint) (string, error) {
	room, err := u.roomRepo.FindByID(ctx, roomID)
	if err != nil {
		return "", err
	}
	if room == nil {
		return "", errors.ErrRoomNotFound
	}

	filter := struct {
		RoomID uint
		Status []string
	}{RoomID: roomID}
	_ = filter

	cal := ics.NewCalendar()
	cal.SetMethod(ics.MethodPublish)
	cal.SetXWRCalName(room.Name)

	blocked, err := u.blockedRepo.FindByRoomAndRange(ctx, roomID, time.Now().AddDate(-1, 0, 0), time.Now().AddDate(2, 0, 0))
	if err != nil {
		return "", err
	}

	for _, b := range blocked {
		event := cal.AddEvent(fmt.Sprintf("%d-%s", roomID, utils.FormatDate(b.Date)))
		event.SetSummary("Blocked")
		event.SetDtStampTime(b.Date)
		event.SetStartAt(b.Date)
		event.SetEndAt(b.Date.AddDate(0, 0, 1))
		if b.Reason != nil {
			event.SetDescription(*b.Reason)
		}
	}

	return cal.Serialize(), nil
}

func (u *IcalUsecase) SyncSingleLinkByID(ctx context.Context, linkID uint) (*entity.IcalSyncResult, error) {
	link, err := u.icalLinkRepo.FindByID(ctx, linkID)
	if err != nil {
		return nil, err
	}
	if link == nil {
		return nil, errors.ErrIcalLinkNotFound
	}
	if link.ImportURL == nil {
		return &entity.IcalSyncResult{LinkID: linkID, RoomID: link.RoomID, Platform: link.Platform, Success: false, Error: "no import URL configured"}, nil
	}

	link.SyncStatus = entity.IcalSyncing
	_ = u.icalLinkRepo.Update(ctx, link)

	cal, err := u.fetcher.FetchFromURL(ctx, *link.ImportURL)
	if err != nil {
		syncErr := err.Error()
		link.SyncStatus = entity.IcalError
		link.SyncError = &syncErr
		_ = u.icalLinkRepo.Update(ctx, link)

		room, _ := u.roomRepo.FindByID(ctx, link.RoomID)
		roomName := ""
		if room != nil {
			roomName = room.Name
		}
		if u.notif != nil {
			_ = u.notif.SendTelegram(ctx, fmt.Sprintf(
				"❌ <b>LỖI ĐỒNG BỘ iCal</b>\n\nPlatform: %s\nPhòng: %s\nLỗi: %s",
				link.Platform, roomName, syncErr,
			))
		}
		return nil, err
	}

	source := "ota_" + link.Platform

	// Tính toán dates từ VEVENT
	newDates := u.extractDatesFromCalendar(cal, link.RoomID, link.Platform)

	// Xóa dates cũ của platform này
	_ = u.blockedRepo.DeleteByRoomAndSource(ctx, link.RoomID, source)

	if len(newDates) > 0 {
		if err := u.blockedRepo.BulkCreate(ctx, newDates); err != nil {
			return nil, err
		}
	}

	if len(newDates) == 0 {
		now := time.Now()
		link.SyncStatus = entity.IcalIdle
		link.LastSyncedAt = &now
		link.SyncError = nil
		_ = u.icalLinkRepo.Update(ctx, link)

		return &entity.IcalSyncResult{
			LinkID:   linkID,
			RoomID:   link.RoomID,
			Platform: link.Platform,
			Success:  true,
			Imported: 0,
		}, nil
	}

	minDate := newDates[0].Date
	maxDate := newDates[0].Date
	for _, d := range newDates {
		if d.Date.Before(minDate) {
			minDate = d.Date
		}
		if d.Date.After(maxDate) {
			maxDate = d.Date
		}
	}

	dateSet := make(map[string]bool, len(newDates))
	for _, d := range newDates {
		dateSet[utils.FormatDate(d.Date)] = true
	}

	bookings, err := u.bookingRepo.FindConfirmedOverlapping(ctx, link.RoomID, minDate, maxDate)
	if err != nil {
		return nil, err
	}
	conflictDates := make([]time.Time, 0)
	seen := make(map[string]bool, 16)
	for _, bk := range bookings {
		dates := utils.GetDateRange(bk.CheckinDate, bk.CheckoutDate)
		for _, d := range dates {
			key := utils.FormatDate(d)
			if !dateSet[key] || seen[key] {
				continue
			}
			seen[key] = true
			conflictDates = append(conflictDates, d)
		}
	}

	if len(conflictDates) > 0 {
		dateStrs := make([]string, len(conflictDates))
		for i, d := range conflictDates {
			dateStrs[i] = utils.FormatDate(d)
		}
		room, _ := u.roomRepo.FindByID(ctx, link.RoomID)
		if u.notif != nil && room != nil {
			_ = u.notif.SendTelegram(ctx, fmt.Sprintf(
				"⚠️ <b>XUNG ĐỘT LỊCH iCal!</b>\n\n🏡 Phòng: %s\n📅 Ngày xung đột: %v\n\nVui lòng kiểm tra và xử lý thủ công!",
				room.Name, dateStrs,
			))
		}
	}

	now := time.Now()
	link.SyncStatus = entity.IcalIdle
	link.LastSyncedAt = &now
	link.SyncError = nil
	_ = u.icalLinkRepo.Update(ctx, link)

	return &entity.IcalSyncResult{
		LinkID:   linkID,
		RoomID:   link.RoomID,
		Platform: link.Platform,
		Success:  true,
		Imported: len(newDates),
	}, nil
}

func (u *IcalUsecase) SyncAllIcalLinks(ctx context.Context) []entity.IcalSyncResult {
	links, err := u.icalLinkRepo.FindActiveImportLinks(ctx)
	if err != nil {
		return nil
	}

	results := make([]entity.IcalSyncResult, 0, len(links))
	for _, link := range links {
		result, err := u.SyncSingleLinkByID(ctx, link.ID)
		if err != nil {
			results = append(results, entity.IcalSyncResult{LinkID: link.ID, RoomID: link.RoomID, Platform: link.Platform, Success: false, Error: err.Error()})
		} else {
			results = append(results, *result)
		}
	}
	return results
}

func (u *IcalUsecase) GetLinksByRoomID(ctx context.Context, roomID uint) ([]entity.IcalLink, error) {
	return u.icalLinkRepo.FindByRoomID(ctx, roomID)
}

func (u *IcalUsecase) GetLinkByID(ctx context.Context, id uint) (*entity.IcalLink, error) {
	return u.icalLinkRepo.FindByID(ctx, id)
}

func (u *IcalUsecase) CreateLink(ctx context.Context, link *entity.IcalLink) error {
	return u.icalLinkRepo.Create(ctx, link)
}

func (u *IcalUsecase) UpdateLink(ctx context.Context, link *entity.IcalLink) error {
	return u.icalLinkRepo.Update(ctx, link)
}

func (u *IcalUsecase) DeleteLink(ctx context.Context, id uint) error {
	return u.icalLinkRepo.Delete(ctx, id)
}

func (u *IcalUsecase) extractDatesFromCalendar(cal *ics.Calendar, roomID uint, platform string) []entity.BlockedDate {
	source := "ota_" + platform
	var result []entity.BlockedDate

	for _, component := range cal.Components {
		event, ok := component.(*ics.VEvent)
		if !ok {
			continue
		}
		start, err1 := event.GetStartAt()
		end, err2 := event.GetEndAt()
		if err1 != nil || err2 != nil {
			continue
		}

		uid := event.GetProperty(ics.ComponentPropertyUniqueId)
		var sourceRef *string
		if uid != nil {
			s := uid.Value
			sourceRef = &s
		}

		summary := ""
		if sp := event.GetProperty(ics.ComponentPropertySummary); sp != nil {
			summary = sp.Value
		}

		reason := fmt.Sprintf("Blocked via %s", platform)
		if summary != "" {
			reason = summary
		}
		reasonCopy := reason

		dates := utils.GetDateRange(start, end)
		for _, d := range dates {
			dd := d
			result = append(result, entity.BlockedDate{
				RoomID:    roomID,
				Date:      dd,
				Source:    entity.BlockSource(source),
				SourceRef: sourceRef,
				Reason:    &reasonCopy,
			})
		}
	}
	return result
}
