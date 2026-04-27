package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/kweaver-ai/kweaver-core/trace-ai/agent-observability/src/boot"
)

// @title Agent Observability API
// @version 1.0
// @description APIs for querying agent traces from OpenSearch.
// @BasePath /api/agent-observability/v1
func main() {
	app := boot.NewApp()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := app.Shutdown(shutdownCtx); err != nil {
			log.Printf("shutdown app: %v", err)
		}
	}()

	log.Printf("agent-observability listening on :8080")
	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
