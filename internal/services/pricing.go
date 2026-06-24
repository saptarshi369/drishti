// Package services holds Drishti's derived computations: pricing, the usage
// rollup fold, and the Overview snapshot assembler. Each is a small, pure,
// independently testable unit.
package services

// rate is a model's price in US dollars per MILLION tokens, by token class.
type rate struct {
	inPerM, outPerM, cacheReadPerM, cacheWritePerM float64
}

// pricing is a tiny local table (zero network, zero deps). Extend as needed;
// unknown models cost 0 so the UI degrades to token-only rather than lying.
var pricing = map[string]rate{
	"claude-opus-4-8":   {inPerM: 15, outPerM: 75, cacheReadPerM: 1.5, cacheWritePerM: 18.75},
	"claude-sonnet-4-6": {inPerM: 3, outPerM: 15, cacheReadPerM: 0.3, cacheWritePerM: 3.75},
	"claude-haiku-4-5":  {inPerM: 1, outPerM: 5, cacheReadPerM: 0.1, cacheWritePerM: 1.25},
}

// Cost estimates the USD cost of a token bundle for model. Returns 0 for an
// unknown model. Inputs are raw token counts; rates are per million.
func Cost(model string, in, out, cacheRead, cacheWrite int64) float64 {
	r, ok := pricing[model]
	if !ok {
		return 0
	}
	const m = 1_000_000.0
	return float64(in)/m*r.inPerM +
		float64(out)/m*r.outPerM +
		float64(cacheRead)/m*r.cacheReadPerM +
		float64(cacheWrite)/m*r.cacheWritePerM
}
