package main

import (
	"golang.org/x/net/html"
	"io"
	"net/http"
	"fmt"
	"strings"
	"github.com/hschaeidt/domquery/tokenutil"
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
	result []*tokenutil.Chain // Contains token results from the matches, based on these the nextQuery will be executed
}

// Processing the search-term then launching the Document or Token search
func (q *Query) Find(term string) []*tokenutil.Chain {
	q.ProcessSearchTerm(term)

	return q.Search()
}

// Takes the decision weither to use RootSearch (DOM) or TokenSearch (List of elements)
// This method also takes care to return the results of the last (sub-)query
func (q *Query) Search() []*tokenutil.Chain {
	var result []*tokenutil.Chain

	if q.hasPrevQuery {
		//result = q.TokenSearch(q.prevQuery.result)
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
func (q *Query) RootSearch() []*tokenutil.Chain {
	for {
		// true by default
		success := true
		tokenType := q.tokenizer.Next()

		if tokenType == html.ErrorToken {
			break
		}

		token := q.tokenizer.Token()

		success = q.Match(token)

		if success == true {
			tokenChain := q.GetTokenChainFromTokenizer(token)
			q.result = append(q.result, tokenChain)
			
			// as suggested by GetTokenChainFromTokenizer() we research in the inner of the chain
			// for other matches
			q.TokenSearch(tokenChain)
		}
	}
	
	return q.result
}

// Token search iterates through a TokenChain to find sub-results
// in the already builded chain. This may be useful in case you match the outer DIV
// of the DOM and still want to get deeper smaller results that may also match your
// search-term
func (q *Query) TokenSearch(tokenChain *tokenutil.Chain) []*tokenutil.Chain {
	// In this case the depth of the chain was already only 1, no further searches are required
	tokenList, _ := tokenChain.Get()
	if len(tokenList) <= 3 {
		return q.result
	}
	
	// this loop skips the first and the last element in the token-chain to avoid
	// endless recursive matches
	for i := 1; i < len(tokenList) - 2; i++ {
		token := tokenList[i]
		success := q.Match(token)
		
		if success == true {
			// slicing out the root element (current element matched)
			//                         ,,,,,,,,,,,,
			tChain := q.GetTokenChain(tokenList[i:])
			q.result = append(q.result, tChain)
			
			// search within the new chain again this will be done recursively upon the deepest level of the chain
			// new results will have their own new chain
			// TODO: upon here this can actually be done in coroutines as we are working with totally independant data
			q.TokenSearch(tChain)
		}
	}

	return q.result
}

// Checks for matches from the parsed search-terms for the given Query object
func (q *Query) Match(token html.Token) bool {
	success := true

	for domType, domValue := range q.match {
		switch {
		case token.Type == html.ErrorToken:
			return false
		case token.Type == html.StartTagToken:
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
func (q *Query) GetTokenChainFromTokenizer(rootToken html.Token) *tokenutil.Chain {
	// Creating a new token chain
	var (
		tokenChain *tokenutil.Chain
		end bool
	)
	
	tokenChain = new(tokenutil.Chain)
	
	if rootToken.Type != html.StartTagToken {
		return nil
	}
	
	// First of all we add the rootToken to our chain
	end = tokenChain.Add(rootToken)

	for {
		// we reached the end of our chain
		if end {
			break;
		}
		
		q.tokenizer.Next()
		
		// Adding next token
		end = tokenChain.Add(q.tokenizer.Token())
	}
	
	// Now we got our chain, the requester has to make sure to search through the token chain for
	// eventual other (inner-)matches
	return tokenChain
}

// This function is used to create a new TokenChain from a re-sliced slice
// of an existing TokenChain
//
// Actually it can be used to build a chain from any slice of html.Token
func (q *Query) GetTokenChain(tokenChain []html.Token) *tokenutil.Chain {
	var (
		tChain *tokenutil.Chain
		end bool
	)
	
	tChain = new(tokenutil.Chain)
	
	for _, token := range tokenChain {
		// we reached the end of our chain
		if end {
			break
		}
		
		end = tChain.Add(token)
	}
	
	return tChain
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

	if err == nil {
		
		q := new(Query)
		q.Load(resp.Body)

		result := q.Find(".gb1")
		
		for _, tokenChain := range result {
			fmt.Println(tokenChain.Get())
		}
		
		defer resp.Body.Close()
	}
}
