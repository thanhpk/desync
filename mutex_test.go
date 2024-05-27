package desync

import (
	"testing"
	"time"
)

func long(mu *Mutex) {
	mu.Lock()
	time.Sleep(15 * time.Second)
	mu.Unlock()
}

func quick(mu *Mutex) {
	mu.Lock()
	time.Sleep(2 * time.Second)
	mu.Unlock()
}
func TestMutex(t *testing.T) {
	mu := &Mutex{}
	go quick(mu)
	go long(mu)
	quick(mu)
	long(mu)
}
