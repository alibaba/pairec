package types

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"
)

type Tags map[string]string

func (t Tags) Validate() error {
	if len(t) == 0 {
		return nil
	}
	for key := range t {
		if strings.HasPrefix(key, "_") {
			return fmt.Errorf("tag key %q contains prefix _ which is reserved", key)
		}
	}
	return nil
}

func (t Tags) Empty() bool {
	if t == nil {
		return true
	}
	return len(t) == 0
}

func (t Tags) Equals(t1 Tags) bool {
	if len(t) != len(t1) {
		return false
	}
	for key, val := range t {
		if t[key] != val {
			return false
		}
	}
	return true
}

func (t Tags) Contains(t1 Tags) bool {
	if len(t) < len(t1) {
		return false
	}
	count := 0
	for key, val := range t1 {
		if t[key] == val {
			count++
		} else {
			return false
		}
	}
	return count == len(t1)
}

func (t Tags) Has(key string) bool {
	_, ok := t[key]
	return ok
}

func (t Tags) Set(key string, value string) {
	t[key] = value
}

func (t Tags) Get(key string) string {
	return t[key]
}

func (t Tags) Diff(t1 Tags) (add Tags, del Tags, update Tags) {
	handled := make(map[string]bool, len(t)+len(t1))
	add, del, update = Tags{}, Tags{}, Tags{}
	for key, val := range t {
		handled[key] = true
		if val1, exist := t1[key]; exist {
			if val1 != val {
				update[key] = val1
			}
		} else {
			del[key] = val
		}
	}
	for key, val := range t1 {
		if !handled[key] {
			add[key] = val
		}
	}
	return
}

func (t Tags) ToJSON() string {
	data, _ := json.Marshal(t)
	return string(data)
}

type DataFrame struct {
	Data []byte

	Index Index

	Tags Tags

	Message string
}

func (f *DataFrame) Empty() bool {
	return len(f.Data) == 0 && len(f.Tags) == 0 && len(f.Message) == 0
}

type DataFrameEncoder interface {
	// Encode encodes DataFrame into bytes.
	Encode(frame DataFrame, w io.Writer) error

	// EncodeList attempts to encode batch of DataFrame into bytes.
	EncodeList(list []DataFrame, w io.Writer) error
}

type DataFrameDecoder interface {
	// Decode decodes DataFrame from bytes.
	Decode([]byte, *DataFrame) error

	// DecodeList attempts to decode DataFrameList from bytes.
	DecodeList([]byte) ([]DataFrame, error)
}

type AttributesEncoder interface {
	Encode(Attributes Attributes, w io.Writer) error
}

type AttributesDecoder interface {
	Decode([]byte, *Attributes) error
}

// DataFrameCodec helps to encode or decode a DataFrame from or to bytes.
type DataFrameCodec interface {
	MediaType() string

	DataFrameEncoder

	DataFrameDecoder
}

// AttributesCodec helps to encode or decode Attributes from or to bytes.
type AttributesCodec interface {
	MediaType() string

	AttributesEncoder

	AttributesDecoder
}

func LargestIndex(dfs []DataFrame) Index {
	var max Index
	for i := range dfs {
		if dfs[i].Index > max {
			max = dfs[i].Index
		}
	}
	return max
}

type Range struct {
	LeftInclude  bool
	RightInclude bool
	PositiveInf  bool

	Begin uint64
	End   uint64
}

func ParseRange(input string) (Range, error) {
	const (
		stateBegin = iota
		stateEnd
		stateLVal
		stateRVal
		stateDelim
		stateInf
	)
	var (
		err             error
		state           = stateBegin
		result          = Range{}
		lval, rval, inf []rune
	)
	// remove all spaces.
	for i, c := range strings.ReplaceAll(input, " ", "") {
		switch state {
		case stateBegin:
			switch c {
			case '(':
				result.LeftInclude = false
				state = stateLVal
			case '[':
				result.LeftInclude = true
				state = stateLVal
			default:
				return result, fmt.Errorf("invalid character '%c' at index %d", c, i)
			}
		case stateLVal:
			switch c {
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				lval = append(lval, c)
			case ',':
				if len(lval) == 0 {
					return result, fmt.Errorf("malformed range left value")
				}
				state = stateDelim
			default:
				return result, fmt.Errorf("invalid character '%c' at index %d", c, i)
			}
		case stateDelim:
			result.Begin, err = strconv.ParseUint(string(lval), 10, 64)
			if err != nil {
				return result, err
			}
			switch c {
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				rval = append(rval, c)
				state = stateRVal
			case '+', 'i', 'I':
				state = stateInf
			default:
				return result, fmt.Errorf("invalid character '%c' at index %d", c, i)
			}
		case stateRVal:
			switch c {
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				rval = append(rval, c)
			case ')', ']':
				state = stateEnd
				result.End, err = strconv.ParseUint(string(rval), 10, 64)
				if err != nil {
					return result, err
				}
				switch c {
				case ']':
					result.RightInclude = true
				case ')':
					result.RightInclude = false
				}
			default:
				return result, fmt.Errorf("invalid character '%c' at index %d", c, i)
			}
		case stateInf:
			switch c {
			case 'i', 'n', 'f', '+':
				inf = append(inf, c)
			case ')':
				state = stateEnd
				if string(inf) == "+inf" || string(inf) == "inf" || string(inf) == "+Inf" || string(inf) == "Inf" {
					result.PositiveInf = true
				} else {
					return result, fmt.Errorf("invalid symbol '%s'", string(inf))
				}
			default:
				return result, fmt.Errorf("invalid character '%c' at index %d", c, i)
			}
		case stateEnd:
			return result, fmt.Errorf("invalid character '%c' at index %d", c, i)
		}
	}
	return result, nil
}

