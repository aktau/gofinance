package bloomberg

import (
	"fmt"
	"strings"
)

var yahooToBloombergMap = map[string]string{
	"AS": "NA",
	"BR": "BB",
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
	return conv(symbol, yahooToBloombergMap, ".", ":")
}

func bloombergToYahoo(symbol string) string {
	return conv(symbol, bloombergToYahooMap, ":", ".")
}

func conv(symbol string, symbmap map[string]string, oldsep string, newsep string) string {
	fmt.Println("CONVERTING, MAP", symbmap)
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
