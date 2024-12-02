package server

import (
	"net/http"
	"strings"
)

type Middleware struct {
	keycloakClient IKeyCloak
}

func NewMiddleware(keycloakClient IKeyCloak) *Middleware {
	return &Middleware{keycloakClient}
}

func (m *Middleware) ValidateToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header is missing", http.StatusUnauthorized)
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			http.Error(w, "Invalid token format", http.StatusUnauthorized)
			return
		}

		token := tokenParts[1]

		if _, ok := m.isValidToken(token); !ok {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (m *Middleware) IsAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header is missing", http.StatusUnauthorized)
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			http.Error(w, "Invalid token format", http.StatusUnauthorized)
			return
		}

		token := tokenParts[1]

		res, ok := m.isValidToken(token)
		if !ok {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		for _, role := range res.RealmsAccess.Roles {
			if role == "admin" {
				next.ServeHTTP(w, r)
				return
			}
		}

		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}

func (m *Middleware) isValidToken(token string) (*UserInfoResponse, bool) {
	res, err := m.keycloakClient.IntrospectToken(token)
	if err != nil {
		return res, false
	}
	return res, true
}
