package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/yaninyzwitty/merch-crud-microservice-go/application"
)

func main() {
	app := application.New()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM)
	defer cancel()

	err := app.Start(ctx)

	if err != nil {
		fmt.Println("failed to start application", err)

	}

}
