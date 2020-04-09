package fio

import (
	"math/rand"
	"time"
)

func GenerateRandomBytes(n uint64) ([]byte, error) {
	data := make([]byte, n)

	rand.Seed(time.Now().Unix())

	_, err := rand.Read(data)
	if err != nil {
		return nil, err
	}
	return data, nil
}
