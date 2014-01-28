package bloomberg

import (
	"fmt"
	"github.com/aktau/gofinance/fquery"
	"time"
)

var VERBOSITY = 0

type Source struct{}

func New() fquery.Source {
	return &Source{}
}

func (s *Source) Quote(symbols []string) ([]fquery.Quote, error) {
	symbols = convertSymbols(symbols)

	slice := make([]fquery.Quote, 0, len(symbols))

	results := make(chan *fquery.Quote, len(symbols))
	errors := make(chan error, len(symbols))

	/* fetch all symbols in parallel */
	for _, symbol := range symbols {
		go func(symbol string) {
			quote, err := getQuote(symbol)
			if err != nil {
				errors <- err
			} else {
				results <- quote
			}
		}(symbol)
	}

	for i := 0; i < len(symbols); i++ {
		select {
		case err := <-errors:
			fmt.Println("bloomberg: error while fetching,", err)
		case r := <-results:
			r.Symbol = bloombergToYahoo(r.Symbol)
			slice = append(slice, *r)
		}
	}

	return slice, nil
}

func (s *Source) Hist(symbols []string) (map[string]fquery.Hist, error) {
	symbols = convertSymbols(symbols)

	m := make(map[string]fquery.Hist, 0)

	results := make(chan *fquery.Hist, len(symbols))
	errors := make(chan error, len(symbols))

	/* fetch all symbols in parallel */
	for _, symbol := range symbols {
		go func(symbol string) {
			quote, err := getHist(symbol)
			if err != nil {
				errors <- err
			} else {
				results <- quote
			}
		}(symbol)
	}

	for i := 0; i < len(symbols); i++ {
		select {
		case err := <-errors:
			fmt.Println("bloomberg: hist error,", err)
		case r := <-results:
			r.Symbol = bloombergToYahoo(r.Symbol)
			m[r.Symbol] = *r
		}
	}

	return m, nil
}

func (s *Source) HistLimit(symbols []string, start time.Time, end time.Time) (map[string]fquery.Hist, error) {
	return nil, fmt.Errorf(fquery.ErrTplNotSupported, s.String(), "histlimit")
}

func (s *Source) DividendHist(symbols []string) (map[string]fquery.DividendHist, error) {
	return nil, fmt.Errorf(fquery.ErrTplNotSupported, s.String(), "dividendhist")
}

func (s *Source) DividendHistLimit(symbols []string, start time.Time, end time.Time) (map[string]fquery.DividendHist, error) {
	return nil, fmt.Errorf(fquery.ErrTplNotSupported, s.String(), "dividendhistlimi")
}

func (s *Source) String() string {
	return "Bloomberg"
}

func vprintln(a ...interface{}) (int, error) {
	if VERBOSITY > 0 {
		return fmt.Println(a...)
	}

	return 0, nil
}

func vprintf(format string, a ...interface{}) (int, error) {
	if VERBOSITY > 0 {
		return fmt.Printf(format, a...)
	}

	return 0, nil
}
