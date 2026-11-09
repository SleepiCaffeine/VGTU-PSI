package game

// Default builds the registry of available games.
func Default() *Registry {
	return NewRegistry(Roulette{})
}
