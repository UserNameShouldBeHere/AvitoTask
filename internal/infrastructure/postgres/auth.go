package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/UserNameShouldBeHere/AvitoTask/internal/domain"
	customErrors "github.com/UserNameShouldBeHere/AvitoTask/internal/errors"
)

type AuthStorage struct {
	pool PgxPool
}

func NewAuthStorage(pool PgxPool) (*AuthStorage, error) {
	return &AuthStorage{
		pool: pool,
	}, nil
}

func (authStorage *AuthStorage) CreateUser(ctx context.Context, userCreds domain.UserCredantials) error {
	hasUser, err := authStorage.HasUser(ctx, userCreds.UserName)
	if err != nil {
		return fmt.Errorf("%w (postgres.CreateUser): %w", customErrors.ErrInternal, err)
	}
	if hasUser {
		return fmt.Errorf("%w (postgres.CreateUser)", customErrors.ErrAlreadyExists)
	}

	_, err = authStorage.pool.Exec(ctx, `
		insert into users(name, password) values ($1, $2);
	`, userCreds.UserName, userCreds.Password)
	if err != nil {
		return fmt.Errorf("%w (postgres.CreateUser): %w", customErrors.ErrFailedToExecuteQuery, err)
	}

	return nil
}

func (authStorage *AuthStorage) GetPassword(ctx context.Context, name string) (string, error) {
	var password string

	err := authStorage.pool.QueryRow(ctx, `
		select password
		from users
		where name = $1;
	`, name).Scan(&password)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", fmt.Errorf("%w (postgres.GetPassword): %w", customErrors.ErrDoesNotExist, err)
		}

		return "", fmt.Errorf("%w (postgres.GetPassword): %w", customErrors.ErrFailedToExecuteQuery, err)
	}

	return password, nil
}

func (authStorage *AuthStorage) HasUser(ctx context.Context, name string) (bool, error) {
	err := authStorage.pool.QueryRow(ctx, `
		select
		from users
		where name = $1;
	`, name).Scan()
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}

		return false, fmt.Errorf("%w (postgres.HasUser): %w", customErrors.ErrFailedToExecuteQuery, err)
	}

	return true, nil
}
