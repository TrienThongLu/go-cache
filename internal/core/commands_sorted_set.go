package core

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/TrienThongLu/goCache/internal/constant"
	"github.com/TrienThongLu/goCache/internal/data_structure"
)

const zSetType = constant.ZSetType

func (w *Worker) cmdZADD(args []string) []byte {
	if len(args) < 3 {
		return Encode(errors.New("ERR wrong number of arguments for 'zadd' command"), false)
	}

	key := args[0]
	numScoreElementArgs := len(args) - 1
	if numScoreElementArgs%2 != 0 {
		return Encode(errors.New("ERR wrong number of (score, member) arg"), false)
	}

	obj := w.dictStore.Get(key)
	if obj == nil {
		w.dictStore.Set(key, data_structure.NewSortedSet(constant.DefaultBPlusTreeDegree), zSetType, 0)
		obj = w.dictStore.Get(key)
	}

	if err := checkType(zSetType, obj.Type); err != nil {
		return Encode(err, false)
	}
	zset := obj.Value.(*data_structure.SortedSet)

	count := 0
	for i := 1; i < len(args); i += 2 {
		score, err := strconv.ParseFloat(args[i], 64)
		if err != nil {
			return Encode(errors.New("ERR score must be floating point number"), false)
		}

		member := args[i+1]
		res := zset.Add(score, member)
		if res != 1 {
			return Encode(errors.New("ERR adding score element failed"), false)
		}

		count++
	}

	return Encode(count, false)
}

func (w *Worker) cmdZSCORE(args []string) []byte {
	if len(args) != 2 {
		return Encode(errors.New("ERR wrong number of arguments for 'zscore' command"), false)
	}

	key, member := args[0], args[1]

	obj := w.dictStore.Get(key)
	if obj == nil {
		return constant.RespNil
	}

	if err := checkType(zSetType, obj.Type); err != nil {
		return Encode(err, false)
	}
	zset := obj.Value.(*data_structure.SortedSet)

	score, exist := zset.GetScore(member)
	if !exist {
		return constant.RespNil
	}

	return Encode(fmt.Sprintf("%f", score), false)
}

func (w *Worker) cmdZRANK(args []string) []byte {
	if len(args) != 2 {
		return Encode(errors.New("ERR wrong number of arguments for 'zrank' command"), false)
	}

	key, member := args[0], args[1]
	obj := w.dictStore.Get(key)
	if obj == nil {
		return constant.RespNil
	}

	if err := checkType(zSetType, obj.Type); err != nil {
		return Encode(err, false)
	}
	zset := obj.Value.(*data_structure.SortedSet)

	rank := zset.GetRank(member)
	return Encode(rank, false)
}
