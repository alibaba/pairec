// This file is auto-generated, don't edit it. Thanks.
package ha3client

import (
	encodeutil "github.com/alibabacloud-go/darabonba-encode-util/client"
	map_ "github.com/alibabacloud-go/darabonba-map/client"
	string_ "github.com/alibabacloud-go/darabonba-string/client"
	util "github.com/alibabacloud-go/tea-utils/service"
	"github.com/alibabacloud-go/tea/tea"
)

type Config struct {
	Endpoint       *string `json:"endpoint,omitempty" xml:"endpoint,omitempty"`
	InstanceId     *string `json:"instanceId,omitempty" xml:"instanceId,omitempty"`
	Protocol       *string `json:"protocol,omitempty" xml:"protocol,omitempty"`
	AccessUserName *string `json:"accessUserName,omitempty" xml:"accessUserName,omitempty"`
	AccessPassWord *string `json:"accessPassWord,omitempty" xml:"accessPassWord,omitempty"`
	UserAgent      *string `json:"userAgent,omitempty" xml:"userAgent,omitempty"`
}

func (s Config) String() string {
	return tea.Prettify(s)
}

func (s Config) GoString() string {
	return s.String()
}

func (s *Config) SetEndpoint(v string) *Config {
	s.Endpoint = &v
	return s
}

func (s *Config) SetInstanceId(v string) *Config {
	s.InstanceId = &v
	return s
}

func (s *Config) SetProtocol(v string) *Config {
	s.Protocol = &v
	return s
}

func (s *Config) SetAccessUserName(v string) *Config {
	s.AccessUserName = &v
	return s
}

func (s *Config) SetAccessPassWord(v string) *Config {
	s.AccessPassWord = &v
	return s
}

func (s *Config) SetUserAgent(v string) *Config {
	s.UserAgent = &v
	return s
}

type HaQuery struct {
	// 搜索主体，不能为空.并且可以指定多个查询条件及其之间的关系.
	Query *string `json:"query,omitempty" xml:"query,omitempty" require:"true"`
	// cluster部分用于指定要查询的集群的名字。它不仅可以同时指定多个集群，还可以指定到集群中的哪些partition获取结果。
	Cluster *string `json:"cluster,omitempty" xml:"cluster,omitempty"`
	// config部分可以指定查询结果的起始位置、返回结果的数量、展现结果的格式、参与精排表达式文档个数等。
	Config *HaQueryconfigClause `json:"config,omitempty" xml:"config,omitempty" require:"true"`
	// 过滤功能支持用户根据查询条件，筛选出用户感兴趣的文档。会在通过query子句查找到的文档进行进一步的过滤，以返回最终所需结果。
	Filter *string `json:"filter,omitempty" xml:"filter,omitempty"`
	// 为便于通过查询语句传递信息给具体的特征函数，用户可以在kvpairs子句中对排序表达式中的可变部分进行参数定义。
	Kvpairs map[string]*string `json:"kvpairs,omitempty" xml:"kvpairs,omitempty"`
	// 用户可以通过查询语句控制结果的排序方式，包括指定排序的字段和升降序。field为要排序的字段，+为按字段值升序排序，-为降序排序
	Sort []*HaQuerySortClause `json:"sort,omitempty" xml:"sort,omitempty" type:"Repeated"`
	// 一个关键词查询后可能会找到数以万计的文档，用户不太可能浏览所有的文档来获取自己需要的信息，有些情况下用户感兴趣的可能是一些统计的信息。
	Aggregate []*HaQueryAggregateClause `json:"aggregate,omitempty" xml:"aggregate,omitempty" type:"Repeated"`
	// 打散子句可以在一定程度上保证展示结果的多样性，以提升用户体验。如一次查询可以查出很多的文档，但是如果某个用户的多个文档分值都比较高，则都排在了前面，导致一页中所展示的结果几乎都属于同一用户，这样既不利于结果展示也不利于用户体验。对此，打散子句可以对每个用户的文档进行抽取，使得每个用户都有展示文档的机会。
	Distinct []*HaQueryDistinctClause `json:"distinct,omitempty" xml:"distinct,omitempty" type:"Repeated"`
	// 扩展 配置参数
	CustomQuery map[string]*string `json:"customConfig,omitempty" xml:"customConfig,omitempty"`
}

func (s HaQuery) String() string {
	return tea.Prettify(s)
}

func (s HaQuery) GoString() string {
	return s.String()
}

func (s *HaQuery) SetQuery(v string) *HaQuery {
	s.Query = &v
	return s
}

func (s *HaQuery) SetCluster(v string) *HaQuery {
	s.Cluster = &v
	return s
}

func (s *HaQuery) SetConfig(v *HaQueryconfigClause) *HaQuery {
	s.Config = v
	return s
}

func (s *HaQuery) SetFilter(v string) *HaQuery {
	s.Filter = &v
	return s
}

func (s *HaQuery) SetKvpairs(v map[string]*string) *HaQuery {
	s.Kvpairs = v
	return s
}

func (s *HaQuery) SetSort(v []*HaQuerySortClause) *HaQuery {
	s.Sort = v
	return s
}

func (s *HaQuery) SetAggregate(v []*HaQueryAggregateClause) *HaQuery {
	s.Aggregate = v
	return s
}

func (s *HaQuery) SetDistinct(v []*HaQueryDistinctClause) *HaQuery {
	s.Distinct = v
	return s
}

