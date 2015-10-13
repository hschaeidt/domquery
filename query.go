package main

import (
	"golang.org/x/net/html"
	"io"
	"net/http"
	"fmt"
	"strings"
)

// Represents a DOM-Query
// Also represents recursively the whole query-chain
type Query struct {
	tokenizer *html.Tokenizer //Contains the tokenized HTML DOM

	hasPrevQuery bool
	hasNextQuery bool // Has next query?
	prevQuery *Query
	nextQuery *Query // Next query object

	match map[string]string //from the mapper some value(s)
	result [][]html.Token // Contains token results from the matches, based on these the nextQuery will be executed
}

// Processing the search-term then launching the Document or Token search
func (q *Query) Find(term string) [][]html.Token {
	q.ProcessSearchTerm(term)

	return q.Search()
}

// Takes the decision weither to use RootSearch (DOM) or TokenSearch (List of elements)
// This method also takes care to return the results of the last (sub-)query
func (q *Query) Search() [][]html.Token {
	var result [][]html.Token

	if q.hasPrevQuery {
		result = q.TokenSearch(q.prevQuery.result)
	} else {
		result = q.RootSearch()
	}

	if q.hasNextQuery {
		result = q.nextQuery.Search()
	}

	return result
}

// Root search represents the search's entrypoint via the DOM
// Delegations for different search methods are made upon here
//
// Root search is only executed for the first query in the query-chain
// All subsequent searches are based on a array of previous resulted tokens
func (q *Query) RootSearch() [][]html.Token {
	var finalTokens [][]html.Token

	for {
		// true by default
		success := true
		tokenType := q.tokenizer.Next()

		if tokenType == html.ErrorToken {
			break
		}

		token := q.tokenizer.Token()

		success = q.Match(token, token.Type)

		if success == true {
			tokenChain := q.GetTokenChainFromTokenizer(token)
			finalTokens = append(finalTokens, tokenChain)
		}
	}

	q.result = finalTokens
	return finalTokens
}

func (q *Query) TokenSearch(tokens [][]html.Token) [][]html.Token {
	var finalTokens [][]html.Token

	for _, tokenChain := range tokens {
		for _, token := range tokenChain {
			success := q.Match(token, token.Type)

			if success == true {
				//tokenChain := q.GetTokenChain(token, tokenChain)
				finalTokens = append(finalTokens, tokenChain)
			}
		}
	}

	return finalTokens
}

// Checks for matches from the parsed search-terms for the given Query object
func (q *Query) Match(token html.Token, tokenType html.TokenType) bool {
	success := true

	for domType, domValue := range q.match {
		switch {
		case tokenType == html.ErrorToken:
			return false
		case tokenType == html.StartTagToken:
			hasAttr := q.HasAttr(token, domType, domValue)

			if !hasAttr {
				// Attribute does not match
				success = false
			}
		default:
			success = false
		}
	}

	return success
}

// Checks weither a (HTML) token has requested attribute matching specified value
func (q *Query) HasAttr(token html.Token, attrType string, searchValue string) bool {
	for _, attr := range token.Attr {
		if attr.Key == attrType && attr.Val == searchValue {
			return true
		}
	}

	return false
}

// Makes a snapshot of the whole token-chain (depth) until reaching the root again
// It takes actually the object wide tokenizer object. So each "Next()" has to be
// sended through "SearchTokens" again, in case another inner match may occure
func (q *Query) GetTokenChainFromTokenizer(rootToken html.Token) []html.Token {
	var tokenChain []html.Token
	depth := 1

	// we expect rootToken to be a start-token, so that we can correctly measure the deepness
	// of the result
	if rootToken.Type != html.StartTagToken {
		return nil
	}

	tokenChain = append(tokenChain, rootToken)

	for {
		tokenType := q.tokenizer.Next()

		// just avoid errors
		if tokenType == html.ErrorToken {
			break
		}

		// we're digging one step deeper
		if tokenType == html.StartTagToken {
			depth++
		}

		// and one step out
		if tokenType == html.EndTagToken {
			depth--
		}

		// push new item to our chain
		tokenChain = append(tokenChain, q.tokenizer.Token())

		// by verifiying against smaller than zero we ensure that the loop
		// makes one more turn to get the rootElements EndTagToken too
		// a correct loop will always end in minus one
		//
		// TODO: this may cause errors by requesting self closing tags later on
		if depth < 0 {
			break
		}
	}

	return tokenChain
}

// Splits the searchterm in a consecutive chain of search queries using search-maps
func (q *Query) ProcessSearchTerm(term string) {
	var (
		queries []string
		subQuery *Query
	)

  // Only split into 2 args, because the next query has to handle its
	// own subqueries by itself (recursion)
	queries = strings.SplitN(term, " ", 2)

	q.CreateSearchMap(queries[0])

	// we got subselects
	if len(queries) > 1 {
		subQuery = new(Query)
		// this will chain the recursively for each consecutive sub-query
		subQuery.CreateSearchMap(queries[1])
	}
}

func (q *Query) CreateSearchMap(query string) {
	if q.match == nil {
		q.match = make(map[string]string)
	}

	if strings.HasPrefix(query, ".") {
		q.match["class"] = strings.TrimPrefix(query, ".")
	} else if strings.HasPrefix(query, "#") {
		q.match["id"] = strings.TrimPrefix(query, "#")
	}
}

// Loads the reader's input into tokenized HTML.
// It can be used to iterate through, finding / changing values.
func (q *Query) Load(reader io.Reader) {
	q.tokenizer = html.NewTokenizer(reader);
}

func main() {
	resp, err := http.Get("https://www.google.de/")

	defer resp.Body.Close()

	if err == nil {
		q := new(Query)
		q.Load(resp.Body)

		result := q.Find(".gb1")
		fmt.Println(result)
	}
}
