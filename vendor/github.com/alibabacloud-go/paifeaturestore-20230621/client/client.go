// This file is auto-generated, don't edit it. Thanks.
/**
 *
 */
package client

import (
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	endpointutil "github.com/alibabacloud-go/endpoint-util/service"
	openapiutil "github.com/alibabacloud-go/openapi-util/service"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
)

type FeatureViewConfigValue struct {
	Partitions map[string]map[string]interface{} `json:"Partitions,omitempty" xml:"Partitions,omitempty"`
}

func (s FeatureViewConfigValue) String() string {
	return tea.Prettify(s)
}

func (s FeatureViewConfigValue) GoString() string {
	return s.String()
}

func (s *FeatureViewConfigValue) SetPartitions(v map[string]map[string]interface{}) *FeatureViewConfigValue {
	s.Partitions = v
	return s
}

type ChangeProjectFeatureEntityHotIdVersionRequest struct {
	Version *string `json:"Version,omitempty" xml:"Version,omitempty"`
}

func (s ChangeProjectFeatureEntityHotIdVersionRequest) String() string {
	return tea.Prettify(s)
}

func (s ChangeProjectFeatureEntityHotIdVersionRequest) GoString() string {
	return s.String()
}

func (s *ChangeProjectFeatureEntityHotIdVersionRequest) SetVersion(v string) *ChangeProjectFeatureEntityHotIdVersionRequest {
	s.Version = &v
	return s
}

type ChangeProjectFeatureEntityHotIdVersionResponseBody struct {
	RequestId *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
}

func (s ChangeProjectFeatureEntityHotIdVersionResponseBody) String() string {
	return tea.Prettify(s)
}

func (s ChangeProjectFeatureEntityHotIdVersionResponseBody) GoString() string {
	return s.String()
}

func (s *ChangeProjectFeatureEntityHotIdVersionResponseBody) SetRequestId(v string) *ChangeProjectFeatureEntityHotIdVersionResponseBody {
	s.RequestId = &v
	return s
}

type ChangeProjectFeatureEntityHotIdVersionResponse struct {
	Headers    map[string]*string                                  `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
	StatusCode *int32                                              `json:"statusCode,omitempty" xml:"statusCode,omitempty" require:"true"`
	Body       *ChangeProjectFeatureEntityHotIdVersionResponseBody `json:"body,omitempty" xml:"body,omitempty" require:"true"`
}

func (s ChangeProjectFeatureEntityHotIdVersionResponse) String() string {
	return tea.Prettify(s)
}

func (s ChangeProjectFeatureEntityHotIdVersionResponse) GoString() string {
	return s.String()
}

func (s *ChangeProjectFeatureEntityHotIdVersionResponse) SetHeaders(v map[string]*string) *ChangeProjectFeatureEntityHotIdVersionResponse {
	s.Headers = v
	return s
}

func (s *ChangeProjectFeatureEntityHotIdVersionResponse) SetStatusCode(v int32) *ChangeProjectFeatureEntityHotIdVersionResponse {
	s.StatusCode = &v
	return s
}

func (s *ChangeProjectFeatureEntityHotIdVersionResponse) SetBody(v *ChangeProjectFeatureEntityHotIdVersionResponseBody) *ChangeProjectFeatureEntityHotIdVersionResponse {
	s.Body = v
	return s
}

type CheckInstanceDatasourceRequest struct {
	Config *string `json:"Config,omitempty" xml:"Config,omitempty"`
	Type   *string `json:"Type,omitempty" xml:"Type,omitempty"`
	Uri    *string `json:"Uri,omitempty" xml:"Uri,omitempty"`
}

func (s CheckInstanceDatasourceRequest) String() string {
	return tea.Prettify(s)
}

func (s CheckInstanceDatasourceRequest) GoString() string {
	return s.String()
}

func (s *CheckInstanceDatasourceRequest) SetConfig(v string) *CheckInstanceDatasourceRequest {
	s.Config = &v
	return s
}

func (s *CheckInstanceDatasourceRequest) SetType(v string) *CheckInstanceDatasourceRequest {
	s.Type = &v
	return s
}

func (s *CheckInstanceDatasourceRequest) SetUri(v string) *CheckInstanceDatasourceRequest {
	s.Uri = &v
	return s
}

type CheckInstanceDatasourceResponseBody struct {
	RequestId *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
	Status    *string `json:"Status,omitempty" xml:"Status,omitempty"`
}

func (s CheckInstanceDatasourceResponseBody) String() string {
	return tea.Prettify(s)
}

func (s CheckInstanceDatasourceResponseBody) GoString() string {
	return s.String()
}

func (s *CheckInstanceDatasourceResponseBody) SetRequestId(v string) *CheckInstanceDatasourceResponseBody {
	s.RequestId = &v
	return s
}

func (s *CheckInstanceDatasourceResponseBody) SetStatus(v string) *CheckInstanceDatasourceResponseBody {
	s.Status = &v
	return s
}

type CheckInstanceDatasourceResponse struct {
	Headers    map[string]*string                   `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
	StatusCode *int32                               `json:"statusCode,omitempty" xml:"statusCode,omitempty" require:"true"`
	Body       *CheckInstanceDatasourceResponseBody `json:"body,omitempty" xml:"body,omitempty" require:"true"`
}

func (s CheckInstanceDatasourceResponse) String() string {
	return tea.Prettify(s)
}

func (s CheckInstanceDatasourceResponse) GoString() string {
	return s.String()
}

func (s *CheckInstanceDatasourceResponse) SetHeaders(v map[string]*string) *CheckInstanceDatasourceResponse {
	s.Headers = v
	return s
}

func (s *CheckInstanceDatasourceResponse) SetStatusCode(v int32) *CheckInstanceDatasourceResponse {
	s.StatusCode = &v
	return s
}

func (s *CheckInstanceDatasourceResponse) SetBody(v *CheckInstanceDatasourceResponseBody) *CheckInstanceDatasourceResponse {
	s.Body = v
	return s
}

type CreateDatasourceRequest struct {
	Config      *string `json:"Config,omitempty" xml:"Config,omitempty"`
	Name        *string `json:"Name,omitempty" xml:"Name,omitempty"`
	Type        *string `json:"Type,omitempty" xml:"Type,omitempty"`
	Uri         *string `json:"Uri,omitempty" xml:"Uri,omitempty"`
	WorkspaceId *string `json:"WorkspaceId,omitempty" xml:"WorkspaceId,omitempty"`
}

func (s CreateDatasourceRequest) String() string {
	return tea.Prettify(s)
}

func (s CreateDatasourceRequest) GoString() string {
	return s.String()
}

func (s *CreateDatasourceRequest) SetConfig(v string) *CreateDatasourceRequest {
	s.Config = &v
	return s
}

func (s *CreateDatasourceRequest) SetName(v string) *CreateDatasourceRequest {
	s.Name = &v
	return s
}

func (s *CreateDatasourceRequest) SetType(v string) *CreateDatasourceRequest {
	s.Type = &v
	return s
}

func (s *CreateDatasourceRequest) SetUri(v string) *CreateDatasourceRequest {
	s.Uri = &v
	return s
}

func (s *CreateDatasourceRequest) SetWorkspaceId(v string) *CreateDatasourceRequest {
	s.WorkspaceId = &v
	return s
}

type CreateDatasourceResponseBody struct {
	DatasourceId *string `json:"DatasourceId,omitempty" xml:"DatasourceId,omitempty"`
	RequestId    *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
}

func (s CreateDatasourceResponseBody) String() string {
	return tea.Prettify(s)
}

func (s CreateDatasourceResponseBody) GoString() string {
	return s.String()
}

func (s *CreateDatasourceResponseBody) SetDatasourceId(v string) *CreateDatasourceResponseBody {
	s.DatasourceId = &v
	return s
}

func (s *CreateDatasourceResponseBody) SetRequestId(v string) *CreateDatasourceResponseBody {
	s.RequestId = &v
	return s
}

type CreateDatasourceResponse struct {
	Headers    map[string]*string            `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
	StatusCode *int32                        `json:"statusCode,omitempty" xml:"statusCode,omitempty" require:"true"`
	Body       *CreateDatasourceResponseBody `json:"body,omitempty" xml:"body,omitempty" require:"true"`
}

func (s CreateDatasourceResponse) String() string {
	return tea.Prettify(s)
}

func (s CreateDatasourceResponse) GoString() string {
	return s.String()
}

func (s *CreateDatasourceResponse) SetHeaders(v map[string]*string) *CreateDatasourceResponse {
	s.Headers = v
	return s
}

func (s *CreateDatasourceResponse) SetStatusCode(v int32) *CreateDatasourceResponse {
	s.StatusCode = &v
	return s
}

func (s *CreateDatasourceResponse) SetBody(v *CreateDatasourceResponseBody) *CreateDatasourceResponse {
	s.Body = v
	return s
}

type CreateFeatureEntityRequest struct {
	JoinId    *string `json:"JoinId,omitempty" xml:"JoinId,omitempty"`
	Name      *string `json:"Name,omitempty" xml:"Name,omitempty"`
	ProjectId *string `json:"ProjectId,omitempty" xml:"ProjectId,omitempty"`
}

func (s CreateFeatureEntityRequest) String() string {
	return tea.Prettify(s)
}

func (s CreateFeatureEntityRequest) GoString() string {
	return s.String()
}

func (s *CreateFeatureEntityRequest) SetJoinId(v string) *CreateFeatureEntityRequest {
	s.JoinId = &v
	return s
}

func (s *CreateFeatureEntityRequest) SetName(v string) *CreateFeatureEntityRequest {
	s.Name = &v
	return s
}

func (s *CreateFeatureEntityRequest) SetProjectId(v string) *CreateFeatureEntityRequest {
	s.ProjectId = &v
	return s
}

type CreateFeatureEntityResponseBody struct {
	FeatureEntityId *string `json:"FeatureEntityId,omitempty" xml:"FeatureEntityId,omitempty"`
	RequestId       *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
}

func (s CreateFeatureEntityResponseBody) String() string {
	return tea.Prettify(s)
}

func (s CreateFeatureEntityResponseBody) GoString() string {
	return s.String()
}

func (s *CreateFeatureEntityResponseBody) SetFeatureEntityId(v string) *CreateFeatureEntityResponseBody {
	s.FeatureEntityId = &v
	return s
}

func (s *CreateFeatureEntityResponseBody) SetRequestId(v string) *CreateFeatureEntityResponseBody {
	s.RequestId = &v
	return s
}

type CreateFeatureEntityResponse struct {
	Headers    map[string]*string               `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
	StatusCode *int32                           `json:"statusCode,omitempty" xml:"statusCode,omitempty" require:"true"`
	Body       *CreateFeatureEntityResponseBody `json:"body,omitempty" xml:"body,omitempty" require:"true"`
}

func (s CreateFeatureEntityResponse) String() string {
	return tea.Prettify(s)
}

func (s CreateFeatureEntityResponse) GoString() string {
	return s.String()
}

func (s *CreateFeatureEntityResponse) SetHeaders(v map[string]*string) *CreateFeatureEntityResponse {
	s.Headers = v
	return s
}

func (s *CreateFeatureEntityResponse) SetStatusCode(v int32) *CreateFeatureEntityResponse {
	s.StatusCode = &v
	return s
}

func (s *CreateFeatureEntityResponse) SetBody(v *CreateFeatureEntityResponseBody) *CreateFeatureEntityResponse {
	s.Body = v
	return s
}

type CreateFeatureViewRequest struct {
	Config               *string                           `json:"Config,omitempty" xml:"Config,omitempty"`
	FeatureEntityId      *string                           `json:"FeatureEntityId,omitempty" xml:"FeatureEntityId,omitempty"`
	Fields               []*CreateFeatureViewRequestFields `json:"Fields,omitempty" xml:"Fields,omitempty" type:"Repeated"`
	Name                 *string                           `json:"Name,omitempty" xml:"Name,omitempty"`
	ProjectId            *string                           `json:"ProjectId,omitempty" xml:"ProjectId,omitempty"`
	RegisterDatasourceId *string                           `json:"RegisterDatasourceId,omitempty" xml:"RegisterDatasourceId,omitempty"`
	RegisterTable        *string                           `json:"RegisterTable,omitempty" xml:"RegisterTable,omitempty"`
	SyncOnlineTable      *bool                             `json:"SyncOnlineTable,omitempty" xml:"SyncOnlineTable,omitempty"`
	TTL                  *int32                            `json:"TTL,omitempty" xml:"TTL,omitempty"`
	Tags                 []*string                         `json:"Tags,omitempty" xml:"Tags,omitempty" type:"Repeated"`
	Type                 *string                           `json:"Type,omitempty" xml:"Type,omitempty"`
	WriteMethod          *string                           `json:"WriteMethod,omitempty" xml:"WriteMethod,omitempty"`
}

func (s CreateFeatureViewRequest) String() string {
	return tea.Prettify(s)
}

func (s CreateFeatureViewRequest) GoString() string {
	return s.String()
}

func (s *CreateFeatureViewRequest) SetConfig(v string) *CreateFeatureViewRequest {
	s.Config = &v
	return s
}

func (s *CreateFeatureViewRequest) SetFeatureEntityId(v string) *CreateFeatureViewRequest {
	s.FeatureEntityId = &v
	return s
}

func (s *CreateFeatureViewRequest) SetFields(v []*CreateFeatureViewRequestFields) *CreateFeatureViewRequest {
	s.Fields = v
	return s
}

func (s *CreateFeatureViewRequest) SetName(v string) *CreateFeatureViewRequest {
	s.Name = &v
	return s
}

func (s *CreateFeatureViewRequest) SetProjectId(v string) *CreateFeatureViewRequest {
	s.ProjectId = &v
	return s
}

func (s *CreateFeatureViewRequest) SetRegisterDatasourceId(v string) *CreateFeatureViewRequest {
	s.RegisterDatasourceId = &v
	return s
}

func (s *CreateFeatureViewRequest) SetRegisterTable(v string) *CreateFeatureViewRequest {
	s.RegisterTable = &v
	return s
}

func (s *CreateFeatureViewRequest) SetSyncOnlineTable(v bool) *CreateFeatureViewRequest {
	s.SyncOnlineTable = &v
	return s
}

func (s *CreateFeatureViewRequest) SetTTL(v int32) *CreateFeatureViewRequest {
	s.TTL = &v
	return s
}

func (s *CreateFeatureViewRequest) SetTags(v []*string) *CreateFeatureViewRequest {
	s.Tags = v
	return s
}

func (s *CreateFeatureViewRequest) SetType(v string) *CreateFeatureViewRequest {
	s.Type = &v
	return s
}

func (s *CreateFeatureViewRequest) SetWriteMethod(v string) *CreateFeatureViewRequest {
	s.WriteMethod = &v
	return s
}

type CreateFeatureViewRequestFields struct {
	Attributes []*string `json:"Attributes,omitempty" xml:"Attributes,omitempty" type:"Repeated"`
	Name       *string   `json:"Name,omitempty" xml:"Name,omitempty"`
	Type       *string   `json:"Type,omitempty" xml:"Type,omitempty"`
}

func (s CreateFeatureViewRequestFields) String() string {
	return tea.Prettify(s)
}

func (s CreateFeatureViewRequestFields) GoString() string {
	return s.String()
}

func (s *CreateFeatureViewRequestFields) SetAttributes(v []*string) *CreateFeatureViewRequestFields {
	s.Attributes = v
	return s
}

func (s *CreateFeatureViewRequestFields) SetName(v string) *CreateFeatureViewRequestFields {
	s.Name = &v
	return s
}

func (s *CreateFeatureViewRequestFields) SetType(v string) *CreateFeatureViewRequestFields {
	s.Type = &v
	return s
}

type CreateFeatureViewResponseBody struct {
	FeatureViewId *string `json:"FeatureViewId,omitempty" xml:"FeatureViewId,omitempty"`
	RequestId     *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
}

func (s CreateFeatureViewResponseBody) String() string {
	return tea.Prettify(s)
}

func (s CreateFeatureViewResponseBody) GoString() string {
	return s.String()
}

func (s *CreateFeatureViewResponseBody) SetFeatureViewId(v string) *CreateFeatureViewResponseBody {
	s.FeatureViewId = &v
	return s
}

func (s *CreateFeatureViewResponseBody) SetRequestId(v string) *CreateFeatureViewResponseBody {
	s.RequestId = &v
	return s
}

type CreateFeatureViewResponse struct {
	Headers    map[string]*string             `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
	StatusCode *int32                         `json:"statusCode,omitempty" xml:"statusCode,omitempty" require:"true"`
	Body       *CreateFeatureViewResponseBody `json:"body,omitempty" xml:"body,omitempty" require:"true"`
}

func (s CreateFeatureViewResponse) String() string {
	return tea.Prettify(s)
}

func (s CreateFeatureViewResponse) GoString() string {
	return s.String()
}

func (s *CreateFeatureViewResponse) SetHeaders(v map[string]*string) *CreateFeatureViewResponse {
	s.Headers = v
	return s
}

func (s *CreateFeatureViewResponse) SetStatusCode(v int32) *CreateFeatureViewResponse {
	s.StatusCode = &v
	return s
}

func (s *CreateFeatureViewResponse) SetBody(v *CreateFeatureViewResponseBody) *CreateFeatureViewResponse {
	s.Body = v
	return s
}

type CreateInstanceRequest struct {
	Type *string `json:"Type,omitempty" xml:"Type,omitempty"`
}

func (s CreateInstanceRequest) String() string {
	return tea.Prettify(s)
}

func (s CreateInstanceRequest) GoString() string {
	return s.String()
}

func (s *CreateInstanceRequest) SetType(v string) *CreateInstanceRequest {
	s.Type = &v
	return s
}

type CreateInstanceResponseBody struct {
	Code       *string `json:"Code,omitempty" xml:"Code,omitempty"`
	InstanceId *string `json:"InstanceId,omitempty" xml:"InstanceId,omitempty"`
	RequestId  *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
}

func (s CreateInstanceResponseBody) String() string {
	return tea.Prettify(s)
}

func (s CreateInstanceResponseBody) GoString() string {
	return s.String()
}

func (s *CreateInstanceResponseBody) SetCode(v string) *CreateInstanceResponseBody {
	s.Code = &v
	return s
}

func (s *CreateInstanceResponseBody) SetInstanceId(v string) *CreateInstanceResponseBody {
	s.InstanceId = &v
	return s
}

func (s *CreateInstanceResponseBody) SetRequestId(v string) *CreateInstanceResponseBody {
	s.RequestId = &v
	return s
}

type CreateInstanceResponse struct {
	Headers    map[string]*string          `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
	StatusCode *int32                      `json:"statusCode,omitempty" xml:"statusCode,omitempty" require:"true"`
	Body       *CreateInstanceResponseBody `json:"body,omitempty" xml:"body,omitempty" require:"true"`
}

func (s CreateInstanceResponse) String() string {
	return tea.Prettify(s)
}

func (s CreateInstanceResponse) GoString() string {
	return s.String()
}

func (s *CreateInstanceResponse) SetHeaders(v map[string]*string) *CreateInstanceResponse {
	s.Headers = v
	return s
}

func (s *CreateInstanceResponse) SetStatusCode(v int32) *CreateInstanceResponse {
	s.StatusCode = &v
	return s
}

func (s *CreateInstanceResponse) SetBody(v *CreateInstanceResponseBody) *CreateInstanceResponse {
	s.Body = v
	return s
}

type CreateLabelTableRequest struct {
	DatasourceId *string                          `json:"DatasourceId,omitempty" xml:"DatasourceId,omitempty"`
	Fields       []*CreateLabelTableRequestFields `json:"Fields,omitempty" xml:"Fields,omitempty" type:"Repeated"`
	Name         *string                          `json:"Name,omitempty" xml:"Name,omitempty"`
	ProjectId    *string                          `json:"ProjectId,omitempty" xml:"ProjectId,omitempty"`
}

func (s CreateLabelTableRequest) String() string {
	return tea.Prettify(s)
}

func (s CreateLabelTableRequest) GoString() string {
	return s.String()
}

func (s *CreateLabelTableRequest) SetDatasourceId(v string) *CreateLabelTableRequest {
	s.DatasourceId = &v
	return s
}

func (s *CreateLabelTableRequest) SetFields(v []*CreateLabelTableRequestFields) *CreateLabelTableRequest {
	s.Fields = v
	return s
}

func (s *CreateLabelTableRequest) SetName(v string) *CreateLabelTableRequest {
	s.Name = &v
	return s
}

func (s *CreateLabelTableRequest) SetProjectId(v string) *CreateLabelTableRequest {
	s.ProjectId = &v
	return s
}

type CreateLabelTableRequestFields struct {
	Attributes []*string `json:"Attributes,omitempty" xml:"Attributes,omitempty" type:"Repeated"`
	Name       *string   `json:"Name,omitempty" xml:"Name,omitempty"`
	Type       *string   `json:"Type,omitempty" xml:"Type,omitempty"`
}

func (s CreateLabelTableRequestFields) String() string {
	return tea.Prettify(s)
}

func (s CreateLabelTableRequestFields) GoString() string {
	return s.String()
}

func (s *CreateLabelTableRequestFields) SetAttributes(v []*string) *CreateLabelTableRequestFields {
	s.Attributes = v
	return s
}

func (s *CreateLabelTableRequestFields) SetName(v string) *CreateLabelTableRequestFields {
	s.Name = &v
	return s
}

func (s *CreateLabelTableRequestFields) SetType(v string) *CreateLabelTableRequestFields {
	s.Type = &v
	return s
}

type CreateLabelTableResponseBody struct {
	LabelTableId *string `json:"LabelTableId,omitempty" xml:"LabelTableId,omitempty"`
	RequestId    *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
}

func (s CreateLabelTableResponseBody) String() string {
	return tea.Prettify(s)
}

func (s CreateLabelTableResponseBody) GoString() string {
	return s.String()
}

func (s *CreateLabelTableResponseBody) SetLabelTableId(v string) *CreateLabelTableResponseBody {
	s.LabelTableId = &v
	return s
}

func (s *CreateLabelTableResponseBody) SetRequestId(v string) *CreateLabelTableResponseBody {
	s.RequestId = &v
	return s
}

type CreateLabelTableResponse struct {
	Headers    map[string]*string            `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
	StatusCode *int32                        `json:"statusCode,omitempty" xml:"statusCode,omitempty" require:"true"`
	Body       *CreateLabelTableResponseBody `json:"body,omitempty" xml:"body,omitempty" require:"true"`
}

func (s CreateLabelTableResponse) String() string {
	return tea.Prettify(s)
}

func (s CreateLabelTableResponse) GoString() string {
	return s.String()
}

func (s *CreateLabelTableResponse) SetHeaders(v map[string]*string) *CreateLabelTableResponse {
	s.Headers = v
	return s
}

func (s *CreateLabelTableResponse) SetStatusCode(v int32) *CreateLabelTableResponse {
	s.StatusCode = &v
	return s
}

func (s *CreateLabelTableResponse) SetBody(v *CreateLabelTableResponseBody) *CreateLabelTableResponse {
	s.Body = v
	return s
}

type CreateModelFeatureRequest struct {
	Features     []*CreateModelFeatureRequestFeatures `json:"Features,omitempty" xml:"Features,omitempty" type:"Repeated"`
	LabelTableId *string                              `json:"LabelTableId,omitempty" xml:"LabelTableId,omitempty"`
	Name         *string                              `json:"Name,omitempty" xml:"Name,omitempty"`
	ProjectId    *string                              `json:"ProjectId,omitempty" xml:"ProjectId,omitempty"`
}

func (s CreateModelFeatureRequest) String() string {
	return tea.Prettify(s)
}

func (s CreateModelFeatureRequest) GoString() string {
	return s.String()
}

func (s *CreateModelFeatureRequest) SetFeatures(v []*CreateModelFeatureRequestFeatures) *CreateModelFeatureRequest {
	s.Features = v
	return s
}

func (s *CreateModelFeatureRequest) SetLabelTableId(v string) *CreateModelFeatureRequest {
	s.LabelTableId = &v
	return s
}

func (s *CreateModelFeatureRequest) SetName(v string) *CreateModelFeatureRequest {
	s.Name = &v
	return s
}

func (s *CreateModelFeatureRequest) SetProjectId(v string) *CreateModelFeatureRequest {
	s.ProjectId = &v
	return s
}

type CreateModelFeatureRequestFeatures struct {
	AliasName     *string `json:"AliasName,omitempty" xml:"AliasName,omitempty"`
	FeatureViewId *string `json:"FeatureViewId,omitempty" xml:"FeatureViewId,omitempty"`
	Name          *string `json:"Name,omitempty" xml:"Name,omitempty"`
	Type          *string `json:"Type,omitempty" xml:"Type,omitempty"`
}

func (s CreateModelFeatureRequestFeatures) String() string {
	return tea.Prettify(s)
}

func (s CreateModelFeatureRequestFeatures) GoString() string {
	return s.String()
}

func (s *CreateModelFeatureRequestFeatures) SetAliasName(v string) *CreateModelFeatureRequestFeatures {
	s.AliasName = &v
	return s
}

func (s *CreateModelFeatureRequestFeatures) SetFeatureViewId(v string) *CreateModelFeatureRequestFeatures {
	s.FeatureViewId = &v
	return s
}

func (s *CreateModelFeatureRequestFeatures) SetName(v string) *CreateModelFeatureRequestFeatures {
	s.Name = &v
	return s
}

func (s *CreateModelFeatureRequestFeatures) SetType(v string) *CreateModelFeatureRequestFeatures {
	s.Type = &v
	return s
}

type CreateModelFeatureResponseBody struct {
	ModelFeatureId *string `json:"ModelFeatureId,omitempty" xml:"ModelFeatureId,omitempty"`
	RequestId      *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
}

func (s CreateModelFeatureResponseBody) String() string {
	return tea.Prettify(s)
}

func (s CreateModelFeatureResponseBody) GoString() string {
	return s.String()
}

func (s *CreateModelFeatureResponseBody) SetModelFeatureId(v string) *CreateModelFeatureResponseBody {
	s.ModelFeatureId = &v
	return s
}

func (s *CreateModelFeatureResponseBody) SetRequestId(v string) *CreateModelFeatureResponseBody {
	s.RequestId = &v
	return s
}

type CreateModelFeatureResponse struct {
	Headers    map[string]*string              `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
	StatusCode *int32                          `json:"statusCode,omitempty" xml:"statusCode,omitempty" require:"true"`
	Body       *CreateModelFeatureResponseBody `json:"body,omitempty" xml:"body,omitempty" require:"true"`
}

func (s CreateModelFeatureResponse) String() string {
	return tea.Prettify(s)
}

func (s CreateModelFeatureResponse) GoString() string {
	return s.String()
}

func (s *CreateModelFeatureResponse) SetHeaders(v map[string]*string) *CreateModelFeatureResponse {
	s.Headers = v
	return s
}

func (s *CreateModelFeatureResponse) SetStatusCode(v int32) *CreateModelFeatureResponse {
	s.StatusCode = &v
	return s
}

func (s *CreateModelFeatureResponse) SetBody(v *CreateModelFeatureResponseBody) *CreateModelFeatureResponse {
	s.Body = v
	return s
}

type CreateProjectRequest struct {
	Description         *string `json:"Description,omitempty" xml:"Description,omitempty"`
	Name                *string `json:"Name,omitempty" xml:"Name,omitempty"`
	OfflineDatasourceId *string `json:"OfflineDatasourceId,omitempty" xml:"OfflineDatasourceId,omitempty"`
	OfflineLifeCycle    *int32  `json:"OfflineLifeCycle,omitempty" xml:"OfflineLifeCycle,omitempty"`
	OnlineDatasourceId  *string `json:"OnlineDatasourceId,omitempty" xml:"OnlineDatasourceId,omitempty"`
	WorkspaceId         *string `json:"WorkspaceId,omitempty" xml:"WorkspaceId,omitempty"`
}

func (s CreateProjectRequest) String() string {
	return tea.Prettify(s)
}

func (s CreateProjectRequest) GoString() string {
	return s.String()
}

func (s *CreateProjectRequest) SetDescription(v string) *CreateProjectRequest {
	s.Description = &v
	return s
}

func (s *CreateProjectRequest) SetName(v string) *CreateProjectRequest {
	s.Name = &v
	return s
}

func (s *CreateProjectRequest) SetOfflineDatasourceId(v string) *CreateProjectRequest {
	s.OfflineDatasourceId = &v
	return s
}

func (s *CreateProjectRequest) SetOfflineLifeCycle(v int32) *CreateProjectRequest {
	s.OfflineLifeCycle = &v
	return s
}

func (s *CreateProjectRequest) SetOnlineDatasourceId(v string) *CreateProjectRequest {
	s.OnlineDatasourceId = &v
	return s
}

func (s *CreateProjectRequest) SetWorkspaceId(v string) *CreateProjectRequest {
	s.WorkspaceId = &v
	return s
}

type CreateProjectResponseBody struct {
	ProjectId *string `json:"ProjectId,omitempty" xml:"ProjectId,omitempty"`
	RequestId *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
}

func (s CreateProjectResponseBody) String() string {
	return tea.Prettify(s)
}

func (s CreateProjectResponseBody) GoString() string {
	return s.String()
}

func (s *CreateProjectResponseBody) SetProjectId(v string) *CreateProjectResponseBody {
	s.ProjectId = &v
	return s
}

func (s *CreateProjectResponseBody) SetRequestId(v string) *CreateProjectResponseBody {
	s.RequestId = &v
	return s
}

type CreateProjectResponse struct {
	Headers    map[string]*string         `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
	StatusCode *int32                     `json:"statusCode,omitempty" xml:"statusCode,omitempty" require:"true"`
	Body       *CreateProjectResponseBody `json:"body,omitempty" xml:"body,omitempty" require:"true"`
}

func (s CreateProjectResponse) String() string {
	return tea.Prettify(s)
}

func (s CreateProjectResponse) GoString() string {
	return s.String()
}

func (s *CreateProjectResponse) SetHeaders(v map[string]*string) *CreateProjectResponse {
	s.Headers = v
	return s
}

func (s *CreateProjectResponse) SetStatusCode(v int32) *CreateProjectResponse {
	s.StatusCode = &v
	return s
}

func (s *CreateProjectResponse) SetBody(v *CreateProjectResponseBody) *CreateProjectResponse {
	s.Body = v
	return s
}

type CreateServiceIdentityRoleRequest struct {
	RoleName *string `json:"RoleName,omitempty" xml:"RoleName,omitempty"`
}

func (s CreateServiceIdentityRoleRequest) String() string {
	return tea.Prettify(s)
}

func (s CreateServiceIdentityRoleRequest) GoString() string {
	return s.String()
}

func (s *CreateServiceIdentityRoleRequest) SetRoleName(v string) *CreateServiceIdentityRoleRequest {
	s.RoleName = &v
	return s
}

type CreateServiceIdentityRoleResponseBody struct {
	Code      *string `json:"Code,omitempty" xml:"Code,omitempty"`
	RequestId *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
	RoleName  *string `json:"RoleName,omitempty" xml:"RoleName,omitempty"`
}

func (s CreateServiceIdentityRoleResponseBody) String() string {
	return tea.Prettify(s)
}

func (s CreateServiceIdentityRoleResponseBody) GoString() string {
	return s.String()
}

func (s *CreateServiceIdentityRoleResponseBody) SetCode(v string) *CreateServiceIdentityRoleResponseBody {
	s.Code = &v
	return s
}

func (s *CreateServiceIdentityRoleResponseBody) SetRequestId(v string) *CreateServiceIdentityRoleResponseBody {
	s.RequestId = &v
	return s
}

func (s *CreateServiceIdentityRoleResponseBody) SetRoleName(v string) *CreateServiceIdentityRoleResponseBody {
	s.RoleName = &v
	return s
}

type CreateServiceIdentityRoleResponse struct {
	Headers    map[string]*string                     `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
	StatusCode *int32                                 `json:"statusCode,omitempty" xml:"statusCode,omitempty" require:"true"`
	Body       *CreateServiceIdentityRoleResponseBody `json:"body,omitempty" xml:"body,omitempty" require:"true"`
}

func (s CreateServiceIdentityRoleResponse) String() string {
	return tea.Prettify(s)
}

func (s CreateServiceIdentityRoleResponse) GoString() string {
	return s.String()
}

func (s *CreateServiceIdentityRoleResponse) SetHeaders(v map[string]*string) *CreateServiceIdentityRoleResponse {
	s.Headers = v
	return s
}

func (s *CreateServiceIdentityRoleResponse) SetStatusCode(v int32) *CreateServiceIdentityRoleResponse {
	s.StatusCode = &v
	return s
}

func (s *CreateServiceIdentityRoleResponse) SetBody(v *CreateServiceIdentityRoleResponseBody) *CreateServiceIdentityRoleResponse {
	s.Body = v
	return s
}

type DeleteDatasourceResponseBody struct {
	RequestId *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
}

func (s DeleteDatasourceResponseBody) String() string {
	return tea.Prettify(s)
}

func (s DeleteDatasourceResponseBody) GoString() string {
	return s.String()
}

func (s *DeleteDatasourceResponseBody) SetRequestId(v string) *DeleteDatasourceResponseBody {
	s.RequestId = &v
	return s
}

type DeleteDatasourceResponse struct {
	Headers    map[string]*string            `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
	StatusCode *int32                        `json:"statusCode,omitempty" xml:"statusCode,omitempty" require:"true"`
	Body       *DeleteDatasourceResponseBody `json:"body,omitempty" xml:"body,omitempty" require:"true"`
}

func (s DeleteDatasourceResponse) String() string {
	return tea.Prettify(s)
}

func (s DeleteDatasourceResponse) GoString() string {
	return s.String()
}

func (s *DeleteDatasourceResponse) SetHeaders(v map[string]*string) *DeleteDatasourceResponse {
	s.Headers = v
	return s
}

func (s *DeleteDatasourceResponse) SetStatusCode(v int32) *DeleteDatasourceResponse {
	s.StatusCode = &v
	return s
}

func (s *DeleteDatasourceResponse) SetBody(v *DeleteDatasourceResponseBody) *DeleteDatasourceResponse {
	s.Body = v
	return s
}

type DeleteFeatureEntityResponseBody struct {
	RequestId *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
}

func (s DeleteFeatureEntityResponseBody) String() string {
	return tea.Prettify(s)
}

func (s DeleteFeatureEntityResponseBody) GoString() string {
	return s.String()
}

func (s *DeleteFeatureEntityResponseBody) SetRequestId(v string) *DeleteFeatureEntityResponseBody {
	s.RequestId = &v
	return s
}

type DeleteFeatureEntityResponse struct {
	Headers    map[string]*string               `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
	StatusCode *int32                           `json:"statusCode,omitempty" xml:"statusCode,omitempty" require:"true"`
	Body       *DeleteFeatureEntityResponseBody `json:"body,omitempty" xml:"body,omitempty" require:"true"`
}

func (s DeleteFeatureEntityResponse) String() string {
	return tea.Prettify(s)
}

func (s DeleteFeatureEntityResponse) GoString() string {
	return s.String()
}

func (s *DeleteFeatureEntityResponse) SetHeaders(v map[string]*string) *DeleteFeatureEntityResponse {
	s.Headers = v
	return s
}

func (s *DeleteFeatureEntityResponse) SetStatusCode(v int32) *DeleteFeatureEntityResponse {
	s.StatusCode = &v
	return s
}

func (s *DeleteFeatureEntityResponse) SetBody(v *DeleteFeatureEntityResponseBody) *DeleteFeatureEntityResponse {
	s.Body = v
	return s
}

type DeleteFeatureViewResponseBody struct {
	RequestId *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
}

func (s DeleteFeatureViewResponseBody) String() string {
	return tea.Prettify(s)
}

func (s DeleteFeatureViewResponseBody) GoString() string {
	return s.String()
}

func (s *DeleteFeatureViewResponseBody) SetRequestId(v string) *DeleteFeatureViewResponseBody {
	s.RequestId = &v
	return s
}

type DeleteFeatureViewResponse struct {
	Headers    map[string]*string             `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
	StatusCode *int32                         `json:"statusCode,omitempty" xml:"statusCode,omitempty" require:"true"`
	Body       *DeleteFeatureViewResponseBody `json:"body,omitempty" xml:"body,omitempty" require:"true"`
}

func (s DeleteFeatureViewResponse) String() string {
	return tea.Prettify(s)
}

func (s DeleteFeatureViewResponse) GoString() string {
	return s.String()
}

func (s *DeleteFeatureViewResponse) SetHeaders(v map[string]*string) *DeleteFeatureViewResponse {
	s.Headers = v
	return s
}

func (s *DeleteFeatureViewResponse) SetStatusCode(v int32) *DeleteFeatureViewResponse {
	s.StatusCode = &v
	return s
}

func (s *DeleteFeatureViewResponse) SetBody(v *DeleteFeatureViewResponseBody) *DeleteFeatureViewResponse {
	s.Body = v
	return s
}

type DeleteLabelTableResponseBody struct {
	RequestId *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
}

func (s DeleteLabelTableResponseBody) String() string {
	return tea.Prettify(s)
}

func (s DeleteLabelTableResponseBody) GoString() string {
	return s.String()
}

func (s *DeleteLabelTableResponseBody) SetRequestId(v string) *DeleteLabelTableResponseBody {
	s.RequestId = &v
	return s
}

type DeleteLabelTableResponse struct {
	Headers    map[string]*string            `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
	StatusCode *int32                        `json:"statusCode,omitempty" xml:"statusCode,omitempty" require:"true"`
	Body       *DeleteLabelTableResponseBody `json:"body,omitempty" xml:"body,omitempty" require:"true"`
}

