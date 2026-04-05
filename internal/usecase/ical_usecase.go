package usecase

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"nomad-residence-be/internal/domain/entity"
	"nomad-residence-be/internal/repository"
	apperrors "nomad-residence-be/pkg/errors"
	"nomad-residence-be/pkg/utils"
	"strings"
	"time"

	ics "github.com/arran4/golang-ical"
)

type IcalSyncResult struct {
	LinkID   uint   `json:"link_id"`
	RoomID   uint   `json:"room_id"`
	Platform string `json:"platform"`
	Success  bool   `json:"success"`
	Error    string `json:"error,omitempty"`
	Imported int    `json:"imported"`
}

type IcalUsecase struct {
	icalRepo        repository.IcalLinkRepository
	blockedDateRepo repository.BlockedDateRepository
	bookingRepo     repository.BookingRepository
	roomRepo        repository.RoomRepository
	logger          *slog.Logger
	httpClient      *http.Client
}

func NewIcalUsecase(
	icalRepo repository.IcalLinkRepository,
	blockedDateRepo repository.BlockedDateRepository,
	bookingRepo repository.BookingRepository,
	roomRepo repository.RoomRepository,
	logger *slog.Logger,
) *IcalUsecase {
	return &IcalUsecase{
		icalRepo:        icalRepo,
		blockedDateRepo: blockedDateRepo,
		bookingRepo:     bookingRepo,
		roomRepo:        roomRepo,
		logger:          logger,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// --- CRUD operations ---

func (uc *IcalUsecase) GetLinksByRoomID(ctx context.Context, roomID uint) ([]entity.IcalLink, error) {
	return uc.icalRepo.FindByRoomID(ctx, roomID)
}

func (uc *IcalUsecase) GetLinkByID(ctx context.Context, id uint) (*entity.IcalLink, error) {
	link, err := uc.icalRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if link == nil {
		return nil, apperrors.ErrIcalLinkNotFound
	}
	return link, nil
}

func (uc *IcalUsecase) CreateLink(ctx context.Context, link *entity.IcalLink) error {
	return uc.icalRepo.Create(ctx, link)
}

func (uc *IcalUsecase) UpdateLink(ctx context.Context, link *entity.IcalLink) error {
	return uc.icalRepo.Update(ctx, link)
}

func (uc *IcalUsecase) DeleteLink(ctx context.Context, id uint) error {
	link, err := uc.icalRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if link == nil {
		return apperrors.ErrIcalLinkNotFound
	}

	if link.ImportURL != nil {
		srcKey := fmt.Sprintf("%s:ical_%d", entity.BlockIcal, link.ID)
		if err := uc.blockedDateRepo.DeleteByRoomAndSource(ctx, link.RoomID, srcKey); err != nil {
			uc.logger.Warn("failed to clean up blocked dates for deleted ical link",
				slog.Uint64("link_id", uint64(id)),
				slog.Any("error", err),
			)
		}
	}

	return uc.icalRepo.Delete(ctx, id)
}

// --- Sync operations ---

// SyncAllIcalLinks fetches and processes all active iCal import links.
// This is called by the cron job scheduler.
func (uc *IcalUsecase) SyncAllIcalLinks(ctx context.Context) []IcalSyncResult {
	links, err := uc.icalRepo.FindActiveImportLinks(ctx)
	if err != nil {
		uc.logger.Error("failed to fetch active ical links", slog.Any("error", err))
		return nil
	}

	results := make([]IcalSyncResult, 0, len(links))
	for _, link := range links {
		result := uc.syncSingleLink(ctx, &link)
		results = append(results, result)
	}

	return results
}

// SyncSingleLinkByID triggers a sync for a specific iCal link.
func (uc *IcalUsecase) SyncSingleLinkByID(ctx context.Context, linkID uint) (*IcalSyncResult, error) {
	link, err := uc.icalRepo.FindByID(ctx, linkID)
	if err != nil {
		return nil, err
	}
	if link == nil {
		return nil, apperrors.ErrIcalLinkNotFound
	}
	if link.ImportURL == nil || *link.ImportURL == "" {
		return nil, apperrors.New(400, "NO_IMPORT_URL", "Link không có URL import")
	}

	result := uc.syncSingleLink(ctx, link)
	return &result, nil
}

func (uc *IcalUsecase) syncSingleLink(ctx context.Context, link *entity.IcalLink) IcalSyncResult {
	result := IcalSyncResult{
		LinkID:   link.ID,
		RoomID:   link.RoomID,
		Platform: link.Platform,
	}

	link.SyncStatus = entity.IcalSyncing
	_ = uc.icalRepo.Update(ctx, link)

	events, err := uc.fetchAndParseIcal(ctx, *link.ImportURL)
	if err != nil {
		result.Error = err.Error()
		uc.markSyncError(ctx, link, err.Error())
		return result
	}

	sourceKey := fmt.Sprintf("%s:ical_%d", entity.BlockIcal, link.ID)
	if err := uc.blockedDateRepo.DeleteByRoomAndSource(ctx, link.RoomID, sourceKey); err != nil {
		uc.logger.Error("failed to clear old ical blocked dates",
			slog.Uint64("link_id", uint64(link.ID)),
			slog.Any("error", err),
		)
	}

	blockedDates := uc.eventsToBlockedDates(link.RoomID, link.ID, events)
	if len(blockedDates) > 0 {
		if err := uc.blockedDateRepo.BulkCreate(ctx, blockedDates); err != nil {
			result.Error = err.Error()
			uc.markSyncError(ctx, link, err.Error())
			return result
		}
	}

	result.Success = true
	result.Imported = len(blockedDates)

	now := time.Now()
	link.SyncStatus = entity.IcalIdle
	link.LastSyncedAt = &now
	link.SyncError = nil
	_ = uc.icalRepo.Update(ctx, link)

	uc.logger.Info("iCal sync completed",
		slog.Uint64("link_id", uint64(link.ID)),
		slog.String("platform", link.Platform),
		slog.Int("imported", len(blockedDates)),
	)

	return result
}

func (uc *IcalUsecase) fetchAndParseIcal(ctx context.Context, url string) ([]*ics.VEvent, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "NomadResidence/1.0 iCal-Sync")

	resp, err := uc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch ical: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ical fetch returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("failed to read ical body: %w", err)
	}

	cal, err := ics.ParseCalendar(strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ical: %w", err)
	}

	return cal.Events(), nil
}

