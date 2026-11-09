package model

// BetRequest is what the player asks to wager.
type BetRequest struct {
	Game             string
	Selection        string
	Stake            int
	ConfirmationID   *int64
	ConfirmationText string
}
