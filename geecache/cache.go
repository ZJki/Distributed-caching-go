package geecache

import (
	"geecache/lru"
	"sync"
)

// cache 用于封装 LRU 缓存。
type cache struct {
	mu         sync.Mutex // 用于保护缓存访问的互斥锁
	lru        *lru.Cache // LRU 缓存实例
	cacheBytes int64      // 缓存总字节数限制
}

// add 将键值对添加到缓存中。
func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		c.lru = lru.New(c.cacheBytes, nil) // 如果缓存为空，创建一个新的 LRU 缓存
	}
	c.lru.Add(key, value) // 向缓存中添加键值对
}

// get 从缓存中获取指定键的值。
func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		return
	}

	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), ok
	}

	return
}