func (uc *IcalUsecase) eventsToBlockedDates(roomID uint, linkID uint, events []*ics.VEvent) []entity.BlockedDate {
	var blocked []entity.BlockedDate
	sourceRef := fmt.Sprintf("ical_%d", linkID)
	now := utils.TruncateToDate(time.Now())

	for _, event := range events {
		dtStart, err := event.GetStartAt()
		if err != nil {
			continue
		}
		dtEnd, err := event.GetEndAt()
		if err != nil {
			dtEnd = dtStart.AddDate(0, 0, 1)
		}

		if dtEnd.Before(now) {
			continue
		}

		summary := ""
		if prop := event.GetProperty(ics.ComponentPropertySummary); prop != nil {
			summary = prop.Value
		}

		dates := utils.GetDateRange(dtStart, dtEnd)
		for _, d := range dates {
			if d.Before(now) {
				continue
			}
			reason := summary
			blocked = append(blocked, entity.BlockedDate{
				RoomID:    roomID,
				Date:      d,
				Source:    entity.BlockIcal,
				SourceRef: &sourceRef,
				Reason:    &reason,
			})
		}
	}

	return blocked
}

func (uc *IcalUsecase) markSyncError(ctx context.Context, link *entity.IcalLink, errMsg string) {
	link.SyncStatus = entity.IcalError
	link.SyncError = &errMsg
	if err := uc.icalRepo.Update(ctx, link); err != nil {
		uc.logger.Error("failed to update ical link sync error",
			slog.Uint64("link_id", uint64(link.ID)),
			slog.Any("error", err),
		)
	}
}

// --- Export ---

// ExportRoomIcal generates an iCal feed for a room's confirmed bookings and blocked dates.
func (uc *IcalUsecase) ExportRoomIcal(ctx context.Context, roomID uint) (string, error) {
	room, err := uc.roomRepo.FindByID(ctx, roomID)
	if err != nil {
		return "", err
	}
	if room == nil {
		return "", apperrors.ErrRoomNotFound
	}

	now := time.Now()
	from := utils.TruncateToDate(now)
	to := from.AddDate(1, 0, 0)

	blockedDates, err := uc.blockedDateRepo.FindByRoomAndRange(ctx, roomID, from, to)
	if err != nil {
		return "", err
	}

	cal := ics.NewCalendar()
	cal.SetMethod(ics.MethodPublish)
	cal.SetProductId(fmt.Sprintf("-//Nomad Residence//%s//EN", room.Name))
	cal.SetName(fmt.Sprintf("Nomad Residence - %s", room.Name))

	dateRanges := uc.mergeDateRanges(blockedDates)
	for i, dr := range dateRanges {
		event := cal.AddEvent(fmt.Sprintf("blocked-%d-%d@nomadresidence", roomID, i))
		event.SetCreatedTime(now)
		event.SetDtStampTime(now)
		event.SetAllDayStartAt(dr.Start)
		event.SetAllDayEndAt(dr.End.AddDate(0, 0, 1))
		event.SetSummary(fmt.Sprintf("Blocked - %s", room.Name))
		event.SetStatus("CONFIRMED")
	}

	return cal.Serialize(), nil
}

type dateRange struct {
	Start  time.Time
	End    time.Time
	Source entity.BlockSource
}

// mergeDateRanges groups consecutive blocked dates into ranges for cleaner iCal export.
func (uc *IcalUsecase) mergeDateRanges(dates []entity.BlockedDate) []dateRange {
	if len(dates) == 0 {
		return nil
	}

	ranges := make([]dateRange, 0)
	current := dateRange{
		Start:  dates[0].Date,
		End:    dates[0].Date,
		Source: dates[0].Source,
	}

	for i := 1; i < len(dates); i++ {
		d := dates[i]
		nextDay := current.End.AddDate(0, 0, 1)

		if d.Date.Equal(nextDay) && d.Source == current.Source {
			current.End = d.Date
		} else {
			ranges = append(ranges, current)
			current = dateRange{
				Start:  d.Date,
				End:    d.Date,
				Source: d.Source,
			}
		}
	}
	ranges = append(ranges, current)

	return ranges
}
