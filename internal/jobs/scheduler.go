package jobs

import (
	"log/slog"
	"nomad-residence-be/config"

	"github.com/robfig/cron/v3"
)

func DefaultSchedulerConfig() config.SchedulerConfig {
	return config.SchedulerConfig{
		BookingJobCron:          "* * * * *",
		IcalSyncIntervalMinutes: 30,
	}
}

type Scheduler struct {
	c      *cron.Cron
	cfg    config.SchedulerConfig
	logger *slog.Logger

	bookingJob *BookingJob
	icalJob    *IcalJob
}

func NewScheduler(
	cfg config.SchedulerConfig,
	bookingJob *BookingJob,
	icalJob *IcalJob,
	logger *slog.Logger,
) *Scheduler {
	c := cron.New(
		cron.WithLogger(cron.DefaultLogger),
		cron.WithChain(
			cron.SkipIfStillRunning(cron.DefaultLogger),
			cron.Recover(cron.DefaultLogger),
		),
	)

	return &Scheduler{
		c:          c,
		cfg:        cfg,
		logger:     logger,
		bookingJob: bookingJob,
		icalJob:    icalJob,
	}
}

func (s *Scheduler) Start() error {
	// Booking lifecycle job
	bookingCron := s.cfg.BookingJobCron
	if bookingCron == "" {
		bookingCron = "* * * * *"
	}

	if _, err := s.c.AddFunc(bookingCron, s.bookingJob.Run); err != nil {
		return err
	}

	s.logger.Info("[CronJob] Cancel Expired Bookings and Mark Completed Bookings scheduled",
		slog.String("cron_expr", bookingCron),
	)

	// iCal sync job
	icalCron := intervalToCron(s.cfg.IcalSyncIntervalMinutes)

	if _, err := s.c.AddFunc(icalCron, s.icalJob.Run); err != nil {
		return err
	}
	s.logger.Info("[CronJob] iCal sync scheduled",
		slog.Int("interval_minutes", s.cfg.IcalSyncIntervalMinutes),
		slog.String("cron_expr", icalCron),
	)

	s.c.Start()
	return nil
}

func (s *Scheduler) Stop() {
	s.logger.Info("Scheduler stopping...")
	ctx := s.c.Stop()
	<-ctx.Done()
	s.logger.Info("Scheduler stopped")
}

func intervalToCron(minutes int) string {
	switch {
	case minutes <= 0:
		return "*/30 * * * *"
	case minutes <= 15:
		return "*/15 * * * *"
	case minutes <= 30:
		return "*/30 * * * *"
	default:
		return "0 * * * *"
	}
}
