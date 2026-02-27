package server

// Server defines the common lifecycle contract for transport servers managed
// by this package.
//
// Implementations are expected to block in [RunServer] until shutdown is
// requested and to release resources in [Shutdown].
type Server interface {
	// RunServer starts serving requests and blocks until the server stops.
	RunServer()

	// Shutdown gracefully stops the server and frees associated resources.
	Shutdown()
}
