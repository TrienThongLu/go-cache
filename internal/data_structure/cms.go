package data_structure

import (
	"math"

	"github.com/spaolacci/murmur3"
)

type CMS struct {
	width   uint64
	depth   uint64
	counter [][]uint64
}

const Log10PointFive = -0.30102999566

func CreateNewCMS(w uint64, d uint64) *CMS {
	cms := &CMS{
		width: w,
		depth: d,
	}

	cms.counter = make([][]uint64, d)
	for i := uint64(0); i < d; i++ {
		cms.counter[i] = make([]uint64, w)
	}

	return cms
}

func CalcCMSDim(errRate float64, errProb float64) (uint64, uint64) {
	w := uint64(math.Ceil(2.0 / errRate))
	d := uint64(math.Ceil(math.Log10(errProb) / Log10PointFive))
	return w, d
}

func (cms *CMS) calcHash(item string, seed uint64) uint64 {
	hasher := murmur3.New64WithSeed(uint32(seed))
	hasher.Write([]byte(item))
	return hasher.Sum64()
}

func (cms *CMS) IncreaseBy(item string, value uint64) uint64 {
	var minCount uint64 = math.MaxUint64

	for i := uint64(0); i < cms.depth; i++ {
		hash := cms.calcHash(item, i)
		j := hash % cms.width

		if math.MaxUint64-cms.counter[i][j] < value {
			cms.counter[i][j] = math.MaxUint64
		} else {
			cms.counter[i][j] += value
		}

		minCount = min(minCount, cms.counter[i][j])
	}

	return minCount
}

func (cms *CMS) Count(item string) uint64 {
	var minCount uint64 = math.MaxUint64

	for i := uint64(0); i < cms.depth; i++ {
		hash := cms.calcHash(item, i)
		j := hash % cms.width
		minCount = min(minCount, cms.counter[i][j])
	}

	return minCount
}
