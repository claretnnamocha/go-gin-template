package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrExpiredToken     = errors.New("token has expired")
	ErrInvalidClaims    = errors.New("invalid token claims")
)

// TokenType represents the type of JWT token
type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

// Claims represents JWT claims
type Claims struct {
	UserID    uint      `json:"user_id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	TokenType TokenType `json:"token_type"`
	jwt.RegisteredClaims
}

// JWTManager handles JWT token operations
type JWTManager struct {
	secret                 []byte
	expirationHours        int
	refreshExpirationHours int
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(secret string, expirationHours, refreshExpirationHours int) *JWTManager {
	return &JWTManager{
		secret:                 []byte(secret),
		expirationHours:        expirationHours,
		refreshExpirationHours: refreshExpirationHours,
	}
}

// GenerateAccessToken generates a new access token
func (m *JWTManager) GenerateAccessToken(userID uint, email, role string) (string, error) {
	return m.generateToken(userID, email, role, AccessToken, m.expirationHours)
}

// GenerateRefreshToken generates a new refresh token
func (m *JWTManager) GenerateRefreshToken(userID uint, email, role string) (string, error) {
	return m.generateToken(userID, email, role, RefreshToken, m.refreshExpirationHours)
}

// GenerateTokenPair generates both access and refresh tokens
func (m *JWTManager) GenerateTokenPair(userID uint, email, role string) (accessToken, refreshToken string, err error) {
	accessToken, err = m.GenerateAccessToken(userID, email, role)
	if err != nil {
		return "", "", err
	}

	refreshToken, err = m.GenerateRefreshToken(userID, email, role)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// generateToken generates a JWT token
func (m *JWTManager) generateToken(userID uint, email, role string, tokenType TokenType, expirationHours int) (string, error) {
	now := time.Now()
	expiresAt := now.Add(time.Duration(expirationHours) * time.Hour)

	claims := &Claims{
		UserID:    userID,
		Email:     email,
		Role:      role,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// ValidateToken validates a JWT token and returns the claims
func (m *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return m.secret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidClaims
	}

	return claims, nil
}

// ValidateAccessToken validates an access token
func (m *JWTManager) ValidateAccessToken(tokenString string) (*Claims, error) {
	claims, err := m.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != AccessToken {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// ValidateRefreshToken validates a refresh token
func (m *JWTManager) ValidateRefreshToken(tokenString string) (*Claims, error) {
	claims, err := m.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != RefreshToken {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// RefreshAccessToken generates a new access token from a refresh token
func (m *JWTManager) RefreshAccessToken(refreshTokenString string) (string, error) {
	claims, err := m.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		return "", err
	}

	return m.GenerateAccessToken(claims.UserID, claims.Email, claims.Role)
}

// GetTokenExpiration returns the expiration time for access tokens
func (m *JWTManager) GetTokenExpiration() time.Duration {
	return time.Duration(m.expirationHours) * time.Hour
}

// GetRefreshTokenExpiration returns the expiration time for refresh tokens
func (m *JWTManager) GetRefreshTokenExpiration() time.Duration {
	return time.Duration(m.refreshExpirationHours) * time.Hour
}