func (s *HaQuery) SetCustomQuery(v map[string]*string) *HaQuery {
	s.CustomQuery = v
	return s
}

type HaQueryconfigClause struct {
	// 从结果集中第 start_offset 开始返回 document
	Start *string `json:"start,omitempty" xml:"start,omitempty" require:"true"`
	// 返回文档的最大数量
	Hit *string `json:"hit,omitempty" xml:"hit,omitempty" require:"true"`
	// 指定用户返回数据格式. 支持 xml 和 json 类型数据返回
	Format *string `json:"format,omitempty" xml:"format,omitempty" require:"true"`
	// 扩展 配置参数
	CustomConfig map[string]*string `json:"customConfig,omitempty" xml:"customConfig,omitempty"`
}

func (s HaQueryconfigClause) String() string {
	return tea.Prettify(s)
}

func (s HaQueryconfigClause) GoString() string {
	return s.String()
}

func (s *HaQueryconfigClause) SetStart(v string) *HaQueryconfigClause {
	s.Start = &v
	return s
}

func (s *HaQueryconfigClause) SetHit(v string) *HaQueryconfigClause {
	s.Hit = &v
	return s
}

func (s *HaQueryconfigClause) SetFormat(v string) *HaQueryconfigClause {
	s.Format = &v
	return s
}

func (s *HaQueryconfigClause) SetCustomConfig(v map[string]*string) *HaQueryconfigClause {
	s.CustomConfig = v
	return s
}

type HaQuerySortClause struct {
	// field为要进行统计的字段名，必须配置属性字段
	SortKey *string `json:"sortKey,omitempty" xml:"sortKey,omitempty" require:"true"`
	// +为按字段值升序排序，-为降序排序
	SortOrder *string `json:"sortOrder,omitempty" xml:"sortOrder,omitempty" require:"true"`
}

func (s HaQuerySortClause) String() string {
	return tea.Prettify(s)
}

func (s HaQuerySortClause) GoString() string {
	return s.String()
}

func (s *HaQuerySortClause) SetSortKey(v string) *HaQuerySortClause {
	s.SortKey = &v
	return s
}

func (s *HaQuerySortClause) SetSortOrder(v string) *HaQuerySortClause {
	s.SortOrder = &v
	return s
}

type HaQueryAggregateClause struct {
	// field为要进行统计的字段名，必须配置属性字段
	GroupKey *string `json:"group_key,omitempty" xml:"group_key,omitempty" require:"true"`
	// func可以为count()、sum(id)、max(id)、min(id)四种系统函数，含义分别为：文档个数、对id字段求和、取id字段最大值、取id字段最小值；支持同时进行多个函数的统计，中间用英文井号（#）分隔；sum、max、min的内容支持基本的算术运算
	AggFun *string `json:"agg_fun,omitempty" xml:"agg_fun,omitempty" require:"true"`
	// 表示分段统计，可用于分布统计，只支持单个range参数。
	Range *string `json:"range,omitempty" xml:"range,omitempty"`
	// 最大返回组数
	MaxGroup *string `json:"max_group,omitempty" xml:"max_group,omitempty"`
	// 表示仅统计满足特定条件的文档
	AggFilter *string `json:"agg_filter,omitempty" xml:"agg_filter,omitempty"`
	// ，抽样统计的阈值。表示该值之前的文档会依次统计，该值之后的文档会进行抽样统计；
	AggSamplerThresHold *string `json:"agg_sampler_threshold,omitempty" xml:"agg_sampler_threshold,omitempty"`
	// 抽样统计的步长，表示从agg_sampler_threshold后的文档将间隔agg_sampler_step个文档统计一次。对于sum和count类型的统计会把阈值后的抽样统计结果最后乘以步长进行估算，估算的结果再加上阈值前的统计结果就是最后的统计结果。
	AggSamplerStep *string `json:"agg_sampler_step,omitempty" xml:"agg_sampler_step,omitempty"`
}

func (s HaQueryAggregateClause) String() string {
	return tea.Prettify(s)
}

func (s HaQueryAggregateClause) GoString() string {
	return s.String()
}

func (s *HaQueryAggregateClause) SetGroupKey(v string) *HaQueryAggregateClause {
	s.GroupKey = &v
	return s
}

func (s *HaQueryAggregateClause) SetAggFun(v string) *HaQueryAggregateClause {
	s.AggFun = &v
	return s
}

func (s *HaQueryAggregateClause) SetRange(v string) *HaQueryAggregateClause {
	s.Range = &v
	return s
}

func (s *HaQueryAggregateClause) SetMaxGroup(v string) *HaQueryAggregateClause {
	s.MaxGroup = &v
	return s
}

func (s *HaQueryAggregateClause) SetAggFilter(v string) *HaQueryAggregateClause {
	s.AggFilter = &v
	return s
}

func (s *HaQueryAggregateClause) SetAggSamplerThresHold(v string) *HaQueryAggregateClause {
	s.AggSamplerThresHold = &v
	return s
}

func (s *HaQueryAggregateClause) SetAggSamplerStep(v string) *HaQueryAggregateClause {
	s.AggSamplerStep = &v
	return s
}

