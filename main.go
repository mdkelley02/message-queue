package main

import (
	"log/slog"

	"github.com/mdkelley02/message-queue/server"
	"github.com/mdkelley02/message-queue/storage"
)

func main() {
	slog.Info("Starting Message Queue")

	s := server.NewServer(server.ServerConfig{
		ServerAddr:      ":8080",
		MetricsAddr:     ":8081",
		MakeStorageFunc: storage.NewStorage,
	})
	if err := s.Start(); err != nil {
		slog.Error("Could not start server: %v", err)
	}

	slog.Info("Stopping Message Queue")
}
