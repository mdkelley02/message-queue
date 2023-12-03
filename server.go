package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type Server struct {
	port            string
	messages        map[string]chan Message
	subscriptions   map[string][]*websocket.Conn
	router          *mux.Router
	storage         map[string]IStorage
	makeStorageFunc func() IStorage
	upgrader        websocket.Upgrader
}

func NewServer(port string, makeStorageFunc func() IStorage) *Server {
	return &Server{
		port:            port,
		messages:        make(map[string]chan Message),
		subscriptions:   make(map[string][]*websocket.Conn),
		router:          mux.NewRouter(),
		storage:         make(map[string]IStorage),
		makeStorageFunc: makeStorageFunc,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
}

func (s *Server) Start() error {
	s.router.HandleFunc("/topics", s.GetTopicsHandler).Methods(http.MethodGet)
	s.router.HandleFunc("/topics/{topic}", s.PublishHandler).Methods(http.MethodPost)
	s.router.HandleFunc("/topics/{topic}/subscribe", s.SubscribeHandler).Methods(http.MethodGet)

	return http.ListenAndServe(s.port, s.router)
}

func getTopicFromUrl(r *http.Request) string {
	vars := mux.Vars(r)
	return vars["topic"]
}
