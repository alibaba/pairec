package types

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"github.com/alibaba/pairec/pkg/eas/types/queue_service_protos"
	"google.golang.org/protobuf/proto"
	"io"
)

type (
	// Protobuf data frame codec implement.
	pbDataFrameCodec struct{}
	// Protobuf attributes codec implement.
	pbAttributesCodec struct{}
	// JSON data frame codec implement.
	jsonDataFrameCodec struct{}
	// JSON attributes codec implement.
	jsonAttributesCodec struct{}
)

const (
	ContentTypeProtobuf   = "application/vnd.google.protobuf"
	ContentTypeFlatbuffer = "application/x-flatbuffers"
	ContentTypeJSON       = "application/json"
)

func DataFrameCodecFor(contentType string) DataFrameCodec {
	switch contentType {
	case ContentTypeProtobuf:
		return &pbDataFrameCodec{}
	case ContentTypeJSON:
		return &jsonDataFrameCodec{}
	default:
		return nil
	}
}

func AttributesCodecFor(contentType string) AttributesCodec {
	switch contentType {
	case ContentTypeProtobuf:
		return &pbAttributesCodec{}
	case ContentTypeJSON:
		return &jsonAttributesCodec{}
	default:
		return nil
	}
}

func (p *pbDataFrameCodec) EncodeList(list []DataFrame, w io.Writer) error {
	dfProto := queue_service_protos.DataFrameListProto{}
	for _, df := range list {
		dfProto.Index = append(dfProto.Index, &queue_service_protos.DataFrameProto{
			Index: df.Index.Uint64(),
			Data:  df.Data,
			Tags:  df.Tags,
		})
	}
	data, err := proto.Marshal(&dfProto)
	if err != nil {
		return err
	}
	_, _ = w.Write(data)
	return nil
}

func (p *pbDataFrameCodec) DecodeList(bytes []byte) ([]DataFrame, error) {
	dfProto := queue_service_protos.DataFrameListProto{}
	if err := proto.Unmarshal(bytes, &dfProto); err != nil {
		return nil, err
	}
	ret := make([]DataFrame, 0, len(dfProto.Index))
	for _, idx := range dfProto.Index {
		ret = append(ret, DataFrame{
			Data:  idx.Data,
			Index: FromUint64(idx.Index),
			Tags:  idx.Tags,
		})
	}
	return ret, nil
}

func (p *pbDataFrameCodec) MediaType() string {
	return ContentTypeProtobuf
}

func (p *pbDataFrameCodec) Encode(frame DataFrame, w io.Writer) error {
	dfProto := queue_service_protos.DataFrameProto{Data: frame.Data, Tags: frame.Tags, Index: frame.Index.Uint64(), Message: frame.Message}
	data, err := proto.Marshal(&dfProto)
	if err != nil {
		return err
	}
	_, _ = w.Write(data)
	return nil
}

func (p *pbDataFrameCodec) Decode(bytes []byte, frame *DataFrame) error {
	dfProto := queue_service_protos.DataFrameProto{}
	if err := proto.Unmarshal(bytes, &dfProto); err != nil {
		return err
	}
	frame.Tags = dfProto.Tags
	frame.Index = FromUint64(dfProto.Index)
	frame.Data = dfProto.Data
	return nil
}

func (p *pbAttributesCodec) MediaType() string {
	return ContentTypeProtobuf
}

func (p *pbAttributesCodec) Encode(a Attributes, w io.Writer) error {
	aProto := queue_service_protos.AttributesProto{Attributes: a}
	data, err := proto.Marshal(&aProto)
	if err != nil {
		return err
	}
	_, _ = w.Write(data)
	return nil
}

func (p *pbAttributesCodec) Decode(bytes []byte, a *Attributes) error {
	aProto := queue_service_protos.AttributesProto{}
	if err := proto.Unmarshal(bytes, &aProto); err != nil {
		return err
	}
	attr := Attributes(aProto.Attributes)
	*a = attr
	return nil
}

func (j *jsonAttributesCodec) MediaType() string {
	return ContentTypeJSON
}

func (j *jsonAttributesCodec) Encode(attr Attributes, w io.Writer) error {
	return json.NewEncoder(w).Encode(attr)
}

func (j *jsonAttributesCodec) Decode(data []byte, attributes *Attributes) error {
	return json.NewDecoder(bytes.NewReader(data)).Decode(attributes)
}

func (j *jsonDataFrameCodec) MediaType() string {
	return ContentTypeJSON
}

type dataFrameJSON struct {
	Index   uint64            `json:"index"`
	Message string            `json:"message,omitempty"`
	Tags    map[string]string `json:"tags"`
	Data    string            `json:"data"`
}

