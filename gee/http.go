package gee

import (
	"fmt"
	"github.com/Godyu97/geecache/consistenthash"
	"log"
	"sync"
	"context"
	"github.com/Godyu97/geecache/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net/http"
	"strings"
)

const (
	defaultBasePath = ""
	defaultReplicas = 50
)

type GrpcPool struct {
	//example.net:8000
	host     string
	basePath string
	l        sync.Mutex
	peers    *consistenthash.Map
	grpcGets map[string]*grpcGet
	pb.UnimplementedGroupCacheServer
}

func NewGrpcPool(host string) *GrpcPool {
	return &GrpcPool{
		host:     host,
		basePath: defaultBasePath,
	}
}

// Log info with server name
func (p *GrpcPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.host, fmt.Sprintf(format, v...))
}

// ServeHTTP handle all http requests
func (p *GrpcPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("GrpcPool serving unexpected path: " + r.URL.Path)
	}
	p.Log("%s %s", r.Method, r.URL.Path)
	// /<basepath>/<groupname>/<key> required
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	key := parts[1]

	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}

	v, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(v.ByteSlice())
}

func (p *GrpcPool) Get(ctx context.Context, in *pb.Request) (out *pb.Response, err error) {
	p.Log("rpc Get %s %s", in.GetGroup(), in.GetKey())
	groupName := in.GetGroup()
	key := in.GetKey()
	group := GetGroup(groupName)
	if group == nil {
		return nil, fmt.Errorf("no such group: %s" + groupName)
	}

	v, err := group.Get(key)
	if err != nil {
		return nil, err
	}
	out = &pb.Response{Value: v.ByteSlice()}
	return out, nil
}

func (p *GrpcPool) Set(peers ...string) {
	p.l.Lock()
	defer p.l.Unlock()
	p.peers = consistenthash.New(defaultReplicas, nil)
	p.peers.Add(peers...)
	p.grpcGets = make(map[string]*grpcGet, len(peers))
	for _, peer := range peers {
		p.grpcGets[peer] = &grpcGet{
			baseURL: peer + p.basePath,
		}
	}
}

func (p *GrpcPool) PickPeer(key string) (PeerGetter, bool) {
	p.Log("Host:%s", p.host)
	p.l.Lock()
	defer p.l.Unlock()
	if peer := p.peers.Get(key); peer != "" && peer != p.host {
		p.Log("Pick peer %s", peer)
		return p.grpcGets[peer], true
	}
	return nil, false
}

type grpcGet struct {
	//example.com:8001/_geecache/
	baseURL string
}

func (h *grpcGet) Get(ctx context.Context, in *pb.Request) (out *pb.Response, err error) {

	//u := fmt.Sprintf(
	//	"%s%s/%s",
	//	h.baseURL,
	//	url.QueryEscape(group),
	//	url.QueryEscape(key),
	//)
	//res, err := http.Get(u)
	//if err != nil {
	//	return nil, err
	//}
	//defer res.Body.Close()
	//
	//if res.StatusCode != http.StatusOK {
	//	return nil, fmt.Errorf("server returned:%s", res.Status)
	//}
	//
	//b, err := io.ReadAll(res.Body)
	//if err != nil {
	//	return nil, fmt.Errorf("reading response body: %v", err)
	//}
	//
	//return b, nil

	//改造成rpc
	out, err = pb.NewGroupCacheClient(GetGrpc(h.baseURL)).Get(ctx, in)
	if err != nil {
		return nil, err
	}
	log.Println("节点间rpc 成功", h.baseURL)
	return out, nil
}

//grpc clients
var conns = make(map[string]*grpc.ClientConn)

func GetGrpc(baseURL string) *grpc.ClientConn {
	conn, ok := conns[baseURL]
	if !ok || conn == nil {
		var err error
		conn, err = grpc.Dial(baseURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			panic(err)
		}
		conns[baseURL] = conn
	}
	return conn
}

func ClosePyApis() {
	for i, _ := range conns {
		conns[i].Close()
	}
}