func (s DeleteLabelTableResponse) String() string {
	return tea.Prettify(s)
}

func (s DeleteLabelTableResponse) GoString() string {
	return s.String()
}

func (s *DeleteLabelTableResponse) SetHeaders(v map[string]*string) *DeleteLabelTableResponse {
	s.Headers = v
	return s
}

func (s *DeleteLabelTableResponse) SetStatusCode(v int32) *DeleteLabelTableResponse {
	s.StatusCode = &v
	return s
}

func (s *DeleteLabelTableResponse) SetBody(v *DeleteLabelTableResponseBody) *DeleteLabelTableResponse {
	s.Body = v
	return s
}

type DeleteModelFeatureResponseBody struct {
	RequestId *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
}

func (s DeleteModelFeatureResponseBody) String() string {
	return tea.Prettify(s)
}

func (s DeleteModelFeatureResponseBody) GoString() string {
	return s.String()
}

func (s *DeleteModelFeatureResponseBody) SetRequestId(v string) *DeleteModelFeatureResponseBody {
	s.RequestId = &v
	return s
}

type DeleteModelFeatureResponse struct {
	Headers    map[string]*string              `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
	StatusCode *int32                          `json:"statusCode,omitempty" xml:"statusCode,omitempty" require:"true"`
	Body       *DeleteModelFeatureResponseBody `json:"body,omitempty" xml:"body,omitempty" require:"true"`
}

func (s DeleteModelFeatureResponse) String() string {
	return tea.Prettify(s)
}

func (s DeleteModelFeatureResponse) GoString() string {
	return s.String()
}

func (s *DeleteModelFeatureResponse) SetHeaders(v map[string]*string) *DeleteModelFeatureResponse {
	s.Headers = v
	return s
}

func (s *DeleteModelFeatureResponse) SetStatusCode(v int32) *DeleteModelFeatureResponse {
	s.StatusCode = &v
	return s
}

func (s *DeleteModelFeatureResponse) SetBody(v *DeleteModelFeatureResponseBody) *DeleteModelFeatureResponse {
	s.Body = v
	return s
}

type DeleteProjectResponseBody struct {
	RequestId *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
}

func (s DeleteProjectResponseBody) String() string {
	return tea.Prettify(s)
}

func (s DeleteProjectResponseBody) GoString() string {
	return s.String()
}

func (s *DeleteProjectResponseBody) SetRequestId(v string) *DeleteProjectResponseBody {
	s.RequestId = &v
	return s
}

type DeleteProjectResponse struct {
	Headers    map[string]*string         `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
	StatusCode *int32                     `json:"statusCode,omitempty" xml:"statusCode,omitempty" require:"true"`
	Body       *DeleteProjectResponseBody `json:"body,omitempty" xml:"body,omitempty" require:"true"`
}

func (s DeleteProjectResponse) String() string {
	return tea.Prettify(s)
}

func (s DeleteProjectResponse) GoString() string {
	return s.String()
}

func (s *DeleteProjectResponse) SetHeaders(v map[string]*string) *DeleteProjectResponse {
	s.Headers = v
	return s
}

func (s *DeleteProjectResponse) SetStatusCode(v int32) *DeleteProjectResponse {
	s.StatusCode = &v
	return s
}

func (s *DeleteProjectResponse) SetBody(v *DeleteProjectResponseBody) *DeleteProjectResponse {
	s.Body = v
	return s
}

type ExportModelFeatureTrainingSetTableRequest struct {
	FeatureViewConfig map[string]*FeatureViewConfigValue                          `json:"FeatureViewConfig,omitempty" xml:"FeatureViewConfig,omitempty"`
	LabelInputConfig  *ExportModelFeatureTrainingSetTableRequestLabelInputConfig  `json:"LabelInputConfig,omitempty" xml:"LabelInputConfig,omitempty" type:"Struct"`
	TrainingSetConfig *ExportModelFeatureTrainingSetTableRequestTrainingSetConfig `json:"TrainingSetConfig,omitempty" xml:"TrainingSetConfig,omitempty" type:"Struct"`
}

func (s ExportModelFeatureTrainingSetTableRequest) String() string {
	return tea.Prettify(s)
}

func (s ExportModelFeatureTrainingSetTableRequest) GoString() string {
	return s.String()
}

func (s *ExportModelFeatureTrainingSetTableRequest) SetFeatureViewConfig(v map[string]*FeatureViewConfigValue) *ExportModelFeatureTrainingSetTableRequest {
	s.FeatureViewConfig = v
	return s
}

func (s *ExportModelFeatureTrainingSetTableRequest) SetLabelInputConfig(v *ExportModelFeatureTrainingSetTableRequestLabelInputConfig) *ExportModelFeatureTrainingSetTableRequest {
	s.LabelInputConfig = v
	return s
}

func (s *ExportModelFeatureTrainingSetTableRequest) SetTrainingSetConfig(v *ExportModelFeatureTrainingSetTableRequestTrainingSetConfig) *ExportModelFeatureTrainingSetTableRequest {
	s.TrainingSetConfig = v
	return s
}

type ExportModelFeatureTrainingSetTableRequestLabelInputConfig struct {
	EventTime  *string                           `json:"EventTime,omitempty" xml:"EventTime,omitempty"`
	Partitions map[string]map[string]interface{} `json:"Partitions,omitempty" xml:"Partitions,omitempty"`
}

func (s ExportModelFeatureTrainingSetTableRequestLabelInputConfig) String() string {
	return tea.Prettify(s)
}

func (s ExportModelFeatureTrainingSetTableRequestLabelInputConfig) GoString() string {
	return s.String()
}

func (s *ExportModelFeatureTrainingSetTableRequestLabelInputConfig) SetEventTime(v string) *ExportModelFeatureTrainingSetTableRequestLabelInputConfig {
	s.EventTime = &v
	return s
}

func (s *ExportModelFeatureTrainingSetTableRequestLabelInputConfig) SetPartitions(v map[string]map[string]interface{}) *ExportModelFeatureTrainingSetTableRequestLabelInputConfig {
	s.Partitions = v
	return s
}

type ExportModelFeatureTrainingSetTableRequestTrainingSetConfig struct {
	Partitions map[string]map[string]interface{} `json:"Partitions,omitempty" xml:"Partitions,omitempty"`
}

func (s ExportModelFeatureTrainingSetTableRequestTrainingSetConfig) String() string {
	return tea.Prettify(s)
}

func (s ExportModelFeatureTrainingSetTableRequestTrainingSetConfig) GoString() string {
	return s.String()
}

func (s *ExportModelFeatureTrainingSetTableRequestTrainingSetConfig) SetPartitions(v map[string]map[string]interface{}) *ExportModelFeatureTrainingSetTableRequestTrainingSetConfig {
	s.Partitions = v
	return s
}

type ExportModelFeatureTrainingSetTableResponseBody struct {
	RequestId *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
}

func (s ExportModelFeatureTrainingSetTableResponseBody) String() string {
	return tea.Prettify(s)
}

func (s ExportModelFeatureTrainingSetTableResponseBody) GoString() string {
	return s.String()
}

func (s *ExportModelFeatureTrainingSetTableResponseBody) SetRequestId(v string) *ExportModelFeatureTrainingSetTableResponseBody {
	s.RequestId = &v
	return s
}

type ExportModelFeatureTrainingSetTableResponse struct {
	Headers    map[string]*string                              `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
	StatusCode *int32                                          `json:"statusCode,omitempty" xml:"statusCode,omitempty" require:"true"`
	Body       *ExportModelFeatureTrainingSetTableResponseBody `json:"body,omitempty" xml:"body,omitempty" require:"true"`
}

func (s ExportModelFeatureTrainingSetTableResponse) String() string {
	return tea.Prettify(s)
}

func (s ExportModelFeatureTrainingSetTableResponse) GoString() string {
	return s.String()
}

func (s *ExportModelFeatureTrainingSetTableResponse) SetHeaders(v map[string]*string) *ExportModelFeatureTrainingSetTableResponse {
	s.Headers = v
	return s
}

func (s *ExportModelFeatureTrainingSetTableResponse) SetStatusCode(v int32) *ExportModelFeatureTrainingSetTableResponse {
	s.StatusCode = &v
	return s
}

func (s *ExportModelFeatureTrainingSetTableResponse) SetBody(v *ExportModelFeatureTrainingSetTableResponseBody) *ExportModelFeatureTrainingSetTableResponse {
	s.Body = v
	return s
}

type GetDatasourceResponseBody struct {
	Config       *string `json:"Config,omitempty" xml:"Config,omitempty"`
	DatasourceId *string `json:"DatasourceId,omitempty" xml:"DatasourceId,omitempty"`
	Name         *string `json:"Name,omitempty" xml:"Name,omitempty"`
	RequestId    *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
	Type         *string `json:"Type,omitempty" xml:"Type,omitempty"`
	Uri          *string `json:"Uri,omitempty" xml:"Uri,omitempty"`
	WorkspaceId  *string `json:"WorkspaceId,omitempty" xml:"WorkspaceId,omitempty"`
}

func (s GetDatasourceResponseBody) String() string {
	return tea.Prettify(s)
}

func (s GetDatasourceResponseBody) GoString() string {
	return s.String()
}

func (s *GetDatasourceResponseBody) SetConfig(v string) *GetDatasourceResponseBody {
	s.Config = &v
	return s
}

func (s *GetDatasourceResponseBody) SetDatasourceId(v string) *GetDatasourceResponseBody {
	s.DatasourceId = &v
	return s
}

func (s *GetDatasourceResponseBody) SetName(v string) *GetDatasourceResponseBody {
	s.Name = &v
	return s
}

func (s *GetDatasourceResponseBody) SetRequestId(v string) *GetDatasourceResponseBody {
	s.RequestId = &v
	return s
}

func (s *GetDatasourceResponseBody) SetType(v string) *GetDatasourceResponseBody {
	s.Type = &v
	return s
}

func (s *GetDatasourceResponseBody) SetUri(v string) *GetDatasourceResponseBody {
	s.Uri = &v
	return s
}

func (s *GetDatasourceResponseBody) SetWorkspaceId(v string) *GetDatasourceResponseBody {
	s.WorkspaceId = &v
	return s
}

type GetDatasourceResponse struct {
	Headers    map[string]*string         `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
	StatusCode *int32                     `json:"statusCode,omitempty" xml:"statusCode,omitempty" require:"true"`
	Body       *GetDatasourceResponseBody `json:"body,omitempty" xml:"body,omitempty" require:"true"`
}

func (s GetDatasourceResponse) String() string {
	return tea.Prettify(s)
}

func (s GetDatasourceResponse) GoString() string {
	return s.String()
}

func (s *GetDatasourceResponse) SetHeaders(v map[string]*string) *GetDatasourceResponse {
	s.Headers = v
	return s
}

func (s *GetDatasourceResponse) SetStatusCode(v int32) *GetDatasourceResponse {
	s.StatusCode = &v
	return s
}

func (s *GetDatasourceResponse) SetBody(v *GetDatasourceResponseBody) *GetDatasourceResponse {
	s.Body = v
	return s
}

type GetDatasourceTableResponseBody struct {
	Fields    []*GetDatasourceTableResponseBodyFields `json:"Fields,omitempty" xml:"Fields,omitempty" type:"Repeated"`
	RequestId *string                                 `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
	TableName *string                                 `json:"TableName,omitempty" xml:"TableName,omitempty"`
}

func (s GetDatasourceTableResponseBody) String() string {
	return tea.Prettify(s)
}

func (s GetDatasourceTableResponseBody) GoString() string {
	return s.String()
}

func (s *GetDatasourceTableResponseBody) SetFields(v []*GetDatasourceTableResponseBodyFields) *GetDatasourceTableResponseBody {
	s.Fields = v
	return s
}

func (s *GetDatasourceTableResponseBody) SetRequestId(v string) *GetDatasourceTableResponseBody {
	s.RequestId = &v
	return s
}

func (s *GetDatasourceTableResponseBody) SetTableName(v string) *GetDatasourceTableResponseBody {
	s.TableName = &v
	return s
}

type GetDatasourceTableResponseBodyFields struct {
	Attributes []*string `json:"Attributes,omitempty" xml:"Attributes,omitempty" type:"Repeated"`
	Name       *string   `json:"Name,omitempty" xml:"Name,omitempty"`
	Type       *string   `json:"Type,omitempty" xml:"Type,omitempty"`
}

func (s GetDatasourceTableResponseBodyFields) String() string {
	return tea.Prettify(s)
}

func (s GetDatasourceTableResponseBodyFields) GoString() string {
	return s.String()
}

func (s *GetDatasourceTableResponseBodyFields) SetAttributes(v []*string) *GetDatasourceTableResponseBodyFields {
	s.Attributes = v
	return s
}

func (s *GetDatasourceTableResponseBodyFields) SetName(v string) *GetDatasourceTableResponseBodyFields {
	s.Name = &v
	return s
}

func (s *GetDatasourceTableResponseBodyFields) SetType(v string) *GetDatasourceTableResponseBodyFields {
	s.Type = &v
	return s
}

type GetDatasourceTableResponse struct {
	Headers    map[string]*string              `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
	StatusCode *int32                          `json:"statusCode,omitempty" xml:"statusCode,omitempty" require:"true"`
	Body       *GetDatasourceTableResponseBody `json:"body,omitempty" xml:"body,omitempty" require:"true"`
}

func (s GetDatasourceTableResponse) String() string {
	return tea.Prettify(s)
}

func (s GetDatasourceTableResponse) GoString() string {
	return s.String()
}

func (s *GetDatasourceTableResponse) SetHeaders(v map[string]*string) *GetDatasourceTableResponse {
	s.Headers = v
	return s
}

func (s *GetDatasourceTableResponse) SetStatusCode(v int32) *GetDatasourceTableResponse {
	s.StatusCode = &v
	return s
}

func (s *GetDatasourceTableResponse) SetBody(v *GetDatasourceTableResponseBody) *GetDatasourceTableResponse {
	s.Body = v
	return s
}

type GetFeatureEntityResponseBody struct {
	GmtCreateTime *string `json:"GmtCreateTime,omitempty" xml:"GmtCreateTime,omitempty"`
	JoinId        *string `json:"JoinId,omitempty" xml:"JoinId,omitempty"`
	Name          *string `json:"Name,omitempty" xml:"Name,omitempty"`
	Owner         *string `json:"Owner,omitempty" xml:"Owner,omitempty"`
	ProjectId     *string `json:"ProjectId,omitempty" xml:"ProjectId,omitempty"`
	ProjectName   *string `json:"ProjectName,omitempty" xml:"ProjectName,omitempty"`
	RequestId     *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
}

func (s GetFeatureEntityResponseBody) String() string {
	return tea.Prettify(s)
}

func (s GetFeatureEntityResponseBody) GoString() string {
	return s.String()
}

func (s *GetFeatureEntityResponseBody) SetGmtCreateTime(v string) *GetFeatureEntityResponseBody {
	s.GmtCreateTime = &v
	return s
}

func (s *GetFeatureEntityResponseBody) SetJoinId(v string) *GetFeatureEntityResponseBody {
	s.JoinId = &v
	return s
}

func (s *GetFeatureEntityResponseBody) SetName(v string) *GetFeatureEntityResponseBody {
	s.Name = &v
	return s
}

func (s *GetFeatureEntityResponseBody) SetOwner(v string) *GetFeatureEntityResponseBody {
	s.Owner = &v
	return s
}

func (s *GetFeatureEntityResponseBody) SetProjectId(v string) *GetFeatureEntityResponseBody {
	s.ProjectId = &v
	return s
}

func (s *GetFeatureEntityResponseBody) SetProjectName(v string) *GetFeatureEntityResponseBody {
	s.ProjectName = &v
	return s
}

func (s *GetFeatureEntityResponseBody) SetRequestId(v string) *GetFeatureEntityResponseBody {
	s.RequestId = &v
	return s
}

type GetFeatureEntityResponse struct {
	Headers    map[string]*string            `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
	StatusCode *int32                        `json:"statusCode,omitempty" xml:"statusCode,omitempty" require:"true"`
	Body       *GetFeatureEntityResponseBody `json:"body,omitempty" xml:"body,omitempty" require:"true"`
}

func (s GetFeatureEntityResponse) String() string {
	return tea.Prettify(s)
}

func (s GetFeatureEntityResponse) GoString() string {
	return s.String()
}

func (s *GetFeatureEntityResponse) SetHeaders(v map[string]*string) *GetFeatureEntityResponse {
	s.Headers = v
	return s
}

func (s *GetFeatureEntityResponse) SetStatusCode(v int32) *GetFeatureEntityResponse {
	s.StatusCode = &v
	return s
}

func (s *GetFeatureEntityResponse) SetBody(v *GetFeatureEntityResponseBody) *GetFeatureEntityResponse {
	s.Body = v
	return s
}

type GetFeatureViewResponseBody struct {
	Config                 *string                             `json:"Config,omitempty" xml:"Config,omitempty"`
	FeatureEntityId        *string                             `json:"FeatureEntityId,omitempty" xml:"FeatureEntityId,omitempty"`
	FeatureEntityName      *string                             `json:"FeatureEntityName,omitempty" xml:"FeatureEntityName,omitempty"`
	Fields                 []*GetFeatureViewResponseBodyFields `json:"Fields,omitempty" xml:"Fields,omitempty" type:"Repeated"`
	GmtCreateTime          *string                             `json:"GmtCreateTime,omitempty" xml:"GmtCreateTime,omitempty"`
	GmtModifiedTime        *string                             `json:"GmtModifiedTime,omitempty" xml:"GmtModifiedTime,omitempty"`
	GmtSyncTime            *string                             `json:"GmtSyncTime,omitempty" xml:"GmtSyncTime,omitempty"`
	JoinId                 *string                             `json:"JoinId,omitempty" xml:"JoinId,omitempty"`
	LastSyncConfig         *string                             `json:"LastSyncConfig,omitempty" xml:"LastSyncConfig,omitempty"`
	Name                   *string                             `json:"Name,omitempty" xml:"Name,omitempty"`
	Owner                  *string                             `json:"Owner,omitempty" xml:"Owner,omitempty"`
	ProjectId              *string                             `json:"ProjectId,omitempty" xml:"ProjectId,omitempty"`
	ProjectName            *string                             `json:"ProjectName,omitempty" xml:"ProjectName,omitempty"`
	RegisterDatasourceId   *string                             `json:"RegisterDatasourceId,omitempty" xml:"RegisterDatasourceId,omitempty"`
	RegisterDatasourceName *string                             `json:"RegisterDatasourceName,omitempty" xml:"RegisterDatasourceName,omitempty"`
	RegisterTable          *string                             `json:"RegisterTable,omitempty" xml:"RegisterTable,omitempty"`
	RequestId              *string                             `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
	SyncOnlineTable        *bool                               `json:"SyncOnlineTable,omitempty" xml:"SyncOnlineTable,omitempty"`
	TTL                    *int32                              `json:"TTL,omitempty" xml:"TTL,omitempty"`
	Tags                   []*string                           `json:"Tags,omitempty" xml:"Tags,omitempty" type:"Repeated"`
	Type                   *string                             `json:"Type,omitempty" xml:"Type,omitempty"`
	WriteMethod            *string                             `json:"WriteMethod,omitempty" xml:"WriteMethod,omitempty"`
}

func (s GetFeatureViewResponseBody) String() string {
	return tea.Prettify(s)
}

func (s GetFeatureViewResponseBody) GoString() string {
	return s.String()
}

func (s *GetFeatureViewResponseBody) SetConfig(v string) *GetFeatureViewResponseBody {
	s.Config = &v
	return s
}

func (s *GetFeatureViewResponseBody) SetFeatureEntityId(v string) *GetFeatureViewResponseBody {
	s.FeatureEntityId = &v
	return s
}

func (s *GetFeatureViewResponseBody) SetFeatureEntityName(v string) *GetFeatureViewResponseBody {
	s.FeatureEntityName = &v
	return s
}

func (s *GetFeatureViewResponseBody) SetFields(v []*GetFeatureViewResponseBodyFields) *GetFeatureViewResponseBody {
	s.Fields = v
	return s
}

func (s *GetFeatureViewResponseBody) SetGmtCreateTime(v string) *GetFeatureViewResponseBody {
	s.GmtCreateTime = &v
	return s
}

func (s *GetFeatureViewResponseBody) SetGmtModifiedTime(v string) *GetFeatureViewResponseBody {
	s.GmtModifiedTime = &v
	return s
}

func (s *GetFeatureViewResponseBody) SetGmtSyncTime(v string) *GetFeatureViewResponseBody {
	s.GmtSyncTime = &v
	return s
}

func (s *GetFeatureViewResponseBody) SetJoinId(v string) *GetFeatureViewResponseBody {
	s.JoinId = &v
	return s
}

func (s *GetFeatureViewResponseBody) SetLastSyncConfig(v string) *GetFeatureViewResponseBody {
	s.LastSyncConfig = &v
	return s
}

func (s *GetFeatureViewResponseBody) SetName(v string) *GetFeatureViewResponseBody {
	s.Name = &v
	return s
}

func (s *GetFeatureViewResponseBody) SetOwner(v string) *GetFeatureViewResponseBody {
	s.Owner = &v
	return s
}

func (s *GetFeatureViewResponseBody) SetProjectId(v string) *GetFeatureViewResponseBody {
	s.ProjectId = &v
	return s
}

func (s *GetFeatureViewResponseBody) SetProjectName(v string) *GetFeatureViewResponseBody {
	s.ProjectName = &v
	return s
}

func (s *GetFeatureViewResponseBody) SetRegisterDatasourceId(v string) *GetFeatureViewResponseBody {
	s.RegisterDatasourceId = &v
	return s
}

func (s *GetFeatureViewResponseBody) SetRegisterDatasourceName(v string) *GetFeatureViewResponseBody {
	s.RegisterDatasourceName = &v
	return s
}

func (s *GetFeatureViewResponseBody) SetRegisterTable(v string) *GetFeatureViewResponseBody {
	s.RegisterTable = &v
	return s
}

func (s *GetFeatureViewResponseBody) SetRequestId(v string) *GetFeatureViewResponseBody {
	s.RequestId = &v
	return s
}

func (s *GetFeatureViewResponseBody) SetSyncOnlineTable(v bool) *GetFeatureViewResponseBody {
	s.SyncOnlineTable = &v
	return s
}

func (s *GetFeatureViewResponseBody) SetTTL(v int32) *GetFeatureViewResponseBody {
	s.TTL = &v
	return s
}

func (s *GetFeatureViewResponseBody) SetTags(v []*string) *GetFeatureViewResponseBody {
	s.Tags = v
	return s
}

func (s *GetFeatureViewResponseBody) SetType(v string) *GetFeatureViewResponseBody {
	s.Type = &v
	return s
}

func (s *GetFeatureViewResponseBody) SetWriteMethod(v string) *GetFeatureViewResponseBody {
	s.WriteMethod = &v
	return s
}

type GetFeatureViewResponseBodyFields struct {
	Attributes []*string `json:"Attributes,omitempty" xml:"Attributes,omitempty" type:"Repeated"`
	Name       *string   `json:"Name,omitempty" xml:"Name,omitempty"`
	Type       *string   `json:"Type,omitempty" xml:"Type,omitempty"`
}

func (s GetFeatureViewResponseBodyFields) String() string {
	return tea.Prettify(s)
}

func (s GetFeatureViewResponseBodyFields) GoString() string {
	return s.String()
}

func (s *GetFeatureViewResponseBodyFields) SetAttributes(v []*string) *GetFeatureViewResponseBodyFields {
	s.Attributes = v
	return s
}

func (s *GetFeatureViewResponseBodyFields) SetName(v string) *GetFeatureViewResponseBodyFields {
	s.Name = &v
	return s
}

func (s *GetFeatureViewResponseBodyFields) SetType(v string) *GetFeatureViewResponseBodyFields {
	s.Type = &v
	return s
}

type GetFeatureViewResponse struct {
	Headers    map[string]*string          `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
	StatusCode *int32                      `json:"statusCode,omitempty" xml:"statusCode,omitempty" require:"true"`
	Body       *GetFeatureViewResponseBody `json:"body,omitempty" xml:"body,omitempty" require:"true"`
}

func (s GetFeatureViewResponse) String() string {
	return tea.Prettify(s)
}

func (s GetFeatureViewResponse) GoString() string {
	return s.String()
}

func (s *GetFeatureViewResponse) SetHeaders(v map[string]*string) *GetFeatureViewResponse {
	s.Headers = v
	return s
}

func (s *GetFeatureViewResponse) SetStatusCode(v int32) *GetFeatureViewResponse {
	s.StatusCode = &v
	return s
}

func (s *GetFeatureViewResponse) SetBody(v *GetFeatureViewResponseBody) *GetFeatureViewResponse {
	s.Body = v
	return s
}

type GetInstanceResponseBody struct {
	GmtCreateTime   *string  `json:"GmtCreateTime,omitempty" xml:"GmtCreateTime,omitempty"`
	GmtModifiedTime *string  `json:"GmtModifiedTime,omitempty" xml:"GmtModifiedTime,omitempty"`
	Message         *string  `json:"Message,omitempty" xml:"Message,omitempty"`
	Progress        *float64 `json:"Progress,omitempty" xml:"Progress,omitempty"`
	RegionId        *string  `json:"RegionId,omitempty" xml:"RegionId,omitempty"`
	RequestId       *string  `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
	Status          *string  `json:"Status,omitempty" xml:"Status,omitempty"`
	Type            *string  `json:"Type,omitempty" xml:"Type,omitempty"`
}

func (s GetInstanceResponseBody) String() string {
	return tea.Prettify(s)
}

func (s GetInstanceResponseBody) GoString() string {
	return s.String()
}

func (s *GetInstanceResponseBody) SetGmtCreateTime(v string) *GetInstanceResponseBody {
	s.GmtCreateTime = &v
	return s
}

func (s *GetInstanceResponseBody) SetGmtModifiedTime(v string) *GetInstanceResponseBody {
	s.GmtModifiedTime = &v
	return s
}

func (s *GetInstanceResponseBody) SetMessage(v string) *GetInstanceResponseBody {
	s.Message = &v
	return s
}

func (s *GetInstanceResponseBody) SetProgress(v float64) *GetInstanceResponseBody {
	s.Progress = &v
	return s
}

func (s *GetInstanceResponseBody) SetRegionId(v string) *GetInstanceResponseBody {
	s.RegionId = &v
	return s
}

func (s *GetInstanceResponseBody) SetRequestId(v string) *GetInstanceResponseBody {
	s.RequestId = &v
	return s
}

func (s *GetInstanceResponseBody) SetStatus(v string) *GetInstanceResponseBody {
	s.Status = &v
	return s
}

func (s *GetInstanceResponseBody) SetType(v string) *GetInstanceResponseBody {
	s.Type = &v
	return s
}

type GetInstanceResponse struct {
	Headers    map[string]*string       `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
	StatusCode *int32                   `json:"statusCode,omitempty" xml:"statusCode,omitempty" require:"true"`
	Body       *GetInstanceResponseBody `json:"body,omitempty" xml:"body,omitempty" require:"true"`
}

func (s GetInstanceResponse) String() string {
	return tea.Prettify(s)
}

func (s GetInstanceResponse) GoString() string {
	return s.String()
}

func (s *GetInstanceResponse) SetHeaders(v map[string]*string) *GetInstanceResponse {
	s.Headers = v
	return s
}

func (s *GetInstanceResponse) SetStatusCode(v int32) *GetInstanceResponse {
	s.StatusCode = &v
	return s
}

func (s *GetInstanceResponse) SetBody(v *GetInstanceResponseBody) *GetInstanceResponse {
	s.Body = v
	return s
}

type GetLabelTableResponseBody struct {
	DatasourceId    *string                            `json:"DatasourceId,omitempty" xml:"DatasourceId,omitempty"`
	DatasourceName  *string                            `json:"DatasourceName,omitempty" xml:"DatasourceName,omitempty"`
	Fields          []*GetLabelTableResponseBodyFields `json:"Fields,omitempty" xml:"Fields,omitempty" type:"Repeated"`
	GmtCreateTime   *string                            `json:"GmtCreateTime,omitempty" xml:"GmtCreateTime,omitempty"`
	GmtModifiedTime *string                            `json:"GmtModifiedTime,omitempty" xml:"GmtModifiedTime,omitempty"`
	Name            *string                            `json:"Name,omitempty" xml:"Name,omitempty"`
	Owner           *string                            `json:"Owner,omitempty" xml:"Owner,omitempty"`
	ProjectId       *string                            `json:"ProjectId,omitempty" xml:"ProjectId,omitempty"`
	ProjectName     *string                            `json:"ProjectName,omitempty" xml:"ProjectName,omitempty"`
	RequestId       *string                            `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
}

func (s GetLabelTableResponseBody) String() string {
	return tea.Prettify(s)
}

func (s GetLabelTableResponseBody) GoString() string {
	return s.String()
}

func (s *GetLabelTableResponseBody) SetDatasourceId(v string) *GetLabelTableResponseBody {
	s.DatasourceId = &v
	return s
}

func (s *GetLabelTableResponseBody) SetDatasourceName(v string) *GetLabelTableResponseBody {
	s.DatasourceName = &v
	return s
}

func (s *GetLabelTableResponseBody) SetFields(v []*GetLabelTableResponseBodyFields) *GetLabelTableResponseBody {
	s.Fields = v
	return s
}

func (s *GetLabelTableResponseBody) SetGmtCreateTime(v string) *GetLabelTableResponseBody {
	s.GmtCreateTime = &v
	return s
}

func (s *GetLabelTableResponseBody) SetGmtModifiedTime(v string) *GetLabelTableResponseBody {
	s.GmtModifiedTime = &v
	return s
}

func (s *GetLabelTableResponseBody) SetName(v string) *GetLabelTableResponseBody {
	s.Name = &v
	return s
}

func (s *GetLabelTableResponseBody) SetOwner(v string) *GetLabelTableResponseBody {
	s.Owner = &v
	return s
}

func (s *GetLabelTableResponseBody) SetProjectId(v string) *GetLabelTableResponseBody {
	s.ProjectId = &v
	return s
}

func (s *GetLabelTableResponseBody) SetProjectName(v string) *GetLabelTableResponseBody {
	s.ProjectName = &v
	return s
}

func (s *GetLabelTableResponseBody) SetRequestId(v string) *GetLabelTableResponseBody {
	s.RequestId = &v
	return s
}

type GetLabelTableResponseBodyFields struct {
	Attributes []*string `json:"Attributes,omitempty" xml:"Attributes,omitempty" type:"Repeated"`
	Name       *string   `json:"Name,omitempty" xml:"Name,omitempty"`
	Type       *string   `json:"Type,omitempty" xml:"Type,omitempty"`
}

func (s GetLabelTableResponseBodyFields) String() string {
	return tea.Prettify(s)
}

func (s GetLabelTableResponseBodyFields) GoString() string {
	return s.String()
}

func (s *GetLabelTableResponseBodyFields) SetAttributes(v []*string) *GetLabelTableResponseBodyFields {
	s.Attributes = v
	return s
}

func (s *GetLabelTableResponseBodyFields) SetName(v string) *GetLabelTableResponseBodyFields {
	s.Name = &v
	return s
}

func (s *GetLabelTableResponseBodyFields) SetType(v string) *GetLabelTableResponseBodyFields {
	s.Type = &v
	return s
}

type GetLabelTableResponse struct {
	Headers    map[string]*string         `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
	StatusCode *int32                     `json:"statusCode,omitempty" xml:"statusCode,omitempty" require:"true"`
	Body       *GetLabelTableResponseBody `json:"body,omitempty" xml:"body,omitempty" require:"true"`
}

func (s GetLabelTableResponse) String() string {
	return tea.Prettify(s)
}

func (s GetLabelTableResponse) GoString() string {
	return s.String()
}

func (s *GetLabelTableResponse) SetHeaders(v map[string]*string) *GetLabelTableResponse {
	s.Headers = v
	return s
}

func (s *GetLabelTableResponse) SetStatusCode(v int32) *GetLabelTableResponse {
	s.StatusCode = &v
	return s
}

func (s *GetLabelTableResponse) SetBody(v *GetLabelTableResponseBody) *GetLabelTableResponse {
	s.Body = v
	return s
}

type GetModelFeatureResponseBody struct {
	Features           []*GetModelFeatureResponseBodyFeatures `json:"Features,omitempty" xml:"Features,omitempty" type:"Repeated"`
	GmtCreateTime      *string                                `json:"GmtCreateTime,omitempty" xml:"GmtCreateTime,omitempty"`
	GmtModifiedTime    *string                                `json:"GmtModifiedTime,omitempty" xml:"GmtModifiedTime,omitempty"`
	LabelTableId       *string                                `json:"LabelTableId,omitempty" xml:"LabelTableId,omitempty"`
	LabelTableName     *string                                `json:"LabelTableName,omitempty" xml:"LabelTableName,omitempty"`
	Name               *string                                `json:"Name,omitempty" xml:"Name,omitempty"`
	Owner              *string                                `json:"Owner,omitempty" xml:"Owner,omitempty"`
	ProjectId          *string                                `json:"ProjectId,omitempty" xml:"ProjectId,omitempty"`
	ProjectName        *string                                `json:"ProjectName,omitempty" xml:"ProjectName,omitempty"`
	Relations          *GetModelFeatureResponseBodyRelations  `json:"Relations,omitempty" xml:"Relations,omitempty" type:"Struct"`
	RequestId          *string                                `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
	TrainingSetFGTable *string                                `json:"TrainingSetFGTable,omitempty" xml:"TrainingSetFGTable,omitempty"`
	TrainingSetTable   *string                                `json:"TrainingSetTable,omitempty" xml:"TrainingSetTable,omitempty"`
}

func (s GetModelFeatureResponseBody) String() string {
	return tea.Prettify(s)
}

func (s GetModelFeatureResponseBody) GoString() string {
	return s.String()
}

func (s *GetModelFeatureResponseBody) SetFeatures(v []*GetModelFeatureResponseBodyFeatures) *GetModelFeatureResponseBody {
	s.Features = v
	return s
}

func (s *GetModelFeatureResponseBody) SetGmtCreateTime(v string) *GetModelFeatureResponseBody {
	s.GmtCreateTime = &v
	return s
}

func (s *GetModelFeatureResponseBody) SetGmtModifiedTime(v string) *GetModelFeatureResponseBody {
	s.GmtModifiedTime = &v
	return s
}

func (s *GetModelFeatureResponseBody) SetLabelTableId(v string) *GetModelFeatureResponseBody {
	s.LabelTableId = &v
	return s
}

func (s *GetModelFeatureResponseBody) SetLabelTableName(v string) *GetModelFeatureResponseBody {
	s.LabelTableName = &v
	return s
}

func (s *GetModelFeatureResponseBody) SetName(v string) *GetModelFeatureResponseBody {
	s.Name = &v
	return s
}

func (s *GetModelFeatureResponseBody) SetOwner(v string) *GetModelFeatureResponseBody {
	s.Owner = &v
	return s
}

func (s *GetModelFeatureResponseBody) SetProjectId(v string) *GetModelFeatureResponseBody {
	s.ProjectId = &v
	return s
}

func (s *GetModelFeatureResponseBody) SetProjectName(v string) *GetModelFeatureResponseBody {
	s.ProjectName = &v
	return s
}

func (s *GetModelFeatureResponseBody) SetRelations(v *GetModelFeatureResponseBodyRelations) *GetModelFeatureResponseBody {
	s.Relations = v
	return s
}

func (s *GetModelFeatureResponseBody) SetRequestId(v string) *GetModelFeatureResponseBody {
	s.RequestId = &v
	return s
}

func (s *GetModelFeatureResponseBody) SetTrainingSetFGTable(v string) *GetModelFeatureResponseBody {
	s.TrainingSetFGTable = &v
	return s
}

func (s *GetModelFeatureResponseBody) SetTrainingSetTable(v string) *GetModelFeatureResponseBody {
	s.TrainingSetTable = &v
	return s
}

type GetModelFeatureResponseBodyFeatures struct {
	AliasName       *string `json:"AliasName,omitempty" xml:"AliasName,omitempty"`
	FeatureViewId   *string `json:"FeatureViewId,omitempty" xml:"FeatureViewId,omitempty"`
	FeatureViewName *string `json:"FeatureViewName,omitempty" xml:"FeatureViewName,omitempty"`
	Name            *string `json:"Name,omitempty" xml:"Name,omitempty"`
	Type            *string `json:"Type,omitempty" xml:"Type,omitempty"`
}

func (s GetModelFeatureResponseBodyFeatures) String() string {
	return tea.Prettify(s)
}

func (s GetModelFeatureResponseBodyFeatures) GoString() string {
	return s.String()
}

func (s *GetModelFeatureResponseBodyFeatures) SetAliasName(v string) *GetModelFeatureResponseBodyFeatures {
	s.AliasName = &v
	return s
}

func (s *GetModelFeatureResponseBodyFeatures) SetFeatureViewId(v string) *GetModelFeatureResponseBodyFeatures {
	s.FeatureViewId = &v
	return s
}

func (s *GetModelFeatureResponseBodyFeatures) SetFeatureViewName(v string) *GetModelFeatureResponseBodyFeatures {
	s.FeatureViewName = &v
	return s
}

