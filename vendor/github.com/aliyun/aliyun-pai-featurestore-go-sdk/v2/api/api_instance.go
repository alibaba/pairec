package api

import (
	"context"
	"errors"
	"fmt"

	paifeaturestore "github.com/alibabacloud-go/paifeaturestore-20230621/client"
)

// Linger please
var (
	_ context.Context
)

type InstanceApiService service

func (a *InstanceApiService) GetInstance() error {
	request := paifeaturestore.ListInstancesRequest{}

	request.SetStatus("Running")

	response, err := a.client.ListInstances(&request)
	if err != nil {
		return err
	}

	if len(response.Body.Instances) == 0 {
		return errors.New("not found PAI-FeatureStore running instance")
	}
	var instanceId string

	for _, instance := range response.Body.Instances {
		instanceId = *instance.InstanceId
		break
	}

	if instanceId == "" {
		return fmt.Errorf("region:%s, not found PAI-FeatureStore instance", a.client.cfg.regionId)
	}

	a.client.instanceId = instanceId
	return nil
}
