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
