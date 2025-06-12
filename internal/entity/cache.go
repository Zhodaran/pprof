package entity

import (
	"log"
	"net/http"
	"sync"
	"time"
)

type Server struct {
	http.Server
	cache *Cache
}

type Cache struct {
	data  map[string]interface{}
	mutex sync.RWMutex
	ttl   time.Duration
}

func NewCache(ttl time.Duration) *Cache {
	return &Cache{
		data: make(map[string]interface{}),
		ttl:  ttl,
	}
}

func (c *Cache) Set(key string, value interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.data[key] = value
	time.AfterFunc(c.ttl, func() {
		c.Remove(key)
	})
}

// Get получает значение из кэша по ключу
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	value, exists := c.data[key]
	return value, exists
}

func (s *Server) Serve() {
	log.Println("Starting server...")
	if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: &v", err)
	}
}

func (c *Cache) Remove(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.data, key)
}
