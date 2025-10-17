package compress

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/pierrec/lz4/v4"
)

func TestLZ4Decode(t *testing.T) {
	str := "hello world"

	buf := bytes.NewBuffer(nil)
	writer := lz4.NewWriter(buf)
	writer.Write([]byte(str))
	writer.Close()
	fmt.Println(string(buf.Bytes()), buf.Len())

	uncompressed, err := LZ4Decode(buf)
	fmt.Println(string(uncompressed), err)
}
