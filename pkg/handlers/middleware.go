package handlers

import (
	"net/http"

	"github.com/nexdb/nexdb/pkg/services/auth"
)

type AuthMiddleware struct {
	*auth.AuthService
}

func (a *AuthMiddleware) IsAuthenticated(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key, _, ok := r.BasicAuth()
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if err := a.Authenticate(key); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
