package gee

import (
	"fmt"
	"github.com/Godyu97/geecache/lru"
	"github.com/Godyu97/geecache/singleflight"
	"log"
	"sync"
	"github.com/Godyu97/geecache/pb"
	"context"
)

type cache struct {
	l          sync.Mutex
	lru        *lru.Cache
	cacheBytes int64
}

func (c *cache) add(key string, v ByteView) {
	c.l.Lock()
	defer c.l.Unlock()
	if c.lru == nil {
		//延迟初始化(Lazy Initialization)意味着该对象的创建将会延迟至第一次使用该对象时。主要用于提高性能，并减少程序内存要求。
		c.lru = lru.New(c.cacheBytes, nil)
	}
	c.lru.Add(key, v)
}

func (c *cache) get(key string) (ByteView, bool) {
	c.l.Lock()
	defer c.l.Unlock()
	if c.lru == nil {
		return ByteView{}, false
	}
	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), ok
	}
	return ByteView{}, false
}

// Getter 如果缓存不存在，应从数据源（文件，数据库等）获取数据并添加到缓存中。
type Getter interface {
	Get(key string) ([]byte, error)
}

// 定义一个函数类型 F，并且实现接口 A 的方法，然后在这个方法中调用自己。这是 Go 语言中将其他函数（参数返回值定义与 F 一致）转换为接口 A 的常用技巧。
type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

// Group
type Group struct {
	name string
	//缓存未命中时获取源数据的回调(callback)
	getter Getter
	//并发安全的单机缓存
	mainCache cache
	//分布式节点选择
	peers  PeerPicker
	loader *singleflight.Group
}

// namespaces
var (
	gl     sync.RWMutex
	groups = make(map[string]*Group)
)

func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("seOoWKPh nil Getter")
	}
	gl.Lock()
	defer gl.Unlock()
	g := &Group{
		name:   name,
		getter: getter,
		mainCache: cache{
			cacheBytes: cacheBytes,
		},
		loader: new(singleflight.Group),
	}
	groups[name] = g
	return g
}

func GetGroup(name string) *Group {
	gl.RLock()
	defer gl.RUnlock()
	return groups[name]
}

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("RFuYlUiN key is required")
	}
	if v, ok := g.mainCache.get(key); ok {
		return v, nil
	}
	return g.load(key)
}

func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	log.Println("getFromPeer ", peer)
	req := &pb.Request{
		Group: g.name,
		Key:   key,
	}
	res, err := peer.Get(context.TODO(), req)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: res.Value}, nil
}

func (g *Group) load(key string) (ByteView, error) {
	v, err := g.loader.Do(key, func() (any, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if v, err := g.getFromPeer(peer, key); err != nil {
					log.Println("[GeeCache] Failed to get from peer", err)
				} else {
					return v, nil
				}
			} else {
				log.Println("[GeeCache] Failed PickPeer ")
			}
		}
		return g.getLocally(key)
	})
	if err != nil {
		return ByteView{}, err
	}
	return v.(ByteView), nil
}

func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err

	}
	v := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, v)
	return v, nil
}

func (g *Group) populateCache(key string, v ByteView) {
	g.mainCache.add(key, v)
}
