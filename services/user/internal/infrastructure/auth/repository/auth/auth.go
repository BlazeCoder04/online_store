package repository

import (
	"context"
	"fmt"

	"github.com/BlazeCoder04/online_store/libs/logger"
	"github.com/BlazeCoder04/online_store/services/user/configs"
	"github.com/BlazeCoder04/online_store/services/user/internal/domain/models"
	domain "github.com/BlazeCoder04/online_store/services/user/internal/domain/ports/auth/repository"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthRepository struct {
	pool   *pgxpool.Pool
	logger logger.Logger
	cfg    *configs.Config
}

func NewAuthRepository(repoLogger logger.Logger, cfg *configs.Config) (domain.AuthRepository, error) {
	repoLogger.Info("Initializing the auth repository")

	repoLogger.Info("Connecting to the database via DSN")
	db, err := pgxpool.New(context.Background(), cfg.PostgresDSN)
	if err != nil {
		repoLogger.Error(ErrConnecting, logger.Field{
			Key:   "error",
			Value: err.Error(),
		})
		return nil, fmt.Errorf("%s: %w", ErrConnecting, err)
	}
	repoLogger.OK("Connection to the database has been completed")

	return &AuthRepository{
		db,
		repoLogger,
		cfg,
	}, nil
}

func (r *AuthRepository) Create(ctx context.Context, email, password, firstName, lastName string) (*models.User, error) {
	user := &models.User{
		Email:     email,
		Password:  password,
		FirstName: firstName,
		LastName:  lastName,
		Role:      models.UserRole,
	}

	query := `
		INSERT INTO users (email, password, first_name, last_name, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	err := r.pool.
		QueryRow(ctx, query, user.Email, user.Password, user.FirstName, user.LastName, user.Role).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *AuthRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User

	query := `
		SELECT id, email, password, first_name, last_name, role, created_at, updated_at, deleted_at
		FROM users
		WHERE email = $1
	`

	err := r.pool.
		QueryRow(ctx, query, email).
		Scan(
			&user.ID, &user.Email, &user.Password, &user.FirstName, &user.LastName, &user.Role,
			&user.CreatedAt, &user.UpdatedAt, &user.DeletedAt,
		)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *AuthRepository) FindByID(ctx context.Context, ID string) (*models.User, error) {
	var user models.User

	query := `
		SELECT id, email, password, first_name, last_name, role, created_at, updated_at, deleted_at
		FROM users
		WHERE id = $1
	`

	err := r.pool.
		QueryRow(ctx, query, ID).
		Scan(
			&user.ID, &user.Email, &user.Password, &user.FirstName, &user.LastName, &user.Role,
			&user.CreatedAt, &user.UpdatedAt, &user.DeletedAt,
		)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
