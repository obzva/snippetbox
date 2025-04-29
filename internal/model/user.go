package model

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID             int
	Name           string
	Email          string
	HashedPassword []byte
	Created        time.Time
}

type UserModel struct {
	DBPool *pgxpool.Pool
}

func (um *UserModel) Insert(ctx context.Context, name, email, password string) error {
	hashedPW, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	stmt := `INSERT INTO "user" (name, email, hashed_password, created)
	VALUES($1, $2, $3, CURRENT_TIMESTAMP)`

	_, err = um.DBPool.Exec(ctx, stmt, name, email, hashedPW)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" { // postgresql error code: unique_violation
				return ErrDuplicateEmail
			}
		}
		return err
	}

	return nil
}

func (um *UserModel) Authenticate(ctx context.Context, email, password string) (int, error) {
	stmt := `SELECT id, hashed_password
	FROM "user"
	WHERE email = $1`

	var id int
	var hashedPW []byte
	if err := um.DBPool.QueryRow(ctx, stmt, email).Scan(&id, &hashedPW); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrInvalidCredentials
		}
		return 0, err
	}

	if err := bcrypt.CompareHashAndPassword(hashedPW, []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, ErrInvalidCredentials
		}
		return 0, err
	}

	return id, nil
}

// check if user with this id exists
func (um *UserModel) Check(ctx context.Context, id int) (bool, error) {
	var ok bool

	stmt := `SELECT EXISTS(
	SELECT id
	FROM "user"
	WHERE id = $1
	)`

	if err := um.DBPool.QueryRow(ctx, stmt, id).Scan(&ok); err != nil {
		return false, err
	}

	return ok, nil
}
