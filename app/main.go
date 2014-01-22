package main

import (
	"fmt"
	"github.com/aktau/gofinance/fquery"
	"github.com/aktau/gofinance/yahoofinance"
	"github.com/wsxiaoys/terminal"
	"github.com/wsxiaoys/terminal/color"
	"math"
)

func main() {
	fmt.Printf("welcome to gofinance %v.%v.%v\n", MAJ_VERSION, MIN_VERSION, MIC_VERSION)

	// s := yahoofinance.NewCvs()
	s := yahoofinance.NewYql()
	divhist(s)
	// hist(s)
	calc(s)
}

func divhist(src fquery.Source) {
	res, err := src.DividendHist([]string{"BELG.BR", "TNET.BR"})
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
			fmt.Println("row:", row.Date.GetTime().Format("02-01-2006"), row.Dividends)
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
	fmt.Println("requesting information on individual stocks...")
	res, err := src.Fetch([]string{
		"VEUR.AS",
		"VJPN.AS",
		"VHYL.AS",
		"AAPL",
		"APC.F",
		// "KO",
		"SAN.MC",
		"BELG.BR",
		"TNET.BR",
	})
	if err != nil {
		fmt.Println("gofinance: could not fetch, ", err)
		return
	}

	desiredTxCostPerc := 0.01
	txCost := 9.75
	maxBidAskSpreadPerc := 0.01
	minDivYield := 0.025

	fmt.Println()
	for _, r := range res {
		nrOfShaderForTxCostPerc := sharesToBuy(r.Ask, txCost, desiredTxCostPerc)
		bidAskSpreadPerc := (r.Ask - r.Bid) / r.Bid

		upDir := r.LastTradePrice >= r.PreviousClose
		upVal := r.LastTradePrice - r.PreviousClose
		upPerc := upVal / r.PreviousClose * 100
		fmt.Printf("name: %v (%v), %v %v %v\n",
			r.Name, r.Symbol,
			binary(fmt.Sprintf("%+.2f", upVal), upDir),
			binary(fmt.Sprintf("%+.2f%%", upPerc), upDir),
			binary(arrow(upDir), upDir))
		bidAskPrint := binary(fmt.Sprintf("%.3f%%", bidAskSpreadPerc*100), bidAskSpreadPerc < maxBidAskSpreadPerc)
		terminal.Stdout.Colorf("bid/ask: @m%v@|/@m%v@|, spread: @m%.3f@| (%v)\n", r.Bid, r.Ask, r.Ask-r.Bid, bidAskPrint)
		if bidAskSpreadPerc < maxBidAskSpreadPerc {
			fmt.Printf("if you want to buy this stock, place a %v at about %v\n", green("limit order"), green("%.2f", (r.Ask+r.Bid)/2))
		} else {
			fmt.Println(red("be cautious, the spread of this stock is rather high"))
		}
		terminal.Stdout.Colorf("prevclose/open/lasttrade: @{m}%v@{|}/@{m}%v@{|}/@{m}%v@{|}\n",
			r.PreviousClose, r.Open, r.LastTradePrice)
		terminal.Stdout.Colorf("day low/high: @{m}%v@{|}/@{m}%v@{|} (@m%.2f@|)\n", r.DayRange.Low, r.DayRange.High, r.DayRange.Diff())
		terminal.Stdout.Colorf("year low/high: @{m}%v@{|}/@{m}%v@{|} (@m%.2f@|)\n", r.YearRange.Low, r.YearRange.High, r.YearRange.Diff())
		terminal.Stdout.Colorf("moving avg. 50/200: @{m}%v@{|}/@{m}%v@{|}\n", r.Ma50, r.Ma200)
		DivYield := binary(fmt.Sprintf("%.2f%%", r.Dividend.Yield*100), r.Dividend.Yield > minDivYield)
		fmt.Println("PURE YIELD", r.Dividend.Yield)
		terminal.Stdout.Colorf("last ex-dividend: @{m}%v@{|}, div. per share: @{m}%v@{|}, div. yield: %v,\n earnings per share: @m%.2f@|, dividend payout ratio: @m%.2f@|\n",
			r.Dividend.ExDate.Format("02/01"), r.Dividend.PerShare, DivYield, r.EarningsPerShare, r.Dividend.PerShare/r.EarningsPerShare)
		terminal.Stdout.Colorf("You would need to buy @{m}%v@{|} (€ @{m}%.2f@{|}) shares of this stock to reach a transaction cost below %v%%\n",
			nrOfShaderForTxCostPerc, nrOfShaderForTxCostPerc*r.Ask, desiredTxCostPerc*100)
		fmt.Print("Richie Rich thinks this is in a ")
		if wouldRichieRichBuy(r) {
			terminal.Stdout.Colorf("@{g}BUY@{|}")
		} else {
			terminal.Stdout.Colorf("@{r}SELL@{|}")
		}
		fmt.Println(" position")
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

/* prints either green or red text to the screen, depending
 * on decision. */
func binary(text string, decision bool) string {
	col := "@g"
	if !decision {
		col = "@r"
	}
	return color.Sprint(col, text)
}

func green(format string, a ...interface{}) string {
	return color.Sprintf("@g"+format, a...)
}

func red(format string, a ...interface{}) string {
	return color.Sprintf("@r"+format, a...)
}

func arrow(decision bool) string {
	if decision {
		return "↑"
	} else {
		return "↓"
	}
}
