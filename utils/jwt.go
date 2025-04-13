package utils

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// GenerateToken creates a JWT token for a user
func GenerateToken(userID string, role string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

// GenerateTokenPair creates both access and refresh tokens
func GenerateTokenPair(userID string, organizationID string, role string) (string, string, error) {
	// Access Token
	accessTokenClaims := jwt.MapClaims{
		"user_id":         userID,
		"organization_id": organizationID,
		"role":            role,
		"type":            "access",
		"exp":             time.Now().Add(time.Hour * 1).Unix(), // Short-lived access token
		"iat":             time.Now().Unix(),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)
	accessTokenString, err := accessToken.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", "", err
	}

	// Refresh Token
	refreshTokenClaims := jwt.MapClaims{
		"user_id": userID,
		"type":    "refresh",
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(), // Longer-lived refresh token
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}

// VerifyRefreshToken validates a refresh token
func VerifyRefreshToken(tokenString string) (*jwt.Token, jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		return nil, nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, nil, fmt.Errorf("invalid refresh token")
	}

	// Additional check to ensure it's a refresh token
	if claims["type"] != "refresh" {
		return nil, nil, fmt.Errorf("not a refresh token")
	}

	return token, claims, nil
}
