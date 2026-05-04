package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	repositories "geo-project/internal/auth/repository"
	apperrors "geo-project/pkg/errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	RegisterService(ctx context.Context, login string, password string, ipAddress string, device string) (*AuthTokens, error)
	LoginService(ctx context.Context, login string, password string, ipAddress string, device string) (*AuthTokens, error)
	RefreshService(ctx context.Context, refreshToken string) (*AuthTokens, error)
	GetSystemStatistics(ctx context.Context) (map[string]int, error)
}

type AuthTokens struct {
	AccessToken  string
	RefreshToken string
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

func generateRefreshToken() (string, error) {
	bytes := make([]byte, 48)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func buildAccessToken(secretKey string, userID int32, role string) (string, error) {
	claims := &JwtCustomClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

func (s *authService) RegisterService(ctx context.Context, login string, password string, ipAddress string, device string) (*AuthTokens, error) {

	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("Error when hashing password: %w", err)
	}

	hashedPassword := string(bytes)

	res, err := s.repo.CreateNewUser(ctx, login, hashedPassword)
	if err != nil {
		if errors.Is(err, repositories.ErrUserAlreadyExists) {
			return nil, apperrors.NewAppError("CONFLICT", "User with this login already exists")
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	err = s.repo.CreateSession(ctx, int(res.ID), ipAddress, device)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	secretKey := s.jwtSecret
	accessToken, err := buildAccessToken(secretKey, int32(res.ID), res.Role.String())
	if err != nil {
		return nil, err
	}

	refreshToken, err := generateRefreshToken()
	if err != nil {
		return nil, err
	}

	if err := s.repo.CreateRefreshToken(ctx, int(res.ID), refreshToken, time.Now().Add(30*24*time.Hour)); err != nil {
		return nil, fmt.Errorf("failed to create refresh token: %w", err)
	}

	return &AuthTokens{AccessToken: accessToken, RefreshToken: refreshToken}, nil

}

func (s *authService) LoginService(ctx context.Context, login string, password string, ipAddress string, device string) (*AuthTokens, error) {
	result, err := s.repo.GetUserByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			return nil, apperrors.NewAppError("UNAUTHORIZED", "Incorrect login or password")
		}
		return nil, fmt.Errorf("failed to retrieve user: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(result.PasswordHash), []byte(password))
	if err != nil {
		return nil, apperrors.NewAppError("UNAUTHORIZED", "Incorrect login or password")
	}

	err = s.repo.CreateSession(ctx, int(result.ID), ipAddress, device)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	secretKey := s.jwtSecret
	accessToken, err := buildAccessToken(secretKey, int32(result.ID), result.Role.String())
	if err != nil {
		return nil, err
	}

	refreshToken, err := generateRefreshToken()
	if err != nil {
		return nil, err
	}

	if err := s.repo.CreateRefreshToken(ctx, int(result.ID), refreshToken, time.Now().Add(30*24*time.Hour)); err != nil {
		return nil, fmt.Errorf("failed to create refresh token: %w", err)
	}

	return &AuthTokens{AccessToken: accessToken, RefreshToken: refreshToken}, nil

}

func (s *authService) RefreshService(ctx context.Context, refreshToken string) (*AuthTokens, error) {
	foundToken, err := s.repo.GetRefreshTokenByToken(ctx, refreshToken)
	if err != nil {
		if errors.Is(err, repositories.ErrRefreshTokenNotFound) {
			return nil, apperrors.NewAppError("UNAUTHORIZED", "Invalid refresh token")
		}
		return nil, fmt.Errorf("failed to retrieve refresh token: %w", err)
	}

	if foundToken.Revoked || foundToken.ExpiresAt.Before(time.Now()) {
		return nil, apperrors.NewAppError("UNAUTHORIZED", "Invalid refresh token")
	}

	user, err := foundToken.QueryUser().Only(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load refresh token user: %w", err)
	}

	accessToken, err := buildAccessToken(s.jwtSecret, int32(user.ID), user.Role.String())
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := generateRefreshToken()
	if err != nil {
		return nil, err
	}

	if err := s.repo.CreateRefreshToken(ctx, int(user.ID), newRefreshToken, time.Now().Add(30*24*time.Hour)); err != nil {
		return nil, fmt.Errorf("failed to create new refresh token: %w", err)
	}

	if err := s.repo.RevokeRefreshToken(ctx, foundToken.ID); err != nil {
		return nil, fmt.Errorf("failed to revoke old refresh token: %w", err)
	}

	return &AuthTokens{AccessToken: accessToken, RefreshToken: newRefreshToken}, nil

}

func (s *authService) GetSystemStatistics(ctx context.Context) (map[string]int, error) {
	stats, err := s.repo.GetRoleStatsRaw(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch system stats: %w", err)
	}

	return stats, nil
}
