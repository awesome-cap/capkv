package _struct

import (
	"sync"
)

type (
	key string
	value interface{}
)

type HashMap struct {
	nodes []*Node
}

type Node struct {
	sync.RWMutex
	header []*Entry
}

type Entry struct {
	K key
	V value
	next *Entry
}

func (m *HashMap) hashCode(k key) int {
	var sum = 0
	for _, c := range k{
		if c < 0 {
			c = -c
		}
		sum += int(c)
	}
	return sum & len(m.nodes)
}

func (m *HashMap) Put(k Object, v Object){

}