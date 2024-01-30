package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// Hash maps bytes to uint32
type Hash func(data []byte) uint32

// Map constains all hashed hvs
type Map struct {
	hash Hash
	//虚拟节点倍数
	replicas int
	//哈希环
	hvs []int // Sorted
	//虚拟节点与真实节点的映射表
	hashMap map[int]string
}

// New creates a Map instance
func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// Add adds some hvs to the hash.
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		//对每一个真实节点 key，对应创建 m.replicas 个虚拟节点，虚拟节点的名称是：strconv.Itoa(i) + key，即通过添加编号的方式区分不同虚拟节点。
		for i := 0; i < m.replicas; i++ {
			//ikey
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.hvs = append(m.hvs, hash)
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.hvs)
}

// Get gets the closest item in the hash to the provided key.
// Get 获取哈希中与所提供的键最接近的项
func (m *Map) Get(key string) string {
	if len(m.hvs) == 0 {
		return ""
	}

	hv := int(m.hash([]byte(key)))
	// 二分查找合适的副本。
	idx := sort.Search(len(m.hvs), func(i int) bool {
		return m.hvs[i] >= hv
	})

	return m.hashMap[m.hvs[idx%len(m.hvs)]]
}
