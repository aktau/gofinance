Gofinance
=========

Financial information retrieval and munging, will support multiple
sources (Yahoo Finance, Bloomberg, ...). Written in Go, based in parts
on [richie rich](https://github.com/aantix/richie_rich).

It's logically composed of a few submodules:

- fquery: provides an interface for querying a financial source
  (`fquery.Source`), but doesn't implement any Source itself.
- yahoofinance: implements _fquery_. Queries **Yahoo Finance** for
  financial data, .
- bloomberg: implements _fquery_. Queries Bloomberg for financial data.
  (NOTE: at the moment it doesn't fetch the same types of data as Yahoo
  Finance, I'm working on a way to derive the missing pieces, and also
  integrate the extra data that Bloomberg gives into fquery, such as 1
  year return %)
- sqlitecache: implements _fquery_. Caches the information returned from
  any `fquery.Source` in a SQLite databse.
- app: a sample application you can compile and run (go build), to see
  what you can do with fquery and its modules.

Requirements
============

- Golang 1.1 (I think, I'm developing on Golang 1.2, but not using any
  features new to 1.2).

Features
========

- Parallel fetching of data (with goroutines)
- (Optional) caching of results (so the sources don't block you),
  configurable expiry time. This also allows for local pre-calculations that
  are too expensive to run on every fetch.

Used libraries
==============

| Package | Description | License |
| --- | --- | --- |
| [code.google.com/p/go.net/html](code.google.com/p/go.net/html) | HTML parser | MIT |
| [github.com/mattn/go-sqlite3](github.com/mattn/go-sqlite3)| Database driver for SQLite | MIT |
| [github.com/coopernurse/gorp](github.com/coopernurse/gorp) | ORM(ish) library| MIT |
| [github.com/wsxiaoys/terminal](github.com/wsxiaoys/terminal) | Ansi terminal colours | BSD (3-clause) |

Todo
====
- bloomberg: Funds/ETFs have a different layout than normal stocks on
  Bloomberg, adapt to that.
- Combine data from Yahoo Finance and Bloomberg. (this should be
  implemented as a Source made out of multiple underlying Sources, like
  the Cache). If Yahoo Finance doesn't have the data on a company,
  Bloomberg usually has it. So there needs to be a function that
  determines if a Quote is "too incomplete". In the same vein, things
  like for example "dividend yield" can sometimes differ quite a bit
  between YF and Bloomberg, we need to find a way to reconcile this.
- Persist historical data locally (avoid getting blocked). This already
  happens for quotes if you query through a cache like the SqliteCache.
- Add an optional ncurses-like userface, for example with termon:
  https://code.google.com/p/termon/
- Plotting in the terminal, how to convert png -> ascii? Can't find easy
  libcaca bindings for now.
- Somehow emulate the google finance stock screener:
  https://www.google.com/finance?ei=8EDhUuCpO4eHwAOklwE#stockscreener
- Calculate dividend yield based on the price you actually paid (in
  aggregate) for stock you own. Why is this handy? For example: suppose
  you paid $20 for a stock that pays out $1 in dividends each year. It's
  dividend yield would be 5% (which is very good). If that stock rises
  to $40 the next year and the dividend payouts stay the same, the
  dividend yield would be 2.5% (which is still good but only half of the
  last time). However, this is not the yield you will be getting if you
  already own the stock (and bought it at $20). The yield would still be
  5%, obviously. So, the yield as it stands is a good measure of what
  you're buying, but not of what you already bought.
- Calculate effective yields and tax drag. This is hard to do (collect
  all the info), but once done it would be wonderful to determine how
  much a fund/stock/ETF is really going to net you. An example: the US
  levies a 30% tax on dividends as far as I know. This can be reduced to
  15% if you or your broker make use of a double-tax treaty. Belgium has
  a 25% dividend tax which is (as of 2014) wholly non-negotiable, you
  always pay it, no matter where else you already paid taxes. In the
  Netherlands it is always 15%. In Ireland it appears to be 20% but I
  don't know if that counts for foreign companies who have their
  "domicile" in Ireland. It's all very confusing. At any rate, the most
  rosy tax climate I could personally get is first 15% (US) then 25%
  (BE). I'm too much of a realist to assume I will get that, but one has
  to take _a_ value. So to calculate the effective yield I would use
  this formula: `yield * tax in source country * tax in receiving
  country`. In the case of for example a company with a dividend yield
  of 2.5%, that would become an effective yield of `2.5% * 0.85 * 0.75 =
  1.6%`. Which is suddenly a whole lot less great. As such, one realizes
  it takes quite a bit more dividend yield to beat a bank account by a
  nice margin. In a slightly worse case, I would also have to pay taxes
  in Holland, as that's where I usually buy shares (on the Amsterdam
  Euronext), so that would become `2.5% * 0.85 * 0.85 * 0.75 = 1.35%`.
  And that's how you hit rock bottom, even with an initially nice
  dividend yield. It gets even worse, if you want to see how your
  purchasing power will evolve, you have to take inflation into account.
  So let's do that. The average inflation in Belgium for 2013 was 1.11%
  (1.43% for Germany). That means that €1 in 2012 is worth `€1 / 1.0111
  = 0.989` in 2013. And so the final formula becomes `2.5% * 0.85 * 0.85
  * 0.75 - 1.11% = 0.24%`. And so, there's nothing left. Note that the
  levied taxes are massively important here. In the optimal case, where
  only Belgium and the US are paid (at the reduced US rate of 15%), we
  get an effective rate of 0.48%, which is not good but already double
  the other one. That said, it is often the case than when viewing yield
  estimates, source tax has already been incorporated! Let's take VUSA
  as an example, a Vanguard tracker for the S&P 500, domiciled in
  Ireland. The US-based variant of this is VOO.  When you look at the
  yield difference between the two you'll notice that VOO has a yield
  of 2.09%, while VUSA has a yield of 1.63%. Why is this less? Because
  the dividend withholding tax has already been deducted (that's 15%
  because Ireland has a treaty with the US). Yet there's bound to be
  some extra cost, as a 15% tax rate would lead to a VUSA yield of 2.09%
  * 0.85 = 1.78%, but the reality says it's 1.63% (which corresponds to
  a "taxation" of 23%). This sounds bad for the european investor. But,
  there's a catch, the US yield is gross, so one has to deduct 30% WHT.
  Which puts the advantage in the European camp. But, every country
  wants another slice of the pie if you're in the EU. So you will in the
  end receive 23% * whatever_your_country_wants%. In my case, Belgium,
  that's 25%. So you lose again. Hopefully the EU will start acting up
  against these practices.
- Dividend payout history (and other indicators) analysis. Even if the
  yield is good, it's possible that the company has just had a really
  bad run, and thus its share price is really low. This _might_ mean
  that it's going to be less succesful (bad), or that the market is just
  in a slump (good). So companies that are just doing bad are to be
  avoiding, more so because they are unlikely to keep up their high
  dividend payments if they are on their way to the poor house. Some
  extra parameters gofinance could use to provide tips:
  - historical dividends (are they growing, for how long?)
  - strong growth and earnings
  - no heavy drops in share price (we can disentangle the "losing
    company" case from the "bear market" case by comparing with the
    general direction of an index, be the index as representative as
    possible).
  Put shortly, fundamentals. One needs to make sure that the dividend is
  not going to be slashed and send the yield to kingdom come.

Copyright
=========

Copyright (c) 2014, Nicolas Hillegeer. All rights reserved.
