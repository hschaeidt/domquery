package main

import (
	"golang.org/x/net/html"
	"io"
	"net/http"
	"fmt"
	"strings"
)

// Contains the tokenized HTML DOM
var tokenizer *html.Tokenizer

// Represents a DOM-Query
// Also represents recursively the whole query-chain
type Query struct {
	hasPrevQuery bool
	hasNextQuery bool // Has next query?
	prevQuery *Query
	nextQuery *Query // Next query object
	
	match map[string]string //from the mapper some value(s)
	result []html.Token // Contains token results from the matches, based on these the nextQuery will be executed
}

// Processing the search-term then launching the Document or Token search
func (q *Query) Find(term string) []html.Token {
	q.ProcessSearchTerm(term)
	
	return q.Search()
}

// Takes the decision weither to use RootSearch (DOM) or TokenSearch (List of elements)
// This method also takes care to return the results of the last (sub-)query
func (q *Query) Search() []html.Token {
	var result []html.Token
	
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
func (q *Query) RootSearch() []html.Token {
	var finalTokens []html.Token
	
	for {
		// true by default
		success := true
		tokenType := tokenizer.Next()
		
		if tokenType == html.ErrorToken {
			return finalTokens
		}
		
		token := tokenizer.Token()
		
		success = q.Match(token, token.Type)
		
		if success == true {
			finalTokens = append(finalTokens, token)
		}
	}
	
	q.result = finalTokens
	return finalTokens
}

func (q *Query) TokenSearch(tokens []html.Token) []html.Token {
	var finalTokens []html.Token
	
	for _, token := range tokens {
		success := q.Match(token, token.Type)
		
		if success == true {
			finalTokens = append(finalTokens, token)
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

// Splits the searchterm in a consecutive chain of search queries using search-maps
func (q *Query) ProcessSearchTerm(term string) {
	var (
		queries []string
		subQuery *Query
	)
	
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
func Load(reader io.Reader) {
	tokenizer = html.NewTokenizer(reader);
}

func main() {
	resp, err := http.Get("https://www.google.de/")
	
	defer resp.Body.Close()
	
	if err == nil {
		Load(resp.Body)
	
		q := new(Query)
	
		result := q.Find(".gb1")
		fmt.Println(result)
	}
}