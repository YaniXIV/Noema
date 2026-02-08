package session

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
)

const CookieName = "noema_judge"

// Sign produces a signed cookie value for the given payload (e.g. judge key).
// Format: base64(payload) + "." + hex(HMAC-SHA256(secret, payload)).
func Sign(secret, payload string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	sig := hex.EncodeToString(mac.Sum(nil))
	encoded := base64.StdEncoding.EncodeToString([]byte(payload))
	return encoded + "." + sig
}

// Verify checks the signed cookie and returns the payload if valid.
func Verify(secret, signed string) (payload string, ok bool) {
	idx := strings.LastIndex(signed, ".")
	if idx == -1 {
		return "", false
	}
	encoded, sigHex := signed[:idx], signed[idx+1:]
	sig, err := hex.DecodeString(sigHex)
	if err != nil || len(sig) != sha256.Size {
		return "", false
	}
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", false
	}
	payload = string(raw)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	expected := mac.Sum(nil)
	if !hmac.Equal(sig, expected) {
		return "", false
	}
	return payload, true
}

// CookieOptions holds options for setting the session cookie.
type CookieOptions struct {
	Secure bool
}

// SetCookieHeader returns a Set-Cookie header value for the given signed value.
// Caller should set HttpOnly, SameSite=Lax, Path=/.
func SetCookieHeader(signedValue string, opts CookieOptions) string {
	secure := ""
	if opts.Secure {
		secure = "; Secure"
	}
	return fmt.Sprintf("%s=%s; Path=/; HttpOnly; SameSite=Lax; Max-Age=86400%s",
		CookieName, signedValue, secure)
}
