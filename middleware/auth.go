package middleware

import (
	"asii/utils"
	"log"
	"net/http"
	"strings"
)

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.Printf("Missing Authorization header")
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 {
			log.Printf("Invalid token format")
			http.Error(w, "Invalid token format", http.StatusUnauthorized)
			return
		}

		claims, err := utils.ValidateToken(bearerToken[1])
		if err != nil {
			log.Printf("Token validation failed: %v", err)
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		log.Printf("User authenticated: ID=%d, Role=%s", claims.UserID, claims.Role)
		ctx := utils.ContextWithUser(r.Context(), claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func AdminOnly(next http.HandlerFunc) http.HandlerFunc {
	return AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		claims := utils.UserFromContext(r.Context())
		if claims.Role != "admin" {
			http.Error(w, "Admin access required", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func TeacherOrAdmin(next http.HandlerFunc) http.HandlerFunc {
	return AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		claims := utils.UserFromContext(r.Context())
		log.Printf("TeacherOrAdmin middleware: User role is %s", claims.Role)
		if claims.Role != "teacher" && claims.Role != "admin" {
			log.Printf("Access denied: User %d with role %s tried to access teacher/admin endpoint",
				claims.UserID, claims.Role)
			http.Error(w, "Teacher or admin access required", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
