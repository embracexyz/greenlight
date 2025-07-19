package data

import (
	"database/sql"
	"errors"
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
}

func NewModels(db *sql.DB) Models {
	return Models{
		MovieModel: NewMovieModel(db),
	}
}
