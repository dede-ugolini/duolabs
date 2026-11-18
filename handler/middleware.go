package handler

import (
	"context"
	"net/http"
	"strings"

	"duolabs/model"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
	userIDKey contextKey = "user_id"
	roleKey   contextKey = "role"
)

// jwtSecret signs and verifies the tokens. The default is only meant for
// development: main replaces it with $JWT_SECRET through SetJWTSecret.
var jwtSecret = []byte("duolabs-secret")

// SetJWTSecret replaces the signing key. Call it once at startup, before
// serving requests.
func SetJWTSecret(secret string) {
	jwtSecret = []byte(secret)
}

// claimsFromRequest validates the Bearer token (HS256 only) and returns its
// claims.
func claimsFromRequest(r *http.Request) (jwt.MapClaims, bool) {
	header := r.Header.Get("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		return nil, false
	}
	token, err := jwt.Parse(strings.TrimPrefix(header, "Bearer "), func(t *jwt.Token) (any, error) {
		return jwtSecret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil || !token.Valid {
		return nil, false
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	return claims, ok
}

// IsAdminRequest reports whether the request carries a valid admin token. It
// never rejects: use it to vary a response, not to protect a route.
func IsAdminRequest(r *http.Request) bool {
	claims, ok := claimsFromRequest(r)
	if !ok {
		return false
	}
	role, _ := claims["role"].(string)
	return role == model.RoleAdmin
}

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := claimsFromRequest(r)
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		id, ok := claims["user_id"].(float64)
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		role, _ := claims["role"].(string)
		ctx := context.WithValue(r.Context(), userIDKey, uint(id))
		ctx = context.WithValue(ctx, roleKey, role)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// AdminMiddleware requires a valid token (401 otherwise) whose role is admin
// (403 otherwise).
func AdminMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if RoleFromContext(r) != model.RoleAdmin {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func RoleFromContext(r *http.Request) string {
	role, _ := r.Context().Value(roleKey).(string)
	return role
}

func UserIDFromContext(r *http.Request) uint {
	id, _ := r.Context().Value(userIDKey).(uint)
	return id
}

// currentUser returns the id of the authenticated user. Without one it
// answers 401 and reports false.
func currentUser(w http.ResponseWriter, r *http.Request) (uint, bool) {
	id := UserIDFromContext(r)
	if id == 0 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return 0, false
	}
	return id, true
}
