package main

import (
	"fmt"
	"github.com/aktau/gofinance/fquery"
	"github.com/aktau/gofinance/yahoofinance"
)

func main() {
	fmt.Printf("welcome to gofinance %v.%v.%v\n", MAJ_VERSION, MIN_VERSION, MIC_VERSION)

	s := &yahoofinance.Source{}
	calc(s)
	// hist(s)
}

func hist(src fquery.Source) {
	res, err := src.Hist([]string{"VEUR.AS"})
	if err != nil {
		fmt.Println("gofinance: could not fetch history, ", err)
	}

	for symb, hist := range res {
		fmt.Println(symb)
		fmt.Println("===========")
		fmt.Println("Length:", len(hist.Entries))
		for _, row := range hist.Entries {
			fmt.Println("row:", row)
		}
	}

	/* with time limits */
	// start := time.Date(2009, time.November, 1, 0, 0, 0, 0, time.UTC)
	// end := time.Date(2011, time.November, 1, 0, 0, 0, 0, time.UTC)
	// _, err = src.HistLimit([]string{"AAPL", "VEUR.AS", "VJPN.AS"}, start, end)
	// if err != nil {
	// 	fmt.Println("gofinance: could not fetch history, ", err)
	// }
}

func calc(src fquery.Source) {
	res, err := src.Fetch([]string{"VEUR.AS", "VJPN.AS"})
	if err != nil {
		fmt.Println("gofinance: could not fetch, ", err)
		return
	}

	fmt.Println()
	for _, r := range res {
		fmt.Printf("name: %v (%v)\n", r.Name, r.Symbol)
		fmt.Printf("bid/ask: %v/%v (spread: %v)\n", r.Bid, r.Ask, r.Ask-r.Bid)
		fmt.Printf("day low/high: %v/%v\n", r.DayRange.Low, r.DayRange.High)
		fmt.Printf("year low/high: %v/%v\n", r.YearRange.Low, r.YearRange.High)
		fmt.Printf("moving avg. 50/200: %v/%v\n", r.Ma50, r.Ma200)
		fmt.Printf("prevclose/open/lasttrade: %v/%v/%v\n",
			r.PreviousClose, r.Open, r.LastTradePrice)
		fmt.Printf("dividend ex: %v\n", r.Dividend.ExDate)
		fmt.Println("======================")
	}
}
