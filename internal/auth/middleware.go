package auth

//import (
//	"myapp/internal/models"
//	"net/http"
//)
//
//func WithRoleMiddleware(next http.Handler) http.Handler {
//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		user, ok := r.Context().Value("user").(models.User)
//		if !ok {
//			http.Error(w, "Unauthorized", http.StatusUnauthorized)
//			return
//		}
//
//		if user.Role == "user" {
//			http.Error(w, "Forbidden: only non-user roles allowed", http.StatusForbidden)
//			return
//		}
//
//		next.ServeHTTP(w, r)
//	})
//}
