package fquery

import (
	"fmt"
	"time"

	"github.com/aktau/gofinance/util"
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
	PeRatio          float64   /* Price / EPS */
	EarningsPerShare float64   /* (net income - spec.dividends) / avg.  outstanding shares */
	DividendPerShare float64   /* total (non-special) dividend payout / total amount of shares */
	DividendYield    float64   /* annual div. per share / price per share */
	DividendExDate   time.Time /* last dividend payout date */

	/* price & derived */
	Bid, Ask              float64
	Open, PreviousClose   float64
	LastTradePrice        float64 // The last trace price and the closing price are usually the same thing. If they vary, the closing price should be used as it refers to the last 'on market' traded price.
	Change, ChangePercent float64

	DayLow, DayHigh   float64
	YearLow, YearHigh float64

	Ma50, Ma200 float64 /* 200- and 50-day moving average */
}

/* will try to calculate the dividend payout ratio, if possible,
 * otherwise returns 0 */
func (q *Quote) DivPayoutRatio() float64 {
	/* total dividends / net income (same period, mostly 1Y):
	 * TODO: implement this (first implement historical data
	 * aggregation) */

	/* DPS / EPS */
	if q.DividendPerShare != 0 && q.EarningsPerShare != 0 {
		return q.DividendPerShare / q.EarningsPerShare
	}
	return 0
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
