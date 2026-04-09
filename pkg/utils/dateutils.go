package utils

import "time"

const DateLayout = "2006-01-02"

var dateLocation = time.FixedZone("Asia/Ho_Chi_Minh", 7*60*60)

func FormatDate(t time.Time) string {
	return t.In(dateLocation).Format(DateLayout)
}

func ParseDate(s string) (time.Time, error) {
	return time.ParseInLocation(DateLayout, s, dateLocation)
}

func GetDateRange(from, to time.Time) []time.Time {
	var dates []time.Time
	d := TruncateToDate(from)
	end := TruncateToDate(to)
	for d.Before(end) {
		dates = append(dates, d)
		d = d.AddDate(0, 0, 1)
	}
	return dates
}

func TruncateToDate(t time.Time) time.Time {
	local := t.In(dateLocation)
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, dateLocation)
}

func AddMinutes(t time.Time, minutes int) time.Time {
	return t.Add(time.Minute * time.Duration(minutes))
}

func NightsBetween(checkin, checkout time.Time) int {
	d := TruncateToDate(checkout).Sub(TruncateToDate(checkin))
	return int(d.Hours() / 24)
}
