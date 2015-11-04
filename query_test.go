package domquery_test

import (
	"strings"
	"testing"

	"github.com/hschaeidt/domquery"
)

func TestFind(t *testing.T) {
	dom := `<html><head></head><body><div class="myClass"><a class="myClass">myLink</a></div><a class="myClass">otherLink</a><a class="otherClass">yetAnotherLink</a></body></html>`
	q := new(domquery.Query)
	q.Load(strings.NewReader(dom))

	result := q.Find("{class}myClass")
	res := result.All()

	if len(res) != 3 {
		t.Errorf("Expected: Result length of 3 - Got: Result length of %d", len(res))
	}

	res1 := res[0]
	res2 := res[1]
	res3 := res[2]

	value, chain := res1.Value()

	if value != "" {
		t.Errorf("Expected: No value - Got: %s", value)
	}

	if chain == nil {
		t.Error("Expected: Next chain - Got: No next chain")
	}

	value, chain = res2.Value()

	if value != "myLink" {
		t.Errorf("Expected: myLink as value - Got: %s as value", value)
	}

	if chain != nil {
		t.Error("Expected: No next chain - Got: Next chain")
	}

	value, chain = res3.Value()

	if value != "otherLink" {
		t.Errorf("Expected: otherLink as value - Got: %s as value", value)
	}

	if chain != nil {
		t.Error("Expected: No next chain - Got: Next chain")
	}
}

func TestFindFindMultipleClasses(t *testing.T) {
	dom := `<html><head></head><body><div class="myClass myClass2"><a class="myClass">myLink</a></div><a class="myClass">otherLink</a><a class="otherClass">yetAnotherLink</a></body></html>`
	q := new(domquery.Query)
	q.Load(strings.NewReader(dom))

	result := q.Find("{class}myClass{class}myClass2")
	res := result.All()

	if len(res) != 1 {
		t.Errorf("Expected: Result length of 1 - Got: Result length of %d", len(res))
	}

	res1 := res[0]

	value, _ := res1.Value()

	if value != "" {
		t.Errorf("Expected: No value - Got: %s", value)
	}
}
