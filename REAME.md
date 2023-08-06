# Debug sync.Mutex
Find out which function hold a mutex for more than 10 seconds

```go
import (
	"time"
	"github.com/thanhpk/desync"
)

func holderA(mu *desync.Mutex) {
	mu.Lock()
	time.Sleep(1 * time.Second)
	mu.Unlock()
}

func holderB(mu *desync.Mutex) {
	mu.Lock()
	time.Sleep(15 * time.Second)
	mu.Unlock()
}

func holderC(mu *desync.Mutex) {
	mu.Lock()
	time.Sleep(2 * time.Second)
	mu.Unlock()
}

func TestMutex(t *testing.T) {
	mu := &desync.Mutex{}
	holderA(mu)
	holderB(mu)
	holderC(mu)
}
```
