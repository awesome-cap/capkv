/**
2 * @Author: Nico
3 * @Date: 2021/6/10 0:45
4 */
package opts

import "errors"

var(
	KeyNotExist = errors.New("key not exist. ")
)

type kv interface {
	Get(key interface{}) interface{}
	Set(key interface{}, value interface{}) uint
	Del(key interface{}) uint
	Exist(key interface{}) bool
}

type dkv struct {
	kvs map[interface{}]interface{}
}

func New() *dkv{
	return &dkv{}
}

func (d *dkv) Get(key interface{}) interface{}{
	return d.kvs[key]
}

func (d *dkv) Set(key interface{}, value interface{}) uint{
	if d.Exist(key) && d.kvs[key] == value{
		return 0
	}
	d.kvs[key] = value
	return 1
}

func (d *dkv) Del(key interface{}) uint{
	if ! d.Exist(key) {
		return 0
	}
	delete(d.kvs, key)
	return 1
}

func (d *dkv) Exist(key interface{}) bool{
	_, exist := d.kvs[key]
	return exist
}





