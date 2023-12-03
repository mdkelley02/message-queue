package main

import (
	"encoding/json"
	"net/http"
	"testing"
)

func Test_integration(t *testing.T) {
	server := NewServer(":8080", NewStorage)
	go func() {
		if err := server.Start(); err != nil {
			t.Fatal(err)
		}
	}()

	t.Run("GetTopics", func(t *testing.T) {
		resp, err := http.Get("http://localhost:8080/topics")
		if err != nil {
			t.Fatal(err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}

		var response GetTopicsResponse
		if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
			t.Fatal(err)
		}

		if len(response.Topics) != 0 {
			t.Fatalf("expected 0 topics, got %d", len(response.Topics))
		}
	})
}
