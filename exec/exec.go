package exec

import (
	"errors"
	"fmt"
	"github.com/awesome-cap/hashmap"
	"strings"
)

var (
	GetHandler = getHandler{}
	SetHandler = setHandler{}
	DelHandler = delHandler{}
)

type Executor struct {
	engine *Engine
	handlers map[string]Handler
}

type Handler interface {
	handle(e *Engine, args []string) ([]string, error)
	size() int
}

func New() *Executor {
	return &Executor{
		engine: &Engine{
			string: hashmap.New(),
		},
		handlers: map[string]Handler{},
	}
}

func assertArgsSize(args []string, s int) error{
	if len(args) < s {
		return errors.New(fmt.Sprintf("Args size err, expect %d", s))
	}
	return nil
}

func (e *Executor) Exec(args []string) ([]string, error){
	err := assertArgsSize(args, 1)
	if err != nil {
		return nil, err
	}
	args[0] = strings.ToLower(args[0])
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

func (e *Executor) RegistryHandler(cmd string, h Handler) {
	e.handlers[strings.ToLower(cmd)] = h
}

type getHandler struct {}

func (h getHandler) handle(e *Engine, args []string) ([]string, error){
	if v, ok := e.Get(args[1]); ok {
		return []string{v}, nil
	}
	return nil, errors.New(fmt.Sprintf("%s not exist. ", args[1]))
}

func (h getHandler) size() int {return 2}

type setHandler struct {}

func (h setHandler) handle(e *Engine, args []string) ([]string, error){
	nx := false
	for i := 3; i < len(args); i ++{
		if strings.ToUpper(args[i]) == "NX" {
			nx = true
		}
	}
	if e.Set(args[1], args[2], 0, nx){
		return []string{"1"}, nil
	}
	return []string{"0"}, nil
}

func (h setHandler) size() int {return 3}

type delHandler struct {}

func (h delHandler) handle(e *Engine, args []string) ([]string, error){
	if e.Del(args[1]) {
		return []string{"1"}, nil
	}
	return []string{"0"}, nil
}

func (h delHandler) size() int {return 2}