type HaQueryDistinctClause struct {
	// 要打散的字段
	DistKey *string `json:"dist_key,omitempty" xml:"dist_key,omitempty" require:"true"`
	// 一轮抽取的文档数
	DistCount *string `json:"dist_count,omitempty" xml:"dist_count,omitempty"`
	// 抽取的轮数
	DistTimes *string `json:"dist_times,omitempty" xml:"dist_times,omitempty"`
	// 是否保留抽取之后剩余的文档。如果为false，为不保留，则搜索结果的total（总匹配结果数）会不准确。
	Reserved *string `json:"reserved,omitempty" xml:"reserved,omitempty"`
	// 过滤条件，被过滤的doc不参与distinct，只在后面的排序中，这些被过滤的doc将和被distinct出来的第一组doc一起参与排序。默认是全部参与distinct。
	DistFilter *string `json:"dist_filter,omitempty" xml:"dist_filter,omitempty"`
	// 当reserved为false时，设置update_total_hit为true，则最终total_hit会减去被distinct丢弃的数目（不一定准确），为false则不减。
	UpdateTotalHit *string `json:"update_total_hit,omitempty" xml:"update_total_hit,omitempty"`
	// 指定档位划分阈值，所有的文档将根据档位划分阈值划分成若干档，每个档位中各自根据distinct参数做distinct，可以不指定该参数，默认是所有文档都在同一档。档位的划分按照文档排序时第一维的排序依据的分数进行划分，两个档位阈值之间用 “|” 分开，档位的个数没有限制。例如：1、grade:3.0 ：表示根据第一维排序依据的分数分成两档，(< 3.0)的是第一档，(>= 3.0) 的是第二档；2、grade:3.0|5.0 ：表示分成三档，(< 3.0)是第一档，(>= 3.0，< 5.0)是第二档，(>= 5.0)是第三档。档位的先后顺序和第一维排序依据的顺序一致，即如果第一维排序依据是降序，则档位也是降序，反之亦然。
	Grade *string `json:"grade,omitempty" xml:"grade,omitempty"`
}

func (s HaQueryDistinctClause) String() string {
	return tea.Prettify(s)
}

func (s HaQueryDistinctClause) GoString() string {
	return s.String()
}

func (s *HaQueryDistinctClause) SetDistKey(v string) *HaQueryDistinctClause {
	s.DistKey = &v
	return s
}

func (s *HaQueryDistinctClause) SetDistCount(v string) *HaQueryDistinctClause {
	s.DistCount = &v
	return s
}

func (s *HaQueryDistinctClause) SetDistTimes(v string) *HaQueryDistinctClause {
	s.DistTimes = &v
	return s
}

func (s *HaQueryDistinctClause) SetReserved(v string) *HaQueryDistinctClause {
	s.Reserved = &v
	return s
}

func (s *HaQueryDistinctClause) SetDistFilter(v string) *HaQueryDistinctClause {
	s.DistFilter = &v
	return s
}

func (s *HaQueryDistinctClause) SetUpdateTotalHit(v string) *HaQueryDistinctClause {
	s.UpdateTotalHit = &v
	return s
}

func (s *HaQueryDistinctClause) SetGrade(v string) *HaQueryDistinctClause {
	s.Grade = &v
	return s
}

type SQLQuery struct {
	// 搜索主体，不能为空.
	Query *string `json:"query,omitempty" xml:"query,omitempty" require:"true"`
	// cluster部分用于指定要查询的集群的名字。它不仅可以同时指定多个集群，还可以指定到集群中的哪些partition获取结果。
	Kvpairs map[string]*string `json:"kvpairs,omitempty" xml:"kvpairs,omitempty"`
}

func (s SQLQuery) String() string {
	return tea.Prettify(s)
}

func (s SQLQuery) GoString() string {
	return s.String()
}

func (s *SQLQuery) SetQuery(v string) *SQLQuery {
	s.Query = &v
	return s
}

func (s *SQLQuery) SetKvpairs(v map[string]*string) *SQLQuery {
	s.Kvpairs = v
	return s
}

type SearchQuery struct {
	// 适配于 Ha3 类型 query. 参数支持子句关键字请参照文档
	Query *string `json:"query,omitempty" xml:"query,omitempty"`
	// 适配于 SQL 类型 query. 参数支持子句关键字请参照文档.
	Sql *string `json:"sql,omitempty" xml:"sql,omitempty"`
}

func (s SearchQuery) String() string {
	return tea.Prettify(s)
}

func (s SearchQuery) GoString() string {
	return s.String()
}

func (s *SearchQuery) SetQuery(v string) *SearchQuery {
	s.Query = &v
	return s
}

func (s *SearchQuery) SetSql(v string) *SearchQuery {
	s.Sql = &v
	return s
}

type SearchRequestModel struct {
	// headers
	Headers map[string]*string `json:"headers,omitempty" xml:"headers,omitempty"`
	// query
	Query *SearchQuery `json:"query,omitempty" xml:"query,omitempty" require:"true"`
}

func (s SearchRequestModel) String() string {
	return tea.Prettify(s)
}

func (s SearchRequestModel) GoString() string {
	return s.String()
}

func (s *SearchRequestModel) SetHeaders(v map[string]*string) *SearchRequestModel {
	s.Headers = v
	return s
}

func (s *SearchRequestModel) SetQuery(v *SearchQuery) *SearchRequestModel {
	s.Query = v
	return s
}

type SearchResponseModel struct {
	// headers
	Headers map[string]*string `json:"headers,omitempty" xml:"headers,omitempty"`
	// body
	Body *string `json:"body,omitempty" xml:"body,omitempty" require:"true"`
}

func (s SearchResponseModel) String() string {
	return tea.Prettify(s)
}

func (s SearchResponseModel) GoString() string {
	return s.String()
}

