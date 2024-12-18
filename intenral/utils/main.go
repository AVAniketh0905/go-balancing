package utils

import (
	"log"
	"math/rand"
)

func GetRandNumber() int {
	return rand.Intn(9000) + 1000
}

type CyclicSlice[T comparable] struct {
	v    []T
	curr int
}

func (cs *CyclicSlice[T]) IsNill() bool {
	return cs.v == nil
}

func (cs *CyclicSlice[T]) Init(v []T) {
	cs.v = v
	log.Println("from cyclic slice", cs.v)
	cs.curr = 0
}

func (cs *CyclicSlice[T]) Get() T {
	curr := cs.curr
	cs.curr += 1
	if cs.curr >= len(cs.v) {
		cs.curr = 0
	}
	return cs.v[curr]
}
