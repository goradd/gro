package test

import "math/rand"

var rng *rand.Rand

// NewSeed returns a new random seed for the random number generator.
func NewSeed() int64 {
	seed := int64(rand.Intn(1000000))
	UseSeed(seed)
	return seed
}

// UseSeed uses the given seed to recreate the results from the random number generator.
func UseSeed(seed int64) {
	source := rand.NewSource(seed)
	rng = rand.New(source)
}
