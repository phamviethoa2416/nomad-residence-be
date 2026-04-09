package port

import (
	"context"

	"github.com/arran4/golang-ical"
)

type IcalFetcher interface {
	FetchFromURL(ctx context.Context, url string) (*ics.Calendar, error)
}
