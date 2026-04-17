package core

import (
	"errors"
	"fmt"
	"math"
	"strconv"

	"github.com/TrienThongLu/goCache/internal/constant"
	"github.com/TrienThongLu/goCache/internal/data_structure"
)

func (w *Worker) cmdCMSINITBYDIM(args []string) []byte {
	if len(args) != 3 {
		return Encode(errors.New("(error) ERR wrong number of arguments for 'CMS.INITBYDIM' command"), false)
	}

	key := args[0]
	width, err := strconv.ParseUint(args[1], 10, 64)
	if err != nil {
		return Encode(fmt.Errorf("width must be a integer number %s", args[1]), false)
	}

	depth, err := strconv.ParseUint(args[2], 10, 64)
	if err != nil {
		return Encode(fmt.Errorf("depth must be a integer number %s", args[2]), false)
	}

	if _, exist := w.cmsStore[key]; exist {
		return Encode(errors.New("CMS: key already exist"), false)
	}

	w.cmsStore[key] = data_structure.CreateNewCMS(uint64(width), uint64(depth))

	return constant.RespOk
}

func (w *Worker) cmdCMSINITBYPROB(args []string) []byte {
	if len(args) != 3 {
		return Encode(errors.New("(error) ERR wrong number of arguments for 'CMS.INITBYPROB' command"), false)
	}

	key := args[0]
	errRate, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return Encode(fmt.Errorf("errRate must be a floating point number %s", args[1]), false)
	}
	if errRate <= 0 || errRate >= 1 {
		return Encode(errors.New("CMS: invalid overestimation value"), false)
	}

	probability, err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		return Encode(fmt.Errorf("probability must be a floating point number %s", args[2]), false)
	}
	if probability <= 0 || probability >= 1 {
		return Encode(errors.New("CMS: invalid probability value"), false)
	}

	if _, exist := w.cmsStore[key]; exist {
		return Encode(errors.New("CMS: key already exist"), false)
	}

	width, depth := data_structure.CalcCMSDim(errRate, probability)
	w.cmsStore[key] = data_structure.CreateNewCMS(width, depth)

	return constant.RespOk
}

func (w *Worker) cmdCMSINCRBY(args []string) []byte {
	if len(args) < 3 || len(args)%2 == 0 {
		return Encode(errors.New("(error) ERR wrong number of arguments for 'CMS.INCRBY' command"), false)
	}

	key := args[0]
	cms, exist := w.cmsStore[key]
	if !exist {
		return Encode(errors.New("CMS: key does not exist"), false)
	}

	var res []string
	for i := 1; i < len(args); i += 2 {
		item := args[i]
		value, err := strconv.ParseUint(args[i+1], 10, 64)
		if err != nil {
			return Encode(fmt.Sprintf("increment mus be a non negative integer number %s", args[1]), false)
		}

		count := cms.IncreaseBy(item, value)
		if count == math.MaxUint64 {
			res = append(res, "CMS: INCRBY overflow")
		}

		res = append(res, fmt.Sprintf("%d", count))
	}

	return Encode(res, false)
}

func (w *Worker) cmdCMSQUERY(args []string) []byte {
	if len(args) < 2 {
		return Encode(errors.New("(error) ERR wrong number of arguments for 'CMS.QUERY' command"), false)
	}

	key := args[0]
	cms, exist := w.cmsStore[key]
	if !exist {
		return Encode(errors.New("CMS: key does not exist"), false)
	}

	var res []string
	for i := 1; i < len(args); i++ {
		item := args[i]
		res = append(res, fmt.Sprintf("%d", cms.Count(item)))
	}

	return Encode(res, false)
}
