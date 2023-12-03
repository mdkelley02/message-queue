package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gorilla/websocket"
)

func (s *Server) GetTopicsHandler(w http.ResponseWriter, r *http.Request) {
	topics := make([]string, 0, len(s.subscriptions))
	for topic := range s.subscriptions {
		topics = append(topics, topic)
	}

	response := GetTopicsResponse{
		Topics: topics,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) PublishHandler(w http.ResponseWriter, r *http.Request) {
	// get topic identifier from url
	topic := getTopicFromUrl(r)
	if topic == "" {
		slog.Error("missing topic")
		http.Error(w, "missing topic", http.StatusBadRequest)
		return
	}

	// read request body
	var request PublishRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		slog.Error("could not read request body: %v", err)
		http.Error(w, "could not read request body", http.StatusBadRequest)
		return
	}

	// initialize memory for topic if it doesn't exist
	s.upsertTopic(topic)

	// write message to storage
	offset, err := s.storage[topic].Put(request.Body)
	if err != nil {
		slog.Error("could not write message to storage: %v", err)
		http.Error(w, "could not write message to storage", http.StatusInternalServerError)
		return
	}

	// write message to topic
	go func() {
		s.messages[topic] <- Message{
			Id:     fmt.Sprintf("%s-%d", topic, offset),
			Body:   request.Body,
			Offset: offset,
		}
	}()

	// write response
	response := PublishResponse{
		Offset: offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) SubscribeHandler(w http.ResponseWriter, r *http.Request) {
	// get topic identifier from url
	topic := getTopicFromUrl(r)
	if topic == "" {
		slog.Error("missing topic")
		http.Error(w, "missing topic", http.StatusBadRequest)
		return
	}

	// initialize memory for topic if it doesn't exist
	s.upsertTopic(topic)

	// upgrade connection to websocket
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("could not upgrade connection: %v", err)
		http.Error(w, "could not upgrade connection", http.StatusInternalServerError)
		return
	}

	// add connection to topic
	s.subscriptions[topic] = append(s.subscriptions[topic], conn)

	// remove connection from topic when done
	defer func() {
		for i, c := range s.subscriptions[topic] {
			if c == conn {
				s.subscriptions[topic] = append(s.subscriptions[topic][:i], s.subscriptions[topic][i+1:]...)
				break
			}
		}
	}()

	for msg := range s.messages[topic] {
		jsonBytes, err := json.Marshal(msg)
		if err != nil {
			slog.Error("could not marshal message: %v", err)
			http.Error(w, "could not marshal message", http.StatusInternalServerError)
			return
		}

		if err := conn.WriteMessage(websocket.TextMessage, jsonBytes); err != nil {
			slog.Error("could not write message to connection: %v", err)
			http.Error(w, "could not write message to connection", http.StatusInternalServerError)
			return
		}
	}

	// subscribe to topic
	for {
		// read message from topic
		message, ok := <-s.messages[topic]
		if !ok {
			slog.Error("could not read message from topic")
			http.Error(w, "could not read message from topic", http.StatusInternalServerError)
			return
		}

		fmt.Println("message", message)

		jsonBytes, err := json.Marshal(message)
		if err != nil {
			slog.Error("could not marshal message: %v", err)
			http.Error(w, "could not marshal message", http.StatusInternalServerError)
			return
		}

		// write message to connection
		if err := conn.WriteMessage(websocket.TextMessage, jsonBytes); err != nil {
			slog.Error("could not write message to connection: %v", err)
			http.Error(w, "could not write message to connection", http.StatusInternalServerError)
			return
		}
	}
}

func (s *Server) upsertTopic(topic string) {
	if _, ok := s.subscriptions[topic]; !ok {
		s.subscriptions[topic] = make([]*websocket.Conn, 0)
	}

	if _, ok := s.messages[topic]; !ok {
		s.messages[topic] = make(chan Message)
	}

	if _, ok := s.storage[topic]; !ok {
		s.storage[topic] = s.makeStorageFunc()
	}
}
