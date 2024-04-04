package application

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type App struct {
	router  http.Handler
	redisDb *redis.Client
}

func New() *App {

	app := &App{
		redisDb: redis.NewClient(&redis.Options{}),
	}

	app.loadRoutes()

	return app

}

// method to start the application

func (a *App) Start(ctx context.Context) error {

	server := &http.Server{
		Addr:    ":3000",
		Handler: a.router,
	}

	err := a.redisDb.Ping(ctx).Err()
	if err != nil {
		return fmt.Errorf("failed to connect to redis: %w", err)
	}

	ch := make(chan error, 1)

	if err != nil {
		return fmt.Errorf("failed to connect to redis: %w", err)
	}

	defer func() {

		if err := a.redisDb.Close(); err != nil {
			fmt.Printf("error closing redis client: %v", err)
		}

	}()

	go func() {
		err = server.ListenAndServe() //start the server concurrently
		if err != nil {
			ch <- fmt.Errorf("error starting server: %v", err)
			close(ch) //signal that no incoming channel data, coz buffered channel

		}

	}()

	if err != nil {
		return fmt.Errorf("error starting server: %v", err)

	}

	select {

	case err = <-ch:
		return err

	case <-ctx.Done():
		timeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		return server.Shutdown(timeout) //gracefully shutdown the server
	}

}
