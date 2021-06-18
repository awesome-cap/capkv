package stroage

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

//const testBatchNum1 = 10000 * 10000 * 1
const testBatchNum1 = 10000 * 1000
const testBatchNum2 = 1000

func TestNewMultiMap(t *testing.T) {
	mm := NewMultiMap(1000 * 10000)
	for i := 0; i < testBatchNum1; i ++ {
		mm.Set(i, i)
	}
	start := time.Now().UnixNano()
	for i := 0; i < testBatchNum1; i ++ {
		mm.Get(i)
	}
	end := time.Now().UnixNano()
	fmt.Println("read", end - start)
}

func TestValues(t *testing.T) {
	mm := values{
		data: map[key]value{},
	}
	start := time.Now().UnixNano()
	for i := 0; i < testBatchNum1; i ++ {
		mm.set(i, i)
	}
	end := time.Now().UnixNano()
	fmt.Println("map rwmutex write", end - start)
	start = time.Now().UnixNano()
	for i := 0; i < testBatchNum1; i ++ {
		mm.get(i)
	}
	end = time.Now().UnixNano()
	fmt.Println("map rwmutex read", end - start)
}

func TestMap(t *testing.T) {
	mm := sync.Map{}
	start := time.Now().UnixNano()
	for i := 0; i < testBatchNum1; i ++ {
		mm.Store(i, i)
	}
	end := time.Now().UnixNano()
	fmt.Println("sync.Map write", end - start)
	start = time.Now().UnixNano()
	for i := 0; i < testBatchNum1; i ++ {
		mm.Load(i)
	}
	end = time.Now().UnixNano()
	fmt.Println("sync.Map read", end - start)
}

func TestConcurrent(t *testing.T){
	mm := NewMultiMap(10 * 10000)
	go func() {
		for i := 0; i < testBatchNum1; i ++{
			mm.Set(i, i)
		}
	}()
	for i := 0; i < testBatchNum1; i ++{
		mm.Get(i)
	}
	t.Log(len(mm.stables))
}