package api

import (
	"context"

	"github.com/antihax/optional"
)

// Linger please
var (
	_ context.Context
)

type FlowCtrlApiService service

/*
FlowCtrlApiService 获取流控计划列表
 * @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 * @param optional nil or *FlowCtrlApiListFlowCtrlPlansOpts - Optional Parameters:
     * @param "SceneId" (optional.Int32) -
     * @param "Status" (optional.String) -
@return ListFlowCtrlPlansResponse
*/

type FlowCtrlApiListFlowCtrlPlansOpts struct {
	SceneId optional.Int32
	Status  optional.String
	Env     optional.String
}

/**
func (a *FlowCtrlApiService) ListFlowCtrlPlans(ctx context.Context, localVarOptionals *FlowCtrlApiListFlowCtrlPlansOpts) (ListFlowCtrlPlansResponse, error) {
	var (
		localVarHttpMethod  = strings.ToUpper("Get")
		localVarPostBody    interface{}
		localVarFileName    string
		localVarFileBytes   []byte
		localVarReturnValue ListFlowCtrlPlansResponse
	)

	// create path and map variables
	localVarPath := a.client.cfg.BasePath + "/flow_ctrl_plans/all"

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	if localVarOptionals != nil && localVarOptionals.SceneId.IsSet() {
		localVarQueryParams.Add("scene_id", parameterToString(localVarOptionals.SceneId.Value(), ""))
	}
	if localVarOptionals != nil && localVarOptionals.Status.IsSet() {
		localVarQueryParams.Add("status", parameterToString(localVarOptionals.Status.Value(), ""))
	}
	if localVarOptionals != nil && localVarOptionals.Env.IsSet() {
		localVarQueryParams.Add("env", parameterToString(localVarOptionals.Env.Value(), ""))
	}
	// to determine the Content-Type header
	localVarHttpContentTypes := []string{}

	// set Content-Type header
	localVarHttpContentType := selectHeaderContentType(localVarHttpContentTypes)
	if localVarHttpContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHttpContentType
	}

	// to determine the Accept header
	localVarHttpHeaderAccepts := []string{"application/json"}

	// set Accept header
	localVarHttpHeaderAccept := selectHeaderAccept(localVarHttpHeaderAccepts)
	if localVarHttpHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHttpHeaderAccept
	}
	r, err := a.client.prepareRequest(ctx, localVarPath, localVarHttpMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, localVarFileName, localVarFileBytes)
	if err != nil {
		return localVarReturnValue, err
	}

	localVarHttpResponse, err := a.client.callAPI(r)
	if err != nil || localVarHttpResponse == nil {
		return localVarReturnValue, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHttpResponse.Body)
	localVarHttpResponse.Body.Close()
	if err != nil {
		return localVarReturnValue, err
	}

	if localVarHttpResponse.StatusCode != 200 {
		err = a.client.decodeResponse(&localVarReturnValue, localVarBody)
		if err != nil {
			err = fmt.Errorf("failed to decode resp, err=%w, body=%s", err, string(localVarBody))
			return localVarReturnValue, err
		}

		return localVarReturnValue, errors.New(fmt.Sprintf("Http Status code:%d", localVarHttpResponse.StatusCode))
	} else {
		err = a.client.decodeResponse(&localVarReturnValue, localVarBody)
		if err != nil {
			err = fmt.Errorf("failed to decode resp, err=%w, body=%s", err, string(localVarBody))
			return localVarReturnValue, err
		}
	}
	return localVarReturnValue, nil
}

**/