func (r Range) String() string {
	var sb strings.Builder
	li := "("
	ri := ")"
	if r.LeftInclude {
		li = "["
	}
	if r.RightInclude && !r.PositiveInf {
		ri = "]"
	}

	sb.WriteString(li)
	sb.WriteString(fmt.Sprintf("%d", r.Begin))
	sb.WriteRune(',')
	if r.PositiveInf {
		sb.WriteString("+inf")
	} else {
		sb.WriteString(fmt.Sprintf("%d", r.End))
	}
	sb.WriteString(ri)
	return sb.String()
}

func (r Range) Empty() bool {
	return r.Begin == 0 && r.End == 0
}

// Watcher is the entity following the stream.
type Watcher interface {
	// Watcher is a kind of DataFrameReader.
	DataFrameReader
	// Close stops Watcher and closes the FrameChan.
	Close()
}

type Attributes map[string]string

// const attributes keys the queue service implement must provide.
const (
	Backend                 = "meta.backend"
	MaxPayloadBytes         = "meta.maxPayloadBytes"
	UserIdentifyHeader      = "meta.header.userIdentifyHeader"
	GroupIdentifyHeader     = "meta.header.groupIdentifyHeader"
	StreamLength            = "stream.length"
	StreamApproximateLength = "stream.approxMaxLength"
)

var MaxIndex = FromUint64(uint64(math.MaxUint64))

// Interface of QueueService. Core abstraction for streaming framework.
type Interface interface {
	// End normally emits 'EOS' symbol to end up the queue asynchronously,
	// but if force set to true, stream ends up directly.
	// Undelivered data will be truncated.
	End(ctx context.Context, force bool) error
	// Truncate truncates data before the specific index.
	Truncate(ctx context.Context, index uint64) error
	// Put appends new data into stream.
	Put(ctx context.Context, data []byte, tags Tags) (index uint64, err error)
	// Get returns data frames from the index of stream in queue.
	// Param length specifies the expected message count.
	// And if timeout is set, this call will block until length got satisfied or
	// timeout timer fires.
	Get(ctx context.Context, index uint64, length int, timeout time.Duration, tags Tags) (dfs []DataFrame, err error)
	// Watch subscribe to queue service, when new data frame is appended through Put method,
	// watcher will emit it through its result channel.
	// Param index specifies the beginning message index of the watch.
	// Param window specifies the largest size the Watcher could transfer at one time.
	Watch(ctx context.Context, index uint64, indexOnly bool, noAck bool, window uint64) (Watcher, error)
	// Commit commits indices to make the corresponding messages marked as consumed.
	Commit(ctx context.Context, del bool, indexes ...uint64) error
	// Del deletes indices to make the corresponding messages deleted from stream.
	Del(ctx context.Context, indexes ...uint64) error
	// Attributes reflects self dynamic attributes by K/V pairs.
	Attributes() Attributes
}

type DataFrameReader interface {
	// FrameChan return a DataFrame channel.
	FrameChan() <-chan DataFrame
}

// User authenticated information.
type User interface {
	// Uid represents the user id.
	Uid() string

	// Gid represents the group id of user.
	Gid() string

	// Token represents the access token of the queue service.
	Token() string
}

type UserAware interface {
	// User returns the user info.
	User() User
}

type UserWithToken interface {
	// Token to access the backend service.
	Token() string
}

const (
	userKey = "__user__"
)

// WithUser saves User into context.
func WithUser(ctx context.Context, user User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

// UserFromContext loads User from context.
func UserFromContext(ctx context.Context) (User, bool) {
	i := ctx.Value(userKey)
	if u, ok := i.(User); ok {
		return u, ok
	}
	return nil, false
}

type WorkerStatus string

const (
	WorkerRunning WorkerStatus = "Running"
	WorkerStopped WorkerStatus = "Stopped"
	WorkerError   WorkerStatus = "Error"
	WorkerUnknown WorkerStatus = "Unknown"
)

type StreamStatus string

const (
	StreamOk     StreamStatus = "OK"
	StreamCancel StreamStatus = "Cancel"
	StreamEnd    StreamStatus = "End"
)

const (
	OffsetEOS Offset = "eos"
)

type Offset string

func (o Offset) IsInf() bool {
	low := strings.ToLower(string(o))
	return low == "inf" || low == "+inf"
}

func (o Offset) Uint64() (uint64, bool) {
	u, err := strconv.ParseUint(string(o), 10, 64)
	if err != nil {
		return 0, false
	}
	return u, true
}

func Compare(o1, o2 Offset) (int, error) {
	uint1, ok1 := o1.Uint64()
	uint2, ok2 := o2.Uint64()
	switch {
	case ok1 && ok2:
		if uint1 > uint2 {
			return 1, nil
		} else if uint1 < uint2 {
			return -1, nil
		} else {
			return 0, nil
		}
	case ok1 && !ok2:
		if o2 == OffsetEOS || o2.IsInf() {
			return -1, nil
		} else {
			return -2, fmt.Errorf("unexpected offset: %v", o2)
		}
	case !ok1 && ok2:
		if o1 == OffsetEOS || o1.IsInf() {
			return 1, nil
		} else {
			return -2, fmt.Errorf("unexpected offset: %v", o1)
		}
	default:
		if o1 == OffsetEOS && o2 == OffsetEOS {
			return 0, nil
		}
		if o1.IsInf() && o2.IsInf() {
			return 0, nil
		}

		return -2, fmt.Errorf("unexpected compare: %s vs %s", o1, o2)
	}
}
