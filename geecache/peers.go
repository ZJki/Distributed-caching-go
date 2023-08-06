package geecache

import pb "geecache/geecachepb"

// PeerPicker 是一个接口，必须被实现，用于定位拥有特定键的对等节点（peer）。
type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

// PeerGetter 是一个接口，必须被对等节点（peer）实现。
type PeerGetter interface {
	Get(in *pb.Request, out *pb.Response) error
}
