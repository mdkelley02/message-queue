package main

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mdkelley02/message-queue/client"
	"github.com/mdkelley02/message-queue/server"
	"github.com/mdkelley02/message-queue/storage"
)

func Test_integration(t *testing.T) {
	s := server.NewServer(":8080", storage.NewStorage)
	go func() {
		if err := s.Start(); err != nil {
			panic(err)
		}
	}()

	t.Run("topic is upserted if it does not exist", func(t *testing.T) {
		topic := "MY_TOPIC_1"
		client := client.NewMessageQueueClient("localhost:8080")

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
		client := client.NewMessageQueueClient("localhost:8080")
		const msg = "{\"message\":\"MY_MESSAGE_2\"}}"
		_, err := client.Publish(topic, msg)
		if err != nil {
			t.Fatal(err)
		}

		err = client.Subscribe(topic, func(msg server.SubscriptionMessage) {
			assert.Equal(t, msg, msg.Value)
		})
		if err != nil {
			t.Fatal(err)
		}

	})
}
