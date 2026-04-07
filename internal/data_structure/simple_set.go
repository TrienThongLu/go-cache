package data_structure

import "log"

type SimpleSet struct {
	dict map[string]struct{}
}

func CreateSimpleSet() *SimpleSet {
	return &SimpleSet{
		dict: make(map[string]struct{}),
	}
}

func (set *SimpleSet) Add(members ...string) int {
	added := 0

	for _, m := range members {
		if _, exist := set.dict[m]; !exist {
			set.dict[m] = struct{}{}
			added++
		}
	}

	return added
}

func (set *SimpleSet) Remove(members ...string) int {
	removed := 0
	for _, m := range members {
		log.Println(m)
		if _, exist := set.dict[m]; exist {
			delete(set.dict, m)
			removed++
		}
	}
	return removed
}

func (set *SimpleSet) IsMember(member string) int {
	_, exist := set.dict[member]
	if exist {
		return 1
	}
	return 0
}

func (set *SimpleSet) Members() []string {
	res := make([]string, 0, len(set.dict))

	for k := range set.dict {
		res = append(res, k)
	}

	return res
}
