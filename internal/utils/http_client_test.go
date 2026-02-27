package utils

import (
	"testing"

	"github.com/go-resty/resty/v2"
)

func TestNewHTTPClient_NotNil(t *testing.T) {
	client := NewHTTPClient()

	if client == nil {
		t.Fatal("expected non-nil *HTTPClient, got nil")
	}

	if client.Client == nil {
		t.Fatal("expected embedded *resty.Client to be non-nil, got nil")
	}
}

func TestNewHTTPClient_Type(t *testing.T) {
	client := NewHTTPClient()

	// Ensure the embedded client is actually a *resty.Client
	if _, ok := interface{}(client.Client).(*resty.Client); !ok {
		t.Fatalf("expected embedded client to be *resty.Client, got %T", client.Client)
	}
}

func TestNewHTTPClient_Independence(t *testing.T) {
	// Create two clients and make sure they don't share the same underlying resty.Client
	client1 := NewHTTPClient()
	client2 := NewHTTPClient()

	if client1.Client == client2.Client {
		t.Fatal("expected NewHTTPClient to return HTTPClients with different *resty.Client instances")
	}
}

func TestHTTPClient_EmbeddedClientUsable(t *testing.T) {
	client := NewHTTPClient()

	// Just check that we can call a basic method on the embedded resty client
	req := client.R()
	if req == nil {
		t.Fatal("expected non-nil request from embedded resty client")
	}
}
