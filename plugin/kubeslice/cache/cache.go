package cache

import (
	"sync"
	"time"

	"github.com/kubeslice/dns/plugin/kubeslice/slice"
)

type EndpointsCache interface {
	GetAll() []slice.Endpoint
	Get(name, slice, namespace string) []slice.Endpoint
	Put(name, slice, namespace string, endpints []slice.Endpoint) error
	Delete(name, slice, namespace string) error
	GetStats() CacheStats
}

type CacheStats struct {
	Hits       int64
	Misses     int64
	Puts       int64
	Deletes    int64
	LastAccess time.Time
	TotalKeys  int
}

const SEP = "|"

func CacheKey(name, slice, namespace string) string {
	return name + SEP + slice + SEP + namespace
}

// Implement EndpointsCache
type endpointsCache struct {
	cache map[string][]slice.Endpoint
	mutex sync.RWMutex
	stats CacheStats
}

func NewEndpointsCache() *endpointsCache {
	return &endpointsCache{
		cache: make(map[string][]slice.Endpoint),
		stats: CacheStats{},
	}
}

func (c *endpointsCache) GetAll() []slice.Endpoint {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	c.stats.LastAccess = time.Now()
	c.stats.Hits++

	eps := []slice.Endpoint{}
	for _, ep := range c.cache {
		eps = append(eps, ep...)
	}

	return eps
}

func (c *endpointsCache) Get(name, slice, namespace string) []slice.Endpoint {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	c.stats.LastAccess = time.Now()
	key := CacheKey(name, slice, namespace)

	if endpoints, exists := c.cache[key]; exists {
		c.stats.Hits++
		return endpoints
	}

	c.stats.Misses++
	return nil
}

func (c *endpointsCache) Put(name, slice, namespace string, eps []slice.Endpoint) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.stats.LastAccess = time.Now()
	c.stats.Puts++

	key := CacheKey(name, slice, namespace)
	c.cache[key] = eps
	c.stats.TotalKeys = len(c.cache)

	return nil
}

func (c *endpointsCache) Delete(name, slice, namespace string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.stats.LastAccess = time.Now()
	c.stats.Deletes++

	key := CacheKey(name, slice, namespace)
	delete(c.cache, key)
	c.stats.TotalKeys = len(c.cache)

	return nil
}

func (c *endpointsCache) GetStats() CacheStats {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	stats := c.stats
	stats.TotalKeys = len(c.cache)
	return stats
}
