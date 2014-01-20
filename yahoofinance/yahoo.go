package yahoofinance

import (
	"github.com/aktau/gofinance/fquery"
	"strconv"
	"strings"
	"time"
)

const (
	// PublicApiUrl        = "http://query.yahooapis.com/v1/public/yql"
	// DatatablesUrl       = "store://datatables.org/alltableswithkeys"
	// ChartUrl            = "http://chart.finance.yahoo.com/z?s=AAPL&t=6m&q=l&l=on&z=s&p=m50,m200"
	HistoricalUrl = "http://ichart.finance.yahoo.com/table.csv"
	// TimeShortFormat     = "Jan 02"
	// TimeYearShortFormat = "Jan 02 2006"
)

// var (
// 	year        = time.Now().Format("2006")
// 	YahooTables = Tables{
// 		Quotes:     "yahoo.finance.quotes",
// 		QuotesList: "yahoo.finance.quoteslist",
// 	}
// )

type Source struct{}

func (s *Source) Fetch(symbols []string) ([]fquery.Result, error) {
	return yqlQuotes(symbols)
}

func (s *Source) Hist(symbols []string) (map[string]fquery.Hist, error) {
	return yqlHist(symbols, nil, nil)
}

func (s *Source) HistLimit(symbols []string, start time.Time, end time.Time) (map[string]fquery.Hist, error) {
	return yqlHist(symbols, &start, &end)
}

func (s *Source) String() string {
	return "Yahoo Finance (YQL)"
}

/* completes data */
func (q *YqlJsonQuote) Process() {
	/* day and year range */
	pc := strings.Split(q.DaysRangeRaw, " - ")
	q.DayLow, _ = strconv.ParseFloat(pc[0], 64)
	q.DayHigh, _ = strconv.ParseFloat(pc[1], 64)

	pc = strings.Split(q.YearRangeRaw, " - ")
	q.YearLow, _ = strconv.ParseFloat(pc[0], 64)
	q.YearHigh, _ = strconv.ParseFloat(pc[1], 64)
}
