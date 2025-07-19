package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/embracexyz/greenlight/internal/validator"
	"github.com/lib/pq"
)

type Movie struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"`
	Title     string    `json:"title"`
	Year      int32     `json:"year,omitempty"`
	Runtime   Runtime   `json:"runtime"`
	Genres    []string  `json:",omitempty"`
	Version   int32     `json:"version"`
}

type MovieModel struct {
	DB *sql.DB
}

func NewMovieModel(db *sql.DB) MovieModel {
	return MovieModel{DB: db}
}

func (m MovieModel) Insert(movie *Movie) error {
	stmt := `
		insert into movies (title, year, runtime, genres)
		values ($1, $2, $3, $4)
		returning id, created_at, version
	`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}
	return m.DB.QueryRowContext(ctx, stmt, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

func (m MovieModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		delete from movies where id = $1
	`
	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowAffected == 0 {
		return ErrRecordNotFound
	}
	return nil

}

func (m MovieModel) Update(movie *Movie) error {
	query := `
		update movies 
		set title= $1, year = $2, runtime = $3, genres = $4, version = version + 1 
		where id = $5 and version = $6
		returning version
	`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres), movie.ID, movie.Version}

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

// 不用uint64 因为pg不支持无符号整数类型，且database/sql最大支持就是int64最大值，uint64可能超过导致panic
func (m MovieModel) Get(id int64) (*Movie, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		select id, created_at, title, year, runtime, genres, version
		from movies
		where id = $1
	`
	var movie Movie

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(&movie.ID, &movie.CreatedAt, &movie.Title, &movie.Year, &movie.Runtime, pq.Array(&movie.Genres), &movie.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &movie, nil

}

func ValidateMove(v *validator.Validator, movie *Movie) {
	v.Check(movie.Title != "", "title", "must be provided")
	v.Check(len(movie.Title) <= 500, "title", "must not be more than 500 bytes long")
	v.Check(movie.Year != 0, "year", "must be provided")
	v.Check(movie.Year >= 1888, "year", "must be greater than 1888")
	v.Check(movie.Year <= int32(time.Now().Year()), "year", "must not be in the future")
	v.Check(movie.Runtime != 0, "runtime", "must be provided")
	v.Check(movie.Runtime > 0, "runtime", "must be a positive integer")
	v.Check(movie.Genres != nil, "genres", "must be provided")
	v.Check(len(movie.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(movie.Genres) <= 5, "genres", "must not contain more than 5 genres")
	v.Check(validator.Unique(movie.Genres), "genres", "must not contain duplicate values")
}
