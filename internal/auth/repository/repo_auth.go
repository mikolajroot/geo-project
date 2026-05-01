package repositories

import (
	"context"
	"errors"
	"fmt"
	"geo-project/internal/auth/ent"
)

type UserRepository interface {
	CreateNewUser(ctx context.Context,login string,passwordHash string) (*ent.User, error)
}

type userRepository struct {
	db *ent.Client
}

func NewUserRepository(db *ent.Client) UserRepository {
	return &userRepository{
		db: db,
	}
}

var ErrUserAlreadyExists = errors.New("user with this login already exists")

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
