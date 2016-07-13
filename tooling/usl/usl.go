// Package usl implements universal scalability law functions as described in https://github.com/VividCortex/ebooks/blob/master/scalability.pdf
package usl

import (
	"math"
)

// ThroughputXN as a function of capacity
func ThroughputXN(capacityN, baseLambda, contentionSigma, crosstalkKappa float64) float64 {
	// X(N) = LambdaN / (1 + Sigma(N - 1) + KappaN(N - 1)
	// N is the number of nodes or CPUs we want to scale to
	// Lambda is the base throughput for a single node
	// Sigma is the amdahls law contention proportion that doesn't scale
	// Kappa is the crosstalk between nodes
	return (baseLambda * capacityN) / (1.0 + contentionSigma*(capacityN-1.0) + crosstalkKappa*capacityN*(capacityN-1.0))
}

// ThroughputMax throughput for a given contention and crosstalk
func ThroughputMax(contentionSigma, crosstalkKappa float64) float64 {
	// Nmax = sqrt((1-Sigma)/Kappa)
	return math.Sqrt((1.0 - contentionSigma) / crosstalkKappa)
}

// ResponseRN time as a function of capacity
func ResponseRN(capacityN, baseLambda, contentionSigma, crosstalkKappa float64) float64 {
	// R(N) = (1 + Sigma(N - 1) + KappaN(N - 1)) / Lambda
	return (1 + contentionSigma*(capacityN-1.0) + crosstalkKappa*capacityN*(capacityN-1.0)) / baseLambda
}

// ResponseRX time as a function of throughput
func ResponseRX(throughputX, baseLambda, contentionSigma, crosstalkKappa float64) float64 {
	if crosstalkKappa == 0.0 {
		// R(X) = (Sigma - 1) / ( SigmaX - Lambda)
		return (contentionSigma - 1.0) / (contentionSigma*throughputX - baseLambda)
	}
	// R(X) = (+/-sqrt(X^2(Kappa^2 + 2Kappa(Sigma -2) + sigma^2) + 2LambdaX(Kappa - Sigma) + Lambda^2) + KappaX + Lambda - SigmaX) / 2KappaX^2
	sign := -1.0
	if throughputX > ThroughputMax(contentionSigma, crosstalkKappa) {
		sign = 1.0
	}
	return (sign*math.Sqrt(throughputX*throughputX*(crosstalkKappa*crosstalkKappa+
		2.0*crosstalkKappa*(contentionSigma-2.0)+contentionSigma*contentionSigma)+
		2.0*baseLambda*throughputX*(crosstalkKappa-contentionSigma)+
		(baseLambda*baseLambda)) + (crosstalkKappa * throughputX) +
		baseLambda - (contentionSigma * throughputX)) /
		(2.0 * crosstalkKappa * throughputX * throughputX)
}

// ThroughputXR as a function of response time
func ThroughputXR(responseR, baseLambda, contentionSigma, crosstalkKappa float64) float64 {
	// X(R) = (sqrt(Sigma^2 + Kappa^2 + 2Kappa(2LambdaR + Sigma - 2)) - Kappa + Sigma)/2KappaR
	return (math.Sqrt(contentionSigma*contentionSigma+crosstalkKappa*crosstalkKappa+
		2.0*crosstalkKappa*(2.0*baseLambda*responseR+contentionSigma-2.0)) - crosstalkKappa + contentionSigma) /
		(2.0 * crosstalkKappa * responseR)
}
