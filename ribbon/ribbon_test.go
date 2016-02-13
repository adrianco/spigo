package ribbon

import (
	"fmt"
	"github.com/adrianco/spigo/gotocol"
	"github.com/adrianco/spigo/names"
	"testing"
	"time"
)

func TestRouter(t *testing.T) {
	r := MakeRouter()
	c := make(chan gotocol.Message)
	n := names.Make("test", "us", "a", "add", "staash", 0)
	now := time.Now()

	r.Add(n, c, now)
	if r.Random() != c {
		t.Errorf("Random failed to get the right channel back for %v", n)
	}
	if r.Named(n) != c {
		t.Errorf("Name failed to get the right channel back for %v", n)
	}
	if r.Pick("staash") != c {
		t.Errorf("Pick failed to get the right channel back for %v", n)
	}
	if r.Pick("junk") != nil {
		t.Errorf("Pick failed to return nil for junk")
	}
	if r.NameChan(c) != n {
		t.Errorf("NameChan failed to return %v for chan", n)
	}

	r.Remove(n)
	if r.Pick("staash") != nil {
		t.Errorf("Pick failed to return nil after removing staash")
	}
	for i := 0; i < 5; i++ {
		r.Add(names.Make("test", "us", "a", "j", "junk", i), nil, now)
	}
	for i := 0; i < 5; i++ {
		r.Add(names.Make("test", "us", "a", "s", "staash", i), c, now)
	}
	fmt.Println(r.Len(), r)
	for i := 0; i < 10; i++ {
		if r.Pick("staash") != c {
			t.Errorf("Pick failed to get the right channel back for %v", n)
		}
		if r.Pick("junk") != nil {
			t.Errorf("Pick failed to return nil for junk")
		}
		fmt.Printf("Random found staash: %v\n", c == r.Random())
	}
}
