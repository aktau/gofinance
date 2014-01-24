package fquery

func QuotesToMap(quotes []Result) map[string]*Result {
	m := make(map[string]*Result)
	for _, quote := range quotes {
		m[quote.Symbol] = &quote
	}
	return m
}
