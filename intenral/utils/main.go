package utils

import (
	"math/rand"
)

func GetRandNumber() int {
	return rand.Intn(9000) + 1000
}
