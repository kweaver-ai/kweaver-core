import intl from 'react-intl-universal';
import { Breadcrumb } from 'antd';
import { Agent } from '@/apis/agent-factory/type';
import {
  getAgentsByPost,
  PublishStatusEnum,
  getPublishedTemplateList,
  getMyCreatedAgentList,
  getMyTemplateList,
  AgentPublishToBeEnum,
  type AgentManagementPermType,
} from '@/apis/agent-factory';
import { AgentActionEnum, TemplateActionEnum, ModeEnum } from './types';

const PublishedStatus = 'published';

// 获取正确的用户信息
export const getUserInfo = (agent: Agent) => {
  return agent.status === PublishedStatus
    ? {
        user_id: agent?.published_by,
        username: agent?.published_by_name,
      }
    : {
        user_id: agent?.updated_by,
        username: agent?.updated_by_name,
      };
};

// Decision Agent-操作菜单
const getDataAgentMenuItems = ({
  agent,
}: {
  mode: ModeEnum;
  agent: Agent;
  perms: AgentManagementPermType;
  customSpaceInfo: { name: string; created_by: string };
  currentUserId: string;
}): any[] => {
  return [
    {
      key: AgentActionEnum.ViewAPI,
      label: intl.get('dataAgent.viewAPI'),
      hidden: agent.publish_info?.is_api_agent === 0, // 未发布到API的，不显示此操作
    },
  ].filter(item => !item.hidden);
};

// 自定义空间 Agent-操作菜单
const getCustomSpaceAgentMenuItems = ({
  agent,
  customSpaceInfo,
  currentUserId,
}: {
  mode: ModeEnum;
  agent: Agent;
  perms: AgentManagementPermType;
  customSpaceInfo: { name: string; created_by: string };
  currentUserId: string;
}): any[] => {
  // 未发布到API的，不显示【查看API】
  const hiddenViewAPI = agent.publish_info?.is_api_agent === 0;
  // 不是空间的创建者，不显示【移除】
  const hiddenRemoveAgent = customSpaceInfo?.created_by !== currentUserId;
  return [
    {
      key: AgentActionEnum.ViewAPI,
      label: intl.get('dataAgent.viewAPI'),
      hidden: hiddenViewAPI,
    },
    {
      type: 'divider',
      hidden: hiddenViewAPI || hiddenRemoveAgent, // 查看API和移除都显示的时候，才显示
    },
    {
      key: AgentActionEnum.RemoveAgent,
      label: intl.get('dataAgent.remove'),
      danger: true,
      hidden: hiddenRemoveAgent,
    },
  ].filter(item => !item.hidden);
};

// 全部模板-操作菜单
const getAllTemplateMenuItems = (): any[] => {
  return [
    { key: TemplateActionEnum.CreateAgentFromTemplate, label: intl.get('dataAgent.useTemplateToCreate') },
    { key: TemplateActionEnum.TemplateConfig, label: intl.get('dataAgent.configInfo') },
  ];
};

// 获取我的模板 卡片的操作菜单
const getMyTemplateDropdownMenuItems = ({
  agent,
  perms,
}: {
  mode: ModeEnum;
  agent: Agent;
  perms: AgentManagementPermType;
  customSpaceInfo: { name: string; created_by: string };
  currentUserId: string;
}) => {
  const isPublished = agent.status === PublishedStatus;

  return [
    { key: TemplateActionEnum.CopyTemplate, label: intl.get('dataAgent.copy') },
    { key: TemplateActionEnum.TemplateConfig, label: intl.get('dataAgent.configInfo') },
    {
      key: TemplateActionEnum.PublishTemplate,
      label: intl.get('dataAgent.publishTemplate'),
      hidden: isPublished || !perms?.agent_tpl?.publish, // 已发布时，隐藏【发布模板】；没有发布权限时，隐藏
    },
    {
      key: TemplateActionEnum.UnpublishTemplate,
      label: intl.get('dataAgent.unpublish'),
      hidden: !agent.published_at || !perms?.agent_tpl?.unpublish, // 当无发布时间时，隐藏；没有取消发布权限时，隐藏
    },
    { type: 'divider' },
    {
      key: TemplateActionEnum.DeleteTemplate,
      label: intl.get('dataAgent.delete'),
      danger: true,
      disabled: isPublished, // 已发布的，禁用【删除】
    },
  ].filter(item => !item.hidden);
};

