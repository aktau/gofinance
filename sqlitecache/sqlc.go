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
	"time"
)

var (
	VERBOSITY = 0
)

type SqliteCache struct {
	fquery.Source

	gorp *gorp.DbMap
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

	c := &SqliteCache{src, dbmap}

	c.gorp.AddTableWithName(fquery.Result{}, "quotes").SetKeys(false, "Symbol")
	c.gorp.AddTableWithName(fquery.HistEntry{}, "histquotes")

	err = dbmap.CreateTablesIfNotExists()
	if err != nil {
		c.Close()
		return nil, err
	}

	return c, nil
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

/* TODO: query on datetime for more recent results only
 * TODO: escape your strings, symbols could be user input */
func (c *SqliteCache) Quote(symbols []string) ([]fquery.Result, error) {
	/* fetch all the quotes we have */
	quotedSymbols := util.MapStr(func(s string) string {
		return `"` + s + `"`
	}, symbols)

	var results []fquery.Result
	_, err := c.gorp.Select(&results,
		fmt.Sprintf("SELECT * FROM quotes WHERE Symbol IN (%v)",
			strings.Join(quotedSymbols, ",")))
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

/*
func (c *SqliteCache) Hist(symbol []string) (map[string]fquery.Hist, error) {
	return nil, fmt.Errorf("not supported")
}

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

/* TODO: this doesn't actually merge yet, just insert */
func (c *SqliteCache) mergeQuotes(quotes ...fquery.Result) error {
	if len(quotes) == 0 {
		return nil
	}

	trans, err := c.gorp.Begin()
	if err != nil {
		return err
	}

	for _, quote := range quotes {
		vprintln("merging quote: ", quote.Symbol)
		err := c.gorp.Insert(&quote)
		vprintln("error?", err)
	}

	return trans.Commit()
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
