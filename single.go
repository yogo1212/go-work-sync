package work_sync

// synchronizes access to a single resource using channels

type SingleWorkFunc[State any] func(state State)

// Handles requests from c until closed.
// Waits for all workers to finish before returning.
func Single[State any](c <-chan SingleWorkFunc[State], state State) {
	for f := range c {
		f(state)
	}
}

func SpawnSingle[State any](state State) (chan<- SingleWorkFunc[State]) {
	c := make(chan SingleWorkFunc[State])
	go Single(c, state)
	return c
}

func DefaultSingle[State any](c <-chan SingleWorkFunc[State]) {
	var state State

	for f := range c {
		f(state)
	}
}

func SpawnDefaultSingle[State any]() (chan<- SingleWorkFunc[State]) {
	c := make(chan SingleWorkFunc[State])
	go DefaultSingle(c)
	return c
}
