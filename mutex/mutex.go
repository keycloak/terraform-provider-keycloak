package mutex

import (
	"log"
	"sync"
)

// KeyValue is a simple key/value
// store for arbitrary mutexes.
type KeyValue struct {
	lock  sync.Mutex
	store map[string]*sync.Mutex
}

// New returns a properly initialized KeyValue mutex.
func New() *KeyValue {
	return &KeyValue{
		store: make(map[string]*sync.Mutex),
	}
}

// Lock the mutex for the given key.
func (m *KeyValue) Lock(key string) {
	log.Printf("[DEBUG] Locking mutex for key: %q", key)
	defer log.Printf("[DEBUG] Locked mutex for key: %q", key)

	m.get(key).Lock()
}

// Unlock the mutex for the given key.
func (m *KeyValue) Unlock(key string) {
	log.Printf("[DEBUG] Unlocking mutex for key: %q", key)
	defer log.Printf("[DEBUG] Unlocked mutex for key: %q", key)

	m.get(key).Unlock()
}

// Returns a mutex for the given key.
func (m *KeyValue) get(key string) *sync.Mutex {
	m.lock.Lock()
	defer m.lock.Unlock()

	mutex, ok := m.store[key]
	if !ok {
		mutex = &sync.Mutex{}
		m.store[key] = mutex
	}

	return mutex
}
