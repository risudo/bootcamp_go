package dao_test

import (
	"context"
	"fmt"
	"math"
	"testing"
	"yatter-backend-go/app/config"
	"yatter-backend-go/app/dao"
	"yatter-backend-go/app/domain/object"
	"yatter-backend-go/app/domain/repository"
	"yatter-backend-go/app/handler/parameters"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

//TODO:失敗したときもロールバックさせる

var preparedAccount = &object.Account{
	Username: "Michael",
}

var preparedStatus = &object.Status{
	Account: preparedAccount,
	Content: "content",
}

const notExistingUser = "notexist"

type mockdao struct {
	db *sqlx.DB
}

func (m *mockdao) Account() repository.Account {
	return dao.NewAccount(m.db)
}

func (m *mockdao) Status() repository.Status {
	return dao.NewStatus(m.db)
}

func (m *mockdao) Relation() repository.Relation {
	return dao.NewRelation(m.db)
}

func (m *mockdao) Attachment() repository.Attachment {
	return dao.NewAttachment(m.db)
}

func initMockDB(config dao.DBConfig) (*sqlx.DB, error) {
	driverName := "mysql"
	db, err := sqlx.Open(driverName, config.FormatDSN())
	if err != nil {
		return nil, fmt.Errorf("sqlx.Open failed: %w", err)
	}

	return db, nil
}

func setupDB() (*mockdao, *sqlx.Tx, error) {
	daoCfg := config.MySQLConfig()
	db, err := initMockDB(daoCfg)
	if err != nil {
		return nil, nil, err
	}
	// トランザクション開始
	tx, _ := db.Beginx()
	// テーブルリセット
	if _, err := db.Exec("SET FOREIGN_KEY_CHECKS=0"); err != nil {
		return nil, nil, err
	}
	for _, table := range []string{"account", "status", "relation", "attachment", "status_contain_attachment"} {
		if _, err := db.Exec("DELETE FROM " + table); err != nil {
			return nil, nil, err
		}
	}
	mockdao := &mockdao{db: db}
	ctx := context.Background()
	preparedAccount.ID, err = mockdao.Account().Insert(ctx, *preparedAccount)
	if err != nil {
		tx.Rollback()
		return nil, nil, err
	}
	preparedStatus.ID, err = mockdao.Status().Insert(ctx, *preparedStatus, nil)
	return mockdao, tx, nil
}

func TestFindByUsername(t *testing.T) {
	m, tx, err := setupDB()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	defer m.db.Close()

	repo := m.Account()
	ctx := context.Background()

	tests := []struct {
		name          string
		userName      string
		expectAccount *object.Account
	}{
		{
			name:          "NotExistingUser",
			userName:      notExistingUser,
			expectAccount: nil,
		},
		{
			name:          "ExistingUser",
			userName:      preparedAccount.Username,
			expectAccount: preparedAccount,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := repo.FindByUsername(ctx, tt.userName)
			if err != nil {
				t.Fatal(err)
			}
			if actual == nil && actual == tt.expectAccount {
				return
			}
			opt := cmpopts.IgnoreFields(object.Account{}, "CreateAt")
			if d := cmp.Diff(actual, tt.expectAccount, opt); len(d) != 0 {
				t.Fatalf("differs: (-got +want)\n%s", d)
			}
		})
	}
}

func TestAccountUpdate(t *testing.T) {
	m, tx, err := setupDB()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	defer m.db.Close()

	repo := m.Account()
	ctx := context.Background()

	displayName := "Mike"
	note := "note"

	tests := []struct {
		name          string
		account       *object.Account
		expectErr     bool
		expectAccount *object.Account
	}{
		{
			name: "Update",
			account: &object.Account{
				ID:          preparedAccount.ID,
				Username:    preparedAccount.Username,
				DisplayName: &displayName,
				Note:        &note,
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotErr bool
			err := repo.Update(ctx, *tt.account)

			if err != nil {
				gotErr = true
			}
			if gotErr != tt.expectErr {
				if err != nil {
					t.Fatal(err)
				} else {
					t.Fatal("expect error")
				}
			}

			updated, err := repo.FindByUsername(ctx, preparedAccount.Username)
			opt := cmpopts.IgnoreFields(object.Account{}, "CreateAt")
			if d := cmp.Diff(updated, tt.account, opt); len(d) != 0 {
				tx.Rollback()
				t.Fatalf("differs: (-got +want)\n%s", d)
			}
		})
	}
}

func TestStatusFindByID(t *testing.T) {
	m, tx, err := setupDB()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	defer m.db.Close()

	repo := m.Status()
	ctx := context.Background()

	tests := []struct {
		name         string
		id           object.StatusID
		expectStatus *object.Status
	}{
		{
			name:         "FindNotExistingID",
			id:           -100,
			expectStatus: nil,
		},
		{
			name:         "FindPreparedStatus",
			id:           preparedStatus.ID,
			expectStatus: preparedStatus,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := repo.FindByID(ctx, tt.id)
			if err != nil {
				t.Fatal(err)
			}
			if actual == nil && actual == tt.expectStatus {
				return
			}
			opt := cmpopts.IgnoreFields(object.Status{}, "CreateAt", "Account")
			if d := cmp.Diff(actual, tt.expectStatus, opt); len(d) != 0 {
				t.Fatalf("differs: (-got +want)\n%s", d)
			}
		})
	}
}

