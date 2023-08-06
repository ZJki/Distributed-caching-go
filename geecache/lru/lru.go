package lru

import "container/list"

// Cache 是一个 LRU 缓存。它不适用于并发访问。
type Cache struct {
	maxBytes int64                    // 缓存最大总字节数限制
	nbytes   int64                    // 当前已使用的字节数
	ll       *list.List               // 双向链表用于维护缓存顺序
	cache    map[string]*list.Element // 字典用于快速查找缓存项
	// 当一个条目被移除时，可以执行此操作（可选）
	OnEvicted func(key string, value Value)
}

// entry 代表缓存中的一个条目
type entry struct {
	key   string
	value Value
}

// Value 用于表示缓存值，并使用 Len 方法来计算字节数
type Value interface {
	Len() int
}

// New 是 Cache 的构造函数
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// Add 向缓存中添加一个值。
func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele) // 移动现有的缓存项到链表头部
		kv := ele.Value.(*entry)
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value // 更新缓存值
	} else {
		ele := c.ll.PushFront(&entry{key, value}) // 在链表头部添加新的缓存项
		c.cache[key] = ele
		c.nbytes += int64(len(key)) + int64(value.Len())
	}
	// 如果缓存总字节数超过限制，移除最老的缓存项
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}

// Get 查找指定键的值
func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele) // 移动现有的缓存项到链表头部
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}

// RemoveOldest 移除最老的缓存项
func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
		// 如果有设置 OnEvicted 回调函数，调用它
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

// Len 返回缓存中的条目数
func (c *Cache) Len() int {
	return c.ll.Len()
}
