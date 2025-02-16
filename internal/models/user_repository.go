package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/devingoodsell/go-links-free/internal/db"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository struct {
	db *db.DB
}

func NewUserRepository(db *db.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *User, password string) error {
	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO users (email, password_hash, is_admin)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`

	err = r.db.QueryRowContext(
		ctx, query,
		user.Email,
		string(hashedPassword),
		user.IsAdmin,
	).Scan(&user.ID, &user.CreatedAt)

	return err
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, email, password_hash, is_admin, created_at, last_login_at
		FROM users
		WHERE email = $1`

	user := &User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.IsAdmin,
		&user.CreatedAt,
		&user.LastLoginAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	if err != nil {
		log.Printf("Database error getting user: %v", err)
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) VerifyPassword(user *User, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	return err == nil
}

func (r *UserRepository) UpdatePassword(ctx context.Context, userID int64, newPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	query := `
		UPDATE users
		SET password_hash = $1, updated_at = NOW()
		WHERE id = $2`

	_, err = r.db.ExecContext(ctx, query, string(hashedPassword), userID)
	return err
}

func (r *UserRepository) SetAdminStatus(ctx context.Context, userID int64, isAdmin bool) error {
	query := `
		UPDATE users
		SET is_admin = $1, updated_at = NOW()
		WHERE id = $2`

	result, err := r.db.ExecContext(ctx, query, isAdmin, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("user not found")
	}

	return nil
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID int64, lastLoginAt *time.Time) error {
	query := `
		UPDATE users 
		SET last_login_at = $1 
		WHERE id = $2
		RETURNING id, last_login_at`

	var updatedID int64
	var updatedTime time.Time
	err := r.db.QueryRowContext(ctx, query, lastLoginAt, userID).Scan(&updatedID, &updatedTime)
	if err != nil {
		return fmt.Errorf("failed to update last login: %v", err)
	}

	return nil
}
