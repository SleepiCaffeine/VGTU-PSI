// Package game holds the games playable with virtual points.
package game

import (
	"errors"
	"fmt"
)

var (
	ErrUnknownGame      = errors.New("unknown game")
	ErrInvalidSelection = errors.New("invalid selection")
)

// RandomSource returns a uniform integer in [0, n).
type RandomSource interface {
	IntN(n int) int
}

// Result is the outcome of one round.
type Result struct {
	Outcome string
	Payout  int
}

// Game resolves rounds on the server.
type Game interface {
	Name() string
	HouseEdge() float64
	ValidateSelection(selection string) error
	Play(stake int, selection string, random RandomSource) Result
}

// Registry finds games by name.
type Registry struct {
	games map[string]Game
}

func NewRegistry(games ...Game) *Registry {
	byName := make(map[string]Game, len(games))
	for _, g := range games {
		byName[g.Name()] = g
	}
	return &Registry{games: byName}
}

func (r *Registry) Lookup(name string) (Game, error) {
	g, found := r.games[name]
	if !found {
		return nil, fmt.Errorf("%w: %s", ErrUnknownGame, name)
	}
	return g, nil
}
