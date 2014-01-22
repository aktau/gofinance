package fquery

import (
	"fmt"
	"github.com/aktau/gofinance/util"
	"time"
)

const (
	ErrTplNotSupported = "source '%s' does not support action '%s'"
)

type Exchange struct {
	Name  string /* eg.: Amsterdam */
	Short string /* eg.: AS */
}

type Range struct {
	Low  float64
	High float64
}

func (r Range) Diff() float64 {
	return r.High - r.Low
}

/* the dividend yield is a very important metric, in the absence of capital
 * gains (the appreciation of the price of the shares), it is the only return
 * on investment of a stock.
 *
 * An example, if X stock trades at $20, and Y stock trades at $40, and they
 * both pay out a dividend of $1 per share, then X has a yield of 0.05 while
 * Y has a yield of 0.025 */
type Dividend struct {
	PerShare float64   /* total (non-special) dividend payout / total amount of shares */
	Yield    float64   /* annual div. per share / price per share */
	ExDate   time.Time /* last dividend payout date */
}

type Result struct {
	Symbol   string /* e.g.: VEUR.AS, Vanguard dev. europe on Amsterdam */
	Name     string
	Exchange Exchange

	/* last actualization of the results */
	Update time.Time

	/* volume */
	Volume         int64 /* outstanding shares */
	AvgDailyVolume int64 /* avg amount of shares traded */

	/* dividend & related */
	Dividend         Dividend
	EarningsPerShare float64

	/* price */
	Bid            float64
	Ask            float64
	Open           float64
	PreviousClose  float64
	LastTradePrice float64

	DayRange  Range
	YearRange Range

	Ma50  float64 /* 50-day moving average */
	Ma200 float64 /* 200-day moving average */

	/* derived from price */
	Change        float64
	ChangePercent float64
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
	Fetch(tickers []string) ([]Result, error)

	Hist(tickers []string) (map[string]Hist, error)
	HistLimit(tickers []string, start time.Time, end time.Time) (map[string]Hist, error)

	DividendHist(tickers []string) (map[string]DividendHist, error)
	DividendHistLimit(tickers []string, start time.Time, end time.Time) (map[string]DividendHist, error)

	/* CompanyToTicker(company string, prefExchange string) string */

	fmt.Stringer
}
