Gofinance
=========

Financial information retrieval and munging. Planning to use Yahoo
Finance solely at first. Written in go, based in parts on [richie
rich](https://github.com/aantix/richie_rich) and

Todo
====
- Try to get data from Bloomberg as well: http://www.openbloomberg.com/
  (reduce dependency on Yahoo Finance)
- Persist data locally to avoid overloading the servers and getting
  blocked

Technical
=========

Yahoo Finance
-------------

There are broadly speaking 2 ways to easily get at the Yahoo Finance
data:

1. Query via YQL (Yahoo Query Language), a SQL lookalike. This happens
   through  a HTTP GET request. The response can be XML or JSON, depending
   on the GET parameters.
2. Request a CSV file, also with a HTTP GET request.

The first approach is a bit more high-level, and actually queries the
CSV file of method (2) under the hood. You'd think that requesting the
CSV file would be faster, since it skips a step, but this appears to be
false according to my testing. This could be for two reasons:

1. The YQL server has preferential access to Yahoo's own APIs
2. The YQL server caches the CSV file somehow (and/or is aware when it
   gets refreshed).

At any rate, **gofinance** implements both methods, and performs
requests in parallel. By default the YQL way is used, although it's easy
to switch.

There seems to be a third kind of API, possibly related with the CSV
one, possibly not, it's "described"
[here](http://www.quantshare.com/sa-426-6-ways-to-download-free-intraday-and-tick-data-for-the-us-stock-market),
the same site also lists some other interesting sources. Worth a look.
