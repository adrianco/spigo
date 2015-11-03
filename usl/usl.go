// package usl implements universal scalability law functions as described in https://github.com/VividCortex/ebooks/blob/master/scalability.pdf
package usl

import (
	"math"
)

// Throughput as a function of capacity
func ThroughputXN(capacityN, baseLambda, contentionRho, crosstalkKappa float64) float64 {
	// X(N) = LambdaN / (1 + Rho(N - 1) + KappaN(N - 1)
	// N is the number of nodes or CPUs we want to scale to
	// Lambda is the base throughput for a single node
	// Rho is the amdahls law contention proportion that doesn't scale
	// Kappa is the crosstalk between nodes
	return (baseLambda * capacityN) / (1.0 + contentionRho*(capacityN-1.0) + crosstalkKappa*capacityN*(capacityN-1.0))
}

func ThroughputMax(contentionRho, crosstalkKappa float64) float64 {
	// Nmax = sqrt((1-Rho)/Kappa)
	return math.Sqrt((1.0 - contentionRho) / crosstalkKappa)
}

func ResponseRN(capacityN, baseLambda, contentionRho, crosstalkKappa float64) float64 {
	// R(N) = (1 + Rho(N - 1) + KappaN(N - 1)) / Lambda
	return (1 + contentionRho*(capacityN-1.0) + crosstalkKappa*capacityN*(capacityN-1.0)) / baseLambda
}

func ResponseRX(throughputX, baseLambda, contentionRho, crosstalkKappa float64) float64 {
	if crosstalkKappa == 0.0 {
		// R(X) = (Rho - 1) / ( RhoX - Lambda)
		return (contentionRho - 1.0) / (contentionRho*throughputX - baseLambda)
	} else {
		// R(X) = (-sqrt(X^2(Kappa^2 + 2Kappa(Rho -2) + rho^2) + 2LambdaX(Kappa - Rho) + Lambda^2) + KappaX + Lambda - RhoX) / 2KappaX^2
		return (-math.Sqrt(throughputX*throughputX*(crosstalkKappa*crosstalkKappa+2.0*crosstalkKappa*(contentionRho-2.0)+contentionRho*contentionRho)+
			2.0*baseLambda*throughputX*(crosstalkKappa-contentionRho)+(baseLambda*baseLambda)) +(crosstalkKappa*throughputX)+baseLambda-(contentionRho*throughputX)) /
			(2.0 * crosstalkKappa * throughputX * throughputX)
	}
}
