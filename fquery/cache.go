package fquery

import (
	"io"
	"time"
)

type Cache interface {
	Source
	io.Closer

	SetQuoteExpiry(dur time.Duration)

	HasQuote(symbol string) bool
	HasHist(symbol string, start *time.Time, end *time.Time) bool
}
