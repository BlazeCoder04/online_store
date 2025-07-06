package repositories

import (
	"context"
	"fmt"

	"github.com/BlazeCoder04/online_store/libs/logger"
	"github.com/BlazeCoder04/online_store/services/user/configs"
	"github.com/BlazeCoder04/online_store/services/user/internal/domain/models"
	domain "github.com/BlazeCoder04/online_store/services/user/internal/domain/ports/repositories"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthRepository struct {
	db     *pgxpool.Pool
	logger logger.Logger
	cfg    *configs.Config
}

const loggerTag = "constructor.auth.repository"

func NewAuthRepository(repoLogger logger.Logger, cfg *configs.Config) (domain.AuthRepository, error) {
	repoLogger.Info(loggerTag, "Initializing the auth repository")

	repoLogger.Info(loggerTag, "Connecting to the database via DSN")
	db, err := pgxpool.New(context.Background(), cfg.PostgresDSN)
	if err != nil {
		repoLogger.Error(loggerTag, ErrConnecting, logger.Field{
			Key:   "error",
			Value: err.Error(),
		})

		return nil, fmt.Errorf("%s: %v", ErrConnecting, err)
	}
	repoLogger.Info(loggerTag, "Connection to the database has been completed")

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

	err := r.db.
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
		SELECT *
		FROM users
		WHERE email = $1
	`

	err := r.db.
		QueryRow(ctx, query, email).
		Scan(&user.ID, &user.Email, &user.Password, &user.FirstName, &user.LastName, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
