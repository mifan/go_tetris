package utils

import (
	"crypto/rand"
	"fmt"
	"os"
	"sync"
)

const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func RandString(n int) string {
	if n <= 0 {
		return ""
	}
	var bytes = make([]byte, n)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return string(bytes)
}

func randStringFromFile() (r string, err error) {
	f, err := os.Open("/dev/urandom")
	if err != nil {
		return
	}
	defer f.Close()
	b := make([]byte, 16)
	f.Read(b)
	r = fmt.Sprintf("%x%x%x%x%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	return
}

var randCache = make(map[string]int)
var mu sync.Mutex

const buffer = 1 << 6
const length = 1 << 5

func addRand() {
	mu.Lock()
	defer mu.Unlock()
	if len(randCache) < buffer {
		for i := 0; i < buffer; i++ {
			r := RandString(length)
			randCache[r]++
		}
	}
}

func getRand() string {
	addRand()
	mu.Lock()
	defer mu.Unlock()
	ran, _ := randStringFromFile()
	for r, c := range randCache {
		if c == 1 {
			ran = r
			goto ret
		}
		delete(randCache, r)
	}
ret:
	return ran
}
