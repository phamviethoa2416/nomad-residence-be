package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"nomad-residence-be/config"
	"nomad-residence-be/internal/api"
	"nomad-residence-be/internal/infrastructure/database"
	"nomad-residence-be/internal/infrastructure/ical"
	"nomad-residence-be/internal/infrastructure/notification"
	"nomad-residence-be/internal/infrastructure/vietqr"
	"nomad-residence-be/internal/jobs"
	"nomad-residence-be/internal/repository"
	"nomad-residence-be/internal/usecase"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	cfgPath := os.Getenv("CONFIG_PATH")
	if cfgPath == "" {
		cfgPath = "."
	}

	appConfig, err := config.LoadConfig(cfgPath)
	if err != nil {
		logger.Error("failed to load config", slog.Any("err", err))
		os.Exit(1)
	}

	if appConfig.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	db, err := database.NewPostgres(appConfig.Database, logger)
	if err != nil {
		logger.Error("failed to connect database", slog.Any("error", err))
		os.Exit(1)
	}

	defer func() {
		_ = database.Close(db)
	}()

	adminRepository := repository.NewAdminRepository(db)
	roomRepository := repository.NewRoomRepository(db)
	bookingRepository := repository.NewBookingRepository(db)
	blockedRepository := repository.NewBlockedDateRepository(db)
	icalLinkRepository := repository.NewIcalLinkRepository(db)
	pricingRepository := repository.NewPricingRuleRepository(db)
	paymentRepository := repository.NewPaymentRepository(db)
	settingRepository := repository.NewSettingRepository(db)
	tx := repository.NewTransactionManager(db)

	notificationService := notification.NewService(appConfig.Notification)
	VietQRService := vietqr.NewService(appConfig.VietQR)
	icalService := ical.NewHTTPFetcher()

	icalSecret := os.Getenv("ICAL_SECRET")
	if icalSecret == "" {
		icalSecret = appConfig.JWT.Secret
	}

	appURL := os.Getenv("APP_URL")
	if appURL == "" {
		appURL = fmt.Sprintf("http://%s:%d", appConfig.Server.Host, appConfig.Server.Port)
	}

	pricingUsecase := usecase.NewPricingUsecase(pricingRepository, roomRepository, bookingRepository, logger)
	roomUsecase := usecase.NewRoomUsecase(roomRepository, blockedRepository, bookingRepository, pricingUsecase, logger)
	bookingUsecase := usecase.NewBookingUsecase(db, tx, bookingRepository, roomRepository, blockedRepository, paymentRepository, pricingUsecase, notificationService, logger)
	paymentUsecase := usecase.NewPaymentUsecase(db, tx, paymentRepository, bookingRepository, blockedRepository, bookingUsecase, logger)
	icalUsecase := usecase.NewIcalUsecase(icalLinkRepository, blockedRepository, bookingRepository, roomRepository, notificationService, icalService, icalSecret, appURL, logger)

	// Scheduler
	scheduler := jobs.NewScheduler(
		appConfig.Scheduler,
		jobs.NewBookingJob(bookingRepository, logger),
		jobs.NewIcalJob(icalUsecase, logger),
		logger,
	)
	if err := scheduler.Start(); err != nil {
		logger.Error("failed to start scheduler", slog.Any("error", err))
		os.Exit(1)
	}
	defer scheduler.Stop()

	engine := api.NewRouter(api.RouterConfig{
		DB: db,

		BookingUsecase: bookingUsecase,
		RoomUsecase:    roomUsecase,
		PricingUsecase: pricingUsecase,
		PaymentUsecase: paymentUsecase,
		IcalUsecase:    icalUsecase,

		AdminRepo:   adminRepository,
		BookingRepo: bookingRepository,
		SettingRepo: settingRepository,
		BlockedRepo: blockedRepository,

		VietQRService: VietQRService,

		JWTSecret:   appConfig.JWT.Secret,
		JWTTokenTTL: appConfig.JWT.TokenTTL,
		IcalSecret:  icalSecret,

		Logger: logger,
	})

	addr := fmt.Sprintf("%s:%d", appConfig.Server.Host, appConfig.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      engine,
		ReadTimeout:  appConfig.Server.ReadTimeout,
		WriteTimeout: appConfig.Server.WriteTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("http server starting", slog.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		logger.Info("shutdown signal received", slog.String("signal", sig.String()))
	case err := <-errCh:
		logger.Error("http server failed", slog.Any("error", err))
	}

	timeout := appConfig.Server.ShutdownTimeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	_ = srv.Shutdown(ctx)
}
