import { get, post, del, put, cacheableGet } from '@/utils/http';
import qs from 'qs';
import {
  AgentConfig,
  AgentGenerationParams,
  ProcessStatuses,
  Product,
  Agent,
  Category,
  AgentDetailType,
  ToolBox,
  AgentVersion,
  SpaceListInfo,
  SpaceMemberTypeEnum,
  CreateSpaceRequestParamType,
  SpaceMemberType,
  SpaceResourceType,
  SpaceResourceEnum,
  AgentPublishToBeEnum,
  AgentPublishToWhereEnum,
  PublishData,
  PublishStatusEnum,
  AgentManagementPermType,
  PublisheAgentInfoType,
  PublishedTemplateInfoType,
  PublishedAgentInfoType,
  GetAgentsByPostRequestBody,
  DolphinTemplateType,
  ImportTypeEnum,
  ImportAgentResult,
  GetToolProcessCategoriesResponseType,
  GetToolProcessStrategiesByCategoryResponseType,
  ToolProcessCategoryType,
  ToolProcessStrategyType,
  ResultProcessStrategyType,
  SelfConfigSchemaType,
} from './type';

export {
  SpaceMemberTypeEnum,
  SpaceResourceEnum,
  AgentPublishToBeEnum,
  AgentPublishToWhereEnum,
  PublishStatusEnum,
  type AgentManagementPermType,
  type PublishData,
  type PublisheAgentInfoType,
  type PublishedTemplateInfoType,
  type DolphinTemplateType,
  ImportTypeEnum,
  type ToolProcessCategoryType,
  type ToolProcessStrategyType,
  type ResultProcessStrategyType,
};

const agentFactoryBaseUrl = '/api/agent-factory/v3';
const agentFactoryBaseUrlV2 = '/api/agent-factory/v2';
const agentConfigBaseUrl = `${agentFactoryBaseUrl}/agent`;
const agentMarketBaseUrl = `${agentFactoryBaseUrl}/agent-market`;
const toolBoxBaseUrl = '/api/agent-factory/v1';
const spaceBaseUrl = `${agentFactoryBaseUrl}/custom-space`;
const agentUrlV3 = `${agentFactoryBaseUrl}/agent`;
const templateUrlV3 = `${agentFactoryBaseUrl}/agent-tpl`;
const anyshareUrl = `${agentFactoryBaseUrl}/anyshare7ds`;

/* agent使用页面获取详情 */
export const getAgentDetailInUsagePage = ({
  id,
  version,
  is_visit = false,
  customSpaceId,
}: {
  id: string;
  version: string;
  is_visit?: boolean;
  customSpaceId?: string;
}): Promise<AgentDetailType> => {
  const queryParams: any = {
    is_visit,
  };
  if (customSpaceId) {
    queryParams.custom_space_id = customSpaceId;
  }
  return get(`${agentMarketBaseUrl}/agent/${id}/version/${version}?${qs.stringify(queryParams)}`);
};

// =================== Agent配置页面相关API =================== //

// 创建Agent
export const createAgent = (data: AgentConfig): Promise<{ id: string }> => {
  return post(agentConfigBaseUrl, { body: data });
};

// 获取agent详情
export const getAgentDetail = (id: string): Promise<AgentConfig> => {
  return get(`${agentConfigBaseUrl}/${id}`);
};

// 获取模板详情
export const getTemplateDetail = (id: string): Promise<AgentConfig> => {
  return get(`${templateUrlV3}/${id}`);
};

// 编辑Agent
export const editAgent = (id: string, data: AgentConfig) => {
  return put(`${agentConfigBaseUrl}/${id}`, { body: data });
};

// 删除Agent
export const deleteAgent = (id: string) => {
  return del(`${agentConfigBaseUrl}/${id}`);
};

// AI自动生成内容
export const aiGenerateContent = (data: AgentGenerationParams): Promise<ReadableStream<Uint8Array>> => {
  return post(`${agentConfigBaseUrl}/ai-autogen`, { body: data });
};

// AI自动生成预设问题
export const aiGeneratePresetQuestions = (data: AgentGenerationParams): Promise<string[]> => {
  return post(`${agentConfigBaseUrl}/ai-autogen`, { body: data });
};

