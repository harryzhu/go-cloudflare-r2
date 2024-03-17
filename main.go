package main

import (
	"sync"

	"cfr2"
)

func main() {
	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		cfr2.StartServer()
	}()

	wg.Wait()
}
