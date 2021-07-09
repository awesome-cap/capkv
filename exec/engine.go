package exec

import (
	"github.com/awesome-cap/hashmap"
	"time"
)

type Engine struct {
	exec *Executor
	string *hashmap.HashMap
}

func (e *Engine) Get(key string) (string, bool){
	v, ok := e.string.Get(key)
	if ok {
		return v.(string), ok
	}
	return "", ok
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