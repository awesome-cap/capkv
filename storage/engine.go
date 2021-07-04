package storage

import (
	"github.com/awesome-cap/hashmap"
	"time"
)

type Engine struct {
	string *hashmap.HashMap
}

func New() *Engine {
	return &Engine{
		string: hashmap.New(),
	}
}

func (e *Engine) Get(key string) (string, bool){
	v, ok := e.string.Get(key)
	return v.(string), ok
}

func (e *Engine) Set(key, value string, ex time.Duration, nx bool) bool {
	if nx {
		return e.string.SetNX(key, value)
	}
	e.string.Set(key, value)
	return true
}

func (e *Engine) Del(key string) bool {
	return e.string.Del(key)
}