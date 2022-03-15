package dao

import (
	"context"
	"fmt"
	"log"
	"yatter-backend-go/app/domain/object"
	"yatter-backend-go/app/domain/repository"

	"github.com/jmoiron/sqlx"
)

type (
	// Implementation for repository.Account
	relation struct {
		db *sqlx.DB
	}
)

func NewRelation(db *sqlx.DB) repository.Relation {
	return &relation{db: db}
}

func (r *relation) Follow(ctx context.Context, followingID object.AccountID, followerID object.AccountID) error {
	// TODO: かぶってたらなにもしない

	const query = "INSERT INTO relation (following_id, follower_id) VALUES(?, ?)"

	_, err := r.db.ExecContext(ctx, query, followingID, followerID)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	log.Println("followerID", followerID)
	log.Println("followingID", followingID)
	return nil
}

func (r *relation) Following(ctx context.Context, followingID object.AccountID) ([]object.Account, error) {
	return nil, nil
}
