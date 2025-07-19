package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/embracexyz/greenlight/internal/data"
	"github.com/embracexyz/greenlight/internal/validator"
)

func (app *application) listMoviesHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title  string
		Genres []string
		data.Filters
	}

	v := validator.New()

	input.Title = app.readString(r.URL.Query(), "title", "")
	input.Genres = app.readCSV(r.URL.Query(), "genres", []string{})
	input.Filters.Page = app.readInt(r.URL.Query(), "page", 1, v)
	input.Filters.PageSize = app.readInt(r.URL.Query(), "page_size", 20, v)

	input.Filters.Sort = app.readString(r.URL.Query(), "sort", "id")
	input.Filters.SortSafelist = []string{"id", "title", "year", "runtime", "-id", "-title", "-year", "-runtime"}

	if data.ValidateFilters(v, &input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.FieldErrors)
		return
	}

	movies, metadata, err := app.models.MovieModel.GetAll(input.Title, input.Genres, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = app.writeJson(w, http.StatusOK, envelope{"movies": movies, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	movie, err := app.models.MovieModel.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJson(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title   string       `json:"title"`
		Year    int32        `json:"year"`
		Runtime data.Runtime `json:"runtime"`
		Genres  []string     `json:"genres"`
	}

	err := app.readJson(w, r, &input)
	if err != nil {
		app.badRequestErrorReponse(w, r, err)
		return
	}

	// valid request
	v := validator.New()
	movie := &data.Movie{
		Title:   input.Title,
		Year:    input.Year,
		Runtime: input.Runtime,
		Genres:  input.Genres,
	}
	if data.ValidateMove(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.FieldErrors)
		return
	}

	// insert
	err = app.models.MovieModel.Insert(movie)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/movies/%d", movie.ID))
	err = app.writeJson(w, http.StatusCreated, envelope{"movie": movie}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateMovieHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 解析url id是否合法，否则return badRequest
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// 2. 查询id是否真实存在，否则return nodFound
	movie, err := app.models.MovieModel.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// 3. 把更新的值，覆盖从数据库查询的字段（其他字段保留）
	var input struct {
		Title   string       `json:"title"`
		Year    int32        `json:"year"`
		Runtime data.Runtime `json:"runtime"`
		Genres  []string     `json:"genres"`
	}

	err = app.readJson(w, r, &input)
	if err != nil {
		app.badRequestErrorReponse(w, r, err)
		return
	}

	movie.Title = input.Title
	movie.Year = input.Year
	movie.Runtime = input.Runtime
	movie.Genres = input.Genres

	// 4. validatror严重，否则return， baserequest
	v := validator.New()
	if data.ValidateMove(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.FieldErrors)
		return
	}

	// 5. 更新，err则判断err类型，返回对应响应，happy path则write更新后的json
	err = app.models.MovieModel.Update(movie)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJson(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) deleteMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.MovieModel.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJson(w, http.StatusOK, envelope{"message": "movie delete successfully!"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) partialUpdateMovieHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 解析url id是否合法，否则return badRequest
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// 2. 查询id是否真实存在，否则return nodFound
	movie, err := app.models.MovieModel.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// 3. 把更新的值，覆盖从数据库查询的字段（其他字段保留）
	// ! 把值类型改为其指针，这样json解析时，没传值的就会保持为nil，根据是否为nil可判断客户端是否传值，只针对传值的字段进行覆盖更新，实现partialUpdate的效果
	//		如果是"key": null，默认json解析器也会忽略该值，认为没传；另注意"key": ""，是传值了，值是空字符串
	var input struct {
		Title   *string       `json:"title"`
		Year    *int32        `json:"year"`
		Runtime *data.Runtime `json:"runtime"`
		Genres  []string      `json:"genres"`
	}

	err = app.readJson(w, r, &input)
	if err != nil {
		app.badRequestErrorReponse(w, r, err)
		return
	}

	if input.Title != nil {
		movie.Title = *input.Title
	}
	if input.Year != nil {
		movie.Year = *input.Year
	}
	if input.Runtime != nil {
		movie.Runtime = *input.Runtime
	}
	if input.Genres != nil {
		movie.Genres = input.Genres
	}

	// 4. validatror严重，否则return， baserequest
	v := validator.New()
	if data.ValidateMove(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.FieldErrors)
		return
	}

	// 5. 更新，err则判断err类型，返回对应响应，happy path则write更新后的json
	err = app.models.MovieModel.Update(movie)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJson(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}
