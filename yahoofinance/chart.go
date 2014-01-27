package yahoofinance

import (
	"net/url"
	"strings"
)

type ChartPeriod string

/* chart parameters described at:
 * https://code.google.com/p/yahoo-finance-managed/wiki/miscapiImageDownload
 * e.g.: http://chart.finance.yahoo.com/z?s=VUSA.AS&t=6m&q=l&l=on&z=l&p=m50,m200&c=VFEM.AS,XMEM.MI */
const (
	ChartUrl = "http://chart.finance.yahoo.com/z?"
)

const (
	Day1      ChartPeriod = "1d"
	Day5                  = "5d"
	Month3                = "3m"
	Month6                = "6m"
	Year1                 = "1y"
	Year2                 = "2y"
	Year5                 = "5y"
	MaxPeriod             = "my"
)

func GenChartUrl(symbol string, period ChartPeriod, compare []string) string {
	const (
		ChartLine   = "l"
		ChartBar    = "b"
		ChartCandle = "c"
	)

	v := url.Values{}
	v.Set("s", symbol)
	v.Set("t", string(period))
	v.Set("c", strings.Join(compare, ","))
	v.Set("q", ChartLine)

	/* p modifiers */
	p := []string{"m50", "m200"}
	v.Set("p", strings.Join(p, ","))
	return ChartUrl + v.Encode()
}
