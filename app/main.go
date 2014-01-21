package main

import (
	"fmt"
	"github.com/aktau/gofinance/fquery"
	"github.com/aktau/gofinance/yahoofinance"
	"math"
)

func main() {
	fmt.Printf("welcome to gofinance %v.%v.%v\n", MAJ_VERSION, MIN_VERSION, MIC_VERSION)

	// s := yahoofinance.NewCvs()
	s := yahoofinance.NewYql()
	divhist(s)
	hist(s)
	calc(s)
}

func divhist(src fquery.Source) {
	res, err := src.DividendHist([]string{"VEUR.AS", "VJPN.AS"})
	if err != nil {
		fmt.Println("gofinance: could not fetch history, ", err)
		return
	}

	fmt.Printf("succesfully fetched %v symbols' dividend history\n", len(res))

	for symb, hist := range res {
		fmt.Println(symb)
		fmt.Println("===========")
		fmt.Println("Length:", len(hist.Dividends))
		for _, row := range hist.Dividends {
			fmt.Println("row:", row)
		}
	}
}

func hist(src fquery.Source) {
	res, err := src.Hist([]string{"VEUR.AS", "VJPN.AS"})
	if err != nil {
		fmt.Println("gofinance: could not fetch history, ", err)
		return
	}

	for symb, hist := range res {
		fmt.Println(symb)
		fmt.Println("===========")
		fmt.Println("Length:", len(hist.Entries))
		for _, row := range hist.Entries {
			fmt.Println("row:", row)
		}
		fmt.Println("Moving average manual calc:", movingAverage(hist))
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

	desiredTxCostPerc := 0.01
	txCost := 9.75

	fmt.Println()
	for _, r := range res {
		nrOfShaderForTxCostPerc := sharesToBuy(r.Ask, txCost, desiredTxCostPerc)

		fmt.Printf("name: %v (%v)\n", r.Name, r.Symbol)
		fmt.Printf("bid/ask: %v/%v (spread: %v)\n", r.Bid, r.Ask, r.Ask-r.Bid)
		fmt.Printf("day low/high: %v/%v\n", r.DayRange.Low, r.DayRange.High)
		fmt.Printf("year low/high: %v/%v\n", r.YearRange.Low, r.YearRange.High)
		fmt.Printf("moving avg. 50/200: %v/%v\n", r.Ma50, r.Ma200)
		fmt.Printf("prevclose/open/lasttrade: %v/%v/%v\n",
			r.PreviousClose, r.Open, r.LastTradePrice)
		fmt.Printf("dividend ex: %v, yield: %v, per share: %v\n",
			r.Dividend.ExDate, r.Dividend.Yield, r.Dividend.PerShare)
		fmt.Printf("You would need to buy %v (€ %v) shares of this stock to reach a transaction cost below %v%%\n",
			nrOfShaderForTxCostPerc, nrOfShaderForTxCostPerc*r.Ask, desiredTxCostPerc*100)
		fmt.Print("Richie Rich thinks this is in a ")
		if wouldRichieRichBuy(r) {
			fmt.Println("BUY position")
		} else {
			fmt.Println("SELL position")
		}
		fmt.Println("======================")
	}
}

/* gives you the number of shares to buy if you want
 * the transaction cost to be less than a certain percentage
 * (0.5% is fantastic, 1% is ok, for example) */
func sharesToBuy(price, txCost, desiredTxCostPerc float64) float64 {
	return math.Ceil((txCost - desiredTxCostPerc*txCost) /
		(desiredTxCostPerc * price))
}

/* you can get the moving average for 50 and 200 days
 * out of the standard stock quote, but if you need
 * something else, just point it at this function */
func movingAverage(hist fquery.Hist) float64 {
	if len(hist.Entries) == 0 {
		return 0
	}

	var sum float64 = 0
	var count float64 = 0

	for _, row := range hist.Entries {
		sum += row.Close
		count++
	}

	return sum / count
}

func wouldRichieRichBuy(res fquery.Result) bool {
	return res.PreviousClose > res.Ma200
}
