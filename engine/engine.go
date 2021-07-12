package engine

import (
	"bytes"
	"github.com/awesome-cap/dkv/ptl"
	"github.com/awesome-cap/hashmap"
	"io"
	"time"
)

type Engine struct {
	string *hashmap.HashMap
}

func (e *Engine) Get(key string) (string, bool) {
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
	return e.string.LogicDel(key)
}

func (e *Engine) Marshal() []byte {
	buf := &bytes.Buffer{}
	// Marshal string
	stringBuf := &bytes.Buffer{}
	e.string.Foreach(func(e *hashmap.Entry) {
		key, value := e.Key().(string), e.Value().(string)
		_ = ptl.WriteUint16(stringBuf, uint16(len(key)))
		buf.WriteString(key)
		_ = ptl.WriteUint64(stringBuf, uint64(len(value)))
		buf.WriteString(value)
	})

	_ = ptl.WriteUint16(buf, 6)
	buf.WriteString("string")
	_ = ptl.WriteUint64(buf, uint64(stringBuf.Len()))
	buf.Write(stringBuf.Bytes())
	return buf.Bytes()
}

func (e *Engine) UnMarshal(reader io.Reader) error {
	for {
		typeSize, err := ptl.ReadUint16(reader)
		if err != nil {
			return err
		}
		typeBytes, err := ptl.ReadBytes(reader, int(typeSize))
		if err != nil {
			return err
		}
		dataSize, err := ptl.ReadUint64(reader)
		if err != nil {
			return err
		}
		switch string(typeBytes) {
		case "string":
			str := hashmap.New()
			readSize := 0
			for readSize < int(dataSize) {
				keySize, err := ptl.ReadUint16(reader)
				if err != nil {
					return err
				}
				keyData, err := ptl.ReadBytes(reader, int(keySize))
				if err != nil {
					return err
				}
				valueSize, err := ptl.ReadUint64(reader)
				if err != nil {
					return err
				}
				valueData, err := ptl.ReadBytes(reader, int(valueSize))
				if err != nil {
					return err
				}
				str.Set(string(keyData), string(valueData))
				readSize += 2 + 8 + int(keySize) + int(valueSize)
			}
			e.string = str
		}
	}
}
