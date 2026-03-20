package main

import (
	"io"
	"log"
	"net"
	"time"
)

func main() {
	listener, err := net.Listen("tcp", ":3000")
	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	log.Println(conn.RemoteAddr())

	for {
		buff := make([]byte, 1024)
		buffLen, err := conn.Read(buff)

		log.Printf("Received: %s", string(buff[:buffLen]))

		if err != nil {
			if err != io.EOF {
				log.Printf("Read error: %v", err)
			}
			return
		}

		time.Sleep(time.Second * 2)
		_, err = conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\nHello world\r\n"))
		if err != nil {
			log.Printf("Write error: %v", err)
			return
		}
	}
}
