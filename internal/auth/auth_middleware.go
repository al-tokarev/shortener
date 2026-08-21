package auth

import (
	"context"
	"net/http"
)

type contextKey string

const userIDKey contextKey = "userID"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, err := GetUserIDFromCookie(r)

		if err != nil {
			if isCookiePresent(r) && !isCookieValid(r) {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			userID = GenerateUserID()
			cookie := CreateSignedCookie(userID)
			http.SetCookie(w, cookie)
		}

		ctx := context.WithValue(r.Context(), userIDKey, userID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey).(string)
	return userID, ok
}

func isCookiePresent(r *http.Request) bool {
	_, err := r.Cookie("user_id")
	return err == nil
}

func isCookieValid(r *http.Request) bool {
	_, err := GetUserIDFromCookie(r)
	return err == nil
}
