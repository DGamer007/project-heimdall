package services

import (
	"errors"
	"fmt"
	"time"

	"heimdall/backend/internal/config"
	"heimdall/backend/internal/domain/entities"
	internal_errors "heimdall/backend/pkg/errors"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTService interface {
	Generate(sub string) (entities.Token, error)
	VerifyAndDecode(tokenString string) (entities.TokenClaims, error)
	Decode(token string) (entities.TokenClaims, error)
}

type jwtService struct {
	Secret string
	TTL    time.Duration
	Issuer string
}

type JWTServiceConfig struct {
	Secret string
	TTL    time.Duration
	Issuer string
}

func NewJWTService(config JWTServiceConfig) *jwtService {
	return &jwtService{
		Secret: config.Secret,
		TTL:    config.TTL,
		Issuer: config.Issuer,
	}
}

func NewAccessTokenJWTService() *jwtService {
	return NewJWTService(JWTServiceConfig{
		Secret: config.AppConfig.Auth.JWT.AccessToken.Secret,
		TTL:    config.AppConfig.Auth.JWT.AccessToken.TTL,
		Issuer: config.AppConfig.Auth.JWT.Issuer,
	})
}

func NewRefreshTokenJWTService() *jwtService {
	return NewJWTService(JWTServiceConfig{
		Secret: config.AppConfig.Auth.JWT.RefreshToken.Secret,
		TTL:    config.AppConfig.Auth.JWT.RefreshToken.TTL,
		Issuer: config.AppConfig.Auth.JWT.Issuer,
	})
}

func (s *jwtService) Generate(sub string) (entities.Token, error) {
	if sub == "" {
		return entities.Token{}, internal_errors.NewValidationError(
			"Subject is required for token generation",
			map[string]any{"field": "sub"},
		)
	}

	claims := s.newTokenClaims(sub)

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss": claims.Issuer,
		"sub": claims.Subject,
		"iat": claims.IssuedAt,
		"exp": claims.ExpiresAt,
		"jti": claims.JTI,
	}).SignedString([]byte(s.Secret))
	if err != nil {
		return entities.Token{}, internal_errors.NewInfrastructureError(
			"Failed to sign token",
			err,
		)
	}

	return entities.Token{
		SignedToken: token,
		Claims:      *claims,
	}, nil
}

func (s *jwtService) VerifyAndDecode(tokenString string) (entities.TokenClaims, error) {
	if tokenString == "" {
		return entities.TokenClaims{}, internal_errors.NewValidationError(
			"token is required for validation",
			map[string]any{"field": "token"},
		)
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, internal_errors.NewAuthenticationError(
				fmt.Sprintf("Unexpected signing method: %v", token.Header["alg"]),
			)
		}
		return []byte(s.Secret), nil
	})

	if err != nil {
		switch {
		case errors.Is(err, jwt.ErrTokenMalformed):
			return entities.TokenClaims{}, internal_errors.NewAuthenticationError("token is malformed")
		case errors.Is(err, jwt.ErrTokenExpired):
			return entities.TokenClaims{}, internal_errors.NewAuthenticationError("token has expired")
		case errors.Is(err, jwt.ErrTokenNotValidYet):
			return entities.TokenClaims{}, internal_errors.NewAuthenticationError("token is not valid yet")
		case errors.Is(err, jwt.ErrTokenSignatureInvalid):
			return entities.TokenClaims{}, internal_errors.NewAuthenticationError("token signature is invalid")
		default:
			return entities.TokenClaims{}, internal_errors.NewInfrastructureError("Failed to parse token", err)
		}
	}

	return s.extractAndValidateClaims(token)
}

func (s *jwtService) Decode(tokenString string) (entities.TokenClaims, error) {
	if tokenString == "" {
		return entities.TokenClaims{}, internal_errors.NewValidationError(
			"token is required for validation",
			map[string]any{"field": "token"},
		)
	}

	token, _, err := jwt.NewParser().ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return entities.TokenClaims{}, internal_errors.NewInfrastructureError("Failed to parse token", err)
	}

	return s.extractAndValidateClaims(token)
}

func (s *jwtService) newTokenClaims(sub string) *entities.TokenClaims {
	jti := uuid.New().String()

	now := time.Now()
	return &entities.TokenClaims{
		Issuer:    config.AppConfig.Auth.JWT.Issuer,
		Subject:   sub,
		JTI:       jti,
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(s.TTL).Unix(),
	}
}

func (s *jwtService) extractAndValidateClaims(token *jwt.Token) (entities.TokenClaims, error) {
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		iss, issOk := claims["iss"].(string)
		sub, subOk := claims["sub"].(string)
		jti, jtiOk := claims["jti"].(string)
		iat, iatOk := claims["iat"].(float64)
		exp, expOk := claims["exp"].(float64)

		if !issOk || !subOk || !iatOk || !expOk || !jtiOk {
			return entities.TokenClaims{}, internal_errors.NewValidationError(
				"token is missing required claims (iss, sub, iat, exp)",
				map[string]any{
					"resource":       "token",
					"missing_claims": "iss, sub, iat, exp",
				},
			)
		}

		tokenClaims := entities.TokenClaims{
			Issuer:    iss,
			Subject:   sub,
			JTI:       jti,
			IssuedAt:  int64(iat),
			ExpiresAt: int64(exp),
		}

		if tokenClaims.Issuer != s.Issuer {
			return entities.TokenClaims{}, internal_errors.NewAuthenticationError("token issuer is invalid")
		}

		return tokenClaims, nil
	}

	return entities.TokenClaims{}, internal_errors.NewAuthenticationError("token claims are invalid")
}