func (s *SearchResponseModel) SetHeaders(v map[string]*string) *SearchResponseModel {
	s.Headers = v
	return s
}

func (s *SearchResponseModel) SetBody(v string) *SearchResponseModel {
	s.Body = &v
	return s
}

type PushDocumentsRequestModel struct {
	// headers
	Headers map[string]*string `json:"headers,omitempty" xml:"headers,omitempty"`
	// body
	Body []map[string]interface{} `json:"body,omitempty" xml:"body,omitempty" require:"true" type:"Repeated"`
}

func (s PushDocumentsRequestModel) String() string {
	return tea.Prettify(s)
}

func (s PushDocumentsRequestModel) GoString() string {
	return s.String()
}

func (s *PushDocumentsRequestModel) SetHeaders(v map[string]*string) *PushDocumentsRequestModel {
	s.Headers = v
	return s
}

func (s *PushDocumentsRequestModel) SetBody(v []map[string]interface{}) *PushDocumentsRequestModel {
	s.Body = v
	return s
}

type PushDocumentsResponseModel struct {
	// headers
	Headers map[string]*string `json:"headers,omitempty" xml:"headers,omitempty"`
	// body
	Body *string `json:"body,omitempty" xml:"body,omitempty" require:"true"`
}

func (s PushDocumentsResponseModel) String() string {
	return tea.Prettify(s)
}

func (s PushDocumentsResponseModel) GoString() string {
	return s.String()
}

func (s *PushDocumentsResponseModel) SetHeaders(v map[string]*string) *PushDocumentsResponseModel {
	s.Headers = v
	return s
}

func (s *PushDocumentsResponseModel) SetBody(v string) *PushDocumentsResponseModel {
	s.Body = &v
	return s
}

type Client struct {
	Endpoint     *string
	InstanceId   *string
	Protocol     *string
	UserAgent    *string
	Credential   *string
	Domainsuffix *string
}

func NewClient(config *Config) (*Client, error) {
	client := new(Client)
	err := client.Init(config)
	return client, err
}

func (client *Client) Init(config *Config) (_err error) {
	if tea.BoolValue(util.IsUnset(tea.ToMap(config))) {
		_err = tea.NewSDKError(map[string]interface{}{
			"name":    "ParameterMissing",
			"message": "'config' can not be unset",
		})
		return _err
	}

	client.Credential = client.GetRealmSignStr(config.AccessUserName, config.AccessPassWord)
	client.Endpoint = config.Endpoint
	client.InstanceId = config.InstanceId
	client.Protocol = config.Protocol
	client.UserAgent = config.UserAgent
	client.Domainsuffix = tea.String("ha.aliyuncs.com")
	return nil
}

func (client *Client) _request(method *string, pathname *string, query map[string]interface{}, headers map[string]*string, body interface{}, runtime *util.RuntimeOptions) (_result map[string]interface{}, _err error) {
	_err = tea.Validate(runtime)
	if _err != nil {
		return _result, _err
	}
	_runtime := map[string]interface{}{
		"timeouted":      "retry",
		"readTimeout":    tea.IntValue(runtime.ReadTimeout),
		"connectTimeout": tea.IntValue(runtime.ConnectTimeout),
		"httpProxy":      tea.StringValue(runtime.HttpProxy),
		"httpsProxy":     tea.StringValue(runtime.HttpsProxy),
		"noProxy":        tea.StringValue(runtime.NoProxy),
		"maxIdleConns":   tea.IntValue(runtime.MaxIdleConns),
		"retry": map[string]interface{}{
			"retryable":   tea.BoolValue(runtime.Autoretry),
			"maxAttempts": tea.IntValue(util.DefaultNumber(runtime.MaxAttempts, tea.Int(5))),
		},
		"backoff": map[string]interface{}{
			"policy": tea.StringValue(util.DefaultString(runtime.BackoffPolicy, tea.String("no"))),
			"period": tea.IntValue(util.DefaultNumber(runtime.BackoffPeriod, tea.Int(1))),
		},
		"ignoreSSL": tea.BoolValue(runtime.IgnoreSSL),
	}

	_resp := make(map[string]interface{})
	for _retryTimes := 0; tea.BoolValue(tea.AllowRetry(_runtime["retry"], tea.Int(_retryTimes))); _retryTimes++ {
		if _retryTimes > 0 {
			_backoffTime := tea.GetBackoffTime(_runtime["backoff"], tea.Int(_retryTimes))
			if tea.IntValue(_backoffTime) > 0 {
				tea.Sleep(_backoffTime)
			}
		}

		_resp, _err = func() (map[string]interface{}, error) {
			request_ := tea.NewRequest()
			request_.Protocol = util.DefaultString(client.Protocol, tea.String("HTTP"))
			request_.Method = method
			request_.Pathname = pathname
			request_.Headers = tea.Merge(map[string]*string{
				"user-agent":    client.GetUserAgent(),
				"host":          util.DefaultString(client.Endpoint, tea.String(tea.StringValue(client.InstanceId)+"."+tea.StringValue(client.Domainsuffix))),
				"authorization": tea.String("Basic " + tea.StringValue(client.Credential)),
				"content-type":  tea.String("application/json; charset=utf-8"),
			}, headers)
			if !tea.BoolValue(util.IsUnset(query)) {
				request_.Query = util.StringifyMapValue(query)
				request_.Headers["X-Opensearch-Request-ID"] = util.GetNonce()
			}

			if !tea.BoolValue(util.IsUnset(body)) {
				request_.Headers["X-Opensearch-Swift-Request-ID"] = util.GetNonce()
				request_.Body = tea.ToReader(util.ToJSONString(body))
			}

			response_, _err := tea.DoRequest(request_, _runtime)
			if _err != nil {
				return _result, _err
			}
			objStr, _err := util.ReadAsString(response_.Body)
			if _err != nil {
				return _result, _err
			}

			if tea.BoolValue(util.Is4xx(response_.StatusCode)) || tea.BoolValue(util.Is5xx(response_.StatusCode)) {
				var rawMsg interface{}
				_, tryErr := func() (_r map[string]interface{}, _e error) {
					defer func() {
						if r := tea.Recover(recover()); r != nil {
							_e = r
						}
					}()
					rawMsg = util.ParseJSON(objStr)
					return nil, nil
				}()
				if tryErr != nil {
					var err = &tea.SDKError{}
					if _t, ok := tryErr.(*tea.SDKError); ok {
						err = _t
					} else {
						err.Message = tea.String(tryErr.Error())
					}
					rawMsg = objStr
				}
				rawMap := map[string]interface{}{
					"errors": rawMsg,
				}
				_err = tea.NewSDKError(map[string]interface{}{
					"message": tea.StringValue(response_.StatusMessage),
					"data":    rawMap,
					"code":    tea.IntValue(response_.StatusCode),
				})
				return _result, _err
			}

			if tea.BoolValue(util.Empty(objStr)) {
				rawbodyMap := map[string]interface{}{
					"status": tea.StringValue(response_.StatusMessage),
					"code":   tea.IntValue(response_.StatusCode),
				}
				_result = make(map[string]interface{})
				_err = tea.Convert(map[string]interface{}{
					"body":    tea.StringValue(util.ToJSONString(rawbodyMap)),
					"headers": response_.Headers,
				}, &_result)
				return _result, _err
			}

			_result = make(map[string]interface{})
			_err = tea.Convert(map[string]interface{}{
				"body":    tea.StringValue(objStr),
				"headers": response_.Headers,
			}, &_result)
			return _result, _err
		}()
		if !tea.BoolValue(tea.Retryable(_err)) {
			break
		}
	}

	return _resp, _err
}

