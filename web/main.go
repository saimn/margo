package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

func main() {
	fmt.Println("please connect to localhost:7777")
	http.HandleFunc("/", rootHandle)
	log.Fatal(http.ListenAndServe(":7777", nil))
}

func rootHandle(w http.ResponseWriter, r *http.Request) {
	// mu.Lock()
	// defer mu.Unlock()
	clients++
	time.Sleep(100 * time.Millisecond)
	fmt.Fprintf(w, "time:  %v\n", time.Now().UTC())
	fmt.Fprintf(w, "conns: %v\n", clients)
	fmt.Printf("%v - %v\n", time.Now().UTC(), clients)
}

var (
	mu      sync.Mutex
	clients = 0
)
