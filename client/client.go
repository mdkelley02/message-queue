package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/mdkelley02/message-queue/server"
)

type IMessageQueueClient interface {
	GetTopics() ([]string, error)
	Publish(topic string, message string) (server.PublishResponse, error)
	Subscribe(topic string, callback func(server.Delivery) error) (chan struct{}, error)
}

type MessageQueueClient struct {
	addr              string
	deadLetterEnabled bool
}

func NewMessageQueueClient(addr string, deadLetterEnabled bool) IMessageQueueClient {
	return &MessageQueueClient{
		addr:              addr,
		deadLetterEnabled: deadLetterEnabled,
	}
}

func (c *MessageQueueClient) GetTopics() ([]string, error) {
	resp, err := http.Get(fmt.Sprintf("http://%s/topics", c.addr))
	if err != nil {
		slog.Error("could not get topics: %v", err)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		slog.Error("could not get topics: %v", err)
		return nil, err
	}

	var response server.GetTopicsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		slog.Error("could not get topics: %v", err)
		return nil, err
	}

	return response.Topics, nil
}

func (c *MessageQueueClient) Publish(topic string, message string) (server.PublishResponse, error) {
	request, err := json.Marshal(server.PublishRequest{
		Body: message,
	})
	if err != nil {
		slog.Error("could not marshal request: %v", err)
		return server.PublishResponse{}, err
	}

	resp, err := http.Post(fmt.Sprintf("http://%s/topics/%s", c.addr, topic), "application/json", bytes.NewReader(request))
	if err != nil {
		slog.Error("could not marshal request: %v", err)
		return server.PublishResponse{}, err
	}

	var response server.PublishResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		slog.Error("could not publish message: %v", err)
		return server.PublishResponse{}, err
	}

	if err != nil {
		slog.Error("could not publish message: %v", err)
		return server.PublishResponse{}, err
	}

	return response, nil
}

func (c *MessageQueueClient) Subscribe(topic string, callback func(server.Delivery) error) (chan struct{}, error) {
	return c.subscribeWithConn(topic, func(conn *websocket.Conn) {
		var message server.Delivery
		if err := conn.ReadJSON(&message); err != nil {
			slog.Error("could not read message: %v", err)
			return
		}

		if err := callback(message); err != nil {
			slog.Error("could not process message: %v", err)

			if c.deadLetterEnabled {
				if _, err := c.Publish(fmt.Sprintf("%s.deadletter", topic), message.Value); err != nil {
					slog.Error("could not publish message to dead letter queue: %v", err)
				}
			}
		}
	})
}

func (c *MessageQueueClient) subscribeWithConn(topic string, callback func(*websocket.Conn)) (chan struct{}, error) {
	conn, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("ws://%s/topics/%s/subscribe", c.addr, topic), nil)
	if err != nil {
		slog.Error("could not subscribe: %v", err)
		return nil, err
	}

	quit := make(chan struct{})

	go func() {
		for {
			select {
			case <-quit:
				return
			default:
				callback(conn)
			}
		}
	}()

	return quit, nil
}