// 批量获取数据处理状态
export const getBatchDataProcessingStatus = (data: {
  agent_uniq_flags: Array<{ agent_id: string; agent_version: string }>;
  is_show_fail_infos: boolean;
}): Promise<ProcessStatuses> => {
  return post(`${agentConfigBaseUrl}/batch-check-index-status`, { body: data });
};

// 发布智能体
export const publishAgent = (
  id: string,
  publishData?: PublishData
): Promise<{ release_id: string; version: string; published_at: number }> => {
  return post(`${agentConfigBaseUrl}/${id}/publish`, { body: publishData });
};

// 取消发布智能体
export const unpublishAgent = (id: string): Promise<void> => {
  return put(`${agentConfigBaseUrl}/${id}/unpublish`);
};

// 获取发布信息
export const getPublishedAgentInfo = (agentId: string): Promise<PublisheAgentInfoType> =>
  get(`${agentConfigBaseUrl}/${agentId}/publish-info`);

// 更新发布信息
export const updateAgentPublishInfo = (agentId: string, publishData: PublishData): Promise<{ published_at: number }> =>
  put(`${agentConfigBaseUrl}/${agentId}/publish-info`, { body: publishData });

// =================== Decision Agent市场页面相关API =================== //

// 获取最近访问的智能体
export const getRecentVisitAgents = ({
  page,
  name = '',
  size,
  custom_space_id = '',
}: {
  page: number;
  name?: string;
  size: number;
  custom_space_id?: string;
}): Promise<{ entries: Agent[]; total: number }> => {
  return get(
    `${agentFactoryBaseUrl}/recent-visit/agent?${qs.stringify({
      page,
      name,
      size,
      custom_space_id,
    })}`
  );
};

// 获取我创建的智能体
export const getMyCreatedAgentList = (params: {
  pagination_marker_str?: string;
  size: number;
  name?: string;
  publish_status?: PublishStatusEnum | '';
  publish_to_be?:
    | AgentPublishToBeEnum.ApiAgent
    | AgentPublishToBeEnum.WebSDKAgent
    | AgentPublishToBeEnum.SkillAgent
    | '';
}): Promise<{ entries: Agent[]; pagination_marker_str: string; is_last_page: boolean }> => {
  return get(`${agentFactoryBaseUrl}/personal-space/agent-list`, { params });
};

// 获取已发布的智能体列表 - POST版本
export const getAgentsByPost = (
  body: GetAgentsByPostRequestBody
): Promise<{ entries: Agent[]; pagination_marker_str: string; is_last_page: boolean }> => {
  return post(`${agentFactoryBaseUrl}/published/agent`, { body });
};

// 获取 agent 应用的api文档
export const getAgentApiDocById = (agentId: string) => get(`${agentFactoryBaseUrlV2}/agent/doc?agent_id=${agentId}`);

// 复制agent为模板并发布
export const publishAgentAsTemplate = ({
  agent_id,
  category_ids,
  business_domain_id,
}: {
  agent_id: string;
  category_ids: string[];
  business_domain_id: string;
}) =>
  post(`${agentFactoryBaseUrl}/agent/${agent_id}/copy2tpl-and-publish`, { body: { category_ids, business_domain_id } });

// 获取已发布智能体信息列表
export const getPublishedAgentInfoList = (agent_keys: string[]): Promise<{ entries: Array<PublishedAgentInfoType> }> =>
  post(`${agentFactoryBaseUrl}/published/agent-info-list`, { body: { agent_keys } });

// =================== 工具箱相关API =================== //

// 获取工具箱列表
export const getToolBoxList = (params?: {
  query?: string;
  order?: 'desc' | 'asc';
  rule?: string;
  size?: number;
  page?: number;
}): Promise<{ data: ToolBox[]; count: number }> => {
  const { query = '', order = 'desc', rule = 'create_time', size = 1000, page = 1 } = params || {};
  return get(`${toolBoxBaseUrl}/tool-boxes/list?query=${query}&order=${order}&rule=${rule}&size=${size}&page=${page}`);
};

// 获取工具箱详情
export const getToolBoxDetail = (boxId: string): Promise<ToolBox> => {
  return get(`${toolBoxBaseUrl}/tool-boxes/${boxId}`);
};

