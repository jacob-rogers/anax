package cache

import (
	"sync"
)

type SimpleMapCache struct {
	Maplock sync.Mutex
	cache   map[string]interface{}
}

// Return the cached object by input key.
func (c *SimpleMapCache) Get(key string) interface{} {
	c.Maplock.Lock()
	defer c.Maplock.Unlock()

	// Create the cache if we need to.
	c.createCache()

	// Return the specified cache entry if it exists.
	if obj, ok := c.cache[key]; !ok {
		return nil
	} else {
		return obj
	}
}

// GetKeys returns a slice containing all map keys in the cache
func (c *SimpleMapCache) GetKeys() []string {
	c.Maplock.Lock()
	defer c.Maplock.Unlock()

	keys := []string{}

	for key := range c.cache {
		keys = append(keys, key)
	}

	return keys
}

// Store the cached object by input key.
func (c *SimpleMapCache) Put(key string, obj interface{}) {
	c.Maplock.Lock()
	defer c.Maplock.Unlock()

	// Create the cache if we need to.
	c.createCache()

	// Store the specified obj with the given key.
	c.cache[key] = obj
}

func (c *SimpleMapCache) Delete(key string) {
	c.Maplock.Lock()
	defer c.Maplock.Unlock()

	if c.cache != nil {
		delete(c.cache, key)
	}
}

// Create the internal map if necessary. This function assumes that the caller already holds the cache lock.
func (c *SimpleMapCache) createCache() {
	if c.cache == nil {
		c.cache = make(map[string]interface{})
	}
}
