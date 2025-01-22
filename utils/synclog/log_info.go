package synclog

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/alibaba/pairec/v2/log"
	unix "golang.org/x/sys/unix"
)

type LogInfo struct {
	logFile     *os.File
	logFileData []byte
	metaInfo    *MetaInfo
	syncLog     *SyncLog
	mu          sync.Mutex
	size        int
}

func OpenLogFile(dirPath string, metaInfo *MetaInfo, syncLog *SyncLog) (*LogInfo, error) {
	logFilePath := fmt.Sprintf("%s/log_sync.data", dirPath)
	logInfo := &LogInfo{metaInfo: metaInfo, size: 0, syncLog: syncLog}
	logFileInfo, err := os.Stat(logFilePath)
	if os.IsNotExist(err) {
		logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_RDWR|os.O_EXCL, 0644)
		if err != nil {
			return nil, fmt.Errorf("create log file %s: %v fail", logFilePath, err)
		}

		logData, err := unix.Mmap(int(logFile.Fd()), 0, logFileSize*2, unix.PROT_WRITE, unix.MAP_SHARED)
		if err != nil {
			return nil, err
		}
		logInfo.logFileData = logData
		logInfo.logFile = logFile
		logInfo.logFile.Seek(0, io.SeekStart)
	} else if err != nil {
		return nil, fmt.Errorf("can't stat log file %s: %v", logFilePath, err)
	} else {
		logFile, err := os.OpenFile(logFilePath, os.O_RDWR, 0644)
		if err != nil {
			return nil, fmt.Errorf("open log file %s: %v fail", logFilePath, err)
		}
		logData, err := unix.Mmap(int(logFile.Fd()), 0, logFileSize*2, unix.PROT_WRITE, unix.MAP_SHARED)
		if err != nil {
			return nil, err
		}

		logInfo.logFileData = logData
		logInfo.logFile = logFile
		size := logFileInfo.Size()

		log.Info(fmt.Sprintf("open log file, dirPath:%s, meta offset:%d, log file size:%d", dirPath, metaInfo.meta.Offset, size))
		if size < int64(metaInfo.meta.Offset) {
			log.Error(fmt.Sprintf("open log file meta error, dirPath:%s, meta offset:%d, log file size:%d", dirPath, metaInfo.meta.Offset, size))
			metaInfo.meta.Offset = 0
		}

		buf := make([]byte, size-int64(metaInfo.meta.Offset))
		copy(buf, logInfo.logFileData[metaInfo.meta.Offset:size])
		//logInfo.logFileData = logInfo.logFileData[:0]
		copy(logInfo.logFileData, buf)

		logInfo.size = len(buf)
		logInfo.logFile.Seek(int64(logInfo.size), io.SeekStart)
		unix.Ftruncate(int(logFile.Fd()), int64(logInfo.size))
		metaInfo.SaveOffset(0)
		if len(buf) > 0 {
			fmt.Printf("reply data size:%d\n", len(buf))
			if err := logInfo.replyRead(buf); err != nil {
				return nil, err
			}
		}
	}
	return logInfo, nil
}
func (l *LogInfo) replyRead(data []byte) error {
	reader := bytes.NewReader(data)
	offset := int32(0)
	for {
		size := int32(-1)
		err := binary.Read(reader, binary.LittleEndian, &size)
		if err == io.EOF {
			break
		}

		if err != nil {
			return err
		}

		buf := make([]byte, size)
		n, err := reader.Read(buf)
		if err != nil {
			return err
		}
		if n != int(size) {
			return fmt.Errorf("read data fail")
		}

		offset += 4 + size
		req := &ReadRequest{
			Offset:   uint32(offset),
			PlayLoad: buf,
		}

		l.syncLog.readCh <- req

	}

	return nil
}

func (l *LogInfo) Write(data []byte) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	n, err := l.logFile.Write(data)
	if err != nil {
		return err
	}
	l.size += n

	req := &ReadRequest{}
	req.Offset = uint32(l.size)
	req.PlayLoad = make([]byte, len(data)-4)
	copy(req.PlayLoad, data[4:])

	select {
	case l.syncLog.readCh <- req:
	default:
		return errors.New("read chan is full")
	}

	if l.size >= logFileSize {
		for {
			if l.size == int(l.metaInfo.meta.Offset) {
				break
			}
			time.Sleep(time.Millisecond * 100)
		}
		fmt.Println("set file to start write")
		unix.Ftruncate(int(l.logFile.Fd()), 0)
		l.logFile.Seek(0, io.SeekStart)
		l.metaInfo.SaveOffset(0)
		l.size = 0
	}
	return nil
}
func (l *LogInfo) Close() error {
	if err := unix.Munmap(l.logFileData); err != nil {
		return err
	}
	return l.logFile.Close()
}
