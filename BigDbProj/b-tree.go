package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

const m = 2

type NodeB struct {
	keys     []string
	children []*NodeB
	leaf     bool
	value    interface{}
}

type BTree struct {
	root *NodeB
}

func NewNodeB(leaf bool, value interface{}) *NodeB {
	return &NodeB{
		keys:     make([]string, 0),
		children: make([]*NodeB, 0),
		leaf:     leaf,
		value:    value,
	}
}

func NewBTree() *BTree {
	return &BTree{
		root: NewNodeB(true, nil),
	}
}

func (t *BTree) Insert(key string, value interface{}) error {
	root := t.root
	if len(root.keys) == (2*m - 1) { //2*t - 1
		newRoot := NewNodeB(false, nil)
		newRoot.children = append(newRoot.children, root)
		t.root = newRoot
		t.splitChild(0)
		t.insertNonFull(newRoot, key)
	} else {
		t.insertNonFull(root, key)
	}
	return nil
}

func (t *BTree) splitChild(i int) {
	child := t.root.children[i]
	newChild := NewNodeB(child.leaf, nil)
	mid := len(child.keys) / 2
	splitKey := child.keys[mid]

	t.root.children = append(t.root.children[:i+1], nil)
	copy(t.root.children[i+2:], t.root.children[i+1:])
	t.root.children[i+1] = newChild

	newChild.keys = append(newChild.keys, child.keys[mid+1:]...)
	child.keys = child.keys[:mid]

	if !child.leaf {
		newChild.children = append(newChild.children, child.children[mid+1:]...)
		child.children = child.children[:mid+1]
	}

	t.root.keys = append(t.root.keys[:i], append([]string{splitKey}, t.root.keys[i:]...)...)
}

func (t *BTree) insertNonFull(node *NodeB, key string) {
	i := len(node.keys) - 1
	if node.leaf {
		for i >= 0 && key < node.keys[i] {
			i--
		}
		node.keys = append(node.keys[:i+1], append([]string{key}, node.keys[i+1:]...)...)
	} else {
		for i >= 0 && key < node.keys[i] {
			i--
		}
		i++
		if len(node.children[i].keys) == (2*m - 1) {
			t.splitChild(i)
			if key > node.keys[i] {
				i++
			}
		}
		t.insertNonFull(node.children[i], key)
	}
}

func (t *BTree) Search(key string) *NodeB {
	return t.search(t.root, key)
}

func (t *BTree) search(node *NodeB, key string) *NodeB {
	if node == nil {
		return nil
	}
	i := 0
	for i < len(node.keys) && key > node.keys[i] {
		i++
	}
	if i < len(node.keys) && key == node.keys[i] {
		return node
	}
	if node.leaf {
		return nil
	}
	return t.search(node.children[i], key)
}

func (t *BTree) Remove(key string) error {
	t.root = t.delete(t.root, key)
	if len(t.root.keys) == 0 && len(t.root.children) == 1 {
		t.root = t.root.children[0]
	}
	return nil
}

func (t *BTree) delete(node *NodeB, key string) *NodeB {
	i := 0
	for i < len(node.keys) && key > node.keys[i] {
		i++
	}
	if i < len(node.keys) && key == node.keys[i] {
		if node.leaf {
			t.removeFromLeaf(node, i)
		} else {
			t.removeFromNonLeaf(node, i)
		}
	} else {
		if node.leaf {
			fmt.Println("Key", key, "not found")
			return node
		}
		var flag bool
		if i == len(node.keys) {
			flag = true
		} else {
			flag = false
		}
		if len(node.children[i].keys) < m {
			t.fill(node, i)
		}
		if flag && i > len(node.keys) {
			node.children[i-1] = t.delete(node.children[i-1], key)
		} else {
			node.children[i] = t.delete(node.children[i], key)
		}
	}
	return node
}

func (t *BTree) removeFromLeaf(node *NodeB, idx int) {
	copy(node.keys[idx:], node.keys[idx+1:])
	node.keys = node.keys[:len(node.keys)-1]
}

func (t *BTree) removeFromNonLeaf(node *NodeB, idx int) {
	key := node.keys[idx]
	if len(node.children[idx].keys) >= m { //t
		pred := t.getPred(node, idx)
		node.keys[idx] = pred
		t.delete(node.children[idx], pred)
	} else if len(node.children[idx+1].keys) >= m { //t
		succ := t.getSucc(node, idx)
		node.keys[idx] = succ
		t.delete(node.children[idx+1], succ)
	} else {
		t.merge(node, idx)
		t.delete(node.children[idx], key)
	}
}

