// Here's an example about how to replicate some concurrency functionality from other programming languages with go's primitives!

package main

import (
	"fmt"
	"sync"
	"time"
)

// python generators
// a function can just return a channel which it serves iterator results on through a goroutine, and then closes.
// If the channel is unbuffered, it's a python 2 synchronous generator.
// if it's buffered, the generator and the program can run simulaniously and it's a python 3 async generator.

func generator() chan int {
	c := make(chan int, 30)

	go func() {
		c <- 1 // "yield 1"
		c <- 2 // etc
		c <- 3
		c <- 4
		c <- 5
		c <- 6
		c <- 7
		c <- 8
		c <- 9
		close(c)
	}()

	return c
}

func main() {
	for v := range generator() {
		fmt.Println(v)
	}
}

// javascript setTimeout, setInterval, clearInterval
// I would do some snazzy es6 async stuff but I don't actually know any of that...

func setTimeout(toRun func(), timeout int) {
	go func() {
		time.Sleep(time.Duration(timeout) * time.Millisecond)
		go toRun()
	}()
}

var timeoutCounter int = 0
var timeoutMap map[int]chan bool = make(map[int]chan bool)
var timeoutMapLock sync.Mutex = sync.Mutex{}

func setInterval(toRun func(), interval int) int {
	timeoutMapLock.Lock()
	defer timeoutMapLock.Unlock()
	quit := make(chan bool)
	key := timeoutCounter
	timeoutCounter++
	timeoutMap[key] = quit

	go func() {
		for {
			select {
			case <-quit:
				break
			case <-time.After(time.Duration(interval) * time.Millisecond):
				go toRun()
			}
		}
	}()

	return key
}

func clearInterval(key int) {
	timeoutMapLock.Lock()
	defer timeoutMapLock.Unlock()
	quit, ok := timeoutMap[key]

	if (ok) {
		quit <- true
		delete(timeoutMap, key)
	}
}
