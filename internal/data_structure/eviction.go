package data_structure

import (
	"sort"
	"time"

	"github.com/TrienThongLu/goCache/internal/config"
)

type EvictionCandidate struct {
	Key            string
	LastAccessTime uint32
	ExpireAt       uint64
}

type EvictionPool struct {
	pool []*EvictionCandidate
}

func now() uint32 {
	return uint32(time.Now().Unix())
}

func createEvictionPool(size int) EvictionPool {
	return EvictionPool{
		pool: make([]*EvictionCandidate, size),
	}
}

func (p *EvictionPool) Push(key string, lastAccessTime uint32, expireAt uint64) {
	newItem := &EvictionCandidate{
		Key:            key,
		ExpireAt:       expireAt,
		LastAccessTime: lastAccessTime,
	}

	exist := false
	for i := 0; i < len(p.pool); i++ {
		if p.pool[i].Key == key {
			exist = true
			p.pool[i] = newItem
		}
	}
	if !exist {
		p.pool = append(p.pool, newItem)
	}

	switch config.EvictionPolicy {
	case "allkeys-lru":
		sort.Sort(LRU(p.pool))
	}

	if len(p.pool) > config.EpoolMaxSize {
		lastIndex := len(p.pool) - 1
		p.pool = p.pool[:lastIndex]
	}
}

func (p *EvictionPool) Pop() *EvictionCandidate {
	if len(p.pool) == 0 {
		return nil
	}

	candidate := p.pool[0]
	p.pool = p.pool[1:]

	return candidate
}

func (p *EvictionPool) Len() int {
	return len(p.pool)
}

var EPool EvictionPool = createEvictionPool(0)
