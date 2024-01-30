package main

import (
	"flag"
	"fmt"
	"github.com/Godyu97/geecache/gee"
	"log"
	"net/http"
	"google.golang.org/grpc"
	"net"
	"github.com/Godyu97/geecache/pb"
)

func createGroup() *gee.Group {
	return gee.NewGroup("scores", 2<<10, gee.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
}

func startCacheServer(addr string, addrs []string, g *gee.Group) {
	//peers := gee.NewHTTPPool(addr)
	//peers.Set(addrs...)
	//g.RegisterPeers(peers)
	//log.Println("gee is running at", addr)
	//log.Fatal(http.ListenAndServe(addr[7:], peers))

	peers := gee.NewGrpcPool(addr)
	peers.Set(addrs...)
	g.RegisterPeers(peers)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Printf("failed to listen: %v", err)
		return
	}
	s := grpc.NewServer()                 // 创建gRPC服务器
	pb.RegisterGroupCacheServer(s, peers) // 在gRPC服务端注册服务
	// 启动服务
	log.Println("gee is running at", addr)
	log.Fatal(s.Serve(lis))
}

func startAPIServer(apiAddr string, g *gee.Group) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			view, err := g.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(view.ByteSlice())

		}))
	log.Println("fontend server is running at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))

}

func Test2() {
	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "gee server port")
	flag.BoolVar(&api, "api", false, "Start a api server?")
	flag.Parse()

	apiAddr := "http://0.0.0.0:9999"
	addrMap := map[int]string{
		8001: "0.0.0.0:8001",
		8002: "0.0.0.0:8002",
		8003: "0.0.0.0:8003",
	}

	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}

	g := createGroup()
	if api {
		go startAPIServer(apiAddr, g)
	}
	startCacheServer(addrMap[port], addrs, g)

	gee.ClosePyApis()
}
