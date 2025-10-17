package compress

import (
	"io"
	"sync"

	"github.com/klauspost/compress/gzip"
)

var gzipReaderPool = sync.Pool{
	// New 函数在池中没有可用对象时被调用，用于创建一个新的对象
	New: func() interface{} {
		// 注意：我们直接返回一个 *gzip.Reader 实例。
		// NewReader() 在这里不适用，因为它需要一个 io.Reader 参数。
		// 我们将在使用时通过 Reset() 方法来提供 reader。
		return new(gzip.Reader)
	},
}

func GzipDecode(reader io.Reader) ([]byte, error) {
	gzipReader := gzipReaderPool.Get().(*gzip.Reader)
	err := gzipReader.Reset(reader)
	if err != nil {
		gzipReaderPool.Put(gzipReader)
		return nil, err
	}

	defer gzipReaderPool.Put(gzipReader)

	decompressed, err := io.ReadAll(gzipReader)
	if err != nil {
		return nil, err
	}

	return decompressed, nil
}
