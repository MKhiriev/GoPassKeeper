package utils

import (
	"github.com/go-resty/resty/v2"
)

// HTTPClient is a wrapper around the resty.Client HTTP client.
// It embeds *resty.Client to expose all of its methods directly,
// while allowing extension with additional application-specific behavior.
//
// Example usage:
//
//	client := utils.NewHTTPClient()
//	resp, err := client.R().Get("https://example.com")
type HTTPClient struct {
	*resty.Client
}

// NewHTTPClient creates and returns a new HTTPClient instance
// with a default-configured underlying resty.Client.
//
// Each call returns an independent client instance with its own
// configuration, connection pool, and state.
//
// Returns:
//
//	*HTTPClient - a ready-to-use HTTP client
//
// Example usage:
//
//	client := utils.NewHTTPClient()
//	resp, err := client.R().
//	    SetHeader("Accept", "application/json").
//	    Get("https://api.example.com/users")
func NewHTTPClient() *HTTPClient {
	return &HTTPClient{Client: resty.New()}
}
