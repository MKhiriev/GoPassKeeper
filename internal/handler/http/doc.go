// Package http implements the HTTP transport layer of the application.
//
// It exposes route wiring, request handlers, and middleware used by the REST
// API. Cross-cutting concerns such as authentication, request tracing, access
// logging, response compression, and integrity checks are handled in this
// package before requests are delegated to the service layer.
package http
