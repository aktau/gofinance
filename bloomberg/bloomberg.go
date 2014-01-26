package bloomberg

import (
	"code.google.com/p/go.net/html"
	"fmt"
	"github.com/aktau/gofinance/fquery"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Source struct{}

type bloomQuote struct {
	Name string

	Volume int64

	Open, PrevClose   float64
	DayLow, DayHigh   float64
	YearLow, YearHigh float64
	LastTradePrice    float64

	YearReturn        float64
	PeRatio           float64
	PeRatioEst        float64
	PeRatioRelToIndex float64

	EarningsPerShare    float64
	EarningsPerShareEst float64

	DividendYield    float64
	DividendGrowth5y float64
	DividendExDate   time.Time
}

func New() fquery.Source {
	return &Source{}
}

func (s *Source) Quote(symbols []string) ([]fquery.Result, error) {
	slice := make([]fquery.Result, 0, len(symbols))

	results := make(chan *fquery.Result, len(symbols))
	errors := make(chan error, len(symbols))

	/* fetch all symbols in parallel */
	for _, symbol := range symbols {
		go func(symbol string) {
			quote, err := single(symbol)
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
			slice = append(slice, *r)
		}
	}

	return slice, nil
}

func (s *Source) Hist(symbols []string) (map[string]fquery.Hist, error) {
	return nil, fmt.Errorf(fquery.ErrTplNotSupported, s.String(), "hist")
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

func single(symbol string) (*fquery.Result, error) {
	resp, err := http.Get("http://www.bloomberg.com/quote/" + symbol)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	/* TODO: detect if fund or plain stock, different layouts... */
	quote := &bloomQuote{}
	walk(doc, quote)

	return &fquery.Result{
		Name:             quote.Name,
		Symbol:           symbol,
		Updated:          time.Now(),
		Volume:           quote.Volume,
		Open:             quote.Open,
		PreviousClose:    quote.PrevClose,
		DayLow:           quote.DayLow,
		DayHigh:          quote.DayHigh,
		YearLow:          quote.YearLow,
		YearHigh:         quote.YearHigh,
		LastTradePrice:   quote.LastTradePrice,
		DividendYield:    quote.DividendYield,
		EarningsPerShare: quote.EarningsPerShare,
		DividendExDate:   quote.DividendExDate,
	}, nil
}

/* could be made mode efficient if we parse layer by layer, and specify for
 * each layer what we expect to find, the first layer would for example be
 * html, then body, then div.key_stat, et cetera */
func walk(n *html.Node, b *bloomQuote) bool {
	if n.Type == html.ElementNode {
		switch {
		case n.Data == "div" && hasClass(n, "ticker_header_top"):
			h2 := findFirstChild(n, "h2")
			b.Name = text(h2)
		case n.Data == "div" && hasClass(n, "key_stat"):
			/* select the table that follows */
			bloomtable(findFirstChild(n, "table"), b)
			return false
		case n.Data == "span" && hasClass(n, "price"):
			b.LastTradePrice = atof(strings.TrimSpace(n.FirstChild.Data))
		case n.Data == "table" && hasClass(n, "snapshot_table"):
			bloomsnapshot(n, b)
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if !walk(c, b) {
			return false
		}
	}

	return true
}

func ntostr(n *html.Node) string {
	return fmt.Sprintf("tag: %v, data: %v, attr: %v", n.Type, n.Data, n.Attr)
}

func bloomsnapshot(n *html.Node, b *bloomQuote) {
	n = findFirstChild(n, "tbody")
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if isTag(c, "tr") {
			for trChld := c.FirstChild; trChld != nil; trChld = trChld.NextSibling {
				if !isTag(trChld, "th") {
					continue
				}

				th := trChld
				td := findFirstSibling(th, "td")

				hdr := text(th)
				val := text(td)

				strstr := strings.Contains
				switch {
				case strstr(hdr, "Open"):
					b.Open = atof(val)
				case strstr(hdr, "Previous") && strstr(hdr, "Close"):
					b.PrevClose = atof(val)
				case strstr(hdr, "1-Yr") && strstr(hdr, "Rtn"):
					b.YearReturn = atof(val[1:len(val)-2]) / 100
				case strstr(hdr, "Volume"):
					val := stripchars(val, ",")
					vol, err := strconv.Atoi(val)
					if err != nil {
						fmt.Println("bloomberg: could not read volume", hdr, val)
					} else {
						b.Volume = int64(vol)
					}
				case strstr(hdr, "Day") && strstr(hdr, "Range"):
					pc := strings.Split(val, " - ")
					b.DayLow = atof(pc[0])
					b.DayHigh = atof(pc[1])
				case strstr(hdr, "52wk") && strstr(hdr, "Range"):
					pc := strings.Split(val, " - ")
					b.YearLow = atof(pc[0])
					b.YearHigh = atof(pc[1])
				}

				trChld = td
			}
		}
	}
}

func bloomtable(n *html.Node, b *bloomQuote) {
	/* it seems like the html package likes turning everything into
	 * well-formed html, which makes things (strangely enough) a bit hard for
	 * me, because I can't guess at what it might have done to the source.
	 * It likes to insert <tbody> tags, for example. */
	n = findFirstChild(n, "tbody")
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if isTag(c, "tr") {
			/* decide where the data goes based on the header */
			th := findFirstChild(c, "th")
			td := findFirstSibling(th, "td")

			hdr := text(th)
			val := text(td)

			strstr := strings.Contains
			switch {
			case strstr(hdr, "Current") && strstr(hdr, "P/E"):
				b.PeRatio = atof(val)
			case strstr(hdr, "Estimated") && strstr(hdr, "P/E"):
				b.PeRatioEst = atof(val)
			case strstr(hdr, "Relative") && strstr(hdr, "P/E"):
				b.PeRatioRelToIndex = atof(val)
			case strstr(hdr, "Earnings") && strstr(hdr, "Per") && strstr(hdr, "Share"):
				b.EarningsPerShare = atof(val)
			case strstr(hdr, "Est.") && strstr(hdr, "EPS"):
				b.EarningsPerShareEst = atof(val)
			case strstr(hdr, "Dividend") && strstr(hdr, "Yield"):
				b.DividendYield = atof(val[0:len(val)-2]) / 100
			case strstr(hdr, "Dividend") && strstr(hdr, "Ex-Date"):
				t, err := time.Parse("02/01/2006", val)
				if err != nil {
					fmt.Println("bloomberg: can't parse time, ", val, err)
				} else {
					b.DividendExDate = t
				}
			case strstr(hdr, "Dividend") && strstr(hdr, "Growth") && strstr(hdr, "5"):
				b.DividendGrowth5y = atof(val[0:len(val)-2]) / 100
			}
		}
	}
}

func atof(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func isTag(n *html.Node, tag string) bool {
	return n.Type == html.ElementNode && n.Data == tag
}

func text(n *html.Node) string {
	return n.FirstChild.Data
}

func hasClass(n *html.Node, class string) bool {
	for _, attr := range n.Attr {
		if attr.Key == "class" {
			return inSlice(class, strings.Fields(attr.Val))
		}
	}
	return false
}

func inSlice(needle string, xs []string) bool {
	for _, x := range xs {
		if needle == x {
			return true
		}
	}

	return false
}

func findFirstChild(n *html.Node, tag string) *html.Node {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == tag {
			return c
		}
	}

	return nil
}

func findFirstSibling(n *html.Node, tag string) *html.Node {
	for ; n != nil; n = n.NextSibling {
		if n.Type == html.ElementNode && n.Data == tag {
			return n
		}
	}

	return nil
}

func printTree(n *html.Node, indent string, level int) {
	if n.Type == html.ElementNode {
		fmt.Println(strings.Repeat(indent, level), n.Data)
	} else if n.Type == html.TextNode {
		fmt.Println(strings.Repeat(indent, level), "(text)", n.Data)
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		printTree(c, indent, level+1)
	}
}

func stripchars(str, chr string) string {
	return strings.Map(func(r rune) rune {
		if strings.IndexRune(chr, r) < 0 {
			return r
		}
		return -1
	}, str)
}
