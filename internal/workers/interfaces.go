// Package workers provides abstractions for managing and running
// background workers in the application.
// It defines the Worker interface and a Workers aggregate that allows
// running multiple workers in a unified way.
package workers

// Worker is the interface that must be implemented by any background worker.
// It defines a single Run method that starts the worker's execution.
//
// Implementations are expected to block for the duration of their work
// or spawn goroutines internally.
//
// Example implementation:
//
//	type MyWorker struct{}
//
//	func (w *MyWorker) Run() {
//	    // start background processing
//	}
type Worker interface {
	Run()
}
