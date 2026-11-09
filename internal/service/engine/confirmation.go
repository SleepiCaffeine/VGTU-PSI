package engine

import (
	"fmt"
	"strings"

	"github.com/SleepiCaffeine/VGTU-PSI/internal/model"
)

// ClickPhrase is the acknowledgement shown for click confirmations.
const ClickPhrase = "Taip, tai nuostolingas statymas"

// minConfirmationWaitSeconds keeps every confirmation from being accepted instantly.
const minConfirmationWaitSeconds = 1

// LossPhrase states the average cost of a stake in Lithuanian number format.
func LossPhrase(stake int, houseEdge float64) string {
	loss := strings.Replace(fmt.Sprintf("%.2f", float64(stake)*houseEdge), ".", ",", 1)
	return fmt.Sprintf("Šis statymas yra nuostolingas: vidutiniškai prarasiu %s taško.", loss)
}

func requiredConfirmation(in BetInput, level, stake int) *model.ConfirmationRequest {
	regularly := in.Session.BetCount%in.Settings.ConfirmEveryBets == 0
	click := &model.ConfirmationRequest{Kind: model.ConfirmationClick, Phrase: ClickPhrase, WaitSeconds: minConfirmationWaitSeconds}
	text := &model.ConfirmationRequest{Kind: model.ConfirmationText, Phrase: LossPhrase(stake, in.HouseEdge), WaitSeconds: minConfirmationWaitSeconds}

	switch {
	case level >= 4:
		return text
	case level == 3 && regularly:
		text.WaitSeconds = max(in.Settings.FrictionWaitSeconds, minConfirmationWaitSeconds)
		return text
	case level == 2 && regularly:
		return click
	case level == 1 && in.Session.BetCount == 0:
		return click
	default:
		return nil
	}
}

func confirmationFailure(attempt ConfirmationAttempt, in BetInput, stake int, request *model.ConfirmationRequest) model.Reason {
	c := attempt.Record
	if c == nil || c.UsedAt != nil || c.SessionID != in.Session.ID || c.Stake != stake || c.Kind != request.Kind {
		return model.ReasonConfirmationInvalid
	}
	if in.Now.Before(c.NotBefore) {
		return model.ReasonConfirmationTooEarly
	}
	if c.Kind == model.ConfirmationText && strings.TrimSpace(attempt.Text) != c.Phrase {
		return model.ReasonConfirmationMismatch
	}
	return ""
}
