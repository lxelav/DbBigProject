package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
)

type StringPools struct {
	Pools map[string]string
	mu    sync.Mutex
}

var instance *StringPools
var once sync.Once

func GetStringPools() *StringPools {
	once.Do(func() {
		instance = &StringPools{
			Pools: make(map[string]string),
		}
	})
	return instance
}

func (sp *StringPools) Get(str string) string {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	if s, exists := sp.Pools[str]; exists {
		return s
	}
	sp.Pools[str] = str
	return str
}

type Tree interface {
	Insert(key string, value interface{}) error
	Get(key string) (interface{}, error)
	GetRange(minValue, maxValue string) ([]string, error)
	Update(key string, value interface{}) error
	Remove(key string) error
	SaveToFile(filename string) error
}

type TreeCollection struct {
	Tree Tree
}

func NewTreeCollection(treeType string) *TreeCollection {
	var tree Tree
	switch treeType {
	case "avl":
		tree = Tree(NewAVLTree())
	case "redblack":
		tree = Tree(NewRedBlackTree())
	case "btree":
		tree = Tree(NewBTree())
	case "map":
		tree = NewMapCollection()
	default:
		tree = NewMapCollection()
	}
	return &TreeCollection{Tree: tree}
}

func (tc *TreeCollection) Insert(key string, value interface{}) error {
	return tc.Tree.Insert(key, value)
}

func (tc *TreeCollection) Get(key string) (interface{}, error) {
	return tc.Tree.Get(key)
}

func (tc *TreeCollection) GetRange(minValue, maxValue string) ([]string, error) {
	return tc.Tree.GetRange(minValue, maxValue)
}

func (tc *TreeCollection) Update(key string, value interface{}) error {
	return tc.Tree.Update(key, value)
}

func (tc *TreeCollection) Remove(key string) error {
	return tc.Tree.Remove(key)
}

func (tc *TreeCollection) SaveToFile(filename string) error {
	return tc.Tree.SaveToFile(filename)
}

type MapCollection struct {
	Data map[string]interface{}
}

func NewMapCollection() *MapCollection {
	return &MapCollection{
		Data: make(map[string]interface{}),
	}
}

func (mc *MapCollection) Insert(key string, value interface{}) error {
	sp := GetStringPools()
	key = sp.Get(key)

	if _, exists := mc.Data[key]; exists {
		return errors.New("Элемент с таким ключом уже существует!")
	}
	mc.Data[key] = value
	fmt.Println("Элемент успешно добавлен с ключом", key)
	return nil
}

func (mc *MapCollection) Get(key string) (interface{}, error) {
	sp := GetStringPools()
	key = sp.Get(key)

	value, exists := mc.Data[key]
	if !exists {
		return nil, errors.New("Элемент не найден!")
	}
	return value, nil
}

func (mc *MapCollection) GetRange(minValue, maxValue string) ([]string, error) {
	sp := GetStringPools()
	minValue = sp.Get(minValue)
	maxValue = sp.Get(maxValue)

	var result []string
	for key := range mc.Data {
		if key >= minValue && key <= maxValue {
			result = append(result, key)
		}
	}
	return result, nil
}

func (mc *MapCollection) Update(key string, value interface{}) error {
	sp := GetStringPools()
	key = sp.Get(key)

	if _, exists := mc.Data[key]; !exists {
		return errors.New("Элемент не найден!")
	}
	mc.Data[key] = value
	fmt.Println("Значение элемента с ключом", key, "успешно обновлено.")
	return nil
}

func (mc *MapCollection) Remove(key string) error {
	sp := GetStringPools()
	key = sp.Get(key)

	if _, exists := mc.Data[key]; !exists {
		return errors.New("Элемент не найден!")
	}
	delete(mc.Data, key)
	return nil
}

