package main

import (
	"log"
	"math/rand"
	"time"
)

func init() {
	rand.NewSource(time.Now().UnixNano())
}

func main() {
	log.Println("Load Balancing")
}
