package ical

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"nomad-residence-be/internal/domain/port"
	"time"

	"github.com/arran4/golang-ical"
)

type httpFetcher struct {
	client *http.Client
}

func NewHTTPFetcher() port.IcalFetcher {
	return &httpFetcher{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (f *httpFetcher) FetchFromURL(ctx context.Context, urlStr string) (*ics.Calendar, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, err
	}
	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("fetch ical: status=%d body=%s", resp.StatusCode, string(body))
	}
	return ics.ParseCalendar(resp.Body)
}
