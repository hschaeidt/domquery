package searchutil

import(
	"github.com/hschaeidt/domquery/tokenutil"
)

// Result represents a util to manipulate result data from Query
type Result struct {
	result []*tokenutil.Chain
}

// Adds token-chain to result
func (this *Result) Add(chain *tokenutil.Chain) {
	this.result = append(this.result, chain)
}

// Returns only the first token-chain out of all matches (results)
func (this *Result) First() *tokenutil.Chain {
	if (len(this.result) > 0) {
		return this.result[0]
	}
	
	return nil
}

func (this *Result) All() []*tokenutil.Chain {
	return this.result
}