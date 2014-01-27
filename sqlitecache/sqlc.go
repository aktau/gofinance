package sqlitecache

import (
	"database/sql"
	"fmt"
	"github.com/aktau/gofinance/fquery"
	"github.com/aktau/gofinance/util"
	"github.com/coopernurse/gorp"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

/* a version of the fquery.HistEntry struct with a symbol appended,
 * so we can query it easier. */
type dbHistEntry struct {
	Symbol   string
	Date     time.Time
	Open     float64
	Close    float64
	AdjClose float64
	High     float64
	Low      float64
	Volume   int64
}

type dbHistMeta struct {
	Symbol string
	Amount int64

	MinDate int64
	MaxDate int64

	/* processed versions of MinDate/MaxDate */
	rMinDate time.Time
	rMaxDate time.Time
}

var (
	VERBOSITY = 0
)

type SqliteCache struct {
	fquery.Source

	gorp        *gorp.DbMap
	quoteExpiry time.Duration
}

func New(path string, src fquery.Source) (*SqliteCache, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.SqliteDialect{}}
	if VERBOSITY >= 2 {
		dbmap.TraceOn("", log.New(os.Stdout, "dbmap: ", log.Lmicroseconds))
	}

	c := &SqliteCache{src, dbmap, 30 * time.Second}

	c.gorp.AddTableWithName(fquery.Quote{}, "quotes").SetKeys(false, "Symbol")
	c.gorp.AddTableWithName(dbHistEntry{}, "histquotes").SetKeys(false, "Symbol", "Date")

	err = c.gorp.CreateTablesIfNotExists()
	if err != nil {
		c.Close()
		return nil, err
	}

	/* support date range queries over all symbols */
	_, err = c.gorp.Exec(`CREATE INDEX IF NOT EXISTS hq_date_idx ON histquotes (Date)`)
	if err != nil {
		c.Close()
		return nil, err
	}

	return c, nil
}

func (c *SqliteCache) SetQuoteExpiry(dur time.Duration) {
	c.quoteExpiry = dur
}

func (c *SqliteCache) HasQuote(symbol string) bool {
	return false
}

func (c *SqliteCache) HasHist(symbol string, start *time.Time, end *time.Time) bool {
	return false
}

func (c *SqliteCache) Close() error {
	return c.gorp.Db.Close()
}

/* TODO: escape your strings, symbols could be user input */
func (c *SqliteCache) Quote(symbols []string) ([]fquery.Quote, error) {
	/* fetch all the quotes we have */
	quotedSymbols := quoteSymbols(symbols)

	var results []fquery.Quote
	cutoff := time.Now().Add(-c.quoteExpiry)
	_, err := c.gorp.Select(&results,
		fmt.Sprintf(
			"SELECT * FROM quotes WHERE Symbol IN (%v) AND Updated >= datetime(%v, 'unixepoch')",
			strings.Join(quotedSymbols, ","), cutoff.Unix()))
	if err != nil {
		/* if an error occured, just patch through to the source */
		vprintln("sqlitecache: error while fetching quotes, ", err, ", will use underlying source")
		return c.Source.Quote(symbols)
	}

	/* in case no error occured, check which ones were not in the cache,
	 * they need to be added to the list of quotes to fetch from the src */
	quoteMap := fquery.QuotesToMap(results)
	toFetch := make([]string, 0, len(symbols))
	for _, symbol := range symbols {
		if _, ok := quoteMap[symbol]; !ok {
			toFetch = append(toFetch, symbol)
			vprintln(symbol, "was NOT fetched from cache!")
		} else {
			vprintln(symbol, "was fetched from cache!")
		}
	}

	/* fetch all missing items, store in cache and add to the results we
	 * already got from the cache */
	fetched, err := c.Source.Quote(toFetch)
	if err != nil {
		vprintf("sqlitecache: error while fetching either of %v: %v\n", toFetch, err)
		return results, nil
	}

	err = c.mergeQuotes(fetched...)
	if err != nil {
		vprintf("sqlitecache: error, could not merge quotes of %v into cache, %v\n", toFetch, err)
	}

	results = append(results, fetched...)

	return results, nil
}

/* I consider this to be an extremely dirty function, it should be split up
 * and possibly (hopefully) simplified*/
