package gee

import (
	"github.com/Godyu97/geecache/pb"
	"context"
)

type PeerGetter interface {
	//Get(group string, key string) ([]byte, error)
	Get(ctx context.Context, in *pb.Request) (out *pb.Response, err error)
}

type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

var _ PeerPicker = (*GrpcPool)(nil)
