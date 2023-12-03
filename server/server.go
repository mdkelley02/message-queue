package server

import (
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/mdkelley02/message-queue/storage"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	metrics "github.com/slok/go-http-metrics/metrics/prometheus"
	"github.com/slok/go-http-metrics/middleware"
	"github.com/slok/go-http-metrics/middleware/std"
)

type Server struct {
	serverAddr      string
	metricsAddr     string
	quitChan        chan struct{}
	messages        map[string]chan Message
	router          *mux.Router
	storage         map[string]storage.IStorage
	makeStorageFunc func() storage.IStorage
	upgrader        websocket.Upgrader
}

func NewServer(port string, makeStorageFunc func() storage.IStorage) *Server {
	return &Server{
		metricsAddr:     ":8081",
		serverAddr:      port,
		messages:        make(map[string]chan Message),
		router:          mux.NewRouter(),
		storage:         make(map[string]storage.IStorage),
		makeStorageFunc: makeStorageFunc,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
}

func (s *Server) Start() error {
	mdlw := middleware.New(middleware.Config{
		Recorder: metrics.NewRecorder(metrics.Config{}),
	})
	s.router.Use(std.HandlerProvider("", mdlw))
	s.router.HandleFunc("/topics", s.GetTopicsHandler).Methods(http.MethodGet)
	s.router.HandleFunc("/topics/{topic}", s.PublishHandler).Methods(http.MethodPost)
	s.router.HandleFunc("/topics/{topic}/subscribe", s.SubscribeHandler).Methods(http.MethodGet)

	go func() {
		slog.Info("starting metrics server")
		if err := http.ListenAndServe(s.metricsAddr, promhttp.Handler()); err != nil {
			slog.Info("metrics server failed: %v", err)
		}
	}()

	go func() {
		slog.Info("starting message queue server")
		if err := http.ListenAndServe(s.serverAddr, s.router); err != nil {
			slog.Error("message queue server failed: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	<-sigChan
	return nil
}