func (c *SqliteCache) Hist(symbols []string) (map[string]fquery.Hist, error) {
	/* fetch all the historical results we have in the cache */
	symbolSet := sliceToSet(symbols)

	/* find out how many entries we have per symbol (we do this so we can
	 * process the cached symbols and fetch the missing ones in parallel */
	sqliteCastToUnix := func(selector string) string {
		return `strftime('%s',` + selector + `)`
	}
	var meta []*dbHistMeta
	_, err := c.gorp.Select(&meta,
		fmt.Sprintf(
			`SELECT Symbol, count(Symbol) AS Amount,
			        %v AS MinDate,
					%v AS MaxDate
			 FROM histquotes
			 WHERE Symbol IN (%v)
			 GROUP BY Symbol`,
			sqliteCastToUnix("min(Date)"),
			sqliteCastToUnix("max(Date)"),
			strings.Join(quoteSymbols(symbols), ",")))
	if err != nil {
		/* if an error occured, just patch through to the source */
		vprintln("sqlitecache: error while fetching history metadata, ", err, ", will use underlying source")
		return c.Source.Hist(symbols)
	}

	/* convert the timestamps from UNIX timestamps to time.Time */
	for _, m := range meta {
		m.rMinDate = time.Unix(m.MinDate, 0)
		m.rMaxDate = time.Unix(m.MaxDate, 0)
	}

	/* convert the meta list to a map, for easier searching */
	metaMap := make(map[string]*dbHistMeta, len(meta))
	for _, m := range meta {
		metaMap[m.Symbol] = m
	}

	yesterday := Yesterday()
	missing := make([]string, 0, len(symbols))
	for _, s := range symbols {
		m, found := metaMap[s]
		if !found {
			vprintln(s, "was NOT fetched from cache!")
			missing = append(missing, s)
		} else {
			if m.rMaxDate.Before(yesterday) {
				vprintln(s, "was NOT fetched from cache! (historical prices "+
					"not recent enough), last available date was",
					m.rMaxDate.Format("02/01/2006"))
				missing = append(missing, s)
				delete(symbolSet, s)
			} else {
				vprintln(s, "was fetched from cache!")
			}
		}
	}

	hist := make(map[string]fquery.Hist, len(symbols))
	var (
		histmutex sync.Mutex
		wg        sync.WaitGroup
	)

	if len(symbolSet) > 0 {
		adjSymbols := setToSlice(symbolSet)
		vprintln("fetching", adjSymbols, "from the SQLite db")

		/* now query the actual entries */
		var results []dbHistEntry
		_, err = c.gorp.Select(&results,
			fmt.Sprintf(
				`SELECT * FROM histquotes WHERE Symbol IN (%v) ORDER BY Symbol, Date`,
				strings.Join(quoteSymbols(adjSymbols), ",")))
		if err != nil {
			/* if an error occured, just patch through to the source */
			vprintln("sqlitecache: error while fetching historical quotes, ", err, ", will use underlying source")
			return c.Source.Hist(symbols)
		}

		/* process the dbHistEntries into fquery.Hist structures */
		wg.Add(1)
		go func() {
			defer wg.Done()

			var (
				cursymb string
				curhist fquery.Hist
			)
			for _, e := range results {
				/* detect transitions between symbols */
				if cursymb != e.Symbol {
					fmt.Println("CHANGING SYMBOL:", cursymb, e.Symbol)
					/* if this is not the first iteration, finish up the hist
					 * entry we've been working on */
					if len(cursymb) != 0 {
						histmutex.Lock()
						hist[cursymb] = curhist
						histmutex.Unlock()
					}

					/* move to the following symbol */
					info := metaMap[e.Symbol]
					cursymb = e.Symbol
					curhist = fquery.Hist{
						Symbol:  e.Symbol,
						From:    info.rMinDate,
						To:      info.rMaxDate,
						Entries: make([]fquery.HistEntry, 0, info.Amount),
					}
				}

				curhist.Entries = append(curhist.Entries, normalizeHistEntry(&e))
			}

			/* add the last one too */
			histmutex.Lock()
			hist[cursymb] = curhist
			histmutex.Unlock()
		}()
	}

	/* fetch the missing entries from the original source */
	if len(missing) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()

			/* fetch the missing symbols */
			missingMap, err := c.Source.Hist(missing)
			if err != nil {
				vprintln("sqlitecache: error occured while fetching missing", missing, "hist. quotes,", err)
			} else {
				missinglist := make([]fquery.Hist, 0, len(missingMap))
				for symbol, history := range missingMap {
					histmutex.Lock()
					hist[symbol] = history
					histmutex.Unlock()

					missinglist = append(missinglist, history)
				}

				/* add the previously missing items to the cache, so we don't have to
				 * request them again */
				c.mergeHistory(missinglist...)
			}
		}()
	}

	wg.Wait()

	return hist, nil
}

