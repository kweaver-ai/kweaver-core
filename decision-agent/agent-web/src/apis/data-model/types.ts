export interface GetMetricModalGroupsParamsType {
  // 排序类型，默认是group_name
  sort?: 'group_name';

  // 排序结果方向，可选asc、desc。 默认asc
  direction?: 'asc' | 'desc';

  // 开始响应的项目的偏移量 范围需大于等于0，默认值0
  offset?: number;

  // 每页最多可返回的项目数； 分页可选1-1000，-1表示不分页； 默认值10
  limit?: number;
}

// 指标模型分组类型
export interface MetricModalGroupType {
  // 分组内指标模型数量
  metric_model_count: number;

  // 分组ID
  id: string;

  // 分组名称，对于默认分组，查询返回结果为空字符串，由前端根据语言环境进行展示
  name: string;

  // 指标模型分组说明
  comment: string;

  // 更新时间
  update_time: number;
}

export interface GetMetricModalGroupsResponseType {
  // 总条数
  total_count: number;

  // 条目列表
  entries: Array<MetricModalGroupType>;
}

export interface GetMetricModelsParamsType {
  // 根据指标模型名称模糊查询，默认为空. 与 name 不能同时存在
  name_pattern?: string;

  // 指标类型，默认为空
  metric_type?: 'atomic' | 'complex';

  // 指标查询语言，默认为空
  query_type?: 'promql' | 'dsl';

  // 排序类型，默认是update_time
  sort?: 'update_time' | 'model_name' | 'group_name';

  // 排序结果方向，可选asc、desc。 默认desc
  direction?: 'asc' | 'desc';

  // 开始响应的项目的偏移量 范围需大于等于0，默认值0
  offset?: number;

  // 每页最多可返回的项目数； 分页可选1-1000，-1表示不分页； 默认值10
  limit?: number;

  // 是否查询指标模型简单信息；true表示查询指标模型部分字段（id,name,tags,metric_type,query_type,update_time）；flase表示查询返回指标模型所有字段信息；默认值flase。（可接受 1, t, T, TRUE, true, True, 0, f, F, FALSE, false, False）
  simple_info?: boolean;

  // 根据指标模型标签精准查询，默认为空.
  tag?: string;

  // 分组ID，查询该分组内指标模型，可以不填，不填则查询所有指标模型；
  group_id?: number;
}

export interface MetricModelType {
  // 指标模型 id
  id: string;

  // 指标模型名称
  name: string;

  // 度量名称
  measure_name: string;

  // 分组ID
  group_id: string;

  // 分组名称，对于默认分组，查询返回结果为空字符串，由前端根据语言环境进行展示
  group_name: string;

  // 标签。 （可以为空）
  tags: Array<string>;

  // 指标模型备注（可以为空）
  comment: string;

  // 指标类型，分原子指标和复合指标。
  metric_type: 'atomic' | 'derived' | 'composite';

  // 指标查询语言。
  query_type?: 'promql' | 'dsl' | 'sql';

  // 计算公式。当不是code的计算公式时，此字段不返回。
  formula?: string;

  // 配置化的指标计算公式。当不是配置化时此字段不返回
  formula_config?: any;

  // 分析维度。即在指标查询时可指定的下钻维度集合
  analysis_dimessions?: Array<any>;

  // 时间字段。当查询语言是 promql 时，此字段值为 @timestamp。当查询语言是 dsl 时序分析时，此字段值为 dsl 语句中的关于时间聚合分桶的聚合名称。当查询语言是sql时，必须指定时间字段。字段列表从数据源中获取。
  date_field: string;

  // 时间格式。当查询语言是 promql 时，此字段值为 epoch_millis。当查询语言是 dsl 时，此字段值为 epoch_millis
  date_format?: 'epoch_millis';

  // 度量字段。当查询语言是 promql 时，此字段值为 value。当查询语言是 dsl 时，此字段值为 dsl 语句中的值聚合名称。
  measure_field: string;

  // 单位类型
  unit_type:
    | 'numUnit'
    | 'storeUnit'
    | 'percent'
    | 'transmissionRate'
    | 'timeUnit'
    | 'currencyUnit'
    | 'percentageUnit'
    | 'countUnit'
    | 'weightUnit'
    | 'ordinalRankUnit';

  // 度量单位
  unit: string;

  // 创建时间
  create_time: number;

  // 更新时间
  update_time: number;

  // 是否内置
  builtin: boolean;

  // 是否日历间隔。0: 非日历间隔； 1: 日历间隔。查询简单信息时无此字段
  is_calendar_interval: number;
}
export interface GetMetricModelsResponseType {
  // 总条数
  total_count: number;

  // 条目列表
  entries: Array<MetricModelType>;
}

export interface GetMetricInfoByIdsParamsType {
  // 指标模型id数组
  ids: string[];

  // 是否包含视图的详细信息，true 表示查询结果包含视图的详细信息；false 表示查询结果不包含视图的详细信息；false。
  include_view?: boolean;
}

export type GetMetricInfoByIdsPesponseType = Array<{
  // 指标模型 id
  id: string;

  // 指标模型名称
  name: string;

  // 很多其它的，省略
  [key: string]: any;
}>;

export interface GetDataDictsParamsType {
  // 根据名称模糊匹配，不能与name同时存在
  name_pattern?: string;

  // 根据名称精确查询，不能与name_pattern同时存在
  name?: string;

  // 排序类型，默认是update_time
  sort?: 'update_time' | 'name';

  // 排序结果方向，可选asc、desc。 默认desc
  direction?: 'asc' | 'desc';

  // 开始响应的项目的偏移量 范围需大于等于0，默认值0
  offset?: number;

  // 每页最多可返回的项目数 分页可选1-1000，-1表示不分页； 默认值-1
  limit?: number;

  // 数据字典类型
  type?: 'kv_dict' | 'dimension_dict';

  // 根据标签名称精确过滤
  tag?: string;
}

interface DataDictType {
  id: string;
  name: string;
}

export interface GetDataDictsResponseType {
  // 总条数
  total_count: number;

  // 条目列表
  entries: Array<DataDictType>;
}

export type GetDataDictInfoByIdsResponseType = Array<{
  // 数据字典 id
  id: string;

  // 字典名称
  name: string;

  // 其它属性，省略
  [key: string]: string;
}>;