func TestStatusDelete(t *testing.T) {
	m, tx, err := setupDB()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	defer m.db.Close()

	repo := m.Status()
	ctx := context.Background()

	tests := []struct {
		name        string
		id          object.StatusID
		expectIsErr bool
	}{
		{
			name:        "delete",
			id:          preparedStatus.ID,
			expectIsErr: false,
		},
		{
			name:        "deleteNotExist",
			id:          preparedStatus.ID,
			expectIsErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Delete(ctx, tt.id)
			if tt.expectIsErr == false && err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestStatusPublicTimeline(t *testing.T) {
	m, tx, err := setupDB()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	defer m.db.Close()

	timeline := object.Timelines{
		{
			Account: preparedAccount,
			Content: "5000兆円欲しい",
		},
		{
			Account: preparedAccount,
			Content: "5億年ぶりに焼肉食べた",
		},
		{
			Account: preparedAccount,
			Content: "スタバわず",
		},
		{
			Account: preparedAccount,
			Content: "駆け出しエンジニアと繋がりたい",
		},
	}
	ctx := context.Background()
	repo := m.Status()
	repo.Delete(ctx, preparedStatus.ID)
	for i, s := range timeline {
		timeline[i].ID, err = repo.Insert(ctx, s, nil)
		if err != nil {
			panic(err)
		}
	}

	tests := []struct {
		name           string
		expectTimeline object.Timelines
		parameter      *object.Parameters
	}{
		{
			name:           "Fetch",
			expectTimeline: timeline,
			parameter:      parameters.Default(),
		},
		{
			name:           "Limit",
			expectTimeline: timeline[0:1],
			parameter: &object.Parameters{
				MaxID:   math.MaxInt64,
				SinceID: 0,
				Limit:   1,
			},
		},
		{
			name:           "MaxID",
			expectTimeline: timeline[0:3],
			parameter: &object.Parameters{
				MaxID:   timeline[len(timeline)-1].ID,
				SinceID: 0,
				Limit:   80,
			},
		},
		{
			name:           "SinceID",
			expectTimeline: timeline[1:],
			parameter: &object.Parameters{
				MaxID:   math.MaxInt64,
				SinceID: timeline[0].ID,
				Limit:   80,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := repo.PublicTimeline(ctx, *tt.parameter)
			if err != nil {
				t.Fatal(err)
			}
			opt := cmpopts.IgnoreTypes(object.DateTime{}, object.StatusID(1))
			if d := cmp.Diff(actual, tt.expectTimeline, opt); len(d) != 0 {
				t.Fatalf("differs: (-got +want)\n%s", d)
			}
		})
	}
}

func TestFollow(t *testing.T) {
	m, tx, err := setupDB()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	defer m.db.Close()
	ctx := context.Background()
	target := &object.Account{
		Username: "john",
	}
	m.Account().Insert(ctx, *target)

	tests := []struct {
		name   string
		f      func(ctx context.Context, lID object.AccountID, tID object.AccountID, m *mockdao)
		expect bool
	}{
		{
			name:   "IsFllowingExpectFalse",
			f:      func(ctx context.Context, lID object.AccountID, tID object.AccountID, m *mockdao) {},
			expect: false,
		},
		{
			name: "Follow",
			f: func(ctx context.Context, lID object.AccountID, tID object.AccountID, m *mockdao) {
				m.Relation().Follow(ctx, lID, tID)
			},
			expect: true,
		},
		{
			name: "UnFollow",
			f: func(ctx context.Context, lID object.AccountID, tID object.AccountID, m *mockdao) {
				m.Relation().Unfollow(ctx, lID, tID)
			},
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.f(ctx, preparedAccount.ID, target.ID, m)
			actual, err := m.Relation().IsFollowing(ctx, preparedAccount.ID, target.ID)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tt.expect, actual)
		})
	}
}

func TestFollowingAndFollowers(t *testing.T) {
	m, tx, err := setupDB()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	defer m.db.Close()
	ctx := context.Background()

	accounts := []object.Account{
		{
			Username: "a",
		},
		{
			Username: "b",
		},
		{
			Username: "c",
		},
	}
	for i, a := range accounts {
		accounts[i].ID, err = m.Account().Insert(ctx, a)
		if err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name      string
		parameter object.Parameters
		f         func(ctx context.Context, id object.AccountID, p object.Parameters) []object.Account
		expect    []object.Account
	}{
		{},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.f(ctx, preparedAccount.ID, tt.parameter)

			if d := cmp.Diff(actual, tt.expect); len(d) != 0 {
				t.Fatalf("differs: (-got +want)\n%s", d)
			}
		})
	}

}
