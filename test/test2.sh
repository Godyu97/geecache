#!/bin/bash
# pkill test2.sh才能触发
trap "rm test2;pkill test2" EXIT

go build -o test2
./test2 -port=8001 &
./test2 -port=8002  -api=1&
./test2 -port=8003 &
sleep 3

echo ">>> start test"
for i in $(seq 1 5);do
    curl "http://localhost:9999/api?key=Jack" &
    curl "http://localhost:9999/api?key=Jack" &
    curl "http://localhost:9999/api?key=Jack" &
    sleep 1
done

wait
