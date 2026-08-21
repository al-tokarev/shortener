package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"
)

const (
	cookieName   = "user_id"
	cookieMaxAge = 60 * 60 * 24 * 30
)

var secretKey = []byte("Sfiuewnben3243dsQWiejfvvw323j!")

func GenerateUserID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func SignData(data string) string {
	h := hmac.New(sha256.New, secretKey)
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func VerifySignature(data, signature string) bool {
	expected := SignData(data)
	return hmac.Equal([]byte(expected), []byte(signature))
}

func CreateSignedCookie(userID string) *http.Cookie {
	signature := SignData(userID)
	value := userID + "|" + signature

	return &http.Cookie{
		Name:     cookieName,
		Value:    value,
		Path:     "/",
		MaxAge:   cookieMaxAge,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}
}

func GetUserIDFromCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return "", fmt.Errorf("cookie not found: %w", err)
	}

	parts := splitCookieValue(cookie.Value)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid cookie format")
	}

	userID, signature := parts[0], parts[1]
	if !VerifySignature(userID, signature) {
		return "", fmt.Errorf("invalid signature")
	}

	return userID, nil
}

func splitCookieValue(value string) []string {
	for i := 0; i < len(value); i++ {
		if value[i] == '|' {
			return []string{value[:i], value[i+1:]}
		}
	}
	return nil
}