func (s *GetModelFeatureResponseBodyFeatures) SetName(v string) *GetModelFeatureResponseBodyFeatures {
	s.Name = &v
	return s
}

func (s *GetModelFeatureResponseBodyFeatures) SetType(v string) *GetModelFeatureResponseBodyFeatures {
	s.Type = &v
	return s
}

type GetModelFeatureResponseBodyRelations struct {
	Domains []*GetModelFeatureResponseBodyRelationsDomains `json:"Domains,omitempty" xml:"Domains,omitempty" type:"Repeated"`
	Links   []*GetModelFeatureResponseBodyRelationsLinks   `json:"Links,omitempty" xml:"Links,omitempty" type:"Repeated"`
}

func (s GetModelFeatureResponseBodyRelations) String() string {
	return tea.Prettify(s)
}

func (s GetModelFeatureResponseBodyRelations) GoString() string {
	return s.String()
}

func (s *GetModelFeatureResponseBodyRelations) SetDomains(v []*GetModelFeatureResponseBodyRelationsDomains) *GetModelFeatureResponseBodyRelations {
	s.Domains = v
	return s
}

func (s *GetModelFeatureResponseBodyRelations) SetLinks(v []*GetModelFeatureResponseBodyRelationsLinks) *GetModelFeatureResponseBodyRelations {
	s.Links = v
	return s
}

type GetModelFeatureResponseBodyRelationsDomains struct {
	DomainType *string `json:"DomainType,omitempty" xml:"DomainType,omitempty"`
	// Domain ID。
	Id   *string `json:"Id,omitempty" xml:"Id,omitempty"`
	Name *string `json:"Name,omitempty" xml:"Name,omitempty"`
}

func (s GetModelFeatureResponseBodyRelationsDomains) String() string {
	return tea.Prettify(s)
}

func (s GetModelFeatureResponseBodyRelationsDomains) GoString() string {
	return s.String()
}

func (s *GetModelFeatureResponseBodyRelationsDomains) SetDomainType(v string) *GetModelFeatureResponseBodyRelationsDomains {
	s.DomainType = &v
	return s
}

func (s *GetModelFeatureResponseBodyRelationsDomains) SetId(v string) *GetModelFeatureResponseBodyRelationsDomains {
	s.Id = &v
	return s
}

func (s *GetModelFeatureResponseBodyRelationsDomains) SetName(v string) *GetModelFeatureResponseBodyRelationsDomains {
	s.Name = &v
	return s
}

type GetModelFeatureResponseBodyRelationsLinks struct {
	From *string `json:"From,omitempty" xml:"From,omitempty"`
	Link *string `json:"Link,omitempty" xml:"Link,omitempty"`
	To   *string `json:"To,omitempty" xml:"To,omitempty"`
}

func (s GetModelFeatureResponseBodyRelationsLinks) String() string {
	return tea.Prettify(s)
}

func (s GetModelFeatureResponseBodyRelationsLinks) GoString() string {
	return s.String()
}

func (s *GetModelFeatureResponseBodyRelationsLinks) SetFrom(v string) *GetModelFeatureResponseBodyRelationsLinks {
	s.From = &v
	return s
}

func (s *GetModelFeatureResponseBodyRelationsLinks) SetLink(v string) *GetModelFeatureResponseBodyRelationsLinks {
	s.Link = &v
	return s
}

func (s *GetModelFeatureResponseBodyRelationsLinks) SetTo(v string) *GetModelFeatureResponseBodyRelationsLinks {
	s.To = &v
	return s
}

type GetModelFeatureResponse struct {
	Headers    map[string]*string           `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
	StatusCode *int32                       `json:"statusCode,omitempty" xml:"statusCode,omitempty" require:"true"`
	Body       *GetModelFeatureResponseBody `json:"body,omitempty" xml:"body,omitempty" require:"true"`
}

func (s GetModelFeatureResponse) String() string {
	return tea.Prettify(s)
}

func (s GetModelFeatureResponse) GoString() string {
	return s.String()
}

func (s *GetModelFeatureResponse) SetHeaders(v map[string]*string) *GetModelFeatureResponse {
	s.Headers = v
	return s
}

func (s *GetModelFeatureResponse) SetStatusCode(v int32) *GetModelFeatureResponse {
	s.StatusCode = &v
	return s
}

func (s *GetModelFeatureResponse) SetBody(v *GetModelFeatureResponseBody) *GetModelFeatureResponse {
	s.Body = v
	return s
}

type GetProjectResponseBody struct {
	Description           *string `json:"Description,omitempty" xml:"Description,omitempty"`
	FeatureEntityCount    *int32  `json:"FeatureEntityCount,omitempty" xml:"FeatureEntityCount,omitempty"`
	FeatureViewCount      *int32  `json:"FeatureViewCount,omitempty" xml:"FeatureViewCount,omitempty"`
	GmtCreateTime         *string `json:"GmtCreateTime,omitempty" xml:"GmtCreateTime,omitempty"`
	GmtModifiedTime       *string `json:"GmtModifiedTime,omitempty" xml:"GmtModifiedTime,omitempty"`
	ModelCount            *int32  `json:"ModelCount,omitempty" xml:"ModelCount,omitempty"`
	Name                  *string `json:"Name,omitempty" xml:"Name,omitempty"`
	OfflineDatasourceId   *string `json:"OfflineDatasourceId,omitempty" xml:"OfflineDatasourceId,omitempty"`
	OfflineDatasourceName *string `json:"OfflineDatasourceName,omitempty" xml:"OfflineDatasourceName,omitempty"`
	OfflineDatasourceType *string `json:"OfflineDatasourceType,omitempty" xml:"OfflineDatasourceType,omitempty"`
	OfflineLifecycle      *int32  `json:"OfflineLifecycle,omitempty" xml:"OfflineLifecycle,omitempty"`
	OnlineDatasourceId    *string `json:"OnlineDatasourceId,omitempty" xml:"OnlineDatasourceId,omitempty"`
	OnlineDatasourceName  *string `json:"OnlineDatasourceName,omitempty" xml:"OnlineDatasourceName,omitempty"`
	OnlineDatasourceType  *string `json:"OnlineDatasourceType,omitempty" xml:"OnlineDatasourceType,omitempty"`
	Owner                 *string `json:"Owner,omitempty" xml:"Owner,omitempty"`
	RequestId             *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
}

func (s GetProjectResponseBody) String() string {
	return tea.Prettify(s)
}

func (s GetProjectResponseBody) GoString() string {
	return s.String()
}

func (s *GetProjectResponseBody) SetDescription(v string) *GetProjectResponseBody {
	s.Description = &v
	return s
}

func (s *GetProjectResponseBody) SetFeatureEntityCount(v int32) *GetProjectResponseBody {
	s.FeatureEntityCount = &v
	return s
}

func (s *GetProjectResponseBody) SetFeatureViewCount(v int32) *GetProjectResponseBody {
	s.FeatureViewCount = &v
	return s
}

func (s *GetProjectResponseBody) SetGmtCreateTime(v string) *GetProjectResponseBody {
	s.GmtCreateTime = &v
	return s
}

func (s *GetProjectResponseBody) SetGmtModifiedTime(v string) *GetProjectResponseBody {
	s.GmtModifiedTime = &v
	return s
}

func (s *GetProjectResponseBody) SetModelCount(v int32) *GetProjectResponseBody {
	s.ModelCount = &v
	return s
}

func (s *GetProjectResponseBody) SetName(v string) *GetProjectResponseBody {
	s.Name = &v
	return s
}

func (s *GetProjectResponseBody) SetOfflineDatasourceId(v string) *GetProjectResponseBody {
	s.OfflineDatasourceId = &v
	return s
}

func (s *GetProjectResponseBody) SetOfflineDatasourceName(v string) *GetProjectResponseBody {
	s.OfflineDatasourceName = &v
	return s
}

func (s *GetProjectResponseBody) SetOfflineDatasourceType(v string) *GetProjectResponseBody {
	s.OfflineDatasourceType = &v
	return s
}

func (s *GetProjectResponseBody) SetOfflineLifecycle(v int32) *GetProjectResponseBody {
	s.OfflineLifecycle = &v
	return s
}

func (s *GetProjectResponseBody) SetOnlineDatasourceId(v string) *GetProjectResponseBody {
	s.OnlineDatasourceId = &v
	return s
}

func (s *GetProjectResponseBody) SetOnlineDatasourceName(v string) *GetProjectResponseBody {
	s.OnlineDatasourceName = &v
	return s
}

func (s *GetProjectResponseBody) SetOnlineDatasourceType(v string) *GetProjectResponseBody {
	s.OnlineDatasourceType = &v
	return s
}

func (s *GetProjectResponseBody) SetOwner(v string) *GetProjectResponseBody {
	s.Owner = &v
	return s
}

func (s *GetProjectResponseBody) SetRequestId(v string) *GetProjectResponseBody {
	s.RequestId = &v
	return s
}

type GetProjectResponse struct {
	Headers    map[string]*string      `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
	StatusCode *int32                  `json:"statusCode,omitempty" xml:"statusCode,omitempty" require:"true"`
	Body       *GetProjectResponseBody `json:"body,omitempty" xml:"body,omitempty" require:"true"`
}

func (s GetProjectResponse) String() string {
	return tea.Prettify(s)
}

func (s GetProjectResponse) GoString() string {
	return s.String()
}

func (s *GetProjectResponse) SetHeaders(v map[string]*string) *GetProjectResponse {
	s.Headers = v
	return s
}

func (s *GetProjectResponse) SetStatusCode(v int32) *GetProjectResponse {
	s.StatusCode = &v
	return s
}

func (s *GetProjectResponse) SetBody(v *GetProjectResponseBody) *GetProjectResponse {
	s.Body = v
	return s
}

type GetProjectFeatureEntityResponseBody struct {
	FeatureEntityId *string `json:"FeatureEntityId,omitempty" xml:"FeatureEntityId,omitempty"`
	JoinId          *string `json:"JoinId,omitempty" xml:"JoinId,omitempty"`
	Name            *string `json:"Name,omitempty" xml:"Name,omitempty"`
	ProjectName     *string `json:"ProjectName,omitempty" xml:"ProjectName,omitempty"`
	RequestId       *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
	WorkspaceId     *string `json:"WorkspaceId,omitempty" xml:"WorkspaceId,omitempty"`
}

func (s GetProjectFeatureEntityResponseBody) String() string {
	return tea.Prettify(s)
}

func (s GetProjectFeatureEntityResponseBody) GoString() string {
	return s.String()
}

func (s *GetProjectFeatureEntityResponseBody) SetFeatureEntityId(v string) *GetProjectFeatureEntityResponseBody {
	s.FeatureEntityId = &v
	return s
}

func (s *GetProjectFeatureEntityResponseBody) SetJoinId(v string) *GetProjectFeatureEntityResponseBody {
	s.JoinId = &v
	return s
}

func (s *GetProjectFeatureEntityResponseBody) SetName(v string) *GetProjectFeatureEntityResponseBody {
	s.Name = &v
	return s
}

func (s *GetProjectFeatureEntityResponseBody) SetProjectName(v string) *GetProjectFeatureEntityResponseBody {
	s.ProjectName = &v
	return s
}

func (s *GetProjectFeatureEntityResponseBody) SetRequestId(v string) *GetProjectFeatureEntityResponseBody {
	s.RequestId = &v
	return s
}

func (s *GetProjectFeatureEntityResponseBody) SetWorkspaceId(v string) *GetProjectFeatureEntityResponseBody {
	s.WorkspaceId = &v
	return s
}

type GetProjectFeatureEntityResponse struct {
	Headers    map[string]*string                   `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
	StatusCode *int32                               `json:"statusCode,omitempty" xml:"statusCode,omitempty" require:"true"`
	Body       *GetProjectFeatureEntityResponseBody `json:"body,omitempty" xml:"body,omitempty" require:"true"`
}

func (s GetProjectFeatureEntityResponse) String() string {
	return tea.Prettify(s)
}

func (s GetProjectFeatureEntityResponse) GoString() string {
	return s.String()
}

func (s *GetProjectFeatureEntityResponse) SetHeaders(v map[string]*string) *GetProjectFeatureEntityResponse {
	s.Headers = v
	return s
}

func (s *GetProjectFeatureEntityResponse) SetStatusCode(v int32) *GetProjectFeatureEntityResponse {
	s.StatusCode = &v
	return s
}

func (s *GetProjectFeatureEntityResponse) SetBody(v *GetProjectFeatureEntityResponseBody) *GetProjectFeatureEntityResponse {
	s.Body = v
	return s
}

type GetProjectFeatureEntityHotIdsResponseBody struct {
	Count         *int32  `json:"Count,omitempty" xml:"Count,omitempty"`
	HotIds        *string `json:"HotIds,omitempty" xml:"HotIds,omitempty"`
	NextSeqNumber *string `json:"NextSeqNumber,omitempty" xml:"NextSeqNumber,omitempty"`
	RequestId     *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
}

func (s GetProjectFeatureEntityHotIdsResponseBody) String() string {
	return tea.Prettify(s)
}

func (s GetProjectFeatureEntityHotIdsResponseBody) GoString() string {
	return s.String()
}

func (s *GetProjectFeatureEntityHotIdsResponseBody) SetCount(v int32) *GetProjectFeatureEntityHotIdsResponseBody {
	s.Count = &v
	return s
}

func (s *GetProjectFeatureEntityHotIdsResponseBody) SetHotIds(v string) *GetProjectFeatureEntityHotIdsResponseBody {
	s.HotIds = &v
	return s
}

func (s *GetProjectFeatureEntityHotIdsResponseBody) SetNextSeqNumber(v string) *GetProjectFeatureEntityHotIdsResponseBody {
	s.NextSeqNumber = &v
	return s
}

func (s *GetProjectFeatureEntityHotIdsResponseBody) SetRequestId(v string) *GetProjectFeatureEntityHotIdsResponseBody {
	s.RequestId = &v
	return s
}

type GetProjectFeatureEntityHotIdsResponse struct {
	Headers    map[string]*string                         `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
	StatusCode *int32                                     `json:"statusCode,omitempty" xml:"statusCode,omitempty" require:"true"`
	Body       *GetProjectFeatureEntityHotIdsResponseBody `json:"body,omitempty" xml:"body,omitempty" require:"true"`
}

func (s GetProjectFeatureEntityHotIdsResponse) String() string {
	return tea.Prettify(s)
}

func (s GetProjectFeatureEntityHotIdsResponse) GoString() string {
	return s.String()
}

func (s *GetProjectFeatureEntityHotIdsResponse) SetHeaders(v map[string]*string) *GetProjectFeatureEntityHotIdsResponse {
	s.Headers = v
	return s
}

func (s *GetProjectFeatureEntityHotIdsResponse) SetStatusCode(v int32) *GetProjectFeatureEntityHotIdsResponse {
	s.StatusCode = &v
	return s
}

func (s *GetProjectFeatureEntityHotIdsResponse) SetBody(v *GetProjectFeatureEntityHotIdsResponseBody) *GetProjectFeatureEntityHotIdsResponse {
	s.Body = v
	return s
}

type GetProjectFeatureViewResponseBody struct {
	Config               *string                                    `json:"Config,omitempty" xml:"Config,omitempty"`
	FeatureEntityId      *string                                    `json:"FeatureEntityId,omitempty" xml:"FeatureEntityId,omitempty"`
	FeatureEntityName    *string                                    `json:"FeatureEntityName,omitempty" xml:"FeatureEntityName,omitempty"`
	FeatureViewId        *string                                    `json:"FeatureViewId,omitempty" xml:"FeatureViewId,omitempty"`
	Fields               []*GetProjectFeatureViewResponseBodyFields `json:"Fields,omitempty" xml:"Fields,omitempty" type:"Repeated"`
	GmtSyncTime          *string                                    `json:"GmtSyncTime,omitempty" xml:"GmtSyncTime,omitempty"`
	JoinId               *string                                    `json:"JoinId,omitempty" xml:"JoinId,omitempty"`
	LastSyncConfig       *string                                    `json:"LastSyncConfig,omitempty" xml:"LastSyncConfig,omitempty"`
	Name                 *string                                    `json:"Name,omitempty" xml:"Name,omitempty"`
	Owner                *string                                    `json:"Owner,omitempty" xml:"Owner,omitempty"`
	ProjectId            *string                                    `json:"ProjectId,omitempty" xml:"ProjectId,omitempty"`
	ProjectName          *string                                    `json:"ProjectName,omitempty" xml:"ProjectName,omitempty"`
	RegisterDatasourceId *string                                    `json:"RegisterDatasourceId,omitempty" xml:"RegisterDatasourceId,omitempty"`
	RegisterTable        *string                                    `json:"RegisterTable,omitempty" xml:"RegisterTable,omitempty"`
	RequestId            *string                                    `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
	SyncOnlineTable      *bool                                      `json:"SyncOnlineTable,omitempty" xml:"SyncOnlineTable,omitempty"`
	TTL                  *int32                                     `json:"TTL,omitempty" xml:"TTL,omitempty"`
	Tags                 []*string                                  `json:"Tags,omitempty" xml:"Tags,omitempty" type:"Repeated"`
	Type                 *string                                    `json:"Type,omitempty" xml:"Type,omitempty"`
	WriteMethod          *string                                    `json:"WriteMethod,omitempty" xml:"WriteMethod,omitempty"`
}

func (s GetProjectFeatureViewResponseBody) String() string {
	return tea.Prettify(s)
}

func (s GetProjectFeatureViewResponseBody) GoString() string {
	return s.String()
}

func (s *GetProjectFeatureViewResponseBody) SetConfig(v string) *GetProjectFeatureViewResponseBody {
	s.Config = &v
	return s
}

func (s *GetProjectFeatureViewResponseBody) SetFeatureEntityId(v string) *GetProjectFeatureViewResponseBody {
	s.FeatureEntityId = &v
	return s
}

func (s *GetProjectFeatureViewResponseBody) SetFeatureEntityName(v string) *GetProjectFeatureViewResponseBody {
	s.FeatureEntityName = &v
	return s
}

func (s *GetProjectFeatureViewResponseBody) SetFeatureViewId(v string) *GetProjectFeatureViewResponseBody {
	s.FeatureViewId = &v
	return s
}

func (s *GetProjectFeatureViewResponseBody) SetFields(v []*GetProjectFeatureViewResponseBodyFields) *GetProjectFeatureViewResponseBody {
	s.Fields = v
	return s
}

func (s *GetProjectFeatureViewResponseBody) SetGmtSyncTime(v string) *GetProjectFeatureViewResponseBody {
	s.GmtSyncTime = &v
	return s
}

func (s *GetProjectFeatureViewResponseBody) SetJoinId(v string) *GetProjectFeatureViewResponseBody {
	s.JoinId = &v
	return s
}

func (s *GetProjectFeatureViewResponseBody) SetLastSyncConfig(v string) *GetProjectFeatureViewResponseBody {
	s.LastSyncConfig = &v
	return s
}

func (s *GetProjectFeatureViewResponseBody) SetName(v string) *GetProjectFeatureViewResponseBody {
	s.Name = &v
	return s
}

func (s *GetProjectFeatureViewResponseBody) SetOwner(v string) *GetProjectFeatureViewResponseBody {
	s.Owner = &v
	return s
}

func (s *GetProjectFeatureViewResponseBody) SetProjectId(v string) *GetProjectFeatureViewResponseBody {
	s.ProjectId = &v
	return s
}

func (s *GetProjectFeatureViewResponseBody) SetProjectName(v string) *GetProjectFeatureViewResponseBody {
	s.ProjectName = &v
	return s
}

func (s *GetProjectFeatureViewResponseBody) SetRegisterDatasourceId(v string) *GetProjectFeatureViewResponseBody {
	s.RegisterDatasourceId = &v
	return s
}

func (s *GetProjectFeatureViewResponseBody) SetRegisterTable(v string) *GetProjectFeatureViewResponseBody {
	s.RegisterTable = &v
	return s
}

func (s *GetProjectFeatureViewResponseBody) SetRequestId(v string) *GetProjectFeatureViewResponseBody {
	s.RequestId = &v
	return s
}

func (s *GetProjectFeatureViewResponseBody) SetSyncOnlineTable(v bool) *GetProjectFeatureViewResponseBody {
	s.SyncOnlineTable = &v
	return s
}

func (s *GetProjectFeatureViewResponseBody) SetTTL(v int32) *GetProjectFeatureViewResponseBody {
	s.TTL = &v
	return s
}

func (s *GetProjectFeatureViewResponseBody) SetTags(v []*string) *GetProjectFeatureViewResponseBody {
	s.Tags = v
	return s
}

func (s *GetProjectFeatureViewResponseBody) SetType(v string) *GetProjectFeatureViewResponseBody {
	s.Type = &v
	return s
}

func (s *GetProjectFeatureViewResponseBody) SetWriteMethod(v string) *GetProjectFeatureViewResponseBody {
	s.WriteMethod = &v
	return s
}

type GetProjectFeatureViewResponseBodyFields struct {
	Attributes []*string `json:"Attributes,omitempty" xml:"Attributes,omitempty" type:"Repeated"`
	Name       *string   `json:"Name,omitempty" xml:"Name,omitempty"`
	Type       *string   `json:"Type,omitempty" xml:"Type,omitempty"`
}

func (s GetProjectFeatureViewResponseBodyFields) String() string {
	return tea.Prettify(s)
}

func (s GetProjectFeatureViewResponseBodyFields) GoString() string {
	return s.String()
}

func (s *GetProjectFeatureViewResponseBodyFields) SetAttributes(v []*string) *GetProjectFeatureViewResponseBodyFields {
	s.Attributes = v
	return s
}

func (s *GetProjectFeatureViewResponseBodyFields) SetName(v string) *GetProjectFeatureViewResponseBodyFields {
	s.Name = &v
	return s
}

func (s *GetProjectFeatureViewResponseBodyFields) SetType(v string) *GetProjectFeatureViewResponseBodyFields {
	s.Type = &v
	return s
}

type GetProjectFeatureViewResponse struct {
	Headers    map[string]*string                 `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
	StatusCode *int32                             `json:"statusCode,omitempty" xml:"statusCode,omitempty" require:"true"`
	Body       *GetProjectFeatureViewResponseBody `json:"body,omitempty" xml:"body,omitempty" require:"true"`
}

func (s GetProjectFeatureViewResponse) String() string {
	return tea.Prettify(s)
}

func (s GetProjectFeatureViewResponse) GoString() string {
	return s.String()
}

func (s *GetProjectFeatureViewResponse) SetHeaders(v map[string]*string) *GetProjectFeatureViewResponse {
	s.Headers = v
	return s
}

func (s *GetProjectFeatureViewResponse) SetStatusCode(v int32) *GetProjectFeatureViewResponse {
	s.StatusCode = &v
	return s
}

func (s *GetProjectFeatureViewResponse) SetBody(v *GetProjectFeatureViewResponseBody) *GetProjectFeatureViewResponse {
	s.Body = v
	return s
}

type GetProjectModelFeatureResponseBody struct {
	Features             []*GetProjectModelFeatureResponseBodyFeatures `json:"Features,omitempty" xml:"Features,omitempty" type:"Repeated"`
	GmtCreateTime        *string                                       `json:"GmtCreateTime,omitempty" xml:"GmtCreateTime,omitempty"`
	GmtModifiedTime      *string                                       `json:"GmtModifiedTime,omitempty" xml:"GmtModifiedTime,omitempty"`
	LabelDatasourceId    *string                                       `json:"LabelDatasourceId,omitempty" xml:"LabelDatasourceId,omitempty"`
	LabelDatasourceTable *string                                       `json:"LabelDatasourceTable,omitempty" xml:"LabelDatasourceTable,omitempty"`
	LabelEventTime       *string                                       `json:"LabelEventTime,omitempty" xml:"LabelEventTime,omitempty"`
	LabelTableId         *string                                       `json:"LabelTableId,omitempty" xml:"LabelTableId,omitempty"`
	ModelFeatureId       *string                                       `json:"ModelFeatureId,omitempty" xml:"ModelFeatureId,omitempty"`
	Name                 *string                                       `json:"Name,omitempty" xml:"Name,omitempty"`
	Owner                *string                                       `json:"Owner,omitempty" xml:"Owner,omitempty"`
	ProjectId            *string                                       `json:"ProjectId,omitempty" xml:"ProjectId,omitempty"`
	ProjectName          *string                                       `json:"ProjectName,omitempty" xml:"ProjectName,omitempty"`
	RequestId            *string                                       `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
	TrainingSetFGTable   *string                                       `json:"TrainingSetFGTable,omitempty" xml:"TrainingSetFGTable,omitempty"`
	TrainingSetTable     *string                                       `json:"TrainingSetTable,omitempty" xml:"TrainingSetTable,omitempty"`
}

func (s GetProjectModelFeatureResponseBody) String() string {
	return tea.Prettify(s)
}

func (s GetProjectModelFeatureResponseBody) GoString() string {
	return s.String()
}

func (s *GetProjectModelFeatureResponseBody) SetFeatures(v []*GetProjectModelFeatureResponseBodyFeatures) *GetProjectModelFeatureResponseBody {
	s.Features = v
	return s
}

func (s *GetProjectModelFeatureResponseBody) SetGmtCreateTime(v string) *GetProjectModelFeatureResponseBody {
	s.GmtCreateTime = &v
	return s
}

func (s *GetProjectModelFeatureResponseBody) SetGmtModifiedTime(v string) *GetProjectModelFeatureResponseBody {
	s.GmtModifiedTime = &v
	return s
}

func (s *GetProjectModelFeatureResponseBody) SetLabelDatasourceId(v string) *GetProjectModelFeatureResponseBody {
	s.LabelDatasourceId = &v
	return s
}

func (s *GetProjectModelFeatureResponseBody) SetLabelDatasourceTable(v string) *GetProjectModelFeatureResponseBody {
	s.LabelDatasourceTable = &v
	return s
}

func (s *GetProjectModelFeatureResponseBody) SetLabelEventTime(v string) *GetProjectModelFeatureResponseBody {
	s.LabelEventTime = &v
	return s
}

func (s *GetProjectModelFeatureResponseBody) SetLabelTableId(v string) *GetProjectModelFeatureResponseBody {
	s.LabelTableId = &v
	return s
}

func (s *GetProjectModelFeatureResponseBody) SetModelFeatureId(v string) *GetProjectModelFeatureResponseBody {
	s.ModelFeatureId = &v
	return s
}

func (s *GetProjectModelFeatureResponseBody) SetName(v string) *GetProjectModelFeatureResponseBody {
	s.Name = &v
	return s
}

func (s *GetProjectModelFeatureResponseBody) SetOwner(v string) *GetProjectModelFeatureResponseBody {
	s.Owner = &v
	return s
}

func (s *GetProjectModelFeatureResponseBody) SetProjectId(v string) *GetProjectModelFeatureResponseBody {
	s.ProjectId = &v
	return s
}

func (s *GetProjectModelFeatureResponseBody) SetProjectName(v string) *GetProjectModelFeatureResponseBody {
	s.ProjectName = &v
	return s
}

func (s *GetProjectModelFeatureResponseBody) SetRequestId(v string) *GetProjectModelFeatureResponseBody {
	s.RequestId = &v
	return s
}

func (s *GetProjectModelFeatureResponseBody) SetTrainingSetFGTable(v string) *GetProjectModelFeatureResponseBody {
	s.TrainingSetFGTable = &v
	return s
}

func (s *GetProjectModelFeatureResponseBody) SetTrainingSetTable(v string) *GetProjectModelFeatureResponseBody {
	s.TrainingSetTable = &v
	return s
}

type GetProjectModelFeatureResponseBodyFeatures struct {
	AliasName       *string `json:"AliasName,omitempty" xml:"AliasName,omitempty"`
	FeatureViewId   *string `json:"FeatureViewId,omitempty" xml:"FeatureViewId,omitempty"`
	FeatureViewName *string `json:"FeatureViewName,omitempty" xml:"FeatureViewName,omitempty"`
	Name            *string `json:"Name,omitempty" xml:"Name,omitempty"`
	Type            *string `json:"Type,omitempty" xml:"Type,omitempty"`
}

func (s GetProjectModelFeatureResponseBodyFeatures) String() string {
	return tea.Prettify(s)
}

func (s GetProjectModelFeatureResponseBodyFeatures) GoString() string {
	return s.String()
}

func (s *GetProjectModelFeatureResponseBodyFeatures) SetAliasName(v string) *GetProjectModelFeatureResponseBodyFeatures {
	s.AliasName = &v
	return s
}

func (s *GetProjectModelFeatureResponseBodyFeatures) SetFeatureViewId(v string) *GetProjectModelFeatureResponseBodyFeatures {
	s.FeatureViewId = &v
	return s
}

func (s *GetProjectModelFeatureResponseBodyFeatures) SetFeatureViewName(v string) *GetProjectModelFeatureResponseBodyFeatures {
	s.FeatureViewName = &v
	return s
}

func (s *GetProjectModelFeatureResponseBodyFeatures) SetName(v string) *GetProjectModelFeatureResponseBodyFeatures {
	s.Name = &v
	return s
}

func (s *GetProjectModelFeatureResponseBodyFeatures) SetType(v string) *GetProjectModelFeatureResponseBodyFeatures {
	s.Type = &v
	return s
}

type GetProjectModelFeatureResponse struct {
	Headers    map[string]*string                  `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
	StatusCode *int32                              `json:"statusCode,omitempty" xml:"statusCode,omitempty" require:"true"`
	Body       *GetProjectModelFeatureResponseBody `json:"body,omitempty" xml:"body,omitempty" require:"true"`
}

func (s GetProjectModelFeatureResponse) String() string {
	return tea.Prettify(s)
}

func (s GetProjectModelFeatureResponse) GoString() string {
	return s.String()
}

func (s *GetProjectModelFeatureResponse) SetHeaders(v map[string]*string) *GetProjectModelFeatureResponse {
	s.Headers = v
	return s
}

func (s *GetProjectModelFeatureResponse) SetStatusCode(v int32) *GetProjectModelFeatureResponse {
	s.StatusCode = &v
	return s
}

func (s *GetProjectModelFeatureResponse) SetBody(v *GetProjectModelFeatureResponseBody) *GetProjectModelFeatureResponse {
	s.Body = v
	return s
}

type GetServiceIdentityRoleResponseBody struct {
	Policy    *string `json:"Policy,omitempty" xml:"Policy,omitempty"`
	RequestId *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
	RoleName  *string `json:"RoleName,omitempty" xml:"RoleName,omitempty"`
}

func (s GetServiceIdentityRoleResponseBody) String() string {
	return tea.Prettify(s)
}

func (s GetServiceIdentityRoleResponseBody) GoString() string {
	return s.String()
}

func (s *GetServiceIdentityRoleResponseBody) SetPolicy(v string) *GetServiceIdentityRoleResponseBody {
	s.Policy = &v
	return s
}

func (s *GetServiceIdentityRoleResponseBody) SetRequestId(v string) *GetServiceIdentityRoleResponseBody {
	s.RequestId = &v
	return s
}

func (s *GetServiceIdentityRoleResponseBody) SetRoleName(v string) *GetServiceIdentityRoleResponseBody {
	s.RoleName = &v
	return s
}

type GetServiceIdentityRoleResponse struct {
	Headers    map[string]*string                  `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
	StatusCode *int32                              `json:"statusCode,omitempty" xml:"statusCode,omitempty" require:"true"`
	Body       *GetServiceIdentityRoleResponseBody `json:"body,omitempty" xml:"body,omitempty" require:"true"`
}

func (s GetServiceIdentityRoleResponse) String() string {
	return tea.Prettify(s)
}

func (s GetServiceIdentityRoleResponse) GoString() string {
	return s.String()
}

func (s *GetServiceIdentityRoleResponse) SetHeaders(v map[string]*string) *GetServiceIdentityRoleResponse {
	s.Headers = v
	return s
}

func (s *GetServiceIdentityRoleResponse) SetStatusCode(v int32) *GetServiceIdentityRoleResponse {
	s.StatusCode = &v
	return s
}

func (s *GetServiceIdentityRoleResponse) SetBody(v *GetServiceIdentityRoleResponseBody) *GetServiceIdentityRoleResponse {
	s.Body = v
	return s
}

type GetTaskResponseBody struct {
	Config          *string `json:"Config,omitempty" xml:"Config,omitempty"`
	GmtCreateTime   *string `json:"GmtCreateTime,omitempty" xml:"GmtCreateTime,omitempty"`
	GmtExecutedTime *string `json:"GmtExecutedTime,omitempty" xml:"GmtExecutedTime,omitempty"`
	GmtFinishedTime *string `json:"GmtFinishedTime,omitempty" xml:"GmtFinishedTime,omitempty"`
	GmtModifiedTime *string `json:"GmtModifiedTime,omitempty" xml:"GmtModifiedTime,omitempty"`
	ObjectId        *string `json:"ObjectId,omitempty" xml:"ObjectId,omitempty"`
	ObjectType      *string `json:"ObjectType,omitempty" xml:"ObjectType,omitempty"`
	ProjectId       *string `json:"ProjectId,omitempty" xml:"ProjectId,omitempty"`
	ProjectName     *string `json:"ProjectName,omitempty" xml:"ProjectName,omitempty"`
	RequestId       *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
	RunningConfig   *string `json:"RunningConfig,omitempty" xml:"RunningConfig,omitempty"`
	Status          *string `json:"Status,omitempty" xml:"Status,omitempty"`
	Type            *string `json:"Type,omitempty" xml:"Type,omitempty"`
}

func (s GetTaskResponseBody) String() string {
	return tea.Prettify(s)
}

func (s GetTaskResponseBody) GoString() string {
	return s.String()
}

func (s *GetTaskResponseBody) SetConfig(v string) *GetTaskResponseBody {
	s.Config = &v
	return s
}

func (s *GetTaskResponseBody) SetGmtCreateTime(v string) *GetTaskResponseBody {
	s.GmtCreateTime = &v
	return s
}

func (s *GetTaskResponseBody) SetGmtExecutedTime(v string) *GetTaskResponseBody {
	s.GmtExecutedTime = &v
	return s
}

func (s *GetTaskResponseBody) SetGmtFinishedTime(v string) *GetTaskResponseBody {
	s.GmtFinishedTime = &v
	return s
}

func (s *GetTaskResponseBody) SetGmtModifiedTime(v string) *GetTaskResponseBody {
	s.GmtModifiedTime = &v
	return s
}

func (s *GetTaskResponseBody) SetObjectId(v string) *GetTaskResponseBody {
	s.ObjectId = &v
	return s
}

func (s *GetTaskResponseBody) SetObjectType(v string) *GetTaskResponseBody {
	s.ObjectType = &v
	return s
}

func (s *GetTaskResponseBody) SetProjectId(v string) *GetTaskResponseBody {
	s.ProjectId = &v
	return s
}

func (s *GetTaskResponseBody) SetProjectName(v string) *GetTaskResponseBody {
	s.ProjectName = &v
	return s
}

func (s *GetTaskResponseBody) SetRequestId(v string) *GetTaskResponseBody {
	s.RequestId = &v
	return s
}

func (s *GetTaskResponseBody) SetRunningConfig(v string) *GetTaskResponseBody {
	s.RunningConfig = &v
	return s
}

func (s *GetTaskResponseBody) SetStatus(v string) *GetTaskResponseBody {
	s.Status = &v
	return s
}

func (s *GetTaskResponseBody) SetType(v string) *GetTaskResponseBody {
	s.Type = &v
	return s
}

type GetTaskResponse struct {
	Headers    map[string]*string   `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
	StatusCode *int32               `json:"statusCode,omitempty" xml:"statusCode,omitempty" require:"true"`
	Body       *GetTaskResponseBody `json:"body,omitempty" xml:"body,omitempty" require:"true"`
}

func (s GetTaskResponse) String() string {
	return tea.Prettify(s)
}

func (s GetTaskResponse) GoString() string {
	return s.String()
}

func (s *GetTaskResponse) SetHeaders(v map[string]*string) *GetTaskResponse {
	s.Headers = v
	return s
}

func (s *GetTaskResponse) SetStatusCode(v int32) *GetTaskResponse {
	s.StatusCode = &v
	return s
}

func (s *GetTaskResponse) SetBody(v *GetTaskResponseBody) *GetTaskResponse {
	s.Body = v
	return s
}

type ListDatasourceTablesRequest struct {
	TableName *string `json:"TableName,omitempty" xml:"TableName,omitempty"`
}

func (s ListDatasourceTablesRequest) String() string {
	return tea.Prettify(s)
}

func (s ListDatasourceTablesRequest) GoString() string {
	return s.String()
}

func (s *ListDatasourceTablesRequest) SetTableName(v string) *ListDatasourceTablesRequest {
	s.TableName = &v
	return s
}

type ListDatasourceTablesResponseBody struct {
	RequestId  *string   `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
	Tables     []*string `json:"Tables,omitempty" xml:"Tables,omitempty" type:"Repeated"`
	TotalCount *int64    `json:"TotalCount,omitempty" xml:"TotalCount,omitempty"`
}

func (s ListDatasourceTablesResponseBody) String() string {
	return tea.Prettify(s)
}

func (s ListDatasourceTablesResponseBody) GoString() string {
	return s.String()
}

func (s *ListDatasourceTablesResponseBody) SetRequestId(v string) *ListDatasourceTablesResponseBody {
	s.RequestId = &v
	return s
}

func (s *ListDatasourceTablesResponseBody) SetTables(v []*string) *ListDatasourceTablesResponseBody {
	s.Tables = v
	return s
}

func (s *ListDatasourceTablesResponseBody) SetTotalCount(v int64) *ListDatasourceTablesResponseBody {
	s.TotalCount = &v
	return s
}

type ListDatasourceTablesResponse struct {
	Headers    map[string]*string                `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
	StatusCode *int32                            `json:"statusCode,omitempty" xml:"statusCode,omitempty" require:"true"`
	Body       *ListDatasourceTablesResponseBody `json:"body,omitempty" xml:"body,omitempty" require:"true"`
}

