package admin

import (
	"log/slog"
	"nomad-residence-be/internal/api/middlewares"
	"nomad-residence-be/internal/api/request"
	"nomad-residence-be/internal/domain/port"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	adminRepo   port.AdminRepository
	jwtSecret   string
	jwtTokenTTL time.Duration
	logger      *slog.Logger
}

func NewAuthHandler(
	adminRepo port.AdminRepository,
	jwtSecret string,
	jwtTokenTTL time.Duration,
	logger *slog.Logger,
) *AuthHandler {
	return &AuthHandler{
		adminRepo:   adminRepo,
		jwtSecret:   jwtSecret,
		jwtTokenTTL: jwtTokenTTL,
		logger:      logger,
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req request.AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	admin, err := h.adminRepo.FindByEmail(c.Request.Context(), req.Email)
	if err != nil {
		h.logger.Error("login: db error", slog.Any("error", err))
		c.JSON(500, gin.H{"error": "Lỗi server, vui lòng thử lại sau"})
		return
	}

	if admin == nil {
		c.JSON(401, gin.H{"error": "Email hoặc mật khẩu không đúng", "code": "INVALID_CREDENTIALS"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(401, gin.H{"error": "Email hoặc mật khẩu không đúng", "code": "INVALID_CREDENTIALS"})
		return
	}

	token, err := middlewares.GenerateToken(admin, h.jwtSecret, h.jwtTokenTTL)
	if err != nil {
		h.logger.Error("login: token generation failed", slog.Any("error", err))
		c.JSON(500, gin.H{"error": "Không thể tạo token"})
		return
	}

	h.logger.Info("audit_admin_login",
		slog.String("action", "login"),
		slog.Uint64("admin_id", uint64(admin.ID)),
		slog.String("email", admin.Email),
		slog.String("request_id", middlewares.GetRequestID(c)),
	)

	c.JSON(200, gin.H{
		"token": token,
		"admin": gin.H{
			"id":        admin.ID,
			"email":     admin.Email,
			"full_name": admin.FullName,
			"role":      admin.Role,
		},
	})
}

func (h *AuthHandler) GetMe(c *gin.Context) {
	adminCtx := middlewares.GetAdminContext(c)
	if adminCtx == nil {
		c.JSON(401, gin.H{"error": "Vui lòng đăng nhập", "code": "UNAUTHORIZED"})
		return
	}

	admin, err := h.adminRepo.FindByID(c.Request.Context(), adminCtx.ID)
	if err != nil {
		h.logger.Error("get_me: db error", slog.Any("error", err))
		c.JSON(500, gin.H{"error": "Lỗi server, vui lòng thử lại sau"})
		return
	}
	if admin == nil {
		c.JSON(404, gin.H{"error": "Không tìm thấy tài khoản", "code": "NOT_FOUND"})
		return
	}

	c.JSON(200, gin.H{
		"id":         admin.ID,
		"email":      admin.Email,
		"full_name":  admin.FullName,
		"phone":      admin.Phone,
		"role":       admin.Role,
		"created_at": admin.CreatedAt,
	})
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req request.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	ctx := middlewares.GetAdminContext(c)
	if ctx == nil {
		c.JSON(401, gin.H{"error": "Vui lòng đăng nhập", "code": "UNAUTHORIZED"})
		return
	}

	admin, err := h.adminRepo.FindByID(c.Request.Context(), ctx.ID)
	if err != nil {
		h.logger.Error("change_password: db error", slog.Any("error", err))
		c.JSON(500, gin.H{"error": "Lỗi server, vui lòng thử lại sau"})
		return
	}
	if admin == nil {
		c.JSON(404, gin.H{"error": "Không tìm thấy tài khoản", "code": "NOT_FOUND"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(req.OldPassword)); err != nil {
		c.JSON(400, gin.H{"error": "Mật khẩu hiện tại không đúng", "code": "INVALID_PASSWORD"})
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), 12)
	if err != nil {
		h.logger.Error("change_password: bcrypt failed", slog.Any("error", err))
		c.JSON(500, gin.H{"error": "Lỗi server, vui lòng thử lại sau"})
		return
	}

	now := time.Now()
	admin.PasswordHash = string(hashed)
	admin.UpdatedAt = now
	admin.PasswordChangedAt = &now
	if err := h.adminRepo.Update(c.Request.Context(), admin); err != nil {
		h.logger.Error("change_password: update failed", slog.Any("error", err))
		c.JSON(500, gin.H{"error": "Lỗi server, vui lòng thử lại sau"})
		return
	}

	h.logger.Info("audit_password_change",
		slog.String("action", "password_change"),
		slog.Uint64("admin_id", uint64(ctx.ID)),
		slog.String("request_id", middlewares.GetRequestID(c)),
	)

	c.JSON(200, gin.H{"message": "Đổi mật khẩu thành công"})
}
