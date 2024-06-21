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

type Mutex struct {
	lastLockId string
	mux        sync.Mutex
}

func (me *Mutex) Lock() {
	lockId := generateID()
	stack := getStack(3)
	me.mux.Lock()
	me.lastLockId = lockId

	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("LOCKRECOVER", stack, r)
			}
		}()
		time.Sleep(10 * time.Second)
		if me == nil {
			fmt.Println("LOCKNIL", stack)
			return
		}

		if me.lastLockId != lockId {
			return
		}
		// still lock
		for i := 0; i < 100; i++ {
			fmt.Println(i, "LOCK HOLD MORE THAN 10 sec", stack)
			time.Sleep(10 * time.Second)
		}
	}()
}

func (me *Mutex) Unlock() {
	me.lastLockId = ""

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
			sb.WriteString(" -> ")
		}
		sb.WriteString(file + ":" + strconv.Itoa(line))
	}
	return sb.String()
}
