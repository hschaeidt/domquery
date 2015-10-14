package tokenutil

import (
	"golang.org/x/net/html"
)

type Chain struct {
	depth int
	collection []html.Token
}

// Adds manually a token to our chain
func (this *Chain) Add(token html.Token) bool {
	end := false
	
	tokenType := token.Type

	// just avoid errors
	if tokenType == html.ErrorToken {
		end = true
	}

	// we're digging one step deeper
	if tokenType == html.StartTagToken {
		this.depth++
	}

	// and one step out
	if tokenType == html.EndTagToken {
		this.depth--
	}

	// push new item to our chain
	this.collection = append(this.collection, token)

	// by verifiying against smaller than zero we ensure that the loop
	// makes one more turn to get the rootElements EndTagToken too
	// a correct loop will always end in minus one
	//
	// TODO: this may cause errors by encoutering self closing tags later on
	if this.depth == 0 {
		end = true
	}
	
	return end
}

// Gets the token chain
func (this *Chain) Get() ([]html.Token, bool) {
	var err = false
	
	if this.collection == nil {
		err = true
	}
	
	return this.collection, err
}
