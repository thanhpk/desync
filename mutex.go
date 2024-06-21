package desync

import (
	"fmt"
	"math/rand"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz")

func generateID() string {
	var sb strings.Builder
	nowstr := strconv.FormatInt(time.Now().UnixNano(), 26)
	sb.WriteString(nowstr)
	for i := 0; i < 20; i++ {
		sb.WriteRune(letterRunes[rand.Intn(len(letterRunes))])
	}
	return sb.String()
}

var g_lock = &sync.Mutex{}
var g_holdingLockM = map[string]int64{}
var g_holdingStackM = map[string]string{}

func init() {
	go func() {
		for {
			time.Sleep(5 * time.Second)
			tooLooong := []string{}
			g_lock.Lock()
			now := time.Now().Unix()
			for id, lockAtSec := range g_holdingLockM {
				if now-lockAtSec > 10 {
					stack := g_holdingStackM[id]
					tooLooong = append(tooLooong, strconv.Itoa(int(now-lockAtSec))+" sec:"+stack)
				}
			}
			g_lock.Unlock()
			for _, l := range tooLooong {
				fmt.Println("ERR LOCK HOLD MORE THAN", l)
			}
		}
	}()

	// clean map
	go func() {
		for {
			time.Sleep(10 * time.Minute)
			var new_holdingLockM = map[string]int64{}
			var new_holdingStackM = map[string]string{}
			g_lock.Lock()
			for id, lockedAt := range g_holdingLockM {
				if lockedAt <= 0 {
					continue
				}
				new_holdingLockM[id] = lockedAt
				new_holdingStackM[id] = g_holdingStackM[id]
			}
			g_holdingStackM = new_holdingStackM
			g_holdingLockM = new_holdingLockM
			g_lock.Unlock()
		}
	}()
}

type Mutex struct {
	mux sync.Mutex
	id  string // internal identification. Generated after the first lock
}

func (me *Mutex) Lock() {
	stack := getStack(3)

	me.mux.Lock()
	var lockId = me.id
	if me.id == "" {
		lockId = generateID()
		me.id = lockId
	}
	now := time.Now().Unix()

	g_lock.Lock()
	g_holdingStackM[lockId] = stack
	g_holdingLockM[lockId] = now
	g_lock.Unlock()
}

func (me *Mutex) Unlock() {
	g_lock.Lock()
	g_holdingLockM[me.id] = 0
	g_lock.Unlock()
	me.mux.Unlock()
}

// getStack returns 10 closest stacktrace, included file paths and line numbers
// it will ignore all system path, path which is vendor is striped to /vendor/
// skip: number of stack ignored
func getStack(skip int) string {
	stack := make([]uintptr, 10)
	var sb strings.Builder
	// skip one system stack, the this current stack line
	length := runtime.Callers(skip, stack[:])
	first := -1
	for i := 0; i < length; i++ {
		pc := stack[i]
		// pc - 1 because the program counters we use are usually return addresses,
		// and we want to show the line that corresponds to the function call
		function := runtime.FuncForPC(pc - 1)
		file, line := function.FileLine(pc - 1)
		if first == -1 {
			first = i
		} else {
			sb.WriteString(";")
		}
		sb.WriteString(file + ":" + strconv.Itoa(line))
	}
	return sb.String()
}