// 获取我的创建 卡片的操作菜单
const getMyAgentDropdownMenu = ({
  agent,
  perms,
}: {
  mode: ModeEnum;
  agent: Agent;
  perms: AgentManagementPermType;
  customSpaceInfo: { name: string; created_by: string };
  currentUserId: string;
}): any[] => {
  const isPublished = agent.status === PublishedStatus;
  // 是否内置agent
  const isBuiltIn = agent.is_built_in === 1;

  return [
    { key: AgentActionEnum.Use, label: intl.get('dataAgent.use') },
    { key: AgentActionEnum.Copy, label: intl.get('dataAgent.copy') },
    { key: AgentActionEnum.ConfigInfo, label: intl.get('dataAgent.configInfo') },
    {
      key: AgentActionEnum.Publish,
      label: intl.get('dataAgent.publish'),
      disabled: agent.version === 'v0' && agent.is_built_in === 1,
      hidden: isPublished || !perms?.agent?.publish, // 已发布时，隐藏【发布】；没有发布权限时，隐藏【发布】
    },
    {
      key: AgentActionEnum.UpdatePublishInfo,
      label: intl.get('dataAgent.updatePublishInfo'),
      hidden: !isPublished || !perms?.agent?.publish, // 未发布时，隐藏【更新发布信息】；没有发布权限时，隐藏【更新发布】
    },
    {
      key: AgentActionEnum.Unpublish,
      label: intl.get('dataAgent.unpublish'),
      hidden: !agent.published_at || !perms?.agent?.unpublish, // 没有发布时间时隐藏；没有 取消发布权限时隐藏此按钮
    },
    {
      key: AgentActionEnum.PublishAsTemplate,
      label: intl.get('dataAgent.publishAsTemplate'),
      hidden: !perms?.agent_tpl?.publish, // 没有发布模板的权限时，隐藏此按钮
    },
    {
      key: AgentActionEnum.ViewAPI,
      label: intl.get('dataAgent.viewAPI'),
      hidden: agent.publish_info?.is_api_agent === 0, // 未发布到API的，不显示此操作
    },
    { type: 'divider' },
    {
      key: AgentActionEnum.Delete,
      label: intl.get('dataAgent.delete'),
      danger: true,
      disabled: isPublished || isBuiltIn, // 已发布的agent，禁用【删除】;内置agent不允许删除
    },
  ].filter(item => !item.hidden);
};

// 获取API Agent页面的卡片下拉菜单
const getAPIDropdownMenuItems = () => {
  return [
    {
      key: AgentActionEnum.ViewAPI,
      label: intl.get('dataAgent.viewAPI'),
    },
  ];
};

export const getMenuItems = (params: {
  mode: ModeEnum;
  agent: Agent;
  perms: AgentManagementPermType;
  customSpaceInfo: { name: string; created_by: string };
  currentUserId: string;
}) => {
  switch (params.mode) {
    case ModeEnum.DataAgent:
      return getDataAgentMenuItems(params);

    case ModeEnum.CustomSpace:
      return getCustomSpaceAgentMenuItems(params);

    case ModeEnum.AllTemplate:
      return getAllTemplateMenuItems();

    case ModeEnum.MyAgent:
      return getMyAgentDropdownMenu(params);

    case ModeEnum.MyTemplate:
      return getMyTemplateDropdownMenuItems(params);

    case ModeEnum.API:
      return getAPIDropdownMenuItems();

    default:
      return [];
  }
};

