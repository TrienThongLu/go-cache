package constant

type DataType string

const (
	StringType DataType = "string"
	SetType    DataType = "set"
	ZSetType   DataType = "zset"
	CMSType    DataType = "cms"
	BFType     DataType = "bf"
)