func (j *jsonDataFrameCodec) Encode(frame DataFrame, w io.Writer) error {
	jd := dataFrameJSON{Data: base64.StdEncoding.EncodeToString(frame.Data), Tags: frame.Tags, Message: frame.Message, Index: frame.Index.Uint64()}
	return json.NewEncoder(w).Encode(jd)
}

func (j *jsonDataFrameCodec) EncodeList(list []DataFrame, w io.Writer) error {
	jds := make([]dataFrameJSON, 0, len(list))
	for _, frame := range list {
		jd := dataFrameJSON{Data: base64.StdEncoding.EncodeToString(frame.Data), Tags: frame.Tags, Message: frame.Message, Index: frame.Index.Uint64()}
		jds = append(jds, jd)
	}
	return json.NewEncoder(w).Encode(jds)
}

func (j *jsonDataFrameCodec) Decode(i []byte, frame *DataFrame) error {
	jd := dataFrameJSON{}
	err := json.NewDecoder(bytes.NewReader(i)).Decode(&jd)
	if err != nil {
		return err
	}
	data, err := base64.StdEncoding.DecodeString(jd.Data)
	if err != nil {
		return err
	}
	*frame = DataFrame{
		Index:   FromUint64(jd.Index),
		Message: jd.Message,
		Tags:    jd.Tags,
		Data:    data,
	}
	return nil
}

func (j *jsonDataFrameCodec) DecodeList(i []byte) ([]DataFrame, error) {
	var jds []dataFrameJSON
	err := json.NewDecoder(bytes.NewReader(i)).Decode(&jds)
	if err != nil {
		return nil, err
	}
	frames := make([]DataFrame, 0, len(jds))
	for _, jd := range jds {
		data, err := base64.StdEncoding.DecodeString(jd.Data)
		if err != nil {
			return nil, err
		}
		frames = append(frames, DataFrame{
			Index:   FromUint64(jd.Index),
			Message: jd.Message,
			Tags:    jd.Tags,
			Data:    data,
		})
	}
	return frames, nil
}

type lengthDelimitedFrameWriter struct {
	w io.Writer
	h [4]byte
}

func NewLengthDelimitedFrameWriter(w io.Writer) io.Writer {
	return &lengthDelimitedFrameWriter{w: w}
}

// Write writes a single frame to the nested writer, prepending it with the length in
// in bytes of data (as a 4 byte, bigendian uint32).
func (w *lengthDelimitedFrameWriter) Write(data []byte) (int, error) {
	binary.BigEndian.PutUint32(w.h[:], uint32(len(data)))
	n, err := w.w.Write(w.h[:])
	if err != nil {
		return 0, err
	}
	if n != len(w.h) {
		return 0, io.ErrShortWrite
	}
	return w.w.Write(data)
}

type lengthDelimitedFrameReader struct {
	r         io.ReadCloser
	remaining int
}

// NewLengthDelimitedFrameReader returns an io.Reader that will decode length-prefixed
// frames off of a stream.
//
// The protocol is:
//
//   stream: message ...
//   message: prefix body
//   prefix: 4 byte uint32 in BigEndian order, denotes length of body
//   body: bytes (0..prefix)
//
// If the buffer passed to Read is not long enough to contain an entire frame, io.ErrShortRead
// will be returned along with the number of bytes read.
func NewLengthDelimitedFrameReader(r io.ReadCloser) io.ReadCloser {
	return &lengthDelimitedFrameReader{r: r}
}

// Read attempts to read an entire frame into data. If that is not possible, io.ErrShortBuffer
// is returned and subsequent calls will attempt to read the last frame. A frame is complete when
// err is nil.
func (r *lengthDelimitedFrameReader) Read(data []byte) (int, error) {
	if r.remaining <= 0 {
		header := [4]byte{}
		n, err := io.ReadAtLeast(r.r, header[:4], 4)
		if err != nil {
			return 0, err
		}
		if n != 4 {
			return 0, io.ErrUnexpectedEOF
		}
		frameLength := int(binary.BigEndian.Uint32(header[:]))
		r.remaining = frameLength
	}

	expect := r.remaining
	max := expect
	if max > len(data) {
		max = len(data)
	}
	n, err := io.ReadAtLeast(r.r, data[:max], max)
	r.remaining -= n
	if err == io.ErrShortBuffer || r.remaining > 0 {
		return n, io.ErrShortBuffer
	}
	if err != nil {
		return n, err
	}
	if n != expect {
		return n, io.ErrUnexpectedEOF
	}

	return n, nil
}

func (r *lengthDelimitedFrameReader) Close() error {
	return r.r.Close()
}
