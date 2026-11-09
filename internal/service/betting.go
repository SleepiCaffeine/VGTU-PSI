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
	loaded, err := loadBetContext(ctx, repos, s.deps.Clock, tenantID, sessionID, request)
	if err != nil {
		return model.BetResult{}, err
	}

	decision := s.deps.Engine.DecideBet(loaded.engineInput(request, g.HouseEdge()))

	if decision.Outcome == model.OutcomeConfirmationRequired {
		id, err := createConfirmation(ctx, repos, loaded.session, decision, loaded.now)
		if err != nil {
			return model.BetResult{}, err
		}
		decision = decision.WithConfirmationID(id)
	}

	plan, session := loaded.plan, loaded.session
	var round *model.RoundResult
	if decision.Accepted() {
		err = consumeConfirmation(ctx, repos, loaded.confirmation, loaded.now)
		if err != nil {
			return model.BetResult{}, err
		}
		played := g.Play(decision.Stake, request.Selection, s.deps.Random)
		plan, session = engine.SettleRound(plan, session, loaded.settings, decision.Stake, played.Payout, loaded.now)
		err = saveRound(ctx, repos, plan, newBet(session, g, request, decision, played, loaded.now))
		if err != nil {
			return model.BetResult{}, err
		}
		round = &model.RoundResult{Outcome: played.Outcome, Payout: played.Payout}
	}

	session, err = persistEffects(ctx, repos, session, decision.Verdict.Effects, loaded.now)
	if err != nil {
		return model.BetResult{}, err
	}
	return model.BetResult{Decision: decision, Round: round, Balance: plan.Balance, Session: session}, nil
}

// betContext is the stored state a bet is decided against.
type betContext struct {
	now          time.Time
	plan         model.Plan
	settings     model.RuleSettings
	session      model.Session
	week         model.WeekActivity
	confirmation *model.Confirmation
}

func loadBetContext(
	ctx context.Context, repos Repositories, clock Clock, tenantID, sessionID int64, request model.BetRequest,
) (betContext, error) {
	plan, now, err := currentPlan(ctx, repos, tenantID, clock)
	if err != nil {
		return betContext{}, err
	}
	settings, err := repos.Settings.Get(ctx, tenantID)
	if err != nil {
		return betContext{}, fmt.Errorf("load settings: %w", err)
	}
	session, err := repos.Sessions.Get(ctx, tenantID, sessionID)
	if err != nil {
		return betContext{}, fmt.Errorf("load session: %w", err)
	}
	week, err := repos.History.WeekActivity(ctx, tenantID, plan.WeekStart, engine.NextWeekStart(plan.WeekStart))
	if err != nil {
		return betContext{}, fmt.Errorf("load week activity: %w", err)
	}
	confirmation, err := submittedConfirmation(ctx, repos, tenantID, request)
	if err != nil {
		return betContext{}, err
	}
	return betContext{
		now: now, plan: plan, settings: settings, session: session, week: week, confirmation: confirmation,
	}, nil
}

func (c betContext) engineInput(request model.BetRequest, houseEdge float64) engine.BetInput {
	input := engine.BetInput{
		Now:            c.now,
		Settings:       c.settings,
		Plan:           c.plan,
		Session:        c.session,
		Week:           c.week,
		RequestedStake: request.Stake,
		HouseEdge:      houseEdge,
	}
	if request.ConfirmationID != nil {
		input.Attempt = &engine.ConfirmationAttempt{Record: c.confirmation, Text: request.ConfirmationText}
	}
	return input
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

func createConfirmation(ctx context.Context, repos Repositories, session model.Session, decision model.BetDecision, now time.Time) (int64, error) {
	confirmation := model.Confirmation{
		TenantID:  session.TenantID,
		SessionID: session.ID,
		Stake:     decision.Stake,
		Kind:      decision.Confirmation.Kind,
		Phrase:    decision.Confirmation.Phrase,
		NotBefore: now.Add(time.Duration(decision.Confirmation.WaitSeconds) * time.Second),
		CreatedAt: now,
	}
	err := repos.Confirmations.Create(ctx, &confirmation)
	if err != nil {
		return 0, fmt.Errorf("create confirmation: %w", err)
	}
	return confirmation.ID, nil
}

func consumeConfirmation(ctx context.Context, repos Repositories, confirmation *model.Confirmation, now time.Time) error {
	if confirmation == nil {
		return nil
	}
	used := *confirmation
	used.UsedAt = &now
	err := repos.Confirmations.Save(ctx, used)
	if err != nil {
		return fmt.Errorf("consume confirmation: %w", err)
	}
	return nil
}

func newBet(session model.Session, g game.Game, request model.BetRequest, decision model.BetDecision, played game.Result, now time.Time) model.Bet {
	return model.Bet{
		TenantID:       session.TenantID,
		SessionID:      session.ID,
		Game:           g.Name(),
		Selection:      request.Selection,
		RequestedStake: request.Stake,
		Stake:          decision.Stake,
		Payout:         played.Payout,
		Outcome:        played.Outcome,
		CreatedAt:      now,
	}
}

func saveRound(ctx context.Context, repos Repositories, plan model.Plan, bet model.Bet) error {
	err := repos.Plans.Save(ctx, plan)
	if err != nil {
		return fmt.Errorf("save plan: %w", err)
	}
	err = repos.Bets.Create(ctx, &bet)
	if err != nil {
		return fmt.Errorf("record bet: %w", err)
	}
	return nil
}

func persistEffects(ctx context.Context, repos Repositories, session model.Session, effects model.Effects, now time.Time) (model.Session, error) {
	updated := engine.ApplyEffects(session, effects, now)
	err := repos.Sessions.Save(ctx, updated)
	if err != nil {
		return model.Session{}, fmt.Errorf("save session: %w", err)
	}
	if effects.AlertReason == "" {
		return updated, nil
	}
	err = repos.Alerts.Create(ctx, &model.Alert{
		TenantID: updated.TenantID, SessionID: updated.ID, Reason: effects.AlertReason, CreatedAt: now,
	})
	if err != nil {
		return model.Session{}, fmt.Errorf("create alert: %w", err)
	}
	return updated, nil
}
