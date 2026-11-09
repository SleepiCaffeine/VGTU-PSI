package engine

// Default builds the engine with the registered rules, in evaluation order.
func Default() *Engine {
	return New(
		[]Rule{
			noOpenSessionRule{},
			noCreditsRule{},
			chasingLockoutRule{},
			breakActiveRule{},
			allowedHoursRule{},
			weeklySessionLimitRule{},
			cooldownRule{},
		},
		[]Rule{
			sessionActiveRule{},
			forcedBreakRule{},
			rapidBettingRule{},
			chasingRule{},
			stakeLimitsRule{},
			frictionRule{},
		},
	)
}
