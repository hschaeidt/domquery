package domquery

import (
	"io"
	"regexp"
	"strings"

	"github.com/hschaeidt/domquery/searchutil"
	"github.com/hschaeidt/domquery/tokenutil"
	"golang.org/x/net/html"
)

// Represents a DOM-Query
// Also represents recursively the whole query-chain
type Query struct {
	tokenizer *html.Tokenizer //Contains the tokenized HTML DOM

	hasPrevQuery bool
	hasNextQuery bool // Has next query?
	prevQuery    *Query
	nextQuery    *Query // Next query object

	match  map[string][]string //from the mapper some value(s)
	result *searchutil.Result  // Contains token results from the matches, based on these the nextQuery will be executed
}

// Processing the search-term then launching the Document or Token search
func (q *Query) Find(term string) *searchutil.Result {
	if q.result == nil {
		q.result = new(searchutil.Result)
	}

	q.ProcessSearchTerm(term, nil)

	return q.Search()
}

// Takes the decision weither to use RootSearch (DOM) or TokenSearch (List of elements)
// This method also takes care to return the results of the last (sub-)query
func (q *Query) Search() *searchutil.Result {
	var result *searchutil.Result

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
func (q *Query) RootSearch() *searchutil.Result {
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
			tokenChain := tokenutil.NewChainFromTokenizer(q.tokenizer)
			q.result.Add(tokenChain)
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
func (q *Query) TokenSearch(tokenChain *tokenutil.Chain) *searchutil.Result {
	var success bool

	for {
		if tokenChain.Next() == nil {
			return q.result
		}

		// we start with the next sub-chain, as the one passed to this func as arg counts already as match
		tokenChain = tokenChain.Next()

		success = q.Match(tokenChain.StartToken())

		if success == true {
			q.result.Add(tokenChain)
			// search within the new chain again this will be done recursively upon the deepest level of the chain
			// new results will have their own new chain
			// TODO: upon here this can actually be done in coroutines as we are working with totally independant data
			q.TokenSearch(tokenChain)
		}
	}
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
func (q *Query) HasAttr(token html.Token, attrType string, searchValue []string) bool {
	iterations := 0
	results := 0
	for _, attr := range token.Attr {
		for _, val := range searchValue {
			if attr.Key == attrType && strings.Contains(attr.Val, val) {
				results++
			}
			iterations++
		}
	}
	if results == 0 && iterations == 0 {
		return false
	}

	return results == iterations
}

// This function is used to create a new TokenChain from a re-sliced slice
// of an existing TokenChain
//
// Actually it can be used to build a chain from any slice of html.Token
func (q *Query) GetTokenChain(tokenChain []html.Token) *tokenutil.Chain {
	var (
		tChain *tokenutil.Chain
		end    bool
	)

	tChain = new(tokenutil.Chain)

	for _, token := range tokenChain {
		// we reached the end of our chain
		if end {
			break
		}

		tChain, end = tChain.Add(token, nil)
	}

	return tChain
}

// Splits the searchterm in a consecutive chain of search queries using search-maps
func (q *Query) ProcessSearchTerm(term string, parent *Query) {
	var (
		queries  []string
		subQuery *Query
	)

	// Only split into 2 args, because the next query has to handle its
	// own subqueries by itself (recursion)
	queries = strings.SplitN(term, " ", 2)

	q.CreateSearchMap(queries[0])

	if parent != nil {
		q.hasPrevQuery = true
		q.prevQuery = parent
	}

	// we got subselects
	if len(queries) > 1 {
		subQuery = new(Query)
		subQuery.hasPrevQuery = true
		subQuery.prevQuery = q

		// this will chain the recursively for each consecutive sub-query
		subQuery.ProcessSearchTerm(queries[1], q)
	}
}

func (q *Query) CreateSearchMap(query string) {
	if q.match == nil {
		q.match = make(map[string][]string)
	}

	reg := regexp.MustCompile("({.*?})([^{]*)")
	matches := reg.FindAllStringSubmatch(query, -1)
	for _, match := range matches {
		for i := 1; i+2 <= len(match); i += 2 {
			index := strings.TrimPrefix(match[i], "{")
			index = strings.TrimSuffix(index, "}")
			// This adds 2 or more elements from the same type in the same query
			// for example "{class}class1{class}class2"
			q.match[index] = append(q.match[index], match[i+1])
		}
	}
}

// Loads the reader's input into tokenized HTML.
// It can be used to iterate through, finding / changing values.
func (q *Query) Load(reader io.Reader) {
	q.tokenizer = html.NewTokenizer(reader)
}
