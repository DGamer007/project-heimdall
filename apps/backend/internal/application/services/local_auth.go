package services

import (
	"heimdall/backend/internal/application/dto"
	"heimdall/backend/internal/domain/entities"
	internal_crypto "heimdall/backend/pkg/crypto"
	internal_errors "heimdall/backend/pkg/errors"
)

type LocalAuthService struct {
	userRepo entities.UserRepository
	sessionService *SessionService
}

func NewLocalAuthService(userRepo entities.UserRepository, sessionService *SessionService) *LocalAuthService {
	return &LocalAuthService{
		userRepo:       userRepo,
		sessionService: sessionService,
	}
}

func (s *LocalAuthService) LoginWithPassword(data dto.LoginWithPasswordPayload) (entities.TokenPair, error) {
	user, err := s.userRepo.FindOneByIdentifier(data.Identifier)
	if err != nil {
		if _, isNotFoundError := err.(*internal_errors.NotFoundError); isNotFoundError {
			return entities.TokenPair{}, internal_errors.NewAuthenticationError("Invalid credentials provided")
		}
		return entities.TokenPair{}, err
	}

	if isValid := internal_crypto.VerifyPassword(user.Password, data.Password); !isValid {
		return entities.TokenPair{}, internal_errors.NewAuthenticationError("Invalid credentials provided")
	}

	var tokenPair entities.TokenPair
	tokenPair, err = s.sessionService.GenerateTokenPair(user.Id)
	if err != nil {
		return entities.TokenPair{}, err
	}

	return tokenPair, nil
}

func (s *LocalAuthService) RegisterWithPassword(data dto.RegisterWithPasswordPayload) (entities.TokenPair, error) {
	if exists, err := s.userRepo.CheckIfExistsByEmail(data.Email); err != nil {
		return entities.TokenPair{}, err
	} else if exists {
		return entities.TokenPair{}, internal_errors.NewDuplicateResourceError("User", "email", data.Email)
	}

	if exists, err := s.userRepo.CheckIfExistsByUserName(data.UserName); err != nil {
		return entities.TokenPair{}, err
	} else if exists {
		return entities.TokenPair{}, internal_errors.NewDuplicateResourceError("User", "username", data.UserName)
	}

	var hashedPassword string
	hashedPassword, err := internal_crypto.HashPassword(data.Password)
	if err != nil {
		return entities.TokenPair{}, internal_errors.NewInfrastructureError("Failed to hash password", err)
	}

	var id string
	id, err = s.userRepo.CreateOne(&entities.User{
		Email:     data.Email,
		Password:  hashedPassword,
		UserName:  data.UserName,
		FirstName: data.FirstName,
		LastName:  data.LastName,
	})
	if err != nil {
		return entities.TokenPair{}, err
	}

	var tokenPair entities.TokenPair
	tokenPair, err = s.sessionService.GenerateTokenPair(id)
	if err != nil {
		return entities.TokenPair{}, err
	}

	return tokenPair, nil
}
