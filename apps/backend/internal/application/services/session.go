package services

import (
	"heimdall/backend/internal/domain/entities"
	internal_errors "heimdall/backend/pkg/errors"
)

type SessionService struct {
	sessionRepo  entities.SessionRepository
	accessToken  JWTService
	refreshToken JWTService
}

func NewSessionService(sr entities.SessionRepository, ats JWTService, rts JWTService) *SessionService {
	return &SessionService{
		sessionRepo:  sr,
		accessToken:  ats,
		refreshToken: rts,
	}
}

func (s *SessionService) GenerateTokenPair(userId string) (entities.TokenPair, error) {
	accessToken, err := s.accessToken.Generate(userId)
	if err != nil {
		return entities.TokenPair{}, err
	}

	var refreshToken entities.Token
	refreshToken, err = s.refreshToken.Generate(userId)
	if err != nil {
		return entities.TokenPair{}, err
	}

	_, err = s.sessionRepo.CreateOne(&entities.Session{
		UserId:         userId,
		RefreshToken:   refreshToken.SignedToken,
		ExpiresAt:      refreshToken.Claims.ExpiresAt,
		AccessTokenJTI: accessToken.Claims.JTI,
	})
	if err != nil {
		return entities.TokenPair{}, err
	}

	return entities.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *SessionService) ReplaceTokenPair(userId, accessTokenJTI string, refreshToken string) (entities.TokenPair, error) {
	if exists, err := s.sessionRepo.CheckIfExistsByIdentifier(userId, accessTokenJTI, refreshToken); err != nil {
		return entities.TokenPair{}, err
	} else if !exists {
		return entities.TokenPair{}, internal_errors.NewNotFoundError("session", "user_id, access_token_jti, refresh_token")
	}

	at, err := s.accessToken.Generate(userId)
	if err != nil {
		return entities.TokenPair{}, err
	}

	var rt entities.Token
	rt, err = s.refreshToken.Generate(userId)
	if err != nil {
		return entities.TokenPair{}, err
	}

	_, err = s.sessionRepo.ReplaceSession(accessTokenJTI, &entities.Session{
		UserId:         userId,
		AccessTokenJTI: at.Claims.JTI,
		RefreshToken:   rt.SignedToken,
		ExpiresAt:      rt.Claims.ExpiresAt,
	})
	if err != nil {
		return entities.TokenPair{}, err
	}

	return entities.TokenPair{
		AccessToken:  at,
		RefreshToken: rt,
	}, nil
}
