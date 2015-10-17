# domquery

domquery is a lightweight implementation in golang for querying and searching through HTML-DOMs.
Based on CSS selectors it gives the opportunity to get data and values from queried elements.

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
package main

import (
	// ...
	"net/http"
	"fmt"
	"github.com/hschaeidt/domquery"
)

func main() {
	// making a request to some website
	resp, err := http.Get("https://www.google.de/")

	// request successful
	if err == nil {
		// Instantiating a new query object
		q := new(domquery.Query)
		// Loading the HTML into the query
		q.Load(resp.Body)
		// Searching through HTML with CSS-selectors
		result := q.Find(".gb1")

		// Printing results
		for _, tokenChain := range result {
			fmt.Println(tokenChain.Value())
		}

		// Closing the request
		defer resp.Body.Close()
	}
}
```
