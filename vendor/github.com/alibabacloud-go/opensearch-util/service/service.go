package service

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"hash"
	"io"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/alibabacloud-go/tea/tea"
)

func GetDate() *string {
	gmt := time.FixedZone("GMT", 0)
	return tea.String(time.Now().In(gmt).Format("2006-01-02T15:04:05Z"))
}

func GetContentMD5(content *string) *string {
	sum := md5.Sum([]byte(tea.StringValue(content)))
	return tea.String(hex.EncodeToString(sum[:]))
}

func GetSignature(request *tea.Request, accessKeyId, accessKeySecret *string) *string {
	sign := "OPENSEARCH " + tea.StringValue(accessKeyId) + ":" + getSignature(request, tea.StringValue(accessKeySecret))
	return tea.String(sign)
}

func Append(in []*string, item *string) []*string {
	in = append(in, item)
	return in
}

func Keys(m map[string]interface{}) []*string {
	keys := make([]*string, 0)
	for key := range m {
		keys = append(keys, tea.String(key))
	}
	return keys
}

func Join(in []*string, spearator *string) *string {
	res := strings.Join(tea.StringSliceValue(in), tea.StringValue(spearator))
	return tea.String(res)
}

func MapToString(in map[string]*string, spearator *string) *string {
	res := ""
	for key, value := range in {
		res += key + tea.StringValue(spearator) + tea.StringValue(value) + ","
	}
	return tea.String(res)
}

func getSignature(request *tea.Request, accessKeySecret string) string {
	resource := tea.StringValue(request.Pathname)
	if !strings.Contains(resource, "?") && len(request.Query) > 0 {
		resource += "?"
	}
	queryKeys := make([]string, len(request.Query))
	for k,_ := range request.Query {
		queryKeys = append(queryKeys, k)
	}
	sort.Strings(queryKeys)
	for _, key := range queryKeys {
		value := request.Query[key]
		if value != nil {
			tmp := url.QueryEscape(tea.StringValue(value))
			tmp = strings.ReplaceAll(tmp, "'", "%27")
			tmp = strings.ReplaceAll(tmp, "+", "%20")
			if strings.HasSuffix(resource, "?") {
				resource = resource + key + "=" + tmp
			} else {
				resource = resource + "&" + key + "=" + tmp
			}
		}
	}
	return getSignedStr(request, resource, accessKeySecret)
}

// Sorter defines the key-value structure for storing the sorted data in signHeader.
type Sorter struct {
	Keys []string
	Vals []string
}

// newSorter is an additional function for function Sign.
func newSorter(m map[string]string) *Sorter {
	hs := &Sorter{
		Keys: make([]string, 0, len(m)),
		Vals: make([]string, 0, len(m)),
	}

	for k, v := range m {
		hs.Keys = append(hs.Keys, k)
		hs.Vals = append(hs.Vals, v)
	}
	return hs
}

// Sort is an additional function for function SignHeader.
func (hs *Sorter) Sort() {
	sort.Sort(hs)
}

// Len is an additional function for function SignHeader.
func (hs *Sorter) Len() int {
	return len(hs.Vals)
}

// Less is an additional function for function SignHeader.
func (hs *Sorter) Less(i, j int) bool {
	return bytes.Compare([]byte(hs.Keys[i]), []byte(hs.Keys[j])) < 0
}

// Swap is an additional function for function SignHeader.
func (hs *Sorter) Swap(i, j int) {
	hs.Vals[i], hs.Vals[j] = hs.Vals[j], hs.Vals[i]
	hs.Keys[i], hs.Keys[j] = hs.Keys[j], hs.Keys[i]
}

func getSignedStr(req *tea.Request, canonicalizedResource, accessKeySecret string) string {
	// Find out the "x-oss-"'s address in header of the request
	temp := make(map[string]string)

	for k, v := range req.Headers {
		if strings.HasPrefix(strings.ToLower(k), "x-opensearch-") {
			temp[strings.ToLower(k)] = tea.StringValue(v)
		}
	}
	hs := newSorter(temp)

	// Sort the temp by the ascending order
	hs.Sort()

	// Get the canonicalizedOSSHeaders
	canonicalizedOSSHeaders := ""
	for i := range hs.Keys {
		canonicalizedOSSHeaders += hs.Keys[i] + ":" + hs.Vals[i] + "\n"
	}

	// Give other parameters values
	// when sign URL, date is expires
	date := tea.StringValue(req.Headers["Date"])
	contentType := tea.StringValue(req.Headers["Content-Type"])
	contentMd5 := tea.StringValue(req.Headers["Content-MD5"])

	signStr := tea.StringValue(req.Method) + "\n" + contentMd5 + "\n" + contentType + "\n" + date + "\n" + canonicalizedOSSHeaders + canonicalizedResource
	h := hmac.New(func() hash.Hash { return sha1.New() }, []byte(accessKeySecret))
	io.WriteString(h, signStr)
	signedStr := base64.StdEncoding.EncodeToString(h.Sum(nil))

	return signedStr
}
