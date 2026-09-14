package fs

import (
	"sync"
	"testing"

	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/domain"
)

// Run with go test -race ./persist/fs to detect unsynchronized project access.
func TestFSClientProjectAccessIsSynchronized(t *testing.T) {
	client := &FSClient{project: &domain.Project{}}
	projects := []*domain.Project{{}, {}}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			client.setProject(projects[i%len(projects)])
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			if client.GetProject() == nil {
				t.Error("GetProject returned nil")
				return
			}
		}
	}()
	wg.Wait()
}

func TestGetFeatureStoreClientUsesConfigKey(t *testing.T) {
	first := &FSClient{projectName: "b"}
	second := &FSClient{projectName: "shared"}
	third := &FSClient{projectName: "shared"}
	projectOnly := &FSClient{projectName: "project-only"}
	fsInstancesMu.Lock()
	original := fsInstances
	fsInstances = map[string]*FSClient{"a": first, "b": second, "c": third, "d": projectOnly}
	fsInstancesMu.Unlock()
	t.Cleanup(func() {
		fsInstancesMu.Lock()
		fsInstances = original
		fsInstancesMu.Unlock()
	})

	for _, test := range []struct {
		name string
		key  string
		want *FSClient
	}{
		{name: "key only", key: "a", want: first},
		{name: "key wins over another project", key: "b", want: second},
		{name: "same project retains separate keys", key: "c", want: third},
		{name: "project only is not a key", key: "project-only"},
		{name: "shared project is not a key", key: "shared"},
		{name: "missing key and project", key: "missing"},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, err := GetFeatureStoreClient(test.key)
			if got != test.want || (err != nil) != (test.want == nil) {
				t.Fatalf("GetFeatureStoreClient(%q) = %p, %v; want %p, error=%t", test.key, got, err, test.want, test.want == nil)
			}
		})
	}
}
