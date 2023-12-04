package storage

import (
	"errors"
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
	t.Run("read previously put message", func(t *testing.T) {
		offset, err := storage.Put(messageBody)
		if err != nil {
			t.Fatal(err)
		}

		item, err := storage.Get(offset)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, messageBody, item)
	})

	t.Run("delete previously written message", func(t *testing.T) {
		offset, err := storage.Put(messageBody)
		if err != nil {
			t.Fatal(err)
		}

		item, err := storage.Get(offset)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, messageBody, item)

		err = storage.Delete(offset)
		if err != nil {
			t.Fatal(err)
		}

		item, err = storage.Get(offset)
		if !errors.Is(err, ErrNotFound) {
			t.Fatal(err)
		}
		assert.Equal(t, "", item)
	})
}
