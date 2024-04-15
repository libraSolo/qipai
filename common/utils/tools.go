package utils

import (
	"golang.org/x/exp/rand"
	"time"
)

func Contains[T string | int](list []T, target T) bool {
	for _, v := range list {
		if v == target {
			return true
		}
	}
	return false
}

func Rand(n int) int {
	rand.Seed(uint64(time.Now().Unix()))
	return rand.Intn(n)
}
