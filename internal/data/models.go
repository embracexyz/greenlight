package data

import (
	"database/sql"
	"errors"
	"time"
)

var (
	ErrRecordNotFound = errors.New("record not found!")
	ErrEditConflict   = errors.New("edit confict!")
)

// 包含所有model，作为统一的引用入口
type Models struct {
	MovieModel interface {
		Insert(*Movie) error
		Delete(int64) error
		Update(*Movie) error
		Get(int64) (*Movie, error)
		GetAll(string, []string, Filters) ([]*Movie, Metadata, error)
	}
	UserModel interface {
		Insert(*User) error
		GetByEmail(string) (*User, error)
		Update(*User) error
		GetForToken(string, string) (*User, error)
	}
	TokenModel interface {
		New(int64, time.Duration, string) (*Token, error)
		Insert(*Token) error
		DeleteAllForUser(string, int64) error
	}
}

func NewModels(db *sql.DB) Models {
	return Models{
		MovieModel: NewMovieModel(db),
		UserModel:  NewUserModel(db),
		TokenModel: NewTokenModel(db),
	}
}
