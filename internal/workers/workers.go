package workers

// Workers is an aggregate that holds a collection of Worker instances
// and allows running all of them together via a single Run call.
type Workers struct {
	// workers is the list of Worker instances managed by this aggregate.
	workers []Worker
}

// Run starts all registered workers sequentially by calling Run on each one.
//
// Workers are executed in the order they were added.
// If a worker blocks indefinitely, subsequent workers will not be started
// until it returns. Consider wrapping long-running workers in goroutines
// within their own Run implementations if concurrent execution is needed.
//
// Example usage:
//
//	w := &Workers{workers: []Worker{worker1, worker2}}
//	w.Run()
func (w *Workers) Run() {
	for _, worker := range w.workers {
		worker.Run()
	}
}
