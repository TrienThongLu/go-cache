package data_structure

import (
	"math"

	"github.com/spaolacci/murmur3"
)

const Ln2 float64 = 0.693147180559945
const Ln2Square float64 = 0.480453013918201
const ABigSeed uint32 = 0x9747b28c

type BloomFilter struct {
	hashes      int
	entries     uint64
	errorRate   float64
	bitPerEntry float64
	bits        uint64
	bytes       uint64
	bf          []uint8
}

func calcBpe(errorRate float64) float64 {
	num := math.Log(errorRate)
	return math.Abs(-(num / Ln2Square))
}

func CreateBloomFilter(entries uint64, errorRate float64) *BloomFilter {
	bloomFilter := BloomFilter{
		entries:   entries,
		errorRate: errorRate,
	}

	bloomFilter.bitPerEntry = calcBpe(errorRate)
	bits := uint64(float64(entries) * bloomFilter.bitPerEntry)
	if bits%64 == 0 {
		bloomFilter.bytes = bits / 8
	} else {
		bloomFilter.bytes = ((bits / 64) + 1) * 8
	}

	bloomFilter.bits = bloomFilter.bytes * 8
	bloomFilter.hashes = int(math.Ceil(bloomFilter.bitPerEntry * Ln2))
	bloomFilter.bf = make([]uint8, bloomFilter.bytes)

	return &bloomFilter
}

func (bloomFilter *BloomFilter) calcHash(entry string) (uint64, uint64) {
	hashser := murmur3.New128WithSeed(ABigSeed)
	hashser.Write([]byte(entry))
	return hashser.Sum128()
}

func (bloomFilter *BloomFilter) Add(entry string) {
	x, y := bloomFilter.calcHash(entry)
	for i := 0; i < bloomFilter.hashes; i++ {
		hash := (x + y*uint64(i)) % bloomFilter.bits
		idx := hash / 8
		bloomFilter.bf[idx] |= 1 << (hash % 8)
	}
}

func (bloomFilter *BloomFilter) Exist(entry string) bool {
	x, y := bloomFilter.calcHash(entry)
	for i := 0; i < bloomFilter.hashes; i++ {
		hash := (x + y*uint64(i)) % bloomFilter.bits
		idx := hash / 8

		if bloomFilter.bf[idx]&(1<<(hash%8)) == 0 {
			return false
		}
	}

	return true
}