// =================== 自定义空间相关API =================== //

// 获取空间列表
export const getSpaceList = ({
  page,
  size,
  name = '',
}: {
  page: number;
  size: number;
  name?: string;
}): Promise<{ total: number; entries: SpaceListInfo[] }> =>
  get(`${spaceBaseUrl}?page=${page}&size=${size}&name=${encodeURIComponent(name)}`);

// 创建空间
export const createSpace = (body: CreateSpaceRequestParamType): Promise<void> => post(spaceBaseUrl, { body });

// 获取空间详情
export const getSpaceInfo = (id: string): Promise<SpaceListInfo> => get(`${spaceBaseUrl}/${id}`);

// 编辑空间
export const editSpace = ({ id, ...body }: { id: string; name?: string; profile?: string }): Promise<void> =>
  put(`${spaceBaseUrl}/${id}`, { body });

// 删除空间
export const deleteSpace = (id: string): Promise<void> => del(`${spaceBaseUrl}/${id}`);

// 获取空间成员列表
export const getSpaceMembers = ({
  id,
  last_id = 0,
  size,
}: {
  id: string;
  last_id?: number;
  size: number;
}): Promise<{ total: number; entries: SpaceMemberType[] }> =>
  get(`${spaceBaseUrl}/${id}/members?last_id=${last_id}&size=${size}`);

// 添加空间成员
export const addSpaceMember = ({
  id,
  members,
}: {
  id: string;
  members: { obj_type: SpaceMemberTypeEnum; obj_id: string }[];
}) => post(`${spaceBaseUrl}/${id}/members`, { body: { members } });

// 删除空间成员
export const deleteSpaceMember = ({ id, member_assoc_id }: { id: string; member_assoc_id: string }): Promise<void> =>
  del(`${spaceBaseUrl}/${id}/members/${member_assoc_id}`);

// 获取空间资源列表
export const getSpaceResources = ({
  id,
  last_id = 0,
  size,
}: {
  id: string;
  last_id: number;
  size: number;
}): Promise<{ total: number; entries: SpaceResourceType[] }> =>
  get(`${spaceBaseUrl}/${id}/resources?last_id=${last_id}&size=${size}`);

// 添加空间资源
export const addSpaceResource = ({
  id,
  resources,
}: {
  id: string;
  resources: { resource_type: SpaceResourceEnum; resource_id: string }[];
}) => post(`${spaceBaseUrl}/${id}/resources`, { body: { resources } });

// 删除空间资源
export const deleteSpaceResource = ({ id, resource_assoc_id }: { id: string; resource_assoc_id: string }) =>
  del(`${spaceBaseUrl}/${id}/resources/${resource_assoc_id}`);

// 根据资源类型和资源ID删除空间下的某个资源
export const deleteSpaceResourceByResourceId = ({
  id,
  resource_type,
  resource_id,
}: {
  id: string;
  resource_type: SpaceResourceEnum;
  resource_id: string;
}) => del(`${spaceBaseUrl}/${id}/resource-type/${resource_type}/resource-id/${resource_id}`);

// =================== 模板 API =================== //

// 复制agent
export const copyAgent = (agent_id: string): Promise<{ id: string; key: string; version: string; name: string }> =>
  post(`${agentUrlV3}/${agent_id}/copy`);

// 复制agent为模板
export const generateAsTemplate = (agent_id: string) => post(`${agentUrlV3}/${agent_id}/copy2tpl`);

// 获取已发布的模板列表
export const getPublishedTemplateList = (params: {
  name?: string;
  category_id?: string;
  pagination_marker_str?: string;
  size: number;
}) => get(`${agentFactoryBaseUrl}/published/agent-tpl`, { params });

// 获取个人空间下的模板列表
export const getMyTemplateList = (params: {
  pagination_marker_str?: string;
  size: number;
  name?: string;
  publish_status?: PublishStatusEnum;
}) => get(`${agentFactoryBaseUrl}/personal-space/agent-tpl-list`, { params });

// 复制智能体模板
export const copyTemplate = (id: string): Promise<{ id: string; key: string; name: string }> =>
  post(`${templateUrlV3}/${id}/copy`);