/**
 * 设置Client UA 配置.
 */
func (client *Client) SetUserAgent(userAgent *string) {
	client.UserAgent = userAgent
}

/**
 * 添加Client UA 配置.
 */
func (client *Client) AppendUserAgent(userAgent *string) {
	client.UserAgent = tea.String(tea.StringValue(client.UserAgent) + " " + tea.StringValue(userAgent))
}

/**
 * 获取Client 配置 UA 配置.
 */
func (client *Client) GetUserAgent() (_result *string) {
	userAgent := util.GetUserAgent(client.UserAgent)
	_result = userAgent
	return _result
}

/**
 * 计算用户请求识别特征, 遵循 Basic Auth 生成规范.
 */
func (client *Client) GetRealmSignStr(accessUserName *string, accessPassWord *string) (_result *string) {
	accessUserNameStr := string_.Trim(accessUserName)
	accessPassWordStr := string_.Trim(accessPassWord)
	realmStr := tea.String(tea.StringValue(accessUserNameStr) + ":" + tea.StringValue(accessPassWordStr))
	_body := encodeutil.Base64EncodeToString(string_.ToBytes(realmStr, tea.String("UTF-8")))
	_result = _body
	return _result
}

func (client *Client) BuildHaSearchQuery(haquery *HaQuery) (_result *string, _err error) {
	if tea.BoolValue(util.IsUnset(haquery.Query)) {
		_err = tea.NewSDKError(map[string]interface{}{
			"name":    "ParameterMissing",
			"message": "'HaQuery.query' can not be unset",
		})
		return _result, _err
	}

	tempString := tea.String("query=" + tea.StringValue(haquery.Query))
	configStr, _err := client.BuildHaQueryconfigClauseStr(haquery.Config)
	if _err != nil {
		return _result, _err
	}
	tempString = tea.String(tea.StringValue(tempString) + "&&cluster=" + tea.StringValue(util.DefaultString(haquery.Cluster, tea.String("general"))))
	tempString = tea.String(tea.StringValue(tempString) + "&&config=" + tea.StringValue(configStr))
	if !tea.BoolValue(util.IsUnset(haquery.Filter)) {
		filterStr := haquery.Filter
		if !tea.BoolValue(util.Empty(filterStr)) {
			fieldValueTrimed := string_.Trim(filterStr)
			tempString = tea.String(tea.StringValue(tempString) + "&&filter=" + tea.StringValue(fieldValueTrimed))
		}

	}

	if !tea.BoolValue(util.IsUnset(haquery.CustomQuery)) {
		for _, keyField := range map_.KeySet(haquery.CustomQuery) {
			fieldValue := haquery.CustomQuery[tea.StringValue(keyField)]
			if !tea.BoolValue(util.Empty(fieldValue)) {
				fieldValueTrimed := string_.Trim(fieldValue)
				keyFieldTrimed := string_.Trim(keyField)
				tempString = tea.String(tea.StringValue(tempString) + "&&" + tea.StringValue(keyFieldTrimed) + "=" + tea.StringValue(fieldValueTrimed))
			}

		}
	}

	if !tea.BoolValue(util.IsUnset(haquery.Sort)) {
		sortStr := client.BuildHaQuerySortClauseStr(haquery.Sort)
		if !tea.BoolValue(util.Empty(sortStr)) {
			tempString = tea.String(tea.StringValue(tempString) + "&&sort=" + tea.StringValue(sortStr))
		}

	}

	if !tea.BoolValue(util.IsUnset(haquery.Aggregate)) {
		aggregateClauseStr, _err := client.BuildHaQueryAggregateClauseStr(haquery.Aggregate)
		if _err != nil {
			return _result, _err
		}
		if !tea.BoolValue(util.Empty(aggregateClauseStr)) {
			tempString = tea.String(tea.StringValue(tempString) + "&&aggregate=" + tea.StringValue(aggregateClauseStr))
		}

	}

	if !tea.BoolValue(util.IsUnset(haquery.Distinct)) {
		distinctClauseStr, _err := client.BuildHaQueryDistinctClauseStr(haquery.Distinct)
		if _err != nil {
			return _result, _err
		}
		if !tea.BoolValue(util.Empty(distinctClauseStr)) {
			tempString = tea.String(tea.StringValue(tempString) + "&&distinct=" + tea.StringValue(distinctClauseStr))
		}
	}

	kvpairs := client.BuildSearcKvPairClauseStr(haquery.Kvpairs)
	if !tea.BoolValue(util.Empty(kvpairs)) {
		tempString = tea.String(tea.StringValue(tempString) + "&&kvpairs=" + tea.StringValue(kvpairs))
	}

	_result = tempString
	return _result, _err
}

