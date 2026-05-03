package service

import (
	"context"
	"errors"
	"fmt"
	repositories "geo-project/internal/auth/repository"
	apperrors "geo-project/pkg/errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	RegisterService(ctx context.Context, login string, password string, ipAddress string, device string) (string, error)
	LoginService(ctx context.Context, login string, password string, ipAddress string, device string) (string, error)
	GetSystemStatistics(ctx context.Context) (map[string]int, error)
}

type authService struct {
	repo      repositories.UserRepository
	jwtSecret string
}

func NewAuthService(repo repositories.UserRepository, jwtSecret string) AuthService {
	return &authService{
		repo:      repo,
		jwtSecret: jwtSecret,
	}
}

type JwtCustomClaims struct {
	UserID int32  `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func (s *authService) RegisterService(ctx context.Context, login string, password string, ipAddress string, device string) (string, error) {

	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("Error when hashing password: %w", err)
	}

	hashedPassword := string(bytes)

	res, err := s.repo.CreateNewUser(ctx, login, hashedPassword)
	if err != nil {
		if errors.Is(err, repositories.ErrUserAlreadyExists) {
			return "", apperrors.NewAppError("CONFLICT", "User with this login already exists")
		}
		return "", fmt.Errorf("failed to create user: %w", err)
	}

	err = s.repo.CreateSession(ctx, int(res.ID), ipAddress, device)
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	claims := &JwtCustomClaims{
		UserID: int32(res.ID),
		Role:   res.Role.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 72)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	secretKey := s.jwtSecret
	signedToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return signedToken, nil

}

func (s *authService) LoginService(ctx context.Context, login string, password string, ipAddress string, device string) (string, error) {
	result, err := s.repo.GetUserByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			return "", apperrors.NewAppError("UNAUTHORIZED", "Incorrect login or password")
		}
		return "", fmt.Errorf("failed to retrieve user: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(result.PasswordHash), []byte(password))
	if err != nil {
		return "", apperrors.NewAppError("UNAUTHORIZED", "Incorrect login or password")
	}

	err = s.repo.CreateSession(ctx, int(result.ID), ipAddress, device)
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}

	claims := &JwtCustomClaims{
		UserID: int32(result.ID),
		Role:   result.Role.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 72)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	secretKey := s.jwtSecret
	signedToken, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	return signedToken, nil

}

func (s *authService) GetSystemStatistics(ctx context.Context) (map[string]int, error) {
	stats, err := s.repo.GetRoleStatsRaw(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch system stats: %w", err)
	}

	return stats, nil
}