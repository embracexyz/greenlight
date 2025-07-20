package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *application) serve() error {
	srv := http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		ErrorLog:     log.New(app.logger, "", 0), // 实现io.Writer接口，借助标准log库包装为log.Logger，从而可以被http.Server使用
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	shutdownErr := make(chan error)
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		// block until a signal received
		s := <-quit
		app.logger.PrintInfo("shutting down server...", map[string]string{
			"signal": s.String(),
		})

		// 开始优雅退出
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		shutdownErr <- srv.Shutdown(ctx)

	}()

	app.logger.PrintInfo("starting %s server on %s", map[string]string{
		"addr": srv.Addr,
		"env":  app.config.env,
	})

	// srv.Shutdown()调用会让srv.ListenAndServe() 立刻返回一个特定err: http.ErrServerClosed（是所期望的正常状态）
	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	// 处理正常退出状态, 判断srv.Shutdown(ctx)是否发生错误
	err = <-shutdownErr
	if err != nil {
		return err
	}
	app.logger.PrintInfo("stoped server.", map[string]string{
		"addr": srv.Addr,
	})
	return nil
}