func (client *Client) BuildHaQueryAggregateClauseStr(Clause []*HaQueryAggregateClause) (_result *string, _err error) {
	_err = nil
	tempClauseString := tea.String("")
	for _, AggregateClause := range Clause {
		tempAggregateClauseString := tea.String("")
		if tea.BoolValue(util.IsUnset(AggregateClause.GroupKey)) || tea.BoolValue(util.IsUnset(AggregateClause.AggFun)) {
			_err := tea.NewSDKError(map[string]interface{}{
				"name":    "ParameterMissing",
				"message": "'HaQueryAggregateClause.groupKey/aggFun' can not be unset",
			})
			return _result, _err
		}

		if !tea.BoolValue(util.Empty(AggregateClause.GroupKey)) && !tea.BoolValue(util.Empty(AggregateClause.AggFun)) {
			groupKeyTrimed := string_.Trim(AggregateClause.GroupKey)
			aggFunTrimed := string_.Trim(AggregateClause.AggFun)
			tempAggregateClauseString = tea.String("group_key:" + tea.StringValue(groupKeyTrimed) + ",agg_fun:" + tea.StringValue(aggFunTrimed))
		}

		if !tea.BoolValue(util.Empty(AggregateClause.Range)) {
			rangeTrimed := string_.Trim(AggregateClause.Range)
			tempAggregateClauseString = tea.String(tea.StringValue(tempAggregateClauseString) + ",range:" + tea.StringValue(rangeTrimed))
		}

		if !tea.BoolValue(util.Empty(AggregateClause.MaxGroup)) {
			maxGroupTrimed := string_.Trim(AggregateClause.MaxGroup)
			tempAggregateClauseString = tea.String(tea.StringValue(tempAggregateClauseString) + ",max_group:" + tea.StringValue(maxGroupTrimed))
		}

		if !tea.BoolValue(util.Empty(AggregateClause.AggFilter)) {
			aggFilterTrimed := string_.Trim(AggregateClause.AggFilter)
			tempAggregateClauseString = tea.String(tea.StringValue(tempAggregateClauseString) + ",agg_filter:" + tea.StringValue(aggFilterTrimed))
		}

		if !tea.BoolValue(util.Empty(AggregateClause.AggSamplerThresHold)) {
			aggSamplerThresHoldTrimed := string_.Trim(AggregateClause.AggSamplerThresHold)
			tempAggregateClauseString = tea.String(tea.StringValue(tempAggregateClauseString) + ",agg_sampler_threshold:" + tea.StringValue(aggSamplerThresHoldTrimed))
		}

		if !tea.BoolValue(util.Empty(AggregateClause.AggSamplerStep)) {
			aggSamplerStepTrimed := string_.Trim(AggregateClause.AggSamplerStep)
			tempAggregateClauseString = tea.String(tea.StringValue(tempAggregateClauseString) + ",agg_sampler_step:" + tea.StringValue(aggSamplerStepTrimed))
		}

		if !tea.BoolValue(util.Empty(tempClauseString)) {
			tempClauseString = tea.String(tea.StringValue(tempClauseString) + ";" + tea.StringValue(tempAggregateClauseString))
		} else {
			tempClauseString = tea.String(tea.StringValue(tempAggregateClauseString))
		}

	}
	_result = tempClauseString
	return _result, _err
}

