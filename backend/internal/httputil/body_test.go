package httputil

import (
	"errors"
	"net/http"
	"testing"
)

func TestIsBodyTooLarge(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil", err: nil, want: false},
		{name: "max bytes error", err: &http.MaxBytesError{Limit: 1}, want: true},
		{name: "http too large string", err: errors.New("http: request body too large"), want: true},
		{name: "multipart too large string", err: errors.New("multipart: message too large"), want: true},
		{name: "other error", err: errors.New("boom"), want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsBodyTooLarge(tc.err); got != tc.want {
				t.Fatalf("IsBodyTooLarge(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}
