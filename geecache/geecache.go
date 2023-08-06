package geecache

import (
	"fmt"
	pb "geecache/geecachepb"
	"geecache/singleflight"
	"log"
	"sync"
)

// Group 是一个缓存命名空间，包含主缓存、数据加载器以及对等节点等信息。
type Group struct {
	name      string
	getter    Getter
	mainCache cache
	peers     PeerPicker
	loader    *singleflight.Group
}

// Getter 定义了获取数据的接口。
type Getter interface {
	Get(key string) ([]byte, error)
}

// GetterFunc 是一个实现 Getter 接口的函数类型。
type GetterFunc func(key string) ([]byte, error)

// Get 实现了 Getter 接口中的函数。
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

// NewGroup 创建一个新的 Group 实例。
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
		loader:    &singleflight.Group{},
	}
	groups[name] = g
	return g
}

// GetGroup 返回之前用 NewGroup 创建的命名组，如果不存在该组则返回 nil。
func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

// Get 从缓存中获取指定键的值。
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	if v, ok := g.mainCache.get(key); ok {
		log.Println("[GeeCache] hit")
		return v, nil
	}

	return g.load(key)
}

// RegisterPeers 注册一个 PeerPicker 用于选择远程对等节点。
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

// load 从本地缓存或对等节点加载数据。
func (g *Group) load(key string) (value ByteView, err error) {
	// 使用 singleflight.Group 确保同一时刻只有一个请求从外部加载数据。
	viewi, err := g.loader.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			// 如果存在对等节点，则尝试从对等节点获取数据。
			if peer, ok := g.peers.PickPeer(key); ok {
				// 通过对等节点获取数据，并在成功获取后返回。
				if value, err = g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
				log.Println("[GeeCache] 从对等节点获取失败", err)
			}
		}

		// 从本地获取数据，如果未找到则通过 getter 获取。
		return g.getLocally(key)
	})

	if err == nil {
		return viewi.(ByteView), nil
	}
	return
}

// populateCache 将数据填充到主缓存中。
func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}

// getLocally 从本地获取数据，如果未找到则通过 getter 获取。
func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	// 将获取到的数据存入主缓存，并返回数据视图。
	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

// getFromPeer 从指定的对等节点获取数据。
func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	// 构造请求对象，包含命名组和键。
	req := &pb.Request{
		Group: g.name,
		Key:   key,
	}
	// 创建响应对象。
	res := &pb.Response{}
	// 通过 PeerGetter 接口从对等节点获取数据。
	err := peer.Get(req, res)
	if err != nil {
		return ByteView{}, err
	}
	// 返回从对等节点获取的数据视图。
	return ByteView{b: res.Value}, nil
}
