#!/bin/bash
# pkill test2.sh才能触发
trap "rm test2;pkill test2" EXIT

go build -o test2
./test2 -port=8001 &
./test2 -port=8002 &
./test2 -port=8003 -api=1 &
sleep 3

echo ">>> start test"
curl "http://localhost:9999/api?key=Tom" &
curl "http://localhost:9999/api?key=Tom" &
curl "http://localhost:9999/api?key=Tom" &
wait
