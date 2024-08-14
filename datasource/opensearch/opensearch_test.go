package opensearch

import (
	"fmt"
	"os"
	"testing"
)

func TestOpenSearchClient(t *testing.T) {
	endpoint := "opensearch-cn-beijing.aliyuncs.com"
	accessId := os.Getenv("ALIBABA_CLOUD_ACCESS_KEY_ID")
	accessKey := os.Getenv("ALIBABA_CLOUD_ACCESS_KEY_SECRET")
	client := NewOpenSearchClient(endpoint, accessId, accessKey)

	fmt.Println(client.OpenSearchClient)
	if client.OpenSearchClient == nil {
		t.Error("client is nil")
	}
}
