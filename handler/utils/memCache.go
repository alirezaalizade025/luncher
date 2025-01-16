package utils

import (
	"sync"
	"time"
)

type KeyValueTTL struct {
	Key       string
	Value     any
	TTL       time.Duration // TTL as a duration
	CreatedAt time.Time     // Time when the item was created
}

type Store struct {
	data  map[string]KeyValueTTL
	mutex sync.RWMutex
} 

func MemCache() *Store {
	return &Store{
		data: make(map[string]KeyValueTTL),
	}
}

// Set adds a key-value pair with TTL as a duration.
func (s *Store) Set(key string, value any, ttl time.Duration) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.data[key] = KeyValueTTL{
		Key:       key,
		Value:     value,
		TTL:       ttl,
		CreatedAt: time.Now(),
	}
}

// Get retrieves a value if it exists and hasn't expired.
func (s *Store) Get(key string) (any, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	item, exists := s.data[key]
	if !exists || time.Since(item.CreatedAt) > item.TTL {
		return nil, false
	}
	return item.Value, true
}

// Delete removes a key-value pair.
func (s *Store) Delete(key string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.data, key)
}

// Cleanup removes expired keys periodically.
func (s *Store) Cleanup() {
	for {
		time.Sleep(1 * time.Minute) // Adjust interval as needed
		// now := time.Now()

		s.mutex.Lock()
		for key, item := range s.data {
			if time.Since(item.CreatedAt) > item.TTL {
				delete(s.data, key)
			}
		}
		s.mutex.Unlock()
	}
}
