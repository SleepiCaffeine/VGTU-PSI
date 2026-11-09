package engine

// Default builds the engine with the registered rules, in evaluation order.
func Default() *Engine {
	return New(
		[]SessionRule{
			noOpenSessionRule{},
			noCreditsRule{},
			chasingLockoutRule{},
			breakActiveRule{},
			allowedHoursRule{},
			weeklySessionLimitRule{},
			cooldownRule{},
		},
		[]BetRule{
			sessionActiveRule{},
			forcedBreakRule{},
			rapidBettingRule{},
			chasingRule{},
		},
	)
}
