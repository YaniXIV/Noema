package session

import (
	"encoding/hex"
	"strings"
	"testing"
)

func TestSignVerifyRoundTrip(t *testing.T) {
	secret := "s3cr3t"
	payload := "judge-key-123"

	signed := Sign(secret, payload)
	got, ok := Verify(secret, signed)
	if !ok {
		t.Fatalf("expected verify ok")
	}
	if got != payload {
		t.Fatalf("expected payload %q, got %q", payload, got)
	}
}

func TestVerifyRejectsTampering(t *testing.T) {
	secret := "secret"
	payload := "abc"

	signed := Sign(secret, payload)
	if !strings.Contains(signed, ".") {
		t.Fatalf("expected signed value to contain dot")
	}
	// Flip the last hex character.
	last := signed[len(signed)-1]
	replacement := byte('0')
	if last == '0' {
		replacement = '1'
	}
	tampered := signed[:len(signed)-1] + string(replacement)

	_, ok := Verify(secret, tampered)
	if ok {
		t.Fatalf("expected tampered signature to fail")
	}
}

func TestVerifyRejectsInvalidFormat(t *testing.T) {
	secret := "secret"

	cases := []string{
		"",                   // empty
		"no-dot",             // missing dot
		"abc.def",            // non-hex signature
		"YWJj." + hex.EncodeToString([]byte("short")), // wrong length
	}

	for _, tc := range cases {
		_, ok := Verify(secret, tc)
		if ok {
			t.Fatalf("expected verify to fail for %q", tc)
		}
	}
}
