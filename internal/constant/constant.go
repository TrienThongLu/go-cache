package constant

import "time"

var RespNil = []byte("$-1\r\n")
var RespOk = []byte("+OK\r\n")
var RespZero = []byte(":0\r\n")
var RespOne = []byte(":1\r\n")
var TtlKeyNotExist = []byte(":-2\r\n")
var TtlKeyExistNoExpire = []byte(":-1\r\n")

const ActiveExpireFrequency = 100 * time.Millisecond
const ActiveExpireSampleSize = 20
const ActiveExpireThreshold = 0.1
const DefaultBPlusTreeDegree = 4

const BfDefaultInitCapacity = 100
const BfDefaultErrRate = 0.01
