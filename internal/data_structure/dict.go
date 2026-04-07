package data_structure

import (
	"log"
	"time"

	"github.com/TrienThongLu/goCache/internal/config"
	"github.com/TrienThongLu/goCache/internal/constant"
)

type Obj struct {
	Value          interface{}
	Type           constant.DataType
	LastAccessTime uint32
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

func newObj(value interface{}, dataType constant.DataType) *Obj {
	return &Obj{
		Value:          value,
		Type:           dataType,
		LastAccessTime: now(),
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
		obj.LastAccessTime = now()
	}

	return obj
}

func (dict *Dict) Len() int {
	return len(dict.dictStore)
}

func (dict *Dict) EvictIfNeeded() {
	if float64(dict.Len())/float64(config.MaxKeyNumber) < config.MinKeyRatioForEviction {
		return
	}

	dict.PopulateEpool()
	dict.evict()
}

func (dict *Dict) evict() {
	log.Print("trigger eviction")
	evictCount := int64(config.EvictionRatio * float64(config.MaxKeyNumber))

	for i := 0; i < int(evictCount) && EPool.Len() > 0; i++ {
		candidate := EPool.Pop()
		dict.Del(candidate.Key)
	}
}

func (dict *Dict) PopulateEpool() {
	remain := config.EpoolLruSampleSize
	for key := range dict.dictStore {
		obj := dict.dictStore[key]

		ttl, exist := dict.GetExpiry(key)
		if !exist {
			ttl = 0
		}

		EPool.Push(key, obj.LastAccessTime, ttl)

		remain--
		if remain == 0 {
			break
		}
	}

	log.Println("EPool:")
	for _, item := range EPool.pool {
		log.Println(item.Key, item.LastAccessTime, item.ExpireAt)
	}
}

func (dict *Dict) Set(key string, value interface{}, dataType constant.DataType, ttlMs int64) {
	dict.EvictIfNeeded()
	dict.dictStore[key] = newObj(value, dataType)

	if ttlMs > 0 {
		dict.SetExpiry(key, ttlMs)
	}
}

func (dict *Dict) Del(key string) bool {
	if _, exist := dict.dictStore[key]; exist {
		log.Printf("delete key %s", key)
		delete(dict.dictStore, key)
		delete(dict.expiredDictStore, key)
		return true
	}

	return false
}
