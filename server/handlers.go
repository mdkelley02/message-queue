package server

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
)

func (s *Server) GetTopicsHandler(w http.ResponseWriter, r *http.Request) {
	topics := make([]string, 0, len(s.messages))
	for topic := range s.messages {
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

	slog.Info(fmt.Sprintf("request: %v", request))

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

	// upgrade connection to websocket
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("could not upgrade connection: %v", err)
		http.Error(w, "could not upgrade connection", http.StatusInternalServerError)
		return
	}

	// initialize memory for topic if it doesn't exist
	s.upsertTopic(topic)

	// subscribe to topic
	for {
		// read message from topic
		message, ok := <-s.messages[topic]
		if !ok {
			slog.Error("could not read message from topic")
			http.Error(w, "could not read message from topic", http.StatusInternalServerError)
			return
		}

		value, err := s.storage[topic].Get(message.Offset)
		if err != nil {
			slog.Error("could not read message from storage: %v", err)
			http.Error(w, "could not read message from storage", http.StatusInternalServerError)
		}

		// delete message from storage
		if err := s.storage[topic].Delete(message.Offset); err != nil {
			slog.Error("could not delete message from storage: %v", err)
			http.Error(w, "could not delete message from storage", http.StatusInternalServerError)
			return
		}

		slog.Info(fmt.Sprintf("value: %s", value))

		subMessage := SubscriptionMessage{
			Topic:     topic,
			MessageId: message.Id,
			Value:     value,
		}

		// write message to connection
		if err := conn.WriteJSON(subMessage); err != nil {
			slog.Error("could not write message to connection: %v", err)
			http.Error(w, "could not write message to connection", http.StatusInternalServerError)
			return
		}
	}
}

func (s *Server) upsertTopic(topic string) {
	if _, ok := s.messages[topic]; !ok {
		s.messages[topic] = make(chan Message)
	}

	if _, ok := s.storage[topic]; !ok {
		s.storage[topic] = s.makeStorageFunc()
	}
}

func getTopicFromUrl(r *http.Request) string {
	vars := mux.Vars(r)
	return vars["topic"]
}
