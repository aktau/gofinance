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
	PublicApiUrl  = "http://query.yahooapis.com/v1/public/yql"
	DatatablesUrl = "store://datatables.org/alltableswithkeys"
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

	Volume             util.NullInt64 `json:"Volume"`
	AverageDailyVolume util.NullInt64 `json:"AverageDailyVolume"`

	Bid            util.NullFloat64 `json:"Bid"`
	Ask            util.NullFloat64 `json:"Ask"`
	Open           util.NullFloat64 `json:"Open"`
	PreviousClose  util.NullFloat64 `json:"PreviousClose"`
	LastTradePrice util.NullFloat64 `json:"LastTradePriceOnly"`

	Ma50  util.NullFloat64 `json:"FiftydayMovingAverage"`
	Ma200 util.NullFloat64 `json:"TwoHundreddayMovingAverage"`

	DayLow       float64 `json:"-"`
	DayHigh      float64 `json:"-"`
	YearLow      float64 `json:"-"`
	YearHigh     float64 `json:"-"`
	DaysRangeRaw string  `json:"DaysRange"`
	YearRangeRaw string  `json:"YearRange"`

	ExDividendDate   *util.MonthDay   `json:"ExDividendDate"`
	DividendPerShare util.NullFloat64 `json:"DividendShare"`
	EarningsPerShare util.NullFloat64 `json:"EarningsShare"`
	DividendYield    util.NullFloat64 `json:"DividendYield"`
	PeRatio          util.NullFloat64 `json:"PERatio"`
}

