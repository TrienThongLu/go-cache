package data_structure

type Item struct {
	Score  float64
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

type BPlusTree struct {
	root   *BPlusNode
	degree int
}

func NewBPlusTree(degree int) *BPlusTree {
	return &BPlusTree{
		root:   newBPlusNode(true),
		degree: degree,
	}
}

type BPlusNode struct {
	items    []*Item
	children []*BPlusNode
	parent   *BPlusNode
	next     *BPlusNode
	isLeaf   bool
}

func newBPlusNode(isLeaf bool) *BPlusNode {
	return &BPlusNode{
		isLeaf: isLeaf,
	}
}

func (tree *BPlusTree) Add(score float64, member string) int {
	if len(member) == 0 {
		return 0
	}

	item := &Item{
		Score:  score,
		Member: member,
	}

	node := tree.root
	for !node.isLeaf {
		i := 0
		for i < len(node.items) && node.items[i].Score <= score {
			i++
		}

		node = node.children[i]
	}

	i := 0
	for _, existingItem := range node.items {
		if existingItem.Member == member {
			existingItem.Score = score
			return 1
		}

		if existingItem.Score > score {
			break
		}
		i++
	}

	node.items = append(node.items[:i], append([]*Item{item}, node.items[i:]...)...)

	if len(node.items) > tree.degree-1 {
		tree.splitNode(node)
	}

	return 1
}

func (tree *BPlusTree) splitNode(node *BPlusNode) {
	if node.parent == nil {
		tree.splitRoot(node)
		return
	}

	if node.isLeaf {
		tree.splitLeaf(node)
	} else {
		tree.splitInternal(node)
	}
}

func (tree *BPlusTree) splitRoot(oldRoot *BPlusNode) {
	newRoot := newBPlusNode(false)
	oldRoot.parent = newRoot
	tree.root = newRoot
	newRoot.children = append(newRoot.children, oldRoot)

	if oldRoot.isLeaf {
		tree.splitLeaf(oldRoot)
	} else {
		tree.splitInternal(oldRoot)
	}
}

func (tree *BPlusTree) splitLeaf(node *BPlusNode) {
	mid := len(node.items) / 2

	newNode := newBPlusNode(node.isLeaf)
	newNode.parent = node.parent
	newNode.next = node.next
	node.next = newNode

	newNode.items = append(newNode.items, node.items[mid:]...)
	node.items = node.items[:mid]

	parent := node.parent
	promotedItem := newNode.items[0]

	childIdx := 0
	for childIdx < len(parent.children) && parent.children[childIdx] != node {
		childIdx++
	}

	parent.items = append(parent.items[:childIdx], append([]*Item{promotedItem}, parent.items[childIdx:]...)...)
	parent.children = append(parent.children[:childIdx+1], append([]*BPlusNode{newNode}, parent.children[childIdx+1:]...)...)

	if len(parent.items) > tree.degree-1 {
		tree.splitNode(parent)
	}
}

func (tree *BPlusTree) splitInternal(node *BPlusNode) {
	mid := len(node.items) / 2

	newNode := newBPlusNode(node.isLeaf)
	newNode.parent = node.parent

	promotedItem := node.items[mid]
	newNode.items = append(newNode.items, node.items[mid+1:]...)
	node.items = node.items[:mid]

	newNode.children = append(newNode.children, node.children[mid+1:]...)
	node.children = node.children[:mid+1]
	for _, child := range newNode.children {
		child.parent = newNode
	}

	parent := node.parent
	childIdx := 0
	for childIdx < len(parent.children) && parent.children[childIdx] != node {
		childIdx++
	}

	parent.items = append(parent.items[:childIdx], append([]*Item{promotedItem}, parent.items[childIdx:]...)...)
	parent.children = append(parent.children[:childIdx+1], append([]*BPlusNode{newNode}, parent.children[childIdx+1:]...)...)

	if len(parent.items) > tree.degree-1 {
		tree.splitNode(parent)
	}
}

func (tree *BPlusTree) Score(member string) (float64, bool) {
	node := tree.root

	for !node.isLeaf {
		node = node.children[0]
	}

	for node != nil {
		for _, item := range node.items {
			if item.Member == member {
				return item.Score, true
			}
		}

		node = node.next
	}

	return 0, false
}

func (tree *BPlusTree) GetRank(member string) int {
	rank := 0

	node := tree.root
	for !node.isLeaf {
		node = node.children[0]
	}

	for node != nil {
		for _, item := range node.items {
			if item.Member == member {
				return rank
			}
			rank++
		}
		node = node.next
	}

	return -1
}