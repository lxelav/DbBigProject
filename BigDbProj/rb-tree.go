package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type Color int

const (
	RED   Color = 0
	BLACK Color = 1
)

// NodeRB представляет узел в Красно-Черном дереве
type NodeRB struct {
	key        string
	value      interface{}
	color      Color
	leftChild  *NodeRB
	rightChild *NodeRB
	parent     *NodeRB
}

// RedBlackTree представляет Красно-Черное дерево
type RedBlackTree struct {
	root *NodeRB
}

// NewRedBlackTree создает новое пустое Красно-Черное дерево
func NewRedBlackTree() *RedBlackTree {
	return &RedBlackTree{}
}

func (rb *RedBlackTree) Insert(key string, value interface{}) error {
	// Implementation of RB Tree Insert operation
	rb.InsertRB(key, value)
	return nil
}

func (rb *RedBlackTree) Get(key string) (interface{}, error) {
	nodeRB := rb.SearchRB(rb.root, key)
	if nodeRB == nil {
		return nil, fmt.Errorf("key not found")
	}
	return nodeRB.value, nil
}

func (rb *RedBlackTree) GetRange(minValue, maxValue string) ([]string, error) {
	var result []string
	var getRangeHelper func(node *NodeRB, minValue, maxValue string)
	getRangeHelper = func(node *NodeRB, minValue, maxValue string) {
		if node == nil {
			return
		}
		if node.key >= minValue {
			getRangeHelper(node.leftChild, minValue, maxValue)
		}
		if node.key >= minValue && node.key <= maxValue {
			result = append(result, node.key)
		}
		if node.key <= maxValue {
			getRangeHelper(node.rightChild, minValue, maxValue)
		}
	}
	getRangeHelper(rb.root, minValue, maxValue)
	return result, nil
}

func (rb *RedBlackTree) Update(key string, value interface{}) error {
	node, err := getNodeRB(rb.root, key)
	if err != nil {
		return err
	}
	node.value = value
	return nil
}

func (rb *RedBlackTree) Remove(key string) error {
	// Implementation of RB Tree Delete operation
	rb.DeleteRB(key)
	return nil
}

func (rb *RedBlackTree) SaveToFile(filename string) error {
	data, err := json.Marshal(rb)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, data, 0644)
}

func (rb *RedBlackTree) LoadFromFile(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, rb)
}

func getNodeRB(root *NodeRB, key string) (*NodeRB, error) {
	if root == nil {
		return nil, fmt.Errorf("key not found")
	}
	if key < root.key {
		return getNodeRB(root.leftChild, key)
	} else if key > root.key {
		return getNodeRB(root.rightChild, key)
	} else {
		return root, nil
	}
}

func (tree *RedBlackTree) InsertRB(key string, value interface{}) {
	newNode := &NodeRB{
		key:        key,
		value:      value,
		color:      RED,
		leftChild:  nil,
		rightChild: nil,
		parent:     nil,
	}
	if tree.root == nil {
		tree.root = newNode
	} else {
		tree.InsertNodeRB(tree.root, newNode)
		tree.FixInsertionRB(newNode)
	}
}

func (tree *RedBlackTree) InsertNodeRB(root, newNode *NodeRB) {
	if newNode.key < root.key {
		if root.leftChild == nil {
			root.leftChild = newNode
			newNode.parent = root
		} else {
			tree.InsertNodeRB(root.leftChild, newNode)
		}
	} else {
		if root.rightChild == nil {
			root.rightChild = newNode
			newNode.parent = root
		} else {
			tree.InsertNodeRB(root.rightChild, newNode)
		}
	}
}

func (tree *RedBlackTree) RotateLeftRB(node *NodeRB) {
	if node == nil || node.rightChild == nil {
		return
	}

	rightChild := node.rightChild
	node.rightChild = rightChild.leftChild
	if rightChild.leftChild != nil {
		rightChild.leftChild.parent = node
	}
	rightChild.parent = node.parent
	if node.parent == nil {
		tree.root = rightChild
	} else if node == node.parent.leftChild {
		node.parent.leftChild = rightChild
	} else {
		node.parent.rightChild = rightChild
	}
	rightChild.leftChild = node
	node.parent = rightChild
}

func (tree *RedBlackTree) RotateRightRB(node *NodeRB) {
	if node == nil || node.leftChild == nil {
		return
	}

	leftChild := node.leftChild
	node.leftChild = leftChild.rightChild
	if leftChild.rightChild != nil {
		leftChild.rightChild.parent = node
	}
	leftChild.parent = node.parent
	if node.parent == nil {
		tree.root = leftChild
	} else if node == node.parent.rightChild {
		node.parent.rightChild = leftChild
	} else {
		node.parent.leftChild = leftChild
	}
	leftChild.rightChild = node
	node.parent = leftChild
}

