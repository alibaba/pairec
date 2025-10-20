package compress

import (
	"bytes"
	"testing"

	"fortio.org/assert"
)

func TestLZ4Decode(t *testing.T) {
	str := "hello world"

	compressData, err := LZ4Encode([]byte(str))
	assert.Equal(t, err, nil)

	uncompressed, err := LZ4Decode(bytes.NewReader(compressData))
	assert.Equal(t, err, nil)
	assert.Equal(t, string(uncompressed), str)
}
