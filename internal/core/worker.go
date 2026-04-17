package core

import (
	"bytes"
	"errors"
	"fmt"
	"log"

	"github.com/TrienThongLu/goCache/internal/data_structure"
)

type cmdFunc func(w *Worker, args []string) []byte

var registry map[string]cmdFunc

func init() {
	registry = map[string]cmdFunc{
		"PING":           (*Worker).cmdPING,
		"INFO":           (*Worker).cmdINFO,
		"SET":            (*Worker).cmdSET,
		"GET":            (*Worker).cmdGET,
		"TTL":            (*Worker).cmdTTL,
		"DEL":            (*Worker).cmdDEL,
		"EXISTS":         (*Worker).cmdEXISTS,
		"EXPIRE":         (*Worker).cmdEXPIRE,
		"SADD":           (*Worker).cmdSADD,
		"SREM":           (*Worker).cmdSREM,
		"SISMEMBER":      (*Worker).cmdSISMEMBER,
		"SMEMBERS":       (*Worker).cmdSMEMBERS,
		"ZADD":           (*Worker).cmdZADD,
		"ZSCORE":         (*Worker).cmdZSCORE,
		"ZRANK":          (*Worker).cmdZRANK,
		"CMS.INITBYPROB": (*Worker).cmdCMSINITBYPROB,
		"CMS.INITBYDIM":  (*Worker).cmdCMSINITBYDIM,
		"CMS.INCRBY":     (*Worker).cmdCMSINCRBY,
		"CMS.QUERY":      (*Worker).cmdCMSQUERY,
		"BF.RESERVE":     (*Worker).cmdBFRESERVE,
		"BF.ADD":         (*Worker).cmdBFADD,
		"BF.EXISTS":      (*Worker).cmdBFEXISTS,
	}
}

type Task struct {
	Command *Command
	ReplyCh chan []byte
}

type Worker struct {
	id        int
	dictStore *data_structure.Dict
	cmsStore  map[string]*data_structure.CMS
	bfStore   map[string]*data_structure.BloomFilter
	TaskCh    chan *Task
}

func NewWorker(id int, bufferSize int) *Worker {
	w := &Worker{
		id:        id,
		dictStore: data_structure.CreateDict(),
		cmsStore:  make(map[string]*data_structure.CMS),
		bfStore:   make(map[string]*data_structure.BloomFilter),
		TaskCh:    make(chan *Task, bufferSize),
	}

	go w.run()
	return w
}

func (w *Worker) ExecuteAndResponse(task *Task) {
	var res []byte

	if cmd, ok := registry[task.Command.Cmd]; ok {
		res = cmd(w, task.Command.Args)
	} else {
		res = Encode(errors.New("CMD NOT FOUND"), false)
	}

	task.ReplyCh <- res
}

func (w *Worker) run() {
	log.Printf("Worker %d started", w.id)

	for task := range w.TaskCh {
		w.ExecuteAndResponse(task)
	}
}

func (w *Worker) cmdPING(args []string) []byte {
	var res []byte

	if len(args) > 1 {
		return Encode(errors.New("ERR wrong number of arguments for 'ping' command"), false)
	}

	if len(args) == 0 {
		res = Encode("PONG", true)
	} else {
		res = Encode(args[0], false)
	}

	return res
}

func (w *Worker) cmdINFO(args []string) []byte {
	var info []byte
	buf := bytes.NewBuffer(info)
	buf.WriteString("# Keyspace\r\n")
	buf.WriteString(fmt.Sprintf("db0:keys=%d,expires=0,avg_ttl=0\r\n", w.dictStore.Len()))
	return Encode(buf.String(), false)
}
