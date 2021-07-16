package engine

import (
	"errors"
	"fmt"
	"strings"
)

var (
	Get = getHandler{}
	Set = setHandler{}
	Del = delHandler{}

	writeable = map[string]bool{
		"set": true, "del": true,
	}
)

type handler interface {
	handle(e *Engine, args []string) ([]string, error)
	size() int
	name() string
}

func assertArgsSize(args []string, s int) error {
	if len(args) < s {
		return errors.New(fmt.Sprintf("Args size err, expect %d", s))
	}
	return nil
}

type getHandler struct{}

func (h getHandler) handle(e *Engine, args []string) ([]string, error) {
	if v, ok := e.Get(args[1]); ok {
		return []string{v}, nil
	}
	return []string{""}, nil
}

func (h getHandler) size() int    { return 2 }
func (h getHandler) name() string { return "get" }

type setHandler struct{}

func (h setHandler) handle(e *Engine, args []string) ([]string, error) {
	nx := false
	for i := 3; i < len(args); i++ {
		if strings.ToUpper(args[i]) == "NX" {
			nx = true
		}
	}
	if e.Set(args[1], args[2], 0, nx) {
		return []string{"1"}, nil
	}
	return []string{"0"}, nil
}

func (h setHandler) size() int    { return 3 }
func (h setHandler) name() string { return "set" }

type delHandler struct{}

func (h delHandler) handle(e *Engine, args []string) ([]string, error) {
	if e.Del(args[1]) {
		return []string{"1"}, nil
	}
	return []string{"0"}, nil
}

func (h delHandler) size() int    { return 2 }
func (h delHandler) name() string { return "del" }
