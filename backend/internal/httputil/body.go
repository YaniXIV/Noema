package httputil

import (
	"errors"
	"net/http"
	"strings"
)

// IsBodyTooLarge reports whether the error indicates the request body exceeded MaxBytesReader.
func IsBodyTooLarge(err error) bool {
	if err == nil {
		return false
	}
	var maxBytesErr *http.MaxBytesError
	if errors.As(err, &maxBytesErr) {
		return true
	}
	return strings.Contains(err.Error(), "request body too large")
}
