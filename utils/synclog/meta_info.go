package synclog

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"

	unix "golang.org/x/sys/unix"
)

type MetaInfo struct {
	metaFile *os.File
	metaData []byte // use mmap to open meta file

	meta *MetaData
}

func OpenMetaFile(dirPath string) (*MetaInfo, error) {
	metaFilePath := fmt.Sprintf("%s/meta", dirPath)
	metaInfo := &MetaInfo{}

	metaFileInfo, err := os.Stat(metaFilePath)
	if os.IsNotExist(err) {
		metaFile, err := os.OpenFile(metaFilePath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
		if err != nil {
			return nil, fmt.Errorf("create meta file %s: %v fail", metaFilePath, err)
		}
		err = unix.Ftruncate(int(metaFile.Fd()), 1024)
		if err != nil {
			return nil, err
		}

		metaData, err := unix.Mmap(int(metaFile.Fd()), 0, 1024, unix.PROT_WRITE, unix.MAP_SHARED)
		if err != nil {
			return nil, err
		}
		metaInfo.metaData = metaData
		meta := NewMetaData()
		metaInfo.metaFile = metaFile
		metaInfo.meta = meta
		metaBytes := meta.ToBytes()

		copy(metaInfo.metaData, metaBytes)

	} else if err != nil {
		return nil, fmt.Errorf("can't stat meta file %s: %v", metaFilePath, err)
	} else {
		metaFile, err := os.OpenFile(metaFilePath, os.O_RDWR, 0644)
		if err != nil {
			return nil, fmt.Errorf("open meta file %s: %v fail", metaFilePath, err)
		}

		metaData, err := unix.Mmap(int(metaFile.Fd()), 0, int(metaFileInfo.Size()), unix.PROT_WRITE, unix.MAP_SHARED)
		if err != nil {
			return nil, err
		}
		metaInfo.metaData = metaData
		metaInfo.metaFile = metaFile
		meta := NewMetaData()
		metaSize := meta.MetaSize()
		buf := make([]byte, metaSize)
		copy(buf, metaInfo.metaData)
		if err := meta.FromBytes(bytes.NewReader(buf)); err != nil {
			return nil, err
		}
		metaInfo.meta = meta
	}
	return metaInfo, nil
}

func (m *MetaInfo) SaveOffset(offset uint32) {
	m.meta.Offset = offset
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)
	binary.Write(buf, binary.LittleEndian, offset)
	copy(m.metaData[m.meta.MetaSize()-4:], buf.Bytes())
}

func (m *MetaInfo) Close() error {
	if err := unix.Munmap(m.metaData); err != nil {
		return err
	}
	return m.metaFile.Close()
}
