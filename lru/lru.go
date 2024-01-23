package lru

import (
	"container/list"
)

// Cache is a LRU cache. It is not safe for concurrent access.
type Cache struct {
	maxBytes  int64
	usedBytes int64
	ll        *list.List
	cache     map[string]*list.Element
	OnEvicted func(key string, v Value)
}

// Value use Len to count how many bytes it takes
type Value interface {
	Len() int64
}

type entry struct {
	key   string
	value Value
}

func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		usedBytes: 0,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// Get look ups a key's value
func (c *Cache) Get(key string) (Value, bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToBack(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return nil, false
}

func (c *Cache) RemoveOldest() {
	ele := c.ll.Front()
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		c.usedBytes = c.usedBytes - (int64(len(kv.key)) + kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

func (c *Cache) Add(key string, v Value) {
	if ele, ok := c.cache[key]; ok {
		//更新key缓存,并将该节点移到队尾
		c.ll.MoveToBack(ele)
		kv := ele.Value.(*entry)
		c.usedBytes = c.usedBytes + v.Len() - kv.value.Len()
		kv.value = v
	} else {
		//新增key ele
		ele := c.ll.PushBack(&entry{
			key:   key,
			value: v,
		})
		c.cache[key] = ele
		c.usedBytes = c.usedBytes + int64(len(key)) + v.Len()
	}
	//可能会淘汰多个节点
	for c.maxBytes != 0 && c.maxBytes < c.usedBytes {
		c.RemoveOldest()
	}
}

func (c *Cache) Len() int {
	return c.ll.Len()
}
