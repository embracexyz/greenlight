package main

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/embracexyz/greenlight/internal/data"
	"github.com/embracexyz/greenlight/internal/validator"
	"github.com/pascaldekloe/jwt"
)

func (app *application) createAuthenticationTokenHandler(w http.ResponseWriter, r *http.Request) {
	// 验证请求体反序列化
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	err := app.readJson(w, r, &input)
	if err != nil {
		app.badRequestErrorReponse(w, r, err)
		return
	}

	// 验证请求体valid
	v := validator.New()
	data.ValidEmail(v, input.Email)
	data.ValidPasswordPlaintext(v, input.Password)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.FieldErrors)
		return
	}

	// 查询用户
	user, err := app.models.UserModel.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidCredentialsResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// 验证密码
	matches, err := user.Password.Matches(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	if !matches {
		app.invalidCredentialsResponse(w, r)
		return
	}

	// 使用jwt token
	var claims jwt.Claims
	claims.Subject = strconv.FormatInt(user.ID, 10)
	claims.Issued = jwt.NewNumericTime(time.Now())
	claims.NotBefore = jwt.NewNumericTime(time.Now())
	claims.Expires = jwt.NewNumericTime(time.Now().Add(24 * time.Hour))
	claims.Issuer = "greenlight.xyz"
	claims.Audiences = []string{"greenlight.xyz"}
	token, err := claims.HMACSign(jwt.HS256, []byte(app.config.jwt.secret))
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// // 为用户生成token
	// token, err := app.models.TokenModel.New(user.ID, 24*3*time.Hour, data.ScopeAuthentication)
	// if err != nil {
	// 	app.serverErrorResponse(w, r, err)
	// 	return
	// }

	// 返回token响应
	err = app.writeJson(w, http.StatusCreated, envelope{"authentication_token": string(token)}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) createPasswordResetTokenHandler(w http.ResponseWriter, r *http.Request) {
	// 验证读request
	var input struct {
		Email string `json:"email"`
	}
	err := app.readJson(w, r, &input)
	if err != nil {
		app.badRequestErrorReponse(w, r, err)
		return
	}

	// 验证email字段
	v := validator.New()
	data.ValidEmail(v, input.Email)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.FieldErrors)
		return
	}

	// 查询用户
	user, err := app.models.UserModel.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddFieldError("email", "no matching email found")
			app.failedValidationResponse(w, r, v.FieldErrors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// 判断是否激活用户
	if !user.Activated {
		v.AddFieldError("email", "user account must be activated")
		app.failedValidationResponse(w, r, v.FieldErrors)
		return
	}

	// 是则进行生成一个临时token返回
	token, err := app.models.TokenModel.New(user.ID, 4*time.Hour, data.ScopePasswordReset)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	//
	message := envelope{
		"password_reset_token": token.Plaintext,
	}

	err = app.writeJson(w, http.StatusCreated, message, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) createActivationTokenHandler(w http.ResponseWriter, r *http.Request) {
	// 验证读request
	var input struct {
		Email string `json:"email"`
	}
	err := app.readJson(w, r, &input)
	if err != nil {
		app.badRequestErrorReponse(w, r, err)
		return
	}

	// 验证email
	v := validator.New()
	if data.ValidEmail(v, input.Email); !v.Valid() {
		app.failedValidationResponse(w, r, v.FieldErrors)
		return
	}

	// 查询用户
	user, err := app.models.UserModel.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddFieldError("email", "no matching email found")
			app.failedValidationResponse(w, r, v.FieldErrors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// 判断是否激活用户
	if user.Activated {
		v.AddFieldError("email", "user already activated")
		app.failedValidationResponse(w, r, v.FieldErrors)
		return
	}

	// 是则进行生成一个临时token返回
	token, err := app.models.TokenModel.New(user.ID, 3*24*time.Hour, data.ScopeActivation)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	message := envelope{
		"activation_token": token.Plaintext,
	}

	err = app.writeJson(w, http.StatusCreated, message, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
