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
	lambda := 1800.0
	rho := []float64{0.0, 0.05, 0.00, 0.05}  // 0.05
	kappa := []float64{0.0, 0.0, 0.02, 0.02} // 0.02
	for i := 0; i < len(rho); i++ {
		fmt.Printf("BaseLambda: %v ContentionRho: %v CrosstalkKappa: %v MaxThroughput:%.2f\n", lambda, rho[i], kappa[i], ThroughputMax(rho[i], kappa[i]))
		for n := 0.0; n <= 20.0; n += 1.0 {
			// ThroughputX(capacityN, baseLambda, contentionRho, crosstalkKappa real)
			fmt.Printf("N:%v\tX(N):%.2f\tX(R(N)):%.2f\tR(N):%.5f\tR(X(N)):%.5f\n", n, ThroughputXN(n, lambda, rho[i], kappa[i]),
				ThroughputXR(ResponseRN(n, lambda, rho[i], kappa[i]), lambda, rho[i], kappa[i]),
				ResponseRN(n, lambda, rho[i], kappa[i]),
				ResponseRX(ThroughputXN(n, lambda, rho[i], kappa[i]), lambda, rho[i], kappa[i]))
		}
	}
}
