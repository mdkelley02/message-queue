package storage

import (
	"errors"
	"sync"
)

var ErrNotFound = errors.New("not found")

type IStorage interface {
	Get(offset int) (string, error)
	Put(string) (int, error)
	Delete(offset int) error
}

type Storage struct {
	lock  *sync.Mutex
	store []string
}

func NewStorage() IStorage {
	return &Storage{
		store: make([]string, 0),
		lock:  &sync.Mutex{},
	}
}

func (s *Storage) Get(offset int) (string, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if offset >= len(s.store) {
		return "", ErrNotFound
	}

	return s.store[offset], nil
}

func (s *Storage) Put(data string) (int, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.store = append(s.store, data)
	return len(s.store) - 1, nil
}

func (s *Storage) Delete(offset int) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if offset >= len(s.store) {
		return ErrNotFound
	}

	s.store[offset] = ""
	return nil
}
