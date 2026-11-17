package platform

import (
	"crypto/subtle"
	"net/http"
)

// RequireUser answers 401 unless the gateway identified a signed-in user.
func RequireUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := IdentityFrom(r).User(); !ok {
			WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Sign in to continue.")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireRole answers 401 for an anonymous caller and 403 for a caller without one of the roles.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := IdentityFrom(r)
			if _, ok := id.User(); !ok {
				WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Sign in to continue.")
				return
			}
			if !id.HasRole(roles...) {
				WriteError(w, http.StatusForbidden, "ACCESS_DENIED", "You do not have permission to access this resource.")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// InternalOnly guards /internal/** endpoints: the caller must send the shared INTERNAL_API_TOKEN. An empty token
// configures nothing and refuses every call, so a forgotten variable fails closed.
func InternalOnly(token string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			given := r.Header.Get(HeaderInternalToken)
			if token == "" || subtle.ConstantTimeCompare([]byte(given), []byte(token)) != 1 {
				WriteError(w, http.StatusForbidden, "ACCESS_DENIED", "internal endpoint")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
