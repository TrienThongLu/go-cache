package server

import (
	"io"
	"log"
	"net"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/TrienThongLu/goCache/internal/config"
	"github.com/TrienThongLu/goCache/internal/constant"
	"github.com/TrienThongLu/goCache/internal/core"
	"github.com/spaolacci/murmur3"
)

var serverStatus int32 = constant.ServerStatusIdle

func WaitForSignal(wg *sync.WaitGroup, signals chan os.Signal) {
	defer wg.Done()

	<-signals

	for {
		if atomic.CompareAndSwapInt32(&serverStatus, constant.ServerStatusIdle, constant.ServerStatusShuttingDown) {
			log.Println("Shutting down gracefully")
			os.Exit(0)
		}
	}
}

type Server struct {
	numWorkers int
	workers    []*core.Worker

	numIOHandlers int
	ioHandlers    []*IOHandler
	nextIOHandler int
}

func NewServer() *Server {
	numCores := runtime.NumCPU()
	numWorkers := numCores / 2
	numIOHandlers := numCores / 2
	log.Printf("Initializing server with %d workers and %d io handler\n", numWorkers, numIOHandlers)

	s := &Server{
		numWorkers:    numWorkers,
		workers:       make([]*core.Worker, numWorkers),
		numIOHandlers: numIOHandlers,
		ioHandlers:    make([]*IOHandler, numIOHandlers),
	}

	for i := 0; i < numWorkers; i++ {
		s.workers[i] = core.NewWorker(i, 1024)
	}

	for i := 0; i < numIOHandlers; i++ {
		handler, err := NewIOHandler(i, s)
		if err != nil {
			log.Fatalf("Failed to create I/O handler %d: %v", i, err)
		}
		s.ioHandlers[i] = handler
	}

	return s
}

func readCommand(fd int) (*core.Command, error) {
	var buf = make([]byte, 512)
	n, err := syscall.Read(fd, buf)
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, io.EOF
	}

	return core.ParseCmd(buf)
}

func readCommandConn(conn net.Conn) (*core.Command, error) {
	var buf = make([]byte, 512)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, io.EOF
	}

	return core.ParseCmd(buf[:n])
}

func (s *Server) getParitionID(key string) int {
	hashser := murmur3.New32()
	hashser.Write([]byte(key))
	return int(hashser.Sum32()) % s.numWorkers
}

func (s *Server) dispatch(task *core.Task) {
	var key string
	workerID := 0

	if len(task.Command.Args) > 0 {
		key = task.Command.Args[0]
		workerID = s.getParitionID(key)
	}

	log.Printf("Task pushed to worker %d", workerID)
	s.workers[workerID].TaskCh <- task
}

func (server *Server) Run(wg *sync.WaitGroup) {
	defer wg.Done()

	lisener, err := net.Listen(config.Protocol, config.Port)
	if err != nil {
		log.Fatal(err)
	}
	defer lisener.Close()

	log.Printf("Server listening on %s", config.Port)

	for {
		conn, err := lisener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}

		handler := server.ioHandlers[server.nextIOHandler%server.numIOHandlers]
		server.nextIOHandler++
		server.nextIOHandler %= server.numIOHandlers

		if err := handler.AddConn(conn); err != nil {
			log.Printf("Failed to add connection to I/O handler %d: %v", handler.id, err)
			conn.Close()
		}
	}
}

// func RunIOMultiplexingServer(wg *sync.WaitGroup) {
// 	defer wg.Done()

// 	log.Println("starting an I/O Multiplexing TCP server on", config.Port)
// 	listener, err := net.Listen(config.Protocol, config.Port)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer listener.Close()

// 	tcpListener, ok := listener.(*net.TCPListener)
// 	if !ok {
// 		log.Fatal("Listener is not a TCPListener")
// 	}
// 	listenerFile, err := tcpListener.File()
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer listenerFile.Close()

// 	serverFd := int(listenerFile.Fd())

// 	ioMultiplexer, err := io_multiplexing.CreateIOMultiplexer()
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer ioMultiplexer.Close()

// 	if err = ioMultiplexer.Monitor(io_multiplexing.Event{
// 		Fd: serverFd,
// 		Op: io_multiplexing.OpRead,
// 	}); err != nil {
// 		log.Fatal(err)
// 	}

// 	handler := core.NewHandler()
// 	lastActiveExpireExecTime := time.Now()
// 	for atomic.LoadInt32(&serverStatus) != constant.ServerStatusShuttingDown {
// 		if time.Now().After(lastActiveExpireExecTime.Add(constant.ActiveExpireFrequency)) {
// 			if !atomic.CompareAndSwapInt32(&serverStatus, constant.ServerStatusIdle, constant.ServerStatusBusy) {
// 				if atomic.LoadInt32(&serverStatus) == constant.ServerStatusShuttingDown {
// 					return
// 				}
// 			}
// 			core.ActiveDeleteExpiredKeys()
// 			atomic.SwapInt32(&serverStatus, constant.ServerStatusIdle)
// 			lastActiveExpireExecTime = time.Now()
// 		}

// 		events, err := ioMultiplexer.Wait()
// 		if err != nil {
// 			continue
// 		}

// 		if !atomic.CompareAndSwapInt32(&serverStatus, constant.ServerStatusIdle, constant.ServerStatusBusy) {
// 			if atomic.LoadInt32(&serverStatus) == constant.ServerStatusShuttingDown {
// 				return
// 			}
// 		}

// 		for i := 0; i < len(events); i++ {
// 			if events[i].Fd == serverFd {
// 				log.Printf("new client is trying to connect")
// 				connFd, _, err := syscall.Accept(serverFd)
// 				if err != nil {
// 					log.Println("err", err)
// 					continue
// 				}
// 				log.Printf("set up a new connection")
// 				if err = ioMultiplexer.Monitor(io_multiplexing.Event{
// 					Fd: connFd,
// 					Op: io_multiplexing.OpRead,
// 				}); err != nil {
// 					log.Fatal(err)
// 				}
// 			} else {
// 				cmd, err := readCommand(events[i].Fd)
// 				if err != nil {
// 					if err == io.EOF || err == syscall.ECONNRESET {
// 						log.Println("client disconnected")
// 						_ = syscall.Close(events[i].Fd)
// 						continue
// 					}
// 					log.Println("read error:", err)
// 					continue
// 				}

// 				if err = handler.ExecuteAndResponse(cmd, events[i].Fd); err != nil {
// 					log.Println("err write:", err)
// 				}
// 			}
// 		}

// 		atomic.SwapInt32(&serverStatus, constant.ServerStatusIdle)
// 	}
// }
