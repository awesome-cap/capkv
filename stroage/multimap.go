package stroage

import (
	"sync"
)

type (
	key interface {}
	value interface {}
	values struct {
		sync.RWMutex
		data map[key]value
	}
	multiMap struct {
		sync.Mutex
		stables []*values
		active *values
		keySizeThreshold int
	}
	collector struct {
		results []result
	}
	result struct {
		value value
		exist bool
	}
)

func NewMultiMap(keySizeThreshold int) *multiMap{
	return &multiMap{
		active: newValues(keySizeThreshold),
		keySizeThreshold: keySizeThreshold,
	}
}

func newValues(capacity int) *values{
	return &values{
		data: make(map[key]value, capacity),
	}
}

func (v *values) set(key key, value value) bool {
	v.Lock()
	defer v.Unlock()
	_, ok := v.data[key]
	v.data[key] = value
	return ! ok
}

func (v *values) get(key key) (value, bool) {
	v.RLock()
	defer v.RUnlock()
	value, ok := v.data[key]
	return value, ok
}

func (v *values) delete(key key) bool {
	v.Lock()
	defer v.Unlock()
	_, ok := v.data[key]
	delete(v.data, key)
	return ok
}

func (v *values) keySize() int{
	return len(v.data)
}

func (m *multiMap) Set(key key, value value) bool {
	if m.active.keySize() > m.keySizeThreshold {
		m.Lock()
		defer m.Unlock()
		m.stables = append(m.stables, m.active)
		m.active = newValues(m.keySizeThreshold)
	}
	return m.active.set(key, value)
}

func (m *multiMap) Get(key key) (value, bool) {
	stablesSize := len(m.stables)
	ct := newCollector(1 + stablesSize)
	ct.collect(m.active, key, 0)
	for i := stablesSize - 1; i >= 0; i -- {
		ct.collect(m.stables[i], key, stablesSize - i)
	}
	return ct.get()
}

func (m *multiMap) Delete(key key) bool {
	return m.active.delete(key)
}

func newCollector(size int) *collector{
	ct := &collector{
		results: make([]result, size),
	}
	return ct
}

func (c *collector) collect(values *values, key key, index int) {
	value, ok := values.get(key)
	c.results[index] = result{
		value: value,
		exist: ok,
	}
}

func (c *collector) get() (value, bool){
	for _, result := range c.results {
		if result.exist {
			return result.value, true
		}
	}
	return nil, false
}
