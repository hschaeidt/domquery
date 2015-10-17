package searchutil_test

import (
  "testing"
  "golang.org/x/net/html"
  "strings"
  "github.com/hschaeidt/domquery/searchutil"
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

func TestResult(t *testing.T) {
  result := new(searchutil.Result)
  chain := getTokenChain()
  result.Add(chain)

  if result.First() != chain {
    t.Error("Expected: First result to equals chain - Got: 2 different chains")
  }

  chain2 := getTokenChain()
  result.Add(chain2)

  if len(result.All()) != 2 {
    t.Errorf("Expected: Length of result objects to be 2 - Got: %s", len(result.All()))
  }
}
