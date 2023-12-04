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
	sigChan         chan os.Signal
	messages        map[string]chan Message
	router          *mux.Router
	storage         map[string]storage.IStorage
	makeStorageFunc func() storage.IStorage
	upgrader        websocket.Upgrader
	deadLetterTopic string
}

type ServerConfig struct {
	DeadLetterTopic          string
	ServerAddr               string
	MetricsAddr              string
	MakeStorageFunc          func() storage.IStorage
	WebsocketReadBufferSize  int
	WebsocketWriteBufferSize int
}

func NewServer(cfg ServerConfig) *Server {
	s := &Server{
		deadLetterTopic: cfg.DeadLetterTopic,
		sigChan:         make(chan os.Signal, 1),
		metricsAddr:     cfg.MetricsAddr,
		serverAddr:      cfg.ServerAddr,
		messages:        make(map[string]chan Message),
		router:          mux.NewRouter(),
		storage:         make(map[string]storage.IStorage),
		makeStorageFunc: cfg.MakeStorageFunc,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  cfg.WebsocketReadBufferSize,
			WriteBufferSize: cfg.WebsocketWriteBufferSize,
		},
	}

	if s.upgrader.ReadBufferSize == 0 {
		s.upgrader.ReadBufferSize = 1024
	}

	if s.upgrader.WriteBufferSize == 0 {
		s.upgrader.WriteBufferSize = 1024
	}

	return s
}

func (s *Server) Start() error {

	//  if metricsAddr is not empty, start metrics server
	if s.metricsAddr != "" {
		s.router.Use(std.HandlerProvider("", middleware.New(middleware.Config{
			Recorder: metrics.NewRecorder(metrics.Config{}),
		})))

		go func() {
			slog.Info("starting metrics server")
			if err := http.ListenAndServe(s.metricsAddr, promhttp.Handler()); err != nil {
				slog.Info("metrics server failed: %v", err)
			}
		}()
	}

	// initialize routes
	s.router.HandleFunc("/topics", s.GetTopicsHandler).Methods(http.MethodGet)
	s.router.HandleFunc("/topics/{topic}", s.PublishHandler).Methods(http.MethodPost)
	s.router.HandleFunc("/topics/{topic}/subscribe", s.SubscribeHandler).Methods(http.MethodGet)

	// start message queue server
	go func() {
		slog.Info("starting message queue server")
		if err := http.ListenAndServe(s.serverAddr, s.router); err != nil {
			slog.Error("message queue server failed: %v", err)
		}
	}()

	signal.Notify(s.sigChan, syscall.SIGTERM, syscall.SIGINT)
	<-s.sigChan

	slog.Info("shutting down message queue server")

	return nil
}

func (s *Server) Stop() {
	s.sigChan <- syscall.SIGTERM
}
