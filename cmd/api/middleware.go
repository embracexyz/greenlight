package main

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/embracexyz/greenlight/internal/data"
	"github.com/embracexyz/greenlight/internal/validator"
	"golang.org/x/time/rate"
)

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (app *application) rateLimit(next http.Handler) http.Handler {
	// 这部分在初始化执行一次，所以limiter或者锁都是全局的
	// 通过闭包方式被引用
	// limiter := rate.NewLimiter(2, 4) // 初始4个token消耗额，一下子消耗完毕后，1/2s补充一个，即1s最多请求2次（连续请求时），且过一段时间不访问，最到能补充到4，如此循环
	// 全局变量
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	if app.config.limiter.enabled {
		go func() {
			for {
				time.Sleep(time.Minute)
				mu.Lock()
				for ip, client := range clients {
					if time.Since(client.lastSeen) > time.Minute*3 {
						delete(clients, ip)
					}
				}
				mu.Unlock()
			}
		}()
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 这里开始是每个reqeust过来，启动一个goroutine处理请求，从一个handler开始直到业务handler都在一个goroutine里（期间不启动其他goroutine的话）
		//		中间件、业务handler、router，都是handler，都实现了ServeHTTP方法，都会被在多个goroutine执行
		// 所以要注意data race场景，
		if app.config.limiter.enabled {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}

			// 访问同一个map，加锁
			mu.Lock()
			if _, ok := clients[ip]; !ok {
				// 新客户端添加一个limiter
				clients[ip] = &client{limiter: rate.NewLimiter(rate.Limit(app.config.limiter.rps), app.config.limiter.burst)}
			}

			clients[ip].lastSeen = time.Now()
			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				app.rateLimmitExceededResponse(w, r)
				return
			}
			mu.Unlock()
		}
		next.ServeHTTP(w, r)
	})
}

// 将用户信息attach到context，即认证为（合法用户or匿名用户）；后续通过user的信息再进行鉴权中间件处理
func (app *application) authentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Authorization")
		// 如果没带就是匿名用户, attach 到context，供工具中间件、或者业务handler处理所需
		authorizationHeader := r.Header.Get("Authorization")
		if authorizationHeader == "" {
			// 更新新的reqeust
			r = app.setContextUser(r, data.AnoymousUser)
			next.ServeHTTP(w, r)
			return
		}

		// 带了就判断是否合法、能否找到用户，找到就把user attach到context上去
		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}
		auth := headerParts[1]

		v := validator.New()
		if data.ValidatorToken(v, auth); !v.Valid() {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		// 根据token查询用户

		user, err := app.models.UserModel.GetForToken(data.ScopeAuthentication, auth)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.invalidAuthenticationTokenResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}

		r = app.setContextUser(r, user)
		next.ServeHTTP(w, r)
	})
}

// 上面只是进行了信息查询、user attach；并未对认证后的用户进行鉴权

// 鉴权：需要登录过(非匿名账户); 被其包裹的中间件，只需是登录用户即可
func (app *application) authenticatedRequired(next http.HandlerFunc) http.HandlerFunc {
	// 注意这里是http.HandlerFunc
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.getContextUser(r)
		if user.IsAnonymous() {
			app.authenticationRequiredResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// 鉴权：需要登录过、且是激活账户
func (app *application) authenticatedActivated(next http.HandlerFunc) http.HandlerFunc {
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.getContextUser(r)
		if !user.Activated {
			app.inactiveAccountResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
	return app.authenticatedRequired(fn)
}