func (s ListDatasourceTablesResponse) String() string {
	return tea.Prettify(s)
}

func (s ListDatasourceTablesResponse) GoString() string {
	return s.String()
}

func (s *ListDatasourceTablesResponse) SetHeaders(v map[string]*string) *ListDatasourceTablesResponse {
	s.Headers = v
	return s
}

func (s *ListDatasourceTablesResponse) SetStatusCode(v int32) *ListDatasourceTablesResponse {
	s.StatusCode = &v
	return s
}

func (s *ListDatasourceTablesResponse) SetBody(v *ListDatasourceTablesResponseBody) *ListDatasourceTablesResponse {
	s.Body = v
	return s
}

type ListDatasourcesRequest struct {
	Name        *string `json:"Name,omitempty" xml:"Name,omitempty"`
	Order       *string `json:"Order,omitempty" xml:"Order,omitempty"`
	PageNumber  *int32  `json:"PageNumber,omitempty" xml:"PageNumber,omitempty"`
	PageSize    *int32  `json:"PageSize,omitempty" xml:"PageSize,omitempty"`
	SortBy      *string `json:"SortBy,omitempty" xml:"SortBy,omitempty"`
	Type        *string `json:"Type,omitempty" xml:"Type,omitempty"`
	WorkspaceId *string `json:"WorkspaceId,omitempty" xml:"WorkspaceId,omitempty"`
}

func (s ListDatasourcesRequest) String() string {
	return tea.Prettify(s)
}

func (s ListDatasourcesRequest) GoString() string {
	return s.String()
}

func (s *ListDatasourcesRequest) SetName(v string) *ListDatasourcesRequest {
	s.Name = &v
	return s
}

func (s *ListDatasourcesRequest) SetOrder(v string) *ListDatasourcesRequest {
	s.Order = &v
	return s
}

func (s *ListDatasourcesRequest) SetPageNumber(v int32) *ListDatasourcesRequest {
	s.PageNumber = &v
	return s
}

func (s *ListDatasourcesRequest) SetPageSize(v int32) *ListDatasourcesRequest {
	s.PageSize = &v
	return s
}

func (s *ListDatasourcesRequest) SetSortBy(v string) *ListDatasourcesRequest {
	s.SortBy = &v
	return s
}

func (s *ListDatasourcesRequest) SetType(v string) *ListDatasourcesRequest {
	s.Type = &v
	return s
}

func (s *ListDatasourcesRequest) SetWorkspaceId(v string) *ListDatasourcesRequest {
	s.WorkspaceId = &v
	return s
}

type ListDatasourcesResponseBody struct {
	Datasources []*ListDatasourcesResponseBodyDatasources `json:"Datasources,omitempty" xml:"Datasources,omitempty" type:"Repeated"`
	RequestId   *string                                   `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
	TotalCount  *int64                                    `json:"TotalCount,omitempty" xml:"TotalCount,omitempty"`
}

func (s ListDatasourcesResponseBody) String() string {
	return tea.Prettify(s)
}

func (s ListDatasourcesResponseBody) GoString() string {
	return s.String()
}

func (s *ListDatasourcesResponseBody) SetDatasources(v []*ListDatasourcesResponseBodyDatasources) *ListDatasourcesResponseBody {
	s.Datasources = v
	return s
}

func (s *ListDatasourcesResponseBody) SetRequestId(v string) *ListDatasourcesResponseBody {
	s.RequestId = &v
	return s
}

func (s *ListDatasourcesResponseBody) SetTotalCount(v int64) *ListDatasourcesResponseBody {
	s.TotalCount = &v
	return s
}

type ListDatasourcesResponseBodyDatasources struct {
	Config          *string `json:"Config,omitempty" xml:"Config,omitempty"`
	DatasourceId    *string `json:"DatasourceId,omitempty" xml:"DatasourceId,omitempty"`
	GmtCreateTime   *string `json:"GmtCreateTime,omitempty" xml:"GmtCreateTime,omitempty"`
	GmtModifiedTime *string `json:"GmtModifiedTime,omitempty" xml:"GmtModifiedTime,omitempty"`
	Name            *string `json:"Name,omitempty" xml:"Name,omitempty"`
	Type            *string `json:"Type,omitempty" xml:"Type,omitempty"`
	Uri             *string `json:"Uri,omitempty" xml:"Uri,omitempty"`
	WorkspaceId     *string `json:"WorkspaceId,omitempty" xml:"WorkspaceId,omitempty"`
}

func (s ListDatasourcesResponseBodyDatasources) String() string {
	return tea.Prettify(s)
}

func (s ListDatasourcesResponseBodyDatasources) GoString() string {
	return s.String()
}

func (s *ListDatasourcesResponseBodyDatasources) SetConfig(v string) *ListDatasourcesResponseBodyDatasources {
	s.Config = &v
	return s
}

func (s *ListDatasourcesResponseBodyDatasources) SetDatasourceId(v string) *ListDatasourcesResponseBodyDatasources {
	s.DatasourceId = &v
	return s
}

func (s *ListDatasourcesResponseBodyDatasources) SetGmtCreateTime(v string) *ListDatasourcesResponseBodyDatasources {
	s.GmtCreateTime = &v
	return s
}

func (s *ListDatasourcesResponseBodyDatasources) SetGmtModifiedTime(v string) *ListDatasourcesResponseBodyDatasources {
	s.GmtModifiedTime = &v
	return s
}

func (s *ListDatasourcesResponseBodyDatasources) SetName(v string) *ListDatasourcesResponseBodyDatasources {
	s.Name = &v
	return s
}

func (s *ListDatasourcesResponseBodyDatasources) SetType(v string) *ListDatasourcesResponseBodyDatasources {
	s.Type = &v
	return s
}

func (s *ListDatasourcesResponseBodyDatasources) SetUri(v string) *ListDatasourcesResponseBodyDatasources {
	s.Uri = &v
	return s
}

func (s *ListDatasourcesResponseBodyDatasources) SetWorkspaceId(v string) *ListDatasourcesResponseBodyDatasources {
	s.WorkspaceId = &v
	return s
}

type ListDatasourcesResponse struct {
	Headers    map[string]*string           `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
	StatusCode *int32                       `json:"statusCode,omitempty" xml:"statusCode,omitempty" require:"true"`
	Body       *ListDatasourcesResponseBody `json:"body,omitempty" xml:"body,omitempty" require:"true"`
}

func (s ListDatasourcesResponse) String() string {
	return tea.Prettify(s)
}

func (s ListDatasourcesResponse) GoString() string {
	return s.String()
}

func (s *ListDatasourcesResponse) SetHeaders(v map[string]*string) *ListDatasourcesResponse {
	s.Headers = v
	return s
}

func (s *ListDatasourcesResponse) SetStatusCode(v int32) *ListDatasourcesResponse {
	s.StatusCode = &v
	return s
}

func (s *ListDatasourcesResponse) SetBody(v *ListDatasourcesResponseBody) *ListDatasourcesResponse {
	s.Body = v
	return s
}

type ListFeatureEntitiesRequest struct {
	FeatureEntityIds []*string `json:"FeatureEntityIds,omitempty" xml:"FeatureEntityIds,omitempty" type:"Repeated"`
	Name             *string   `json:"Name,omitempty" xml:"Name,omitempty"`
	Order            *string   `json:"Order,omitempty" xml:"Order,omitempty"`
	Owner            *string   `json:"Owner,omitempty" xml:"Owner,omitempty"`
	PageNumber       *int32    `json:"PageNumber,omitempty" xml:"PageNumber,omitempty"`
	PageSize         *int32    `json:"PageSize,omitempty" xml:"PageSize,omitempty"`
	ProjectId        *string   `json:"ProjectId,omitempty" xml:"ProjectId,omitempty"`
	SortBy           *string   `json:"SortBy,omitempty" xml:"SortBy,omitempty"`
}

func (s ListFeatureEntitiesRequest) String() string {
	return tea.Prettify(s)
}

func (s ListFeatureEntitiesRequest) GoString() string {
	return s.String()
}

func (s *ListFeatureEntitiesRequest) SetFeatureEntityIds(v []*string) *ListFeatureEntitiesRequest {
	s.FeatureEntityIds = v
	return s
}

func (s *ListFeatureEntitiesRequest) SetName(v string) *ListFeatureEntitiesRequest {
	s.Name = &v
	return s
}

func (s *ListFeatureEntitiesRequest) SetOrder(v string) *ListFeatureEntitiesRequest {
	s.Order = &v
	return s
}

func (s *ListFeatureEntitiesRequest) SetOwner(v string) *ListFeatureEntitiesRequest {
	s.Owner = &v
	return s
}

func (s *ListFeatureEntitiesRequest) SetPageNumber(v int32) *ListFeatureEntitiesRequest {
	s.PageNumber = &v
	return s
}

func (s *ListFeatureEntitiesRequest) SetPageSize(v int32) *ListFeatureEntitiesRequest {
	s.PageSize = &v
	return s
}

func (s *ListFeatureEntitiesRequest) SetProjectId(v string) *ListFeatureEntitiesRequest {
	s.ProjectId = &v
	return s
}

func (s *ListFeatureEntitiesRequest) SetSortBy(v string) *ListFeatureEntitiesRequest {
	s.SortBy = &v
	return s
}

type ListFeatureEntitiesShrinkRequest struct {
	FeatureEntityIdsShrink *string `json:"FeatureEntityIds,omitempty" xml:"FeatureEntityIds,omitempty"`
	Name                   *string `json:"Name,omitempty" xml:"Name,omitempty"`
	Order                  *string `json:"Order,omitempty" xml:"Order,omitempty"`
	Owner                  *string `json:"Owner,omitempty" xml:"Owner,omitempty"`
	PageNumber             *int32  `json:"PageNumber,omitempty" xml:"PageNumber,omitempty"`
	PageSize               *int32  `json:"PageSize,omitempty" xml:"PageSize,omitempty"`
	ProjectId              *string `json:"ProjectId,omitempty" xml:"ProjectId,omitempty"`
	SortBy                 *string `json:"SortBy,omitempty" xml:"SortBy,omitempty"`
}

func (s ListFeatureEntitiesShrinkRequest) String() string {
	return tea.Prettify(s)
}

func (s ListFeatureEntitiesShrinkRequest) GoString() string {
	return s.String()
}

func (s *ListFeatureEntitiesShrinkRequest) SetFeatureEntityIdsShrink(v string) *ListFeatureEntitiesShrinkRequest {
	s.FeatureEntityIdsShrink = &v
	return s
}

func (s *ListFeatureEntitiesShrinkRequest) SetName(v string) *ListFeatureEntitiesShrinkRequest {
	s.Name = &v
	return s
}

func (s *ListFeatureEntitiesShrinkRequest) SetOrder(v string) *ListFeatureEntitiesShrinkRequest {
	s.Order = &v
	return s
}

func (s *ListFeatureEntitiesShrinkRequest) SetOwner(v string) *ListFeatureEntitiesShrinkRequest {
	s.Owner = &v
	return s
}

func (s *ListFeatureEntitiesShrinkRequest) SetPageNumber(v int32) *ListFeatureEntitiesShrinkRequest {
	s.PageNumber = &v
	return s
}

func (s *ListFeatureEntitiesShrinkRequest) SetPageSize(v int32) *ListFeatureEntitiesShrinkRequest {
	s.PageSize = &v
	return s
}

func (s *ListFeatureEntitiesShrinkRequest) SetProjectId(v string) *ListFeatureEntitiesShrinkRequest {
	s.ProjectId = &v
	return s
}

func (s *ListFeatureEntitiesShrinkRequest) SetSortBy(v string) *ListFeatureEntitiesShrinkRequest {
	s.SortBy = &v
	return s
}

type ListFeatureEntitiesResponseBody struct {
	FeatureEntities []*ListFeatureEntitiesResponseBodyFeatureEntities `json:"FeatureEntities,omitempty" xml:"FeatureEntities,omitempty" type:"Repeated"`
	RequestId       *string                                           `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
	TotalCount      *int32                                            `json:"TotalCount,omitempty" xml:"TotalCount,omitempty"`
}

func (s ListFeatureEntitiesResponseBody) String() string {
	return tea.Prettify(s)
}

func (s ListFeatureEntitiesResponseBody) GoString() string {
	return s.String()
}

func (s *ListFeatureEntitiesResponseBody) SetFeatureEntities(v []*ListFeatureEntitiesResponseBodyFeatureEntities) *ListFeatureEntitiesResponseBody {
	s.FeatureEntities = v
	return s
}

func (s *ListFeatureEntitiesResponseBody) SetRequestId(v string) *ListFeatureEntitiesResponseBody {
	s.RequestId = &v
	return s
}

func (s *ListFeatureEntitiesResponseBody) SetTotalCount(v int32) *ListFeatureEntitiesResponseBody {
	s.TotalCount = &v
	return s
}

type ListFeatureEntitiesResponseBodyFeatureEntities struct {
	FeatureEntityId *string `json:"FeatureEntityId,omitempty" xml:"FeatureEntityId,omitempty"`
	GmtCreateTime   *string `json:"GmtCreateTime,omitempty" xml:"GmtCreateTime,omitempty"`
	JoinId          *string `json:"JoinId,omitempty" xml:"JoinId,omitempty"`
	Name            *string `json:"Name,omitempty" xml:"Name,omitempty"`
	Owner           *string `json:"Owner,omitempty" xml:"Owner,omitempty"`
	ProjectId       *string `json:"ProjectId,omitempty" xml:"ProjectId,omitempty"`
	ProjectName     *string `json:"ProjectName,omitempty" xml:"ProjectName,omitempty"`
}

func (s ListFeatureEntitiesResponseBodyFeatureEntities) String() string {
	return tea.Prettify(s)
}

func (s ListFeatureEntitiesResponseBodyFeatureEntities) GoString() string {
	return s.String()
}

func (s *ListFeatureEntitiesResponseBodyFeatureEntities) SetFeatureEntityId(v string) *ListFeatureEntitiesResponseBodyFeatureEntities {
	s.FeatureEntityId = &v
	return s
}

func (s *ListFeatureEntitiesResponseBodyFeatureEntities) SetGmtCreateTime(v string) *ListFeatureEntitiesResponseBodyFeatureEntities {
	s.GmtCreateTime = &v
	return s
}

func (s *ListFeatureEntitiesResponseBodyFeatureEntities) SetJoinId(v string) *ListFeatureEntitiesResponseBodyFeatureEntities {
	s.JoinId = &v
	return s
}

func (s *ListFeatureEntitiesResponseBodyFeatureEntities) SetName(v string) *ListFeatureEntitiesResponseBodyFeatureEntities {
	s.Name = &v
	return s
}

func (s *ListFeatureEntitiesResponseBodyFeatureEntities) SetOwner(v string) *ListFeatureEntitiesResponseBodyFeatureEntities {
	s.Owner = &v
	return s
}

func (s *ListFeatureEntitiesResponseBodyFeatureEntities) SetProjectId(v string) *ListFeatureEntitiesResponseBodyFeatureEntities {
	s.ProjectId = &v
	return s
}

func (s *ListFeatureEntitiesResponseBodyFeatureEntities) SetProjectName(v string) *ListFeatureEntitiesResponseBodyFeatureEntities {
	s.ProjectName = &v
	return s
}

type ListFeatureEntitiesResponse struct {
	Headers    map[string]*string               `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
	StatusCode *int32                           `json:"statusCode,omitempty" xml:"statusCode,omitempty" require:"true"`
	Body       *ListFeatureEntitiesResponseBody `json:"body,omitempty" xml:"body,omitempty" require:"true"`
}

func (s ListFeatureEntitiesResponse) String() string {
	return tea.Prettify(s)
}

func (s ListFeatureEntitiesResponse) GoString() string {
	return s.String()
}

func (s *ListFeatureEntitiesResponse) SetHeaders(v map[string]*string) *ListFeatureEntitiesResponse {
	s.Headers = v
	return s
}

func (s *ListFeatureEntitiesResponse) SetStatusCode(v int32) *ListFeatureEntitiesResponse {
	s.StatusCode = &v
	return s
}

func (s *ListFeatureEntitiesResponse) SetBody(v *ListFeatureEntitiesResponseBody) *ListFeatureEntitiesResponse {
	s.Body = v
	return s
}

type ListFeatureViewFieldRelationshipsResponseBody struct {
	Relationships []*ListFeatureViewFieldRelationshipsResponseBodyRelationships `json:"Relationships,omitempty" xml:"Relationships,omitempty" type:"Repeated"`
	RequestId     *string                                                       `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
}

func (s ListFeatureViewFieldRelationshipsResponseBody) String() string {
	return tea.Prettify(s)
}

func (s ListFeatureViewFieldRelationshipsResponseBody) GoString() string {
	return s.String()
}

func (s *ListFeatureViewFieldRelationshipsResponseBody) SetRelationships(v []*ListFeatureViewFieldRelationshipsResponseBodyRelationships) *ListFeatureViewFieldRelationshipsResponseBody {
	s.Relationships = v
	return s
}

func (s *ListFeatureViewFieldRelationshipsResponseBody) SetRequestId(v string) *ListFeatureViewFieldRelationshipsResponseBody {
	s.RequestId = &v
	return s
}

type ListFeatureViewFieldRelationshipsResponseBodyRelationships struct {
	FeatureName      *string                                                             `json:"FeatureName,omitempty" xml:"FeatureName,omitempty"`
	Models           []*ListFeatureViewFieldRelationshipsResponseBodyRelationshipsModels `json:"Models,omitempty" xml:"Models,omitempty" type:"Repeated"`
	OfflineTableName *string                                                             `json:"OfflineTableName,omitempty" xml:"OfflineTableName,omitempty"`
	OnlineTableName  *string                                                             `json:"OnlineTableName,omitempty" xml:"OnlineTableName,omitempty"`
}

func (s ListFeatureViewFieldRelationshipsResponseBodyRelationships) String() string {
	return tea.Prettify(s)
}

func (s ListFeatureViewFieldRelationshipsResponseBodyRelationships) GoString() string {
	return s.String()
}

func (s *ListFeatureViewFieldRelationshipsResponseBodyRelationships) SetFeatureName(v string) *ListFeatureViewFieldRelationshipsResponseBodyRelationships {
	s.FeatureName = &v
	return s
}

func (s *ListFeatureViewFieldRelationshipsResponseBodyRelationships) SetModels(v []*ListFeatureViewFieldRelationshipsResponseBodyRelationshipsModels) *ListFeatureViewFieldRelationshipsResponseBodyRelationships {
	s.Models = v
	return s
}

func (s *ListFeatureViewFieldRelationshipsResponseBodyRelationships) SetOfflineTableName(v string) *ListFeatureViewFieldRelationshipsResponseBodyRelationships {
	s.OfflineTableName = &v
	return s
}

func (s *ListFeatureViewFieldRelationshipsResponseBodyRelationships) SetOnlineTableName(v string) *ListFeatureViewFieldRelationshipsResponseBodyRelationships {
	s.OnlineTableName = &v
	return s
}

type ListFeatureViewFieldRelationshipsResponseBodyRelationshipsModels struct {
	ModelId   *string `json:"ModelId,omitempty" xml:"ModelId,omitempty"`
	ModelName *string `json:"ModelName,omitempty" xml:"ModelName,omitempty"`
}

func (s ListFeatureViewFieldRelationshipsResponseBodyRelationshipsModels) String() string {
	return tea.Prettify(s)
}

func (s ListFeatureViewFieldRelationshipsResponseBodyRelationshipsModels) GoString() string {
	return s.String()
}

func (s *ListFeatureViewFieldRelationshipsResponseBodyRelationshipsModels) SetModelId(v string) *ListFeatureViewFieldRelationshipsResponseBodyRelationshipsModels {
	s.ModelId = &v
	return s
}

func (s *ListFeatureViewFieldRelationshipsResponseBodyRelationshipsModels) SetModelName(v string) *ListFeatureViewFieldRelationshipsResponseBodyRelationshipsModels {
	s.ModelName = &v
	return s
}

type ListFeatureViewFieldRelationshipsResponse struct {
	Headers    map[string]*string                             `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
	StatusCode *int32                                         `json:"statusCode,omitempty" xml:"statusCode,omitempty" require:"true"`
	Body       *ListFeatureViewFieldRelationshipsResponseBody `json:"body,omitempty" xml:"body,omitempty" require:"true"`
}

func (s ListFeatureViewFieldRelationshipsResponse) String() string {
	return tea.Prettify(s)
}

func (s ListFeatureViewFieldRelationshipsResponse) GoString() string {
	return s.String()
}

func (s *ListFeatureViewFieldRelationshipsResponse) SetHeaders(v map[string]*string) *ListFeatureViewFieldRelationshipsResponse {
	s.Headers = v
	return s
}

func (s *ListFeatureViewFieldRelationshipsResponse) SetStatusCode(v int32) *ListFeatureViewFieldRelationshipsResponse {
	s.StatusCode = &v
	return s
}

func (s *ListFeatureViewFieldRelationshipsResponse) SetBody(v *ListFeatureViewFieldRelationshipsResponseBody) *ListFeatureViewFieldRelationshipsResponse {
	s.Body = v
	return s
}

type ListFeatureViewRelationshipsResponseBody struct {
	Relationships []*ListFeatureViewRelationshipsResponseBodyRelationships `json:"Relationships,omitempty" xml:"Relationships,omitempty" type:"Repeated"`
	RequestId     *string                                                  `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
}

func (s ListFeatureViewRelationshipsResponseBody) String() string {
	return tea.Prettify(s)
}

func (s ListFeatureViewRelationshipsResponseBody) GoString() string {
	return s.String()
}

func (s *ListFeatureViewRelationshipsResponseBody) SetRelationships(v []*ListFeatureViewRelationshipsResponseBodyRelationships) *ListFeatureViewRelationshipsResponseBody {
	s.Relationships = v
	return s
}

func (s *ListFeatureViewRelationshipsResponseBody) SetRequestId(v string) *ListFeatureViewRelationshipsResponseBody {
	s.RequestId = &v
	return s
}

type ListFeatureViewRelationshipsResponseBodyRelationships struct {
	FeatureViewName *string                                                        `json:"FeatureViewName,omitempty" xml:"FeatureViewName,omitempty"`
	Models          []*ListFeatureViewRelationshipsResponseBodyRelationshipsModels `json:"Models,omitempty" xml:"Models,omitempty" type:"Repeated"`
	ProjectName     *string                                                        `json:"ProjectName,omitempty" xml:"ProjectName,omitempty"`
}

func (s ListFeatureViewRelationshipsResponseBodyRelationships) String() string {
	return tea.Prettify(s)
}

func (s ListFeatureViewRelationshipsResponseBodyRelationships) GoString() string {
	return s.String()
}

func (s *ListFeatureViewRelationshipsResponseBodyRelationships) SetFeatureViewName(v string) *ListFeatureViewRelationshipsResponseBodyRelationships {
	s.FeatureViewName = &v
	return s
}

func (s *ListFeatureViewRelationshipsResponseBodyRelationships) SetModels(v []*ListFeatureViewRelationshipsResponseBodyRelationshipsModels) *ListFeatureViewRelationshipsResponseBodyRelationships {
	s.Models = v
	return s
}

func (s *ListFeatureViewRelationshipsResponseBodyRelationships) SetProjectName(v string) *ListFeatureViewRelationshipsResponseBodyRelationships {
	s.ProjectName = &v
	return s
}

type ListFeatureViewRelationshipsResponseBodyRelationshipsModels struct {
	ModelId   *string `json:"ModelId,omitempty" xml:"ModelId,omitempty"`
	ModelName *string `json:"ModelName,omitempty" xml:"ModelName,omitempty"`
}

func (s ListFeatureViewRelationshipsResponseBodyRelationshipsModels) String() string {
	return tea.Prettify(s)
}

func (s ListFeatureViewRelationshipsResponseBodyRelationshipsModels) GoString() string {
	return s.String()
}

func (s *ListFeatureViewRelationshipsResponseBodyRelationshipsModels) SetModelId(v string) *ListFeatureViewRelationshipsResponseBodyRelationshipsModels {
	s.ModelId = &v
	return s
}

func (s *ListFeatureViewRelationshipsResponseBodyRelationshipsModels) SetModelName(v string) *ListFeatureViewRelationshipsResponseBodyRelationshipsModels {
	s.ModelName = &v
	return s
}

type ListFeatureViewRelationshipsResponse struct {
	Headers    map[string]*string                        `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
	StatusCode *int32                                    `json:"statusCode,omitempty" xml:"statusCode,omitempty" require:"true"`
	Body       *ListFeatureViewRelationshipsResponseBody `json:"body,omitempty" xml:"body,omitempty" require:"true"`
}

func (s ListFeatureViewRelationshipsResponse) String() string {
	return tea.Prettify(s)
}

func (s ListFeatureViewRelationshipsResponse) GoString() string {
	return s.String()
}

func (s *ListFeatureViewRelationshipsResponse) SetHeaders(v map[string]*string) *ListFeatureViewRelationshipsResponse {
	s.Headers = v
	return s
}

func (s *ListFeatureViewRelationshipsResponse) SetStatusCode(v int32) *ListFeatureViewRelationshipsResponse {
	s.StatusCode = &v
	return s
}

func (s *ListFeatureViewRelationshipsResponse) SetBody(v *ListFeatureViewRelationshipsResponseBody) *ListFeatureViewRelationshipsResponse {
	s.Body = v
	return s
}

type ListFeatureViewsRequest struct {
	FeatureName    *string   `json:"FeatureName,omitempty" xml:"FeatureName,omitempty"`
	FeatureViewIds []*string `json:"FeatureViewIds,omitempty" xml:"FeatureViewIds,omitempty" type:"Repeated"`
	Order          *string   `json:"Order,omitempty" xml:"Order,omitempty"`
	Owner          *string   `json:"Owner,omitempty" xml:"Owner,omitempty"`
	PageNumber     *int32    `json:"PageNumber,omitempty" xml:"PageNumber,omitempty"`
	PageSize       *int32    `json:"PageSize,omitempty" xml:"PageSize,omitempty"`
	ProjectId      *string   `json:"ProjectId,omitempty" xml:"ProjectId,omitempty"`
	SortBy         *string   `json:"SortBy,omitempty" xml:"SortBy,omitempty"`
	Tag            *string   `json:"Tag,omitempty" xml:"Tag,omitempty"`
	Type           *string   `json:"Type,omitempty" xml:"Type,omitempty"`
}

func (s ListFeatureViewsRequest) String() string {
	return tea.Prettify(s)
}

func (s ListFeatureViewsRequest) GoString() string {
	return s.String()
}

func (s *ListFeatureViewsRequest) SetFeatureName(v string) *ListFeatureViewsRequest {
	s.FeatureName = &v
	return s
}

func (s *ListFeatureViewsRequest) SetFeatureViewIds(v []*string) *ListFeatureViewsRequest {
	s.FeatureViewIds = v
	return s
}

func (s *ListFeatureViewsRequest) SetOrder(v string) *ListFeatureViewsRequest {
	s.Order = &v
	return s
}

func (s *ListFeatureViewsRequest) SetOwner(v string) *ListFeatureViewsRequest {
	s.Owner = &v
	return s
}

func (s *ListFeatureViewsRequest) SetPageNumber(v int32) *ListFeatureViewsRequest {
	s.PageNumber = &v
	return s
}

func (s *ListFeatureViewsRequest) SetPageSize(v int32) *ListFeatureViewsRequest {
	s.PageSize = &v
	return s
}

func (s *ListFeatureViewsRequest) SetProjectId(v string) *ListFeatureViewsRequest {
	s.ProjectId = &v
	return s
}

func (s *ListFeatureViewsRequest) SetSortBy(v string) *ListFeatureViewsRequest {
	s.SortBy = &v
	return s
}

func (s *ListFeatureViewsRequest) SetTag(v string) *ListFeatureViewsRequest {
	s.Tag = &v
	return s
}

func (s *ListFeatureViewsRequest) SetType(v string) *ListFeatureViewsRequest {
	s.Type = &v
	return s
}

type ListFeatureViewsShrinkRequest struct {
	FeatureName          *string `json:"FeatureName,omitempty" xml:"FeatureName,omitempty"`
	FeatureViewIdsShrink *string `json:"FeatureViewIds,omitempty" xml:"FeatureViewIds,omitempty"`
	Order                *string `json:"Order,omitempty" xml:"Order,omitempty"`
	Owner                *string `json:"Owner,omitempty" xml:"Owner,omitempty"`
	PageNumber           *int32  `json:"PageNumber,omitempty" xml:"PageNumber,omitempty"`
	PageSize             *int32  `json:"PageSize,omitempty" xml:"PageSize,omitempty"`
	ProjectId            *string `json:"ProjectId,omitempty" xml:"ProjectId,omitempty"`
	SortBy               *string `json:"SortBy,omitempty" xml:"SortBy,omitempty"`
	Tag                  *string `json:"Tag,omitempty" xml:"Tag,omitempty"`
	Type                 *string `json:"Type,omitempty" xml:"Type,omitempty"`
}

func (s ListFeatureViewsShrinkRequest) String() string {
	return tea.Prettify(s)
}

func (s ListFeatureViewsShrinkRequest) GoString() string {
	return s.String()
}

func (s *ListFeatureViewsShrinkRequest) SetFeatureName(v string) *ListFeatureViewsShrinkRequest {
	s.FeatureName = &v
	return s
}

func (s *ListFeatureViewsShrinkRequest) SetFeatureViewIdsShrink(v string) *ListFeatureViewsShrinkRequest {
	s.FeatureViewIdsShrink = &v
	return s
}

func (s *ListFeatureViewsShrinkRequest) SetOrder(v string) *ListFeatureViewsShrinkRequest {
	s.Order = &v
	return s
}

func (s *ListFeatureViewsShrinkRequest) SetOwner(v string) *ListFeatureViewsShrinkRequest {
	s.Owner = &v
	return s
}

func (s *ListFeatureViewsShrinkRequest) SetPageNumber(v int32) *ListFeatureViewsShrinkRequest {
	s.PageNumber = &v
	return s
}

func (s *ListFeatureViewsShrinkRequest) SetPageSize(v int32) *ListFeatureViewsShrinkRequest {
	s.PageSize = &v
	return s
}

func (s *ListFeatureViewsShrinkRequest) SetProjectId(v string) *ListFeatureViewsShrinkRequest {
	s.ProjectId = &v
	return s
}

func (s *ListFeatureViewsShrinkRequest) SetSortBy(v string) *ListFeatureViewsShrinkRequest {
	s.SortBy = &v
	return s
}

func (s *ListFeatureViewsShrinkRequest) SetTag(v string) *ListFeatureViewsShrinkRequest {
	s.Tag = &v
	return s
}

func (s *ListFeatureViewsShrinkRequest) SetType(v string) *ListFeatureViewsShrinkRequest {
	s.Type = &v
	return s
}

type ListFeatureViewsResponseBody struct {
	FeatureViews []*ListFeatureViewsResponseBodyFeatureViews `json:"FeatureViews,omitempty" xml:"FeatureViews,omitempty" type:"Repeated"`
	RequestId    *string                                     `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
	TotalCount   *int64                                      `json:"TotalCount,omitempty" xml:"TotalCount,omitempty"`
}

func (s ListFeatureViewsResponseBody) String() string {
	return tea.Prettify(s)
}

func (s ListFeatureViewsResponseBody) GoString() string {
	return s.String()
}

func (s *ListFeatureViewsResponseBody) SetFeatureViews(v []*ListFeatureViewsResponseBodyFeatureViews) *ListFeatureViewsResponseBody {
	s.FeatureViews = v
	return s
}

func (s *ListFeatureViewsResponseBody) SetRequestId(v string) *ListFeatureViewsResponseBody {
	s.RequestId = &v
	return s
}

func (s *ListFeatureViewsResponseBody) SetTotalCount(v int64) *ListFeatureViewsResponseBody {
	s.TotalCount = &v
	return s
}

type ListFeatureViewsResponseBodyFeatureViews struct {
	FeatureEntityName      *string `json:"FeatureEntityName,omitempty" xml:"FeatureEntityName,omitempty"`
	FeatureViewId          *string `json:"FeatureViewId,omitempty" xml:"FeatureViewId,omitempty"`
	GmtCreateTime          *string `json:"GmtCreateTime,omitempty" xml:"GmtCreateTime,omitempty"`
	GmtModifiedTime        *string `json:"GmtModifiedTime,omitempty" xml:"GmtModifiedTime,omitempty"`
	Name                   *string `json:"Name,omitempty" xml:"Name,omitempty"`
	Owner                  *string `json:"Owner,omitempty" xml:"Owner,omitempty"`
	ProjectId              *string `json:"ProjectId,omitempty" xml:"ProjectId,omitempty"`
	ProjectName            *string `json:"ProjectName,omitempty" xml:"ProjectName,omitempty"`
	RegisterDatasourceId   *string `json:"RegisterDatasourceId,omitempty" xml:"RegisterDatasourceId,omitempty"`
	RegisterDatasourceName *string `json:"RegisterDatasourceName,omitempty" xml:"RegisterDatasourceName,omitempty"`
	RegisterTable          *string `json:"RegisterTable,omitempty" xml:"RegisterTable,omitempty"`
	TTL                    *int32  `json:"TTL,omitempty" xml:"TTL,omitempty"`
	Type                   *string `json:"Type,omitempty" xml:"Type,omitempty"`
}

func (s ListFeatureViewsResponseBodyFeatureViews) String() string {
	return tea.Prettify(s)
}

func (s ListFeatureViewsResponseBodyFeatureViews) GoString() string {
	return s.String()
}

func (s *ListFeatureViewsResponseBodyFeatureViews) SetFeatureEntityName(v string) *ListFeatureViewsResponseBodyFeatureViews {
	s.FeatureEntityName = &v
	return s
}

func (s *ListFeatureViewsResponseBodyFeatureViews) SetFeatureViewId(v string) *ListFeatureViewsResponseBodyFeatureViews {
	s.FeatureViewId = &v
	return s
}

func (s *ListFeatureViewsResponseBodyFeatureViews) SetGmtCreateTime(v string) *ListFeatureViewsResponseBodyFeatureViews {
	s.GmtCreateTime = &v
	return s
}

func (s *ListFeatureViewsResponseBodyFeatureViews) SetGmtModifiedTime(v string) *ListFeatureViewsResponseBodyFeatureViews {
	s.GmtModifiedTime = &v
	return s
}

func (s *ListFeatureViewsResponseBodyFeatureViews) SetName(v string) *ListFeatureViewsResponseBodyFeatureViews {
	s.Name = &v
	return s
}

func (s *ListFeatureViewsResponseBodyFeatureViews) SetOwner(v string) *ListFeatureViewsResponseBodyFeatureViews {
	s.Owner = &v
	return s
}

func (s *ListFeatureViewsResponseBodyFeatureViews) SetProjectId(v string) *ListFeatureViewsResponseBodyFeatureViews {
	s.ProjectId = &v
	return s
}

func (s *ListFeatureViewsResponseBodyFeatureViews) SetProjectName(v string) *ListFeatureViewsResponseBodyFeatureViews {
	s.ProjectName = &v
	return s
}

func (s *ListFeatureViewsResponseBodyFeatureViews) SetRegisterDatasourceId(v string) *ListFeatureViewsResponseBodyFeatureViews {
	s.RegisterDatasourceId = &v
	return s
}

func (s *ListFeatureViewsResponseBodyFeatureViews) SetRegisterDatasourceName(v string) *ListFeatureViewsResponseBodyFeatureViews {
	s.RegisterDatasourceName = &v
	return s
}

func (s *ListFeatureViewsResponseBodyFeatureViews) SetRegisterTable(v string) *ListFeatureViewsResponseBodyFeatureViews {
	s.RegisterTable = &v
	return s
}

func (s *ListFeatureViewsResponseBodyFeatureViews) SetTTL(v int32) *ListFeatureViewsResponseBodyFeatureViews {
	s.TTL = &v
	return s
}

func (s *ListFeatureViewsResponseBodyFeatureViews) SetType(v string) *ListFeatureViewsResponseBodyFeatureViews {
	s.Type = &v
	return s
}

type ListFeatureViewsResponse struct {
	Headers    map[string]*string            `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
	StatusCode *int32                        `json:"statusCode,omitempty" xml:"statusCode,omitempty" require:"true"`
	Body       *ListFeatureViewsResponseBody `json:"body,omitempty" xml:"body,omitempty" require:"true"`
}

func (s ListFeatureViewsResponse) String() string {
	return tea.Prettify(s)
}

func (s ListFeatureViewsResponse) GoString() string {
	return s.String()
}

func (s *ListFeatureViewsResponse) SetHeaders(v map[string]*string) *ListFeatureViewsResponse {
	s.Headers = v
	return s
}

func (s *ListFeatureViewsResponse) SetStatusCode(v int32) *ListFeatureViewsResponse {
	s.StatusCode = &v
	return s
}

