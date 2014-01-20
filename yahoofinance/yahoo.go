package yahoofinance

import (
	"encoding/json"
	"fmt"
	"github.com/aktau/gofinance/fquery"
	"github.com/aktau/gofinance/util"
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
	HistoricalUrl       = "http://ichart.finance.yahoo.com/table.csv"
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

	ExDividendDate *util.MonthDay `json:"ExDividendDate"`
}

type YqlJsonQuoteResult struct {
	Quote []YqlJsonQuote `json:"quote"`
}

// type YqlJsonHistRow struct {
// 	Date     time.Time `json:"Date"`
// 	Open     float64   `json:"Open,string"`
// 	Close    float64   `json:"Close,string"`
// 	AdjClose float64   `json:"AdjClose,string"`
// 	High     float64   `json:"High,string"`
// 	Low      float64   `json:"Low,string"`
// 	Volume   int64     `json:"Volume,string"`
// }

type YqlJsonHistResult struct {
	Rows []fquery.HistEntry `json:"quote"`
}

type YqlJsonMeta struct {
	Count   int       `json:"count"`
	Created time.Time `json:"created"`
	Lang    string    `json:"lang"`
}

type YqlJsonQuoteQuery struct {
	Count   int                `json:"count"`
	Created time.Time          `json:"created"`
	Lang    string             `json:"lang"`
	Results YqlJsonQuoteResult `json:"results"`
}

type YqlJsonQuoteResponse struct {
	Query YqlJsonQuoteQuery `json:"query"`
}

type YqlJsonHistQuery struct {
	Count   int               `json:"count"`
	Created time.Time         `json:"created"`
	Lang    string            `json:"lang"`
	Results YqlJsonHistResult `json:"results"`
}

type YqlJsonHistResponse struct {
	Query YqlJsonHistQuery `json:"query"`
}

func (s *Source) Fetch(symbols []string) ([]fquery.Result, error) {
	quotedSymbols := stringMap(func(s string) string {
		return `"` + s + `"`
	}, symbols)
	query := fmt.Sprintf(`SELECT * FROM %s WHERE symbol IN (%s)`,
		YahooTables.Quotes, strings.Join(quotedSymbols, ","))

	raw, err := Yql(query)
	if err != nil {
		return nil, err
	}

	fmt.Print(string(raw))

	var resp YqlJsonQuoteResponse
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

	return results, nil
}

func (s *Source) Hist(symbols []string) (map[string]fquery.Hist, error) {
	return hist(symbols, nil, nil)
}

func (s *Source) HistLimit(symbols []string, start time.Time, end time.Time) (map[string]fquery.Hist, error) {
	return hist(symbols, &start, &end)
}

func hist(symbols []string, start *time.Time, end *time.Time) (map[string]fquery.Hist, error) {
	if start == nil {
		t := time.Date(1900, time.January, 1, 0, 0, 0, 0, time.UTC)
		start = &t
	}

	if end == nil {
		t := time.Date(2099, time.January, 1, 0, 0, 0, 0, time.UTC)
		end = &t
	}

	res := make(map[string]fquery.Hist)
	for _, symbol := range symbols {
		query := fmt.Sprintf(
			`SELECT * FROM yahoo.finance.historicaldata WHERE symbol="%s"`,
			symbol)

		query += fmt.Sprintf(` AND startDate = "%v-%v-%v"`,
			start.Year(), int(start.Month()), start.Day())
		query += fmt.Sprintf(` AND endDate = "%v-%v-%v"`,
			end.Year(), int(end.Month()), end.Day())

		fmt.Println("yahoo-finance: query = ", query)
		raw, err := Yql(query)
		if err != nil {
			return nil, err
		}

		fmt.Print("raw:", string(raw))
		var resp YqlJsonHistResponse
		err = json.Unmarshal(raw, &resp)
		if err != nil {
			return nil, err
		}

		fmt.Println("ROWS", resp.Query.Results.Rows)
		fmt.Println("REST", resp.Query)
		res[symbol] = fquery.Hist{
			Entries: resp.Query.Results.Rows,
		}
	}

	return res, nil
}

/* makes yql query directly from the csv-file, instead of via
 * yahoo.financial.historicaldata */
func pureYqlHist(symbols []string, start *time.Time, end *time.Time) (map[string]fquery.Hist, error) {
	v := url.Values{}

	if start != nil {
		v.Set("a", strconv.Itoa(int(start.Month())-1))
		v.Set("b", strconv.Itoa(start.Day()))
		v.Set("c", strconv.Itoa(start.Year()))
	}

	if end != nil {
		v.Set("d", strconv.Itoa(int(end.Month())-1))
		v.Set("e", strconv.Itoa(end.Day()))
		v.Set("f", strconv.Itoa(end.Year()))
	}

	res := make(map[string]fquery.Hist)
	for _, symbol := range symbols {
		v.Set("s", symbol)
		csv := HistoricalUrl + "?" + v.Encode()
		query := fmt.Sprintf(
			`SELECT * FROM csv WHERE url='%s' AND
			columns="Date,Open,High,Low,Close,Volume,AdjClose"`,
			csv)

		fmt.Println("yahoo-finance: query = ", query)
		raw, err := Yql(query)
		if err != nil {
			return nil, err
		}

		fmt.Print("raw:", string(raw))
		var resp YqlJsonHistResponse
		err = json.Unmarshal(raw, &resp)
		if err != nil {
			return nil, err
		}

		res[symbol] = fquery.Hist{
			Entries: resp.Query.Results.Rows,
		}
	}

	return res, nil
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

	url := PublicApiUrl + "?" + v.Encode()
	fmt.Println("yql: firing HTTP GET at ", url)
	resp, err := http.Get(url)
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
