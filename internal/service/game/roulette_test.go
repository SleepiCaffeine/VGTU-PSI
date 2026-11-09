package game_test

import (
	"errors"
	"math"
	"testing"

	"github.com/SleepiCaffeine/VGTU-PSI/internal/service/game"
)

type fixedRandom int

func (f fixedRandom) IntN(int) int { return int(f) }

func TestRoulettePlay(t *testing.T) {
	tests := []struct {
		name       string
		selection  string
		pocket     int
		wantPayout int
	}{
		{"red wins on red pocket", "red", 1, 40},
		{"red loses on black pocket", "red", 2, 0},
		{"black wins on black pocket", "black", 2, 40},
		{"zero loses for red", "red", 0, 0},
		{"zero loses for black", "black", 0, 0},
		{"straight number wins", "17", 17, 720},
		{"straight number loses", "17", 18, 0},
		{"zero can be bet directly", "0", 0, 720},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := game.Roulette{}.Play(20, tt.selection, fixedRandom(tt.pocket))
			if got.Payout != tt.wantPayout {
				t.Errorf("payout = %d, want %d", got.Payout, tt.wantPayout)
			}
		})
	}
}

func TestRouletteValidateSelection(t *testing.T) {
	tests := []struct {
		selection string
		wantErr   error
	}{
		{"red", nil},
		{"black", nil},
		{"0", nil},
		{"36", nil},
		{"37", game.ErrInvalidSelection},
		{"-1", game.ErrInvalidSelection},
		{"green", game.ErrInvalidSelection},
	}

	for _, tt := range tests {
		t.Run(tt.selection, func(t *testing.T) {
			err := game.Roulette{}.ValidateSelection(tt.selection)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("err = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestRouletteExpectedReturnMatchesHouseEdge(t *testing.T) {
	roulette := game.Roulette{}
	for _, selection := range []string{"red", "black", "17"} {
		totalPayout := 0
		for pocket := range 37 {
			totalPayout += roulette.Play(1, selection, fixedRandom(pocket)).Payout
		}
		edge := 1 - float64(totalPayout)/37
		if math.Abs(edge-roulette.HouseEdge()) > 1e-9 {
			t.Errorf("%s edge = %f, want %f", selection, edge, roulette.HouseEdge())
		}
	}
}

func TestRegistryLookup(t *testing.T) {
	registry := game.Default()

	_, err := registry.Lookup("roulette")
	if err != nil {
		t.Fatalf("roulette not registered: %v", err)
	}

	_, err = registry.Lookup("keno")
	if !errors.Is(err, game.ErrUnknownGame) {
		t.Errorf("err = %v, want ErrUnknownGame", err)
	}
}
