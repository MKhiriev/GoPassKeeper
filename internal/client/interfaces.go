package client

// Client defines the minimal lifecycle contract for runnable client
// applications.
type Client interface {
	// Run starts the client application and blocks until exit.
	Run() error
}
