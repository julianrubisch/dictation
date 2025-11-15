package main

import (
	"math/rand"
	"time"
)

// shuffleWords shuffles a slice of words using Fisher-Yates algorithm
// This function takes a slice (Go's dynamic array type) and returns
// a new shuffled slice without modifying the original.
func shuffleWords(words []string) []string {
	// make() creates a slice with the specified length
	// We copy the original to avoid mutating it
	shuffled := make([]string, len(words))
	copy(shuffled, words)

	// Create a new random number generator seeded with current time
	// This ensures different shuffles each run
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	
	// Fisher-Yates shuffle: iterate backwards, swap each element
	// with a random element from the unshuffled portion
	for i := len(shuffled) - 1; i > 0; i-- {
		j := r.Intn(i + 1)  // Random index from 0 to i
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]  // Swap
	}

	return shuffled
}