func (client *Client) BuildHaQueryDistinctClauseStr(Clause []*HaQueryDistinctClause) (_result *string, _err error) {
	tempClauseString := tea.String("")
	_err = nil
	for _, DistinctClause := range Clause {
		tempDistinctClauseString := tea.String("")
		if tea.BoolValue(util.IsUnset(DistinctClause.DistKey)) {
			_err = tea.NewSDKError(map[string]interface{}{
				"name":    "ParameterMissing",
				"message": "'HaQueryDistinctClause.distKey' can not be unset",
			})
			return _result, _err
		}

		if !tea.BoolValue(util.Empty(DistinctClause.DistKey)) {
			distKeyTrimed := string_.Trim(DistinctClause.DistKey)
			tempDistinctClauseString = tea.String("dist_key:" + tea.StringValue(distKeyTrimed))
		}

		if !tea.BoolValue(util.Empty(DistinctClause.DistCount)) {
			distCountTrimed := string_.Trim(DistinctClause.DistCount)
			tempDistinctClauseString = tea.String(tea.StringValue(tempDistinctClauseString) + ",dist_count:" + tea.StringValue(distCountTrimed))
		}

		if !tea.BoolValue(util.Empty(DistinctClause.DistTimes)) {
			distTimesTrimed := string_.Trim(DistinctClause.DistTimes)
			tempDistinctClauseString = tea.String(tea.StringValue(tempDistinctClauseString) + ",dist_times:" + tea.StringValue(distTimesTrimed))
		}

		if !tea.BoolValue(util.Empty(DistinctClause.Reserved)) {
			reservedTrimed := string_.Trim(DistinctClause.Reserved)
			tempDistinctClauseString = tea.String(tea.StringValue(tempDistinctClauseString) + ",reserved:" + tea.StringValue(reservedTrimed))
		}

		if !tea.BoolValue(util.Empty(DistinctClause.DistFilter)) {
			distFilterTrimed := string_.Trim(DistinctClause.DistFilter)
			tempDistinctClauseString = tea.String(tea.StringValue(tempDistinctClauseString) + ",dist_filter:" + tea.StringValue(distFilterTrimed))
		}

		if !tea.BoolValue(util.Empty(DistinctClause.UpdateTotalHit)) {
			updateTotalHitTrimed := string_.Trim(DistinctClause.UpdateTotalHit)
			tempDistinctClauseString = tea.String(tea.StringValue(tempDistinctClauseString) + ",update_total_hit:" + tea.StringValue(updateTotalHitTrimed))
		}

		if !tea.BoolValue(util.Empty(DistinctClause.Grade)) {
			gradeTrimed := string_.Trim(DistinctClause.Grade)
			tempDistinctClauseString = tea.String(tea.StringValue(tempDistinctClauseString) + ",grade:" + tea.StringValue(gradeTrimed))
		}

		if !tea.BoolValue(util.Empty(tempClauseString)) {
			tempClauseString = tea.String(tea.StringValue(tempClauseString) + ";" + tea.StringValue(tempDistinctClauseString))
		} else {
			tempClauseString = tea.String(tea.StringValue(tempDistinctClauseString))
		}

	}
	_result = tempClauseString
	return _result, _err
}

func (client *Client) BuildHaQuerySortClauseStr(Clause []*HaQuerySortClause) (_result *string) {
	tempClauseString := tea.String("")
	for _, SortClause := range Clause {
		fieldValueTrimed := string_.Trim(SortClause.SortOrder)
		keyFieldTrimed := string_.Trim(SortClause.SortKey)
		if tea.BoolValue(util.EqualString(fieldValueTrimed, tea.String("+"))) || tea.BoolValue(util.EqualString(fieldValueTrimed, tea.String("-"))) {
			if !tea.BoolValue(util.Empty(fieldValueTrimed)) && !tea.BoolValue(util.Empty(keyFieldTrimed)) {
				if tea.BoolValue(util.Empty(tempClauseString)) {
					tempClauseString = tea.String(tea.StringValue(fieldValueTrimed) + tea.StringValue(keyFieldTrimed))
				} else {
					tempClauseString = tea.String(tea.StringValue(tempClauseString) + ";" + tea.StringValue(fieldValueTrimed) + tea.StringValue(keyFieldTrimed))
				}

			}

		}

	}
	_result = tempClauseString
	return _result
}

func (client *Client) BuildHaQueryconfigClauseStr(Clause *HaQueryconfigClause) (_result *string, _err error) {
	_err = nil
	tempClauseString := tea.String("")
	if tea.BoolValue(util.IsUnset(tea.ToMap(Clause))) {
		_err = tea.NewSDKError(map[string]interface{}{
			"name":    "ParameterMissing",
			"message": "'HaQueryconfigClause' can not be unset",
		})
		return _result, _err
	}

	if tea.BoolValue(util.IsUnset(Clause.Start)) {
		Clause.Start = nil
	}

	if tea.BoolValue(util.IsUnset(Clause.Hit)) {
		Clause.Hit = nil
	}

	if tea.BoolValue(util.IsUnset(Clause.Format)) {
		Clause.Format = nil
	}
	tempClauseString = tea.String("start:" + tea.StringValue(util.DefaultString(Clause.Start, tea.String("0"))))
	tempClauseString = tea.String(tea.StringValue(tempClauseString) + ",hit:" + tea.StringValue(util.DefaultString(Clause.Hit, tea.String("10"))))
	tempClauseString = tea.String(tea.StringValue(tempClauseString) + ",format:" + tea.StringValue(string_.ToLower(util.DefaultString(Clause.Format, tea.String("json")))))
	if !tea.BoolValue(util.IsUnset(Clause.CustomConfig)) {
		for _, keyField := range map_.KeySet(Clause.CustomConfig) {
			fieldValue := Clause.CustomConfig[tea.StringValue(keyField)]
			if !tea.BoolValue(util.Empty(fieldValue)) {
				fieldValueTrimed := string_.Trim(fieldValue)
				keyFieldTrimed := string_.Trim(keyField)
				if !tea.BoolValue(util.Empty(tempClauseString)) {
					tempClauseString = tea.String(tea.StringValue(tempClauseString) + "," + tea.StringValue(keyFieldTrimed) + ":" + tea.StringValue(fieldValueTrimed))
				} else {
					tempClauseString = tea.String(tea.StringValue(keyFieldTrimed) + ":" + tea.StringValue(fieldValueTrimed))
				}

			}

		}
	}

	_result = tempClauseString
	return _result, _err
}

