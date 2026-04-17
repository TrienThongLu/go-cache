package core

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/TrienThongLu/goCache/internal/constant"
	"github.com/TrienThongLu/goCache/internal/data_structure"
)

func (w *Worker) cmdBFRESERVE(args []string) []byte {
	if len(args) != 3 && len(args) != 5 {
		return Encode(errors.New("(error) ERR wrong number of arguments for 'BF.RESERVE' command"), false)
	}

	key := args[0]
	errRate, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return Encode(fmt.Errorf("error rate must be a floating point number %s", args[1]), false)
	}

	capacity, err := strconv.ParseUint(args[2], 10, 64)
	if err != nil {
		return Encode(fmt.Errorf("error rate must be a positive integer number %s", args[2]), false)
	}

	if _, exist := w.bfStore[key]; exist {
		return Encode(fmt.Errorf("Bloom filter with key '%s' already exist", key), false)
	}

	w.bfStore[key] = data_structure.CreateBloomFilter(capacity, errRate)

	return constant.RespOk
}

func (w *Worker) cmdBFADD(args []string) []byte {
	if len(args) != 2 {
		return Encode(errors.New("(error) ERR wrong number of arguments for 'BF.ADD' command"), false)
	}

	key := args[0]
	bf, exist := w.bfStore[key]
	if !exist {
		bf = data_structure.CreateBloomFilter(constant.BfDefaultInitCapacity,
			constant.BfDefaultErrRate)
		w.bfStore[key] = bf
	}

	bf.Add(args[1])
	return constant.RespOk
}

func (w *Worker) cmdBFEXISTS(args []string) []byte {
	if len(args) != 2 {
		return Encode(errors.New("(error) ERR wrong number of arguments for 'BF.EXIST' command"), false)
	}

	key := args[0]
	bf, exist := w.bfStore[key]
	if !exist {
		return Encode(fmt.Errorf("Bloom filter with key '%s' is not exist", key), false)
	}

	if !bf.Exist(args[1]) {
		return constant.RespZero
	}

	return constant.RespOne
}
