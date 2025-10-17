package compress

import (
	"io"

	"github.com/pierrec/lz4"
)

func LZ4Decode(reader io.Reader) ([]byte, error) {
	lz4Reader := lz4.NewReader(reader)
	return io.ReadAll(lz4Reader)

}
