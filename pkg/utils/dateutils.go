package utils

import "time"

const DateLayout = "2006-01-02"

func FormatDate(t time.Time) string {
	return t.Format(DateLayout)
}

func ParseDate(s string) (time.Time, error) {
	return time.ParseInLocation(DateLayout, s, time.UTC)
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
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func AddMinutes(t time.Time, minutes int) time.Time {
	return t.Add(time.Minute * time.Duration(minutes))
}

func NightsBetween(checkin, checkout time.Time) int {
	d := TruncateToDate(checkout).Sub(TruncateToDate(checkin))
	return int(d.Hours() / 24)
}
