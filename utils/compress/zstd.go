package compress

import (
	"io"

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
