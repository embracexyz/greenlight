package data

import (
	"context"
	"database/sql"
	"time"

	"github.com/lib/pq"
)

type Permisions []string

func (p Permisions) Include(code string) bool {
	for i := range p {
		if code == p[i] {
			return true
		}
	}
	return false
}

type PermisionModel struct {
	DB *sql.DB
}

func NewPermisionModel(db *sql.DB) PermisionModel {
	return PermisionModel{DB: db}
}

func (p PermisionModel) GetAllForUser(userID int64) (Permisions, error) {
	query := `
		select permissions.code
		from permissions
		inner join users_permissions on users_permissions.permission_id = permissions.id
		inner join users on users_permissions.user_id = users.id
		where users.id = $1
	`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := p.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions Permisions
	for rows.Next() {
		var permission string
		err = rows.Scan(&permission)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return permissions, nil
}

func (p PermisionModel) AddForUser(userID int64, codes ...string) error {
	query := `
		insert into users_permissions
		select $1, permissions.id from permissions where permissions.code = any($2)
	`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := p.DB.ExecContext(ctx, query, userID, pq.Array(codes))
	return err
}