func (t *BTree) getPred(node *NodeB, idx int) string {
	cur := node.children[idx]
	for !cur.leaf {
		cur = cur.children[len(cur.children)-1]
	}
	return cur.keys[len(cur.keys)-1]
}

func (t *BTree) getSucc(node *NodeB, idx int) string {
	cur := node.children[idx+1]
	for !cur.leaf {
		cur = cur.children[0]
	}
	return cur.keys[0]
}

func (t *BTree) fill(node *NodeB, idx int) {
	if idx != 0 && len(node.children[idx-1].keys) >= m { //t
		t.borrowFromPrev(node, idx)
	} else if idx != len(node.keys) && len(node.children[idx+1].keys) >= m { //t
		t.borrowFromNext(node, idx)
	} else {
		if idx != len(node.keys) {
			t.merge(node, idx)
		} else {
			t.merge(node, idx-1)
		}
	}
}

func (t *BTree) borrowFromPrev(node *NodeB, idx int) {
	child := node.children[idx]
	sibling := node.children[idx-1]

	// Перемещаем ключ из родительского узла в конец child
	child.keys = append([]string{node.keys[idx-1]}, child.keys...)

	// Если не лист, перемещаем последнего ребенка из sibling в начало child
	if !child.leaf {
		child.children = append([]*NodeB{sibling.children[len(sibling.children)-1]}, child.children...)
	}
	node.keys[idx-1] = sibling.keys[len(sibling.keys)-1]
	sibling.keys = sibling.keys[:len(sibling.keys)-1]
	if !sibling.leaf {
		sibling.children = sibling.children[:len(sibling.children)-1]
	}
}

func (t *BTree) borrowFromNext(node *NodeB, idx int) {
	child := node.children[idx]
	sibling := node.children[idx+1]

	// Перемещаем ключ из родительского узла в конец child
	child.keys = append(child.keys, node.keys[idx])

	// Если не лист, перемещаем первого ребенка из sibling в конец child
	if !child.leaf {
		child.children = append(child.children, sibling.children[0])
	}
	node.keys[idx] = sibling.keys[0]
	sibling.keys = sibling.keys[1:]
	if !sibling.leaf {
		sibling.children = sibling.children[1:]
	}
}

func (t *BTree) merge(node *NodeB, idx int) {
	child := node.children[idx]
	sibling := node.children[idx+1]

	child.keys = append(child.keys, node.keys[idx])
	child.keys = append(child.keys, sibling.keys...)
	if !child.leaf {
		child.children = append(child.children, sibling.children...)
	}

	node.keys = append(node.keys[:idx], node.keys[idx+1:]...)
	node.children = append(node.children[:idx+1], node.children[idx+2:]...)
}

func (t *BTree) Get(key string) (interface{}, error) {
	node := t.Search(key)
	if node == nil {
		return nil, fmt.Errorf("key not found")
	}
	return node, nil
}

func (t *BTree) GetRange(minValue, maxValue string) ([]string, error) {
	keysInRange := make([]string, 0)
	t.traverseRange(t.root, minValue, maxValue, &keysInRange)
	return keysInRange, nil
}

func (t *BTree) traverseRange(node *NodeB, minValue, maxValue string, keysInRange *[]string) {
	if node == nil {
		return
	}

	i := 0
	for i < len(node.keys) && node.keys[i] < minValue {
		i++
	}

	for ; i < len(node.keys); i++ {
		if node.children[i] != nil {
			t.traverseRange(node.children[i], minValue, maxValue, keysInRange)
		}
		if node.keys[i] >= minValue && node.keys[i] <= maxValue {
			*keysInRange = append(*keysInRange, node.keys[i])
		}
		if node.keys[i] > maxValue {
			break
		}
	}

	if node.children[i] != nil {
		t.traverseRange(node.children[i], minValue, maxValue, keysInRange)
	}
}

func (t *BTree) Update(key string, value interface{}) error {
	node := t.Search(key)
	if node == nil {
		return fmt.Errorf("key not found")
	}
	node.value = value
	return nil
}

func (t *BTree) SaveToFile(filename string) error {
	data, err := json.Marshal(t)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, data, 0644)
}

func (t *BTree) LoadFromFile(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, t)
}
