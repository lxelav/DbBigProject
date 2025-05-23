package main

import (
	"fmt"
	"time"
)

type Command interface {
	Execute(dataExists *bool, dataToModify *TData)
}

type TData struct {
	Key       string
	Value     interface{}
	Timestamp time.Time
}

type InsertCommand struct {
	InitialVersion TData
}

func (c *InsertCommand) Execute(dataExists *bool, dataToModify *TData) {
	if *dataExists {
		panic("attempt to insert already existent data")
	}
	*dataToModify = c.InitialVersion
	*dataExists = true
	fmt.Printf("Вставка данных: ключ = %s, значение = %v, время = %s\n", dataToModify.Key, dataToModify.Value, time.Now().Format(time.RFC3339))
}

type UpdateCommand struct {
	UpdateExpression string
}

func (c *UpdateCommand) Execute(dataExists *bool, dataToModify *TData) {
	if !*dataExists {
		panic("attempt to modify non-existent data")
	}
	dataToModify.Value = c.UpdateExpression
	dataToModify.Timestamp = time.Now()
	fmt.Printf("Обновление данных: ключ = %s, новое значение = %v, время = %s\n", dataToModify.Key, dataToModify.Value, dataToModify.Timestamp.Format(time.RFC3339))
}

type DisposeCommand struct{}

func (c *DisposeCommand) Execute(dataExists *bool, dataToModify *TData) {
	if !*dataExists {
		panic("attempt to dispose non-existent data")
	}
	*dataExists = false
	fmt.Printf("Удаление данных: ключ = %s, время = %s\n", dataToModify.Key, time.Now().Format(time.RFC3339))
}

type ChainOfResponsibilityHandler struct {
	Command                 Command
	DateTimeActivityStarted int64
	NextHandler             *ChainOfResponsibilityHandler
}

func (h *ChainOfResponsibilityHandler) Handle(dataExists *bool, dataToModify *TData, dateTimeTarget int64) {
	if dateTimeTarget <= h.DateTimeActivityStarted {
		return
	}
	h.Command.Execute(dataExists, dataToModify)
	if h.NextHandler != nil {
		h.NextHandler.Handle(dataExists, dataToModify, dateTimeTarget)
	}
}

type ChainOfResponsibility struct {
	FirstHandler *ChainOfResponsibilityHandler
	LastHandler  *ChainOfResponsibilityHandler
}

func (c *ChainOfResponsibility) AddHandler(command Command) {
	dateTimeActivityStarted := time.Now().Unix()
	addedHandler := &ChainOfResponsibilityHandler{
		Command:                 command,
		DateTimeActivityStarted: dateTimeActivityStarted,
		NextHandler:             nil,
	}
	if c.LastHandler == nil {
		c.FirstHandler = addedHandler
		c.LastHandler = addedHandler
	} else {
		c.LastHandler.NextHandler = addedHandler
		c.LastHandler = addedHandler
	}
}
