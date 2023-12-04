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
	rwLock *sync.RWMutex
	store  []string
}

func NewStorage() IStorage {
	return &Storage{
		store:  make([]string, 0),
		rwLock: &sync.RWMutex{},
	}
}

func (s *Storage) Get(offset int) (string, error) {
	s.rwLock.RLock()
	defer s.rwLock.RUnlock()

	if offset >= len(s.store) || offset < 0 {
		return "", ErrNotFound
	}

	value := s.store[offset]
	if value == "" {
		return "", ErrNotFound
	}

	return value, nil
}

func (s *Storage) Put(data string) (int, error) {
	s.rwLock.Lock()
	defer s.rwLock.Unlock()

	s.store = append(s.store, data)
	return len(s.store) - 1, nil
}

func (s *Storage) Delete(offset int) error {
	s.rwLock.Lock()
	defer s.rwLock.Unlock()

	if offset >= len(s.store) {
		return ErrNotFound
	}

	s.store[offset] = ""
	return nil
}
