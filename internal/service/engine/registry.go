package engine

// Default builds the engine with the registered rules, in evaluation order.
func Default() *Engine {
	return New(
		[]Rule{
			noOpenSessionRule{},
			noCreditsRule{},
			allowedHoursRule{},
			weeklySessionLimitRule{},
			cooldownRule{},
		},
		[]Rule{
			sessionActiveRule{},
			forcedBreakRule{},
			chasingRule{},
			stakeLimitsRule{},
			frictionRule{},
		},
	)
}