func (mc *MapCollection) SaveToFile(filename string) error {
	data, err := json.Marshal(mc.Data)
	if err != nil {
		return err
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	return err
}

type AllPools struct {
	Pools map[string]*Pools
}

func InitPools() *AllPools {
	return &AllPools{
		Pools: make(map[string]*Pools),
	}
}

func (ap *AllPools) ShowAll() {
	fmt.Println("Текущие пулы и схемы:")
	for poolName, pool := range ap.Pools {
		fmt.Printf("Пул: %s\n", poolName)
		for schemaName := range pool.schema {
			fmt.Printf("  Схема: %s\n", schemaName)
		}
	}
}

func (ap *AllPools) AddPools(name string) {
	if _, exists := ap.Pools[name]; exists {
		fmt.Println("Пул с именем", name, "уже существует.")
	} else {
		ap.Pools[name] = NewPools()
		fmt.Println("Добавлен пул с именем", name)
	}
	ap.ShowAll()
}

func (ap *AllPools) RemovePools(name string) {
	if pool, exists := ap.Pools[name]; exists {
		for schemaName := range pool.schema {
			schema := pool.schema[schemaName]
			for collectionName := range schema.Collection {
				schema.RemoveCollection(collectionName)
			}
			pool.RemoveSchema(schemaName)
		}
		delete(ap.Pools, name)
		fmt.Println("Пул с именем", name, "удален.")
	} else {
		fmt.Println("Пул с именем", name, "не существует.")
	}
	ap.ShowAll()
}

func (ap *AllPools) GetPools(name string) (*Pools, error) {
	returnEl, ok := ap.Pools[name]
	if !ok {
		return nil, errors.New("Элемент не найден!")
	}
	return returnEl, nil
}

func (ap *AllPools) GetRange(minValue, maxValue string) ([]*Pools, error) {
	var result []*Pools
	for name, pool := range ap.Pools {
		if name >= minValue && name <= maxValue {
			result = append(result, pool)
		}
	}
	return result, nil
}

func (ap *AllPools) SaveToFile(filename string) error {
	for _, pool := range ap.Pools {
		if err := pool.SaveToFile(filename); err != nil {
			return err
		}
	}

	data, err := json.Marshal(ap)
	if err != nil {
		return err
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	return err
}

type Pools struct {
	schema map[string]*Schema
}

func NewPools() *Pools {
	return &Pools{
		schema: make(map[string]*Schema),
	}
}

func (p *Pools) GetSchema(schemaName string) (*Schema, error) {
	returnEl, ok := p.schema[schemaName]
	if !ok {
		return nil, errors.New("Элемент не найден!")
	}
	return returnEl, nil
}

func (p *Pools) AddSchema(name string) {
	if _, exists := p.schema[name]; exists {
		fmt.Println("Схема с именем", name, "уже существует в пуле.")
	} else {
		p.schema[name] = InitSchema()
		fmt.Println("Схема с именем", name, "добавлена в пул.")
	}
	p.ShowSchemas()
}

func (p *Pools) RemoveSchema(name string) {
	if schema, exists := p.schema[name]; exists {
		for collectionName := range schema.Collection {
			schema.RemoveCollection(collectionName)
		}
		// Удаляем схему из пула
		delete(p.schema, name)
		fmt.Println("Схема с именем", name, "удалена из пула.")
	} else {
		fmt.Println("Схема с именем", name, "не найдена в пуле.")
	}
	p.ShowSchemas()
}

func (p *Pools) ShowSchemas() {
	fmt.Println("Текущие схемы в пуле:")
	for schemaName := range p.schema {
		fmt.Printf("  Схема: %s\n", schemaName)
	}
}

func (p *Pools) SaveToFile(filename string) error {
	type tempStruct struct {
		Schema map[string]*Schema `json:"schema"`
	}

	temp := &tempStruct{
		Schema: p.schema,
	}

	for _, schema := range p.schema {
		if err := schema.SaveToFile(filename); err != nil {
			return err
		}
	}

	data, err := json.Marshal(temp)
	if err != nil {
		return err
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	return err
}

type Schema struct {
	Collection map[string]TreeCollection
}

func InitSchema() *Schema {
	return &Schema{
		Collection: make(map[string]TreeCollection),
	}
}

func (s *Schema) GetCollection(name string) (TreeCollection, error) {
	returnEl, ok := s.Collection[name]
	if !ok {
		return TreeCollection{}, errors.New("Элемент не найден!")
	}
	return returnEl, nil
}

func (p *Pools) AddCollection(schemaName, collectionName string, collection TreeCollection) error {
	schema, err := p.GetSchema(schemaName)
	if err != nil {
		return err
	}

	if _, exists := schema.Collection[collectionName]; exists {
		return errors.New("Коллекция с таким именем уже существует!")
	}

	schema.Collection[collectionName] = collection
	fmt.Printf("Коллекция с именем %s добавлена в схему %s в пуле\n", collectionName, schemaName)
	schema.ShowCollections()
	return nil
}

func (s *Schema) RemoveCollection(name string) {
	if _, exists := s.Collection[name]; exists {
		delete(s.Collection, name)
		fmt.Println("Коллекция с именем", name, "удалена из схемы.")
	} else {
		fmt.Println("Коллекция с именем", name, "не найдена в схеме.")
	}
	s.ShowCollections()
}

func (s *Schema) ShowCollections() {
	fmt.Println("Текущие коллекции в схеме:")
	for collectionName := range s.Collection {
		fmt.Printf("  Коллекция: %s\n", collectionName)
	}
}

func (s *Schema) SaveToFile(filename string) error {
	for _, collection := range s.Collection {
		if err := collection.SaveToFile(filename); err != nil {
			return err
		}
	}

	data, err := json.Marshal(s)
	if err != nil {
		return err
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	return err
}
