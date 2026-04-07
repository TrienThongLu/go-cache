package core

import (
	"errors"

	"github.com/TrienThongLu/goCache/internal/constant"
	"github.com/TrienThongLu/goCache/internal/data_structure"
)

const setType = constant.SetType

type cmdSADD struct{}

func (cmd cmdSADD) run(args []string) []byte {
	if len(args) < 2 {
		return Encode(errors.New("ERR wrong number of arguments for 'sadd' command"), false)
	}

	key := args[0]
	obj := dictStore.Get(key)
	if obj == nil {
		dictStore.Set(key, data_structure.CreateSimpleSet(), setType, 0)
		obj = dictStore.Get(key)
	}

	set := obj.Value.(*data_structure.SimpleSet)
	return Encode(set.Add(args[1:]...), false)
}

type cmdSREM struct{}

func (cmd cmdSREM) run(args []string) []byte {
	if len(args) < 2 {
		return Encode(errors.New("ERR wrong number of arguments for 'srem' command"), false)
	}

	key := args[0]
	obj := dictStore.Get(key)
	if obj == nil {
		return Encode(0, false)
	}

	set := obj.Value.(*data_structure.SimpleSet)
	return Encode(set.Remove(args[1:]...), false)
}

type cmdSISMEMBER struct{}

func (cmd cmdSISMEMBER) run(args []string) []byte {
	if len(args) != 2 {
		return Encode(errors.New("ERR wrong number of arguments for 'sismember' command"), false)
	}

	key := args[0]
	obj := dictStore.Get(key)
	if obj == nil {
		return Encode(0, false)
	}

	set := obj.Value.(*data_structure.SimpleSet)
	return Encode(set.IsMember(args[1]), false)
}

type cmdSMEMBERS struct{}

func (cmd cmdSMEMBERS) run(args []string) []byte {
	if len(args) != 1 {
		return Encode(errors.New("ERR wrong number of arguments for 'smembers' command"), false)
	}

	key := args[0]
	obj := dictStore.Get(key)
	if obj == nil {
		return Encode(0, false)
	}

	set := obj.Value.(*data_structure.SimpleSet)
	return Encode(set.Members(), false)
}
