package main

import (
	"github.com/Godyu97/geecache/gee"
	"log"
	"fmt"
	"net/http"
)

func Test1() {
	gee.NewGroup("scores", 2<<10, gee.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))

	addr := "0.0.0.0:9999"
	peers := gee.NewGrpcPool(addr)
	log.Println("geecache is running at", addr)
	log.Fatal(http.ListenAndServe(addr, peers))
}