/* completes data */
func (q *YqlJsonQuote) Process() {
	/* day and year range */
	pc := strings.Split(q.DaysRangeRaw, " - ")
	if len(pc) == 2 {
		q.DayLow, _ = strconv.ParseFloat(pc[0], 64)
		q.DayHigh, _ = strconv.ParseFloat(pc[1], 64)
	}

	if len(pc) == 2 {
		pc = strings.Split(q.YearRangeRaw, " - ")
		q.YearLow, _ = strconv.ParseFloat(pc[0], 64)
		q.YearHigh, _ = strconv.ParseFloat(pc[1], 64)
	}
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

type YqlJsonSingleQuoteResponse struct {
	Query struct {
		YqlJsonMeta
		Results struct {
			Quote YqlJsonQuote `json:"quote"`
		}
	}
}

type histResult interface {
	Entries() []fquery.HistEntry
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

type respDividendHistory struct {
	Query struct {
		YqlJsonMeta
		Results struct {
			Rows []fquery.DividendEntry `json:"row"`
		}
	}
}

type divHistResult interface {
	Entries() []fquery.DividendEntry
}

func (r *respDividendHistory) Entries() []fquery.DividendEntry {
	return r.Query.Results.Rows
}

func yqlQuotes(symbols []string) ([]fquery.Quote, error) {
	if len(symbols) == 0 {
		return nil, nil
	}

	quotedSymbols := util.MapStr(func(s string) string {
		return `"` + s + `"`
	}, symbols)
	query := fmt.Sprintf(`SELECT * FROM %s WHERE symbol IN (%s)`,
		YahooTables.Quotes, strings.Join(quotedSymbols, ","))
	fmt.Println("Quotes query = ", query)

	raw, err := Yql(query)
	if err != nil {
		return nil, err
	}

	/* json responses for just a single symbols are slightly different from
	 * the ones for multiple symbols. */
	var quotes []YqlJsonQuote
	if len(symbols) == 1 {
		var sresp YqlJsonSingleQuoteResponse
		err = json.Unmarshal(raw, &sresp)
		if err != nil {
			return nil, err
		}
		quotes = []YqlJsonQuote{sresp.Query.Results.Quote}
	} else {
		var resp YqlJsonQuoteResponse
		err = json.Unmarshal(raw, &resp)
		if err != nil {
			return nil, err
		}
		quotes = resp.Query.Results.Quote
	}

	results := make([]fquery.Quote, 0, len(quotes))
	for _, rawres := range quotes {
		rawres.Process()

		res := fquery.Quote{
			Name:             rawres.Name,
			Symbol:           rawres.Symbol,
			Updated:          time.Now(),
			Volume:           int64(rawres.Volume),
			AvgDailyVolume:   int64(rawres.AverageDailyVolume),
			Bid:              float64(rawres.Bid),
			Ask:              float64(rawres.Ask),
			Open:             float64(rawres.Open),
			PreviousClose:    float64(rawres.PreviousClose),
			LastTradePrice:   float64(rawres.LastTradePrice),
			Ma50:             float64(rawres.Ma50),
			Ma200:            float64(rawres.Ma200),
			DayLow:           float64(rawres.DayLow),
			DayHigh:          float64(rawres.DayHigh),
			YearLow:          float64(rawres.YearLow),
			YearHigh:         float64(rawres.YearHigh),
			EarningsPerShare: float64(rawres.EarningsPerShare),
			DividendPerShare: float64(rawres.DividendPerShare),
			DividendYield:    float64(rawres.DividendYield / 100),
			PeRatio:          float64(rawres.PeRatio),
		}
		if rawres.ExDividendDate != nil {
			res.DividendExDate = rawres.ExDividendDate.GetTime()
		}
		results = append(results, res)
	}

	return results, nil
}

func yqlHist(symbols []string, start *time.Time, end *time.Time) (map[string]fquery.Hist, error) {
	if start == nil {
		t := time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)
		start = &t
	}
	startq := fmt.Sprintf(` AND startDate = "%v-%v-%v"`,
		start.Year(), int(start.Month()), start.Day())

	if end == nil {
		t := time.Now()
		end = &t
	}
	endq := fmt.Sprintf(` AND endDate = "%v-%v-%v"`,
		end.Year(), int(end.Month()), end.Day())

	queryGen := func(symbol string) string {
		return fmt.Sprintf(
			`SELECT * FROM yahoo.finance.historicaldata WHERE symbol="%s"`,
			symbol) + startq + endq
	}

	makeMarshal := func() interface{} {
		var resp YqlJsonHistResponse
		return &resp
	}

	res := make(map[string]fquery.Hist)

	parallelFetch(queryGen, makeMarshal, addHistToMap(res), symbols)
	return res, nil
}

func yqlDivHist(symbols []string, start *time.Time, end *time.Time) (map[string]fquery.DividendHist, error) {
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

	/* ask for the dividend history */
	v.Set("g", "v")

	queryGen := func(symbol string) string {
		/* make a copy of the url parameters since we're going to be
		 * modifying it and this will run in parallel */
		params := v
		params.Set("s", symbol)
		csv := HistoricalUrl + "?" + params.Encode()
		return fmt.Sprintf(
			`SELECT * FROM csv(2,0) WHERE url='%s' AND
			columns="Date,Dividends"`, csv)
	}

	makeMarshal := func() interface{} {
		var resp respDividendHistory
		return &resp
	}

	res := make(map[string]fquery.DividendHist)
	/* add will be called serially, so no need for synchronizing */
	add := func(work workPair) {
		if w, ok := work.Result.(divHistResult); ok {
			res[work.Symbol] = fquery.DividendHist{
				Symbol:    work.Symbol,
				Dividends: w.Entries(),
			}
		}
	}

	parallelFetch(queryGen, makeMarshal, add, symbols)
	return res, nil
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

	makeMarshal := func() interface{} {
		var resp YqlJsonPureHistResponse
		return &resp
	}

	res := make(map[string]fquery.Hist)
	parallelFetch(queryGen, makeMarshal, addHistToMap(res), symbols)
	return res, nil
}

func addHistToMap(m map[string]fquery.Hist) func(workPair) {
	return func(work workPair) {
		/* ugh, no generics, at least I could keep it to the parts that
		 * aren't going to happen much */
		if w, ok := work.Result.(histResult); ok {
			entries := w.Entries()
			var (
				from time.Time
				to   time.Time
			)
			if len(entries) > 0 {
				from = entries[len(entries)-1].Date.GetTime()
				to = entries[0].Date.GetTime()
			}
			m[work.Symbol] = fquery.Hist{
				Symbol:  work.Symbol,
				From:    from,
				To:      to,
				Entries: w.Entries(),
			}
		}
	}
}

type workPair struct {
	Symbol string
	Result interface{}
}

func parallelFetch(queryGen func(string) string, makeUnmarshal func() interface{}, add func(workPair), symbols []string) {
	results := make(chan workPair, len(symbols))
	errors := make(chan error, len(symbols))

	for _, symbol := range symbols {
		go func(symbol string) {
			query := queryGen(symbol)
			resp := makeUnmarshal()

			err := fetchAndUnmarshall(query, resp)
			if err != nil {
				errors <- err
			} else {
				results <- workPair{symbol, resp}
			}
		}(symbol)
	}

	for i := 0; i < len(symbols); i++ {
		select {
		case err := <-errors:
			fmt.Println("yql: error while fetching,", err)
		case r := <-results:
			add(r)
		}
	}
}

func fetchAndUnmarshall(query string, target interface{}) error {
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