func (s *ListFeatureViewsResponse) SetBody(v *ListFeatureViewsResponseBody) *ListFeatureViewsResponse {
	s.Body = v
	return s
}

type ListInstancesRequest struct {
	Order      *string `json:"Order,omitempty" xml:"Order,omitempty"`
	PageNumber *int32  `json:"PageNumber,omitempty" xml:"PageNumber,omitempty"`
	PageSize   *int32  `json:"PageSize,omitempty" xml:"PageSize,omitempty"`
	SortBy     *string `json:"SortBy,omitempty" xml:"SortBy,omitempty"`
	Status     *string `json:"Status,omitempty" xml:"Status,omitempty"`
}

func (s ListInstancesRequest) String() string {
	return tea.Prettify(s)
}

func (s ListInstancesRequest) GoString() string {
	return s.String()
}

func (s *ListInstancesRequest) SetOrder(v string) *ListInstancesRequest {
	s.Order = &v
	return s
}

func (s *ListInstancesRequest) SetPageNumber(v int32) *ListInstancesRequest {
	s.PageNumber = &v
	return s
}

func (s *ListInstancesRequest) SetPageSize(v int32) *ListInstancesRequest {
	s.PageSize = &v
	return s
}

func (s *ListInstancesRequest) SetSortBy(v string) *ListInstancesRequest {
	s.SortBy = &v
	return s
}

func (s *ListInstancesRequest) SetStatus(v string) *ListInstancesRequest {
	s.Status = &v
	return s
}

type ListInstancesResponseBody struct {
	Instances  []*ListInstancesResponseBodyInstances `json:"Instances,omitempty" xml:"Instances,omitempty" type:"Repeated"`
	RequestId  *string                               `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
	TotalCount *int64                                `json:"TotalCount,omitempty" xml:"TotalCount,omitempty"`
}

func (s ListInstancesResponseBody) String() string {
	return tea.Prettify(s)
}

func (s ListInstancesResponseBody) GoString() string {
	return s.String()
}

func (s *ListInstancesResponseBody) SetInstances(v []*ListInstancesResponseBodyInstances) *ListInstancesResponseBody {
	s.Instances = v
	return s
}

func (s *ListInstancesResponseBody) SetRequestId(v string) *ListInstancesResponseBody {
	s.RequestId = &v
	return s
}

func (s *ListInstancesResponseBody) SetTotalCount(v int64) *ListInstancesResponseBody {
	s.TotalCount = &v
	return s
}

type ListInstancesResponseBodyInstances struct {
	GmtCreateTime   *string `json:"GmtCreateTime,omitempty" xml:"GmtCreateTime,omitempty"`
	GmtModifiedTime *string `json:"GmtModifiedTime,omitempty" xml:"GmtModifiedTime,omitempty"`
	InstanceId      *string `json:"InstanceId,omitempty" xml:"InstanceId,omitempty"`
	RegionId        *string `json:"RegionId,omitempty" xml:"RegionId,omitempty"`
	Status          *string `json:"Status,omitempty" xml:"Status,omitempty"`
	Type            *string `json:"Type,omitempty" xml:"Type,omitempty"`
}

func (s ListInstancesResponseBodyInstances) String() string {
	return tea.Prettify(s)
}

func (s ListInstancesResponseBodyInstances) GoString() string {
	return s.String()
}

func (s *ListInstancesResponseBodyInstances) SetGmtCreateTime(v string) *ListInstancesResponseBodyInstances {
	s.GmtCreateTime = &v
	return s
}

func (s *ListInstancesResponseBodyInstances) SetGmtModifiedTime(v string) *ListInstancesResponseBodyInstances {
	s.GmtModifiedTime = &v
	return s
}

func (s *ListInstancesResponseBodyInstances) SetInstanceId(v string) *ListInstancesResponseBodyInstances {
	s.InstanceId = &v
	return s
}

func (s *ListInstancesResponseBodyInstances) SetRegionId(v string) *ListInstancesResponseBodyInstances {
	s.RegionId = &v
	return s
}

func (s *ListInstancesResponseBodyInstances) SetStatus(v string) *ListInstancesResponseBodyInstances {
	s.Status = &v
	return s
}

func (s *ListInstancesResponseBodyInstances) SetType(v string) *ListInstancesResponseBodyInstances {
	s.Type = &v
	return s
}

type ListInstancesResponse struct {
	Headers    map[string]*string         `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
	StatusCode *int32                     `json:"statusCode,omitempty" xml:"statusCode,omitempty" require:"true"`
	Body       *ListInstancesResponseBody `json:"body,omitempty" xml:"body,omitempty" require:"true"`
}

func (s ListInstancesResponse) String() string {
	return tea.Prettify(s)
}

func (s ListInstancesResponse) GoString() string {
	return s.String()
}

func (s *ListInstancesResponse) SetHeaders(v map[string]*string) *ListInstancesResponse {
	s.Headers = v
	return s
}

func (s *ListInstancesResponse) SetStatusCode(v int32) *ListInstancesResponse {
	s.StatusCode = &v
	return s
}

func (s *ListInstancesResponse) SetBody(v *ListInstancesResponseBody) *ListInstancesResponse {
	s.Body = v
	return s
}

type ListLabelTablesRequest struct {
	LabelTableIds []*string `json:"LabelTableIds,omitempty" xml:"LabelTableIds,omitempty" type:"Repeated"`
	Name          *string   `json:"Name,omitempty" xml:"Name,omitempty"`
	Order         *string   `json:"Order,omitempty" xml:"Order,omitempty"`
	Owner         *string   `json:"Owner,omitempty" xml:"Owner,omitempty"`
	PageNumber    *int64    `json:"PageNumber,omitempty" xml:"PageNumber,omitempty"`
	PageSize      *int64    `json:"PageSize,omitempty" xml:"PageSize,omitempty"`
	ProjectId     *string   `json:"ProjectId,omitempty" xml:"ProjectId,omitempty"`
	SortBy        *string   `json:"SortBy,omitempty" xml:"SortBy,omitempty"`
}

func (s ListLabelTablesRequest) String() string {
	return tea.Prettify(s)
}

func (s ListLabelTablesRequest) GoString() string {
	return s.String()
}

func (s *ListLabelTablesRequest) SetLabelTableIds(v []*string) *ListLabelTablesRequest {
	s.LabelTableIds = v
	return s
}

func (s *ListLabelTablesRequest) SetName(v string) *ListLabelTablesRequest {
	s.Name = &v
	return s
}

func (s *ListLabelTablesRequest) SetOrder(v string) *ListLabelTablesRequest {
	s.Order = &v
	return s
}

func (s *ListLabelTablesRequest) SetOwner(v string) *ListLabelTablesRequest {
	s.Owner = &v
	return s
}

func (s *ListLabelTablesRequest) SetPageNumber(v int64) *ListLabelTablesRequest {
	s.PageNumber = &v
	return s
}

func (s *ListLabelTablesRequest) SetPageSize(v int64) *ListLabelTablesRequest {
	s.PageSize = &v
	return s
}

func (s *ListLabelTablesRequest) SetProjectId(v string) *ListLabelTablesRequest {
	s.ProjectId = &v
	return s
}

func (s *ListLabelTablesRequest) SetSortBy(v string) *ListLabelTablesRequest {
	s.SortBy = &v
	return s
}

type ListLabelTablesShrinkRequest struct {
	LabelTableIdsShrink *string `json:"LabelTableIds,omitempty" xml:"LabelTableIds,omitempty"`
	Name                *string `json:"Name,omitempty" xml:"Name,omitempty"`
	Order               *string `json:"Order,omitempty" xml:"Order,omitempty"`
	Owner               *string `json:"Owner,omitempty" xml:"Owner,omitempty"`
	PageNumber          *int64  `json:"PageNumber,omitempty" xml:"PageNumber,omitempty"`
	PageSize            *int64  `json:"PageSize,omitempty" xml:"PageSize,omitempty"`
	ProjectId           *string `json:"ProjectId,omitempty" xml:"ProjectId,omitempty"`
	SortBy              *string `json:"SortBy,omitempty" xml:"SortBy,omitempty"`
}

func (s ListLabelTablesShrinkRequest) String() string {
	return tea.Prettify(s)
}

func (s ListLabelTablesShrinkRequest) GoString() string {
	return s.String()
}

func (s *ListLabelTablesShrinkRequest) SetLabelTableIdsShrink(v string) *ListLabelTablesShrinkRequest {
	s.LabelTableIdsShrink = &v
	return s
}

func (s *ListLabelTablesShrinkRequest) SetName(v string) *ListLabelTablesShrinkRequest {
	s.Name = &v
	return s
}

func (s *ListLabelTablesShrinkRequest) SetOrder(v string) *ListLabelTablesShrinkRequest {
	s.Order = &v
	return s
}

func (s *ListLabelTablesShrinkRequest) SetOwner(v string) *ListLabelTablesShrinkRequest {
	s.Owner = &v
	return s
}

func (s *ListLabelTablesShrinkRequest) SetPageNumber(v int64) *ListLabelTablesShrinkRequest {
	s.PageNumber = &v
	return s
}

func (s *ListLabelTablesShrinkRequest) SetPageSize(v int64) *ListLabelTablesShrinkRequest {
	s.PageSize = &v
	return s
}

func (s *ListLabelTablesShrinkRequest) SetProjectId(v string) *ListLabelTablesShrinkRequest {
	s.ProjectId = &v
	return s
}

func (s *ListLabelTablesShrinkRequest) SetSortBy(v string) *ListLabelTablesShrinkRequest {
	s.SortBy = &v
	return s
}

type ListLabelTablesResponseBody struct {
	LabelTables []*ListLabelTablesResponseBodyLabelTables `json:"LabelTables,omitempty" xml:"LabelTables,omitempty" type:"Repeated"`
	RequestId   *string                                   `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
	TotalCount  *int64                                    `json:"TotalCount,omitempty" xml:"TotalCount,omitempty"`
}

func (s ListLabelTablesResponseBody) String() string {
	return tea.Prettify(s)
}

func (s ListLabelTablesResponseBody) GoString() string {
	return s.String()
}

func (s *ListLabelTablesResponseBody) SetLabelTables(v []*ListLabelTablesResponseBodyLabelTables) *ListLabelTablesResponseBody {
	s.LabelTables = v
	return s
}

func (s *ListLabelTablesResponseBody) SetRequestId(v string) *ListLabelTablesResponseBody {
	s.RequestId = &v
	return s
}

func (s *ListLabelTablesResponseBody) SetTotalCount(v int64) *ListLabelTablesResponseBody {
	s.TotalCount = &v
	return s
}

type ListLabelTablesResponseBodyLabelTables struct {
	DatasourceId    *string `json:"DatasourceId,omitempty" xml:"DatasourceId,omitempty"`
	DatasourceName  *string `json:"DatasourceName,omitempty" xml:"DatasourceName,omitempty"`
	GmtCreateTime   *string `json:"GmtCreateTime,omitempty" xml:"GmtCreateTime,omitempty"`
	GmtModifiedTime *string `json:"GmtModifiedTime,omitempty" xml:"GmtModifiedTime,omitempty"`
	LabelTableId    *string `json:"LabelTableId,omitempty" xml:"LabelTableId,omitempty"`
	Name            *string `json:"Name,omitempty" xml:"Name,omitempty"`
	Owner           *string `json:"Owner,omitempty" xml:"Owner,omitempty"`
	ProjectId       *string `json:"ProjectId,omitempty" xml:"ProjectId,omitempty"`
	ProjectName     *string `json:"ProjectName,omitempty" xml:"ProjectName,omitempty"`
}

func (s ListLabelTablesResponseBodyLabelTables) String() string {
	return tea.Prettify(s)
}

func (s ListLabelTablesResponseBodyLabelTables) GoString() string {
	return s.String()
}

func (s *ListLabelTablesResponseBodyLabelTables) SetDatasourceId(v string) *ListLabelTablesResponseBodyLabelTables {
	s.DatasourceId = &v
	return s
}

func (s *ListLabelTablesResponseBodyLabelTables) SetDatasourceName(v string) *ListLabelTablesResponseBodyLabelTables {
	s.DatasourceName = &v
	return s
}

func (s *ListLabelTablesResponseBodyLabelTables) SetGmtCreateTime(v string) *ListLabelTablesResponseBodyLabelTables {
	s.GmtCreateTime = &v
	return s
}

func (s *ListLabelTablesResponseBodyLabelTables) SetGmtModifiedTime(v string) *ListLabelTablesResponseBodyLabelTables {
	s.GmtModifiedTime = &v
	return s
}

func (s *ListLabelTablesResponseBodyLabelTables) SetLabelTableId(v string) *ListLabelTablesResponseBodyLabelTables {
	s.LabelTableId = &v
	return s
}

func (s *ListLabelTablesResponseBodyLabelTables) SetName(v string) *ListLabelTablesResponseBodyLabelTables {
	s.Name = &v
	return s
}

func (s *ListLabelTablesResponseBodyLabelTables) SetOwner(v string) *ListLabelTablesResponseBodyLabelTables {
	s.Owner = &v
	return s
}

func (s *ListLabelTablesResponseBodyLabelTables) SetProjectId(v string) *ListLabelTablesResponseBodyLabelTables {
	s.ProjectId = &v
	return s
}

func (s *ListLabelTablesResponseBodyLabelTables) SetProjectName(v string) *ListLabelTablesResponseBodyLabelTables {
	s.ProjectName = &v
	return s
}

type ListLabelTablesResponse struct {
	Headers    map[string]*string           `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
	StatusCode *int32                       `json:"statusCode,omitempty" xml:"statusCode,omitempty" require:"true"`
	Body       *ListLabelTablesResponseBody `json:"body,omitempty" xml:"body,omitempty" require:"true"`
}

func (s ListLabelTablesResponse) String() string {
	return tea.Prettify(s)
}

func (s ListLabelTablesResponse) GoString() string {
	return s.String()
}

func (s *ListLabelTablesResponse) SetHeaders(v map[string]*string) *ListLabelTablesResponse {
	s.Headers = v
	return s
}

func (s *ListLabelTablesResponse) SetStatusCode(v int32) *ListLabelTablesResponse {
	s.StatusCode = &v
	return s
}

func (s *ListLabelTablesResponse) SetBody(v *ListLabelTablesResponseBody) *ListLabelTablesResponse {
	s.Body = v
	return s
}

type ListModelFeaturesRequest struct {
	ModelFeatureIds []*string `json:"ModelFeatureIds,omitempty" xml:"ModelFeatureIds,omitempty" type:"Repeated"`
	Name            *string   `json:"Name,omitempty" xml:"Name,omitempty"`
	Order           *string   `json:"Order,omitempty" xml:"Order,omitempty"`
	Owner           *string   `json:"Owner,omitempty" xml:"Owner,omitempty"`
	PageNumber      *string   `json:"PageNumber,omitempty" xml:"PageNumber,omitempty"`
	PageSize        *string   `json:"PageSize,omitempty" xml:"PageSize,omitempty"`
	ProjectId       *string   `json:"ProjectId,omitempty" xml:"ProjectId,omitempty"`
	SortBy          *string   `json:"SortBy,omitempty" xml:"SortBy,omitempty"`
}

func (s ListModelFeaturesRequest) String() string {
	return tea.Prettify(s)
}

func (s ListModelFeaturesRequest) GoString() string {
	return s.String()
}

func (s *ListModelFeaturesRequest) SetModelFeatureIds(v []*string) *ListModelFeaturesRequest {
	s.ModelFeatureIds = v
	return s
}

func (s *ListModelFeaturesRequest) SetName(v string) *ListModelFeaturesRequest {
	s.Name = &v
	return s
}

func (s *ListModelFeaturesRequest) SetOrder(v string) *ListModelFeaturesRequest {
	s.Order = &v
	return s
}

func (s *ListModelFeaturesRequest) SetOwner(v string) *ListModelFeaturesRequest {
	s.Owner = &v
	return s
}

func (s *ListModelFeaturesRequest) SetPageNumber(v string) *ListModelFeaturesRequest {
	s.PageNumber = &v
	return s
}

func (s *ListModelFeaturesRequest) SetPageSize(v string) *ListModelFeaturesRequest {
	s.PageSize = &v
	return s
}

func (s *ListModelFeaturesRequest) SetProjectId(v string) *ListModelFeaturesRequest {
	s.ProjectId = &v
	return s
}

func (s *ListModelFeaturesRequest) SetSortBy(v string) *ListModelFeaturesRequest {
	s.SortBy = &v
	return s
}

type ListModelFeaturesShrinkRequest struct {
	ModelFeatureIdsShrink *string `json:"ModelFeatureIds,omitempty" xml:"ModelFeatureIds,omitempty"`
	Name                  *string `json:"Name,omitempty" xml:"Name,omitempty"`
	Order                 *string `json:"Order,omitempty" xml:"Order,omitempty"`
	Owner                 *string `json:"Owner,omitempty" xml:"Owner,omitempty"`
	PageNumber            *string `json:"PageNumber,omitempty" xml:"PageNumber,omitempty"`
	PageSize              *string `json:"PageSize,omitempty" xml:"PageSize,omitempty"`
	ProjectId             *string `json:"ProjectId,omitempty" xml:"ProjectId,omitempty"`
	SortBy                *string `json:"SortBy,omitempty" xml:"SortBy,omitempty"`
}

func (s ListModelFeaturesShrinkRequest) String() string {
	return tea.Prettify(s)
}

func (s ListModelFeaturesShrinkRequest) GoString() string {
	return s.String()
}

func (s *ListModelFeaturesShrinkRequest) SetModelFeatureIdsShrink(v string) *ListModelFeaturesShrinkRequest {
	s.ModelFeatureIdsShrink = &v
	return s
}

func (s *ListModelFeaturesShrinkRequest) SetName(v string) *ListModelFeaturesShrinkRequest {
	s.Name = &v
	return s
}

func (s *ListModelFeaturesShrinkRequest) SetOrder(v string) *ListModelFeaturesShrinkRequest {
	s.Order = &v
	return s
}

func (s *ListModelFeaturesShrinkRequest) SetOwner(v string) *ListModelFeaturesShrinkRequest {
	s.Owner = &v
	return s
}

func (s *ListModelFeaturesShrinkRequest) SetPageNumber(v string) *ListModelFeaturesShrinkRequest {
	s.PageNumber = &v
	return s
}

func (s *ListModelFeaturesShrinkRequest) SetPageSize(v string) *ListModelFeaturesShrinkRequest {
	s.PageSize = &v
	return s
}

func (s *ListModelFeaturesShrinkRequest) SetProjectId(v string) *ListModelFeaturesShrinkRequest {
	s.ProjectId = &v
	return s
}

func (s *ListModelFeaturesShrinkRequest) SetSortBy(v string) *ListModelFeaturesShrinkRequest {
	s.SortBy = &v
	return s
}

type ListModelFeaturesResponseBody struct {
	ModelFeatures []*ListModelFeaturesResponseBodyModelFeatures `json:"ModelFeatures,omitempty" xml:"ModelFeatures,omitempty" type:"Repeated"`
	RequestId     *string                                       `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
	TotalCount    *int64                                        `json:"TotalCount,omitempty" xml:"TotalCount,omitempty"`
}

func (s ListModelFeaturesResponseBody) String() string {
	return tea.Prettify(s)
}

func (s ListModelFeaturesResponseBody) GoString() string {
	return s.String()
}

func (s *ListModelFeaturesResponseBody) SetModelFeatures(v []*ListModelFeaturesResponseBodyModelFeatures) *ListModelFeaturesResponseBody {
	s.ModelFeatures = v
	return s
}

func (s *ListModelFeaturesResponseBody) SetRequestId(v string) *ListModelFeaturesResponseBody {
	s.RequestId = &v
	return s
}

func (s *ListModelFeaturesResponseBody) SetTotalCount(v int64) *ListModelFeaturesResponseBody {
	s.TotalCount = &v
	return s
}

type ListModelFeaturesResponseBodyModelFeatures struct {
	GmtCreateTime   *string `json:"GmtCreateTime,omitempty" xml:"GmtCreateTime,omitempty"`
	GmtModifiedTime *string `json:"GmtModifiedTime,omitempty" xml:"GmtModifiedTime,omitempty"`
	LabelTableName  *string `json:"LabelTableName,omitempty" xml:"LabelTableName,omitempty"`
	ModelFeatureId  *string `json:"ModelFeatureId,omitempty" xml:"ModelFeatureId,omitempty"`
	Name            *string `json:"Name,omitempty" xml:"Name,omitempty"`
	Owner           *string `json:"Owner,omitempty" xml:"Owner,omitempty"`
	ProjectId       *string `json:"ProjectId,omitempty" xml:"ProjectId,omitempty"`
	ProjectName     *string `json:"ProjectName,omitempty" xml:"ProjectName,omitempty"`
}

func (s ListModelFeaturesResponseBodyModelFeatures) String() string {
	return tea.Prettify(s)
}

func (s ListModelFeaturesResponseBodyModelFeatures) GoString() string {
	return s.String()
}

func (s *ListModelFeaturesResponseBodyModelFeatures) SetGmtCreateTime(v string) *ListModelFeaturesResponseBodyModelFeatures {
	s.GmtCreateTime = &v
	return s
}

func (s *ListModelFeaturesResponseBodyModelFeatures) SetGmtModifiedTime(v string) *ListModelFeaturesResponseBodyModelFeatures {
	s.GmtModifiedTime = &v
	return s
}

func (s *ListModelFeaturesResponseBodyModelFeatures) SetLabelTableName(v string) *ListModelFeaturesResponseBodyModelFeatures {
	s.LabelTableName = &v
	return s
}

func (s *ListModelFeaturesResponseBodyModelFeatures) SetModelFeatureId(v string) *ListModelFeaturesResponseBodyModelFeatures {
	s.ModelFeatureId = &v
	return s
}

func (s *ListModelFeaturesResponseBodyModelFeatures) SetName(v string) *ListModelFeaturesResponseBodyModelFeatures {
	s.Name = &v
	return s
}

func (s *ListModelFeaturesResponseBodyModelFeatures) SetOwner(v string) *ListModelFeaturesResponseBodyModelFeatures {
	s.Owner = &v
	return s
}

func (s *ListModelFeaturesResponseBodyModelFeatures) SetProjectId(v string) *ListModelFeaturesResponseBodyModelFeatures {
	s.ProjectId = &v
	return s
}

func (s *ListModelFeaturesResponseBodyModelFeatures) SetProjectName(v string) *ListModelFeaturesResponseBodyModelFeatures {
	s.ProjectName = &v
	return s
}

type ListModelFeaturesResponse struct {
	Headers    map[string]*string             `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
	StatusCode *int32                         `json:"statusCode,omitempty" xml:"statusCode,omitempty" require:"true"`
	Body       *ListModelFeaturesResponseBody `json:"body,omitempty" xml:"body,omitempty" require:"true"`
}

func (s ListModelFeaturesResponse) String() string {
	return tea.Prettify(s)
}

func (s ListModelFeaturesResponse) GoString() string {
	return s.String()
}

func (s *ListModelFeaturesResponse) SetHeaders(v map[string]*string) *ListModelFeaturesResponse {
	s.Headers = v
	return s
}

func (s *ListModelFeaturesResponse) SetStatusCode(v int32) *ListModelFeaturesResponse {
	s.StatusCode = &v
	return s
}

func (s *ListModelFeaturesResponse) SetBody(v *ListModelFeaturesResponseBody) *ListModelFeaturesResponse {
	s.Body = v
	return s
}

type ListProjectFeatureViewOwnersResponseBody struct {
	Owners    []*string `json:"Owners,omitempty" xml:"Owners,omitempty" type:"Repeated"`
	RequestId *string   `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
}

func (s ListProjectFeatureViewOwnersResponseBody) String() string {
	return tea.Prettify(s)
}

func (s ListProjectFeatureViewOwnersResponseBody) GoString() string {
	return s.String()
}

func (s *ListProjectFeatureViewOwnersResponseBody) SetOwners(v []*string) *ListProjectFeatureViewOwnersResponseBody {
	s.Owners = v
	return s
}

func (s *ListProjectFeatureViewOwnersResponseBody) SetRequestId(v string) *ListProjectFeatureViewOwnersResponseBody {
	s.RequestId = &v
	return s
}

type ListProjectFeatureViewOwnersResponse struct {
	Headers    map[string]*string                        `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
	StatusCode *int32                                    `json:"statusCode,omitempty" xml:"statusCode,omitempty" require:"true"`
	Body       *ListProjectFeatureViewOwnersResponseBody `json:"body,omitempty" xml:"body,omitempty" require:"true"`
}

func (s ListProjectFeatureViewOwnersResponse) String() string {
	return tea.Prettify(s)
}

func (s ListProjectFeatureViewOwnersResponse) GoString() string {
	return s.String()
}

func (s *ListProjectFeatureViewOwnersResponse) SetHeaders(v map[string]*string) *ListProjectFeatureViewOwnersResponse {
	s.Headers = v
	return s
}

func (s *ListProjectFeatureViewOwnersResponse) SetStatusCode(v int32) *ListProjectFeatureViewOwnersResponse {
	s.StatusCode = &v
	return s
}

func (s *ListProjectFeatureViewOwnersResponse) SetBody(v *ListProjectFeatureViewOwnersResponseBody) *ListProjectFeatureViewOwnersResponse {
	s.Body = v
	return s
}

type ListProjectFeatureViewTagsResponseBody struct {
	RequestId *string   `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
	Tags      []*string `json:"Tags,omitempty" xml:"Tags,omitempty" type:"Repeated"`
}

func (s ListProjectFeatureViewTagsResponseBody) String() string {
	return tea.Prettify(s)
}

func (s ListProjectFeatureViewTagsResponseBody) GoString() string {
	return s.String()
}

func (s *ListProjectFeatureViewTagsResponseBody) SetRequestId(v string) *ListProjectFeatureViewTagsResponseBody {
	s.RequestId = &v
	return s
}

func (s *ListProjectFeatureViewTagsResponseBody) SetTags(v []*string) *ListProjectFeatureViewTagsResponseBody {
	s.Tags = v
	return s
}

type ListProjectFeatureViewTagsResponse struct {
	Headers    map[string]*string                      `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
	StatusCode *int32                                  `json:"statusCode,omitempty" xml:"statusCode,omitempty" require:"true"`
	Body       *ListProjectFeatureViewTagsResponseBody `json:"body,omitempty" xml:"body,omitempty" require:"true"`
}

func (s ListProjectFeatureViewTagsResponse) String() string {
	return tea.Prettify(s)
}

func (s ListProjectFeatureViewTagsResponse) GoString() string {
	return s.String()
}

func (s *ListProjectFeatureViewTagsResponse) SetHeaders(v map[string]*string) *ListProjectFeatureViewTagsResponse {
	s.Headers = v
	return s
}

func (s *ListProjectFeatureViewTagsResponse) SetStatusCode(v int32) *ListProjectFeatureViewTagsResponse {
	s.StatusCode = &v
	return s
}

func (s *ListProjectFeatureViewTagsResponse) SetBody(v *ListProjectFeatureViewTagsResponseBody) *ListProjectFeatureViewTagsResponse {
	s.Body = v
	return s
}

type ListProjectFeatureViewsResponseBody struct {
	FeatureViews []*ListProjectFeatureViewsResponseBodyFeatureViews `json:"FeatureViews,omitempty" xml:"FeatureViews,omitempty" type:"Repeated"`
	RequestId    *string                                            `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
	TotalCount   *int64                                             `json:"TotalCount,omitempty" xml:"TotalCount,omitempty"`
}

func (s ListProjectFeatureViewsResponseBody) String() string {
	return tea.Prettify(s)
}

func (s ListProjectFeatureViewsResponseBody) GoString() string {
	return s.String()
}

func (s *ListProjectFeatureViewsResponseBody) SetFeatureViews(v []*ListProjectFeatureViewsResponseBodyFeatureViews) *ListProjectFeatureViewsResponseBody {
	s.FeatureViews = v
	return s
}

func (s *ListProjectFeatureViewsResponseBody) SetRequestId(v string) *ListProjectFeatureViewsResponseBody {
	s.RequestId = &v
	return s
}

func (s *ListProjectFeatureViewsResponseBody) SetTotalCount(v int64) *ListProjectFeatureViewsResponseBody {
	s.TotalCount = &v
	return s
}

type ListProjectFeatureViewsResponseBodyFeatureViews struct {
	FeatureViewId *string                                                    `json:"FeatureViewId,omitempty" xml:"FeatureViewId,omitempty"`
	Features      []*ListProjectFeatureViewsResponseBodyFeatureViewsFeatures `json:"Features,omitempty" xml:"Features,omitempty" type:"Repeated"`
	Name          *string                                                    `json:"Name,omitempty" xml:"Name,omitempty"`
}

func (s ListProjectFeatureViewsResponseBodyFeatureViews) String() string {
	return tea.Prettify(s)
}

func (s ListProjectFeatureViewsResponseBodyFeatureViews) GoString() string {
	return s.String()
}

func (s *ListProjectFeatureViewsResponseBodyFeatureViews) SetFeatureViewId(v string) *ListProjectFeatureViewsResponseBodyFeatureViews {
	s.FeatureViewId = &v
	return s
}

func (s *ListProjectFeatureViewsResponseBodyFeatureViews) SetFeatures(v []*ListProjectFeatureViewsResponseBodyFeatureViewsFeatures) *ListProjectFeatureViewsResponseBodyFeatureViews {
	s.Features = v
	return s
}

func (s *ListProjectFeatureViewsResponseBodyFeatureViews) SetName(v string) *ListProjectFeatureViewsResponseBodyFeatureViews {
	s.Name = &v
	return s
}

type ListProjectFeatureViewsResponseBodyFeatureViewsFeatures struct {
	Attributes []*string `json:"Attributes,omitempty" xml:"Attributes,omitempty" type:"Repeated"`
	Name       *string   `json:"Name,omitempty" xml:"Name,omitempty"`
	Type       *string   `json:"Type,omitempty" xml:"Type,omitempty"`
}

func (s ListProjectFeatureViewsResponseBodyFeatureViewsFeatures) String() string {
	return tea.Prettify(s)
}

func (s ListProjectFeatureViewsResponseBodyFeatureViewsFeatures) GoString() string {
	return s.String()
}

func (s *ListProjectFeatureViewsResponseBodyFeatureViewsFeatures) SetAttributes(v []*string) *ListProjectFeatureViewsResponseBodyFeatureViewsFeatures {
	s.Attributes = v
	return s
}

func (s *ListProjectFeatureViewsResponseBodyFeatureViewsFeatures) SetName(v string) *ListProjectFeatureViewsResponseBodyFeatureViewsFeatures {
	s.Name = &v
	return s
}

func (s *ListProjectFeatureViewsResponseBodyFeatureViewsFeatures) SetType(v string) *ListProjectFeatureViewsResponseBodyFeatureViewsFeatures {
	s.Type = &v
	return s
}

type ListProjectFeatureViewsResponse struct {
	Headers    map[string]*string                   `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
	StatusCode *int32                               `json:"statusCode,omitempty" xml:"statusCode,omitempty" require:"true"`
	Body       *ListProjectFeatureViewsResponseBody `json:"body,omitempty" xml:"body,omitempty" require:"true"`
}

func (s ListProjectFeatureViewsResponse) String() string {
	return tea.Prettify(s)
}

func (s ListProjectFeatureViewsResponse) GoString() string {
	return s.String()
}

func (s *ListProjectFeatureViewsResponse) SetHeaders(v map[string]*string) *ListProjectFeatureViewsResponse {
	s.Headers = v
	return s
}

func (s *ListProjectFeatureViewsResponse) SetStatusCode(v int32) *ListProjectFeatureViewsResponse {
	s.StatusCode = &v
	return s
}

func (s *ListProjectFeatureViewsResponse) SetBody(v *ListProjectFeatureViewsResponseBody) *ListProjectFeatureViewsResponse {
	s.Body = v
	return s
}

type ListProjectsRequest struct {
	Name        *string   `json:"Name,omitempty" xml:"Name,omitempty"`
	Order       *string   `json:"Order,omitempty" xml:"Order,omitempty"`
	Owner       *string   `json:"Owner,omitempty" xml:"Owner,omitempty"`
	PageNumber  *int32    `json:"PageNumber,omitempty" xml:"PageNumber,omitempty"`
	PageSize    *int32    `json:"PageSize,omitempty" xml:"PageSize,omitempty"`
	ProjectIds  []*string `json:"ProjectIds,omitempty" xml:"ProjectIds,omitempty" type:"Repeated"`
	SortBy      *string   `json:"SortBy,omitempty" xml:"SortBy,omitempty"`
	WorkspaceId *string   `json:"WorkspaceId,omitempty" xml:"WorkspaceId,omitempty"`
}

func (s ListProjectsRequest) String() string {
	return tea.Prettify(s)
}

func (s ListProjectsRequest) GoString() string {
	return s.String()
}

func (s *ListProjectsRequest) SetName(v string) *ListProjectsRequest {
	s.Name = &v
	return s
}

func (s *ListProjectsRequest) SetOrder(v string) *ListProjectsRequest {
	s.Order = &v
	return s
}

func (s *ListProjectsRequest) SetOwner(v string) *ListProjectsRequest {
	s.Owner = &v
	return s
}

func (s *ListProjectsRequest) SetPageNumber(v int32) *ListProjectsRequest {
	s.PageNumber = &v
	return s
}

func (s *ListProjectsRequest) SetPageSize(v int32) *ListProjectsRequest {
	s.PageSize = &v
	return s
}

func (s *ListProjectsRequest) SetProjectIds(v []*string) *ListProjectsRequest {
	s.ProjectIds = v
	return s
}

func (s *ListProjectsRequest) SetSortBy(v string) *ListProjectsRequest {
	s.SortBy = &v
	return s
}

func (s *ListProjectsRequest) SetWorkspaceId(v string) *ListProjectsRequest {
	s.WorkspaceId = &v
	return s
}

type ListProjectsShrinkRequest struct {
	Name             *string `json:"Name,omitempty" xml:"Name,omitempty"`
	Order            *string `json:"Order,omitempty" xml:"Order,omitempty"`
	Owner            *string `json:"Owner,omitempty" xml:"Owner,omitempty"`
	PageNumber       *int32  `json:"PageNumber,omitempty" xml:"PageNumber,omitempty"`
	PageSize         *int32  `json:"PageSize,omitempty" xml:"PageSize,omitempty"`
	ProjectIdsShrink *string `json:"ProjectIds,omitempty" xml:"ProjectIds,omitempty"`
	SortBy           *string `json:"SortBy,omitempty" xml:"SortBy,omitempty"`
	WorkspaceId      *string `json:"WorkspaceId,omitempty" xml:"WorkspaceId,omitempty"`
}

func (s ListProjectsShrinkRequest) String() string {
	return tea.Prettify(s)
}

func (s ListProjectsShrinkRequest) GoString() string {
	return s.String()
}

func (s *ListProjectsShrinkRequest) SetName(v string) *ListProjectsShrinkRequest {
	s.Name = &v
	return s
}

func (s *ListProjectsShrinkRequest) SetOrder(v string) *ListProjectsShrinkRequest {
	s.Order = &v
	return s
}

func (s *ListProjectsShrinkRequest) SetOwner(v string) *ListProjectsShrinkRequest {
	s.Owner = &v
	return s
}

func (s *ListProjectsShrinkRequest) SetPageNumber(v int32) *ListProjectsShrinkRequest {
	s.PageNumber = &v
	return s
}

func (s *ListProjectsShrinkRequest) SetPageSize(v int32) *ListProjectsShrinkRequest {
	s.PageSize = &v
	return s
}

func (s *ListProjectsShrinkRequest) SetProjectIdsShrink(v string) *ListProjectsShrinkRequest {
	s.ProjectIdsShrink = &v
	return s
}

func (s *ListProjectsShrinkRequest) SetSortBy(v string) *ListProjectsShrinkRequest {
	s.SortBy = &v
	return s
}

func (s *ListProjectsShrinkRequest) SetWorkspaceId(v string) *ListProjectsShrinkRequest {
	s.WorkspaceId = &v
	return s
}

