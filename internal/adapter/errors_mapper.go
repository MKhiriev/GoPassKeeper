package adapter

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-resty/resty/v2"
)

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