// 发布模板
export const publishTemplate = ({
  id,
  category_ids,
  business_domain_id,
}: {
  id: string;
  category_ids: string[];
  business_domain_id: string;
}): Promise<{ published_at: number }> =>
  post(`${templateUrlV3}/${id}/publish`, { body: { category_ids, business_domain_id } });

// 取消发布模板
export const unpublishTemplate = (id: string) => put(`${templateUrlV3}/${id}/unpublish`);

// 删除模板
export const deleteTemplate = (id: string) => del(`${templateUrlV3}/${id}`);

// 获取模板发布信息
export const getPublishedTemplateInfo = (id: string): Promise<PublishedTemplateInfoType> =>
  get(`${templateUrlV3}/${id}/publish-info`);

// 编辑模板
export const editTemplate = (id: string, data: AgentConfig) => {
  return put(`${templateUrlV3}/${id}`, { body: data });
};

// 获取已发布模板详情
export const getPublishedTemplateDetail = (tpl_id: string): Promise<AgentDetailType> =>
  get(`${agentFactoryBaseUrl}/published/agent-tpl/${tpl_id}`);

// =================== Agent其他相关API =================== //

// 获取产品列表
export const getProductList = (): Promise<{ entries: Product[]; total: number }> => {
  return get(`${agentFactoryBaseUrl}/product`, {});
};

// 获取智能体分类
export const getAgentCategoryList = (): Promise<Category[]> => {
  return get(`${agentFactoryBaseUrl}/category`);
};

// 获取智能体版本列表
export const getAgentVersionList = (id: string): Promise<{ entries: AgentVersion[]; total: number }> => {
  return get(`${agentFactoryBaseUrl}/agent/${id}/release-history`);
};

// 获取用户拥有的管理权限状态
export const getAgentManagementPerm = (): Promise<AgentManagementPermType> =>
  cacheableGet(`${agentFactoryBaseUrl}/agent-permission/management/user-status`, { expires: 1000 * 60 * 5 });

/** 获取所有的文件扩展名 */
export const getAllFileExt = () => {
  return get(`${agentFactoryBaseUrl}/agent/temp-zone/file-ext-map`);
};

/** 获取dolphin模板列表。config和新建/编辑agent的传参相比，少了pre_dolphin和post_dolphin */
export const getDolphinTemplateList = (body: {
  config: any;
  built_in_agent_key?: string;
}): Promise<{
  pre_dolphin: DolphinTemplateType[];
  post_dolphin: DolphinTemplateType[];
}> => post(`${agentFactoryBaseUrl}/agent/dolphin-tpl/list`, { body });

/** 导出agent */
export const exportAgent = (agent_ids: string[]): Promise<any> =>
  post(`${agentFactoryBaseUrl}/agent-inout/export`, {
    body: { agent_ids },
    returnFullResponse: true,
    timeout: 3 * 60 * 1000,
  });

/** 导入agent */
export const importAgent = (body: FormData): Promise<ImportAgentResult> =>
  post(`${agentFactoryBaseUrl}/agent-inout/import`, { body, timeout: 3 * 60 * 1000 });

/** 获取结果处理策略分类列表 */
export const getToolProcessCategories = (): Promise<GetToolProcessCategoriesResponseType> =>
  cacheableGet(`${agentFactoryBaseUrl}/tool-result-process-strategy/category`, { expires: 1000 * 60 * 5 });

/** 根据分类获取结果处理策略列表 */
export const getToolProcessStrategiesByCategoryId = (
  id: string
): Promise<GetToolProcessStrategiesByCategoryResponseType> =>
  cacheableGet(`${agentFactoryBaseUrl}/tool-result-process-strategy/category/${id}/strategy`, {
    expires: 1000 * 60 * 5,
  });

export const getInFoByPath = async (body: any) => {
  return post(`${anyshareUrl}/getinfobypath`, { body });
};

export const getDirList = async (body: any) => {
  return post(`${anyshareUrl}/dir/list`, { body });
};

// 获取SELF_CONFIG字段结构
export const getSelfConfigSchema = (): Promise<SelfConfigSchemaType> =>
  cacheableGet(`${agentFactoryBaseUrl}/agent-self-config-fields`, {
    expires: 1000 * 60 * 5,
  });
