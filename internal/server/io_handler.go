package server

import (
	"io"
	"log"
	"net"
	"sync"
	"syscall"

	"github.com/TrienThongLu/goCache/internal/core"
	"github.com/TrienThongLu/goCache/internal/core/io_multiplexing"
)

type IOHandler struct {
	id            int
	ioMultiplexer io_multiplexing.IOMultiplexer
	mu            sync.Mutex
	server        *Server
	conns         map[int]net.Conn
}

func NewIOHandler(id int, server *Server) (*IOHandler, error) {
	multiplexer, err := io_multiplexing.CreateIOMultiplexer()
	if err != nil {
		return nil, err
	}

	ioHandler := &IOHandler{
		id:            id,
		ioMultiplexer: multiplexer,
		server:        server,
		conns:         make(map[int]net.Conn),
	}

	go ioHandler.run()

	return ioHandler, nil
}

func (h *IOHandler) AddConn(conn net.Conn) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	tcpConn := conn.(*net.TCPConn)
	rawConn, err := tcpConn.SyscallConn()
	if err != nil {
		return err
	}

	var connFd int
	err = rawConn.Control(func(fd uintptr) {
		connFd = int(fd)
		log.Printf("I/O Handler %d is monitoring fd %d", h.id, connFd)
		h.conns[connFd] = conn
		h.ioMultiplexer.Monitor(io_multiplexing.Event{
			Fd: connFd,
			Op: io_multiplexing.OpRead,
		})
	})

	return err
}

func (h *IOHandler) closeConn(fd int) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if conn, ok := h.conns[fd]; ok {
		conn.Close()
		delete(h.conns, fd)
	}
}

func (h *IOHandler) run() {
	log.Printf("I/O handler %d started", h.id)

	for {
		events, err := h.ioMultiplexer.Wait()
		if err != nil {
			continue
		}

		for _, event := range events {
			connFd := event.Fd
			h.mu.Lock()
			conn, ok := h.conns[connFd]
			h.mu.Unlock()
			if !ok {
				continue
			}

			cmd, err := readCommandConn(conn)
			if err != nil {
				if err == io.EOF || err == syscall.ECONNRESET {
					log.Printf("Connection closed by client: %v", err)
				} else {
					log.Printf("Error reading command: %v", err)
				}

				h.closeConn(connFd)
				continue
			}

			replyCh := make(chan []byte, 1)
			task := &core.Task{
				Command: cmd,
				ReplyCh: replyCh,
			}

			h.server.dispatch(task)
			res := <-replyCh

			if _, err := conn.Write(res); err != nil {
				log.Println("err write:", err)
			}
		}
	}
}
