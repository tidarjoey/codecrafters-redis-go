package store

import (
	"errors"
	"sync"
	"time"
)

type Type int

const (
	TypeString Type = iota
	TypeList
)

var ErrWrongType = errors.New("WRONGTYPE Operation against a key holding the wrong kind of value")

type value struct {
	typ    Type
	str    string
	list   []string
	expiry time.Time // zero means no expiry
}

func (v *value) expired() bool {
	return !v.expiry.IsZero() && time.Now().After(v.expiry)
}

type Store struct {
	mu   sync.RWMutex
	data map[string]*value
}

func New() *Store {
	return &Store{data: make(map[string]*value)}
}

func (s *Store) Set(key, str string, expiry time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = &value{typ: TypeString, str: str, expiry: expiry}
}

func (s *Store) Get(key string) (string, bool, error) {
	s.mu.RLock()
	v, ok := s.data[key]
	s.mu.RUnlock()
	if !ok {
		return "", false, nil
	}
	if v.expired() {
		s.mu.Lock()
		delete(s.data, key)
		s.mu.Unlock()
		return "", false, nil
	}
	if v.typ != TypeString {
		return "", false, ErrWrongType
	}
	return v.str, true, nil
}

func (s *Store) RPush(key string, elements ...string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.data[key]
	if ok && !v.expired() && v.typ != TypeList {
		return 0, ErrWrongType
	}
	if !ok || v.expired() {
		s.data[key] = &value{typ: TypeList}
		v = s.data[key]
	}
	v.list = append(v.list, elements...)
	return len(v.list), nil
}
