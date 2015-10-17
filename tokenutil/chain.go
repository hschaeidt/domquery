package tokenutil

import (
	"golang.org/x/net/html"
)

type Chain struct {
	startToken html.Token
	textToken html.Token
	next *Chain
	prev *Chain
	endToken html.Token
	depth int
}

// Adds recursively tokens to the chain whereas for each StartTagToken a new
// sub-chain will be created and linked to the current chain.
//
// This gives more control to the manipulation of the completed chain
func (this *Chain) Add(token html.Token, prev *Chain) (*Chain, bool) {
	chain := this
	end := false
	tokenType := token.Type

	// This is only true by default if called by recursion within this func
	if prev != nil {
		this.prev = prev
	}

	// 2 StartTokens in a row encountered, creating new sub-chain linked to this chain
	if tokenType == html.StartTagToken && this.depth > 0 {
		chain = new(Chain)
		this.next = chain
		return chain.Add(token, this)
	}

	// Setting chains basic values
	switch tokenType {
	case html.StartTagToken:
		this.startToken = token
		this.depth++
	case html.EndTagToken:
		this.endToken = token
		this.depth--
		if this.prev != nil {
			return this.prev, false
		}
	case html.TextToken:
		this.textToken = token
	}

	// We are back at the root level
	// This is true whenever a EntToken has been encountered
	if this.depth == 0 {
		end = true
	}

	return chain, end
}

// Collects a chains value
// Always returns 2 values, a string (Text-value / field-value etc) and a chain ref
// The chain ref will be the pointer to the next chain within the current chain
func (this *Chain) Value() (string, *Chain) {
	var (
		value string
		chain *Chain
	)

	value = this.textToken.Data

	if this.next != nil {
		chain = this.next
	}

	return value, chain
}

func (this *Chain) Next() *Chain {
	return this.next
}

func (this *Chain) Prev() *Chain {
	return this.prev
}

func (this *Chain) StartToken() html.Token {
	return this.startToken
}

func (this *Chain) TextToken() html.Token {
	return this.textToken
}

func (this *Chain) EndToken() html.Token {
	return this.endToken
}

func (this *Chain) GetRootChain() *Chain {
	chain := this

	for {
		if chain.prev == nil {
			return chain
		}

		chain = chain.prev
	}
}

// Creates a new Chain object out of a tokenizer pointer
func NewChainFromTokenizer(tokenizer *html.Tokenizer) *Chain {
	// Creating a new token chain
	var (
		rootToken html.Token
		tokenChain *Chain
		end bool
	)

	rootToken = tokenizer.Token()
	tokenChain = new(Chain)

	if rootToken.Type != html.StartTagToken {
		return nil
	}

	// First of all we add the rootToken to our chain
	tokenChain, end = tokenChain.Add(rootToken, nil)

	for {
		// we reached the end of our chain
		if end {
			break;
		}

		tokenizer.Next()

		// Adding next token
		tokenChain, end = tokenChain.Add(tokenizer.Token(), nil)
	}

	// Now we got our chain, the requester has to make sure to search through the token chain for
	// eventual other (inner-)matches
	return tokenChain.GetRootChain()
}
