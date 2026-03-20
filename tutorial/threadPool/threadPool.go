package main

import (
	"io"
	"log"
	"net"
	"sync"
)

type Job struct {
	conn net.Conn
}

type Worker struct {
	id      int
	jobChan chan Job
	wg      *sync.WaitGroup
}

type Pool struct {
	jobQueue chan Job
	workers  []*Worker
	wg       sync.WaitGroup
}

func NewWorker(id int, jobChan chan Job, wg *sync.WaitGroup) *Worker {
	return &Worker{
		id:      id,
		jobChan: jobChan,
		wg:      wg,
	}
}

func (w *Worker) Start() {
	go func() {
		defer w.wg.Done()
		for job := range w.jobChan {
			log.Printf("Worker %d is handling job from %s", w.id, job.conn.RemoteAddr())
			handleConnection(job.conn)
		}
	}()
}

func NewPool(numOfWorker, limit int) *Pool {
	return &Pool{
		jobQueue: make(chan Job, limit),
		workers:  make([]*Worker, numOfWorker),
	}
}

func (p *Pool) AddJob(conn net.Conn) {
	p.jobQueue <- Job{conn: conn}
}

func (p *Pool) Start() {
	for i := 0; i < len(p.workers); i++ {
		p.wg.Add(1)
		worker := NewWorker(i, p.jobQueue, &p.wg)
		p.workers[i] = worker
		worker.Start()
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	log.Println("handle conn from", conn.RemoteAddr())

	for {
		cmd, err := readCommand(conn)
		if err != nil {
			conn.Close()
			log.Println("Client disconnected", conn.RemoteAddr())
			if err != io.EOF {
				break
			}
		}

		if err = respond(cmd, conn); err != nil {
			log.Println("Err write:", err)
		}
	}
}

func readCommand(conn net.Conn) (string, error) {
	buff := make([]byte, 1024)
	n, err := conn.Read(buff)
	if err != nil {
		return "", err
	}

	return string(buff[:n]), nil
}

func respond(cmd string, conn net.Conn) error {
	if _, err := conn.Write([]byte(cmd)); err != nil {
		return err
	}

	return nil
}

func main() {
	listener, err := net.Listen("tcp", ":3000")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	pool := NewPool(3, 2)
	pool.Start()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		pool.AddJob(conn)
	}
}
