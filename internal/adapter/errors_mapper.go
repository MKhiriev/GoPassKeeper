// SPDX-License-Identifier: Apache-2.0
// Copyright 2026 Rasul Khiriev

package adapter

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-resty/resty/v2"
)

// mapHTTPError converts a resty HTTP response into an error value. It returns
// nil for any 2xx status code. For known error codes it wraps the corresponding
// sentinel (e.g. [ErrConflict] for 409) with the trimmed response body as
// additional context. For unrecognised non-2xx codes it returns a plain
// "http <code>: <body>" error.
func mapHTTPError(resp *resty.Response) error {
	if resp.StatusCode() >= http.StatusOK && resp.StatusCode() < http.StatusMultipleChoices {
		return nil
	}

	body := strings.TrimSpace(string(resp.Body()))

	switch resp.StatusCode() {
	case http.StatusBadRequest:
		return fmt.Errorf("%w: %s", ErrBadRequest, body)
	case http.StatusUnauthorized:
		return fmt.Errorf("%w: %s", ErrUnauthorized, body)
	case http.StatusForbidden:
		return fmt.Errorf("%w: %s", ErrForbidden, body)
	case http.StatusNotFound:
		return fmt.Errorf("%w: %s", ErrNotFound, body)
	case http.StatusConflict:
		return fmt.Errorf("%w: %s", ErrConflict, body)
	case http.StatusBadGateway:
		return fmt.Errorf("%w: %s", ErrBadGateway, body)
	case http.StatusInternalServerError:
		return fmt.Errorf("%w: %s", ErrInternalServerError, body)
	default:
		if body == "" {
			body = http.StatusText(resp.StatusCode())
		}
		return fmt.Errorf("http %d: %s", resp.StatusCode(), body)
	}
}
