package validator

import (
	"reflect"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

var (
	bookingCodeRegex = regexp.MustCompile(`^BK-[A-Z0-9]{6,}\d+$`)
	timeHHMMRegex    = regexp.MustCompile(`^([01]\d|2[0-3]):([0-5]\d)$`)
	slugRegex        = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
)

func Register() {
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return
	}

	mustRegister(v, "booking_code", validateBookingCode)
	mustRegister(v, "time_hhmm", validateTimeHHMM)
	mustRegister(v, "slug", validateSlug)
	mustRegister(v, "password_strength", validatePasswordStrength)
	mustRegister(v, "date_gte_today", validateDateGteToday)

	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" || name == "" {
			name = strings.SplitN(fld.Tag.Get("form"), ",", 2)[0]
		}
		if name == "-" || name == "" {
			return ""
		}
		return name
	})
}

func mustRegister(v *validator.Validate, tag string, fn validator.Func) {
	if err := v.RegisterValidation(tag, fn); err != nil {
		panic("validator: failed to register " + tag + ": " + err.Error())
	}
}

func validateBookingCode(fl validator.FieldLevel) bool {
	return bookingCodeRegex.MatchString(fl.Field().String())
}

func validateTimeHHMM(fl validator.FieldLevel) bool {
	s := fl.Field().String()
	if s == "" {
		return true
	}
	return timeHHMMRegex.MatchString(s)
}

func validateSlug(fl validator.FieldLevel) bool {
	return slugRegex.MatchString(fl.Field().String())
}

func validatePasswordStrength(fl validator.FieldLevel) bool {
	s := fl.Field().String()
	var hasUpper, hasLower, hasDigit bool
	for _, r := range s {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}
	return hasUpper && hasLower && hasDigit
}

func validateDateGteToday(fl validator.FieldLevel) bool {
	s := fl.Field().String()
	if s == "" {
		return true
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return false
	}
	today := time.Now().UTC().Truncate(24 * time.Hour)
	return !t.Before(today)
}

func NormalizePhone(phone string) string {
	phone = strings.TrimSpace(phone)
	phone = strings.ReplaceAll(phone, " ", "")
	if strings.HasPrefix(phone, "+84") {
		phone = "0" + phone[3:]
	}
	return phone
}

func ValidateDateRange(checkinStr, checkoutStr string) (field, msg string) {
	if checkinStr == "" || checkoutStr == "" {
		return "", ""
	}
	checkin, err1 := time.Parse("2006-01-02", checkinStr)
	checkout, err2 := time.Parse("2006-01-02", checkoutStr)
	if err1 != nil || err2 != nil {
		return "", ""
	}
	if !checkin.Before(checkout) {
		return "checkout_date", "Ngày checkout phải sau ngày checkin"
	}
	return "", ""
}

func ValidateDateRangeFromTo(fromStr, toStr string) (field, msg string) {
	if fromStr == "" || toStr == "" {
		return "", ""
	}
	from, err1 := time.Parse("2006-01-02", fromStr)
	to, err2 := time.Parse("2006-01-02", toStr)
	if err1 != nil || err2 != nil {
		return "", ""
	}
	if from.After(to) {
		return "to", "Ngày bắt đầu phải trước ngày đến"
	}
	return "", ""
}
