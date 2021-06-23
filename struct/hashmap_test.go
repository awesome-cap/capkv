package _struct

import (
	"fmt"
	"github.com/orcaman/concurrent-map"
	"log"
	"strconv"
	"testing"
	"time"
)

func TestNewHashMap(t *testing.T) {
	hm := NewHashMap()
	batch := 1000000
	start := time.Now().UnixNano()
	for i := 0; i < batch; i ++ {
		hm.Set(strconv.Itoa(i), i)
	}
	end := time.Now().UnixNano()
	fmt.Println("set:", (end - start) / 1e6)
	for i := 0; i < batch; i ++ {
		v, e := hm.Get(strconv.Itoa(i))
		if ! e || v != i{
			log.Fatal("data err " + strconv.Itoa(i))
		}
	}
}

func TestNewMap(t *testing.T) {
	hm := map[string]int{}
	batch := 1000000
	start := time.Now().UnixNano()
	for i := 0; i < batch; i ++ {
		hm[strconv.Itoa(i)] = i
	}
	end := time.Now().UnixNano()
	fmt.Println("set:", (end - start) / 1e6)
}

func TestCurrentNewMap(t *testing.T) {
	hm := cmap.New()
	batch := 1000000
	start := time.Now().UnixNano()
	for i := 0; i < batch; i ++ {
		hm.Set(strconv.Itoa(i), i)
	}
	end := time.Now().UnixNano()
	fmt.Println("set:", (end - start) / 1e6)
}

