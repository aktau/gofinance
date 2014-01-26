package fquery

func QuotesToMap(quotes []Quote) map[string]*Quote {
	m := make(map[string]*Quote)
	for _, quote := range quotes {
		m[quote.Symbol] = &quote
	}
	return m
}
