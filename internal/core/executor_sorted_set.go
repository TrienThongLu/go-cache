package core

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/TrienThongLu/goCache/internal/constant"
	"github.com/TrienThongLu/goCache/internal/data_structure"
)

type cmdZADD struct{}

func (cmd cmdZADD) run(args []string) []byte {
	if len(args) < 3 {
		return Encode(errors.New("ERR wrong number of arguments for 'zadd' command"), false)
	}

	key := args[0]
	numScoreElementArgs := len(args) - 1
	if numScoreElementArgs%2 != 0 {
		return Encode(errors.New("ERR wrong number of (score, member) arg"), false)
	}

	zset, exist := zsetStore[key]
	if !exist {
		zset = data_structure.NewSortedSet(constant.DefaultBPlusTreeDegree)
		zsetStore[key] = zset
	}

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

type cmdZSCORE struct{}

func (cmd cmdZSCORE) run(args []string) []byte {
	if len(args) != 2 {
		return Encode(errors.New("ERR wrong number of arguments for 'zscore' command"), false)
	}

	key, member := args[0], args[1]
	zset, exist := zsetStore[key]
	if !exist {
		return constant.RespNil
	}

	score, exist := zset.GetScore(member)
	if !exist {
		return constant.RespNil
	}

	return Encode(fmt.Sprintf("%f", score), false)
}

type cmdZRANK struct{}

func (cmd cmdZRANK) run(args []string) []byte {
	if len(args) != 2 {
		return Encode(errors.New("ERR wrong number of arguments for 'zrank' command"), false)
	}

	key, member := args[0], args[1]
	zset, exist := zsetStore[key]
	if !exist {
		return constant.RespNil
	}

	rank := zset.GetRank(member)
	return Encode(rank, false)
}
