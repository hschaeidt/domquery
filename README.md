# domquery

domquery is a lightweight implementation in golang for querying and searching through HTML-DOMs.
Based on the jQuery-like selectors it gives the opportunity to get data and values from queried elements.

domquery is heavely based on golangs [html package](https://godoc.org/golang.org/x/net/html) for searching and parsing through HTML in form of HTML Tokens.
Elements returned from the lib are objects of type html.Token (golangs html package) or tokenutil.Chain (github.com/hschaeidt/domquery/tokenutil)

# Usage

To use the library, first import the main package into your code

```go
// github.com/me/application_where_to_use_domquery/main.go
package main

import (
	// ...
	"github.com/hschaeidt/domquery"
)
```

Then by creating a new query object it is possible to iterate through the loaded HTML

```go
// ...
// Initializing a query
q := new(domquery.Query)
q.Load(myIOReader)

// Finding HTML
result := q.Find(".myClass")

// result is a 2D array of values and token chains
for _, tokenChain := range result {
	value, chain := tokenChain.Value()
	fmt.Println(value) // prints the value of the text token if available
	if chain != nil {
		// do something with the sub-chain
	}
}

// prints all results to stdout
// ...
```