package middlewares

import (
	"errors"
	"nomad-residence-be/internal/domain/entity"
	"nomad-residence-be/internal/domain/port"
	apperrors "nomad-residence-be/pkg/errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type AdminClaims struct {
	ID           uint   `json:"id"`
	Email        string `json:"email"`
	Role         string `json:"role"`
	PwdChangedAt int64  `json:"pwdChangedAt,omitempty"`
	jwt.RegisteredClaims
}

type AdminContext struct {
	ID       uint
	Email    string
	FullName string
	Role     string
}

func GenerateToken(admin *entity.Admin, secret string, ttl time.Duration) (string, error) {
	now := time.Now()
	var pwdChanged int64
	if admin.PasswordChangedAt != nil {
		pwdChanged = admin.PasswordChangedAt.Unix()
	}
	claims := AdminClaims{
		ID:           admin.ID,
		Email:        admin.Email,
		Role:         string(admin.Role),
		PwdChangedAt: pwdChanged,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func Authenticate(adminRepo port.AdminRepository, jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			AbortWithError(c, apperrors.New(
				401,
				"UNAUTHORIZED",
				"Vui lòng đăng nhập"),
			)
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenStr == "" {
			AbortWithError(c, apperrors.New(
				401,
				"INVALID_TOKEN",
				"Token không tồn tại"),
			)
			return
		}

		claims := &AdminClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(jwtSecret), nil
		})

		if err != nil {
			if errors.Is(err, jwt.ErrTokenExpired) {
				AbortWithError(c, apperrors.New(
					401, "TOKEN_EXPIRED",
					"Phiên đăng nhập đã hết hạn, vui lòng đăng nhập lại"),
				)
				return
			}
			AbortWithError(c, apperrors.New(
				401,
				"INVALID_TOKEN",
				"Token không hợp lệ, vui lòng đăng nhập lại"),
			)
			return
		}

		if !token.Valid || claims.ID == 0 {
			AbortWithError(c, apperrors.New(
				401,
				"INVALID_TOKEN",
				"Token không hợp lệ, vui lòng đăng nhập lại"),
			)
			return
		}

		ctx := c.Request.Context()
		admin, err := adminRepo.FindByID(ctx, claims.ID)
		if err != nil {
			AbortWithError(c, apperrors.New(
				500,
				"INTERNAL_SERVER_ERROR",
				"Lỗi server"),
			)
			return
		}
		if admin == nil {
			AbortWithError(c, apperrors.New(
				401,
				"INVALID_TOKEN",
				"Tài khoản không tồn tại"),
			)
			return
		}

		if claims.PwdChangedAt > 0 && admin.PasswordChangedAt != nil {
			if admin.PasswordChangedAt.Unix() > claims.PwdChangedAt {
				AbortWithError(c, apperrors.New(
					401,
					"TOKEN_REVOKED",
					"Phiên đăng nhập đã bị thu hồi, vui lòng đăng nhập lại"),
				)
				return
			}
		}

		c.Set("admin_claims", &AdminContext{
			ID:       admin.ID,
			Email:    admin.Email,
			FullName: admin.FullName,
			Role:     string(admin.Role),
		})

		c.Next()
	}
}

func GetAdminContext(c *gin.Context) *AdminContext {
	val, exists := c.Get("admin_claims")
	if !exists {
		return nil
	}
	admin, _ := val.(*AdminContext)
	return admin
}
