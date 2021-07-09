package net

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"strings"
)

var (
	sign = []byte{11, 24, 21, 126, 127}
	lenByteSize = 4
	LengthError = errors.New("Data length error. ")
)

func encode(args []string) []byte{
	lenBytes := make([]byte, lenByteSize)
	dataBytes, _ := json.Marshal(args)
	binary.BigEndian.PutUint32(lenBytes, uint32(len(dataBytes)))
	data := make([]byte, 0)
	data = append(data, sign...)
	data = append(data, lenBytes...)
	data = append(data, dataBytes...)
	return data
}

func decode(data []byte) ([]string, int, error){
	start := strings.Index(string(data), string(sign))
	if start == -1 {
		return nil, 0, LengthError
	}
	start += len(sign)
	if len(data) < start +lenByteSize {
		return nil, 0, LengthError
	}
	dataSize := int(binary.BigEndian.Uint32(data[start:start +lenByteSize]))
	start += lenByteSize
	if len(data) < start + dataSize {
		return nil, 0, LengthError
	}
	args := make([]string, 0)
	err := json.Unmarshal(data[start: start + dataSize], &args)
	if err != nil{
		return nil, 0, err
	}
	return args, start + dataSize, nil
}