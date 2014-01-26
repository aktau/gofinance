package fquery

import (
	"fmt"
	"github.com/aktau/gofinance/util"
	"time"
)

const (
	ErrTplNotSupported = "source '%s' does not support action '%s'"
)

type Quote struct {
	Symbol   string /* e.g.: VEUR.AS, Vanguard dev. europe on Amsterdam */
	Name     string
	Exchange string

	/* last actualization of the results */
	Updated time.Time

	/* volume */
	Volume         int64 /* outstanding shares */
	AvgDailyVolume int64 /* avg amount of shares traded */

	/* dividend & related */
	PeRatio          float64
	EarningsPerShare float64
	DividendPerShare float64   /* total (non-special) dividend payout / total amount of shares */
	DividendYield    float64   /* annual div. per share / price per share */
	DividendExDate   time.Time /* last dividend payout date */

	/* price & derived */
	Bid, Ask              float64
	Open, PreviousClose   float64
	LastTradePrice        float64
	Change, ChangePercent float64

	DayLow, DayHigh   float64
	YearLow, YearHigh float64

	Ma50, Ma200 float64 /* 200- and 50-day moving average */
}

type Hist struct {
	Symbol  string
	From    time.Time
	To      time.Time
	Entries []HistEntry
}

type DividendHist struct {
	Symbol    string
	Dividends []DividendEntry
}

type DividendEntry struct {
	Date      util.YearMonthDay
	Dividends float64 `json:",string"`
}

type HistEntry struct {
	Date     util.YearMonthDay `json:"Date"`
	Open     float64           `json:"Open,string"`
	Close    float64           `json:"Close,string"`
	AdjClose float64           `json:"AdjClose,string"`
	High     float64           `json:"High,string"`
	Low      float64           `json:"Low,string"`
	Volume   int64             `json:"Volume,string"`
}

type Source interface {
	Quote(symbols []string) ([]Quote, error)

	Hist(symbols []string) (map[string]Hist, error)
	HistLimit(symbols []string, start time.Time, end time.Time) (map[string]Hist, error)

	DividendHist(symbols []string) (map[string]DividendHist, error)
	DividendHistLimit(symbols []string, start time.Time, end time.Time) (map[string]DividendHist, error)

	fmt.Stringer
}

type Cache interface {
	Source

	SetQuoteExpiry(dur time.Duration)

	HasQuote(symbol string) bool
	HasHist(symbol string, start *time.Time, end *time.Time) bool
}
