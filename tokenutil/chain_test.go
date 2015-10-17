package tokenutil_test

import (
  "testing"
  "strings"
  "golang.org/x/net/html"
  "github.com/hschaeidt/domquery/tokenutil"
)

// Builds a clean prepared token chain for all test methods
func getTokenChain() *tokenutil.Chain {
  // example DOM
  dom := `<div class="myClass1"><b>myValue1</b></div><span class="myClass2">myValue2</span>`
  // tokenizer takes a io.Reader instance as arg
  tokenizer := html.NewTokenizer(strings.NewReader(dom));
  // simulate one next, as the chain expects a "used" tokenizer
  tokenizer.Next()
  // building a new chain from our tokenizer
  chain := tokenutil.NewChainFromTokenizer(tokenizer)

  return chain
}

// Testing if the StartToken matchs the one we expect from our DOM data
func TestStartToken(t *testing.T) {
  chain := getTokenChain()
  startToken := chain.StartToken()

  if startToken.Type != html.StartTagToken {
    t.Errorf("Expected: StartTagToken - Got: %s", startToken.Type)
  }

  if startToken.Data != "div" {
    t.Errorf("Expected: div - Got: %s", startToken.Data)
  }

  correctElement := false
  for _, attr := range startToken.Attr {
    if attr.Key == "class" && attr.Val == "myClass1" {
			correctElement = true
		}
  }

  if !correctElement {
    t.Errorf("Expected StartToken to have the class: myClass1")
  }
}

// Testing if the EndToken data matchs with the StartToken data
func TestEndToken(t *testing.T) {
  chain := getTokenChain()
  startToken := chain.StartToken()
  endToken := chain.EndToken()

  if endToken.Type != html.EndTagToken {
    t.Errorf("Expected: EndTagToken - Got: %s", endToken.Type)
  }

  if endToken.Data != startToken.Data {
    t.Errorf("Expected EndToken for: %s - Got: %s", startToken.Data, endToken.Data)
  }
}

// Testing that TextToken is empty
func TestTextToken(t *testing.T) {
  chain := getTokenChain()
  textToken := chain.TextToken()

  if textToken.Data != "" {
    t.Errorf("Expected: No TextToken - Got TextToken data: %s", textToken.Data)
  }
}

func TestNextAndPrev(t *testing.T) {
  chain := getTokenChain()

  if chain.Next() == nil {
    t.Error("Expected: Next chain - Got: nil")
  }

  if chain.Prev() != nil {
    t.Error("Expected: No prev chain - Got: Prev chain")
  }
}

func TestRootChain(t *testing.T)  {
  chain := getTokenChain()

  if chain.Next().GetRootChain() != chain {
    t.Errorf("Chain not equals root chain")
  }
}
