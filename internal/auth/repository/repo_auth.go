package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"geo-project/internal/auth/ent"
	"geo-project/internal/auth/ent/refreshtoken"
	"geo-project/internal/auth/ent/user"
	"time"
)

type UserRepository interface {
	CreateNewUser(ctx context.Context, login string, passwordHash string) (*ent.User, error)
	GetUserByLogin(ctx context.Context, login string) (*ent.User, error)
	CreateSession(ctx context.Context, userID int, ipAddress string, device string) error
	GetRoleStatsRaw(ctx context.Context) (map[string]int, error)
	CreateRefreshToken(ctx context.Context, userID int, token string, expiresAt time.Time) error
	GetRefreshTokenByToken(ctx context.Context, token string) (*ent.RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, tokenID int) error
}

type userRepository struct {
	db    *ent.Client
	sqlDB *sql.DB
}

func NewUserRepository(db *ent.Client, sqlDB *sql.DB) UserRepository {
	return &userRepository{
		db:    db,
		sqlDB: sqlDB,
	}
}

var ErrUserAlreadyExists = errors.New("user with this login already exists")
var ErrUserNotFound = errors.New("User login doesnt exists")
var ErrRefreshTokenNotFound = errors.New("refresh token doesnt exists")

func (r *userRepository) CreateNewUser(ctx context.Context, login string, passwordHash string) (*ent.User, error) {

	result, err := r.db.User.Create().
		SetLogin(login).
		SetPasswordHash(passwordHash).
		Save(ctx)

	if err != nil {
		if ent.IsConstraintError(err) {
			return nil, ErrUserAlreadyExists
		}

		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return result, nil
}

func (r *userRepository) GetUserByLogin(ctx context.Context, login string) (*ent.User, error) {
	result, err := r.db.User.Query().Where(user.LoginEQ(login)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return result, nil
}

func (r *userRepository) CreateSession(ctx context.Context, userID int, ipAddress string, device string) error {
	_, err := r.db.Session.Create().
		SetUserID(userID).
		SetIPAddress(ipAddress).
		SetDevice(device).
		Save(ctx)

	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

func (r *userRepository) GetRoleStatsRaw(ctx context.Context) (map[string]int, error) {
	query := `SELECT role, COUNT(*) FROM users GROUP BY role;`

	rows, err := r.sqlDB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("raw query failed: %w", err)
	}
	defer rows.Close()

	stats := make(map[string]int)
	for rows.Next() {
		var role string
		var count int64

		if err := rows.Scan(&role, &count); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		stats[role] = int(count)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration failed: %w", err)
	}

	return stats, nil
}

func (r *userRepository) CreateRefreshToken(ctx context.Context, userID int, token string, expiresAt time.Time) error {
	_, err := r.db.RefreshToken.Create().
		SetUserID(userID).
		SetToken(token).
		SetExpiresAt(expiresAt).
		Save(ctx)

	if err != nil {
		if ent.IsConstraintError(err) {
			return fmt.Errorf("refresh token already exists: %w", err)
		}
		return fmt.Errorf("failed to create refresh token: %w", err)
	}

	return nil
}

func (r *userRepository) GetRefreshTokenByToken(ctx context.Context, token string) (*ent.RefreshToken, error) {
	result, err := r.db.RefreshToken.Query().Where(refreshtoken.TokenEQ(token)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrRefreshTokenNotFound
		}
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}

	return result, nil
}

func (r *userRepository) RevokeRefreshToken(ctx context.Context, tokenID int) error {
	err := r.db.RefreshToken.UpdateOneID(tokenID).
		SetRevoked(true).
		Exec(ctx)

	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	return nil
}
