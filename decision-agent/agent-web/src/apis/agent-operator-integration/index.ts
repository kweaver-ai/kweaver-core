import { get, businessDomainHeaderKey } from '@/utils/http';
import {
  ToolListParams,
  BoxToolListResponse,
  GlobalToolListResponse,
  ToolResponse,
  ToolBoxMarketListParams,
  ToolBoxMarketListResponse,
} from './type';

const toolBoxIntegrationBaseUrl = '/api/agent-operator-integration/v1';

// =================== 基于toolbox.yaml的新工具箱API =================== //

// 获取工具箱内工具列表
export const getBoxToolList = (
  boxId: string,
  params?: ToolListParams,
  businessDomainIds?: string[] // 支持查询多个业务域下的资源
): Promise<BoxToolListResponse> => {
  const searchParams = new URLSearchParams();

  if (params?.page) searchParams.append('page', params.page.toString());
  if (params?.page_size) searchParams.append('page_size', params.page_size.toString());
  if (params?.sort_by) searchParams.append('sort_by', params.sort_by);
  if (params?.sort_order) searchParams.append('sort_order', params.sort_order);
  if (params?.name) searchParams.append('name', params.name);
  if (params?.status) searchParams.append('status', params.status);
  if (params?.user_id) searchParams.append('user_id', params.user_id);
  if (params?.all !== undefined) searchParams.append('all', params.all.toString());

  const queryString = searchParams.toString();
  return get(`${toolBoxIntegrationBaseUrl}/tool-box/${boxId}/tools/list${queryString ? `?${queryString}` : ''}`, {
    headers: {
      [businessDomainHeaderKey]: businessDomainIds?.join(',') || '',
    },
  });
};
export const getToolBoxListFromMarks = (params?: any, businessDomainIds?: string[]): Promise<BoxToolListResponse> => {
  const searchParams = new URLSearchParams();

  if (params?.page) searchParams.append('page', params.page.toString());
  if (params?.page_size) searchParams.append('page_size', params.page_size.toString());
  if (params?.status) searchParams.append('status', params.status);
  if (params?.all !== undefined) searchParams.append('all', params.all.toString());

  const queryString = searchParams.toString();
  return get(`${toolBoxIntegrationBaseUrl}/tool-box/market${queryString ? `?${queryString}` : ''}`, {
    headers: {
      [businessDomainHeaderKey]: businessDomainIds?.join(',') || '',
    },
  });
};

// 全局工具列表查询
export const getGlobalMarketToolList = (
  params?: any,
  businessDomainIds?: string[]
): Promise<GlobalToolListResponse> => {
  const searchParams = new URLSearchParams();

  if (params?.page) searchParams.append('page', params.page.toString());
  if (params?.page_size) searchParams.append('page_size', params.page_size.toString());
  if (params?.sort_by) searchParams.append('sort_by', params.sort_by);
  if (params?.sort_order) searchParams.append('sort_order', params.sort_order);
  if (params?.tool_name) searchParams.append('tool_name', params.tool_name);
  if (params?.status) searchParams.append('status', params.status);
  if (params?.all !== undefined) searchParams.append('all', params.all.toString());

  const queryString = searchParams.toString();
  return get(`${toolBoxIntegrationBaseUrl}/tool-box/market/tools${queryString ? `?${queryString}` : ''}`, {
    headers: {
      [businessDomainHeaderKey]: businessDomainIds?.join(',') || '',
    },
  });
};

// 获取单个工具信息
export const getToolById = (boxId: string, toolId: string, businessDomainIds?: string[]): Promise<ToolResponse> => {
  return get(`${toolBoxIntegrationBaseUrl}/tool-box/${boxId}/tool/${toolId}`, {
    headers: {
      [businessDomainHeaderKey]: businessDomainIds?.join(',') || '',
    },
  });
};

// 获取工具箱市场工具列表
export const getToolBoxMarketList = (
  params: ToolBoxMarketListParams,
  businessDomainIds?: string[]
): Promise<ToolBoxMarketListResponse> => {
  // 将数组转换为逗号分隔的字符串作为路径参数
  const boxIdsString = params.box_ids.join(',');
  return get(`${toolBoxIntegrationBaseUrl}/tool-box/market/${boxIdsString}/${params.fields}`, {
    headers: {
      [businessDomainHeaderKey]: businessDomainIds?.join(',') || '',
    },
  });
};
