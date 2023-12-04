package server

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

func (s *Server) GetTopicsHandler(w http.ResponseWriter, r *http.Request) {
	response := GetTopicsResponse{
		Topics: make([]string, 0, len(s.messages)),
	}

	for topic := range s.messages {
		response.Topics = append(response.Topics, topic)
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

	// publish message to topic
	publishResp, err := s.publishMessage(topic, request)
	if err != nil {
		slog.Error("could not publish message: %v", err)
		http.Error(w, "could not publish message", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(publishResp)
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

	// create topic if it does not exist
	s.upsertTopic(topic)

	topicStorage := s.storage[topic]

	// subscribe to topic
	for {
		// read message from topic
		message, ok := <-s.messages[topic]
		if !ok {
			slog.Error("could not read message from topic")
			http.Error(w, "could not read message from topic", http.StatusInternalServerError)
			return
		}

		value, err := topicStorage.Get(message.Offset)
		if err != nil {
			slog.Error("could not read message from storage: %v", err)
			http.Error(w, "could not read message from storage", http.StatusInternalServerError)
		}

		// delete message from storage
		if err := topicStorage.Delete(message.Offset); err != nil {
			slog.Error("could not delete message from storage: %v", err)
			http.Error(w, "could not delete message from storage", http.StatusInternalServerError)
			return
		}

		// write message to connection
		if err := conn.WriteJSON(Delivery{
			Topic:     topic,
			MessageId: message.Id,
			Value:     value,
		}); err != nil {
			slog.Error("could not write message to connection: %v", err)
			http.Error(w, "could not write message to connection", http.StatusInternalServerError)
			return
		}
	}
}
