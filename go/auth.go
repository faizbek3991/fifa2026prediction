package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type contextKey string

const usernameContextKey contextKey = "username"

func contextWithUsername(ctx context.Context, username string) context.Context {
	return context.WithValue(ctx, usernameContextKey, username)
}

func usernameFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(usernameContextKey).(string); ok {
		return v
	}
	return ""
}

const (
	sessionCookieName = "session_token"
	sessionTTL        = 24 * time.Hour
)

func hashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(b), err
}

func checkPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func newSessionToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// isSecureRequest reports whether the connection is over TLS (directly or via
// a trusted reverse proxy setting X-Forwarded-Proto), used to decide whether
// to mark the session cookie Secure. On plain HTTP (local dev) Secure would
// stop the browser from ever sending the cookie back.
func isSecureRequest(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	return r.Header.Get("X-Forwarded-Proto") == "https"
}

func setSessionCookie(w http.ResponseWriter, r *http.Request, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   isSecureRequest(r),
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(sessionTTL),
	})
}

func clearSessionCookie(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   isSecureRequest(r),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

// requireAuth gates a handler behind a valid session cookie. API requests
// (as identified by the Accept header) get a 401 JSON body; page requests
// get redirected to the login page.
func requireAuth(db *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookieName)
		if err != nil {
			redirectToLogin(w, r)
			return
		}
		username, err := getSessionUsername(db, cookie.Value)
		if err != nil {
			redirectToLogin(w, r)
			return
		}
		ctx := contextWithUsername(r.Context(), username)
		next(w, r.WithContext(ctx))
	}
}

func redirectToLogin(w http.ResponseWriter, r *http.Request) {
	if isAPIRequest(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func isAPIRequest(r *http.Request) bool {
	return len(r.URL.Path) >= 5 && r.URL.Path[:5] == "/api/"
}

// --- Simple in-memory login rate limiter (per IP) ---

type rateLimiter struct {
	mu       sync.Mutex
	attempts map[string][]time.Time
	max      int
	window   time.Duration
}

func newRateLimiter(max int, window time.Duration) *rateLimiter {
	return &rateLimiter{attempts: make(map[string][]time.Time), max: max, window: window}
}

func (rl *rateLimiter) allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)
	kept := rl.attempts[key][:0]
	for _, t := range rl.attempts[key] {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	if len(kept) >= rl.max {
		rl.attempts[key] = kept
		return false
	}
	rl.attempts[key] = append(kept, now)
	return true
}

func clientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	return r.RemoteAddr
}

func logAuthEvent(msg, username, ip string, ok bool) {
	if ok {
		slog.Info(msg, slog.String("username", username), slog.String("ip", ip))
	} else {
		slog.Warn(msg, slog.String("username", username), slog.String("ip", ip))
	}
}
