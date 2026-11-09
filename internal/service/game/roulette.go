package game

import (
	"fmt"
	"strconv"
)

const (
	rouletteName    = "roulette"
	roulettePockets = 37
	colorPayout     = 2
	numberPayout    = 36
	selectionRed    = "red"
	selectionBlack  = "black"
)

var redPockets = map[int]bool{
	1: true, 3: true, 5: true, 7: true, 9: true, 12: true, 14: true, 16: true, 18: true,
	19: true, 21: true, 23: true, 25: true, 27: true, 30: true, 32: true, 34: true, 36: true,
}

// Roulette is European single-zero roulette with colour and straight-number bets.
type Roulette struct{}

func (Roulette) Name() string { return rouletteName }

func (Roulette) HouseEdge() float64 { return 1.0 / roulettePockets }

func (Roulette) ValidateSelection(selection string) error {
	if selection == selectionRed || selection == selectionBlack {
		return nil
	}
	number, err := strconv.Atoi(selection)
	if err != nil || number < 0 || number >= roulettePockets {
		return fmt.Errorf("%w: %q", ErrInvalidSelection, selection)
	}
	return nil
}

func (Roulette) Play(stake int, selection string, random RandomSource) Result {
	pocket := random.IntN(roulettePockets)
	return Result{
		Outcome: strconv.Itoa(pocket),
		Payout:  stake * payoutMultiplier(selection, pocket),
	}
}

func payoutMultiplier(selection string, pocket int) int {
	switch selection {
	case selectionRed:
		return colorMultiplier(redPockets[pocket])
	case selectionBlack:
		return colorMultiplier(pocket != 0 && !redPockets[pocket])
	case strconv.Itoa(pocket):
		return numberPayout
	default:
		return 0
	}
}

func colorMultiplier(won bool) int {
	if won {
		return colorPayout
	}
	return 0
}
