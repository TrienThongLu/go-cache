package data_structure

import "time"

type Obj struct {
	Value interface{}
}

type Dict struct {
	dictStore        map[string]*Obj
	expiredDictStore map[string]uint64
}

func CreateDict() *Dict {
	return &Dict{
		dictStore:        make(map[string]*Obj),
		expiredDictStore: make(map[string]uint64),
	}
}

func newObj(value interface{}) *Obj {
	return &Obj{
		Value: value,
	}
}

func (dict *Dict) SetExpiry(key string, ttlMs int64) {
	dict.expiredDictStore[key] = uint64(time.Now().UnixMilli()) + uint64(ttlMs)
}

func (dict *Dict) GetExpiry(key string) (uint64, bool) {
	exp, exist := dict.expiredDictStore[key]
	return exp, exist
}

func (dict *Dict) GetExpireDictStore() map[string]uint64 {
	return dict.expiredDictStore
}

func (dict *Dict) hasExpired(key string) bool {
	exp, exist := dict.GetExpiry(key)

	if !exist {
		return false
	}

	return exp <= uint64(time.Now().UnixMilli())
}

func (dict *Dict) Get(key string) *Obj {
	obj := dict.dictStore[key]

	if obj != nil {
		if dict.hasExpired(key) {
			dict.Del(key)
			return nil
		}
	}

	return obj
}

func (dict *Dict) Set(key string, value interface{}, ttlMs int64) {
	dict.dictStore[key] = newObj(value)

	if ttlMs > 0 {
		dict.SetExpiry(key, ttlMs)
	}
}

func (dict *Dict) Del(key string) bool {
	if _, exist := dict.dictStore[key]; exist {
		delete(dict.dictStore, key)
		delete(dict.expiredDictStore, key)
		return true
	}

	return false
}
