package core

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

const CRLF string = "\r\n"

var RespNil = []byte("$-1\r\n")

func readSimpleString(data []byte) (string, int, error) {
	pos := 1

	for data[pos] != '\r' {
		pos++
	}

	return string(data[1:pos]), pos + 2, nil
}

func readBulkString(data []byte) (string, int, error) {
	length, pos := readLen(data)
	return string(data[pos:(pos + length)]), pos + length + 2, nil
}

func readInt64(data []byte) (int64, int, error) {
	pos := 1

	var sign int64 = 1
	if data[pos] == '-' {
		sign = -1
		pos++
	} else if data[pos] == '+' {
		pos++
	}

	var num int64 = 0
	for data[pos] != '\r' {
		num = num*10 + int64(data[pos]-'0')
		pos++
	}

	return num * sign, pos + 2, nil
}

func readError(data []byte) (string, int, error) {
	return readSimpleString(data)
}

func readArray(data []byte) (interface{}, int, error) {
	length, pos := readLen(data)
	var res []interface{} = make([]interface{}, length)

	for i := range res {
		element, delta, err := DecodeOne(data[pos:])
		if err != nil {
			return nil, 0, err
		}

		res[i] = element
		pos += delta
	}

	return res, pos, nil
}

func readLen(data []byte) (int, int) {
	len, pos, _ := readInt64(data)

	return int(len), pos
}

func DecodeOne(data []byte) (interface{}, int, error) {
	if len(data) == 0 {
		return nil, 0, errors.New("No data")
	}

	switch data[0] {
	case '+':
		return readSimpleString(data)
	case '$':
		return readBulkString(data)
	case ':':
		return readInt64(data)
	case '-':
		return readError(data)
	case '*':
		return readArray(data)
	}

	return nil, 0, nil
}

func Decode(data []byte) (interface{}, error) {
	res, _, err := DecodeOne(data)
	return res, err
}

func encodeSimpleString(v interface{}) []byte {
	return []byte(fmt.Sprintf("+%s%s", v, CRLF))
}

func encodeBulkString(v string) []byte {
	return []byte(fmt.Sprintf("$%d%s%s%s", len(v), CRLF, v, CRLF))
}

func encodeInt64(v interface{}) []byte {
	return []byte(fmt.Sprintf(":%d\r\n", v))
}

func encodeError(v error) []byte {
	return []byte(fmt.Sprintf("-%s\r\n", v))
}

func encodeStringArray(sa []string) []byte {
	var b []byte
	buf := bytes.NewBuffer(b)
	for _, s := range sa {
		buf.Write(encodeBulkString(s))
	}
	return []byte(fmt.Sprintf("*%d\r\n%s", len(sa), buf.Bytes()))
}

func Encode(value interface{}, isSimpleString bool) []byte {
	switch v := value.(type) {
	case string:
		if isSimpleString {
			return encodeSimpleString(v)
		}
		return encodeBulkString(v)
	case int64, int32, int16, int8, int:
		return encodeInt64(v)
	case error:
		return encodeError(v)
	case []string:
		return encodeStringArray(value.([]string))
	case [][]string:
		var b []byte
		buf := bytes.NewBuffer(b)
		for _, sa := range value.([][]string) {
			buf.Write(encodeStringArray(sa))
		}
		return []byte(fmt.Sprintf("*%d\r\n%s", len(value.([][]string)), buf.Bytes()))
	case []interface{}:
		var b []byte
		buf := bytes.NewBuffer(b)
		for _, x := range value.([]interface{}) {
			buf.Write(Encode(x, false))
		}
		return []byte(fmt.Sprintf("*%d\r\n%s", len(value.([]interface{})), buf.Bytes()))
	default:
		return RespNil
	}
}

func ParseCmd(data []byte) (*Command, error) {
	value, err := Decode(data)
	if err != nil {
		return nil, err
	}

	array := value.([]interface{})
	tokens := make([]string, len(array))
	for i := range tokens {
		tokens[i] = array[i].(string)
	}

	res := &Command{Cmd: strings.ToUpper(tokens[0]), Args: tokens[1:]}
	return res, nil
}