type ListProjectsResponseBody struct {
	Projects   []*ListProjectsResponseBodyProjects `json:"Projects,omitempty" xml:"Projects,omitempty" type:"Repeated"`
	RequestId  *string                             `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
	TotalCount *int64                              `json:"TotalCount,omitempty" xml:"TotalCount,omitempty"`
}

func (s ListProjectsResponseBody) String() string {
	return tea.Prettify(s)
}

func (s ListProjectsResponseBody) GoString() string {
	return s.String()
}

func (s *ListProjectsResponseBody) SetProjects(v []*ListProjectsResponseBodyProjects) *ListProjectsResponseBody {
	s.Projects = v
	return s
}

func (s *ListProjectsResponseBody) SetRequestId(v string) *ListProjectsResponseBody {
	s.RequestId = &v
	return s
}

func (s *ListProjectsResponseBody) SetTotalCount(v int64) *ListProjectsResponseBody {
	s.TotalCount = &v
	return s
}

type ListProjectsResponseBodyProjects struct {
	Description           *string `json:"Description,omitempty" xml:"Description,omitempty"`
	FeatureEntityCount    *int32  `json:"FeatureEntityCount,omitempty" xml:"FeatureEntityCount,omitempty"`
	FeatureViewCount      *int32  `json:"FeatureViewCount,omitempty" xml:"FeatureViewCount,omitempty"`
	GmtCreateTime         *string `json:"GmtCreateTime,omitempty" xml:"GmtCreateTime,omitempty"`
	GmtModifiedTime       *string `json:"GmtModifiedTime,omitempty" xml:"GmtModifiedTime,omitempty"`
	ModelCount            *int32  `json:"ModelCount,omitempty" xml:"ModelCount,omitempty"`
	Name                  *string `json:"Name,omitempty" xml:"Name,omitempty"`
	OfflineDatasourceId   *string `json:"OfflineDatasourceId,omitempty" xml:"OfflineDatasourceId,omitempty"`
	OfflineDatasourceName *string `json:"OfflineDatasourceName,omitempty" xml:"OfflineDatasourceName,omitempty"`
	OfflineDatasourceType *string `json:"OfflineDatasourceType,omitempty" xml:"OfflineDatasourceType,omitempty"`
	OfflineLifecycle      *int32  `json:"OfflineLifecycle,omitempty" xml:"OfflineLifecycle,omitempty"`
	OnlineDatasourceId    *string `json:"OnlineDatasourceId,omitempty" xml:"OnlineDatasourceId,omitempty"`
	OnlineDatasourceName  *string `json:"OnlineDatasourceName,omitempty" xml:"OnlineDatasourceName,omitempty"`
	OnlineDatasourceType  *string `json:"OnlineDatasourceType,omitempty" xml:"OnlineDatasourceType,omitempty"`
	Owner                 *string `json:"Owner,omitempty" xml:"Owner,omitempty"`
	ProjectId             *string `json:"ProjectId,omitempty" xml:"ProjectId,omitempty"`
}

func (s ListProjectsResponseBodyProjects) String() string {
	return tea.Prettify(s)
}

func (s ListProjectsResponseBodyProjects) GoString() string {
	return s.String()
}

func (s *ListProjectsResponseBodyProjects) SetDescription(v string) *ListProjectsResponseBodyProjects {
	s.Description = &v
	return s
}

func (s *ListProjectsResponseBodyProjects) SetFeatureEntityCount(v int32) *ListProjectsResponseBodyProjects {
	s.FeatureEntityCount = &v
	return s
}

func (s *ListProjectsResponseBodyProjects) SetFeatureViewCount(v int32) *ListProjectsResponseBodyProjects {
	s.FeatureViewCount = &v
	return s
}

func (s *ListProjectsResponseBodyProjects) SetGmtCreateTime(v string) *ListProjectsResponseBodyProjects {
	s.GmtCreateTime = &v
	return s
}

func (s *ListProjectsResponseBodyProjects) SetGmtModifiedTime(v string) *ListProjectsResponseBodyProjects {
	s.GmtModifiedTime = &v
	return s
}

func (s *ListProjectsResponseBodyProjects) SetModelCount(v int32) *ListProjectsResponseBodyProjects {
	s.ModelCount = &v
	return s
}

func (s *ListProjectsResponseBodyProjects) SetName(v string) *ListProjectsResponseBodyProjects {
	s.Name = &v
	return s
}

func (s *ListProjectsResponseBodyProjects) SetOfflineDatasourceId(v string) *ListProjectsResponseBodyProjects {
	s.OfflineDatasourceId = &v
	return s
}

func (s *ListProjectsResponseBodyProjects) SetOfflineDatasourceName(v string) *ListProjectsResponseBodyProjects {
	s.OfflineDatasourceName = &v
	return s
}

func (s *ListProjectsResponseBodyProjects) SetOfflineDatasourceType(v string) *ListProjectsResponseBodyProjects {
	s.OfflineDatasourceType = &v
	return s
}

func (s *ListProjectsResponseBodyProjects) SetOfflineLifecycle(v int32) *ListProjectsResponseBodyProjects {
	s.OfflineLifecycle = &v
	return s
}

func (s *ListProjectsResponseBodyProjects) SetOnlineDatasourceId(v string) *ListProjectsResponseBodyProjects {
	s.OnlineDatasourceId = &v
	return s
}

func (s *ListProjectsResponseBodyProjects) SetOnlineDatasourceName(v string) *ListProjectsResponseBodyProjects {
	s.OnlineDatasourceName = &v
	return s
}

func (s *ListProjectsResponseBodyProjects) SetOnlineDatasourceType(v string) *ListProjectsResponseBodyProjects {
	s.OnlineDatasourceType = &v
	return s
}

func (s *ListProjectsResponseBodyProjects) SetOwner(v string) *ListProjectsResponseBodyProjects {
	s.Owner = &v
	return s
}

func (s *ListProjectsResponseBodyProjects) SetProjectId(v string) *ListProjectsResponseBodyProjects {
	s.ProjectId = &v
	return s
}

type ListProjectsResponse struct {
	Headers    map[string]*string        `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
	StatusCode *int32                    `json:"statusCode,omitempty" xml:"statusCode,omitempty" require:"true"`
	Body       *ListProjectsResponseBody `json:"body,omitempty" xml:"body,omitempty" require:"true"`
}

func (s ListProjectsResponse) String() string {
	return tea.Prettify(s)
}

func (s ListProjectsResponse) GoString() string {
	return s.String()
}

func (s *ListProjectsResponse) SetHeaders(v map[string]*string) *ListProjectsResponse {
	s.Headers = v
	return s
}

func (s *ListProjectsResponse) SetStatusCode(v int32) *ListProjectsResponse {
	s.StatusCode = &v
	return s
}

func (s *ListProjectsResponse) SetBody(v *ListProjectsResponseBody) *ListProjectsResponse {
	s.Body = v
	return s
}

type ListTaskLogsRequest struct {
	PageNumber *int32 `json:"PageNumber,omitempty" xml:"PageNumber,omitempty"`
	PageSize   *int32 `json:"PageSize,omitempty" xml:"PageSize,omitempty"`
}

func (s ListTaskLogsRequest) String() string {
	return tea.Prettify(s)
}

func (s ListTaskLogsRequest) GoString() string {
	return s.String()
}

func (s *ListTaskLogsRequest) SetPageNumber(v int32) *ListTaskLogsRequest {
	s.PageNumber = &v
	return s
}

func (s *ListTaskLogsRequest) SetPageSize(v int32) *ListTaskLogsRequest {
	s.PageSize = &v
	return s
}

type ListTaskLogsResponseBody struct {
	Logs       []*string `json:"Logs,omitempty" xml:"Logs,omitempty" type:"Repeated"`
	RequestId  *string   `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
	TotalCount *int32    `json:"TotalCount,omitempty" xml:"TotalCount,omitempty"`
}

func (s ListTaskLogsResponseBody) String() string {
	return tea.Prettify(s)
}

func (s ListTaskLogsResponseBody) GoString() string {
	return s.String()
}

func (s *ListTaskLogsResponseBody) SetLogs(v []*string) *ListTaskLogsResponseBody {
	s.Logs = v
	return s
}

func (s *ListTaskLogsResponseBody) SetRequestId(v string) *ListTaskLogsResponseBody {
	s.RequestId = &v
	return s
}

func (s *ListTaskLogsResponseBody) SetTotalCount(v int32) *ListTaskLogsResponseBody {
	s.TotalCount = &v
	return s
}

type ListTaskLogsResponse struct {
	Headers    map[string]*string        `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
	StatusCode *int32                    `json:"statusCode,omitempty" xml:"statusCode,omitempty" require:"true"`
	Body       *ListTaskLogsResponseBody `json:"body,omitempty" xml:"body,omitempty" require:"true"`
}

func (s ListTaskLogsResponse) String() string {
	return tea.Prettify(s)
}

func (s ListTaskLogsResponse) GoString() string {
	return s.String()
}

func (s *ListTaskLogsResponse) SetHeaders(v map[string]*string) *ListTaskLogsResponse {
	s.Headers = v
	return s
}

func (s *ListTaskLogsResponse) SetStatusCode(v int32) *ListTaskLogsResponse {
	s.StatusCode = &v
	return s
}

func (s *ListTaskLogsResponse) SetBody(v *ListTaskLogsResponseBody) *ListTaskLogsResponse {
	s.Body = v
	return s
}

type ListTasksRequest struct {
	ObjectId   *string   `json:"ObjectId,omitempty" xml:"ObjectId,omitempty"`
	ObjectType *string   `json:"ObjectType,omitempty" xml:"ObjectType,omitempty"`
	PageNumber *int32    `json:"PageNumber,omitempty" xml:"PageNumber,omitempty"`
	PageSize   *int32    `json:"PageSize,omitempty" xml:"PageSize,omitempty"`
	ProjectId  *string   `json:"ProjectId,omitempty" xml:"ProjectId,omitempty"`
	Status     *string   `json:"Status,omitempty" xml:"Status,omitempty"`
	TaskIds    []*string `json:"TaskIds,omitempty" xml:"TaskIds,omitempty" type:"Repeated"`
	Type       *string   `json:"Type,omitempty" xml:"Type,omitempty"`
}

func (s ListTasksRequest) String() string {
	return tea.Prettify(s)
}

func (s ListTasksRequest) GoString() string {
	return s.String()
}

func (s *ListTasksRequest) SetObjectId(v string) *ListTasksRequest {
	s.ObjectId = &v
	return s
}

func (s *ListTasksRequest) SetObjectType(v string) *ListTasksRequest {
	s.ObjectType = &v
	return s
}

func (s *ListTasksRequest) SetPageNumber(v int32) *ListTasksRequest {
	s.PageNumber = &v
	return s
}

func (s *ListTasksRequest) SetPageSize(v int32) *ListTasksRequest {
	s.PageSize = &v
	return s
}

func (s *ListTasksRequest) SetProjectId(v string) *ListTasksRequest {
	s.ProjectId = &v
	return s
}

func (s *ListTasksRequest) SetStatus(v string) *ListTasksRequest {
	s.Status = &v
	return s
}

func (s *ListTasksRequest) SetTaskIds(v []*string) *ListTasksRequest {
	s.TaskIds = v
	return s
}

func (s *ListTasksRequest) SetType(v string) *ListTasksRequest {
	s.Type = &v
	return s
}

type ListTasksShrinkRequest struct {
	ObjectId      *string `json:"ObjectId,omitempty" xml:"ObjectId,omitempty"`
	ObjectType    *string `json:"ObjectType,omitempty" xml:"ObjectType,omitempty"`
	PageNumber    *int32  `json:"PageNumber,omitempty" xml:"PageNumber,omitempty"`
	PageSize      *int32  `json:"PageSize,omitempty" xml:"PageSize,omitempty"`
	ProjectId     *string `json:"ProjectId,omitempty" xml:"ProjectId,omitempty"`
	Status        *string `json:"Status,omitempty" xml:"Status,omitempty"`
	TaskIdsShrink *string `json:"TaskIds,omitempty" xml:"TaskIds,omitempty"`
	Type          *string `json:"Type,omitempty" xml:"Type,omitempty"`
}

func (s ListTasksShrinkRequest) String() string {
	return tea.Prettify(s)
}

func (s ListTasksShrinkRequest) GoString() string {
	return s.String()
}

func (s *ListTasksShrinkRequest) SetObjectId(v string) *ListTasksShrinkRequest {
	s.ObjectId = &v
	return s
}

func (s *ListTasksShrinkRequest) SetObjectType(v string) *ListTasksShrinkRequest {
	s.ObjectType = &v
	return s
}

func (s *ListTasksShrinkRequest) SetPageNumber(v int32) *ListTasksShrinkRequest {
	s.PageNumber = &v
	return s
}

func (s *ListTasksShrinkRequest) SetPageSize(v int32) *ListTasksShrinkRequest {
	s.PageSize = &v
	return s
}

func (s *ListTasksShrinkRequest) SetProjectId(v string) *ListTasksShrinkRequest {
	s.ProjectId = &v
	return s
}

func (s *ListTasksShrinkRequest) SetStatus(v string) *ListTasksShrinkRequest {
	s.Status = &v
	return s
}

func (s *ListTasksShrinkRequest) SetTaskIdsShrink(v string) *ListTasksShrinkRequest {
	s.TaskIdsShrink = &v
	return s
}

func (s *ListTasksShrinkRequest) SetType(v string) *ListTasksShrinkRequest {
	s.Type = &v
	return s
}

type ListTasksResponseBody struct {
	RequestId  *string                       `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
	Tasks      []*ListTasksResponseBodyTasks `json:"Tasks,omitempty" xml:"Tasks,omitempty" type:"Repeated"`
	TotalCount *int32                        `json:"TotalCount,omitempty" xml:"TotalCount,omitempty"`
}

func (s ListTasksResponseBody) String() string {
	return tea.Prettify(s)
}

func (s ListTasksResponseBody) GoString() string {
	return s.String()
}

func (s *ListTasksResponseBody) SetRequestId(v string) *ListTasksResponseBody {
	s.RequestId = &v
	return s
}

func (s *ListTasksResponseBody) SetTasks(v []*ListTasksResponseBodyTasks) *ListTasksResponseBody {
	s.Tasks = v
	return s
}

func (s *ListTasksResponseBody) SetTotalCount(v int32) *ListTasksResponseBody {
	s.TotalCount = &v
	return s
}

type ListTasksResponseBodyTasks struct {
	GmtCreateTime   *string `json:"GmtCreateTime,omitempty" xml:"GmtCreateTime,omitempty"`
	GmtExecutedTime *string `json:"GmtExecutedTime,omitempty" xml:"GmtExecutedTime,omitempty"`
	GmtFinishedTime *string `json:"GmtFinishedTime,omitempty" xml:"GmtFinishedTime,omitempty"`
	ObjectId        *string `json:"ObjectId,omitempty" xml:"ObjectId,omitempty"`
	ObjectType      *string `json:"ObjectType,omitempty" xml:"ObjectType,omitempty"`
	ProjectId       *string `json:"ProjectId,omitempty" xml:"ProjectId,omitempty"`
	ProjectName     *string `json:"ProjectName,omitempty" xml:"ProjectName,omitempty"`
	Status          *string `json:"Status,omitempty" xml:"Status,omitempty"`
	TaskId          *string `json:"TaskId,omitempty" xml:"TaskId,omitempty"`
	Type            *string `json:"Type,omitempty" xml:"Type,omitempty"`
}

func (s ListTasksResponseBodyTasks) String() string {
	return tea.Prettify(s)
}

func (s ListTasksResponseBodyTasks) GoString() string {
	return s.String()
}

func (s *ListTasksResponseBodyTasks) SetGmtCreateTime(v string) *ListTasksResponseBodyTasks {
	s.GmtCreateTime = &v
	return s
}

func (s *ListTasksResponseBodyTasks) SetGmtExecutedTime(v string) *ListTasksResponseBodyTasks {
	s.GmtExecutedTime = &v
	return s
}

func (s *ListTasksResponseBodyTasks) SetGmtFinishedTime(v string) *ListTasksResponseBodyTasks {
	s.GmtFinishedTime = &v
	return s
}

func (s *ListTasksResponseBodyTasks) SetObjectId(v string) *ListTasksResponseBodyTasks {
	s.ObjectId = &v
	return s
}

func (s *ListTasksResponseBodyTasks) SetObjectType(v string) *ListTasksResponseBodyTasks {
	s.ObjectType = &v
	return s
}

func (s *ListTasksResponseBodyTasks) SetProjectId(v string) *ListTasksResponseBodyTasks {
	s.ProjectId = &v
	return s
}

func (s *ListTasksResponseBodyTasks) SetProjectName(v string) *ListTasksResponseBodyTasks {
	s.ProjectName = &v
	return s
}

func (s *ListTasksResponseBodyTasks) SetStatus(v string) *ListTasksResponseBodyTasks {
	s.Status = &v
	return s
}

func (s *ListTasksResponseBodyTasks) SetTaskId(v string) *ListTasksResponseBodyTasks {
	s.TaskId = &v
	return s
}

func (s *ListTasksResponseBodyTasks) SetType(v string) *ListTasksResponseBodyTasks {
	s.Type = &v
	return s
}

type ListTasksResponse struct {
	Headers    map[string]*string     `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
	StatusCode *int32                 `json:"statusCode,omitempty" xml:"statusCode,omitempty" require:"true"`
	Body       *ListTasksResponseBody `json:"body,omitempty" xml:"body,omitempty" require:"true"`
}

func (s ListTasksResponse) String() string {
	return tea.Prettify(s)
}

func (s ListTasksResponse) GoString() string {
	return s.String()
}

func (s *ListTasksResponse) SetHeaders(v map[string]*string) *ListTasksResponse {
	s.Headers = v
	return s
}

func (s *ListTasksResponse) SetStatusCode(v int32) *ListTasksResponse {
	s.StatusCode = &v
	return s
}

func (s *ListTasksResponse) SetBody(v *ListTasksResponseBody) *ListTasksResponse {
	s.Body = v
	return s
}

type PublishFeatureViewTableRequest struct {
	Config          *string                           `json:"Config,omitempty" xml:"Config,omitempty"`
	EventTime       *string                           `json:"EventTime,omitempty" xml:"EventTime,omitempty"`
	Mode            *string                           `json:"Mode,omitempty" xml:"Mode,omitempty"`
	OfflineToOnline *bool                             `json:"OfflineToOnline,omitempty" xml:"OfflineToOnline,omitempty"`
	Partitions      map[string]map[string]interface{} `json:"Partitions,omitempty" xml:"Partitions,omitempty"`
}

func (s PublishFeatureViewTableRequest) String() string {
	return tea.Prettify(s)
}

func (s PublishFeatureViewTableRequest) GoString() string {
	return s.String()
}

func (s *PublishFeatureViewTableRequest) SetConfig(v string) *PublishFeatureViewTableRequest {
	s.Config = &v
	return s
}

func (s *PublishFeatureViewTableRequest) SetEventTime(v string) *PublishFeatureViewTableRequest {
	s.EventTime = &v
	return s
}

func (s *PublishFeatureViewTableRequest) SetMode(v string) *PublishFeatureViewTableRequest {
	s.Mode = &v
	return s
}

func (s *PublishFeatureViewTableRequest) SetOfflineToOnline(v bool) *PublishFeatureViewTableRequest {
	s.OfflineToOnline = &v
	return s
}

func (s *PublishFeatureViewTableRequest) SetPartitions(v map[string]map[string]interface{}) *PublishFeatureViewTableRequest {
	s.Partitions = v
	return s
}

type PublishFeatureViewTableResponseBody struct {
	RequestId *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
}

func (s PublishFeatureViewTableResponseBody) String() string {
	return tea.Prettify(s)
}

func (s PublishFeatureViewTableResponseBody) GoString() string {
	return s.String()
}

func (s *PublishFeatureViewTableResponseBody) SetRequestId(v string) *PublishFeatureViewTableResponseBody {
	s.RequestId = &v
	return s
}

type PublishFeatureViewTableResponse struct {
	Headers    map[string]*string                   `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
	StatusCode *int32                               `json:"statusCode,omitempty" xml:"statusCode,omitempty" require:"true"`
	Body       *PublishFeatureViewTableResponseBody `json:"body,omitempty" xml:"body,omitempty" require:"true"`
}

func (s PublishFeatureViewTableResponse) String() string {
	return tea.Prettify(s)
}

func (s PublishFeatureViewTableResponse) GoString() string {
	return s.String()
}

func (s *PublishFeatureViewTableResponse) SetHeaders(v map[string]*string) *PublishFeatureViewTableResponse {
	s.Headers = v
	return s
}

func (s *PublishFeatureViewTableResponse) SetStatusCode(v int32) *PublishFeatureViewTableResponse {
	s.StatusCode = &v
	return s
}

func (s *PublishFeatureViewTableResponse) SetBody(v *PublishFeatureViewTableResponseBody) *PublishFeatureViewTableResponse {
	s.Body = v
	return s
}

type UpdateDatasourceRequest struct {
	Config *string `json:"Config,omitempty" xml:"Config,omitempty"`
	Name   *string `json:"Name,omitempty" xml:"Name,omitempty"`
	Uri    *string `json:"Uri,omitempty" xml:"Uri,omitempty"`
}

func (s UpdateDatasourceRequest) String() string {
	return tea.Prettify(s)
}

func (s UpdateDatasourceRequest) GoString() string {
	return s.String()
}

func (s *UpdateDatasourceRequest) SetConfig(v string) *UpdateDatasourceRequest {
	s.Config = &v
	return s
}

func (s *UpdateDatasourceRequest) SetName(v string) *UpdateDatasourceRequest {
	s.Name = &v
	return s
}

func (s *UpdateDatasourceRequest) SetUri(v string) *UpdateDatasourceRequest {
	s.Uri = &v
	return s
}

type UpdateDatasourceResponseBody struct {
	RequestId *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
}

func (s UpdateDatasourceResponseBody) String() string {
	return tea.Prettify(s)
}

func (s UpdateDatasourceResponseBody) GoString() string {
	return s.String()
}

func (s *UpdateDatasourceResponseBody) SetRequestId(v string) *UpdateDatasourceResponseBody {
	s.RequestId = &v
	return s
}

type UpdateDatasourceResponse struct {
	Headers    map[string]*string            `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
	StatusCode *int32                        `json:"statusCode,omitempty" xml:"statusCode,omitempty" require:"true"`
	Body       *UpdateDatasourceResponseBody `json:"body,omitempty" xml:"body,omitempty" require:"true"`
}

func (s UpdateDatasourceResponse) String() string {
	return tea.Prettify(s)
}

func (s UpdateDatasourceResponse) GoString() string {
	return s.String()
}

func (s *UpdateDatasourceResponse) SetHeaders(v map[string]*string) *UpdateDatasourceResponse {
	s.Headers = v
	return s
}

func (s *UpdateDatasourceResponse) SetStatusCode(v int32) *UpdateDatasourceResponse {
	s.StatusCode = &v
	return s
}

func (s *UpdateDatasourceResponse) SetBody(v *UpdateDatasourceResponseBody) *UpdateDatasourceResponse {
	s.Body = v
	return s
}

type UpdateLabelTableRequest struct {
	DatasourceId *string                          `json:"DatasourceId,omitempty" xml:"DatasourceId,omitempty"`
	Fields       []*UpdateLabelTableRequestFields `json:"Fields,omitempty" xml:"Fields,omitempty" type:"Repeated"`
	Name         *string                          `json:"Name,omitempty" xml:"Name,omitempty"`
}

func (s UpdateLabelTableRequest) String() string {
	return tea.Prettify(s)
}

func (s UpdateLabelTableRequest) GoString() string {
	return s.String()
}

func (s *UpdateLabelTableRequest) SetDatasourceId(v string) *UpdateLabelTableRequest {
	s.DatasourceId = &v
	return s
}

func (s *UpdateLabelTableRequest) SetFields(v []*UpdateLabelTableRequestFields) *UpdateLabelTableRequest {
	s.Fields = v
	return s
}

func (s *UpdateLabelTableRequest) SetName(v string) *UpdateLabelTableRequest {
	s.Name = &v
	return s
}

type UpdateLabelTableRequestFields struct {
	Attributes []*string `json:"Attributes,omitempty" xml:"Attributes,omitempty" type:"Repeated"`
	Name       *string   `json:"Name,omitempty" xml:"Name,omitempty"`
	Type       *string   `json:"Type,omitempty" xml:"Type,omitempty"`
}

func (s UpdateLabelTableRequestFields) String() string {
	return tea.Prettify(s)
}

func (s UpdateLabelTableRequestFields) GoString() string {
	return s.String()
}

func (s *UpdateLabelTableRequestFields) SetAttributes(v []*string) *UpdateLabelTableRequestFields {
	s.Attributes = v
	return s
}

func (s *UpdateLabelTableRequestFields) SetName(v string) *UpdateLabelTableRequestFields {
	s.Name = &v
	return s
}

func (s *UpdateLabelTableRequestFields) SetType(v string) *UpdateLabelTableRequestFields {
	s.Type = &v
	return s
}

type UpdateLabelTableResponseBody struct {
	RequestId *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
}

func (s UpdateLabelTableResponseBody) String() string {
	return tea.Prettify(s)
}

func (s UpdateLabelTableResponseBody) GoString() string {
	return s.String()
}

func (s *UpdateLabelTableResponseBody) SetRequestId(v string) *UpdateLabelTableResponseBody {
	s.RequestId = &v
	return s
}

type UpdateLabelTableResponse struct {
	Headers    map[string]*string            `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
	StatusCode *int32                        `json:"statusCode,omitempty" xml:"statusCode,omitempty" require:"true"`
	Body       *UpdateLabelTableResponseBody `json:"body,omitempty" xml:"body,omitempty" require:"true"`
}

func (s UpdateLabelTableResponse) String() string {
	return tea.Prettify(s)
}

func (s UpdateLabelTableResponse) GoString() string {
	return s.String()
}

func (s *UpdateLabelTableResponse) SetHeaders(v map[string]*string) *UpdateLabelTableResponse {
	s.Headers = v
	return s
}

func (s *UpdateLabelTableResponse) SetStatusCode(v int32) *UpdateLabelTableResponse {
	s.StatusCode = &v
	return s
}

func (s *UpdateLabelTableResponse) SetBody(v *UpdateLabelTableResponseBody) *UpdateLabelTableResponse {
	s.Body = v
	return s
}

type UpdateModelFeatureRequest struct {
	Features     []*UpdateModelFeatureRequestFeatures `json:"Features,omitempty" xml:"Features,omitempty" type:"Repeated"`
	LabelTableId *string                              `json:"LabelTableId,omitempty" xml:"LabelTableId,omitempty"`
}

func (s UpdateModelFeatureRequest) String() string {
	return tea.Prettify(s)
}

func (s UpdateModelFeatureRequest) GoString() string {
	return s.String()
}

func (s *UpdateModelFeatureRequest) SetFeatures(v []*UpdateModelFeatureRequestFeatures) *UpdateModelFeatureRequest {
	s.Features = v
	return s
}

func (s *UpdateModelFeatureRequest) SetLabelTableId(v string) *UpdateModelFeatureRequest {
	s.LabelTableId = &v
	return s
}

type UpdateModelFeatureRequestFeatures struct {
	AliasName     *string `json:"AliasName,omitempty" xml:"AliasName,omitempty"`
	FeatureViewId *string `json:"FeatureViewId,omitempty" xml:"FeatureViewId,omitempty"`
	Name          *string `json:"Name,omitempty" xml:"Name,omitempty"`
	Type          *string `json:"Type,omitempty" xml:"Type,omitempty"`
}

func (s UpdateModelFeatureRequestFeatures) String() string {
	return tea.Prettify(s)
}

func (s UpdateModelFeatureRequestFeatures) GoString() string {
	return s.String()
}

func (s *UpdateModelFeatureRequestFeatures) SetAliasName(v string) *UpdateModelFeatureRequestFeatures {
	s.AliasName = &v
	return s
}

func (s *UpdateModelFeatureRequestFeatures) SetFeatureViewId(v string) *UpdateModelFeatureRequestFeatures {
	s.FeatureViewId = &v
	return s
}

func (s *UpdateModelFeatureRequestFeatures) SetName(v string) *UpdateModelFeatureRequestFeatures {
	s.Name = &v
	return s
}

func (s *UpdateModelFeatureRequestFeatures) SetType(v string) *UpdateModelFeatureRequestFeatures {
	s.Type = &v
	return s
}

type UpdateModelFeatureResponseBody struct {
	RequestId *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
}

func (s UpdateModelFeatureResponseBody) String() string {
	return tea.Prettify(s)
}

func (s UpdateModelFeatureResponseBody) GoString() string {
	return s.String()
}

func (s *UpdateModelFeatureResponseBody) SetRequestId(v string) *UpdateModelFeatureResponseBody {
	s.RequestId = &v
	return s
}

type UpdateModelFeatureResponse struct {
	Headers    map[string]*string              `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
	StatusCode *int32                          `json:"statusCode,omitempty" xml:"statusCode,omitempty" require:"true"`
	Body       *UpdateModelFeatureResponseBody `json:"body,omitempty" xml:"body,omitempty" require:"true"`
}

func (s UpdateModelFeatureResponse) String() string {
	return tea.Prettify(s)
}

func (s UpdateModelFeatureResponse) GoString() string {
	return s.String()
}

func (s *UpdateModelFeatureResponse) SetHeaders(v map[string]*string) *UpdateModelFeatureResponse {
	s.Headers = v
	return s
}

func (s *UpdateModelFeatureResponse) SetStatusCode(v int32) *UpdateModelFeatureResponse {
	s.StatusCode = &v
	return s
}

func (s *UpdateModelFeatureResponse) SetBody(v *UpdateModelFeatureResponseBody) *UpdateModelFeatureResponse {
	s.Body = v
	return s
}

type UpdateProjectRequest struct {
	Description *string `json:"Description,omitempty" xml:"Description,omitempty"`
	Name        *string `json:"Name,omitempty" xml:"Name,omitempty"`
}

func (s UpdateProjectRequest) String() string {
	return tea.Prettify(s)
}

func (s UpdateProjectRequest) GoString() string {
	return s.String()
}

func (s *UpdateProjectRequest) SetDescription(v string) *UpdateProjectRequest {
	s.Description = &v
	return s
}

func (s *UpdateProjectRequest) SetName(v string) *UpdateProjectRequest {
	s.Name = &v
	return s
}

type UpdateProjectResponseBody struct {
	RequestId *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
}

func (s UpdateProjectResponseBody) String() string {
	return tea.Prettify(s)
}

func (s UpdateProjectResponseBody) GoString() string {
	return s.String()
}

func (s *UpdateProjectResponseBody) SetRequestId(v string) *UpdateProjectResponseBody {
	s.RequestId = &v
	return s
}

type UpdateProjectResponse struct {
	Headers    map[string]*string         `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
	StatusCode *int32                     `json:"statusCode,omitempty" xml:"statusCode,omitempty" require:"true"`
	Body       *UpdateProjectResponseBody `json:"body,omitempty" xml:"body,omitempty" require:"true"`
}

func (s UpdateProjectResponse) String() string {
	return tea.Prettify(s)
}

func (s UpdateProjectResponse) GoString() string {
	return s.String()
}

func (s *UpdateProjectResponse) SetHeaders(v map[string]*string) *UpdateProjectResponse {
	s.Headers = v
	return s
}

func (s *UpdateProjectResponse) SetStatusCode(v int32) *UpdateProjectResponse {
	s.StatusCode = &v
	return s
}

func (s *UpdateProjectResponse) SetBody(v *UpdateProjectResponseBody) *UpdateProjectResponse {
	s.Body = v
	return s
}

type WriteFeatureViewTableRequest struct {
	Mode          *string                                    `json:"Mode,omitempty" xml:"Mode,omitempty"`
	Partitions    map[string]map[string]interface{}          `json:"Partitions,omitempty" xml:"Partitions,omitempty"`
	UrlDatasource *WriteFeatureViewTableRequestUrlDatasource `json:"UrlDatasource,omitempty" xml:"UrlDatasource,omitempty" type:"Struct"`
}

func (s WriteFeatureViewTableRequest) String() string {
	return tea.Prettify(s)
}

func (s WriteFeatureViewTableRequest) GoString() string {
	return s.String()
}

func (s *WriteFeatureViewTableRequest) SetMode(v string) *WriteFeatureViewTableRequest {
	s.Mode = &v
	return s
}

func (s *WriteFeatureViewTableRequest) SetPartitions(v map[string]map[string]interface{}) *WriteFeatureViewTableRequest {
	s.Partitions = v
	return s
}

func (s *WriteFeatureViewTableRequest) SetUrlDatasource(v *WriteFeatureViewTableRequestUrlDatasource) *WriteFeatureViewTableRequest {
	s.UrlDatasource = v
	return s
}

type WriteFeatureViewTableRequestUrlDatasource struct {
	Delimiter  *string `json:"Delimiter,omitempty" xml:"Delimiter,omitempty"`
	OmitHeader *bool   `json:"OmitHeader,omitempty" xml:"OmitHeader,omitempty"`
	Path       *string `json:"Path,omitempty" xml:"Path,omitempty"`
}

func (s WriteFeatureViewTableRequestUrlDatasource) String() string {
	return tea.Prettify(s)
}

func (s WriteFeatureViewTableRequestUrlDatasource) GoString() string {
	return s.String()
}

func (s *WriteFeatureViewTableRequestUrlDatasource) SetDelimiter(v string) *WriteFeatureViewTableRequestUrlDatasource {
	s.Delimiter = &v
	return s
}

func (s *WriteFeatureViewTableRequestUrlDatasource) SetOmitHeader(v bool) *WriteFeatureViewTableRequestUrlDatasource {
	s.OmitHeader = &v
	return s
}

func (s *WriteFeatureViewTableRequestUrlDatasource) SetPath(v string) *WriteFeatureViewTableRequestUrlDatasource {
	s.Path = &v
	return s
}

type WriteFeatureViewTableResponseBody struct {
	RequestId *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
}

func (s WriteFeatureViewTableResponseBody) String() string {
	return tea.Prettify(s)
}

func (s WriteFeatureViewTableResponseBody) GoString() string {
	return s.String()
}

func (s *WriteFeatureViewTableResponseBody) SetRequestId(v string) *WriteFeatureViewTableResponseBody {
	s.RequestId = &v
	return s
}

type WriteFeatureViewTableResponse struct {
	Headers    map[string]*string                 `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
	StatusCode *int32                             `json:"statusCode,omitempty" xml:"statusCode,omitempty" require:"true"`
	Body       *WriteFeatureViewTableResponseBody `json:"body,omitempty" xml:"body,omitempty" require:"true"`
}

func (s WriteFeatureViewTableResponse) String() string {
	return tea.Prettify(s)
}

func (s WriteFeatureViewTableResponse) GoString() string {
	return s.String()
}

func (s *WriteFeatureViewTableResponse) SetHeaders(v map[string]*string) *WriteFeatureViewTableResponse {
	s.Headers = v
	return s
}

func (s *WriteFeatureViewTableResponse) SetStatusCode(v int32) *WriteFeatureViewTableResponse {
	s.StatusCode = &v
	return s
}

func (s *WriteFeatureViewTableResponse) SetBody(v *WriteFeatureViewTableResponseBody) *WriteFeatureViewTableResponse {
	s.Body = v
	return s
}

type WriteProjectFeatureEntityHotIdsRequest struct {
	HotIds  *string `json:"HotIds,omitempty" xml:"HotIds,omitempty"`
	Version *string `json:"Version,omitempty" xml:"Version,omitempty"`
}

func (s WriteProjectFeatureEntityHotIdsRequest) String() string {
	return tea.Prettify(s)
}

func (s WriteProjectFeatureEntityHotIdsRequest) GoString() string {
	return s.String()
}

func (s *WriteProjectFeatureEntityHotIdsRequest) SetHotIds(v string) *WriteProjectFeatureEntityHotIdsRequest {
	s.HotIds = &v
	return s
}

func (s *WriteProjectFeatureEntityHotIdsRequest) SetVersion(v string) *WriteProjectFeatureEntityHotIdsRequest {
	s.Version = &v
	return s
}

type WriteProjectFeatureEntityHotIdsResponseBody struct {
	RequestId *string `json:"RequestId,omitempty" xml:"RequestId,omitempty"`
}

func (s WriteProjectFeatureEntityHotIdsResponseBody) String() string {
	return tea.Prettify(s)
}

func (s WriteProjectFeatureEntityHotIdsResponseBody) GoString() string {
	return s.String()
}

func (s *WriteProjectFeatureEntityHotIdsResponseBody) SetRequestId(v string) *WriteProjectFeatureEntityHotIdsResponseBody {
	s.RequestId = &v
	return s
}

type WriteProjectFeatureEntityHotIdsResponse struct {
	Headers    map[string]*string                           `json:"headers,omitempty" xml:"headers,omitempty" require:"true"`
	StatusCode *int32                                       `json:"statusCode,omitempty" xml:"statusCode,omitempty" require:"true"`
	Body       *WriteProjectFeatureEntityHotIdsResponseBody `json:"body,omitempty" xml:"body,omitempty" require:"true"`
}

func (s WriteProjectFeatureEntityHotIdsResponse) String() string {
	return tea.Prettify(s)
}

func (s WriteProjectFeatureEntityHotIdsResponse) GoString() string {
	return s.String()
}

func (s *WriteProjectFeatureEntityHotIdsResponse) SetHeaders(v map[string]*string) *WriteProjectFeatureEntityHotIdsResponse {
	s.Headers = v
	return s
}

func (s *WriteProjectFeatureEntityHotIdsResponse) SetStatusCode(v int32) *WriteProjectFeatureEntityHotIdsResponse {
	s.StatusCode = &v
	return s
}

func (s *WriteProjectFeatureEntityHotIdsResponse) SetBody(v *WriteProjectFeatureEntityHotIdsResponseBody) *WriteProjectFeatureEntityHotIdsResponse {
	s.Body = v
	return s
}

type Client struct {
	openapi.Client
}

func NewClient(config *openapi.Config) (*Client, error) {
	client := new(Client)
	err := client.Init(config)
	return client, err
}

func (client *Client) Init(config *openapi.Config) (_err error) {
	_err = client.Client.Init(config)
	if _err != nil {
		return _err
	}
	client.EndpointRule = tea.String("")
	_err = client.CheckConfig(config)
	if _err != nil {
		return _err
	}
	client.Endpoint, _err = client.GetEndpoint(tea.String("paifeaturestore"), client.RegionId, client.EndpointRule, client.Network, client.Suffix, client.EndpointMap, client.Endpoint)
	if _err != nil {
		return _err
	}

	return nil
}

