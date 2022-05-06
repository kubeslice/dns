package cache

import (
	"github.com/kubeslice/dns/plugin/kubeslice/slice"
)

type EndpointsCache interface {
	GetAll() []slice.Endpoint
	Get(name, slice, namespace string) []slice.Endpoint
	Put(name, slice, namespace string, endpints []slice.Endpoint) error
	Delete(name, slice, namespace string) error
}

const SEP = "|"

func CacheKey(name, slice, namespace string) string {
	return name + SEP + slice + SEP + namespace
}

// Implement EndpointsCache
type endpointsCache struct {
	cache map[string][]slice.Endpoint
}

func NewEndpointsCache() *endpointsCache {
	return &endpointsCache{
		cache: make(map[string][]slice.Endpoint),
	}
}

func (c *endpointsCache) GetAll() []slice.Endpoint {

	eps := []slice.Endpoint{}

	for _, ep := range c.cache {
		eps = append(eps, ep...)
	}

	return eps
}

func (c *endpointsCache) Get(name, slice, namespace string) []slice.Endpoint {
	key := CacheKey(name, slice, namespace)
	return c.cache[key]
}

func (c *endpointsCache) Put(name, slice, namespace string, eps []slice.Endpoint) error {
	key := CacheKey(name, slice, namespace)
	c.cache[key] = eps
	return nil
}

func (c *endpointsCache) Delete(name, slice, namespace string) error {
	key := CacheKey(name, slice, namespace)
	delete(c.cache, key)
	return nil
}
