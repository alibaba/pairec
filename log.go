package pairec

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/alibaba/pairec/v2/datasource/sls"
	"github.com/alibaba/pairec/v2/log"
	"github.com/alibaba/pairec/v2/recconf"
)

const (
	LOG_DEBUG = "DEBUG"
	LOG_INFO  = "INFO"
	LOG_ERROR = "ERROR"
	LOG_FATAL = "FATAL"

	LOG_OUTPUT_CONSOLE = "console"
)

var logDir string

func ClearDir(config recconf.LogConfig) {
	if config.LogLevel == LOG_DEBUG {
		flag.Set("v", "1")
	}

	if config.Output == LOG_OUTPUT_CONSOLE {
		flag.Set("logtostderr", "true")
	}

	if f := flag.Lookup("log_dir"); f != nil {
		logDir = f.Value.String()
	}

	if logDir != "" {
		go clearLoop(config)
	}

	if config.SLSName != "" {
		if slsclient, err := sls.GetSlsClient(config.SLSName); err == nil {
			log.RegisterSlsClient(slsclient)
		}
	}
}

type fileInfo struct {
	file   os.FileInfo
	delete bool
}

type fileInfoSlice []*fileInfo

func (us fileInfoSlice) Len() int {
	return len(us)
}
func (us fileInfoSlice) Less(i, j int) bool {
	return us[i].file.ModTime().Unix() < us[j].file.ModTime().Unix()
}
func (us fileInfoSlice) Swap(i, j int) {
	tmp := us[i]
	us[i] = us[j]
	us[j] = tmp
}
func clearLoop(config recconf.LogConfig) {
	fileInfoList := make([]*fileInfo, 0)
	for {
		fileInfoList = fileInfoList[:0]

		fileInfos, err := ioutil.ReadDir(logDir)
		if err != nil {
			fmt.Println(err)
			continue
		}

		pointTime := time.Now().Unix() - int64(config.RetensionDays*86400)
		totalSize := int64(0)
		for _, file := range fileInfos {
			if file.ModTime().Unix() < pointTime {
				path := filepath.Join(logDir, file.Name())
				err := os.Remove(path)
				if err != nil {
					fmt.Println(err)
				}
				continue
			}
			// omit symlink file
			if file.Mode() == os.ModeSymlink {
				continue
			}
			totalSize += file.Size()
			info := &fileInfo{
				delete: false,
				file:   file,
			}
			fileInfoList = append(fileInfoList, info)
		}

		sizeThreshold := int64(float64(config.DiskSize*1024*1024*1024) * 0.8)
		// sizeThreshold := int64(0)
		if totalSize > sizeThreshold {
			sort.Sort(fileInfoSlice(fileInfoList))
			for _, info := range fileInfoList {
				info.delete = true
				totalSize -= info.file.Size()

				if totalSize < sizeThreshold {
					break
				}
			}

			for _, info := range fileInfoList {
				if info.delete {
					path := filepath.Join(logDir, info.file.Name())
					err := os.Remove(path)
					if err != nil {
						fmt.Println(err)
					}
				}

			}
		}

		time.Sleep(10 * time.Second)
	}
}
