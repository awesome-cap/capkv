package engine

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/awesome-cap/dkv/config"
	"github.com/awesome-cap/dkv/ptl"
	"github.com/awesome-cap/hashmap"
	"io"
	"strings"
	"time"
)

type Engine struct {
	lsn      uint64
	storage  *Storage
	handlers map[string]handler

	string *hashmap.HashMap
}

func New(conf config.Config) (*Engine, error) {
	s, err := newStorage(conf.Storage)
	if err != nil {
		return nil, err
	}
	e := &Engine{
		storage:  s,
		handlers: map[string]handler{},
		string:   hashmap.New(),
	}
	_ = s.loadDB(e)
	_ = s.loadLog(e)
	//s.startDaemon(e)
	return e, nil
}

func (e *Engine) Registry(h handler) {
	e.handlers[strings.ToLower(h.name())] = h
}

func (e *Engine) Exec(args []string) ([]string, error) {
	err := assertArgsSize(args, 1)
	if err != nil {
		return nil, err
	}
	args[0] = strings.ToLower(args[0])
	handler, ok := e.handlers[args[0]]
	if !ok {
		return nil, errors.New(fmt.Sprintf("Invalid cmd %s", args[0]))
	}
	err = assertArgsSize(args, handler.size())
	if err != nil {
		return nil, err
	}
	if writeable[args[0]] {
		e.lsn, err = e.storage.logging(args)
		if err != nil {
			return nil, err
		}
	}
	return handler.handle(e, args)
}

func (e *Engine) exec(args []string) {
	if handler, ok := e.handlers[args[0]]; ok {
		_, _ = handler.handle(e, args)
	}
}

func (e *Engine) Get(key string) (string, bool) {
	v, ok := e.string.Get(key)
	if ok {
		return v.(string), ok
	}
	v, ok = e.storage.Get(key)
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
		stringBuf.WriteString(key)
		_ = ptl.WriteUint64(stringBuf, uint64(len(value)))
		stringBuf.WriteString(value)
	})

	_ = ptl.WriteUint64(buf, e.lsn)
	_ = ptl.WriteUint16(buf, 6)
	buf.WriteString("string")
	_ = ptl.WriteUint64(buf, uint64(stringBuf.Len()))
	buf.Write(stringBuf.Bytes())
	return buf.Bytes()
}

func (e *Engine) UnMarshal(reader io.Reader) error {
	lsn, err := ptl.ReadUint64(reader)

	if err != nil {
		return err
	}
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
	e.lsn = lsn
	return nil
}
