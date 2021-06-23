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
	overload float32
	loadFactory int
	size int
	capacity int
}

type Node struct {
	sync.RWMutex
	header *Entry
	tail *Entry
	size int
}

type Entry struct {
	K key
	V value
	hash int
	next *Entry
}

func NewHashMap() *HashMap{
	defaultCapacity := 16
	return &HashMap{
		nodes: make([]*Node, defaultCapacity),
		overload: 0.75,
		loadFactory: 60,
		capacity: defaultCapacity,
	}
}

func (m *HashMap) hash(k key) int {
	hash := uint32(2166136261)
	const prime32 = uint32(16777619)
	keyLength := len(k)
	for i := 0; i < keyLength; i++ {
		hash *= prime32
		hash ^= uint32(k[i])
	}
	return int(hash)
}

func (m *HashMap) index(hash int) int{
	return hash & (m.capacity - 1)
}

func (m *HashMap) Set(k key, v value) value {
	if m.size > int(float32(len(m.nodes) * m.loadFactory) * m.overload){
		m.resize()
	}

	h := m.hash(k)
	i := m.index(h)
	n := m.nodes[i]
	e := &Entry{K: k, V: v, hash: h}
	if n == nil {
		m.nodes[i] = &Node{header: e, tail: e}
		m.size ++
		m.nodes[i].size ++
		return v
	}
	if n.header == nil {
		n.header = e
		n.tail = e
		m.size ++
		n.size ++
		return v
	}
	next := n.header
	if next.K == k{
		next.V = v
		return v
	}
	for next.next != nil {
		next = next.next
		if next.K == k{
			next.V = v
			return v
		}
	}
	next.next = e
	n.tail = e
	m.size ++
	n.size ++
	return v
}

func (m *HashMap) resize() {
	m.capacity = m.capacity * 2
	nodes := make([]*Node, m.capacity)
	for _, old := range m.nodes {
		if old != nil {
			next := old.header
			for next != nil {
				i := m.index(next.hash)
				n := nodes[i]
				if n == nil {
					nodes[i] = &Node{header: next, tail: next}
					nodes[i].size ++
				}else{
					n.tail.next = next
					n.tail = next
					n.size ++
				}
				temp := next.next
				next.next = nil
				next = temp
			}
		}
	}
	m.nodes = nodes
}

func (m *HashMap) Get(k key) (value, bool) {
	i := m.index(m.hash(k))
	n := m.nodes[i]
	if n != nil {
		next := n.header
		for next != nil {
			if next.K == k {
				return next.V, true
			}
			next = next.next
		}
	}
	return nil, false
}

func (m *HashMap) Del(k key) bool {
	i := m.index(m.hash(k))
	n := m.nodes[i]
	if n != nil && n.header != nil {
		if n.header.K == k {
			n.header = nil
			n.tail = nil
			m.size --
			n.size --
			return true
		}
		prev := n.header
		next := prev.next
		for next != nil {
			if next.K == k {
				prev.next = nil
				n.tail = prev
				m.size --
				n.size --
				return true
			}
			prev = next
			next = prev.next
		}
	}
	return false
}