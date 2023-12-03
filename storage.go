package main

import (
	"errors"
	"sync"
)

var ErrNotFound = errors.New("not found")

type IStorage interface {
	Get(offset int) ([]byte, error)
	Put([]byte) (int, error)
	Delete(offset int) error
}

type Storage struct {
	lock  *sync.Mutex
	store [][]byte
}

func NewStorage() IStorage {
	return &Storage{
		store: make([][]byte, 0),
		lock:  &sync.Mutex{},
	}
}

func (s *Storage) Get(offset int) ([]byte, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if offset >= len(s.store) {
		return nil, ErrNotFound
	}

	return s.store[offset], nil
}

func (s *Storage) Put(data []byte) (int, error) {
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

	s.store[offset] = nil
	return nil
}
