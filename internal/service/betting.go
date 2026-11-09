package service

import (
	"context"
	"fmt"
	"time"

	"github.com/SleepiCaffeine/VGTU-PSI/internal/model"
	"github.com/SleepiCaffeine/VGTU-PSI/internal/service/engine"
	"github.com/SleepiCaffeine/VGTU-PSI/internal/service/game"
)

type BettingService struct {
	deps Dependencies
}

func NewBettingService(deps Dependencies) *BettingService {
	return &BettingService{deps: deps}
}

func (s *BettingService) PlaceBet(ctx context.Context, tenantID, sessionID int64, request model.BetRequest) (model.BetResult, error) {
	g, err := s.deps.Games.Lookup(request.Game)
	if err != nil {
		return model.BetResult{}, err
	}
	err = g.ValidateSelection(request.Selection)
	if err != nil {
		return model.BetResult{}, err
	}

	var result model.BetResult
	err = s.deps.UnitOfWork.Do(ctx, func(ctx context.Context, repos Repositories) error {
		var err error
		result, err = s.placeBet(ctx, repos, tenantID, sessionID, request, g)
		return err
	})
	return result, err
}

func (s *BettingService) placeBet(
	ctx context.Context, repos Repositories, tenantID, sessionID int64, request model.BetRequest, g game.Game,
) (model.BetResult, error) {
	now := s.deps.Clock.Now()
	plan, err := currentPlan(ctx, repos, tenantID, now)
	if err != nil {
		return model.BetResult{}, err
	}
	settings, err := repos.Settings.Get(ctx, tenantID)
	if err != nil {
		return model.BetResult{}, fmt.Errorf("load settings: %w", err)
	}
	session, err := repos.Sessions.Get(ctx, tenantID, sessionID)
	if err != nil {
		return model.BetResult{}, fmt.Errorf("load session: %w", err)
	}
	confirmation, err := submittedConfirmation(ctx, repos, tenantID, request)
	if err != nil {
		return model.BetResult{}, err
	}

	decision := s.deps.Engine.DecideBet(engine.Input{
		Now: now, Settings: settings, Plan: plan, Session: &session,
		RequestedStake: request.Stake, HouseEdge: g.HouseEdge(),
		ConfirmationSubmitted: request.ConfirmationID != nil,
		Confirmation:          confirmation,
		ConfirmationText:      request.ConfirmationText,
	})

	result := model.BetResult{Decision: decision}
	if decision.Outcome == model.OutcomeConfirmationRequired {
		err = createConfirmation(ctx, repos, tenantID, session.ID, &result.Decision, now)
		if err != nil {
			return model.BetResult{}, err
		}
	}
	if decision.Accepted() {
		result.Round, err = s.playRound(ctx, repos, roundContext{
			plan: &plan, session: &session, settings: settings, confirmation: confirmation,
			request: request, decision: decision, game: g, now: now,
		})
		if err != nil {
			return model.BetResult{}, err
		}
	}

	err = persistEffects(ctx, repos, &session, decision, now)
	if err != nil {
		return model.BetResult{}, err
	}
	result.Balance = plan.Balance
	result.Session = session
	return result, nil
}

func submittedConfirmation(ctx context.Context, repos Repositories, tenantID int64, request model.BetRequest) (*model.Confirmation, error) {
	if request.ConfirmationID == nil {
		return nil, nil
	}
	confirmation, err := repos.Confirmations.Find(ctx, tenantID, *request.ConfirmationID)
	if err != nil {
		return nil, fmt.Errorf("load confirmation: %w", err)
	}
	return confirmation, nil
}

func createConfirmation(ctx context.Context, repos Repositories, tenantID, sessionID int64, decision *model.Decision, now time.Time) error {
	confirmation := model.Confirmation{
		TenantID:  tenantID,
		SessionID: sessionID,
		Stake:     decision.Stake,
		Kind:      decision.Confirmation.Kind,
		Phrase:    decision.Confirmation.Phrase,
		NotBefore: now.Add(time.Duration(decision.Confirmation.WaitSeconds) * time.Second),
		CreatedAt: now,
	}
	err := repos.Confirmations.Create(ctx, &confirmation)
	if err != nil {
		return fmt.Errorf("create confirmation: %w", err)
	}
	decision.Confirmation.ID = confirmation.ID
	return nil
}

type roundContext struct {
	plan         *model.Plan
	session      *model.Session
	settings     model.RuleSettings
	confirmation *model.Confirmation
	request      model.BetRequest
	decision     model.Decision
	game         game.Game
	now          time.Time
}

func (s *BettingService) playRound(ctx context.Context, repos Repositories, round roundContext) (*model.RoundResult, error) {
	if round.decision.ConfirmationUsed {
		round.confirmation.UsedAt = &round.now
		err := repos.Confirmations.Save(ctx, *round.confirmation)
		if err != nil {
			return nil, fmt.Errorf("consume confirmation: %w", err)
		}
	}

	played := round.game.Play(round.decision.Stake, round.request.Selection, s.deps.Random)
	engine.SettleRound(round.plan, round.session, round.settings, round.decision.Stake, played.Payout, round.now)

	err := repos.Plans.Save(ctx, *round.plan)
	if err != nil {
		return nil, fmt.Errorf("save plan: %w", err)
	}
	err = repos.Bets.Create(ctx, &model.Bet{
		TenantID:       round.session.TenantID,
		SessionID:      round.session.ID,
		Game:           round.game.Name(),
		Selection:      round.request.Selection,
		RequestedStake: round.decision.RequestedStake,
		Stake:          round.decision.Stake,
		Payout:         played.Payout,
		Outcome:        played.Outcome,
		CreatedAt:      round.now,
	})
	if err != nil {
		return nil, fmt.Errorf("record bet: %w", err)
	}
	return &model.RoundResult{Outcome: played.Outcome, Payout: played.Payout}, nil
}

func persistEffects(ctx context.Context, repos Repositories, session *model.Session, decision model.Decision, now time.Time) error {
	engine.ApplyEffects(session, decision.Effects, now)
	err := repos.Sessions.Save(ctx, *session)
	if err != nil {
		return fmt.Errorf("save session: %w", err)
	}
	if !decision.Effects.NotifySupervisor {
		return nil
	}
	err = repos.Alerts.Create(ctx, &model.Alert{
		TenantID: session.TenantID, SessionID: session.ID, Reason: decision.Reason, CreatedAt: now,
	})
	if err != nil {
		return fmt.Errorf("create alert: %w", err)
	}
	return nil
}
