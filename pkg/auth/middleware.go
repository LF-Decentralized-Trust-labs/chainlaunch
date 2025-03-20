package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"os"
	"strings"
)

const (
	SessionCookieName = "session_id"
)

// getEncryptionKey returns the session encryption key from environment variable
func getEncryptionKey() []byte {
	key := os.Getenv("SESSION_ENCRYPTION_KEY")
	if key == "" {
		// Fallback to a default key if not set (not recommended for production)
		key = "default-encryption-key-12345"
	}
	return []byte(key)
}

// signSessionID creates a signed version of the session ID
func signSessionID(sessionID string) string {
	mac := hmac.New(sha256.New, getEncryptionKey())
	mac.Write([]byte(sessionID))
	signature := mac.Sum(nil)
	return base64.URLEncoding.EncodeToString(signature)
}

// verifySessionID verifies the signature of a signed session ID
func verifySessionID(sessionID, signature string) bool {
	expectedSignature := signSessionID(sessionID)
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// parseBasicAuth extracts credentials from Basic Auth header
func parseBasicAuth(r *http.Request) (username, password string, ok bool) {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Basic ") {
		return "", "", false
	}

	payload, err := base64.StdEncoding.DecodeString(auth[6:])
	if err != nil {
		return "", "", false
	}

	pair := strings.SplitN(string(payload), ":", 2)
	if len(pair) != 2 {
		return "", "", false
	}

	return pair[0], pair[1], true
}

// AuthMiddleware creates a new authentication middleware
func AuthMiddleware(authService *AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			// Skip authentication for login endpoint and swagger docs
			if r.URL.Path == "/api/auth/login" ||
				strings.HasPrefix(r.URL.Path, "/swagger/") {
				next.ServeHTTP(w, r)
				return
			}

			var session *Session
			var err error

			// Try Basic Auth first
			if username, password, ok := parseBasicAuth(r); ok {
				session, err = authService.Login(ctx, username, password)
				if err != nil {
					http.Error(w, "Invalid credentials", http.StatusUnauthorized)
					return
				}
			} else {
				// Try Cookie Auth
				cookie, err := r.Cookie(SessionCookieName)
				if err != nil {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}

				// Split cookie value into session ID and signature
				parts := strings.Split(cookie.Value, ".")
				if len(parts) != 2 {
					http.Error(w, "Invalid session format", http.StatusUnauthorized)
					return
				}

				sessionID, signature := parts[0], parts[1]

				// Verify signature
				if !verifySessionID(sessionID, signature) {
					http.Error(w, "Invalid session signature", http.StatusUnauthorized)
					return
				}

				// Validate session
				session, err = authService.ValidateSession(ctx, sessionID)
				if err != nil {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
			}

			// Add session to request context
			ctx = WithSession(ctx, session)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}
