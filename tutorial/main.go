package main

import (
	"fmt"
	"sync"
)

var (
	counter int = 0
	mutex   sync.Mutex
)

func Increment() {
	mutex.Lock()
	defer mutex.Unlock()

	counter++
}

func Increment1() {
	counter++
}

func main() {
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			if i%2 == 0 {
				Increment()
			}
		}()
	}

	wg.Wait()
	fmt.Println("Final value:", counter)
}
