package pairecservice

//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
//
// Code generated by Alibaba Cloud SDK Code Generator.
// Changes may cause incorrect behavior and will be lost if the code is regenerated.

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/responses"
)

// OnlineExperiment invokes the pairecservice.OnlineExperiment API synchronously
func (client *Client) OnlineExperiment(request *OnlineExperimentRequest) (response *OnlineExperimentResponse, err error) {
	response = CreateOnlineExperimentResponse()
	err = client.DoAction(request, response)
	return
}

// OnlineExperimentWithChan invokes the pairecservice.OnlineExperiment API asynchronously
func (client *Client) OnlineExperimentWithChan(request *OnlineExperimentRequest) (<-chan *OnlineExperimentResponse, <-chan error) {
	responseChan := make(chan *OnlineExperimentResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.OnlineExperiment(request)
		if err != nil {
			errChan <- err
		} else {
			responseChan <- response
		}
	})
	if err != nil {
		errChan <- err
		close(responseChan)
		close(errChan)
	}
	return responseChan, errChan
}

// OnlineExperimentWithCallback invokes the pairecservice.OnlineExperiment API asynchronously
func (client *Client) OnlineExperimentWithCallback(request *OnlineExperimentRequest, callback func(response *OnlineExperimentResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *OnlineExperimentResponse
		var err error
		defer close(result)
		response, err = client.OnlineExperiment(request)
		callback(response, err)
		result <- 1
	})
	if err != nil {
		defer close(result)
		callback(nil, err)
		result <- 0
	}
	return result
}

// OnlineExperimentRequest is the request struct for api OnlineExperiment
type OnlineExperimentRequest struct {
	*requests.RoaRequest
	Body         string `position:"Body" name:"body"`
	ExperimentId string `position:"Path" name:"ExperimentId"`
}

// OnlineExperimentResponse is the response struct for api OnlineExperiment
type OnlineExperimentResponse struct {
	*responses.BaseResponse
	RequestId string `json:"RequestId" xml:"RequestId"`
}

// CreateOnlineExperimentRequest creates a request to invoke OnlineExperiment API
func CreateOnlineExperimentRequest() (request *OnlineExperimentRequest) {
	request = &OnlineExperimentRequest{
		RoaRequest: &requests.RoaRequest{},
	}
	request.InitWithApiInfo("PaiRecService", "2022-12-13", "OnlineExperiment", "/api/v1/experiments/[ExperimentId]/action/online", "", "")
	request.Method = requests.POST
	return
}

// CreateOnlineExperimentResponse creates a response to parse from OnlineExperiment response
func CreateOnlineExperimentResponse() (response *OnlineExperimentResponse) {
	response = &OnlineExperimentResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
