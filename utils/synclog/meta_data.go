package synclog

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

const (
	Meta_Head = "SYNCLOG"
)

type MetaData struct {
	//headSize uint8
	Version uint8
	Offset  uint32
	Head    string
}

func NewMetaData() *MetaData {
	return &MetaData{
		Head:    Meta_Head,
		Version: 1,
		Offset:  0,
	}
}

func (m *MetaData) MetaSize() int {
	return len(m.Head) + 1 + 4
}
func (m *MetaData) ToBytes() []byte {
	buf := bytes.NewBufferString(m.Head)

	binary.Write(buf, binary.LittleEndian, m.Version)
	binary.Write(buf, binary.LittleEndian, m.Offset)

	return buf.Bytes()
}

func (m *MetaData) FromBytes(reader io.Reader) error {

	head := make([]byte, len(Meta_Head))
	if _, err := reader.Read(head); err != nil {
		return err
	}

	if string(head) != Meta_Head {
		return errors.New("head string is not match")
	}
	m.Head = string(head)

	binary.Read(reader, binary.LittleEndian, &m.Version)
	binary.Read(reader, binary.LittleEndian, &m.Offset)

	//m.headSize = uint8(len(m.Head) + 1 + 4)
	return nil
}