/*
func (c *SqliteCache) HistLimit(symbol []string, start time.Time, end time.Time) (map[string]fquery.Hist, error) {
	return nil, fmt.Errorf("not supported")
}

func (c *SqliteCache) DividendHist(symbol []string) (map[string]fquery.DividendHist, error) {
	return nil, fmt.Errorf("not supported")
}

func (c *SqliteCache) DividendHistLimit(symbol []string, start time.Time, end time.Time) (map[string]fquery.DividendHist, error) {
	return nil, fmt.Errorf("not supported")
}
*/

func (c *SqliteCache) String() string {
	return "SQLite cache, backed by: " + c.String()
}

/* TODO: actually use the transaction, or don't even start it... */
func (c *SqliteCache) mergeQuotes(quotes ...fquery.Quote) error {
	if len(quotes) == 0 {
		return nil
	}

	trans, err := c.gorp.Begin()
	if err != nil {
		return err
	}

	for _, quote := range quotes {
		vprintln("merging quote: ", quote.Symbol)
		count, err := c.gorp.Update(&quote)
		if err == nil && count == 1 {
			continue
		}
		vprintln("sqlitecache: error while UPDATE'ing symbol", quote.Symbol,
			"err:", err, ", count:", count)

		/* update didn't work, so try insert */
		err = c.gorp.Insert(&quote)
		if err != nil {
			vprintln("sqlitecache: error while INSERTing", err)
		}
	}

	return trans.Commit()
}

func (c *SqliteCache) mergeHistory(hists ...fquery.Hist) error {
	if len(hists) == 0 {
		return nil
	}

	tx, err := c.gorp.Begin()
	if err != nil {
		return err
	}

	for _, hist := range hists {
		/* delete values we're going to override anyway */
		vprintln("deleting potentially already stored history for symbol",
			hist.Symbol, "from", hist.From, "to", hist.To)
		delquery := `DELETE
				     FROM  histquotes
					 WHERE Symbol = '%v'
					 AND   Date BETWEEN datetime(%v, 'unixepoch')
							    AND     datetime(%v, 'unixepoch')`
		delquery = fmt.Sprintf(delquery, hist.Symbol, hist.From.Unix(), hist.To.Unix())
		_, err := tx.Exec(delquery)
		if err != nil {
			vprintln("sqlitecache: mergeHistory: error while deleting, query:",
				delquery, "error:", err)
		}

		vprintln("merging history:", hist.Symbol)
		for _, entry := range hist.Entries {
			/* insert */
			err := tx.Insert(newDbHistEntry(hist.Symbol, &entry))
			if err != nil {
				vprintln("sqlitecache: mergeHistory: error while INSERTing", err)
			}
		}
	}

	return tx.Commit()
}

func vprintln(a ...interface{}) (int, error) {
	if VERBOSITY > 0 {
		return fmt.Println(a...)
	}

	return 0, nil
}

func vprintf(format string, a ...interface{}) (int, error) {
	if VERBOSITY > 0 {
		return fmt.Printf(format, a...)
	}

	return 0, nil
}

func quoteSymbols(symbols []string) []string {
	return util.MapStr(func(s string) string {
		return `'` + s + `'`
	}, symbols)
}

func normalizeHistEntry(e *dbHistEntry) fquery.HistEntry {
	return fquery.HistEntry{
		Date:     util.YearMonthDay(e.Date),
		Open:     e.Open,
		Close:    e.Close,
		AdjClose: e.AdjClose,
		Low:      e.Low,
		High:     e.High,
		Volume:   e.Volume,
	}
}

func newDbHistEntry(symbol string, e *fquery.HistEntry) *dbHistEntry {
	return &dbHistEntry{
		Symbol:   symbol,
		Date:     time.Time(e.Date),
		Open:     e.Open,
		Close:    e.Close,
		AdjClose: e.AdjClose,
		Low:      e.Low,
		High:     e.High,
		Volume:   e.Volume,
	}
}

func Yesterday() time.Time {
	const day = 24 * time.Hour
	return time.Now().Add(-1 * day)
}

func sliceToSet(xs []string) map[string]bool {
	m := make(map[string]bool, len(xs))
	for _, x := range xs {
		m[x] = true
	}
	return m
}

func setToSlice(m map[string]bool) []string {
	xs := make([]string, len(m))
	idx := 0
	for str := range m {
		xs[idx] = str
		idx++
	}
	return xs
}
