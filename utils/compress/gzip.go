package compress

import (
	"bytes"
	"io"
	"sync"

	"github.com/klauspost/compress/gzip"
)

var gzipReaderPool = sync.Pool{
	New: func() interface{} {
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

func GzipEncode(body []byte) ([]byte, error) {
	var buf bytes.Buffer

	gzipWriter := gzip.NewWriter(&buf)

	gzipWriter.Reset(&buf)
	if _, err := gzipWriter.Write(body); err != nil {
		gzipWriter.Close()
		return nil, err
	}
	if err := gzipWriter.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
