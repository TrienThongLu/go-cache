package core

import (
	"errors"

	"github.com/TrienThongLu/goCache/internal/constant"
)

type Command struct {
	Cmd  string
	Args []string
}

func checkType(dataType constant.DataType, currentType constant.DataType) error {
	if dataType != currentType {
		return errors.New("(error) WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	return nil
}
