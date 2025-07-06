package services

import (
	"heimdall/backend/internal/domain/entities"
	internal_errors "heimdall/backend/pkg/errors"
)

type AuthService struct {
	sessionService *SessionService
}

func NewAuthService(ss *SessionService) *AuthService {
	return &AuthService{
		sessionService: ss,
	}
}

func (s *AuthService) RefreshTokenPair(userId string, accessTokenJTI string, refreshToken string) (entities.TokenPair, error) {
	return s.sessionService.ReplaceTokenPair(userId, accessTokenJTI, refreshToken)
}

func (s *AuthService) Logout(userId string, accessTokenJTI string) (bool, error) {
	sessionDeleted, err := s.sessionService.sessionRepo.DeleteOneByAccessTokenJTI(accessTokenJTI)
	if err != nil {
		return false, err
	}

	if !sessionDeleted {
		return false, internal_errors.NewNotFoundError("session", accessTokenJTI)
	}

	return true, nil
}
