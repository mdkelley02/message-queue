package server

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func getTopicFromUrl(r *http.Request) string {
	vars := mux.Vars(r)
	return vars["topic"]
}

func (s *Server) upsertTopic(topic string) {
	if _, ok := s.messages[topic]; !ok {
		s.messages[topic] = make(chan Message)
	}

	if _, ok := s.storage[topic]; !ok {
		s.storage[topic] = s.makeStorageFunc()
	}
}

func (s *Server) publishMessage(topic string, req PublishRequest) (PublishResponse, error) {
	// create topic if it doesn't exist
	s.upsertTopic(topic)

	// write message to storage
	offset, err := s.storage[topic].Put(req.Body)
	if err != nil {
		return PublishResponse{}, err
	}

	msg := Message{
		Id:     fmt.Sprintf("%s-%d", topic, offset),
		Offset: offset,
	}

	// send message to subscribers
	go func() {
		s.messages[topic] <- msg
	}()

	return PublishResponse{
		Offset:    msg.Offset,
		MessageId: msg.Id,
	}, nil
}
