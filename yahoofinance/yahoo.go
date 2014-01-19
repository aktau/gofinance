package yahoofinance

import (
	"encoding/json"
	"fmt"
	"github.com/aktau/gofinance/fquery"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	PublicApiUrl        = "http://query.yahooapis.com/v1/public/yql"
	DatatablesUrl       = "store://datatables.org/alltableswithkeys"
	ChartUrl            = "http://chart.finance.yahoo.com/z?s=AAPL&t=6m&q=l&l=on&z=s&p=m50,m200"
	HistoricalUrl       = "http://ichart.finance.yahoo.com/table.csv?s=%s&c=%s"
	TimeShortFormat     = "Jan 02"
	TimeYearShortFormat = "Jan 02 2006"
)

var (
	year        = time.Now().Format("2006")
	YahooTables = Tables{
		Quotes:     "yahoo.finance.quotes",
		QuotesList: "yahoo.finance.quoteslist",
	}
)

type Source struct{}

type Tables struct {
	Quotes     string
	QuotesList string
}

type ShortTime time.Time

func (jt *ShortTime) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	t, err := time.Parse(TimeShortFormat, s)
	if err != nil {
		return err
	}

	*jt = (ShortTime)(t)
	return nil
}

func (jt ShortTime) GetTime() time.Time {
	return (time.Time)(jt)
}

type YqlJsonQuote struct {
	Name   string `json:"Name"`
	Symbol string `json:"Symbol"`

	Bid            float64 `json:"Bid,string"`
	Ask            float64 `json:"Ask,string"`
	Open           float64 `json:"Open,string"`
	PreviousClose  float64 `json:"PreviousClose,string"`
	LastTradePrice float64 `json:"LastTradePriceOnly,string"`

	Ma50  float64 `json:"FiftydayMovingAverage,string"`
	Ma200 float64 `json:"TwoHundreddayMovingAverage,string"`

	DayLow       float64 `json:"-"`
	DayHigh      float64 `json:"-"`
	YearLow      float64 `json:"-"`
	YearHigh     float64 `json:"-"`
	DaysRangeRaw string  `json:"DaysRange"`
	YearRangeRaw string  `json:"YearRange"`

	ExDividendDate *ShortTime `json:"ExDividendDate"`
}

type YqlJsonResults struct {
	Quote []YqlJsonQuote `json:"quote"`
}

type YqlJsonQuery struct {
	Count   int
	Created time.Time
	Lang    string
	Results YqlJsonResults
}

type YqlJsonResponse struct {
	Query YqlJsonQuery `json:"query"`
}

func (s *Source) Fetch(symbols []string) ([]fquery.Result, error) {
	quotedSymbols := stringMap(func(s string) string {
		return `"` + s + `"`
	}, symbols)
	query := fmt.Sprintf(`SELECT * FROM %s WHERE symbol IN (%s)`,
		YahooTables.Quotes, strings.Join(quotedSymbols, ","))

	fmt.Println("yahoo-finance: query = ", query)
	raw, err := Yql(query)
	if err != nil {
		return nil, err
	}

	fmt.Print(string(raw))

	var resp YqlJsonResponse
	err = json.Unmarshal(raw, &resp)
	if err != nil {
		return nil, err
	}

	results := make([]fquery.Result, 0, len(resp.Query.Results.Quote))
	for _, rawres := range resp.Query.Results.Quote {
		rawres.Process()

		res := fquery.Result{
			Name:           rawres.Name,
			Symbol:         rawres.Symbol,
			Bid:            rawres.Bid,
			Ask:            rawres.Ask,
			Open:           rawres.Open,
			PreviousClose:  rawres.PreviousClose,
			LastTradePrice: rawres.LastTradePrice,
			Ma50:           rawres.Ma50,
			Ma200:          rawres.Ma200,
			DayRange:       fquery.Range{rawres.DayLow, rawres.DayHigh},
			YearRange:      fquery.Range{rawres.YearLow, rawres.YearHigh},
		}
		if rawres.ExDividendDate != nil {
			res.Dividend = fquery.Dividend{ExDate: rawres.ExDividendDate.GetTime()}
		}
		results = append(results, res)
	}

	// return nil, fmt.Errorf(fquery.ErrTplNotSupported, s.String(), "fetch")
	return results, nil
}

/* Retrieves historical stock data for the provided symbol.
 * Historical data includes date, open, close, high, low, volume
 * and adjusted close. */
func (s *Source) Hist(symbols []string, start time.Time, end time.Time) ([][]fquery.Historical, error) {
	/* could alternatively query from yahoo.finance.historicaldata
	 * e.g.: http://query.yahooapis.com/v1/public/yql?q=select%20*%20from%20yahoo.finance.historical */
	for _, symbol := range symbols {
		startYear := ""
		csv := fmt.Sprintf(HistoricalUrl, symbol, startYear)
		query := `SELECT *
				  FROM   csv
				  WHERE  url='%s'
				  AND    columns="Date,Open,High,Low,Close,Volume,AdjClose"`
		Yql(fmt.Sprintf(query, csv))
	}

	return nil, fmt.Errorf(fquery.ErrTplNotSupported, s.String(), "history")
}

func (s *Source) String() string {
	return "Yahoo Finance (YQL)"
}

func Csv() ([]byte, error) {
	return nil, nil
}

func Yql(query string) ([]byte, error) {
	v := url.Values{}
	v.Set("q", query)
	v.Set("format", "json")
	v.Set("env", DatatablesUrl)

	resp, err := http.Get(PublicApiUrl + "?" + v.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	httpBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	/* the first row includes column headers, ignore */
	return httpBody, nil
}

func stringMap(mapping func(string) string, xs []string) []string {
	mxs := make([]string, 0, len(xs))
	for _, s := range xs {
		mxs = append(mxs, mapping(s))
	}
	return mxs
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
