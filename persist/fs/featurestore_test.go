package fs

import (
	"sync"
	"testing"

	"github.com/aliyun/aliyun-pai-featurestore-go-sdk/v2/domain"
)

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