func (client *Client) BuildSQLSearchQuery(sqlquery *SQLQuery) (_result *string, _err error) {
	_err = nil
	if tea.BoolValue(util.IsUnset(sqlquery.Query)) {
		_err = tea.NewSDKError(map[string]interface{}{
			"name":    "ParameterMissing",
			"message": "'SQLQuery.query' can not be unset",
		})
		return _result, _err
	}

	tempString := tea.String("query=" + tea.StringValue(sqlquery.Query))
	kvpairs := client.BuildSearcKvPairClauseStr(sqlquery.Kvpairs)
	if !tea.BoolValue(util.Empty(kvpairs)) {
		tempString = tea.String(tea.StringValue(tempString) + "&&kvpair=" + tea.StringValue(kvpairs))
	}

	_result = tempString
	return _result, _err
}

func (client *Client) BuildSearcKvPairClauseStr(kvPair map[string]*string) (_result *string) {
	tempkvpairsString := tea.String("__ops_request_id:" + tea.StringValue(util.GetNonce()))
	if !tea.BoolValue(util.IsUnset(kvPair)) {
		for _, keyField := range map_.KeySet(kvPair) {
			fieldValue := kvPair[tea.StringValue(keyField)]
			if !tea.BoolValue(util.Empty(fieldValue)) {
				fieldValueTrimed := string_.Trim(fieldValue)
				keyFieldTrimed := string_.Trim(keyField)
				tempkvpairsString = tea.String(tea.StringValue(tempkvpairsString) + "," + tea.StringValue(keyFieldTrimed) + ":" + tea.StringValue(fieldValueTrimed))
			}

		}
	}

	_result = tempkvpairsString
	return _result
}

/**
 * 系统提供了丰富的搜索语法以满足用户各种场景下的搜索需求。
 */
func (client *Client) SearchEx(request *SearchRequestModel, runtime *util.RuntimeOptions) (_result *SearchResponseModel, _err error) {
	_result = &SearchResponseModel{}
	_body, _err := client._request(tea.String("GET"), tea.String("/query"), tea.ToMap(request.Query), request.Headers, nil, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

/**
 * 系统提供了丰富的搜索语法以满足用户各种场景下的搜索需求。
 */
func (client *Client) Search(request *SearchRequestModel) (_result *SearchResponseModel, _err error) {
	runtime := &util.RuntimeOptions{
		ConnectTimeout: tea.Int(5000),
		ReadTimeout:    tea.Int(10000),
		Autoretry:      tea.Bool(false),
		IgnoreSSL:      tea.Bool(false),
		MaxIdleConns:   tea.Int(50),
	}
	_result = &SearchResponseModel{}
	_body, _err := client.SearchWithOptions(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

/**
 * 系统提供了丰富的搜索语法以满足用户各种场景下的搜索需求,及传入运行时参数.
 */
func (client *Client) SearchWithOptions(request *SearchRequestModel, runtime *util.RuntimeOptions) (_result *SearchResponseModel, _err error) {
	_result = &SearchResponseModel{}
	_body, _err := client.SearchEx(request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

/**
 * 支持新增、更新、删除 等操作，以及对应批量操作
 */
func (client *Client) PushDocumentEx(dataSourceName *string, request *PushDocumentsRequestModel, runtime *util.RuntimeOptions) (_result *PushDocumentsResponseModel, _err error) {
	_result = &PushDocumentsResponseModel{}
	_body, _err := client._request(tea.String("POST"), tea.String("/update/"+tea.StringValue(dataSourceName)+"/actions/bulk"), nil, request.Headers, request.Body, runtime)
	if _err != nil {
		return _result, _err
	}
	_err = tea.Convert(_body, &_result)
	return _result, _err
}

/**
 * 支持新增、更新、删除 等操作，以及对应批量操作
 */
func (client *Client) PushDocuments(dataSourceName *string, keyField *string, request *PushDocumentsRequestModel) (_result *PushDocumentsResponseModel, _err error) {
	runtime := &util.RuntimeOptions{
		ConnectTimeout: tea.Int(5000),
		ReadTimeout:    tea.Int(10000),
		Autoretry:      tea.Bool(false),
		IgnoreSSL:      tea.Bool(false),
		MaxIdleConns:   tea.Int(50),
	}
	_result = &PushDocumentsResponseModel{}
	_body, _err := client.PushDocumentsWithOptions(dataSourceName, keyField, request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}

/**
 * 支持新增、更新、删除 等操作，以及对应批量操作,及传入运行时参数.
 */
func (client *Client) PushDocumentsWithOptions(dataSourceName *string, keyField *string, request *PushDocumentsRequestModel, runtime *util.RuntimeOptions) (_result *PushDocumentsResponseModel, _err error) {
	request.Headers = map[string]*string{
		"X-Opensearch-Swift-PK-Field": keyField,
	}
	_result = &PushDocumentsResponseModel{}
	_body, _err := client.PushDocumentEx(dataSourceName, request, runtime)
	if _err != nil {
		return _result, _err
	}
	_result = _body
	return _result, _err
}