func (tree *RedBlackTree) FixInsertionRB(node *NodeRB) {
	for node != nil && node != tree.root && node.parent.color == RED {
		if node.parent == node.parent.parent.leftChild {
			uncle := node.parent.parent.rightChild
			if uncle != nil && uncle.color == RED {
				node.parent.color = BLACK
				uncle.color = BLACK
				node.parent.parent.color = RED
				node = node.parent.parent
			} else {
				if node == node.parent.rightChild {
					node = node.parent
					tree.RotateLeftRB(node)
				}
				node.parent.color = BLACK
				node.parent.parent.color = RED
				tree.RotateRightRB(node.parent.parent)
			}
		} else {
			uncle := node.parent.parent.leftChild
			if uncle != nil && uncle.color == RED {
				node.parent.color = BLACK
				uncle.color = BLACK
				node.parent.parent.color = RED
				node = node.parent.parent
			} else {
				if node == node.parent.leftChild {
					node = node.parent
					tree.RotateRightRB(node)
				}
				node.parent.color = BLACK
				node.parent.parent.color = RED
				tree.RotateLeftRB(node.parent.parent)
			}
		}
	}
	tree.root.color = BLACK
}

func (tree *RedBlackTree) DeleteRB(key string) {
	nodeToDelete := tree.SearchRB(tree.root, key)
	if nodeToDelete == nil {
		return
	}

	var child *NodeRB
	if nodeToDelete.leftChild == nil || nodeToDelete.rightChild == nil {
		child = nodeToDelete
	} else {
		child = tree.Successor(nodeToDelete)
	}

	var replacement *NodeRB
	if child.leftChild != nil {
		replacement = child.leftChild
	} else {
		replacement = child.rightChild
	}

	if replacement != nil {
		replacement.parent = child.parent
	}

	if child.parent == nil {
		tree.root = replacement
	} else if child == child.parent.leftChild {
		child.parent.leftChild = replacement
	} else {
		child.parent.rightChild = replacement
	}

	if child != nodeToDelete {
		nodeToDelete.key = child.key
		nodeToDelete.value = child.value
	}

	if child.color == BLACK && replacement != nil {
		tree.FixDeletionRB(replacement)
	}
}

func (tree *RedBlackTree) SearchRB(root *NodeRB, key string) *NodeRB {
	if root == nil || root.key == key {
		return root
	}

	if root.key < key {
		return tree.SearchRB(root.rightChild, key)
	}
	return tree.SearchRB(root.leftChild, key)
}

func (tree *RedBlackTree) Successor(node *NodeRB) *NodeRB {
	if node.rightChild != nil {
		return tree.Minimum(node.rightChild)
	}

	parent := node.parent
	for parent != nil && node == parent.rightChild {
		node = parent
		parent = parent.parent
	}
	return parent
}

func (tree *RedBlackTree) Minimum(node *NodeRB) *NodeRB {
	for node.leftChild != nil {
		node = node.leftChild
	}
	return node
}

func (tree *RedBlackTree) FixDeletionRB(node *NodeRB) {
	for node != nil && node != tree.root && node.color == BLACK {
		if node == node.parent.leftChild {
			sibling := node.parent.rightChild
			if sibling.color == RED {
				sibling.color = BLACK
				node.parent.color = RED
				tree.RotateLeftRB(node.parent)
				sibling = node.parent.rightChild
			}
			if sibling.leftChild.color == BLACK && sibling.rightChild.color == BLACK {
				sibling.color = RED
				node = node.parent
			} else {
				if sibling.rightChild.color == BLACK {
					sibling.leftChild.color = BLACK
					sibling.color = RED
					tree.RotateRightRB(sibling)
					sibling = node.parent.rightChild
				}
				sibling.color = node.parent.color
				node.parent.color = BLACK
				sibling.rightChild.color = BLACK
				tree.RotateLeftRB(node.parent)
				node = tree.root
			}
		} else {
			sibling := node.parent.leftChild
			if sibling.color == RED {
				sibling.color = BLACK
				node.parent.color = RED
				tree.RotateRightRB(node.parent)
				sibling = node.parent.leftChild
			}
			if sibling.rightChild.color == BLACK && sibling.leftChild.color == BLACK {
				sibling.color = RED
				node = node.parent
			} else {
				if sibling.leftChild.color == BLACK {
					sibling.rightChild.color = BLACK
					sibling.color = RED
					tree.RotateLeftRB(sibling)
					sibling = node.parent.leftChild
				}
				sibling.color = node.parent.color
				node.parent.color = BLACK
				sibling.leftChild.color = BLACK
				tree.RotateRightRB(node.parent)
				node = tree.root
			}
		}
	}
	if node != nil {
		node.color = BLACK
	}
}

// RedBlackCollection представляет коллекцию на основе Красно-Черного дерева
type RedBlackCollection struct {
	tree *RedBlackTree
}

func NewRedBlackCollection() *RedBlackCollection {
	return &RedBlackCollection{
		tree: NewRedBlackTree(),
	}
}

func (rb *RedBlackCollection) Insert(key string, value interface{}) error {
	return rb.tree.Insert(key, value)
}

func (rb *RedBlackCollection) Get(key string) (interface{}, error) {
	return rb.tree.Get(key)
}

func (rb *RedBlackCollection) GetRange(minValue, maxValue string) ([]string, error) {
	return rb.tree.GetRange(minValue, maxValue)
}

func (rb *RedBlackCollection) Update(key string, value interface{}) error {
	return rb.tree.Update(key, value)
}

func (rb *RedBlackCollection) Remove(key string) error {
	return rb.tree.Remove(key)
}

func (rb *RedBlackCollection) SaveToFile(filename string) error {
	data, err := json.Marshal(rb)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, data, 0644)
}

func (rb *RedBlackCollection) LoadFromFile(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, rb)
}
