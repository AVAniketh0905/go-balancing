package main

import (
	"log"
	"math/rand"
	"time"

	_ "net/http/pprof"
)

func init() {
	rand.NewSource(time.Now().UnixNano())
}

func main() {
	log.Println("Load Balancing")
}
