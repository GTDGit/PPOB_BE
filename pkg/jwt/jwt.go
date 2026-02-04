package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)

// Claims represents JWT claims
type Claims struct {
	UserID   string `json:"userId"`
	DeviceID string `json:"deviceId"`
	Type     string `json:"type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

// TokenPair represents access and refresh tokens
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int // seconds
}

// Generator handles JWT generation and validation
type Generator struct {
	secret       string
	accessTTL    time.Duration
	refreshTTL   time.Duration
}

// NewGenerator creates a new JWT generator
func NewGenerator(secret string, accessTTL, refreshTTL time.Duration) *Generator {
	return &Generator{
		secret:     secret,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

// GenerateTokenPair generates both access and refresh tokens
func (g *Generator) GenerateTokenPair(userID, deviceID string) (*TokenPair, error) {
	accessToken, err := g.generateToken(userID, deviceID, "access", g.accessTTL)
	if err != nil {
		return nil, err
	}

	refreshToken, err := g.generateToken(userID, deviceID, "refresh", g.refreshTTL)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(g.accessTTL.Seconds()),
	}, nil
}

// GenerateAccessToken generates only an access token
func (g *Generator) GenerateAccessToken(userID, deviceID string) (string, error) {
	return g.generateToken(userID, deviceID, "access", g.accessTTL)
}

// GenerateRefreshToken generates only a refresh token
func (g *Generator) GenerateRefreshToken(userID, deviceID string) (string, error) {
	return g.generateToken(userID, deviceID, "refresh", g.refreshTTL)
}

func (g *Generator) generateToken(userID, deviceID, tokenType string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID:   userID,
		DeviceID: deviceID,
		Type:     tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			Issuer:    "ppob.id",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(g.secret))
}

// ValidateToken validates a token and returns its claims
func (g *Generator) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(g.secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// ValidateAccessToken validates an access token
func (g *Generator) ValidateAccessToken(tokenString string) (*Claims, error) {
	claims, err := g.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.Type != "access" {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// ValidateRefreshToken validates a refresh token
func (g *Generator) ValidateRefreshToken(tokenString string) (*Claims, error) {
	claims, err := g.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.Type != "refresh" {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// GetAccessTTL returns the access token TTL
func (g *Generator) GetAccessTTL() time.Duration {
	return g.accessTTL
}

// GetRefreshTTL returns the refresh token TTL
func (g *Generator) GetRefreshTTL() time.Duration {
	return g.refreshTTL
}
