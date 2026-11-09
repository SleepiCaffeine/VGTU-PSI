// Package random provides the random source used to resolve game rounds.
package random

import "math/rand/v2"

type Source struct{}

func (Source) IntN(n int) int {
	return rand.IntN(n)
}
