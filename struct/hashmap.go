package _struct

import "sync"

type HashMap struct {
	nodes []*Node
	loadFactor float32
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
	K string
	V interface{}
	hash int
	next *Entry
	prev *Entry
}

func NewHashMap() *HashMap{
	defaultCapacity := 16
	return &HashMap{
		nodes: initNodes(defaultCapacity),
		loadFactor: 0.75,
		capacity: defaultCapacity,
	}
}

func initNodes(capacity int) (nodes []*Node){
	nodes = make([]*Node, capacity)
	for i := 0; i < capacity; i ++{
		nodes[i] = &Node{}
	}
	return
}

func (m *HashMap) hash(k string) int {
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

func (m *HashMap) Set(k string, v interface{}) interface{} {
	if m.size > int(float32(len(m.nodes)) * m.loadFactor * 3){
		m.resize()
	}
	h := m.hash(k)
	m.setNodeEntry(m.nodes[m.index(h)], &Entry{K: k, V: v, hash: h})
	return v
}

func (m *HashMap) setNodeEntry(n *Node, e *Entry){
	if n.header == nil {
		n.header = e
		n.tail = e
	}else if e.hash <= n.header.hash {
		e.next = n.header
		n.header.prev = e
		n.header = e
	}else if e.hash >= n.tail.hash{
		n.tail.next = e
		e.prev = n.tail
		n.tail = e
	}else{
		next := n.header
		for next != nil && next.hash < e.hash{
			if next.K == e.K{
				next.V = e.V
				return
			}
			next = next.next
		}
		if next != nil && next.prev != nil {
			next.prev.next = e
			e.prev = next.prev
			e.next = next
			next.prev = e
		}
	}
	n.size ++
	m.size ++
}

func (m *HashMap) resize() {
	m.capacity = m.capacity * 2
	nodes := initNodes(m.capacity)
	for _, old := range m.nodes {
		next := old.header
		for next != nil {
			m.setNodeEntry(nodes[m.index(next.hash)], next.clone())
			next = next.next
		}
	}
	m.nodes = nodes
}

func (m *HashMap) getNodeEntry(n *Node, k string) *Entry {
	if n != nil {
		h := n.header
		t := n.tail
		for h != nil && t != nil && t.hash >= h.hash {
			if h.K == k {
				return h
			}
			if t.K == k {
				return t
			}
			h = h.next
			t = t.prev
		}
	}
	return nil
}

func (m *HashMap) Get(k string) (interface{}, bool) {
	n := m.nodes[m.index(m.hash(k))]
	if n != nil {
		e := m.getNodeEntry(n, k)
		if e != nil {
			return e.V, true
		}
	}
	return nil, false
}

func (m *HashMap) Del(k string) bool {
	n := m.nodes[m.index(m.hash(k))]
	if n != nil{
		e := m.getNodeEntry(n, k)
		if e != nil {
			if e.prev == nil && e.next == nil{
				n.header = nil
				n.tail = nil
			}else if e.prev == nil {
				n.header = e.next
				e.next.prev = nil
			}else if e.next == nil {
				n.tail = e.prev
				e.prev.next = nil
			}else{
				e.prev.next = e.next
				e.next.prev = e.prev
			}
			m.size --
			n.size --
		}
	}
	return false
}

func (e *Entry) clone() *Entry{
	return &Entry{
		K: e.K,
		V: e.V,
		hash: e.hash,
	}
}