package ptl

import (
	"bytes"
	"encoding/binary"
	"io"
)

func WriteUint16(writer io.Writer, i uint16) error {
	data := make([]byte, 2)
	binary.BigEndian.PutUint16(data, i)
	_, err := writer.Write(data)
	return err
}

func WriteUint32(writer io.Writer, i uint32) error {
	data := make([]byte, 4)
	binary.BigEndian.PutUint32(data, i)
	_, err := writer.Write(data)
	return err
}

func WriteUint64(writer io.Writer, i uint64) error {
	data := make([]byte, 8)
	binary.BigEndian.PutUint64(data, i)
	_, err := writer.Write(data)
	return err
}

func ReadUint16(reader io.Reader) (uint16, error) {
	data := make([]byte, 2)
	_, err := io.ReadFull(reader, data)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(data), nil
}

func ReadUint32(reader io.Reader) (uint32, error) {
	data := make([]byte, 4)
	_, err := io.ReadFull(reader, data)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(data), nil
}

func ReadUint64(reader io.Reader) (uint64, error) {
	data := make([]byte, 8)
	_, err := io.ReadFull(reader, data)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(data), nil
}

func ReadBytes(reader io.Reader, size int) ([]byte, error) {
	data := make([]byte, size)
	_, err := io.ReadFull(reader, data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func UnMarshal(reader io.Reader) ([]string, error) {
	count, err := ReadUint16(reader)
	if err != nil {
		return nil, err
	}
	args := make([]string, count)
	for i := 0; i < int(count); i++ {
		size, err := ReadUint32(reader)
		if err != nil {
			return nil, err
		}
		data, err := ReadBytes(reader, int(size))
		if err != nil {
			return nil, err
		}
		args[i] = string(data)
	}
	return args, nil
}

func Marshal(args []string) ([]byte, error) {
	buf := &bytes.Buffer{}
	err := WriteUint16(buf, uint16(len(args)))
	if err != nil {
		return nil, err
	}
	for _, data := range args {
		err = WriteUint32(buf, uint32(len(data)))
		if err != nil {
			return nil, err
		}
		_, err := buf.Write([]byte(data))
		if err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

func MarshalWrappedLSN(id uint64, args []string) ([]byte, error) {
	buf := &bytes.Buffer{}
	err := WriteUint64(buf, id)
	if err != nil {
		return nil, err
	}
	data, err := Marshal(args)
	if err != nil {
		return nil, err
	}
	buf.Write(data)
	return buf.Bytes(), nil
}

func UnMarshalWrappedLSN(reader io.Reader) (uint64, []string, error) {
	lsn, err := ReadUint64(reader)
	if err != nil {
		return 0, nil, err
	}
	args, err := UnMarshal(reader)
	if err != nil {
		return 0, nil, err
	}
	return lsn, args, nil
}
