package api

import (
	"log/slog"
	"nomad-residence-be/internal/api/handler/admin"
	"nomad-residence-be/internal/api/handler/public"
	"nomad-residence-be/internal/api/middlewares"
	"nomad-residence-be/internal/domain/port"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RouterConfig struct {
	DB *gorm.DB

	BookingUsecase port.BookingUsecase
	RoomUsecase    port.RoomUsecase
	PricingUsecase port.PricingUsecase
	PaymentUsecase port.PaymentUsecase
	IcalUsecase    port.IcalUsecase

	AdminRepo   port.AdminRepository
	BookingRepo port.BookingRepository
	SettingRepo port.SettingRepository
	BlockedRepo port.BlockedDateRepository

	VietQRService port.VietQRService

	JWTSecret   string
	JWTTokenTTL time.Duration
	IcalSecret  string

	Logger *slog.Logger
}

func NewRouter(cfg RouterConfig) *gin.Engine {
	engine := gin.New()

	engine.Use(middlewares.Recovery(cfg.Logger))
	engine.Use(middlewares.RequestID())
	engine.Use(middlewares.Logger(cfg.Logger))
	engine.Use(middlewares.CORS(middlewares.DefaultCORSConfig()))
	engine.Use(middlewares.Timeout(30 * time.Second))
	engine.Use(middlewares.RateLimit(middlewares.DefaultRateLimitConfig()))
	engine.Use(middlewares.ErrorHandler())

	roomHandler := public.NewRoomHandler(cfg.RoomUsecase, cfg.PricingUsecase, cfg.Logger)
	bookingHandler := public.NewBookingHandler(cfg.BookingUsecase, cfg.Logger)
	paymentHandler := public.NewPaymentHandler(cfg.PaymentUsecase, cfg.BookingUsecase, cfg.VietQRService, cfg.Logger)
	icalHandler := public.NewIcalHandler(cfg.IcalUsecase, cfg.IcalSecret, cfg.Logger)

	authHandler := admin.NewAuthHandler(cfg.AdminRepo, cfg.JWTSecret, cfg.JWTTokenTTL, cfg.Logger)
	adminRoomHandler := admin.NewRoomHandler(cfg.RoomUsecase, cfg.Logger)
	adminBookingHandler := admin.NewBookingHandler(cfg.BookingUsecase, cfg.Logger)
	adminPricingHandler := admin.NewPricingHandler(cfg.PricingUsecase, cfg.Logger)
	adminIcalHandler := admin.NewIcalHandler(cfg.IcalUsecase, icalHandler, cfg.Logger)
	adminBlockedHandler := admin.NewBlockedDateHandler(cfg.RoomUsecase, cfg.BlockedRepo, cfg.BookingRepo, cfg.Logger)
	adminDashboardHandler := admin.NewDashboardHandler(cfg.DB, cfg.SettingRepo, cfg.Logger)

	v1 := engine.Group("/api/v1")

	// Rooms (public)
	rooms := v1.Group("/rooms")
	{
		rooms.GET("", roomHandler.ListRooms)
		rooms.GET("/:slug", roomHandler.GetRoomDetail)
		rooms.GET("/:slug/calendar", roomHandler.GetRoomCalendar)
	}

	// Bookings (guest)
	bookings := v1.Group("/bookings")
	{
		bookings.POST("", bookingHandler.CreateBooking)
		bookings.GET("/lookup", bookingHandler.LookupBooking)
	}

	// Payments (guest)
	payments := v1.Group("/payments")
	{
		payments.GET("/vietqr", paymentHandler.GetQRPayment)
		payments.GET("/status", paymentHandler.CheckPaymentStatus)
		payments.POST("/vietqr/webhook",
			middlewares.RateLimit(middlewares.StrictRateLimitConfig()),
			paymentHandler.VietQRWebhook,
		)
	}

	v1.GET("/ical/:room_id/calendar.ics", icalHandler.ExportIcal)

	adminGroup := v1.Group("/admin")

	auth := adminGroup.Group("/auth")
	{
		auth.POST("/login",
			middlewares.RateLimit(middlewares.StrictRateLimitConfig()),
			authHandler.Login,
		)
		auth.GET("/me",
			middlewares.Authenticate(cfg.AdminRepo, cfg.JWTSecret),
			authHandler.GetMe,
		)
		auth.POST("/change-password",
			middlewares.Authenticate(cfg.AdminRepo, cfg.JWTSecret),
			authHandler.ChangePassword,
		)
	}

	protected := adminGroup.Group("")
	protected.Use(middlewares.Authenticate(cfg.AdminRepo, cfg.JWTSecret))
	{
		// Dashboard
		protected.GET("/dashboard", adminDashboardHandler.GetDashboard)
		protected.GET("/settings", adminDashboardHandler.GetSettings)
		protected.PUT("/settings", adminDashboardHandler.UpdateSettings)

		// Bookings
		bk := protected.Group("/bookings")
		{
			bk.GET("", adminBookingHandler.ListBookings)
			bk.POST("/manual", adminBookingHandler.CreateManualBooking)
			bk.GET("/:id", adminBookingHandler.GetBooking)
			bk.POST("/:id/confirm", adminBookingHandler.ConfirmBooking)
			bk.POST("/:id/cancel", adminBookingHandler.CancelBooking)
		}

		// Rooms + sub-resources
		rm := protected.Group("/rooms")
		{
			rm.GET("", adminRoomHandler.ListRooms)
			rm.POST("", adminRoomHandler.CreateRoom)
			rm.GET("/:id", adminRoomHandler.GetRoom)
			rm.PUT("/:id", adminRoomHandler.UpdateRoom)
			rm.DELETE("/:id", adminRoomHandler.DeleteRoom)

			// Images
			rm.POST("/:id/images", adminRoomHandler.AddImage)
			rm.PATCH("/:id/images/reorder", adminRoomHandler.ReorderImages)

			// Amenities
			rm.PUT("/:id/amenities", adminRoomHandler.UpdateAmenities)

			// Pricing rules
			rm.GET("/:id/pricing", adminPricingHandler.ListRules)
			rm.POST("/:id/pricing", adminPricingHandler.CreateRule)
			rm.PATCH("/:id/pricing/:ruleId", adminPricingHandler.UpdateRule)
			rm.DELETE("/:id/pricing/:ruleId", adminPricingHandler.DeleteRule)

			// iCal links per room
			rm.GET("/:id/ical", adminIcalHandler.ListIcalLinks)
			rm.POST("/:id/ical", adminIcalHandler.AddIcalLink)
			rm.GET("/:id/ical/export-url", adminIcalHandler.GetExportURL)

			// Blocked dates per room
			rm.GET("/:id/blocked-dates", adminBlockedHandler.ListBlockedDates)
			rm.POST("/:id/blocked-dates", adminBlockedHandler.BlockDates)
		}

		protected.DELETE("/images/:imageId", adminRoomHandler.DeleteImage)

		protected.DELETE("/ical-links/:linkId", adminIcalHandler.DeleteIcalLink)
		protected.POST("/ical-links/:linkId/sync", adminIcalHandler.SyncIcalLink)

		protected.DELETE("/blocked-dates/:id", adminBlockedHandler.UnblockDate)

		super := protected.Group("")
		super.Use(middlewares.Authorize("superadmin"))
		{
			_ = super
		}
	}

	return engine
}
