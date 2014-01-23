package yahoofinance

import (
	"fmt"
	"github.com/aktau/gofinance/fquery"
	"time"
)

const (
	HistoricalUrl = "http://ichart.finance.yahoo.com/table.csv"
)

const (
	TypeCsv = iota
	TypeYql
)

type Source struct {
	srcType int
}

func NewCvs() fquery.Source {
	return &Source{srcType: TypeCsv}
}

func NewYql() fquery.Source {
	return &Source{srcType: TypeYql}
}

func (s *Source) Quote(symbols []string) ([]fquery.Result, error) {
	switch s.srcType {
	case TypeCsv:
		return csvQuotes(symbols)
	case TypeYql:
		return yqlQuotes(symbols)
	}

	return nil, fmt.Errorf("yahoo finance: unknown backend type: %v", s.srcType)
}

func (s *Source) Hist(symbols []string) (map[string]fquery.Hist, error) {
	return yqlHist(symbols, nil, nil)
}

func (s *Source) HistLimit(symbols []string, start time.Time, end time.Time) (map[string]fquery.Hist, error) {
	return yqlHist(symbols, &start, &end)
}

func (s *Source) DividendHist(symbols []string) (map[string]fquery.DividendHist, error) {
	return yqlDivHist(symbols, nil, nil)
}

func (s *Source) DividendHistLimit(symbols []string, start time.Time, end time.Time) (map[string]fquery.DividendHist, error) {
	return yqlDivHist(symbols, &start, &end)
}

func (s *Source) String() string {
	return "Yahoo Finance"
}
