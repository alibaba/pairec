package synclog

import (
	"bytes"
	"encoding/binary"
	"os"
	"sync"
	"time"

	"github.com/pkg/errors"
)

var (
	logFileSize = 2 << 30 // 2G
)

type ReadFunc func([]byte) error
type SyncLog struct {
	dirPath  string
	metaInfo *MetaInfo

	logInfo  *LogInfo
	readCh   chan *ReadRequest
	readFunc ReadFunc
	closeCh  chan struct{}
}

func NewSyncLog(dirPath string, f ReadFunc) *SyncLog {
	_, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		os.MkdirAll(dirPath, 0744)
	}

	return &SyncLog{
		dirPath:  dirPath,
		readCh:   make(chan *ReadRequest, 10000),
		closeCh:  make(chan struct{}, 1),
		readFunc: f,
	}
}

func (l *SyncLog) Init() error {
	info, err := OpenMetaFile(l.dirPath)
	if err != nil {
		return err
	}
	l.metaInfo = info

	logInfo, err := OpenLogFile(l.dirPath, info, l)
	if err != nil {
		return err
	}

	l.logInfo = logInfo

	if err := syncDir(l.dirPath); err != nil {
		return err
	}

	go l.loopRead()

	return nil
}

// Write write loger to sync log
func (l *SyncLog) Write(loger SyncLoger) error {

	playData := loger.Format()
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	binary.Write(buf, binary.LittleEndian, int32(len(playData)))
	buf.Write(playData)

	if err := l.logInfo.Write(buf.Bytes()); err != nil {
		return err
	}

	bufferPool.Put(buf)
	return nil
}
func (l *SyncLog) loopRead() {
	retryTimes := 10
	for {
		select {
		case req := <-l.readCh:
			for i := 0; i < retryTimes; i++ {
				if err := l.readFunc(req.PlayLoad); err == nil {
					break
				}
				time.Sleep(time.Second)
			}

			l.metaInfo.SaveOffset(req.Offset)
		case <-l.closeCh:
			return
		}
	}
}

func (l *SyncLog) Close() error {
	close(l.closeCh)
	if err := l.metaInfo.Close(); err != nil {
		return err
	}
	if err := l.logInfo.Close(); err != nil {
		return err
	}
	return nil
}

type SyncLoger interface {
	Format() []byte
	Parse([]byte) error
}
type SyncLogKVItem struct {
	Key   []byte
	Value []byte
}

var bufferPool sync.Pool

func init() {
	bufferPool = sync.Pool{
		New: func() any {
			return bytes.NewBuffer(nil)
		},
	}
}
func (kv *SyncLogKVItem) Format() []byte {
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)

	binary.Write(buf, binary.LittleEndian, int32(len(kv.Key)))
	buf.Write(kv.Key)
	binary.Write(buf, binary.LittleEndian, int32(len(kv.Value)))
	buf.Write(kv.Value)

	return buf.Bytes()
}

func (kv *SyncLogKVItem) Parse(data []byte) error {

	reader := bytes.NewReader(data)
	var keySize int32
	if err := binary.Read(reader, binary.LittleEndian, &keySize); err != nil {
		return err
	}

	kv.Key = make([]byte, keySize)

	if _, err := reader.Read(kv.Key); err != nil {
		return err
	}

	var valueSize int32
	if err := binary.Read(reader, binary.LittleEndian, &valueSize); err != nil {
		return err
	}

	kv.Value = make([]byte, valueSize)

	if _, err := reader.Read(kv.Value); err != nil {
		return err
	}

	return nil
}

type ReadRequest struct {
	Offset   uint32
	PlayLoad []byte
}

// When you create or delete a file, you have to ensure the directory entry for the file is synced
// in order to guarantee the file is visible (if the system crashes). (See the man page for fsync,
// or see https://github.com/coreos/etcd/issues/6368 for an example.)
func syncDir(dir string) error {
	f, err := os.Open(dir)
	if err != nil {
		return errors.Wrapf(err, "While opening directory: %s.", dir)
	}
	err = f.Sync()
	closeErr := f.Close()
	if err != nil {
		return errors.Wrapf(err, "While syncing directory: %s.", dir)
	}
	return errors.Wrapf(closeErr, "While closing directory: %s.", dir)
}
