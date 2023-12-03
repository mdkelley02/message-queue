package main

import (
	"log/slog"

	"github.com/mdkelley02/message-queue/server"
	"github.com/mdkelley02/message-queue/storage"
)

func main() {
	slog.Info("Starting Message Queue")
	s := server.NewServer(":8080", storage.NewStorage)
	if err := s.Start(); err != nil {
		slog.Error("Could not start server: %v", err)
	}
	slog.Info("Stopping Message Queue")
}
