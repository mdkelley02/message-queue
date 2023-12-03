package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_storage(t *testing.T) {
	storage := NewStorage()
	messageBody := `
		{
			"body": "MY_MESSAGE_1"
		}
	`
	offset, err := storage.Put(messageBody)
	if err != nil {
		t.Fatal(err)
	}

	item, err := storage.Get(offset)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, messageBody, item)
}
