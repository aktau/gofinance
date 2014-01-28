package bloomberg

import (
	"encoding/json"
	"fmt"
	"github.com/aktau/gofinance/fquery"
	"github.com/aktau/gofinance/util"
	"net/http"
	"time"
)

const (
	HIST_URL = "http://www.bloomberg.com/markets/chart/data/%s/%s"
)

type bloomHistValues [2]float64

type bloomHist struct {
	DataValues []bloomHistValues `json:"data_values"`
}

func getHist(symbol string) (*fquery.Hist, error) {
	url := fmt.Sprintf(HIST_URL, "1Y", symbol)
	vprintln("bloomberg: fetching historical, ", url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("bloomberg: %v, error while fetching, url: %v, error: %v", symbol, url, err)
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("bloomberg: %v, json new decoder error, url: %v, error: %v", symbol, url, err)
	}

	var v bloomHist
	if err := dec.Decode(&v); err != nil {
		return nil, fmt.Errorf("bloomberg: %v, json decode error, url: %v, error: %v", symbol, url, err)
	}

	var (
		minDate time.Time
		maxDate time.Time
	)
	entries := make([]fquery.HistEntry, 0, 365)
	for idx, e := range v.DataValues {
		t := time.Unix(int64(e[0])/1000, 0)

		if idx == 0 {
			minDate = t
		}
		if idx == len(v.DataValues)-1 {
			maxDate = t
		}

		entries = append(entries, fquery.HistEntry{
			Date:  util.YearMonthDay(t),
			Close: e[1],
		})
	}

	return &fquery.Hist{
		Symbol:  symbol,
		From:    minDate,
		To:      maxDate,
		Entries: entries,
	}, nil
}
