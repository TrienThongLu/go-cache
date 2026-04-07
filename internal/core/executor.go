package core

import (
	"bytes"
	"errors"
	"fmt"
	"syscall"
)

type cmd interface {
	run(args []string) []byte
}

type Handler struct {
	registry map[string]cmd
}

func NewHandler() *Handler {
	return &Handler{
		registry: map[string]cmd{
			"PING":           cmdPING{},
			"INFO":           cmdINFO{},
			"SET":            cmdSET{},
			"GET":            cmdGET{},
			"TTL":            cmdTTL{},
			"DEL":            cmdDEL{},
			"EXISTS":         cmdEXISTS{},
			"EXPIRE":         cmdEXPIRE{},
			"SADD":           cmdSADD{},
			"SREM":           cmdSREM{},
			"SISMEMBER":      cmdSISMEMBER{},
			"SMEMBERS":       cmdSMEMBERS{},
			"ZADD":           cmdZADD{},
			"ZSCORE":         cmdZSCORE{},
			"ZRANK":          cmdZRANK{},
			"CMS.INITBYPROB": cmdCMSINITBYPROB{},
			"CMS.INITBYDIM":  cmdCMSINITBYDIM{},
			"CMS.INCRBY":     cmdCMSINCRBY{},
			"CMS.QUERY":      cmdCMSQUERY{},
			"BF.RESERVE":     cmdBFRESERVE{},
			"BF.ADD":         cmdBFADD{},
			"BF.EXISTS":      cmdBFEXISTS{},
		},
	}
}

func (handler *Handler) ExecuteAndResponse(command *Command, connFd int) error {
	var res []byte

	if cmd, ok := handler.registry[command.Cmd]; ok {
		res = cmd.run(command.Args)
	} else {
		res = Encode(errors.New("CMD NOT FOUND"), false)
	}

	_, err := syscall.Write(connFd, res)

	return err
}

type cmdPING struct{}

func (cmd cmdPING) run(args []string) []byte {
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

type cmdINFO struct{}

func (cmd cmdINFO) run(args []string) []byte {
	var info []byte
	buf := bytes.NewBuffer(info)
	buf.WriteString("# Keyspace\r\n")
	buf.WriteString(fmt.Sprintf("db0:keys=%d,expires=0,avg_ttl=0\r\n", dictStore.Len()))
	return Encode(buf.String(), false)
}
