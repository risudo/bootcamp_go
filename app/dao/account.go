package dao

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"yatter-backend-go/app/domain/object"
	"yatter-backend-go/app/domain/repository"

	"github.com/jmoiron/sqlx"
)

type (
	// Implementation for repository.Account
	account struct {
		db *sqlx.DB
	}
)

// Create accout repository
func NewAccount(db *sqlx.DB) repository.Account {
	return &account{db: db}
}

// FindByUsername : ユーザ名からユーザを取得
func (r *account) FindByUsername(ctx context.Context, username string) (*object.Account, error) {
	entity := new(object.Account)
	const query = "SELECT * FROM account WHERE username = ?"

	err := r.db.QueryRowxContext(ctx, query, username).StructScan(entity)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("%w", err)
	}

	return entity, nil
}

// アカウントを作成
func (r *account) Insert(ctx context.Context, a object.Account) error {
	const query = "INSERT INTO account (username, password_hash) VALUES (?, ?)"

	_, err := r.db.ExecContext(ctx, query, a.Username, a.PasswordHash)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

func (r *account) Update(ctx context.Context, a object.Account) error {
	const query = `
	UPDATE account
	SET
		display_name = ?,
		note = ?,
		avatar = ?,
		header = ?
	WHERE username = ?
	`
	_, err := r.db.ExecContext(ctx, query, a.DisplayName, a.Note, a.Avatar, a.Header, a.Username)
	if err != nil {
		return err
	}
	return nil
}
