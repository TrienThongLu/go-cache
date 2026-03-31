package core

import (
	"errors"

	"github.com/TrienThongLu/goCache/internal/data_structure"
)

type cmdSADD struct{}

func (cmd cmdSADD) run(args []string) []byte {
	if len(args) < 2 {
		return Encode(errors.New("ERR wrong number of arguments for 'sadd' command"), false)
	}

	key := args[0]
	set, exist := setStore[key]
	if !exist {
		set = data_structure.CreateSimpleSet(key)
		setStore[key] = set
	}

	return Encode(set.Add(args[1:]...), false)
}

type cmdSREM struct{}

func (cmd cmdSREM) run(args []string) []byte {
	if len(args) < 2 {
		return Encode(errors.New("ERR wrong number of arguments for 'srem' command"), false)
	}

	key := args[0]
	set, exist := setStore[key]
	if !exist {
		return Encode(0, false)
	}

	return Encode(set.Remove(args[1:]...), false)
}

type cmdSISMEMBER struct{}

func (cmd cmdSISMEMBER) run(args []string) []byte {
	if len(args) != 2 {
		return Encode(errors.New("ERR wrong number of arguments for 'sismember' command"), false)
	}

	key := args[0]
	set, exist := setStore[key]
	if !exist {
		return Encode(0, false)
	}

	return Encode(set.IsMember(args[1]), false)
}

type cmdSMEMBERS struct{}

func (cmd cmdSMEMBERS) run(args []string) []byte {
	if len(args) != 1 {
		return Encode(errors.New("ERR wrong number of arguments for 'smembers' command"), false)
	}

	key := args[0]
	set, exist := setStore[key]
	if !exist {
		return Encode(make([]string, 0), false)
	}

	return Encode(set.Members(), false)
}
