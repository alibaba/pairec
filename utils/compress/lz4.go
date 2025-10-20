package compress

import (
	"bytes"
	"io"

	"github.com/pierrec/lz4"
)

func LZ4Decode(reader io.Reader) ([]byte, error) {
	lz4Reader := lz4.NewReader(reader)
	return io.ReadAll(lz4Reader)

}

func LZ4Encode(body []byte) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	writer := lz4.NewWriter(buf)
	if _, err := writer.Write(body); err != nil {
		_ = writer.Close()
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
