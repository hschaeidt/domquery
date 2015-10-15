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