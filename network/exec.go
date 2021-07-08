package network

import (
	"errors"
	"fmt"
	"github.com/awesome-cap/dkv/storage"
)

type Executor struct {
	engine *storage.Engine
	handlers map[string]Handler
}

type Handler interface {
	handle(e *storage.Engine, args []string) (interface{}, error)
	size() int
}

type GetHandler struct {}

func assertArgsSize(args []string, s int) error{
	if len(args) < s {
		return errors.New(fmt.Sprintf("Args size err, expect %d", s))
	}
	return nil
}

func (e *Executor) Exec(args []string) (interface{}, error){
	err := assertArgsSize(args, 1)
	if err != nil {
		return nil, err
	}
	handler, ok := e.handlers[args[0]]
	if ! ok {
		return nil, errors.New(fmt.Sprintf("Invalid cmd %s", args[0]))
	}
	err = assertArgsSize(args, handler.size())
	if err != nil {
		return nil, err
	}
	return handler.handle(e.engine, args)
}

func (h *GetHandler) handle(e *storage.Engine, args []string) (interface{}, error){
	if v, ok := e.Get(args[1]); ok {
		return v, nil
	}
	return nil, errors.New(fmt.Sprintf("%s not exist. ", args[1]))
}

func (h *GetHandler) size() int {
	return 2
}
