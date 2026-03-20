package data_structure

type Item struct {
    Score float64
    Member string
}

func (item *Item) compareTo(other *Item) int {
    if item.Score < other.Score {
        return -1
    } 

    if item.Score > other.Score {
        return 1
    }

    if item.Member < other.Member {
        return -1
    }

    if item.Member > other.Member {
        return 1
    }

    return 0
}