// tests for usl
package usl

import (
	"fmt"
	"testing"
)

func TestUSL(t *testing.T) {
	for k := 1.0; k >= 0.00001; k = k / 2.0 {
		fmt.Printf("Kappa: %.5f Max: %.2f\n", k, ThroughputMax(0.05, k))
	}
	for n := 0.0; n <= 20.0; n += 1.0 {
		// ThroughputX(capacityN, baseLambda, contentionRho, crosstalkKappa real)
		fmt.Printf("N:%v X:%.2f R:%.5f %.5f\n", n, ThroughputXN(n, 1800.0, 0.05, 0.02), ResponseRN(n, 1800.0, 0.05, 0.02),
			ResponseRX(ThroughputXN(n, 1800.0, 0.05, 0.02), 1800.0, 0.05, 0.02) )
	}
}
