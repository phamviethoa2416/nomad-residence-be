package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"nomad-residence-be/config"
	"nomad-residence-be/internal/domain/entity"
	"time"

	"gopkg.in/gomail.v2"
)

type Service struct {
	cfg    config.NotificationConfig
	dialer *gomail.Dialer
}

func NewService(cfg config.NotificationConfig) *Service {
	var dialer *gomail.Dialer
	if cfg.Email.User != "" && cfg.Email.Password != "" {
		dialer = gomail.NewDialer(
			cfg.Email.Host,
			cfg.Email.Port,
			cfg.Email.User,
			cfg.Email.Password,
		)
		dialer.SSL = cfg.Email.Port == 465
	}
	return &Service{cfg: cfg, dialer: dialer}
}

func (s *Service) SendBookingConfirmationEmail(ctx context.Context, booking *entity.Booking) error {
	if booking.GuestEmail == nil {
		return nil
	}
	if s.dialer == nil {
		return nil // SMTP not configured
	}

	checkin := booking.CheckinDate.Format("02/01/2006")
	checkout := booking.CheckoutDate.Format("02/01/2006")
	amount := formatVND(booking.TotalAmount.IntPart())
	roomName := ""
	if booking.Room.ID != 0 {
		roomName = booking.Room.Name
	}

	html := fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
		  <h2 style="color: #2563eb;">Đặt phòng thành công!</h2>
		  <p>Xin chào <strong>%s</strong>,</p>
		  <p>Đơn đặt phòng của bạn đã được xác nhận. Chi tiết:</p>
		  <table style="width:100%%; border-collapse: collapse; margin: 20px 0;">
			<tr style="background: #f3f4f6;">
			  <td style="padding: 10px; border: 1px solid #e5e7eb;"><strong>Mã đặt phòng</strong></td>
			  <td style="padding: 10px; border: 1px solid #e5e7eb;">%s</td>
			</tr>
			<tr>
			  <td style="padding: 10px; border: 1px solid #e5e7eb;"><strong>Phòng</strong></td>
			  <td style="padding: 10px; border: 1px solid #e5e7eb;">%s</td>
			</tr>
			<tr style="background: #f3f4f6;">
			  <td style="padding: 10px; border: 1px solid #e5e7eb;"><strong>Nhận phòng</strong></td>
			  <td style="padding: 10px; border: 1px solid #e5e7eb;">%s</td>
			</tr>
			<tr>
			  <td style="padding: 10px; border: 1px solid #e5e7eb;"><strong>Trả phòng</strong></td>
			  <td style="padding: 10px; border: 1px solid #e5e7eb;">%s</td>
			</tr>
			<tr style="background: #f3f4f6;">
			  <td style="padding: 10px; border: 1px solid #e5e7eb;"><strong>Số đêm</strong></td>
			  <td style="padding: 10px; border: 1px solid #e5e7eb;">%d đêm</td>
			</tr>
			<tr>
			  <td style="padding: 10px; border: 1px solid #e5e7eb;"><strong>Tổng tiền</strong></td>
			  <td style="padding: 10px; border: 1px solid #e5e7eb; color: #dc2626;"><strong>%sđ</strong></td>
			</tr>
		  </table>
		  <p>Cảm ơn bạn đã tin tưởng và lựa chọn Nomad Residence!</p>
		</div>`,
		template.HTMLEscapeString(booking.GuestName),
		template.HTMLEscapeString(booking.BookingCode),
		template.HTMLEscapeString(roomName),
		checkin,
		checkout,
		booking.NumNights,
		amount,
	)

	m := gomail.NewMessage()
	m.SetHeader("From", s.cfg.Email.From)
	m.SetHeader("To", *booking.GuestEmail)
	m.SetHeader("Subject", fmt.Sprintf("[%s] Xác nhận đặt phòng thành công", booking.BookingCode))
	m.SetBody("text/html", html)

	return s.dialer.DialAndSend(m)
}

func (s *Service) SendBookingCancellationEmail(ctx context.Context, booking *entity.Booking) error {
	if booking.GuestEmail == nil {
		return nil
	}

	if s.dialer == nil {
		return nil
	}

	cancelReason := ""
	if booking.CancelReason != nil {
		cancelReason = fmt.Sprintf("<p>Lý do: %s</p>", template.HTMLEscapeString(*booking.CancelReason))
	}

	html := fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
		  <h2 style="color: #dc2626;">Đặt phòng đã bị hủy</h2>
		  <p>Xin chào <strong>%s</strong>,</p>
		  <p>Đặt phòng <strong>%s</strong> đã bị hủy.</p>
		  %s
		  <p>Nếu có thắc mắc, vui lòng liên hệ chúng tôi.</p>
		</div>`,
		template.HTMLEscapeString(booking.GuestName),
		template.HTMLEscapeString(booking.BookingCode),
		cancelReason,
	)

	m := gomail.NewMessage()
	m.SetHeader("From", s.cfg.Email.From)
	m.SetHeader("To", *booking.GuestEmail)
	m.SetHeader("Subject", fmt.Sprintf("[%s] Đặt phòng đã bị hủy", booking.BookingCode))
	m.SetBody("text/html", html)

	return s.dialer.DialAndSend(m)
}

func (s *Service) SendTelegram(ctx context.Context, message string) error {
	cfg := s.cfg.Telegram
	if cfg.BotToken == "" || cfg.ChatID == "" {
		return nil // not configured
	}

	payload := map[string]interface{}{
		"chat_id":    cfg.ChatID,
		"text":       message,
		"parse_mode": "HTML",
	}
	body, _ := json.Marshal(payload)

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", cfg.BotToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("Failed to close Telegram response body: %v\n", err)
		}
	}(resp.Body)
	return nil
}

func (s *Service) NotifyAdminBookingConfirmed(ctx context.Context, booking *entity.Booking) error {
	checkin := booking.CheckinDate.Format("02/01/2006")
	checkout := booking.CheckoutDate.Format("02/01/2006")
	amount := formatVND(booking.TotalAmount.IntPart())
	roomName := fmt.Sprintf("%d", booking.RoomID)
	if booking.Room.ID != 0 {
		roomName = booking.Room.Name
	}

	guestNote := ""
	if booking.GuestNote != nil {
		guestNote = fmt.Sprintf("\n📝 Ghi chú: %s", *booking.GuestNote)
	}

	msg := fmt.Sprintf(
		"🏠 <b>ĐẶT PHÒNG MỚI!</b>\n\n"+
			"📋 Mã: <code>%s</code>\n"+
			"🏡 Phòng: %s\n"+
			"👤 Khách: %s\n"+
			"📞 SĐT: %s\n"+
			"📅 Nhận phòng: %s\n"+
			"📅 Trả phòng: %s\n"+
			"🌙 Số đêm: %d\n"+
			"💰 Tổng tiền: %sđ\n"+
			"📌 Nguồn: %s%s",
		booking.BookingCode,
		roomName,
		booking.GuestName,
		booking.GuestPhone,
		checkin,
		checkout,
		booking.NumNights,
		amount,
		booking.Source,
		guestNote,
	)

	return s.SendTelegram(ctx, msg)
}

func formatVND(amount int64) string {
	s := fmt.Sprintf("%d", amount)
	result := make([]byte, 0, len(s)+len(s)/3)
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result = append(result, '.')
		}
		result = append(result, byte(c))
	}
	return string(result)
}
