package server

import (
	"io"
	"log"
	"net"
	"syscall"
	"time"

	"github.com/TrienThongLu/goCache/internal/config"
	"github.com/TrienThongLu/goCache/internal/constant"
	"github.com/TrienThongLu/goCache/internal/core"
	io_multiplexing "github.com/TrienThongLu/goCache/internal/core/io_multiplexing"
)

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

func RunIOMultiplexingServer() {
	log.Println("starting an I/O Multiplexing TCP server on", config.Port)
	listener, err := net.Listen(config.Protocol, config.Port)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	tcpListener, ok := listener.(*net.TCPListener)
	if !ok {
		log.Fatal("Listener is not a TCPListener")
	}
	listenerFile, err := tcpListener.File()
	if err != nil {
		log.Fatal(err)
	}
	defer listenerFile.Close()

	serverFd := int(listenerFile.Fd())

	ioMultiplexer, err := io_multiplexing.CreateIOMultiplexer()
	if err != nil {
		log.Fatal(err)
	}
	defer ioMultiplexer.Close()

	if err = ioMultiplexer.Monitor(io_multiplexing.Event{
		Fd: serverFd,
		Op: io_multiplexing.OpRead,
	}); err != nil {
		log.Fatal(err)
	}

	handler := core.NewHandler()
	lastActiveExpireExecTime := time.Now()
	for {
		if time.Now().After(lastActiveExpireExecTime.Add(constant.ActiveExpireFrequency)) {
			core.ActiveDeleteExpiredKeys()
			lastActiveExpireExecTime = time.Now()
		}

		events, err := ioMultiplexer.Wait()
		if err != nil {
			continue
		}

		for i := 0; i < len(events); i++ {
			if events[i].Fd == serverFd {
				log.Printf("new client is trying to connect")
				connFd, _, err := syscall.Accept(serverFd)
				if err != nil {
					log.Println("err", err)
					continue
				}
				log.Printf("set up a new connection")
				if err = ioMultiplexer.Monitor(io_multiplexing.Event{
					Fd: connFd,
					Op: io_multiplexing.OpRead,
				}); err != nil {
					log.Fatal(err)
				}
			} else {
				cmd, err := readCommand(events[i].Fd)
				if err != nil {
					if err == io.EOF || err == syscall.ECONNRESET {
						log.Println("client disconnected")
						_ = syscall.Close(events[i].Fd)
						continue
					}
					log.Println("read error:", err)
					continue
				}

				if err = handler.ExecuteAndResponse(cmd, events[i].Fd); err != nil {
					log.Println("err write:", err)
				}
			}
		}
	}
}
