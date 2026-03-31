package core

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/TrienThongLu/goCache/internal/constant"
)

type cmdSET struct{}

func (cmd cmdSET) run(args []string) []byte {
	if len(args) != 2 && len(args) != 4 {
		return Encode(errors.New("ERR wrong number of arguments for 'set' command"), false)
	}

	if len(args) > 2 && strings.ToUpper(args[2]) != "EX" {
		return Encode(errors.New("ERR wrong EX argument"), false)
	}

	var key, value string
	var ttlMs int64 = -1

	key, value = args[0], args[1]
	if len(args) > 2 {
		ttlSec, err := strconv.ParseInt(args[3], 10, 64)
		if err != nil {
			return Encode(errors.New("ERR value is not an integer or out of range"), false)
		}

		ttlMs = ttlSec * 1000
	}

	dictStore.Set(key, value, ttlMs)

	return constant.RespOk
}

type cmdGET struct{}

func (cmd cmdGET) run(args []string) []byte {
	if len(args) != 1 {
		return Encode(errors.New("ERR wrong number of arguments for 'get' command"), false)
	}

	key := args[0]
	obj := dictStore.Get(key)
	if obj == nil {
		return constant.RespNil
	}

	return Encode(obj.Value, false)
}

type cmdTTL struct{}

func (cmd cmdTTL) run(args []string) []byte {
	if len(args) != 1 {
		return Encode(errors.New("ERR wrong number of arguments for 'ttl' command"), false)
	}

	key := args[0]
	obj := dictStore.Get(key)
	if obj == nil {
		return constant.TtlKeyNotExist
	}

	exp, exist := dictStore.GetExpiry(key)
	if !exist {
		return constant.TtlKeyExistNoExpire
	}

	remainMs := exp - uint64(time.Now().UnixMilli())
	return Encode(int64(remainMs/1000), false)
}

type cmdDEL struct{}

func (cmd cmdDEL) run(args []string) []byte {
	if len(args) == 0 {
		return Encode(errors.New("ERR wrong number of arguments for 'del' command"), false)
	}

	res := 0
	for _, key := range args {
		obj := dictStore.Get(key)
		if obj != nil {
			dictStore.Del(key)
			res++
		}
	}

	return Encode(res, false)
}

type cmdEXISTS struct{}

func (cmd cmdEXISTS) run(args []string) []byte {
	if len(args) == 0 {
		return Encode(errors.New("ERR wrong number of arguments for 'exists' command"), false)
	}

	res := 0
	for _, key := range args {
		obj := dictStore.Get(key)
		if obj != nil {
			res++
		}
	}

	return Encode(res, false)
}

type cmdEXPIRE struct{}

func (cmd cmdEXPIRE) run(args []string) []byte {
	if len(args) != 2 {
		return Encode(errors.New("ERR wrong number of arguments for 'expire' command"), false)
	}

	key := args[0]
	obj := dictStore.Get(key)
	if obj == nil {
		return Encode(0, false)
	}

	ttlSec, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return Encode(errors.New("ERR value is not an integer or out of range"), false)
	}
	ttlMs := ttlSec * 1000

	dictStore.SetExpiry(key, ttlMs)

	return Encode(1, false)
}
