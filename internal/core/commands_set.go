package core

import (
	"errors"

	"github.com/TrienThongLu/goCache/internal/constant"
	"github.com/TrienThongLu/goCache/internal/data_structure"
)

const setType = constant.SetType

func (w *Worker) cmdSADD(args []string) []byte {
	if len(args) < 2 {
		return Encode(errors.New("ERR wrong number of arguments for 'sadd' command"), false)
	}

	key := args[0]
	obj := w.dictStore.Get(key)
	if obj == nil {
		w.dictStore.Set(key, data_structure.CreateSimpleSet(), setType, 0)
		obj = w.dictStore.Get(key)
	}

	if err := checkType(setType, obj.Type); err != nil {
		return Encode(err, false)
	}
	set := obj.Value.(*data_structure.SimpleSet)

	return Encode(set.Add(args[1:]...), false)
}

func (w *Worker) cmdSREM(args []string) []byte {
	if len(args) < 2 {
		return Encode(errors.New("ERR wrong number of arguments for 'srem' command"), false)
	}

	key := args[0]
	obj := w.dictStore.Get(key)
	if obj == nil {
		return Encode(0, false)
	}

	if err := checkType(setType, obj.Type); err != nil {
		return Encode(err, false)
	}
	set := obj.Value.(*data_structure.SimpleSet)

	return Encode(set.Remove(args[1:]...), false)
}

func (w *Worker) cmdSISMEMBER(args []string) []byte {
	if len(args) != 2 {
		return Encode(errors.New("ERR wrong number of arguments for 'sismember' command"), false)
	}

	key := args[0]
	obj := w.dictStore.Get(key)
	if obj == nil {
		return Encode(0, false)
	}

	if err := checkType(setType, obj.Type); err != nil {
		return Encode(err, false)
	}
	set := obj.Value.(*data_structure.SimpleSet)

	return Encode(set.IsMember(args[1]), false)
}

func (w *Worker) cmdSMEMBERS(args []string) []byte {
	if len(args) != 1 {
		return Encode(errors.New("ERR wrong number of arguments for 'smembers' command"), false)
	}

	key := args[0]
	obj := w.dictStore.Get(key)
	if obj == nil {
		return Encode(0, false)
	}

	if err := checkType(setType, obj.Type); err != nil {
		return Encode(err, false)
	}
	set := obj.Value.(*data_structure.SimpleSet)

	return Encode(set.Members(), false)
}
