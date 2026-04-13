package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	fmt.Println("Process ID:", os.Getpid())
	// 1. Create a channel to receive system signals.
	// This channel will be the asynchronous endpoint for signal notifications.
	// think it as a queue
	sigs := make(chan os.Signal, 1)

	// 2. Register the signals we want to catch.
	// We want to catch SIGINT (interrupt from terminal)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// 3. Create a channel to block the main goroutine until a shutdown is complete.
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		// block
		fmt.Println("[MAIN] Waiting for work or signal...")
		sig := <-sigs
		fmt.Printf("\n\n[HANDLER] Received signal: %v\n", sig)
		// Signal the main goroutine that the handler has finished and it's safe to exit.
		os.Exit(0)
	}()

	wg.Wait()
}
