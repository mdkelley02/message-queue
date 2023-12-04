package client

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/mdkelley02/message-queue/server"
	"github.com/mdkelley02/message-queue/storage"
)

func Test_integration(t *testing.T) {
	s := server.NewServer(server.ServerConfig{
		ServerAddr:      ":8080",
		MakeStorageFunc: storage.NewStorage,
	})
	go func() {
		if err := s.Start(); err != nil {
			panic(err)
		}
	}()

	t.Run("topic is upserted if it does not exist, topic is found in GetTopics response after creation", func(t *testing.T) {
		topic := "MY_TOPIC_1"
		client := NewMessageQueueClient("localhost:8080", false)

		pubResp, err := client.Publish(topic, "MY_MESSAGE_1")
		if err != nil {
			t.Fatal(err)
		}

		if pubResp.Offset != 0 {
			t.Fatalf("expected offset to be 0, got %d", pubResp.Offset)
		}

		topics, err := client.GetTopics()
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, []string{topic}, topics)
	})

	t.Run("message is published to topic and subscriber receives it", func(t *testing.T) {
		topic := "MY_TOPIC_2"
		const msg = "{\"message\":\"MY_MESSAGE_2\"}}"

		client := NewMessageQueueClient("localhost:8080", false)

		_, err := client.Subscribe(topic, func(d server.Delivery) error {
			if d.Value != msg {
				t.Fatalf("expected message to be %s, got %s", msg, d.Value)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}

		if _, err := client.Publish(topic, msg); err != nil {
			t.Fatal(err)
		}

		time.Sleep(100 * time.Millisecond)
	})
}
