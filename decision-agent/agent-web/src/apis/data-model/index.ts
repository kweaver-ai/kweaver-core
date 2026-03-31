import { get } from '@/utils/http';
import type {
  GetMetricModalGroupsParamsType,
  GetMetricModalGroupsResponseType,
  MetricModalGroupType,
  GetMetricModelsParamsType,
  MetricModelType,
  GetMetricInfoByIdsParamsType,
  GetMetricInfoByIdsPesponseType,
  GetMetricModelsResponseType,
  GetDataDictsParamsType,
  GetDataDictsResponseType,
  GetDataDictInfoByIdsResponseType,
} from './types';

export { type MetricModalGroupType, type MetricModelType };

const dataModelUrl = '/api/mdl-data-model/v1';
const dataDictUrl = `${dataModelUrl}/data-dicts`;

// 查询分组列表
export const getMetricModalGroups = (
  params?: GetMetricModalGroupsParamsType
): Promise<GetMetricModalGroupsResponseType> => get(`${dataModelUrl}/metric-model-groups`, { params });

// 查询指标模型列表
export const getMetricModels = (params: GetMetricModelsParamsType): Promise<GetMetricModelsResponseType> =>
  get(`${dataModelUrl}/metric-models`, { params });

// 按ids批量取指标模型对象信息
export const getMetricInfoByIds = ({
  ids,
  include_view,
}: GetMetricInfoByIdsParamsType): Promise<GetMetricInfoByIdsPesponseType> =>
  get(`${dataModelUrl}/metric-models/${ids.join(',')}`, { params: { include_view } });

// 查询数据字典列表
export const getDataDicts = (params?: GetDataDictsParamsType): Promise<GetDataDictsResponseType> =>
  get(`${dataDictUrl}`, { params });

// 根据 ids 批量查询数据字典信息
export const getDataDictInfoByIds = (ids: string[]): Promise<GetDataDictInfoByIdsResponseType> =>
  get(`${dataDictUrl}/${ids.join(',')}`);
