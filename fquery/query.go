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

type Dividend struct {
	Yield    float64
	PerShare float64
	ExDate   time.Time
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

	/* dividend */
	Dividend Dividend

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
	From    time.Time
	To      time.Time
	Entries []HistEntry
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

	/* CompanyToTicker(company string, prefExchange string) string */

	fmt.Stringer
}
