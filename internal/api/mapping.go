package api

import (
	"github.com/SleepiCaffeine/VGTU-PSI/internal/api/gen"
	"github.com/SleepiCaffeine/VGTU-PSI/internal/model"
)

func toState(state model.State) gen.State {
	response := gen.State{
		InitialAllocation:      state.Plan.InitialAllocation,
		Allocation:             state.Plan.Allocation,
		Balance:                state.Plan.Balance,
		WeekNumber:             state.Plan.WeekNumber,
		ProgressPercent:        state.ProgressPercent,
		FrictionLevel:          state.FrictionLevel,
		ChasingSignalsThisWeek: state.ChasingSignalsThisWeek,
		NextAllocationAt:       state.NextAllocationAt,
	}
	if state.ActiveSession != nil {
		active := toSession(*state.ActiveSession)
		response.ActiveSession = &active
	}
	return response
}

func toSession(session model.Session) gen.Session {
	response := gen.Session{
		Id:             session.ID,
		StartedAt:      session.StartedAt,
		EndedAt:        session.EndedAt,
		NetLoss:        session.NetLoss,
		BetCount:       session.BetCount,
		ChasingSignals: session.ChasingSignals,
		RapidFlags:     session.RapidFlags,
		BreakUntil:     session.BreakUntil,
	}
	if session.EndReason != "" {
		reason := string(session.EndReason)
		response.EndReason = &reason
	}
	return response
}

func toDecision(decision model.Decision) gen.Decision {
	response := gen.Decision{
		Outcome:          string(decision.Outcome),
		Reason:           string(decision.Reason),
		RequestedStake:   decision.RequestedStake,
		Stake:            decision.Stake,
		FrictionLevel:    decision.FrictionLevel,
		NextAllocationAt: decision.NextAllocationAt,
	}
	if decision.Confirmation != nil {
		response.Confirmation = &gen.ConfirmationRequest{
			Id:          decision.Confirmation.ID,
			Kind:        string(decision.Confirmation.Kind),
			Phrase:      decision.Confirmation.Phrase,
			WaitSeconds: decision.Confirmation.WaitSeconds,
		}
	}
	return response
}

func toSessionStartResult(result model.SessionStartResult) gen.SessionStartResult {
	response := gen.SessionStartResult{Decision: toDecision(result.Decision)}
	if result.Session != nil {
		session := toSession(*result.Session)
		response.Session = &session
	}
	return response
}

func toBetResult(result model.BetResult) gen.BetResult {
	response := gen.BetResult{
		Decision: toDecision(result.Decision),
		Balance:  result.Balance,
		Session:  toSession(result.Session),
	}
	if result.Round != nil {
		response.Round = &gen.Round{Outcome: result.Round.Outcome, Payout: result.Round.Payout}
	}
	return response
}