export const getHeaderText = (
  mode: ModeEnum,
  { customSpaceName, navigateToSpaces }: { customSpaceName: string; navigateToSpaces: () => void }
) => {
  const texts: any = {
    [ModeEnum.DataAgent]: intl.get('dataAgent.agent'),
    [ModeEnum.CustomSpace]: (
      <Breadcrumb
        items={[
          {
            title: (
              <span className="dip-pointer dip-font-weight-700 dip-font-16">{intl.get('dataAgent.customSpace')}</span>
            ),
            onClick: navigateToSpaces,
          },
          {
            title: (
              <div className="dip-font-16 dip-ellipsis" title={customSpaceName}>
                {customSpaceName}
              </div>
            ),
          },
        ]}
        separator=">"
      />
    ),
    [ModeEnum.MyAgent]: intl.get('dataAgent.myCreation'),
    [ModeEnum.AllTemplate]: intl.get('dataAgent.template'),
    [ModeEnum.MyTemplate]: intl.get('dataAgent.myCreation'),
    [ModeEnum.API]: 'API',
  };

  return texts[mode] || '';
};

export const getSearchPlaceholder = (mode: ModeEnum) => {
  const texts: any = {
    [ModeEnum.DataAgent]: intl.get('dataAgent.searchAgentName'),
    [ModeEnum.CustomSpace]: intl.get('dataAgent.searchAgentName'),
    [ModeEnum.MyAgent]: intl.get('dataAgent.searchAgentName'),
    [ModeEnum.AllTemplate]: intl.get('dataAgent.searchTemplateName'),
    [ModeEnum.MyTemplate]: intl.get('dataAgent.searchTemplateName'),
    [ModeEnum.API]: intl.get('dataAgent.searchApiName'),
  };

  return texts[mode] || '';
};

export const fetchData = async ({
  mode,
  category_id,
  size,
  name,
  custom_space_id = '',
  publish_status,
  publish_to_be,
  pagination_marker_str,
}: {
  mode: ModeEnum;
  category_id: string;
  size: number;
  name: string;
  custom_space_id?: string;
  publish_status: PublishStatusEnum;
  publish_to_be: AgentPublishToBeEnum;
  pagination_marker_str?: string;
}) => {
  let entries, nextPaginationMarkerStr, is_last_page;

  switch (mode) {
    case ModeEnum.DataAgent:
    case ModeEnum.CustomSpace:
      ({
        entries,
        pagination_marker_str: nextPaginationMarkerStr,
        is_last_page,
      } = await getAgentsByPost({
        pagination_marker_str,
        category_id,
        size,
        name,
        custom_space_id,
        is_to_square: mode === ModeEnum.DataAgent ? 1 : 0,
      }));
      entries = entries.map(item => ({ ...item, status: PublishedStatus }));
      break;

    case ModeEnum.MyAgent:
      ({
        entries,
        pagination_marker_str: nextPaginationMarkerStr,
        is_last_page,
      } = await getMyCreatedAgentList({
        pagination_marker_str,
        size,
        name,
        publish_status,
        publish_to_be: publish_to_be as any,
      }));
      break;

    case ModeEnum.AllTemplate:
      ({
        entries,
        pagination_marker_str: nextPaginationMarkerStr,
        is_last_page,
      } = await getPublishedTemplateList({
        pagination_marker_str,
        category_id,
        size,
        name,
      }));
      entries = entries.map(item => ({ ...item, status: PublishedStatus }));
      break;

    case ModeEnum.MyTemplate:
      ({
        entries,
        pagination_marker_str: nextPaginationMarkerStr,
        is_last_page,
      } = await getMyTemplateList({
        pagination_marker_str,
        size,
        name,
        publish_status,
      }));
      break;

    case ModeEnum.API:
      ({
        entries,
        pagination_marker_str: nextPaginationMarkerStr,
        is_last_page,
      } = await getAgentsByPost({
        pagination_marker_str,
        category_id,
        size,
        name,
        publish_to_be: AgentPublishToBeEnum.ApiAgent,
      }));
      entries = entries.map(item => ({ ...item, status: PublishedStatus }));
      break;

    default:
      entries = [];
  }

  return { entries, nextPaginationMarkerStr: nextPaginationMarkerStr || '', is_last_page };
};
