package fquery

import (
	"fmt"
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
	Yield  float64
	ExDate time.Time
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

type Historical struct {
	Date time.Time
}

type Source interface {
	Fetch(tickers []string) ([]Result, error)
	Hist(tickers []string, start time.Time, end time.Time) ([][]Historical, error)

	/* CompanyToTicker(company string, prefExchange string) string */

	fmt.Stringer
}
