package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
)

type Node struct {
	key    string
	value  interface{}
	height int
	left   *Node
	right  *Node
}

type AVLTree struct {
	root *Node
}

func NewAVLTree() *AVLTree {
	return &AVLTree{}
}

func (avl *AVLTree) Insert(key string, value interface{}) error {
	var err error
	avl.root, err = insert(avl.root, key, value)
	return err
}

func (avl *AVLTree) Get(key string) (interface{}, error) {
	node, err := getNode(avl.root, key)
	if err != nil {
		return nil, err
	}
	return node.value, nil
}

func (avl *AVLTree) GetRange(minValue, maxValue string) ([]string, error) {
	var result []string
	var getRangeHelper func(node *Node, minValue, maxValue string)
	getRangeHelper = func(node *Node, minValue, maxValue string) {
		if node == nil {
			return
		}
		if node.key >= minValue {
			getRangeHelper(node.left, minValue, maxValue)
		}
		if node.key >= minValue && node.key <= maxValue {
			result = append(result, node.key)
		}
		if node.key <= maxValue {
			getRangeHelper(node.right, minValue, maxValue)
		}
	}
	getRangeHelper(avl.root, minValue, maxValue)
	return result, nil
}

func (avl *AVLTree) Update(key string, value interface{}) error {
	node, err := getNode(avl.root, key)
	if err != nil {
		return err
	}
	node.value = value
	return nil
}

func (avl *AVLTree) Remove(key string) error {
	var err error
	avl.root, err = deleteNode(avl.root, key)
	return err
}

func (avl *AVLTree) SaveToFile(filename string) error {
	data, err := json.Marshal(avl)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, data, 0644)
}

func (avl *AVLTree) LoadFromFile(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, avl)
}

func height(node *Node) int {
	if node == nil {
		return 0
	}
	return node.height
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// rightRotate выполняет правое вращение поддерева с корнем y
func rightRotate(y *Node) *Node {
	x := y.left
	T2 := x.right

	// Выполняем вращение
	x.right = y
	y.left = T2

	// Обновляем высоты
	y.height = max(height(y.left), height(y.right)) + 1
	x.height = max(height(x.left), height(x.right)) + 1

	// Возвращаем новый корень
	return x
}

// leftRotate выполняет левое вращение поддерева с корнем x
func leftRotate(x *Node) *Node {
	y := x.right
	T2 := y.left

	// Выполняем вращение
	y.left = x
	x.right = T2

	// Обновляем высоты
	x.height = max(height(x.left), height(x.right)) + 1
	y.height = max(height(y.left), height(y.right)) + 1

	// Возвращаем новый корень
	return y
}

// getBalance получает балансирующий фактор узла
func getBalance(node *Node) int {
	if node == nil {
		return 0
	}
	return height(node.left) - height(node.right)
}

func insert(node *Node, key string, value interface{}) (*Node, error) {
	if node == nil {
		return &Node{key: key, value: value, height: 1}, nil
	}

	if key < node.key {
		var err error
		node.left, err = insert(node.left, key, value)
		if err != nil {
			return nil, err
		}
	} else if key > node.key {
		var err error
		node.right, err = insert(node.right, key, value)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("Элемент с таким ключом уже существует!")
	}

	node.height = 1 + max(height(node.left), height(node.right))

	balance := getBalance(node)

	if balance > 1 && key < node.left.key {
		return rightRotate(node), nil
	}

	if balance < -1 && key > node.right.key {
		return leftRotate(node), nil
	}

	if balance > 1 && key > node.left.key {
		node.left = leftRotate(node.left)
		return rightRotate(node), nil
	}

	if balance < -1 && key < node.right.key {
		node.right = rightRotate(node.right)
		return leftRotate(node), nil
	}

	return node, nil
}

func minValueNode(node *Node) *Node {
	current := node
	for current.left != nil {
		current = current.left
	}
	return current
}

func deleteNode(root *Node, key string) (*Node, error) {
	if root == nil {
		return root, errors.New("Элемент не найден!")
	}

	if key < root.key {
		var err error
		root.left, err = deleteNode(root.left, key)
		if err != nil {
			return nil, err
		}
	} else if key > root.key {
		var err error
		root.right, err = deleteNode(root.right, key)
		if err != nil {
			return nil, err
		}
	} else {
		if (root.left == nil) || (root.right == nil) {
			var temp *Node
			if root.left != nil {
				temp = root.left
			} else {
				temp = root.right
			}

			if temp == nil {
				temp = root
				root = nil
			} else {
				*root = *temp
			}
		} else {
			temp := minValueNode(root.right)
			root.key = temp.key
			root.value = temp.value
			var err error
			root.right, err = deleteNode(root.right, temp.key)
			if err != nil {
				return nil, err
			}
		}
	}

	if root == nil {
		return root, nil
	}

	root.height = max(height(root.left), height(root.right)) + 1

	balance := getBalance(root)

	if balance > 1 && getBalance(root.left) >= 0 {
		return rightRotate(root), nil
	}

	if balance > 1 && getBalance(root.left) < 0 {
		root.left = leftRotate(root.left)
		return rightRotate(root), nil
	}

	if balance < -1 && getBalance(root.right) <= 0 {
		return leftRotate(root), nil
	}

	if balance < -1 && getBalance(root.right) > 0 {
		root.right = rightRotate(root.right)
		return leftRotate(root), nil
	}

	return root, nil
}

func getNode(node *Node, key string) (*Node, error) {
	if node == nil {
		return nil, errors.New("Элемент не найден!")
	}

	if key < node.key {
		return getNode(node.left, key)
	} else if key > node.key {
		return getNode(node.right, key)
	} else {
		return node, nil
	}
}

type AVLCollection struct {
	tree *AVLTree
}

func NewAVLCollection() *AVLCollection {
	return &AVLCollection{
		tree: NewAVLTree(),
	}
}

func (avl *AVLCollection) Insert(key string, value interface{}) error {
	var err error
	avl.tree.root, err = insert(avl.tree.root, key, value)
	return err
}

func (avl *AVLCollection) Get(key string) (interface{}, error) {
	node, err := getNode(avl.tree.root, key)
	if err != nil {
		return nil, err
	}
	return node.value, nil
}

func (avl *AVLCollection) GetRange(minValue, maxValue string) ([]string, error) {
	var result []string
	var getRangeHelper func(node *Node, minValue, maxValue string)
	getRangeHelper = func(node *Node, minValue, maxValue string) {
		if node == nil {
			return
		}
		if node.key >= minValue {
			getRangeHelper(node.left, minValue, maxValue)
		}
		if node.key >= minValue && node.key <= maxValue {
			result = append(result, node.key) // конвертируем node.key в строку
		}
		if node.key <= maxValue {
			getRangeHelper(node.right, minValue, maxValue)
		}
	}
	getRangeHelper(avl.tree.root, minValue, maxValue)
	return result, nil
}

func (avl *AVLCollection) Update(key string, value interface{}) error {
	node, err := getNode(avl.tree.root, key)
	if err != nil {
		return err
	}
	node.value = value
	return nil
}

func (avl *AVLCollection) Remove(key string) error {
	var err error
	avl.tree.root, err = deleteNode(avl.tree.root, key)
	return err
}

func (avl *AVLCollection) SaveToFile(filename string) error {
	data, err := json.Marshal(avl)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, data, 0644)
}

func (avl *AVLCollection) LoadFromFile(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, avl)
}
