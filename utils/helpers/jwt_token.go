package helpers

import (
	"errors"
	"fmt"
	"github.com/gofrs/uuid/v5"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

var (
	ErrInvalidToken      = errors.New("token tidak valid")
	ErrExpiredToken      = errors.New("token telah kadaluarsa")
	ErrTokenNotProvided  = errors.New("token tidak ditemukan")
	ErrInvalidSignMethod = errors.New("metode signing tidak valid")
)

type ClaimsToken struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Role   string    `json:"role"`
	jwt.RegisteredClaims
}

type JWTHelper struct {
	jwtSecret          string
	accessTokenExpiry  time.Duration
	refreshTokenExpiry time.Duration
	issuer             string
}

func NewJWTHelper(jwtSecret string, accessTokenExp int, refreshTokenExp int, issuer string) *JWTHelper {
	return &JWTHelper{
		jwtSecret:          jwtSecret,
		accessTokenExpiry:  time.Duration(accessTokenExp*24) * time.Hour,
		refreshTokenExpiry: time.Duration(refreshTokenExp*24) * time.Hour,
		issuer:             issuer,
	}
}

// GenerateAccessToken membuat token akses baru
func (j *JWTHelper) GenerateAccessToken(userID uuid.UUID, email, role string) (string, time.Time, error) {
	expiryTime := time.Now().Add(j.accessTokenExpiry)

	claims := &ClaimsToken{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiryTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    j.issuer,
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(j.jwtSecret))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("gagal membuat access token: %w", err)
	}

	return signedToken, expiryTime, nil
}

// GenerateRefreshToken membuat token refresh baru
func (j *JWTHelper) GenerateRefreshToken(userID uuid.UUID) (string, time.Time, error) {
	expiryTime := time.Now().Add(j.refreshTokenExpiry)

	claims := &ClaimsToken{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiryTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    j.issuer,
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(j.jwtSecret))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("gagal membuat refresh token: %w", err)
	}

	return signedToken, expiryTime, nil
}

// ValidateAccessToken validasi access token
func (j *JWTHelper) ValidateAccessToken(tokenString string) (*ClaimsToken, error) {
	if tokenString == "" {
		return nil, ErrTokenNotProvided
	}

	token, err := jwt.ParseWithClaims(tokenString, &ClaimsToken{}, func(token *jwt.Token) (interface{}, error) {
		// Validasi metode signing
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidSignMethod
		}
		return []byte(j.jwtSecret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*ClaimsToken)
	if !ok {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

// Ekstrak token claims
func (j *JWTHelper) ExtractTokenClaims(tokenString string) (*ClaimsToken, error) {
	if tokenString == "" {
		return nil, ErrTokenNotProvided
	}

	// Parse token tanpa validasi
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &ClaimsToken{})
	if err != nil {
		return nil, fmt.Errorf("gagal parse token: %w", err)
	}

	// Ekstrak claims
	claims, ok := token.Claims.(*ClaimsToken)
	if !ok {
		return nil, fmt.Errorf("gagal ekstrak claims")
	}

	return claims, nil
}
