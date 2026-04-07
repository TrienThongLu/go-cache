package core

import "github.com/TrienThongLu/goCache/internal/data_structure"

var dictStore *data_structure.Dict
var cmsStore map[string]*data_structure.CMS
var bfStore map[string]*data_structure.BloomFilter

func init() {
	dictStore = data_structure.CreateDict()
	cmsStore = make(map[string]*data_structure.CMS)
	bfStore = make(map[string]*data_structure.BloomFilter)
}
