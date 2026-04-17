package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/TrienThongLu/goCache/internal/server"
)

func main() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	var wg sync.WaitGroup
	wg.Add(2)

	s := server.NewServer()
	go s.Run(&wg)
	go server.WaitForSignal(&wg, signals)

	wg.Wait()
}
