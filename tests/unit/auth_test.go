package unit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go-gin-template/pkg/auth"
)

func TestJWTManager_GenerateAccessToken(t *testing.T) {
	// Arrange
	jwtManager := auth.NewJWTManager("test-secret", 24, 168)

	// Act
	token, err := jwtManager.GenerateAccessToken(1, "test@example.com", "user")

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestJWTManager_GenerateRefreshToken(t *testing.T) {
	// Arrange
	jwtManager := auth.NewJWTManager("test-secret", 24, 168)

	// Act
	token, err := jwtManager.GenerateRefreshToken(1, "test@example.com", "user")

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestJWTManager_GenerateTokenPair(t *testing.T) {
	// Arrange
	jwtManager := auth.NewJWTManager("test-secret", 24, 168)

	// Act
	accessToken, refreshToken, err := jwtManager.GenerateTokenPair(1, "test@example.com", "user")

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)
	assert.NotEqual(t, accessToken, refreshToken)
}

func TestJWTManager_ValidateAccessToken(t *testing.T) {
	// Arrange
	jwtManager := auth.NewJWTManager("test-secret", 24, 168)
	token, _ := jwtManager.GenerateAccessToken(1, "test@example.com", "user")

	// Act
	claims, err := jwtManager.ValidateAccessToken(token)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, uint(1), claims.UserID)
	assert.Equal(t, "test@example.com", claims.Email)
	assert.Equal(t, "user", claims.Role)
	assert.Equal(t, auth.AccessToken, claims.TokenType)
}

func TestJWTManager_ValidateRefreshToken(t *testing.T) {
	// Arrange
	jwtManager := auth.NewJWTManager("test-secret", 24, 168)
	token, _ := jwtManager.GenerateRefreshToken(1, "test@example.com", "user")

	// Act
	claims, err := jwtManager.ValidateRefreshToken(token)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, uint(1), claims.UserID)
	assert.Equal(t, auth.RefreshToken, claims.TokenType)
}

func TestJWTManager_ValidateAccessToken_WithRefreshToken(t *testing.T) {
	// Arrange - try to validate refresh token as access token
	jwtManager := auth.NewJWTManager("test-secret", 24, 168)
	refreshToken, _ := jwtManager.GenerateRefreshToken(1, "test@example.com", "user")

	// Act
	claims, err := jwtManager.ValidateAccessToken(refreshToken)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Equal(t, auth.ErrInvalidToken, err)
}

func TestJWTManager_ValidateRefreshToken_WithAccessToken(t *testing.T) {
	// Arrange - try to validate access token as refresh token
	jwtManager := auth.NewJWTManager("test-secret", 24, 168)
	accessToken, _ := jwtManager.GenerateAccessToken(1, "test@example.com", "user")

	// Act
	claims, err := jwtManager.ValidateRefreshToken(accessToken)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Equal(t, auth.ErrInvalidToken, err)
}

func TestJWTManager_ValidateToken_InvalidToken(t *testing.T) {
	// Arrange
	jwtManager := auth.NewJWTManager("test-secret", 24, 168)

	// Act
	claims, err := jwtManager.ValidateToken("invalid-token")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestJWTManager_ValidateToken_WrongSecret(t *testing.T) {
	// Arrange
	jwtManager1 := auth.NewJWTManager("secret-1", 24, 168)
	jwtManager2 := auth.NewJWTManager("secret-2", 24, 168)
	token, _ := jwtManager1.GenerateAccessToken(1, "test@example.com", "user")

	// Act
	claims, err := jwtManager2.ValidateAccessToken(token)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestJWTManager_RefreshAccessToken(t *testing.T) {
	// Arrange
	jwtManager := auth.NewJWTManager("test-secret", 24, 168)
	refreshToken, _ := jwtManager.GenerateRefreshToken(1, "test@example.com", "user")

	// Act
	newAccessToken, err := jwtManager.RefreshAccessToken(refreshToken)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, newAccessToken)

	// Verify the new access token is valid
	claims, err := jwtManager.ValidateAccessToken(newAccessToken)
	assert.NoError(t, err)
	assert.Equal(t, uint(1), claims.UserID)
}

func TestJWTManager_GetTokenExpiration(t *testing.T) {
	// Arrange
	jwtManager := auth.NewJWTManager("test-secret", 24, 168)

	// Act
	expiration := jwtManager.GetTokenExpiration()
	refreshExpiration := jwtManager.GetRefreshTokenExpiration()

	// Assert
	assert.Equal(t, 24*time.Hour, expiration)
	assert.Equal(t, 168*time.Hour, refreshExpiration)
}

// Table-driven test example
func TestJWTManager_ValidateToken_TableDriven(t *testing.T) {
	jwtManager := auth.NewJWTManager("test-secret", 24, 168)

	testCases := []struct {
		name      string
		token     string
		expectErr bool
	}{
		{
			name:      "empty token",
			token:     "",
			expectErr: true,
		},
		{
			name:      "invalid format",
			token:     "not.a.valid.jwt.token",
			expectErr: true,
		},
		{
			name:      "random string",
			token:     "randomstring",
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			claims, err := jwtManager.ValidateToken(tc.token)

			if tc.expectErr {
				assert.Error(t, err)
				assert.Nil(t, claims)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
			}
		})
	}
}
