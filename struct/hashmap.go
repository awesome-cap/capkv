package _struct

import (
	"sync"
	"sync/atomic"
)

type HashMap struct {
	sync.RWMutex

	nodes []*Node
	loadFactor float64
	capacity int
	size int64
}

type Node struct {
	sync.Mutex
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

func hash(k string) int {
	hash := uint32(2166136261)
	const prime32 = uint32(16777619)
	keyLength := len(k)
	for i := 0; i < keyLength; i++ {
		hash *= prime32
		hash ^= uint32(k[i])
	}
	return int(hash)
}

func indexOf(hash int, capacity int) int{
	return hash & (capacity - 1)
}

func (m *HashMap) Set(k string, v interface{}) interface{} {
	m.resize()
	h := hash(k)
	n := m.nodes[indexOf(h, m.capacity)]
	n.Lock()
	defer n.Unlock()
	if m.setNodeEntry(n, &Entry{K: k, V: v, hash: h}) {
		n.size ++
		atomic.AddInt64(&m.size, 1)
	}
	return v
}

func (m *HashMap) setNodeEntry(n *Node, e *Entry) bool{
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
				return false
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
	return true
}

func (m *HashMap) dilate() bool {
	return m.size > int64(float64(len(m.nodes)) * m.loadFactor * 3)
}

func (m *HashMap) resize() {
	if m.dilate() {
		m.Lock()
		defer m.Unlock()
		if m.dilate() {
			m.doResize()
		}
	}
}

func (m *HashMap) doResize()  {
	capacity := m.capacity * 2
	nodes := initNodes(capacity)
	size := int64(0)
	for _, old := range m.nodes {
		next := old.header
		for next != nil {
			newNode := nodes[indexOf(next.hash, capacity)]
			if m.setNodeEntry(newNode, next.clone()) {
				newNode.size ++
				size ++
			}
			next = next.next
		}
	}
	m.capacity = capacity
	m.nodes = nodes
	m.size = size
}

func (m *HashMap) getNodeEntry(n *Node, k string) *Entry {
	if n != nil {
		next := n.header
		h := hash(k)
		for next != nil && next.hash <= h {
			if next.K == k {
				return next
			}
			next = next.next
		}
	}
	return nil
}

func (m *HashMap) Get(k string) (interface{}, bool) {
	n := m.nodes[indexOf(hash(k), m.capacity)]
	if n != nil {
		e := m.getNodeEntry(n, k)
		if e != nil {
			return e.V, true
		}
	}
	return nil, false
}

func (m *HashMap) Del(k string) bool {
	m.RLock()
	defer m.RUnlock()

	n := m.nodes[indexOf(hash(k), m.capacity)]
	n.Lock()
	defer n.Unlock()
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
		n.size --
		atomic.AddInt64(&m.size, -1)
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