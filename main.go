package main

import "log/slog"

func main() {
	slog.Info("Starting Message Queue")
	server := NewServer(":8080", NewStorage)
	if err := server.Start(); err != nil {
		slog.Error("Could not start server: %v", err)
	}
	slog.Info("Stopping Message Queue")
}
