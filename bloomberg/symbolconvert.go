package bloomberg

import (
	"strings"
)

var yahooToBloombergMap = map[string]string{
	"AS": "NA", /* Amsterdam Euronext */
	"BR": "BB", /* Brussels Euronext */
	"L":  "LN", /* London Stock Exchange */
	"MI": "IM", /* Milan */
	"SI": "SP", /* Singapore */
	"DE": "GR", /* Xetra Germany */
	"SA": "BZ", /* Sao Paulo, Brasil */
	"MC": "SM", /* Madrid, Spain */
	"MX": "MM", /* Mexico */
}

var bloombergToYahooMap map[string]string

func init() {
	bloombergToYahooMap = revmap(yahooToBloombergMap)
}

func revmap(orig map[string]string) map[string]string {
	rev := make(map[string]string, len(orig))
	for key, val := range orig {
		rev[val] = key
	}
	return rev
}

func yahooToBloomberg(symbol string) string {
	/* symbols with = are assumed to be currencies */
	if strings.HasSuffix(symbol, "=X") {
		return strings.TrimSuffix(symbol, "=X") + ":CUR"
	}

	/* symbols without a specific exchange are assumed to be US-based */
	if !strings.Contains(symbol, ".") {
		return symbol + ":US"
	}

	return conv(symbol, yahooToBloombergMap, ".", ":")
}

func bloombergToYahoo(symbol string) string {
	/* symbols with = are assumed to be currencies */
	if strings.HasSuffix(symbol, ":CUR") {
		return strings.Split(symbol, ":")[0] + "=X"
	}

	/* US symbols are special since yahoo doesn't suffix them, but bloomberg
	 * does */
	if strings.Contains(symbol, ":US") {
		return strings.Split(symbol, ":")[0]
	}

	return conv(symbol, bloombergToYahooMap, ":", ".")
}

func conv(symbol string, symbmap map[string]string, oldsep string, newsep string) string {
	if strings.Contains(symbol, oldsep) {
		parts := strings.Split(symbol, oldsep)
		if len(parts) != 2 {
			vprintln("bloomberg: can't convert symbol, ", symbol)
			return symbol
		}

		if ext, found := symbmap[parts[1]]; found {
			symbol = parts[0] + newsep + ext
		} else {
			vprintln("bloomberg: unknown symbol extension ", symbol)
		}
	}

	return symbol
}

func convertSymbols(symbols []string) []string {
	c := make([]string, 0, len(symbols))
	for _, symbol := range symbols {
		c = append(c, yahooToBloomberg(symbol))
	}
	return c
}
