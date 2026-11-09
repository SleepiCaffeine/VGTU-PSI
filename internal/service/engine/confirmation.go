package engine

import (
	"fmt"
	"strings"

	"github.com/SleepiCaffeine/VGTU-PSI/internal/model"
)

// ClickPhrase is the acknowledgement shown for click confirmations.
const ClickPhrase = "Taip, tai nuostolingas statymas"

// LossPhrase states the average cost of a stake in Lithuanian number format.
func LossPhrase(stake int, houseEdge float64) string {
	loss := strings.Replace(fmt.Sprintf("%.2f", float64(stake)*houseEdge), ".", ",", 1)
	return fmt.Sprintf("Šis statymas yra nuostolingas: vidutiniškai prarasiu %s taško.", loss)
}

type frictionRule struct{}

func (frictionRule) Name() string { return "friction" }

func (frictionRule) Evaluate(in Input, d *model.Decision) bool {
	request := requiredConfirmation(in, d)
	if request == nil {
		return false
	}
	if !in.ConfirmationSubmitted {
		d.Confirmation = request
		return stopForConfirmation(d, model.OutcomeConfirmationRequired, model.ReasonConfirmationNeeded)
	}
	failure := confirmationFailure(in, d.Stake, request)
	if failure != "" {
		return stopForConfirmation(d, model.OutcomeConfirmationFailed, failure)
	}
	d.ConfirmationUsed = true
	return false
}

func requiredConfirmation(in Input, d *model.Decision) *model.ConfirmationRequest {
	regularly := in.Session.BetCount%in.Settings.ConfirmEveryBets == 0
	click := &model.ConfirmationRequest{Kind: model.ConfirmationClick, Phrase: ClickPhrase}
	text := &model.ConfirmationRequest{Kind: model.ConfirmationText, Phrase: LossPhrase(d.Stake, in.HouseEdge)}

	switch {
	case d.FrictionLevel >= 4:
		return text
	case d.FrictionLevel == 3 && regularly:
		text.WaitSeconds = in.Settings.FrictionWaitSeconds
		return text
	case d.FrictionLevel == 2 && regularly:
		return click
	case d.FrictionLevel == 1 && in.Session.BetCount == 0:
		return click
	default:
		return nil
	}
}

func confirmationFailure(in Input, stake int, request *model.ConfirmationRequest) model.Reason {
	c := in.Confirmation
	if c == nil || c.UsedAt != nil || c.SessionID != in.Session.ID || c.Stake != stake || c.Kind != request.Kind {
		return model.ReasonConfirmationInvalid
	}
	if in.Now.Before(c.NotBefore) {
		return model.ReasonConfirmationTooEarly
	}
	if c.Kind == model.ConfirmationText && strings.TrimSpace(in.ConfirmationText) != c.Phrase {
		return model.ReasonConfirmationMismatch
	}
	return ""
}

// stopForConfirmation discards effects so that a retried bet cannot count them twice.
func stopForConfirmation(d *model.Decision, outcome model.Outcome, reason model.Reason) bool {
	d.Outcome = outcome
	d.Reason = reason
	d.Effects = model.Effects{}
	return true
}
