package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type ctxKey string

const (
	UserIDKey   ctxKey = "userID"
	UsernameKey ctxKey = "username"
)

func UserIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(UserIDKey).(string)
	return v
}

func UsernameFromContext(ctx context.Context) string {
	v, _ := ctx.Value(UsernameKey).(string)
	return v
}

func parseToken(secret, tokenStr string) (userID, username string, ok bool) {
	parsed, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secret), nil
	})
	if err != nil || !parsed.Valid {
		return "", "", false
	}
	claims, okc := parsed.Claims.(jwt.MapClaims)
	if !okc {
		return "", "", false
	}
	uid, _ := claims["uid"].(string)
	uname, _ := claims["uname"].(string)
	if uid == "" {
		return "", "", false
	}
	return uid, uname, true
}

func bearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if strings.HasPrefix(strings.ToLower(h), "bearer ") {
		return strings.TrimSpace(h[7:])
	}
	return ""
}

func withUser(r *http.Request, userID, username string) *http.Request {
	ctx := context.WithValue(r.Context(), UserIDKey, userID)
	ctx = context.WithValue(ctx, UsernameKey, username)
	return r.WithContext(ctx)
}

func RequireAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := bearerToken(r)
			if token == "" {
				writeAuthError(w, "请先登录")
				return
			}
			uid, uname, ok := parseToken(secret, token)
			if !ok {
				writeAuthError(w, "登录已过期，请重新登录")
				return
			}
			next.ServeHTTP(w, withUser(r, uid, uname))
		})
	}
}

func writeAuthError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"success":false,"message":"` + message + `"}`))
}

func OptionalAuth(secret string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if token := bearerToken(r); token != "" {
			if uid, uname, ok := parseToken(secret, token); ok {
				r = withUser(r, uid, uname)
			}
		}
		next(w, r)
	}
}
