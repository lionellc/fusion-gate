package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/lionellc/fusion-gate/internal/config"
)

func main() {
	cfg := config.Load("config/config.yaml")

	engine, cleanup, err := wireApp(cfg)
	if err != nil {
		log.Fatalf("Failed to wire app: %v", err)
	}

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)

	go engine.Run(addr)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	log.Println("Shutting down...")
	cleanup()

}