func (client *Client) GetEndpoint(productId *string, regionId *string, endpointRule *string, network *string, suffix *string, endpointMap map[string]*string, endpoint *string) (_result *string, _err error) {
	if !tea.BoolValue(util.Empty(endpoint)) {
		_result = endpoint
		return _result, _err
	}

	if !tea.BoolValue(util.IsUnset(endpointMap)) && !tea.BoolValue(util.Empty(endpointMap[tea.StringValue(regionId)])) {
		_result = endpointMap[tea.StringValue(regionId)]
		return _result, _err
	}

	_body, _err := endpointutil.GetEndpointRules(productId, regionId, endpointRule, network, suffix)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) ChangeProjectFeatureEntityHotIdVersionWithOptions(InstanceId *string, ProjectId *string, FeatureEntityName *string, request *ChangeProjectFeatureEntityHotIdVersionRequest, headers map[string]*string, runtime *util.RuntimeOptions) (_result *ChangeProjectFeatureEntityHotIdVersionResponse, _err error) {
	_err = util.ValidateModel(request)
	if _err != nil {
		return _result, _err
	}
	body := map[string]interface{}{}
	if !tea.BoolValue(util.IsUnset(request.Version)) {
		body["Version"] = request.Version
	}

	req := &openapi.OpenApiRequest{
		Headers: headers,
		Body:    openapiutil.ParseToMap(body),
	}
	params := &openapi.Params{
		Action:      tea.String("ChangeProjectFeatureEntityHotIdVersion"),
		Version:     tea.String("2023-06-21"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/api/v1/instances/" + tea.StringValue(openapiutil.GetEncodeParam(InstanceId)) + "/projects/" + tea.StringValue(openapiutil.GetEncodeParam(ProjectId)) + "/featureentities/" + tea.StringValue(openapiutil.GetEncodeParam(FeatureEntityName)) + "/action/changehotidversion"),
		Method:      tea.String("POST"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("json"),
	}
	_result = &ChangeProjectFeatureEntityHotIdVersionResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) ChangeProjectFeatureEntityHotIdVersion(InstanceId *string, ProjectId *string, FeatureEntityName *string, request *ChangeProjectFeatureEntityHotIdVersionRequest) (_result *ChangeProjectFeatureEntityHotIdVersionResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &ChangeProjectFeatureEntityHotIdVersionResponse{}
	_body, _err := client.ChangeProjectFeatureEntityHotIdVersionWithOptions(InstanceId, ProjectId, FeatureEntityName, request, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) CheckInstanceDatasourceWithOptions(InstanceId *string, request *CheckInstanceDatasourceRequest, headers map[string]*string, runtime *util.RuntimeOptions) (_result *CheckInstanceDatasourceResponse, _err error) {
	_err = util.ValidateModel(request)
	if _err != nil {
		return _result, _err
	}
	body := map[string]interface{}{}
	if !tea.BoolValue(util.IsUnset(request.Config)) {
		body["Config"] = request.Config
	}

	if !tea.BoolValue(util.IsUnset(request.Type)) {
		body["Type"] = request.Type
	}

	if !tea.BoolValue(util.IsUnset(request.Uri)) {
		body["Uri"] = request.Uri
	}

	req := &openapi.OpenApiRequest{
		Headers: headers,
		Body:    openapiutil.ParseToMap(body),
	}
	params := &openapi.Params{
		Action:      tea.String("CheckInstanceDatasource"),
		Version:     tea.String("2023-06-21"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/api/v1/instances/" + tea.StringValue(openapiutil.GetEncodeParam(InstanceId)) + "/action/checkdatasource"),
		Method:      tea.String("POST"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("json"),
	}
	_result = &CheckInstanceDatasourceResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) CheckInstanceDatasource(InstanceId *string, request *CheckInstanceDatasourceRequest) (_result *CheckInstanceDatasourceResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &CheckInstanceDatasourceResponse{}
	_body, _err := client.CheckInstanceDatasourceWithOptions(InstanceId, request, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) CreateDatasourceWithOptions(InstanceId *string, request *CreateDatasourceRequest, headers map[string]*string, runtime *util.RuntimeOptions) (_result *CreateDatasourceResponse, _err error) {
	_err = util.ValidateModel(request)
	if _err != nil {
		return _result, _err
	}
	body := map[string]interface{}{}
	if !tea.BoolValue(util.IsUnset(request.Config)) {
		body["Config"] = request.Config
	}

	if !tea.BoolValue(util.IsUnset(request.Name)) {
		body["Name"] = request.Name
	}

	if !tea.BoolValue(util.IsUnset(request.Type)) {
		body["Type"] = request.Type
	}

	if !tea.BoolValue(util.IsUnset(request.Uri)) {
		body["Uri"] = request.Uri
	}

	if !tea.BoolValue(util.IsUnset(request.WorkspaceId)) {
		body["WorkspaceId"] = request.WorkspaceId
	}

	req := &openapi.OpenApiRequest{
		Headers: headers,
		Body:    openapiutil.ParseToMap(body),
	}
	params := &openapi.Params{
		Action:      tea.String("CreateDatasource"),
		Version:     tea.String("2023-06-21"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/api/v1/instances/" + tea.StringValue(openapiutil.GetEncodeParam(InstanceId)) + "/datasources"),
		Method:      tea.String("POST"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("json"),
	}
	_result = &CreateDatasourceResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) CreateDatasource(InstanceId *string, request *CreateDatasourceRequest) (_result *CreateDatasourceResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &CreateDatasourceResponse{}
	_body, _err := client.CreateDatasourceWithOptions(InstanceId, request, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) CreateFeatureEntityWithOptions(InstanceId *string, request *CreateFeatureEntityRequest, headers map[string]*string, runtime *util.RuntimeOptions) (_result *CreateFeatureEntityResponse, _err error) {
	_err = util.ValidateModel(request)
	if _err != nil {
		return _result, _err
	}
	body := map[string]interface{}{}
	if !tea.BoolValue(util.IsUnset(request.JoinId)) {
		body["JoinId"] = request.JoinId
	}

	if !tea.BoolValue(util.IsUnset(request.Name)) {
		body["Name"] = request.Name
	}

	if !tea.BoolValue(util.IsUnset(request.ProjectId)) {
		body["ProjectId"] = request.ProjectId
	}

	req := &openapi.OpenApiRequest{
		Headers: headers,
		Body:    openapiutil.ParseToMap(body),
	}
	params := &openapi.Params{
		Action:      tea.String("CreateFeatureEntity"),
		Version:     tea.String("2023-06-21"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/api/v1/instances/" + tea.StringValue(openapiutil.GetEncodeParam(InstanceId)) + "/featureentities"),
		Method:      tea.String("POST"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("json"),
	}
	_result = &CreateFeatureEntityResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) CreateFeatureEntity(InstanceId *string, request *CreateFeatureEntityRequest) (_result *CreateFeatureEntityResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &CreateFeatureEntityResponse{}
	_body, _err := client.CreateFeatureEntityWithOptions(InstanceId, request, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) CreateFeatureViewWithOptions(InstanceId *string, request *CreateFeatureViewRequest, headers map[string]*string, runtime *util.RuntimeOptions) (_result *CreateFeatureViewResponse, _err error) {
	_err = util.ValidateModel(request)
	if _err != nil {
		return _result, _err
	}
	body := map[string]interface{}{}
	if !tea.BoolValue(util.IsUnset(request.Config)) {
		body["Config"] = request.Config
	}

	if !tea.BoolValue(util.IsUnset(request.FeatureEntityId)) {
		body["FeatureEntityId"] = request.FeatureEntityId
	}

	if !tea.BoolValue(util.IsUnset(request.Fields)) {
		body["Fields"] = request.Fields
	}

	if !tea.BoolValue(util.IsUnset(request.Name)) {
		body["Name"] = request.Name
	}

	if !tea.BoolValue(util.IsUnset(request.ProjectId)) {
		body["ProjectId"] = request.ProjectId
	}

	if !tea.BoolValue(util.IsUnset(request.RegisterDatasourceId)) {
		body["RegisterDatasourceId"] = request.RegisterDatasourceId
	}

	if !tea.BoolValue(util.IsUnset(request.RegisterTable)) {
		body["RegisterTable"] = request.RegisterTable
	}

	if !tea.BoolValue(util.IsUnset(request.SyncOnlineTable)) {
		body["SyncOnlineTable"] = request.SyncOnlineTable
	}

	if !tea.BoolValue(util.IsUnset(request.TTL)) {
		body["TTL"] = request.TTL
	}

	if !tea.BoolValue(util.IsUnset(request.Tags)) {
		body["Tags"] = request.Tags
	}

	if !tea.BoolValue(util.IsUnset(request.Type)) {
		body["Type"] = request.Type
	}

	if !tea.BoolValue(util.IsUnset(request.WriteMethod)) {
		body["WriteMethod"] = request.WriteMethod
	}

	req := &openapi.OpenApiRequest{
		Headers: headers,
		Body:    openapiutil.ParseToMap(body),
	}
	params := &openapi.Params{
		Action:      tea.String("CreateFeatureView"),
		Version:     tea.String("2023-06-21"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/api/v1/instances/" + tea.StringValue(openapiutil.GetEncodeParam(InstanceId)) + "/featureviews"),
		Method:      tea.String("POST"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("json"),
	}
	_result = &CreateFeatureViewResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) CreateFeatureView(InstanceId *string, request *CreateFeatureViewRequest) (_result *CreateFeatureViewResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &CreateFeatureViewResponse{}
	_body, _err := client.CreateFeatureViewWithOptions(InstanceId, request, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) CreateInstanceWithOptions(request *CreateInstanceRequest, headers map[string]*string, runtime *util.RuntimeOptions) (_result *CreateInstanceResponse, _err error) {
	_err = util.ValidateModel(request)
	if _err != nil {
		return _result, _err
	}
	body := map[string]interface{}{}
	if !tea.BoolValue(util.IsUnset(request.Type)) {
		body["Type"] = request.Type
	}

	req := &openapi.OpenApiRequest{
		Headers: headers,
		Body:    openapiutil.ParseToMap(body),
	}
	params := &openapi.Params{
		Action:      tea.String("CreateInstance"),
		Version:     tea.String("2023-06-21"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/api/v1/instances"),
		Method:      tea.String("POST"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("json"),
	}
	_result = &CreateInstanceResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) CreateInstance(request *CreateInstanceRequest) (_result *CreateInstanceResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &CreateInstanceResponse{}
	_body, _err := client.CreateInstanceWithOptions(request, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) CreateLabelTableWithOptions(InstanceId *string, request *CreateLabelTableRequest, headers map[string]*string, runtime *util.RuntimeOptions) (_result *CreateLabelTableResponse, _err error) {
	_err = util.ValidateModel(request)
	if _err != nil {
		return _result, _err
	}
	body := map[string]interface{}{}
	if !tea.BoolValue(util.IsUnset(request.DatasourceId)) {
		body["DatasourceId"] = request.DatasourceId
	}

	if !tea.BoolValue(util.IsUnset(request.Fields)) {
		body["Fields"] = request.Fields
	}

	if !tea.BoolValue(util.IsUnset(request.Name)) {
		body["Name"] = request.Name
	}

	if !tea.BoolValue(util.IsUnset(request.ProjectId)) {
		body["ProjectId"] = request.ProjectId
	}

	req := &openapi.OpenApiRequest{
		Headers: headers,
		Body:    openapiutil.ParseToMap(body),
	}
	params := &openapi.Params{
		Action:      tea.String("CreateLabelTable"),
		Version:     tea.String("2023-06-21"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/api/v1/instances/" + tea.StringValue(openapiutil.GetEncodeParam(InstanceId)) + "/labeltables"),
		Method:      tea.String("POST"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("json"),
	}
	_result = &CreateLabelTableResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) CreateLabelTable(InstanceId *string, request *CreateLabelTableRequest) (_result *CreateLabelTableResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &CreateLabelTableResponse{}
	_body, _err := client.CreateLabelTableWithOptions(InstanceId, request, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) CreateModelFeatureWithOptions(InstanceId *string, request *CreateModelFeatureRequest, headers map[string]*string, runtime *util.RuntimeOptions) (_result *CreateModelFeatureResponse, _err error) {
	_err = util.ValidateModel(request)
	if _err != nil {
		return _result, _err
	}
	body := map[string]interface{}{}
	if !tea.BoolValue(util.IsUnset(request.Features)) {
		body["Features"] = request.Features
	}

	if !tea.BoolValue(util.IsUnset(request.LabelTableId)) {
		body["LabelTableId"] = request.LabelTableId
	}

	if !tea.BoolValue(util.IsUnset(request.Name)) {
		body["Name"] = request.Name
	}

	if !tea.BoolValue(util.IsUnset(request.ProjectId)) {
		body["ProjectId"] = request.ProjectId
	}

	req := &openapi.OpenApiRequest{
		Headers: headers,
		Body:    openapiutil.ParseToMap(body),
	}
	params := &openapi.Params{
		Action:      tea.String("CreateModelFeature"),
		Version:     tea.String("2023-06-21"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/api/v1/instances/" + tea.StringValue(openapiutil.GetEncodeParam(InstanceId)) + "/modelfeatures"),
		Method:      tea.String("POST"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("json"),
	}
	_result = &CreateModelFeatureResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) CreateModelFeature(InstanceId *string, request *CreateModelFeatureRequest) (_result *CreateModelFeatureResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &CreateModelFeatureResponse{}
	_body, _err := client.CreateModelFeatureWithOptions(InstanceId, request, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) CreateProjectWithOptions(InstanceId *string, request *CreateProjectRequest, headers map[string]*string, runtime *util.RuntimeOptions) (_result *CreateProjectResponse, _err error) {
	_err = util.ValidateModel(request)
	if _err != nil {
		return _result, _err
	}
	body := map[string]interface{}{}
	if !tea.BoolValue(util.IsUnset(request.Description)) {
		body["Description"] = request.Description
	}

	if !tea.BoolValue(util.IsUnset(request.Name)) {
		body["Name"] = request.Name
	}

	if !tea.BoolValue(util.IsUnset(request.OfflineDatasourceId)) {
		body["OfflineDatasourceId"] = request.OfflineDatasourceId
	}

	if !tea.BoolValue(util.IsUnset(request.OfflineLifeCycle)) {
		body["OfflineLifeCycle"] = request.OfflineLifeCycle
	}

	if !tea.BoolValue(util.IsUnset(request.OnlineDatasourceId)) {
		body["OnlineDatasourceId"] = request.OnlineDatasourceId
	}

	if !tea.BoolValue(util.IsUnset(request.WorkspaceId)) {
		body["WorkspaceId"] = request.WorkspaceId
	}

	req := &openapi.OpenApiRequest{
		Headers: headers,
		Body:    openapiutil.ParseToMap(body),
	}
	params := &openapi.Params{
		Action:      tea.String("CreateProject"),
		Version:     tea.String("2023-06-21"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/api/v1/instances/" + tea.StringValue(openapiutil.GetEncodeParam(InstanceId)) + "/projects"),
		Method:      tea.String("POST"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("json"),
	}
	_result = &CreateProjectResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) CreateProject(InstanceId *string, request *CreateProjectRequest) (_result *CreateProjectResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &CreateProjectResponse{}
	_body, _err := client.CreateProjectWithOptions(InstanceId, request, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) CreateServiceIdentityRoleWithOptions(request *CreateServiceIdentityRoleRequest, headers map[string]*string, runtime *util.RuntimeOptions) (_result *CreateServiceIdentityRoleResponse, _err error) {
	_err = util.ValidateModel(request)
	if _err != nil {
		return _result, _err
	}
	body := map[string]interface{}{}
	if !tea.BoolValue(util.IsUnset(request.RoleName)) {
		body["RoleName"] = request.RoleName
	}

	req := &openapi.OpenApiRequest{
		Headers: headers,
		Body:    openapiutil.ParseToMap(body),
	}
	params := &openapi.Params{
		Action:      tea.String("CreateServiceIdentityRole"),
		Version:     tea.String("2023-06-21"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/api/v1/serviceidentityroles"),
		Method:      tea.String("POST"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("json"),
	}
	_result = &CreateServiceIdentityRoleResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) CreateServiceIdentityRole(request *CreateServiceIdentityRoleRequest) (_result *CreateServiceIdentityRoleResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &CreateServiceIdentityRoleResponse{}
	_body, _err := client.CreateServiceIdentityRoleWithOptions(request, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) DeleteDatasourceWithOptions(InstanceId *string, DatasourceId *string, headers map[string]*string, runtime *util.RuntimeOptions) (_result *DeleteDatasourceResponse, _err error) {
	req := &openapi.OpenApiRequest{
		Headers: headers,
	}
	params := &openapi.Params{
		Action:      tea.String("DeleteDatasource"),
		Version:     tea.String("2023-06-21"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/api/v1/instances/" + tea.StringValue(openapiutil.GetEncodeParam(InstanceId)) + "/datasources/" + tea.StringValue(openapiutil.GetEncodeParam(DatasourceId))),
		Method:      tea.String("DELETE"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("json"),
	}
	_result = &DeleteDatasourceResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) DeleteDatasource(InstanceId *string, DatasourceId *string) (_result *DeleteDatasourceResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &DeleteDatasourceResponse{}
	_body, _err := client.DeleteDatasourceWithOptions(InstanceId, DatasourceId, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) DeleteFeatureEntityWithOptions(InstanceId *string, FeatureEntityId *string, headers map[string]*string, runtime *util.RuntimeOptions) (_result *DeleteFeatureEntityResponse, _err error) {
	req := &openapi.OpenApiRequest{
		Headers: headers,
	}
	params := &openapi.Params{
		Action:      tea.String("DeleteFeatureEntity"),
		Version:     tea.String("2023-06-21"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/api/v1/instances/" + tea.StringValue(openapiutil.GetEncodeParam(InstanceId)) + "/featureentities/" + tea.StringValue(openapiutil.GetEncodeParam(FeatureEntityId))),
		Method:      tea.String("DELETE"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("json"),
	}
	_result = &DeleteFeatureEntityResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) DeleteFeatureEntity(InstanceId *string, FeatureEntityId *string) (_result *DeleteFeatureEntityResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &DeleteFeatureEntityResponse{}
	_body, _err := client.DeleteFeatureEntityWithOptions(InstanceId, FeatureEntityId, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) DeleteFeatureViewWithOptions(InstanceId *string, FeatureViewId *string, headers map[string]*string, runtime *util.RuntimeOptions) (_result *DeleteFeatureViewResponse, _err error) {
	req := &openapi.OpenApiRequest{
		Headers: headers,
	}
	params := &openapi.Params{
		Action:      tea.String("DeleteFeatureView"),
		Version:     tea.String("2023-06-21"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/api/v1/instances/" + tea.StringValue(openapiutil.GetEncodeParam(InstanceId)) + "/featureviews/" + tea.StringValue(openapiutil.GetEncodeParam(FeatureViewId))),
		Method:      tea.String("DELETE"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("json"),
	}
	_result = &DeleteFeatureViewResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) DeleteFeatureView(InstanceId *string, FeatureViewId *string) (_result *DeleteFeatureViewResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &DeleteFeatureViewResponse{}
	_body, _err := client.DeleteFeatureViewWithOptions(InstanceId, FeatureViewId, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) DeleteLabelTableWithOptions(InstanceId *string, LabelTableId *string, headers map[string]*string, runtime *util.RuntimeOptions) (_result *DeleteLabelTableResponse, _err error) {
	req := &openapi.OpenApiRequest{
		Headers: headers,
	}
	params := &openapi.Params{
		Action:      tea.String("DeleteLabelTable"),
		Version:     tea.String("2023-06-21"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/api/v1/instances/" + tea.StringValue(openapiutil.GetEncodeParam(InstanceId)) + "/labeltables/" + tea.StringValue(openapiutil.GetEncodeParam(LabelTableId))),
		Method:      tea.String("DELETE"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("json"),
	}
	_result = &DeleteLabelTableResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) DeleteLabelTable(InstanceId *string, LabelTableId *string) (_result *DeleteLabelTableResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &DeleteLabelTableResponse{}
	_body, _err := client.DeleteLabelTableWithOptions(InstanceId, LabelTableId, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) DeleteModelFeatureWithOptions(InstanceId *string, ModelFeatureId *string, headers map[string]*string, runtime *util.RuntimeOptions) (_result *DeleteModelFeatureResponse, _err error) {
	req := &openapi.OpenApiRequest{
		Headers: headers,
	}
	params := &openapi.Params{
		Action:      tea.String("DeleteModelFeature"),
		Version:     tea.String("2023-06-21"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/api/v1/instances/" + tea.StringValue(openapiutil.GetEncodeParam(InstanceId)) + "/modelfeatures/" + tea.StringValue(openapiutil.GetEncodeParam(ModelFeatureId))),
		Method:      tea.String("DELETE"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("json"),
	}
	_result = &DeleteModelFeatureResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) DeleteModelFeature(InstanceId *string, ModelFeatureId *string) (_result *DeleteModelFeatureResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &DeleteModelFeatureResponse{}
	_body, _err := client.DeleteModelFeatureWithOptions(InstanceId, ModelFeatureId, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) DeleteProjectWithOptions(InstanceId *string, ProjectId *string, headers map[string]*string, runtime *util.RuntimeOptions) (_result *DeleteProjectResponse, _err error) {
	req := &openapi.OpenApiRequest{
		Headers: headers,
	}
	params := &openapi.Params{
		Action:      tea.String("DeleteProject"),
		Version:     tea.String("2023-06-21"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/api/v1/instances/" + tea.StringValue(openapiutil.GetEncodeParam(InstanceId)) + "/projects/" + tea.StringValue(openapiutil.GetEncodeParam(ProjectId))),
		Method:      tea.String("DELETE"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("json"),
	}
	_result = &DeleteProjectResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) DeleteProject(InstanceId *string, ProjectId *string) (_result *DeleteProjectResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &DeleteProjectResponse{}
	_body, _err := client.DeleteProjectWithOptions(InstanceId, ProjectId, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) ExportModelFeatureTrainingSetTableWithOptions(InstanceId *string, ModelFeatureId *string, request *ExportModelFeatureTrainingSetTableRequest, headers map[string]*string, runtime *util.RuntimeOptions) (_result *ExportModelFeatureTrainingSetTableResponse, _err error) {
	_err = util.ValidateModel(request)
	if _err != nil {
		return _result, _err
	}
	body := map[string]interface{}{}
	if !tea.BoolValue(util.IsUnset(request.FeatureViewConfig)) {
		body["FeatureViewConfig"] = request.FeatureViewConfig
	}

	if !tea.BoolValue(util.IsUnset(request.LabelInputConfig)) {
		body["LabelInputConfig"] = request.LabelInputConfig
	}

	if !tea.BoolValue(util.IsUnset(request.TrainingSetConfig)) {
		body["TrainingSetConfig"] = request.TrainingSetConfig
	}

	req := &openapi.OpenApiRequest{
		Headers: headers,
		Body:    openapiutil.ParseToMap(body),
	}
	params := &openapi.Params{
		Action:      tea.String("ExportModelFeatureTrainingSetTable"),
		Version:     tea.String("2023-06-21"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/api/v1/instances/" + tea.StringValue(openapiutil.GetEncodeParam(InstanceId)) + "/modelfeatures/" + tea.StringValue(openapiutil.GetEncodeParam(ModelFeatureId)) + "/action/exporttrainingsettable"),
		Method:      tea.String("POST"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("json"),
	}
	_result = &ExportModelFeatureTrainingSetTableResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) ExportModelFeatureTrainingSetTable(InstanceId *string, ModelFeatureId *string, request *ExportModelFeatureTrainingSetTableRequest) (_result *ExportModelFeatureTrainingSetTableResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &ExportModelFeatureTrainingSetTableResponse{}
	_body, _err := client.ExportModelFeatureTrainingSetTableWithOptions(InstanceId, ModelFeatureId, request, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) GetDatasourceWithOptions(InstanceId *string, DatasourceId *string, headers map[string]*string, runtime *util.RuntimeOptions) (_result *GetDatasourceResponse, _err error) {
	req := &openapi.OpenApiRequest{
		Headers: headers,
	}
	params := &openapi.Params{
		Action:      tea.String("GetDatasource"),
		Version:     tea.String("2023-06-21"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/api/v1/instances/" + tea.StringValue(openapiutil.GetEncodeParam(InstanceId)) + "/datasources/" + tea.StringValue(openapiutil.GetEncodeParam(DatasourceId))),
		Method:      tea.String("GET"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("json"),
	}
	_result = &GetDatasourceResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) GetDatasource(InstanceId *string, DatasourceId *string) (_result *GetDatasourceResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &GetDatasourceResponse{}
	_body, _err := client.GetDatasourceWithOptions(InstanceId, DatasourceId, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) GetDatasourceTableWithOptions(InstanceId *string, DatasourceId *string, TableName *string, headers map[string]*string, runtime *util.RuntimeOptions) (_result *GetDatasourceTableResponse, _err error) {
	req := &openapi.OpenApiRequest{
		Headers: headers,
	}
	params := &openapi.Params{
		Action:      tea.String("GetDatasourceTable"),
		Version:     tea.String("2023-06-21"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/api/v1/instances/" + tea.StringValue(openapiutil.GetEncodeParam(InstanceId)) + "/datasources/" + tea.StringValue(openapiutil.GetEncodeParam(DatasourceId)) + "/tables/" + tea.StringValue(openapiutil.GetEncodeParam(TableName))),
		Method:      tea.String("GET"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("json"),
	}
	_result = &GetDatasourceTableResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) GetDatasourceTable(InstanceId *string, DatasourceId *string, TableName *string) (_result *GetDatasourceTableResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &GetDatasourceTableResponse{}
	_body, _err := client.GetDatasourceTableWithOptions(InstanceId, DatasourceId, TableName, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) GetFeatureEntityWithOptions(InstanceId *string, FeatureEntityId *string, headers map[string]*string, runtime *util.RuntimeOptions) (_result *GetFeatureEntityResponse, _err error) {
	req := &openapi.OpenApiRequest{
		Headers: headers,
	}
	params := &openapi.Params{
		Action:      tea.String("GetFeatureEntity"),
		Version:     tea.String("2023-06-21"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/api/v1/instances/" + tea.StringValue(openapiutil.GetEncodeParam(InstanceId)) + "/featureentities/" + tea.StringValue(openapiutil.GetEncodeParam(FeatureEntityId))),
		Method:      tea.String("GET"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("json"),
	}
	_result = &GetFeatureEntityResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) GetFeatureEntity(InstanceId *string, FeatureEntityId *string) (_result *GetFeatureEntityResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &GetFeatureEntityResponse{}
	_body, _err := client.GetFeatureEntityWithOptions(InstanceId, FeatureEntityId, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) GetFeatureViewWithOptions(InstanceId *string, FeatureViewId *string, headers map[string]*string, runtime *util.RuntimeOptions) (_result *GetFeatureViewResponse, _err error) {
	req := &openapi.OpenApiRequest{
		Headers: headers,
	}
	params := &openapi.Params{
		Action:      tea.String("GetFeatureView"),
		Version:     tea.String("2023-06-21"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/api/v1/instances/" + tea.StringValue(openapiutil.GetEncodeParam(InstanceId)) + "/featureviews/" + tea.StringValue(openapiutil.GetEncodeParam(FeatureViewId))),
		Method:      tea.String("GET"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("json"),
	}
	_result = &GetFeatureViewResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) GetFeatureView(InstanceId *string, FeatureViewId *string) (_result *GetFeatureViewResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &GetFeatureViewResponse{}
	_body, _err := client.GetFeatureViewWithOptions(InstanceId, FeatureViewId, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) GetInstanceWithOptions(InstanceId *string, headers map[string]*string, runtime *util.RuntimeOptions) (_result *GetInstanceResponse, _err error) {
	req := &openapi.OpenApiRequest{
		Headers: headers,
	}
	params := &openapi.Params{
		Action:      tea.String("GetInstance"),
		Version:     tea.String("2023-06-21"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/api/v1/instances/" + tea.StringValue(openapiutil.GetEncodeParam(InstanceId))),
		Method:      tea.String("GET"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("json"),
	}
	_result = &GetInstanceResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) GetInstance(InstanceId *string) (_result *GetInstanceResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &GetInstanceResponse{}
	_body, _err := client.GetInstanceWithOptions(InstanceId, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) GetLabelTableWithOptions(InstanceId *string, LabelTableId *string, headers map[string]*string, runtime *util.RuntimeOptions) (_result *GetLabelTableResponse, _err error) {
	req := &openapi.OpenApiRequest{
		Headers: headers,
	}
	params := &openapi.Params{
		Action:      tea.String("GetLabelTable"),
		Version:     tea.String("2023-06-21"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/api/v1/instances/" + tea.StringValue(openapiutil.GetEncodeParam(InstanceId)) + "/labeltables/" + tea.StringValue(openapiutil.GetEncodeParam(LabelTableId))),
		Method:      tea.String("GET"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("json"),
	}
	_result = &GetLabelTableResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) GetLabelTable(InstanceId *string, LabelTableId *string) (_result *GetLabelTableResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &GetLabelTableResponse{}
	_body, _err := client.GetLabelTableWithOptions(InstanceId, LabelTableId, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) GetModelFeatureWithOptions(InstanceId *string, ModelFeatureId *string, headers map[string]*string, runtime *util.RuntimeOptions) (_result *GetModelFeatureResponse, _err error) {
	req := &openapi.OpenApiRequest{
		Headers: headers,
	}
	params := &openapi.Params{
		Action:      tea.String("GetModelFeature"),
		Version:     tea.String("2023-06-21"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/api/v1/instances/" + tea.StringValue(openapiutil.GetEncodeParam(InstanceId)) + "/modelfeatures/" + tea.StringValue(openapiutil.GetEncodeParam(ModelFeatureId))),
		Method:      tea.String("GET"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("json"),
	}
	_result = &GetModelFeatureResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) GetModelFeature(InstanceId *string, ModelFeatureId *string) (_result *GetModelFeatureResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &GetModelFeatureResponse{}
	_body, _err := client.GetModelFeatureWithOptions(InstanceId, ModelFeatureId, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) GetProjectWithOptions(InstanceId *string, ProjectId *string, headers map[string]*string, runtime *util.RuntimeOptions) (_result *GetProjectResponse, _err error) {
	req := &openapi.OpenApiRequest{
		Headers: headers,
	}
	params := &openapi.Params{
		Action:      tea.String("GetProject"),
		Version:     tea.String("2023-06-21"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/api/v1/instances/" + tea.StringValue(openapiutil.GetEncodeParam(InstanceId)) + "/projects/" + tea.StringValue(openapiutil.GetEncodeParam(ProjectId))),
		Method:      tea.String("GET"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("json"),
	}
	_result = &GetProjectResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) GetProject(InstanceId *string, ProjectId *string) (_result *GetProjectResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &GetProjectResponse{}
	_body, _err := client.GetProjectWithOptions(InstanceId, ProjectId, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) GetProjectFeatureEntityWithOptions(InstanceId *string, ProjectId *string, FeatureEntityName *string, headers map[string]*string, runtime *util.RuntimeOptions) (_result *GetProjectFeatureEntityResponse, _err error) {
	req := &openapi.OpenApiRequest{
		Headers: headers,
	}
	params := &openapi.Params{
		Action:      tea.String("GetProjectFeatureEntity"),
		Version:     tea.String("2023-06-21"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/api/v1/instances/" + tea.StringValue(openapiutil.GetEncodeParam(InstanceId)) + "/projects/" + tea.StringValue(openapiutil.GetEncodeParam(ProjectId)) + "/featureentities/" + tea.StringValue(openapiutil.GetEncodeParam(FeatureEntityName))),
		Method:      tea.String("GET"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("json"),
	}
	_result = &GetProjectFeatureEntityResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) GetProjectFeatureEntity(InstanceId *string, ProjectId *string, FeatureEntityName *string) (_result *GetProjectFeatureEntityResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &GetProjectFeatureEntityResponse{}
	_body, _err := client.GetProjectFeatureEntityWithOptions(InstanceId, ProjectId, FeatureEntityName, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) GetProjectFeatureEntityHotIdsWithOptions(InstanceId *string, ProjectId *string, NextSeqNumber *string, FeatureEntityName *string, headers map[string]*string, runtime *util.RuntimeOptions) (_result *GetProjectFeatureEntityHotIdsResponse, _err error) {
	req := &openapi.OpenApiRequest{
		Headers: headers,
	}
	params := &openapi.Params{
		Action:      tea.String("GetProjectFeatureEntityHotIds"),
		Version:     tea.String("2023-06-21"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/api/v1/instances/" + tea.StringValue(openapiutil.GetEncodeParam(InstanceId)) + "/projects/" + tea.StringValue(openapiutil.GetEncodeParam(ProjectId)) + "/featureentities/" + tea.StringValue(openapiutil.GetEncodeParam(FeatureEntityName)) + "/hotids/" + tea.StringValue(openapiutil.GetEncodeParam(NextSeqNumber))),
		Method:      tea.String("GET"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("json"),
	}
	_result = &GetProjectFeatureEntityHotIdsResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) GetProjectFeatureEntityHotIds(InstanceId *string, ProjectId *string, NextSeqNumber *string, FeatureEntityName *string) (_result *GetProjectFeatureEntityHotIdsResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &GetProjectFeatureEntityHotIdsResponse{}
	_body, _err := client.GetProjectFeatureEntityHotIdsWithOptions(InstanceId, ProjectId, NextSeqNumber, FeatureEntityName, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) GetProjectFeatureViewWithOptions(InstanceId *string, ProjectId *string, FeatureViewName *string, headers map[string]*string, runtime *util.RuntimeOptions) (_result *GetProjectFeatureViewResponse, _err error) {
	req := &openapi.OpenApiRequest{
		Headers: headers,
	}
	params := &openapi.Params{
		Action:      tea.String("GetProjectFeatureView"),
		Version:     tea.String("2023-06-21"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/api/v1/instances/" + tea.StringValue(openapiutil.GetEncodeParam(InstanceId)) + "/projects/" + tea.StringValue(openapiutil.GetEncodeParam(ProjectId)) + "/featureviews/" + tea.StringValue(openapiutil.GetEncodeParam(FeatureViewName))),
		Method:      tea.String("GET"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("json"),
	}
	_result = &GetProjectFeatureViewResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) GetProjectFeatureView(InstanceId *string, ProjectId *string, FeatureViewName *string) (_result *GetProjectFeatureViewResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &GetProjectFeatureViewResponse{}
	_body, _err := client.GetProjectFeatureViewWithOptions(InstanceId, ProjectId, FeatureViewName, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) GetProjectModelFeatureWithOptions(InstanceId *string, ProjectId *string, ModelFeatureName *string, headers map[string]*string, runtime *util.RuntimeOptions) (_result *GetProjectModelFeatureResponse, _err error) {
	req := &openapi.OpenApiRequest{
		Headers: headers,
	}
	params := &openapi.Params{
		Action:      tea.String("GetProjectModelFeature"),
		Version:     tea.String("2023-06-21"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/api/v1/instances/" + tea.StringValue(openapiutil.GetEncodeParam(InstanceId)) + "/projects/" + tea.StringValue(openapiutil.GetEncodeParam(ProjectId)) + "/modelfeatures/" + tea.StringValue(openapiutil.GetEncodeParam(ModelFeatureName))),
		Method:      tea.String("GET"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("json"),
	}
	_result = &GetProjectModelFeatureResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) GetProjectModelFeature(InstanceId *string, ProjectId *string, ModelFeatureName *string) (_result *GetProjectModelFeatureResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &GetProjectModelFeatureResponse{}
	_body, _err := client.GetProjectModelFeatureWithOptions(InstanceId, ProjectId, ModelFeatureName, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) GetServiceIdentityRoleWithOptions(RoleName *string, headers map[string]*string, runtime *util.RuntimeOptions) (_result *GetServiceIdentityRoleResponse, _err error) {
	req := &openapi.OpenApiRequest{
		Headers: headers,
	}
	params := &openapi.Params{
		Action:      tea.String("GetServiceIdentityRole"),
		Version:     tea.String("2023-06-21"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/api/v1/serviceidentityroles/" + tea.StringValue(openapiutil.GetEncodeParam(RoleName))),
		Method:      tea.String("GET"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("json"),
	}
	_result = &GetServiceIdentityRoleResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) GetServiceIdentityRole(RoleName *string) (_result *GetServiceIdentityRoleResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &GetServiceIdentityRoleResponse{}
	_body, _err := client.GetServiceIdentityRoleWithOptions(RoleName, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) GetTaskWithOptions(InstanceId *string, TaskId *string, headers map[string]*string, runtime *util.RuntimeOptions) (_result *GetTaskResponse, _err error) {
	req := &openapi.OpenApiRequest{
		Headers: headers,
	}
	params := &openapi.Params{
		Action:      tea.String("GetTask"),
		Version:     tea.String("2023-06-21"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/api/v1/instances/" + tea.StringValue(openapiutil.GetEncodeParam(InstanceId)) + "/tasks/" + tea.StringValue(openapiutil.GetEncodeParam(TaskId))),
		Method:      tea.String("GET"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("json"),
	}
	_result = &GetTaskResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) GetTask(InstanceId *string, TaskId *string) (_result *GetTaskResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &GetTaskResponse{}
	_body, _err := client.GetTaskWithOptions(InstanceId, TaskId, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) ListDatasourceTablesWithOptions(InstanceId *string, DatasourceId *string, request *ListDatasourceTablesRequest, headers map[string]*string, runtime *util.RuntimeOptions) (_result *ListDatasourceTablesResponse, _err error) {
	_err = util.ValidateModel(request)
	if _err != nil {
		return _result, _err
	}
	query := map[string]interface{}{}
	if !tea.BoolValue(util.IsUnset(request.TableName)) {
		query["TableName"] = request.TableName
	}

	req := &openapi.OpenApiRequest{
		Headers: headers,
		Query:   openapiutil.Query(query),
	}
	params := &openapi.Params{
		Action:      tea.String("ListDatasourceTables"),
		Version:     tea.String("2023-06-21"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/api/v1/instances/" + tea.StringValue(openapiutil.GetEncodeParam(InstanceId)) + "/datasources/" + tea.StringValue(openapiutil.GetEncodeParam(DatasourceId)) + "/tables"),
		Method:      tea.String("GET"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("json"),
	}
	_result = &ListDatasourceTablesResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) ListDatasourceTables(InstanceId *string, DatasourceId *string, request *ListDatasourceTablesRequest) (_result *ListDatasourceTablesResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &ListDatasourceTablesResponse{}
	_body, _err := client.ListDatasourceTablesWithOptions(InstanceId, DatasourceId, request, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) ListDatasourcesWithOptions(InstanceId *string, request *ListDatasourcesRequest, headers map[string]*string, runtime *util.RuntimeOptions) (_result *ListDatasourcesResponse, _err error) {
	_err = util.ValidateModel(request)
	if _err != nil {
		return _result, _err
	}
	query := map[string]interface{}{}
	if !tea.BoolValue(util.IsUnset(request.Name)) {
		query["Name"] = request.Name
	}

	if !tea.BoolValue(util.IsUnset(request.Order)) {
		query["Order"] = request.Order
	}

	if !tea.BoolValue(util.IsUnset(request.PageNumber)) {
		query["PageNumber"] = request.PageNumber
	}

	if !tea.BoolValue(util.IsUnset(request.PageSize)) {
		query["PageSize"] = request.PageSize
	}

	if !tea.BoolValue(util.IsUnset(request.SortBy)) {
		query["SortBy"] = request.SortBy
	}

	if !tea.BoolValue(util.IsUnset(request.Type)) {
		query["Type"] = request.Type
	}

	if !tea.BoolValue(util.IsUnset(request.WorkspaceId)) {
		query["WorkspaceId"] = request.WorkspaceId
	}

	req := &openapi.OpenApiRequest{
		Headers: headers,
		Query:   openapiutil.Query(query),
	}
	params := &openapi.Params{
		Action:      tea.String("ListDatasources"),
		Version:     tea.String("2023-06-21"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/api/v1/instances/" + tea.StringValue(openapiutil.GetEncodeParam(InstanceId)) + "/datasources"),
		Method:      tea.String("GET"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("json"),
	}
	_result = &ListDatasourcesResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) ListDatasources(InstanceId *string, request *ListDatasourcesRequest) (_result *ListDatasourcesResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &ListDatasourcesResponse{}
	_body, _err := client.ListDatasourcesWithOptions(InstanceId, request, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) ListFeatureEntitiesWithOptions(InstanceId *string, tmpReq *ListFeatureEntitiesRequest, headers map[string]*string, runtime *util.RuntimeOptions) (_result *ListFeatureEntitiesResponse, _err error) {
	_err = util.ValidateModel(tmpReq)
	if _err != nil {
		return _result, _err
	}
	request := &ListFeatureEntitiesShrinkRequest{}
	openapiutil.Convert(tmpReq, request)
	if !tea.BoolValue(util.IsUnset(tmpReq.FeatureEntityIds)) {
		request.FeatureEntityIdsShrink = openapiutil.ArrayToStringWithSpecifiedStyle(tmpReq.FeatureEntityIds, tea.String("FeatureEntityIds"), tea.String("simple"))
	}

	query := map[string]interface{}{}
	if !tea.BoolValue(util.IsUnset(request.FeatureEntityIdsShrink)) {
		query["FeatureEntityIds"] = request.FeatureEntityIdsShrink
	}

	if !tea.BoolValue(util.IsUnset(request.Name)) {
		query["Name"] = request.Name
	}

	if !tea.BoolValue(util.IsUnset(request.Order)) {
		query["Order"] = request.Order
	}

	if !tea.BoolValue(util.IsUnset(request.Owner)) {
		query["Owner"] = request.Owner
	}

	if !tea.BoolValue(util.IsUnset(request.PageNumber)) {
		query["PageNumber"] = request.PageNumber
	}

	if !tea.BoolValue(util.IsUnset(request.PageSize)) {
		query["PageSize"] = request.PageSize
	}

	if !tea.BoolValue(util.IsUnset(request.ProjectId)) {
		query["ProjectId"] = request.ProjectId
	}

	if !tea.BoolValue(util.IsUnset(request.SortBy)) {
		query["SortBy"] = request.SortBy
	}

	req := &openapi.OpenApiRequest{
		Headers: headers,
		Query:   openapiutil.Query(query),
	}
	params := &openapi.Params{
		Action:      tea.String("ListFeatureEntities"),
		Version:     tea.String("2023-06-21"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/api/v1/instances/" + tea.StringValue(openapiutil.GetEncodeParam(InstanceId)) + "/featureentities"),
		Method:      tea.String("GET"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("json"),
	}
	_result = &ListFeatureEntitiesResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) ListFeatureEntities(InstanceId *string, request *ListFeatureEntitiesRequest) (_result *ListFeatureEntitiesResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &ListFeatureEntitiesResponse{}
	_body, _err := client.ListFeatureEntitiesWithOptions(InstanceId, request, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) ListFeatureViewFieldRelationshipsWithOptions(InstanceId *string, FeatureViewId *string, FieldName *string, headers map[string]*string, runtime *util.RuntimeOptions) (_result *ListFeatureViewFieldRelationshipsResponse, _err error) {
	req := &openapi.OpenApiRequest{
		Headers: headers,
	}
	params := &openapi.Params{
		Action:      tea.String("ListFeatureViewFieldRelationships"),
		Version:     tea.String("2023-06-21"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/api/v1/instances/" + tea.StringValue(openapiutil.GetEncodeParam(InstanceId)) + "/featureviews/" + tea.StringValue(openapiutil.GetEncodeParam(FeatureViewId)) + "/fields/" + tea.StringValue(openapiutil.GetEncodeParam(FieldName)) + "/relationships"),
		Method:      tea.String("GET"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("json"),
	}
	_result = &ListFeatureViewFieldRelationshipsResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) ListFeatureViewFieldRelationships(InstanceId *string, FeatureViewId *string, FieldName *string) (_result *ListFeatureViewFieldRelationshipsResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &ListFeatureViewFieldRelationshipsResponse{}
	_body, _err := client.ListFeatureViewFieldRelationshipsWithOptions(InstanceId, FeatureViewId, FieldName, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) ListFeatureViewRelationshipsWithOptions(InstanceId *string, FeatureViewId *string, headers map[string]*string, runtime *util.RuntimeOptions) (_result *ListFeatureViewRelationshipsResponse, _err error) {
	req := &openapi.OpenApiRequest{
		Headers: headers,
	}
	params := &openapi.Params{
		Action:      tea.String("ListFeatureViewRelationships"),
		Version:     tea.String("2023-06-21"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/api/v1/instances/" + tea.StringValue(openapiutil.GetEncodeParam(InstanceId)) + "/featureviews/" + tea.StringValue(openapiutil.GetEncodeParam(FeatureViewId)) + "/relationships"),
		Method:      tea.String("GET"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("json"),
	}
	_result = &ListFeatureViewRelationshipsResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) ListFeatureViewRelationships(InstanceId *string, FeatureViewId *string) (_result *ListFeatureViewRelationshipsResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &ListFeatureViewRelationshipsResponse{}
	_body, _err := client.ListFeatureViewRelationshipsWithOptions(InstanceId, FeatureViewId, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) ListFeatureViewsWithOptions(InstanceId *string, tmpReq *ListFeatureViewsRequest, headers map[string]*string, runtime *util.RuntimeOptions) (_result *ListFeatureViewsResponse, _err error) {
	_err = util.ValidateModel(tmpReq)
	if _err != nil {
		return _result, _err
	}
	request := &ListFeatureViewsShrinkRequest{}
	openapiutil.Convert(tmpReq, request)
	if !tea.BoolValue(util.IsUnset(tmpReq.FeatureViewIds)) {
		request.FeatureViewIdsShrink = openapiutil.ArrayToStringWithSpecifiedStyle(tmpReq.FeatureViewIds, tea.String("FeatureViewIds"), tea.String("simple"))
	}

	query := map[string]interface{}{}
	if !tea.BoolValue(util.IsUnset(request.FeatureName)) {
		query["FeatureName"] = request.FeatureName
	}

	if !tea.BoolValue(util.IsUnset(request.FeatureViewIdsShrink)) {
		query["FeatureViewIds"] = request.FeatureViewIdsShrink
	}

	if !tea.BoolValue(util.IsUnset(request.Order)) {
		query["Order"] = request.Order
	}

	if !tea.BoolValue(util.IsUnset(request.Owner)) {
		query["Owner"] = request.Owner
	}

	if !tea.BoolValue(util.IsUnset(request.PageNumber)) {
		query["PageNumber"] = request.PageNumber
	}

	if !tea.BoolValue(util.IsUnset(request.PageSize)) {
		query["PageSize"] = request.PageSize
	}

	if !tea.BoolValue(util.IsUnset(request.ProjectId)) {
		query["ProjectId"] = request.ProjectId
	}

	if !tea.BoolValue(util.IsUnset(request.SortBy)) {
		query["SortBy"] = request.SortBy
	}

	if !tea.BoolValue(util.IsUnset(request.Tag)) {
		query["Tag"] = request.Tag
	}

	if !tea.BoolValue(util.IsUnset(request.Type)) {
		query["Type"] = request.Type
	}

	req := &openapi.OpenApiRequest{
		Headers: headers,
		Query:   openapiutil.Query(query),
	}
	params := &openapi.Params{
		Action:      tea.String("ListFeatureViews"),
		Version:     tea.String("2023-06-21"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/api/v1/instances/" + tea.StringValue(openapiutil.GetEncodeParam(InstanceId)) + "/featureviews"),
		Method:      tea.String("GET"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("json"),
	}
	_result = &ListFeatureViewsResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) ListFeatureViews(InstanceId *string, request *ListFeatureViewsRequest) (_result *ListFeatureViewsResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &ListFeatureViewsResponse{}
	_body, _err := client.ListFeatureViewsWithOptions(InstanceId, request, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) ListInstancesWithOptions(request *ListInstancesRequest, headers map[string]*string, runtime *util.RuntimeOptions) (_result *ListInstancesResponse, _err error) {
	_err = util.ValidateModel(request)
	if _err != nil {
		return _result, _err
	}
	query := map[string]interface{}{}
	if !tea.BoolValue(util.IsUnset(request.Order)) {
		query["Order"] = request.Order
	}

	if !tea.BoolValue(util.IsUnset(request.PageNumber)) {
		query["PageNumber"] = request.PageNumber
	}

	if !tea.BoolValue(util.IsUnset(request.PageSize)) {
		query["PageSize"] = request.PageSize
	}

	if !tea.BoolValue(util.IsUnset(request.SortBy)) {
		query["SortBy"] = request.SortBy
	}

	if !tea.BoolValue(util.IsUnset(request.Status)) {
		query["Status"] = request.Status
	}

	req := &openapi.OpenApiRequest{
		Headers: headers,
		Query:   openapiutil.Query(query),
	}
	params := &openapi.Params{
		Action:      tea.String("ListInstances"),
		Version:     tea.String("2023-06-21"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/api/v1/instances"),
		Method:      tea.String("GET"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("json"),
	}
	_result = &ListInstancesResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) ListInstances(request *ListInstancesRequest) (_result *ListInstancesResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &ListInstancesResponse{}
	_body, _err := client.ListInstancesWithOptions(request, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) ListLabelTablesWithOptions(InstanceId *string, tmpReq *ListLabelTablesRequest, headers map[string]*string, runtime *util.RuntimeOptions) (_result *ListLabelTablesResponse, _err error) {
	_err = util.ValidateModel(tmpReq)
	if _err != nil {
		return _result, _err
	}
	request := &ListLabelTablesShrinkRequest{}
	openapiutil.Convert(tmpReq, request)
	if !tea.BoolValue(util.IsUnset(tmpReq.LabelTableIds)) {
		request.LabelTableIdsShrink = openapiutil.ArrayToStringWithSpecifiedStyle(tmpReq.LabelTableIds, tea.String("LabelTableIds"), tea.String("simple"))
	}

	query := map[string]interface{}{}
	if !tea.BoolValue(util.IsUnset(request.LabelTableIdsShrink)) {
		query["LabelTableIds"] = request.LabelTableIdsShrink
	}

	if !tea.BoolValue(util.IsUnset(request.Name)) {
		query["Name"] = request.Name
	}

	if !tea.BoolValue(util.IsUnset(request.Order)) {
		query["Order"] = request.Order
	}

	if !tea.BoolValue(util.IsUnset(request.Owner)) {
		query["Owner"] = request.Owner
	}

	if !tea.BoolValue(util.IsUnset(request.PageNumber)) {
		query["PageNumber"] = request.PageNumber
	}

	if !tea.BoolValue(util.IsUnset(request.PageSize)) {
		query["PageSize"] = request.PageSize
	}

	if !tea.BoolValue(util.IsUnset(request.ProjectId)) {
		query["ProjectId"] = request.ProjectId
	}

	if !tea.BoolValue(util.IsUnset(request.SortBy)) {
		query["SortBy"] = request.SortBy
	}

	req := &openapi.OpenApiRequest{
		Headers: headers,
		Query:   openapiutil.Query(query),
	}
	params := &openapi.Params{
		Action:      tea.String("ListLabelTables"),
		Version:     tea.String("2023-06-21"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/api/v1/instances/" + tea.StringValue(openapiutil.GetEncodeParam(InstanceId)) + "/labeltables"),
		Method:      tea.String("GET"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("json"),
	}
	_result = &ListLabelTablesResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) ListLabelTables(InstanceId *string, request *ListLabelTablesRequest) (_result *ListLabelTablesResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &ListLabelTablesResponse{}
	_body, _err := client.ListLabelTablesWithOptions(InstanceId, request, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) ListModelFeaturesWithOptions(InstanceId *string, tmpReq *ListModelFeaturesRequest, headers map[string]*string, runtime *util.RuntimeOptions) (_result *ListModelFeaturesResponse, _err error) {
	_err = util.ValidateModel(tmpReq)
	if _err != nil {
		return _result, _err
	}
	request := &ListModelFeaturesShrinkRequest{}
	openapiutil.Convert(tmpReq, request)
	if !tea.BoolValue(util.IsUnset(tmpReq.ModelFeatureIds)) {
		request.ModelFeatureIdsShrink = openapiutil.ArrayToStringWithSpecifiedStyle(tmpReq.ModelFeatureIds, tea.String("ModelFeatureIds"), tea.String("simple"))
	}

	query := map[string]interface{}{}
	if !tea.BoolValue(util.IsUnset(request.ModelFeatureIdsShrink)) {
		query["ModelFeatureIds"] = request.ModelFeatureIdsShrink
	}

	if !tea.BoolValue(util.IsUnset(request.Name)) {
		query["Name"] = request.Name
	}

	if !tea.BoolValue(util.IsUnset(request.Order)) {
		query["Order"] = request.Order
	}

	if !tea.BoolValue(util.IsUnset(request.Owner)) {
		query["Owner"] = request.Owner
	}

	if !tea.BoolValue(util.IsUnset(request.PageNumber)) {
		query["PageNumber"] = request.PageNumber
	}

	if !tea.BoolValue(util.IsUnset(request.PageSize)) {
		query["PageSize"] = request.PageSize
	}

	if !tea.BoolValue(util.IsUnset(request.ProjectId)) {
		query["ProjectId"] = request.ProjectId
	}

	if !tea.BoolValue(util.IsUnset(request.SortBy)) {
		query["SortBy"] = request.SortBy
	}

	req := &openapi.OpenApiRequest{
		Headers: headers,
		Query:   openapiutil.Query(query),
	}
	params := &openapi.Params{
		Action:      tea.String("ListModelFeatures"),
		Version:     tea.String("2023-06-21"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/api/v1/instances/" + tea.StringValue(openapiutil.GetEncodeParam(InstanceId)) + "/modelfeatures"),
		Method:      tea.String("GET"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("json"),
	}
	_result = &ListModelFeaturesResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) ListModelFeatures(InstanceId *string, request *ListModelFeaturesRequest) (_result *ListModelFeaturesResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &ListModelFeaturesResponse{}
	_body, _err := client.ListModelFeaturesWithOptions(InstanceId, request, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) ListProjectFeatureViewOwnersWithOptions(InstanceId *string, ProjectId *string, headers map[string]*string, runtime *util.RuntimeOptions) (_result *ListProjectFeatureViewOwnersResponse, _err error) {
	req := &openapi.OpenApiRequest{
		Headers: headers,
	}
	params := &openapi.Params{
		Action:      tea.String("ListProjectFeatureViewOwners"),
		Version:     tea.String("2023-06-21"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/api/v1/instances/" + tea.StringValue(openapiutil.GetEncodeParam(InstanceId)) + "/projects/" + tea.StringValue(openapiutil.GetEncodeParam(ProjectId)) + "/featureviewowners"),
		Method:      tea.String("GET"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("json"),
	}
	_result = &ListProjectFeatureViewOwnersResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) ListProjectFeatureViewOwners(InstanceId *string, ProjectId *string) (_result *ListProjectFeatureViewOwnersResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &ListProjectFeatureViewOwnersResponse{}
	_body, _err := client.ListProjectFeatureViewOwnersWithOptions(InstanceId, ProjectId, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) ListProjectFeatureViewTagsWithOptions(InstanceId *string, ProjectId *string, headers map[string]*string, runtime *util.RuntimeOptions) (_result *ListProjectFeatureViewTagsResponse, _err error) {
	req := &openapi.OpenApiRequest{
		Headers: headers,
	}
	params := &openapi.Params{
		Action:      tea.String("ListProjectFeatureViewTags"),
		Version:     tea.String("2023-06-21"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/api/v1/instances/" + tea.StringValue(openapiutil.GetEncodeParam(InstanceId)) + "/projects/" + tea.StringValue(openapiutil.GetEncodeParam(ProjectId)) + "/featureviewtags"),
		Method:      tea.String("GET"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("json"),
	}
	_result = &ListProjectFeatureViewTagsResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) ListProjectFeatureViewTags(InstanceId *string, ProjectId *string) (_result *ListProjectFeatureViewTagsResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &ListProjectFeatureViewTagsResponse{}
	_body, _err := client.ListProjectFeatureViewTagsWithOptions(InstanceId, ProjectId, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) ListProjectFeatureViewsWithOptions(InstanceId *string, ProjectId *string, headers map[string]*string, runtime *util.RuntimeOptions) (_result *ListProjectFeatureViewsResponse, _err error) {
	req := &openapi.OpenApiRequest{
		Headers: headers,
	}
	params := &openapi.Params{
		Action:      tea.String("ListProjectFeatureViews"),
		Version:     tea.String("2023-06-21"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/api/v1/instances/" + tea.StringValue(openapiutil.GetEncodeParam(InstanceId)) + "/projects/" + tea.StringValue(openapiutil.GetEncodeParam(ProjectId)) + "/featureviews"),
		Method:      tea.String("GET"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("json"),
	}
	_result = &ListProjectFeatureViewsResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) ListProjectFeatureViews(InstanceId *string, ProjectId *string) (_result *ListProjectFeatureViewsResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &ListProjectFeatureViewsResponse{}
	_body, _err := client.ListProjectFeatureViewsWithOptions(InstanceId, ProjectId, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) ListProjectsWithOptions(InstanceId *string, tmpReq *ListProjectsRequest, headers map[string]*string, runtime *util.RuntimeOptions) (_result *ListProjectsResponse, _err error) {
	_err = util.ValidateModel(tmpReq)
	if _err != nil {
		return _result, _err
	}
	request := &ListProjectsShrinkRequest{}
	openapiutil.Convert(tmpReq, request)
	if !tea.BoolValue(util.IsUnset(tmpReq.ProjectIds)) {
		request.ProjectIdsShrink = openapiutil.ArrayToStringWithSpecifiedStyle(tmpReq.ProjectIds, tea.String("ProjectIds"), tea.String("simple"))
	}

	query := map[string]interface{}{}
	if !tea.BoolValue(util.IsUnset(request.Name)) {
		query["Name"] = request.Name
	}

	if !tea.BoolValue(util.IsUnset(request.Order)) {
		query["Order"] = request.Order
	}

	if !tea.BoolValue(util.IsUnset(request.Owner)) {
		query["Owner"] = request.Owner
	}

	if !tea.BoolValue(util.IsUnset(request.PageNumber)) {
		query["PageNumber"] = request.PageNumber
	}

	if !tea.BoolValue(util.IsUnset(request.PageSize)) {
		query["PageSize"] = request.PageSize
	}

	if !tea.BoolValue(util.IsUnset(request.ProjectIdsShrink)) {
		query["ProjectIds"] = request.ProjectIdsShrink
	}

	if !tea.BoolValue(util.IsUnset(request.SortBy)) {
		query["SortBy"] = request.SortBy
	}

	if !tea.BoolValue(util.IsUnset(request.WorkspaceId)) {
		query["WorkspaceId"] = request.WorkspaceId
	}

	req := &openapi.OpenApiRequest{
		Headers: headers,
		Query:   openapiutil.Query(query),
	}
	params := &openapi.Params{
		Action:      tea.String("ListProjects"),
		Version:     tea.String("2023-06-21"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/api/v1/instances/" + tea.StringValue(openapiutil.GetEncodeParam(InstanceId)) + "/projects"),
		Method:      tea.String("GET"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("json"),
	}
	_result = &ListProjectsResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) ListProjects(InstanceId *string, request *ListProjectsRequest) (_result *ListProjectsResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &ListProjectsResponse{}
	_body, _err := client.ListProjectsWithOptions(InstanceId, request, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) ListTaskLogsWithOptions(InstanceId *string, TaskId *string, request *ListTaskLogsRequest, headers map[string]*string, runtime *util.RuntimeOptions) (_result *ListTaskLogsResponse, _err error) {
	_err = util.ValidateModel(request)
	if _err != nil {
		return _result, _err
	}
	query := map[string]interface{}{}
	if !tea.BoolValue(util.IsUnset(request.PageNumber)) {
		query["PageNumber"] = request.PageNumber
	}

	if !tea.BoolValue(util.IsUnset(request.PageSize)) {
		query["PageSize"] = request.PageSize
	}

	req := &openapi.OpenApiRequest{
		Headers: headers,
		Query:   openapiutil.Query(query),
	}
	params := &openapi.Params{
		Action:      tea.String("ListTaskLogs"),
		Version:     tea.String("2023-06-21"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/api/v1/instances/" + tea.StringValue(openapiutil.GetEncodeParam(InstanceId)) + "/tasks/" + tea.StringValue(openapiutil.GetEncodeParam(TaskId)) + "/logs"),
		Method:      tea.String("GET"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("json"),
	}
	_result = &ListTaskLogsResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) ListTaskLogs(InstanceId *string, TaskId *string, request *ListTaskLogsRequest) (_result *ListTaskLogsResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &ListTaskLogsResponse{}
	_body, _err := client.ListTaskLogsWithOptions(InstanceId, TaskId, request, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) ListTasksWithOptions(InstanceId *string, tmpReq *ListTasksRequest, headers map[string]*string, runtime *util.RuntimeOptions) (_result *ListTasksResponse, _err error) {
	_err = util.ValidateModel(tmpReq)
	if _err != nil {
		return _result, _err
	}
	request := &ListTasksShrinkRequest{}
	openapiutil.Convert(tmpReq, request)
	if !tea.BoolValue(util.IsUnset(tmpReq.TaskIds)) {
		request.TaskIdsShrink = openapiutil.ArrayToStringWithSpecifiedStyle(tmpReq.TaskIds, tea.String("TaskIds"), tea.String("simple"))
	}

	query := map[string]interface{}{}
	if !tea.BoolValue(util.IsUnset(request.ObjectId)) {
		query["ObjectId"] = request.ObjectId
	}

	if !tea.BoolValue(util.IsUnset(request.ObjectType)) {
		query["ObjectType"] = request.ObjectType
	}

	if !tea.BoolValue(util.IsUnset(request.PageNumber)) {
		query["PageNumber"] = request.PageNumber
	}

	if !tea.BoolValue(util.IsUnset(request.PageSize)) {
		query["PageSize"] = request.PageSize
	}

	if !tea.BoolValue(util.IsUnset(request.ProjectId)) {
		query["ProjectId"] = request.ProjectId
	}

	if !tea.BoolValue(util.IsUnset(request.Status)) {
		query["Status"] = request.Status
	}

	if !tea.BoolValue(util.IsUnset(request.TaskIdsShrink)) {
		query["TaskIds"] = request.TaskIdsShrink
	}

	if !tea.BoolValue(util.IsUnset(request.Type)) {
		query["Type"] = request.Type
	}

	req := &openapi.OpenApiRequest{
		Headers: headers,
		Query:   openapiutil.Query(query),
	}
	params := &openapi.Params{
		Action:      tea.String("ListTasks"),
		Version:     tea.String("2023-06-21"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/api/v1/instances/" + tea.StringValue(openapiutil.GetEncodeParam(InstanceId)) + "/tasks"),
		Method:      tea.String("GET"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("json"),
	}
	_result = &ListTasksResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) ListTasks(InstanceId *string, request *ListTasksRequest) (_result *ListTasksResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &ListTasksResponse{}
	_body, _err := client.ListTasksWithOptions(InstanceId, request, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) PublishFeatureViewTableWithOptions(InstanceId *string, FeatureViewId *string, request *PublishFeatureViewTableRequest, headers map[string]*string, runtime *util.RuntimeOptions) (_result *PublishFeatureViewTableResponse, _err error) {
	_err = util.ValidateModel(request)
	if _err != nil {
		return _result, _err
	}
	body := map[string]interface{}{}
	if !tea.BoolValue(util.IsUnset(request.Config)) {
		body["Config"] = request.Config
	}

	if !tea.BoolValue(util.IsUnset(request.EventTime)) {
		body["EventTime"] = request.EventTime
	}

	if !tea.BoolValue(util.IsUnset(request.Mode)) {
		body["Mode"] = request.Mode
	}

	if !tea.BoolValue(util.IsUnset(request.OfflineToOnline)) {
		body["OfflineToOnline"] = request.OfflineToOnline
	}

	if !tea.BoolValue(util.IsUnset(request.Partitions)) {
		body["Partitions"] = request.Partitions
	}

	req := &openapi.OpenApiRequest{
		Headers: headers,
		Body:    openapiutil.ParseToMap(body),
	}
	params := &openapi.Params{
		Action:      tea.String("PublishFeatureViewTable"),
		Version:     tea.String("2023-06-21"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/api/v1/instances/" + tea.StringValue(openapiutil.GetEncodeParam(InstanceId)) + "/featureviews/" + tea.StringValue(openapiutil.GetEncodeParam(FeatureViewId)) + "/action/publishtable"),
		Method:      tea.String("POST"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("json"),
	}
	_result = &PublishFeatureViewTableResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) PublishFeatureViewTable(InstanceId *string, FeatureViewId *string, request *PublishFeatureViewTableRequest) (_result *PublishFeatureViewTableResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &PublishFeatureViewTableResponse{}
	_body, _err := client.PublishFeatureViewTableWithOptions(InstanceId, FeatureViewId, request, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) UpdateDatasourceWithOptions(InstanceId *string, DatasourceId *string, request *UpdateDatasourceRequest, headers map[string]*string, runtime *util.RuntimeOptions) (_result *UpdateDatasourceResponse, _err error) {
	_err = util.ValidateModel(request)
	if _err != nil {
		return _result, _err
	}
	body := map[string]interface{}{}
	if !tea.BoolValue(util.IsUnset(request.Config)) {
		body["Config"] = request.Config
	}

	if !tea.BoolValue(util.IsUnset(request.Name)) {
		body["Name"] = request.Name
	}

	if !tea.BoolValue(util.IsUnset(request.Uri)) {
		body["Uri"] = request.Uri
	}

	req := &openapi.OpenApiRequest{
		Headers: headers,
		Body:    openapiutil.ParseToMap(body),
	}
	params := &openapi.Params{
		Action:      tea.String("UpdateDatasource"),
		Version:     tea.String("2023-06-21"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/api/v1/instances/" + tea.StringValue(openapiutil.GetEncodeParam(InstanceId)) + "/datasources/" + tea.StringValue(openapiutil.GetEncodeParam(DatasourceId))),
		Method:      tea.String("PUT"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("json"),
	}
	_result = &UpdateDatasourceResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) UpdateDatasource(InstanceId *string, DatasourceId *string, request *UpdateDatasourceRequest) (_result *UpdateDatasourceResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &UpdateDatasourceResponse{}
	_body, _err := client.UpdateDatasourceWithOptions(InstanceId, DatasourceId, request, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) UpdateLabelTableWithOptions(InstanceId *string, LabelTableId *string, request *UpdateLabelTableRequest, headers map[string]*string, runtime *util.RuntimeOptions) (_result *UpdateLabelTableResponse, _err error) {
	_err = util.ValidateModel(request)
	if _err != nil {
		return _result, _err
	}
	body := map[string]interface{}{}
	if !tea.BoolValue(util.IsUnset(request.DatasourceId)) {
		body["DatasourceId"] = request.DatasourceId
	}

	if !tea.BoolValue(util.IsUnset(request.Fields)) {
		body["Fields"] = request.Fields
	}

	if !tea.BoolValue(util.IsUnset(request.Name)) {
		body["Name"] = request.Name
	}

	req := &openapi.OpenApiRequest{
		Headers: headers,
		Body:    openapiutil.ParseToMap(body),
	}
	params := &openapi.Params{
		Action:      tea.String("UpdateLabelTable"),
		Version:     tea.String("2023-06-21"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/api/v1/instances/" + tea.StringValue(openapiutil.GetEncodeParam(InstanceId)) + "/labeltables/" + tea.StringValue(openapiutil.GetEncodeParam(LabelTableId))),
		Method:      tea.String("PUT"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("json"),
	}
	_result = &UpdateLabelTableResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) UpdateLabelTable(InstanceId *string, LabelTableId *string, request *UpdateLabelTableRequest) (_result *UpdateLabelTableResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &UpdateLabelTableResponse{}
	_body, _err := client.UpdateLabelTableWithOptions(InstanceId, LabelTableId, request, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) UpdateModelFeatureWithOptions(InstanceId *string, ModelFeatureId *string, request *UpdateModelFeatureRequest, headers map[string]*string, runtime *util.RuntimeOptions) (_result *UpdateModelFeatureResponse, _err error) {
	_err = util.ValidateModel(request)
	if _err != nil {
		return _result, _err
	}
	body := map[string]interface{}{}
	if !tea.BoolValue(util.IsUnset(request.Features)) {
		body["Features"] = request.Features
	}

	if !tea.BoolValue(util.IsUnset(request.LabelTableId)) {
		body["LabelTableId"] = request.LabelTableId
	}

	req := &openapi.OpenApiRequest{
		Headers: headers,
		Body:    openapiutil.ParseToMap(body),
	}
	params := &openapi.Params{
		Action:      tea.String("UpdateModelFeature"),
		Version:     tea.String("2023-06-21"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/api/v1/instances/" + tea.StringValue(openapiutil.GetEncodeParam(InstanceId)) + "/modelfeatures/" + tea.StringValue(openapiutil.GetEncodeParam(ModelFeatureId))),
		Method:      tea.String("PUT"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("json"),
	}
	_result = &UpdateModelFeatureResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) UpdateModelFeature(InstanceId *string, ModelFeatureId *string, request *UpdateModelFeatureRequest) (_result *UpdateModelFeatureResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &UpdateModelFeatureResponse{}
	_body, _err := client.UpdateModelFeatureWithOptions(InstanceId, ModelFeatureId, request, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) UpdateProjectWithOptions(InstanceId *string, ProjectId *string, request *UpdateProjectRequest, headers map[string]*string, runtime *util.RuntimeOptions) (_result *UpdateProjectResponse, _err error) {
	_err = util.ValidateModel(request)
	if _err != nil {
		return _result, _err
	}
	body := map[string]interface{}{}
	if !tea.BoolValue(util.IsUnset(request.Description)) {
		body["Description"] = request.Description
	}

	if !tea.BoolValue(util.IsUnset(request.Name)) {
		body["Name"] = request.Name
	}

	req := &openapi.OpenApiRequest{
		Headers: headers,
		Body:    openapiutil.ParseToMap(body),
	}
	params := &openapi.Params{
		Action:      tea.String("UpdateProject"),
		Version:     tea.String("2023-06-21"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/api/v1/instances/" + tea.StringValue(openapiutil.GetEncodeParam(InstanceId)) + "/projects/" + tea.StringValue(openapiutil.GetEncodeParam(ProjectId))),
		Method:      tea.String("PUT"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("json"),
	}
	_result = &UpdateProjectResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) UpdateProject(InstanceId *string, ProjectId *string, request *UpdateProjectRequest) (_result *UpdateProjectResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &UpdateProjectResponse{}
	_body, _err := client.UpdateProjectWithOptions(InstanceId, ProjectId, request, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) WriteFeatureViewTableWithOptions(InstanceId *string, FeatureViewId *string, request *WriteFeatureViewTableRequest, headers map[string]*string, runtime *util.RuntimeOptions) (_result *WriteFeatureViewTableResponse, _err error) {
	_err = util.ValidateModel(request)
	if _err != nil {
		return _result, _err
	}
	body := map[string]interface{}{}
	if !tea.BoolValue(util.IsUnset(request.Mode)) {
		body["Mode"] = request.Mode
	}

	if !tea.BoolValue(util.IsUnset(request.Partitions)) {
		body["Partitions"] = request.Partitions
	}

	if !tea.BoolValue(util.IsUnset(request.UrlDatasource)) {
		body["UrlDatasource"] = request.UrlDatasource
	}

	req := &openapi.OpenApiRequest{
		Headers: headers,
		Body:    openapiutil.ParseToMap(body),
	}
	params := &openapi.Params{
		Action:      tea.String("WriteFeatureViewTable"),
		Version:     tea.String("2023-06-21"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/api/v1/instances/" + tea.StringValue(openapiutil.GetEncodeParam(InstanceId)) + "/featureviews/" + tea.StringValue(openapiutil.GetEncodeParam(FeatureViewId)) + "/action/writetable"),
		Method:      tea.String("POST"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("json"),
	}
	_result = &WriteFeatureViewTableResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) WriteFeatureViewTable(InstanceId *string, FeatureViewId *string, request *WriteFeatureViewTableRequest) (_result *WriteFeatureViewTableResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &WriteFeatureViewTableResponse{}
	_body, _err := client.WriteFeatureViewTableWithOptions(InstanceId, FeatureViewId, request, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

func (client *Client) WriteProjectFeatureEntityHotIdsWithOptions(InstanceId *string, ProjectId *string, FeatureEntityName *string, request *WriteProjectFeatureEntityHotIdsRequest, headers map[string]*string, runtime *util.RuntimeOptions) (_result *WriteProjectFeatureEntityHotIdsResponse, _err error) {
	_err = util.ValidateModel(request)
	if _err != nil {
		return _result, _err
	}
	body := map[string]interface{}{}
	if !tea.BoolValue(util.IsUnset(request.HotIds)) {
		body["HotIds"] = request.HotIds
	}

	if !tea.BoolValue(util.IsUnset(request.Version)) {
		body["Version"] = request.Version
	}

	req := &openapi.OpenApiRequest{
		Headers: headers,
		Body:    openapiutil.ParseToMap(body),
	}
	params := &openapi.Params{
		Action:      tea.String("WriteProjectFeatureEntityHotIds"),
		Version:     tea.String("2023-06-21"),
		Protocol:    tea.String("HTTPS"),
		Pathname:    tea.String("/api/v1/instances/" + tea.StringValue(openapiutil.GetEncodeParam(InstanceId)) + "/projects/" + tea.StringValue(openapiutil.GetEncodeParam(ProjectId)) + "/featureentities/" + tea.StringValue(openapiutil.GetEncodeParam(FeatureEntityName)) + "/action/writehotids"),
		Method:      tea.String("POST"),
		AuthType:    tea.String("AK"),
		Style:       tea.String("ROA"),
		ReqBodyType: tea.String("json"),
		BodyType:    tea.String("json"),
	}
	_result = &WriteProjectFeatureEntityHotIdsResponse{}
	_body, _err := client.CallApi(params, req, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

func (client *Client) WriteProjectFeatureEntityHotIds(InstanceId *string, ProjectId *string, FeatureEntityName *string, request *WriteProjectFeatureEntityHotIdsRequest) (_result *WriteProjectFeatureEntityHotIdsResponse, _err error) {
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	_result = &WriteProjectFeatureEntityHotIdsResponse{}
	_body, _err := client.WriteProjectFeatureEntityHotIdsWithOptions(InstanceId, ProjectId, FeatureEntityName, request, headers, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}
