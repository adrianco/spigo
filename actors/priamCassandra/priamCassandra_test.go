package priamCassandra

import (
	"fmt"
	"github.com/adrianco/spigo/tooling/gotocol"
	"testing"
)

func TestPriamCassandra(t *testing.T) {
	x := 16
	cass := make(map[string]chan gotocol.Message, x)
	for i := 0; i < x; i++ {
		cass[fmt.Sprintf("cass%v", i)] = nil
	}
	s := Distribute(cass)
	fmt.Println(s)
	r := RingConfig(s)
	c := make(map[int]int, x)
	for i := 0; i < 1000; i++ {
		w := fmt.Sprintf("whynot%v%v", i, i*i)
		h := ringHash(w)
		f := r.Find(h)
		c[f]++
		// fmt.Printf("%v %v %v\n", w, f, h)
	}
	fmt.Println(c)
}
