package service

import (
	"context"
	"errors"
	"go-chat/internal/auth"
	"go-chat/internal/dto"
	"go-chat/internal/models"
	"go-chat/internal/repositories"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidInput        = errors.New("invalid input")
	ErrEmailAlradyTaken    = errors.New("email alrady taken")
	ErrUsernsmeAlradyTaken = errors.New("username alrady taken")
	ErrInvalidCredentials  = errors.New("invalid email or password")
)

type AuthService struct {
	userRepo  *repositories.UserRepository
	jwtSecret string
	jwtTTL    time.Duration
}

func NewAuthService(userRepo *repositories.UserRepository, jwtSecret string, jwtTTL time.Duration) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
		jwtTTL:    jwtTTL,
	}
}

func (s *AuthService) Register(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, error) {
	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Password = strings.TrimSpace(req.Password)

	if len(req.Username) < 3 || len(req.Username) > 20 {
		return nil, errors.New("username must be at least 3 characters")
	}

	if !strings.Contains(req.Email, "@") {
		return nil, errors.New("invalid email format")
	}

	if len(req.Password) < 6 {
		return nil, errors.New("password must be at least 6 characters")
	}

	_, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err == nil {
		return nil, ErrEmailAlradyTaken
	}

	if !errors.Is(err, repositories.ErrNotFound) {
		return nil, err
	}

	_, err = s.userRepo.FindByUsername(ctx, req.Username)
	if !errors.Is(err, repositories.ErrNotFound) {
		return nil, ErrUsernsmeAlradyTaken
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(passwordHash),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	token, err := auth.GenerateToken(s.jwtSecret, user, s.jwtTTL)

	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		Token: token,
		User:  toUserResponse(user),
	}, nil
}

func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error) {
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Password = strings.TrimSpace(req.Password)

	if req.Email == "" || req.Password == "" {
		return nil, ErrInvalidCredentials
	}

	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if errors.Is(err, repositories.ErrNotFound) {
		return nil, ErrInvalidCredentials
	}

	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := auth.GenerateToken(s.jwtSecret, user, s.jwtTTL)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		Token: token,
		User:  toUserResponse(user),
	}, nil

}

func (s *AuthService) GetUserByID(ctx context.Context, userID int64) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	response := toUserResponse(user)

	return &response, nil
}

func toUserResponse(user *models.User) dto.UserResponse {
	return dto.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}
}
