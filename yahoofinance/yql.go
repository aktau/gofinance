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

type YqlJsonMeta struct {
	Count   int       `json:"count"`
	Created time.Time `json:"created"`
	Lang    string    `json:"lang"`
}

type YqlJsonQuoteResponse struct {
	Query struct {
		YqlJsonMeta
		Results struct {
			Quote []YqlJsonQuote `json:"quote"`
		}
	}
}

type YqlJsonHistResponse struct {
	Query struct {
		YqlJsonMeta
		Results struct {
			Rows []fquery.HistEntry `json:"quote"`
		}
	}
}

func (r *YqlJsonHistResponse) Entries() []fquery.HistEntry {
	return r.Query.Results.Rows
}

type YqlJsonPureHistResponse struct {
	Query struct {
		YqlJsonMeta
		Results struct {
			Rows []fquery.HistEntry `json:"row"`
		}
	}
}

func (r *YqlJsonPureHistResponse) Entries() []fquery.HistEntry {
	return r.Query.Results.Rows
}

func yqlQuotes(symbols []string) ([]fquery.Result, error) {
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

func yqlHist(symbols []string, start *time.Time, end *time.Time) (map[string]fquery.Hist, error) {
	if start == nil {
		t := time.Date(1900, time.January, 1, 0, 0, 0, 0, time.UTC)
		start = &t
	}
	startq := fmt.Sprintf(` AND startDate = "%v-%v-%v"`,
		start.Year(), int(start.Month()), start.Day())

	if end == nil {
		t := time.Date(2099, time.January, 1, 0, 0, 0, 0, time.UTC)
		end = &t
	}
	endq := fmt.Sprintf(` AND endDate = "%v-%v-%v"`,
		end.Year(), int(end.Month()), end.Day())

	queryGen := func(symbol string) string {
		return fmt.Sprintf(
			`SELECT * FROM yahoo.finance.historicaldata WHERE symbol="%s"`,
			symbol) + startq + endq
	}

	makeMarshal := func() histResult {
		var resp YqlJsonHistResponse
		return &resp
	}

	return parallelHistFetch(queryGen, makeMarshal, symbols), nil
}

/* makes yql query directly from the csv-file, instead of via
 * the yahoo.financial.historicaldata predefined table */
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

	queryGen := func(symbol string) string {
		/* make a copy of the url parameters since we're going to be
		 * modifying it and this will run in parallel */
		params := v
		params.Set("s", symbol)
		csv := HistoricalUrl + "?" + params.Encode()
		return fmt.Sprintf(
			`SELECT * FROM csv(2,0) WHERE url='%s' AND
			columns="Date,Open,High,Low,Close,Volume,AdjClose"`,
			csv)
	}

	makeMarshal := func() histResult {
		var resp YqlJsonPureHistResponse
		return &resp
	}

	return parallelHistFetch(queryGen, makeMarshal, symbols), nil
}

type histResult interface {
	Entries() []fquery.HistEntry
}

func parallelHistFetch(queryGen func(string) string, makeUnmarshal func() histResult, symbols []string) map[string]fquery.Hist {
	res := make(map[string]fquery.Hist)

	results := make(chan fquery.Hist, len(symbols))
	errors := make(chan error, len(symbols))

	for _, symbol := range symbols {
		go func(symbol string) {
			query := queryGen(symbol)
			resp := makeUnmarshal()

			err := histFetchAndUnmarshall(query, resp)
			if err != nil {
				errors <- err
			} else {
				results <- fquery.Hist{
					Entries: resp.Entries(),
					Symbol:  symbol,
				}
			}
		}(symbol)
	}

	for i := 0; i < len(symbols); i++ {
		select {
		case err := <-errors:
			fmt.Println("yql: error while fetching history,", err)
		case r := <-results:
			res[r.Symbol] = r
		}
	}

	return res
}

func histFetchAndUnmarshall(query string, target histResult) error {
	fmt.Println("yahoo-finance: query = ", query)
	raw, err := Yql(query)
	if err != nil {
		return err
	}

	err = json.Unmarshal(raw, target)
	if err != nil {
		return err
	}

	return nil
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
