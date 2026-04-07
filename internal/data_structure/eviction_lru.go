package data_structure

type LRU []*EvictionCandidate

func (l LRU) Len() int {
	return len(l)
}

func (l LRU) Less(i, j int) bool {
	return l[i].LastAccessTime < l[j].LastAccessTime
}

func (l LRU) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}
