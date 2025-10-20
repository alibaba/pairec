package compress

import (
	"io"
	"sync"

	"github.com/klauspost/compress/zstd"
)

var zstdDecoder, _ = zstd.NewReader(nil, zstd.WithDecoderConcurrency(0))

func ZstdDecode(reader io.Reader) ([]byte, error) {
	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return zstdDecoder.DecodeAll(body, nil)
}

var zstdWriterPool = sync.Pool{
	New: func() interface{} {
		writer, _ := zstd.NewWriter(nil, zstd.WithEncoderConcurrency(1))
		return writer
	},
}

func ZstdEncode(body []byte) ([]byte, error) {
	writer := zstdWriterPool.Get().(*zstd.Encoder)
	defer zstdWriterPool.Put(writer)

	return writer.EncodeAll(body, make([]byte, 0, len(body))), nil
}
