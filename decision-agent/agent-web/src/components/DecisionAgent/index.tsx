import { useState, useEffect, useRef, useMemo, useCallback } from 'react';
import { useNavigate, useParams, useLocation } from 'react-router-dom';
import _, { uniq } from 'lodash';
import classNames from 'classnames';
import intl from 'react-intl-universal';
import { Row, Col, Button, message, Spin, Modal, Empty, Select, Tooltip, Carousel } from 'antd';
import { LeftOutlined, RightOutlined, ReloadOutlined, SyncOutlined, VerticalAlignTopOutlined } from '@ant-design/icons';
import {
  getRecentVisitAgents,
  getAgentCategoryList,
  deleteAgent,
  unpublishAgent,
  getBatchDataProcessingStatus,
  copyAgent,
  copyTemplate,
  PublishStatusEnum,
  getSpaceInfo,
  AgentPublishToBeEnum,
  unpublishTemplate,
  deleteTemplate,
  getAgentManagementPerm,
  type AgentManagementPermType,
  deleteSpaceResourceByResourceId,
  SpaceResourceEnum,
  getMyTemplateList,
  exportAgent,
} from '@/apis/agent-factory';
import type { Agent, Category } from '@/apis/agent-factory/type';
import PlusIcon from '@/assets/icons/plus.svg';
import { useUserAvatars } from '@/hooks/useUserAvatars';
import { useMicroWidgetProps } from '@/hooks';
import empty from '@/assets/images/empty.png';
import ExportIcon from '@/assets/icons/export.svg';
import ImportIcon from '@/assets/icons/import.svg';
import { downloadFile, getFilenameFromContentDisposition } from '@/utils/file';
import { useSize } from '@/hooks';
import { formatTimeSlash } from '@/utils/handle-function/FormatTime';
import { PublishSettingsModal, PublishModeEnum } from '@/components/AgentPublish';
import LoadFailed from '@/components/LoadFailed';
import SpaceAgentAddButton from '@/components/SpaceAgentAddButton';
import BaseCard, { computeColumnCount, gap } from '@/components/BaseCard';
import SkeletonGrid from '@/components/BaseCard/SkeletonGrid';
import GradientContainer from '../GradientContainer';
import MyCreatedTab from './MyCreatedTab';
import { handleImportAgent } from './import-agent';
import { getMenuItems, getSearchPlaceholder, fetchData, getUserInfo } from './utils';
import { ModeEnum, AgentActionEnum, TemplateActionEnum } from './types';
import styles from './index.module.less';
import qs from 'qs';
import Header from './Header';
import ResizeObserver from '@/components/ResizeObserver';
import type { CarouselRef } from 'antd/es/carousel';
import { useInfiniteScroll } from 'ahooks';
import SearchInput from '@/components/SearchInput';
import { getParam, isJSONString } from '@/utils/handle-function';
import { nanoid } from 'nanoid';

export { ModeEnum };

interface DataAgentsProps {
  mode?: ModeEnum;
}

const DecisionAgent = ({ mode: modeFromProps = ModeEnum.DataAgent }: DataAgentsProps) => {
  const filterParams = useMemo(() => {
    const data = getParam('filterParams');
    const decodedData = decodeURIComponent(data);
    // console.log('filterParams', decodedData);
    if (isJSONString(decodedData)) {
      const url = new URL(window.location.href);
      url.searchParams.delete('filterParams');
      window.history.replaceState({}, '', url.toString());
      return JSON.parse(decodedData);
    }
    return {};
  }, []);


  const publishModeRef = useRef<PublishModeEnum | undefined>(undefined);
  const nextPaginationMarkerStrRef = useRef<string>(''); // 分页marker，用于获取下一批数据
  // 分类-外部容器
  const containerRef = useRef<HTMLDivElement | null>(null);
  // 分类-可滚动容器
  const scrollableRef = useRef<HTMLDivElement | null>(null);
  // 分类-当前滚动位置
  const scrollPositionRef = useRef<number>(0);

  const { width: containerWidth } = useSize(containerRef.current);
  const { width: contentWidth } = useSize(scrollableRef.current);

  const { customSpaceId } = useParams();
  const navigate = useNavigate();
  const location = useLocation();
  const microWidgetProps = useMicroWidgetProps();
  const [modal, contextHolder] = Modal.useModal();
  const [recentLoading, setRecentLoading] = useState(true);
  const [selectedCategory, setSelectedCategory] = useState<Category>({
    category_id: filterParams.categoryId ?? '',
    name: '全部',
  });
  const [recentAgents, setRecentAgents] = useState<Agent[]>([]);
  const [categories, setCategories] = useState<Category[]>([]);
  const [recentError, setRecentError] = useState<string | null>(null);
  const [categoryError, setCategoryError] = useState<string | null>(null);
  const [mode, setMode] = useState<ModeEnum>(filterParams?.mode ?? modeFromProps);
  // 在我的创建页面是否显示【DataAgent】、【我的模板】的Tab
  const [showMyCreatedTab, setShowMyCreatedTab] = useState<boolean>(false);
  // 分类-箭头是否显示
  const [showScrollArrows, setShowScrollArrows] = useState<{ left: boolean; right: boolean }>({
    left: false,
    right: false,
  });

  // 搜索相关状态
  const [searchName, setSearchName] = useState<string>(filterParams?.searchValue ?? '');
  const [publishStatusFilter, setPublishStatusFilter] = useState<PublishStatusEnum>(
    filterParams?.publishStatus ?? PublishStatusEnum.All
  );
  const [publishCategoryFilter, setPublishCategoryFilter] = useState<AgentPublishToBeEnum>(
    filterParams?.publishCategory ?? AgentPublishToBeEnum.All
  );

  // 发布设置相关状态
  const [publishModalVisible, setPublishModalVisible] = useState(false);
  const [selectedAgent, setSelectedAgent] = useState<Agent | null>(null);

  // 自定义空间信息
  const [customSpaceInfo, setCustomSpaceInfo] = useState<{ name: string; created_by: string }>({
    name: '',
    created_by: '',
  });

  // 权限
  const [perms, setPerms] = useState<AgentManagementPermType>({
    custom_space: {
      create: false,
    },
    agent: {
      publish: false,
      unpublish: false,
      unpublish_other_user_agent: false,
      publish_to_be_api_agent: false,
      publish_to_be_data_flow_agent: false,
      publish_to_be_skill_agent: false,
      publish_to_be_web_sdk_agent: false,
      create_system_agent: false,
    },
    agent_tpl: {
      publish: false,
      unpublish: false,
      unpublish_other_user_agent_tpl: false,
    },
  });

  // 导出模式
  const [isExportMode, setIsExportMode] = useState<boolean>(false);
  // 被选中用于导出的项目id
  const [selectedIdsForExport, setSelectedIdsForExport] = useState<string[]>([]);

  const { userAvatars, addUserIds } = useUserAvatars([], 20);

  // 添加处理状态相关状态
  const [processingStatuses, setProcessingStatuses] = useState<{ [key: string]: number }>({});
  const processingStatusTimer = useRef<NodeJS.Timeout | null>(null);

  const agentListPageSize = 48;

  // 高亮的agent/templage id
  const [highlightId, setHighlightId] = useState<string>('');

  const [countOfRow, setCountOfRow] = useState<number>(0);
  const recentAgentSlideRef = useRef<CarouselRef | null>(null);
  const [recentAgentSlideIndex, setRecentAgentSlideIndex] = useState<number>(-1);

  // 是否显示最近访问：只有Decision Agent页面显示
  const showRecent = useMemo(() => mode === ModeEnum.DataAgent, [mode]);

  const pageContainerRef = useRef<HTMLDivElement>(null);
  const sectionRef = useRef<HTMLElement>(null);

  // 是否显示分类: Decision Agent、模板、自定义空间、API Agent页面显示
  const showCategory = useMemo(
    () => [ModeEnum.DataAgent, ModeEnum.AllTemplate, ModeEnum.CustomSpace, ModeEnum.API].includes(mode),
    [mode]
  );

  // 搜索框的placeholder
  const searchPlaceholder = useMemo(() => getSearchPlaceholder(mode), [mode]);

  // 当前登录的用户id
  const currentUserId = useMemo(() => microWidgetProps?.config?.userInfo?.id, [microWidgetProps]);

  const [isStuck, setIsStuck] = useState(false);
  const stickySentinelRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (showRecent) {
      const observer = new IntersectionObserver(
        ([entry]) => {
          // 当哨兵离开视图（即滚出容器顶部）时，认为已吸顶
          setIsStuck(!entry.isIntersecting);
        },
        {
          root: pageContainerRef.current,
          threshold: [1],
        }
      );

      if (stickySentinelRef.current) {
        observer.observe(stickySentinelRef.current);
      }

      return () => observer.disconnect();
    }
  }, []);

  const getFilterParams = () => {
    const filterParams: any = {};
    if (searchName) {
      filterParams.searchValue = searchName;
    }
    if (publishStatusFilter) {
      filterParams.publishStatus = publishStatusFilter;
    }
    if (publishCategoryFilter) {
      filterParams.publishCategory = publishCategoryFilter;
    }
    if (mode) {
      filterParams.mode = mode;
    }
    if (selectedCategory.category_id) {
      filterParams.categoryId = selectedCategory.category_id;
    }
    return filterParams;
  };

  // 跳转到agent的使用界面
  const navigateToAgentUsage = ({ id, version, status }: { id: string; version: string; status: string }) => {
    const filterParams = getFilterParams();
    const params: any = {
      id,
      version: status === 'unpublished' ? 'v0' : version,
      agentAppType: 'common',
      preRoute: location.pathname,
      customSpaceId,
    };
    if (!_.isEmpty(filterParams)) {
      params.filterParams = encodeURIComponent(JSON.stringify(filterParams));
    }
    // 自定义空间的agent列表页面，跳转到agent使用页面，使用相对路径导航：前面不用加 /
    navigate(`${mode === ModeEnum.CustomSpace ? '' : '/'}usage?${qs.stringify(params)}`);
  };

  // 跳转到agent的编辑界面
  const navigateToAgentEditConfig = ({ id }: { id: string }) => {
    let url = `/config?agentId=${id}`;
    const filterParams = getFilterParams();
    if (!_.isEmpty(filterParams)) {
      url += `&filterParams=${encodeURIComponent(JSON.stringify(filterParams))}`;
    }
    navigate(url);
  };

  useEffect(() => {
    if (showRecent) {
      fetchRecentAgents();
    }
  }, []);

  // 获取最近访问数据函数
  const fetchRecentAgents = async () => {
    try {
      setRecentLoading(true);
      setRecentError(null);
      // 最近访问，只获取20条数据
      const { entries } = await getRecentVisitAgents({
        page: 1,
        size: 20,
      });

      if (entries && entries.length > 0) {
        setRecentAgents(entries);
        // 批量获取用户头像
        if ([ModeEnum.DataAgent, ModeEnum.AllTemplate].includes(mode)) {
          // Decision Agent、全部模板，要显示用户头像
          const userIds = uniq(entries.map(agent => getUserInfo(agent).user_id).filter(userId => userId));
          addUserIds(userIds);
        }
      }
    } catch (ex: any) {
      if (ex?.description) {
        message.error(ex.description);
      }
      setRecentError('获取最近访问数据时发生错误');
    } finally {
      setRecentLoading(false);
    }
  };

  const fetchCategoryAgentsList = async () =>
    new Promise(async resolve => {
      try {
        const { entries, nextPaginationMarkerStr, is_last_page } = await fetchData({
          pagination_marker_str: nextPaginationMarkerStrRef.current,
          mode,
          category_id: selectedCategory.category_id,
          size: agentListPageSize,
          name: searchName,
          custom_space_id: customSpaceId,
          publish_status: publishStatusFilter,
          publish_to_be: publishCategoryFilter,
        });
        nextPaginationMarkerStrRef.current = nextPaginationMarkerStr;

        // 已发布的页面，才显示用户头像
        if ([ModeEnum.DataAgent, ModeEnum.AllTemplate, ModeEnum.CustomSpace].includes(mode)) {
          const userIds = uniq(entries.map((agent: any) => getUserInfo(agent).user_id).filter((userId: any) => userId));
          addUserIds(userIds as any);
        }
        resolve({ list: entries, isNoMore: is_last_page || entries.length === 0 });
      } catch (ex: any) {
        if (ex?.description) {
          message.error(ex.description);
        }
        setCategoryError('获取数据时发生错误');
        resolve({ list: [], isNoMore: true });
      }
    });

  const {
    data: agentList,
    loading: agentListInitLoading,
    loadingMore: agentListLoadingMore,
    reload: refreshCategoryList,
    noMore: agentListNoMore,
    mutate: forceUpdateAgentList,
  } = useInfiniteScroll(d => fetchCategoryAgentsList(d), {
    target: pageContainerRef,
    isNoMore: (d: any) => d?.isNoMore,
    threshold: 100,
    reloadDeps: [
      searchName,
      selectedCategory.category_id,
      publishStatusFilter,
      publishCategoryFilter,
      mode,
      customSpaceId,
    ],
  });

  // 重试函数
  const retryRecentAgents = () => {
    fetchRecentAgents();
  };

  const retryCategoryAgents = () => {
    refreshCategoryList();
  };

  // 获取分类数据
  useEffect(() => {
    const fetchCategoryList = async () => {
      try {
        const categories = await getAgentCategoryList();
        setCategories(categories);
      } catch (ex: any) {
        if (ex?.description) {
          message.error(ex.description);
        }
        setCategoryError('获取数据时发生错误');
      }
    };

    if (showCategory) {
      // 只有已发布的 DataAgent、全部模板、自定义空间-agent页面 会展示分类信息
      fetchCategoryList();
    }
  }, []);

  // 处理分类点击
  const handleCategoryClick = (category: Category) => {
    nextPaginationMarkerStrRef.current = '';
    setSelectedCategory(category);
  };

  // 处理创建新Agent的点击
  const handleCreateClick = () => {
    navigate('/config');
  };

  // 设置高亮（场景：比如复制之后，当前页面刷新了，为了让用户能看到复制出来的agent/template，将其设置为高亮显示）
  const changeHighlightId = (id: string) => {
    setHighlightId(id);

    // 2秒后 高亮消失
    setTimeout(() => {
      setHighlightId('');
    }, 2000);
  };

  // 处理菜单点击
  const handleMenuClick = async ({ key }: { key: string }, agent: Agent) => {
    switch (key) {
      case AgentActionEnum.Copy:
        // 复制agent
        try {
          const { id } = await copyAgent(agent.id);
          message.success(intl.get('dataAgent.operationSuccess'));

          if (mode === ModeEnum.MyAgent) {
            // 我的创建页面，需要刷新列表
            nextPaginationMarkerStrRef.current = '';
            await refreshCategoryList();
            // 高亮生成的agent
            changeHighlightId(id);
          }
        } catch (ex: any) {
          if (ex?.description) {
            message.error(ex.description);
          }
        }
        break;

      case AgentActionEnum.ViewAPI:
        // 查看API
        microWidgetProps?.history.navigateToMicroWidget({
          name: 'agent-list',
          path: `/api-doc?id=${agent?.id}&name=${encodeURIComponent(agent?.name)}&version=${agent?.version}&hidesidebar=true&hideHeaderPath=true`,
          isNewTab: true,
        });
        return;

      case AgentActionEnum.RemoveAgent:
        // 移除Agent（自定义空间）:
        try {
          await deleteSpaceResourceByResourceId({
            id: customSpaceId!,
            resource_type: SpaceResourceEnum.DataAgent,
            resource_id: agent.id,
          });
          message.success(intl.get('dataAgent.operationSuccess'));
          nextPaginationMarkerStrRef.current = '';
          await refreshCategoryList();
        } catch (ex: any) {
          if (ex?.description) {
            message.error(ex.description);
          }
        }

        break;

      case AgentActionEnum.Use:
        // 去使用
        navigateToAgentUsage(agent);
        break;

      case AgentActionEnum.ConfigInfo:
        // 跳转到配置信息详情页面
        microWidgetProps?.history.navigateToMicroWidget({
          name: 'my-agent-list',
          path: `/detail/${agent.id}?hidesidebar=true&hideHeaderPath=true`,
          isNewTab: true,
        });
        break;

      case TemplateActionEnum.TemplateConfig:
        // 跳转到配置信息详情页面
        microWidgetProps?.history.navigateToMicroWidget({
          name: mode === ModeEnum.MyTemplate ? 'my-agent-list' : 'agent-template',
          path: `/template-detail/${agent.tpl_id ?? agent.id}?hidesidebar=true&hideHeaderPath=true`,
          isNewTab: true,
        });
        break;

      case AgentActionEnum.Publish:
        // 打开发布设置弹窗
        setSelectedAgent(agent);
        publishModeRef.current = PublishModeEnum.PublishAgent;
        setPublishModalVisible(true);
        break;

      case AgentActionEnum.UpdatePublishInfo:
        // 更新发布信息
        setSelectedAgent(agent);
        publishModeRef.current = PublishModeEnum.UpdatePublishAgent;
        setPublishModalVisible(true);
        break;

      case AgentActionEnum.Unpublish:
        // 取消发布Agent确认
        modal.confirm({
          title: intl.get('dataAgent.confirmCancelPublish'),
          content: intl.get('dataAgent.confirmUnpublishAgent'),
          centered: true,
          okButtonProps: { className: 'dip-min-width-72' },
          cancelButtonProps: { className: 'dip-min-width-72' },
          onOk() {
            const hide = message.loading(intl.get('dataAgent.cancelPublishing'), 0);
            unpublishAgent(agent.id)
              .then(() => {
                hide();
                message.success(intl.get('dataAgent.operationSuccess'));

                if (!_.isEmpty(agentList?.list)) {
                  const newAgentList = _.cloneDeep(agentList);
                  newAgentList.list = newAgentList.list.map((item: any) =>
                    item.id === agent.id
                      ? {
                          ...item,
                          status: 'unpublished',
                          // 取消发布成功后，需要重置publish_info
                          publish_info: { is_api_agent: 0, is_web_sdk_agent: 0, is_skill_agent: 0 },
                          published_at: 0,
                        }
                      : item
                  );
                  forceUpdateAgentList(newAgentList);
                }
              })
              .catch((ex: any) => {
                hide();
                if (ex?.description) {
                  message.error(ex.description);
                }
              });
          },
          onCancel() {},
          footer: (_, { OkBtn, CancelBtn }) => (
            <div className="dip-flex-content-end">
              <OkBtn />
              <CancelBtn />
            </div>
          ),
        });
        break;

      case AgentActionEnum.PublishAsTemplate:
        // 将agent发布为模板
        setSelectedAgent(agent);
        publishModeRef.current = PublishModeEnum.PublishAgentAsTemplate;
        setPublishModalVisible(true);
        break;

      case AgentActionEnum.Delete:
        // 检查是否已发布
        if (agent.status === 'published') {
          // 如果已发布，提示需要先取消发布
          modal.warning({
            title: intl.get('dataAgent.cannotDelete'),
            content: intl.get('dataAgent.cannotDeleteWithPublished', { name: agent.name }),
            okText: intl.get('dataAgent.gotIt'),
            centered: true,
          });
          return;
        }

        // 删除Agent确认
        modal.confirm({
          title: intl.get('dataAgent.confirmDelete'),
          content: intl.get('dataAgent.confirmDeleteAgent', { name: agent.name }),
          centered: true,
          okButtonProps: { className: 'dip-min-width-72' },
          cancelButtonProps: { className: 'dip-min-width-72' },
          onOk() {
            const hide = message.loading(intl.get('dataAgent.deleting'), 0);
            deleteAgent(agent.id)
              .then(async () => {
                hide();
                message.success(intl.get('dataAgent.operationSuccess'));

                // 重新获取分类数据，根据总条数重置分页
                try {
                  nextPaginationMarkerStrRef.current = '';
                  refreshCategoryList();
                } catch (ex: any) {
                  if (ex?.description) {
                    message.error(ex.description);
                  }
                  setCategoryError('获取数据时发生错误');
                }

                // 如果删除的Agent在最近访问列表中，也更新最近访问列表
                if (recentAgents.some(recent => recent.id === agent.id)) {
                  fetchRecentAgents();
                }
              })
              .catch((ex: any) => {
                hide();
                if (ex?.description) {
                  message.error(ex.description);
                }
              });
          },
          onCancel() {},
          footer: (_, { OkBtn, CancelBtn }) => (
            <div className="dip-flex-content-end">
              <OkBtn />
              <CancelBtn />
            </div>
          ),
        });
        break;

      case TemplateActionEnum.CreateAgentFromTemplate: {
        const url = `/config?templateId=${(agent as any).tpl_id}&mode=createAgent`;
        // const filterParams = getFilterParams();
        // if (!_.isEmpty(filterParams)) {
        //   url += `&filterParams=${encodeURIComponent(JSON.stringify(filterParams))}`;
        // }
        // 使用此模板创建Agent
        // navigate(url);
        microWidgetProps?.history.navigateToMicroWidget({
          name: 'my-agent-list',
          path: url,
          isNewTab: true,
        });
        break;
      }

      case TemplateActionEnum.CopyTemplate:
        // 复制模板
        try {
          const { id } = await copyTemplate(agent.id);
          message.success(intl.get('dataAgent.operationSuccess'));

          // 我的模板页面，需要刷新列表
          nextPaginationMarkerStrRef.current = '';
          await refreshCategoryList();
          // 高亮生成的模板
          changeHighlightId(id);
        } catch (ex: any) {
          if (ex?.description) {
            message.error(ex.description);
          }
        }
        break;

      case TemplateActionEnum.PublishTemplate:
        // 发布模板
        setSelectedAgent(agent);
        publishModeRef.current = PublishModeEnum.PublishTemplate;
        setPublishModalVisible(true);
        break;

      case TemplateActionEnum.UnpublishTemplate:
        // 取消发布模板
        modal.confirm({
          title: intl.get('dataAgent.confirmCancelPublish'),
          content: intl.get('dataAgent.cancelPublishTemplateWarning'),
          centered: true,
          okButtonProps: { className: 'dip-min-width-72' },
          cancelButtonProps: { className: 'dip-min-width-72' },
          onOk() {
            const hide = message.loading(intl.get('dataAgent.cancelPublishing'), 0);
            unpublishTemplate(agent.id)
              .then(() => {
                hide();
                message.success(intl.get('dataAgent.operationSuccess'));

                if (!_.isEmpty(agentList?.list)) {
                  const newAgentList = _.cloneDeep(agentList);
                  newAgentList.list = newAgentList.list.map((item: any) =>
                    item.id === agent.id ? { ...item, status: 'unpublished', published_at: 0 } : item
                  );
                  forceUpdateAgentList(newAgentList);
                }
              })
              .catch((ex: any) => {
                hide();
                if (ex?.description) {
                  message.error(ex.description);
                }
              });
          },
          onCancel() {},
          footer: (_, { OkBtn, CancelBtn }) => (
            <div className="dip-flex-content-end">
              <OkBtn />
              <CancelBtn />
            </div>
          ),
        });
        break;

      case TemplateActionEnum.DeleteTemplate:
        // 删除模板
        // 检查是否已发布
        if (agent.status === 'published') {
          // 如果已发布，提示需要先取消发布
          modal.warning({
            title: intl.get('dataAgent.cannotDelete'),
            content: intl.get('dataAgent.cannotDeleteWithPublished', { name: agent.name }),
            okText: intl.get('dataAgent.gotIt'),
            centered: true,
          });
          return;
        }

        // 删除确认
        modal.confirm({
          title: intl.get('dataAgent.confirmDelete'),
          content: intl.get('dataAgent.confirmDeleteAgent', { name: agent.name }),
          centered: true,
          okButtonProps: { className: 'dip-min-width-72' },
          cancelButtonProps: { className: 'dip-min-width-72' },
          onOk() {
            const hide = message.loading(intl.get('dataAgent.deleting'), 0);
            deleteTemplate(agent.id)
              .then(async () => {
                hide();
                message.success(intl.get('dataAgent.operationSuccess'));

                try {
                  nextPaginationMarkerStrRef.current = '';
                  refreshCategoryList();
                } catch (ex: any) {
                  if (ex?.description) {
                    message.error(ex.description);
                  }
                  setCategoryError('获取数据时发生错误');
                }

                // 如果删除的Agent在最近访问列表中，也更新最近访问列表
                if (recentAgents.some(recent => recent.id === agent.id)) {
                  fetchRecentAgents();
                }
              })
              .catch((ex: any) => {
                hide();
                if (ex?.description) {
                  message.error(ex.description);
                }
              });
          },
          onCancel() {},
          footer: (_, { OkBtn, CancelBtn }) => (
            <div className="dip-flex-content-end">
              <OkBtn />
              <CancelBtn />
            </div>
          ),
        });
        break;

      default:
        break;
    }
  };

  // 处理发布确认
  const handlePublishSubmit = ({
    publish_to_bes,
    published_at,
  }: {
    publish_to_bes?: AgentPublishToBeEnum[];
    published_at?: number;
  }) => {
    // 清理状态
    setSelectedAgent(null);
    setPublishModalVisible(false);

    // 将agent发布为模板，无需更新当前列表
    if (publishModeRef.current === PublishModeEnum.PublishAgentAsTemplate) return;

    if (!_.isEmpty(agentList?.list)) {
      const newAgentList = _.cloneDeep(agentList);
      newAgentList.list = newAgentList.list.map((item: any) =>
        item.id === selectedAgent!.id
          ? {
              ...item,
              status: 'published',
              ...(publish_to_bes
                ? {
                    // 发布成功后，更新publish_info
                    publish_info: {
                      is_api_agent: publish_to_bes.includes(AgentPublishToBeEnum.ApiAgent) ? 1 : 0,
                      is_sdk_agent: publish_to_bes.includes(AgentPublishToBeEnum.WebSDKAgent) ? 1 : 0,
                      is_skill_agent: publish_to_bes.includes(AgentPublishToBeEnum.SkillAgent) ? 1 : 0,
                    },
                  }
                : {}),
              ...(published_at
                ? {
                    published_at,
                  }
                : {}),
            }
          : item
      );
      forceUpdateAgentList(newAgentList);
    }
  };

  // 关闭发布设置弹窗
  const handlePublishCancel = () => {
    setPublishModalVisible(false);
    setSelectedAgent(null);
  };

  // 获取数据处理状态
  const fetchProcessingStatuses = async (agents: Agent[]) => {
    if (!agents || agents.length === 0) return;

    try {
      // 最多只取当前分页大小的数量，与懒加载保持一致
      const agentsToCheck = agents.slice(0, agentListPageSize);
      const agentFlags = agentsToCheck.map(agent => ({
        agent_id: agent.id,
        agent_version: 'v0',
      }));

      const response = await getBatchDataProcessingStatus({
        agent_uniq_flags: agentFlags,
        is_show_fail_infos: true,
      });

      if (response && response.entries) {
        const newStatuses: { [key: string]: number } = {};
        response.entries.forEach(entry => {
          newStatuses[entry.agent_id] = entry.progress;
        });
        setProcessingStatuses(prev => ({ ...prev, ...newStatuses }));

        // 检查是否所有项都是100%，如果是则停止轮询
        const allCompleted = response.entries.every(entry => entry.progress === 100 || entry.progress === -1);
        if (allCompleted && processingStatusTimer.current) {
          stopPollingProcessingStatuses();
        }
      }
    } catch (error) {
      console.error('Failed to fetch processing statuses:', error);
    }
  };

  // 开始轮询处理状态
  const startPollingProcessingStatuses = () => {
    // 先清除可能存在的定时器
    if (processingStatusTimer.current) {
      clearInterval(processingStatusTimer.current);
    }

    // 立即获取一次
    if (agentList?.list.length > 0) {
      fetchProcessingStatuses(agentList?.list);
    }

    // 设置定时器，每10秒轮询一次
    processingStatusTimer.current = setInterval(() => {
      if (agentList?.list.length > 0) {
        fetchProcessingStatuses(agentList?.list);
      }
    }, 10000);
  };

  // 停止轮询
  const stopPollingProcessingStatuses = () => {
    if (processingStatusTimer.current) {
      clearInterval(processingStatusTimer.current);
      processingStatusTimer.current = null;
    }
  };

  // 点击agent
  const handleClickAgent = useCallback(
    (agent: any) => {
      switch (mode) {
        case ModeEnum.MyTemplate: {
          // 跳转到编辑模板页面
          let url = `/config?templateId=${agent.id}&mode=editTemplate`;
          const filterParams = getFilterParams();
          if (!_.isEmpty(filterParams)) {
            url += `&filterParams=${encodeURIComponent(JSON.stringify(filterParams))}`;
          }
          navigate(url);
          break;
        }

        case ModeEnum.MyAgent:
          // 跳转到编辑agent页面
          navigateToAgentEditConfig(agent);
          break;

        case ModeEnum.CustomSpace:
        case ModeEnum.DataAgent:
          // 跳转到agent使用页面
          navigateToAgentUsage(agent);
          break;

        default:
          break;
      }
    },
    [mode, navigate, navigateToAgentEditConfig, navigateToAgentUsage]
  );

  // 获取卡片处理状态标签
  const getCardProcessingStatus = useCallback(
    (agentId: string) => {
      if (mode !== ModeEnum.MyAgent) return null;
      if (processingStatuses[agentId] === undefined) return null;

      const progress = processingStatuses[agentId];
      const isProcessing = progress >= 0 && progress < 100;

      const getProcessingStatus = () => {
        if (progress === 100) {
          return intl.get('dataAgent.dataProcessComplete');
        } else if (progress === -1) {
          return intl.get('dataAgent.dataProcessFailed');
        } else if (isProcessing) {
          return intl.get('dataAgent.dataProcessing');
        }
        return null;
      };

      const status = getProcessingStatus();
      const isError = progress === -1;

      if (status) {
        return (
          <div
            className={classNames(styles.processingStatus, 'dip-ellipsis', 'dip-border-radius-2', {
              [styles['processingStatusError']]: isError,
            })}
          >
            {(isProcessing || isError) && (
              <Tooltip title={status}>
                <SyncOutlined spin={isProcessing} />
              </Tooltip>
            )}
          </div>
        );
      }

      return null;
    },
    [processingStatuses, mode]
  );

  // 获取卡片的发布文字：我的创建页面，始终显示发布状态；其它页面，只显示未发布状态
  const getCardPublishStatus = useCallback(
    (agentStatus: string) => {
      if (agentStatus === 'unpublished') {
        return <span className={styles['unpublished']}>{intl.get('dataAgent.unpublished')}</span>;
      }
      if (agentStatus === 'published_edited') {
        return <span className={styles['publishedEdited']}>{intl.get('dataAgent.publishedEdited')}</span>;
      }
      if ([ModeEnum.MyAgent, ModeEnum.MyTemplate].includes(mode)) {
        return <span className={styles['published']}>{intl.get('dataAgent.published')}</span>;
      }
    },
    [mode]
  );

  // // 在获取分类数据后开始轮询处理状态
  // useEffect(() => {
  //   // 只有我的agent，才需要查询处理状态
  //   if (mode !== ModeEnum.MyAgent) return;
  //
  //   if (agentList?.list.length > 0) {
  //     startPollingProcessingStatuses();
  //   }
  //
  //   return () => {
  //     stopPollingProcessingStatuses();
  //   };
  // }, [mode, agentList?.list]);

  // Function to render Agent cards
  const renderAgentCard = (agent: Agent, cls: string = '') => {
    const time =
      agent.status === 'published'
        ? intl.get('dataAgent.publishTime') + formatTimeSlash(agent.published_at)
        : intl.get('dataAgent.updateTime') + formatTimeSlash(agent.updated_at);

    const userInfo = getUserInfo(agent);
    const menuItems = getMenuItems({ mode, agent, perms, customSpaceInfo, currentUserId });

    const isMine = [ModeEnum.MyAgent, ModeEnum.MyTemplate].includes(mode);
    const showUserInfo = !isMine;

    return (
      <BaseCard
        checkable={isExportMode}
        checked={selectedIdsForExport.includes(agent.id)}
        checkboxDisabled={isExportMode && agent?.is_built_in === 1}
        onCheckedChange={toggleExportSelection}
        bordered={false}
        // 全部模板页面，卡片不可点击，故样式设为cursor: default
        className={classNames(mode === ModeEnum.AllTemplate ? 'dip-default' : '', cls)}
        hoverable
        isHighlighted={highlightId === agent.id}
        item={agent}
        time={time}
        name={agent.name}
        getNameSuffixIcon={getCardProcessingStatus}
        getNameSuffixStatus={getCardPublishStatus}
        profile={agent.profile}
        userAvatar={showUserInfo ? userAvatars[userInfo.user_id] : undefined}
        userName={showUserInfo ? userInfo?.username : undefined}
        menuItems={isExportMode ? [] : menuItems}
        onClickMenu={isExportMode ? undefined : handleMenuClick}
        onClick={isExportMode ? undefined : handleClickAgent}
      />
    );
  };

  // 渲染Category代理列表
  const renderCategoryAgents = () => {
    if (agentList?.list?.length === 0) {
      // 内容为空的提示语：
      // 1. 搜索时：搜索结果为空
      // 2. 选中分类，或者 筛选：暂无数据
      // 3. 我的创建：暂无创建
      // 4. 我的模板 或者 全部模板：暂无模板
      // 5. 广场：当前没有可用的 Decision Agent \n 立即新建一个，开启您的智能体验。
      const emptyText = searchName ? (
        intl.get('dataAgent.searchResultIsEmpty')
      ) : selectedCategory?.category_id ||
        publishStatusFilter !== PublishStatusEnum.All ||
        publishCategoryFilter !== AgentPublishToBeEnum.All ? (
        intl.get('dataAgent.noData')
      ) : mode === ModeEnum.MyAgent ? (
        <div>
          {intl.get('dataAgent.notYetCreated')}
          <div className="dip-text-blue-link dip-mt-6" onClick={handleCreateClick}>
            {intl.get('dataAgent.createNewAgent')}
          </div>
        </div>
      ) : [ModeEnum.MyTemplate, ModeEnum.AllTemplate].includes(mode) ? (
        intl.get('dataAgent.noTemplate')
      ) : mode === ModeEnum.DataAgent ? (
        <div>{intl.get('dataAgent.noAvailableAgentsCurrently')}</div>
      ) : (
        intl.get('dataAgent.noData')
      );

      return (
        <div className={styles.emptyStateContainer}>
          <Empty image={empty} description={emptyText} />
        </div>
      );
    }

    return (
      <>
        <Row gutter={[gap, gap]}>
          {agentList?.list?.map((agent: any) => (
            <Col key={agent.id} span={24 / countOfRow}>
              {renderAgentCard(agent)}
            </Col>
          ))}
        </Row>
        {agentListLoadingMore && (
          <div style={{ textAlign: 'center', padding: '16px 0' }}>
            <Spin size="small" />
            <span style={{ marginLeft: '8px' }}>{intl.get('dataAgent.loading')}</span>
          </div>
        )}
        {agentListNoMore && (
          <div style={{ textAlign: 'center', padding: '16px 0' }}>
            <span className="dip-text-color-45">--- 没有更多了 ---</span>
          </div>
        )}
      </>
    );
  };

  // 处理刷新按钮点击
  const handleRefresh = () => {
    nextPaginationMarkerStrRef.current = '';
    refreshCategoryList();
  };

  // 切换选中状态
  const toggleExportSelection = useCallback((_, agent: any) => {
    setSelectedIdsForExport(prev => {
      if (prev.includes(agent.id)) {
        return prev.filter(item => item !== agent.id);
      }

      if (prev.length === 500) {
        message.info(intl.get('dataAgent.maximumSelectionReached'));

        return prev;
      }

      return [...prev, agent.id];
    });
  }, []);

  // 导出agent
  const handleExportAgent = async () => {
    try {
      const { data, headers } = await exportAgent(selectedIdsForExport);
      const fileName = getFilenameFromContentDisposition(headers['content-disposition']);

      downloadFile(data, fileName, { type: headers['content-type'] });
      message.success(intl.get('dataAgent.exportSuccess'));
      setIsExportMode(false);
      setSelectedIdsForExport([]);
    } catch (ex: any) {
      if (ex.description) {
        message.error(ex.description);
      }
    }
  };

  const updateScrollPosition = (newPosition: number) => {
    if (!scrollableRef.current || !containerWidth || !contentWidth) return;

    scrollPositionRef.current = newPosition;
    scrollableRef.current.style.transform = `translateX(-${scrollPositionRef.current}px)`;

    setShowScrollArrows({
      left: scrollPositionRef.current > 0,
      right: scrollPositionRef.current < contentWidth - containerWidth,
    });
  };

  // 点击分类的左箭头，向左滑动
  const scrollLeft = () => {
    const newPosition = Math.max(scrollPositionRef.current - containerWidth, 0);
    updateScrollPosition(newPosition);
  };

  // 点击分类的右箭头，向右滑动
  const scrollRight = () => {
    const newPosition = Math.min(scrollPositionRef.current + containerWidth, contentWidth - containerWidth);
    updateScrollPosition(newPosition);
  };

  useEffect(() => {
    if (mode === ModeEnum.CustomSpace && customSpaceId) {
      // 获取当前的自定义空间名称
      getSpaceInfo(customSpaceId).then(
        info => {
          setCustomSpaceInfo(info);
        },
        (ex: any) => {
          if (ex?.description) {
            message.error(ex.description);
          }
        }
      );
    }
  }, []);

  useEffect(() => {
    const fetchPerms = async () => {
      let perms;
      try {
        // 获取权限
        perms = await getAgentManagementPerm();
        setPerms(perms);
      } catch (ex: any) {
        if (ex?.description) {
          message.error(ex.description);
        }
      }

      // 有模板发布权限，或者我的模板有数据，则显示MyCreatedTab
      if (perms?.agent_tpl.publish) {
        setShowMyCreatedTab(true);
      } else {
        try {
          // 这里仅仅是为了判断是否要显示我的模板 才调用接口获取我的模板内容
          const { entries } = await getMyTemplateList({ size: 1 });
          if (entries?.length) {
            setShowMyCreatedTab(true);
          }
        } catch {}
      }
    };

    fetchPerms();
  }, []);

  useEffect(() => {
    if (containerWidth && contentWidth && scrollableRef.current) {
      const canScroll = contentWidth > containerWidth;
      setShowScrollArrows({ left: false, right: canScroll });

      if (scrollPositionRef.current > contentWidth - containerWidth) {
        scrollPositionRef.current = contentWidth - containerWidth;
        scrollableRef.current.style.transform = `translateX(-${scrollPositionRef.current}px)`;
      }
    } else {
      setShowScrollArrows({ left: false, right: false });
    }
  }, [containerWidth, contentWidth]);

  const recentAgentList: any = useMemo(() => {
    if (countOfRow > recentAgents.length) {
      const tempArr = Array.from({ length: countOfRow - recentAgents.length }).map(() => ({ id: nanoid() }));
      return [...recentAgents, ...tempArr];
    }
    return recentAgents;
  }, [countOfRow, recentAgents]);

  const filterStatusOptions = useMemo(() => {
    const res = [
      { label: intl.get('dataAgent.all'), value: PublishStatusEnum.All },
      { label: intl.get('dataAgent.published'), value: PublishStatusEnum.Published },
      { label: intl.get('dataAgent.unpublished'), value: PublishStatusEnum.Draft },
    ];
    if (mode === ModeEnum.MyAgent) {
      res.push({
        label: intl.get('dataAgent.publishedEdited'),
        value: PublishStatusEnum.PublishedEdited,
      });
    }
    return res;
  }, [mode]);

  return (
    <GradientContainer className={classNames(styles.listPageContainer, 'dip-flex-column')}>
      <ResizeObserver
        onResize={({ width }) => {
          if (width > 0) {
            let count = computeColumnCount(width);
            if (count > 4) {
              count = 4;
            }
            if (count > 6) {
              count = 6;
            }
            setCountOfRow(count);
          }
        }}
      >
        <div className="dip-flex-item-full-height dip-flex-column">
          {/* 如果不是广场页面（没有最近访问），则 Header 固定在顶部不随页面滚动 */}
          {!showRecent && <Header mode={mode} isExportMode={isExportMode} onCreate={handleCreateClick} />}

          <div
            ref={showRecent ? pageContainerRef : null}
            className={classNames('dip-flex-column dip-overflow-auto', showRecent ? 'dip-flex-item-full-height' : '')}
          >
            {/* 只有广场页面，Header 在滚动容器内，随页面滚动 */}
            {showRecent && <Header mode={mode} isExportMode={isExportMode} onCreate={handleCreateClick} />}

            {/* 最近访问 */}
            {showRecent && Boolean(recentLoading || recentError || recentAgents?.length) && (
              <section ref={sectionRef} className={styles.recentAgents}>
                <div className={classNames(styles.sectionTitle, 'dip-mb-16 dip-pl-16 dip-pr-16')}>
                  {intl.get('dataAgent.recentVisits')}
                </div>
                {recentLoading ? (
                  <div className="dip-pl-16 dip-pr-16">
                    <SkeletonGrid countOfRow={countOfRow} avatarShape="square" />
                  </div>
                ) : recentError ? (
                  <LoadFailed
                    className={classNames(styles.emptyStateContainer, `dip-m-0 dip-p-0 dip-mr-${gap}`)}
                    onRetry={retryRecentAgents}
                  />
                ) : (
                  <div className={styles.recentAgentsContainer}>
                    {recentAgentSlideIndex > 0 && (
                      <Button
                        className={`${styles.navArrow} ${styles.leftArrow}`}
                        icon={<LeftOutlined />}
                        onClick={() => {
                          recentAgentSlideRef.current?.prev();
                        }}
                      />
                    )}
                    <Carousel
                      dots={false}
                      ref={recentAgentSlideRef}
                      infinite={false}
                      slidesToShow={countOfRow}
                      slidesToScroll={countOfRow}
                      afterChange={current => {
                        setRecentAgentSlideIndex(current);
                      }}
                    >
                      {recentAgentList.map((agent: any) => (
                        <div key={agent.id} style={{ marginRight: 16, marginLeft: 16 }}>
                          {_.isEmpty(agent.name) ? <div /> : renderAgentCard(agent, 'dip-ml-8 dip-mr-8')}
                        </div>
                      ))}
                    </Carousel>
                    {recentAgentSlideIndex < recentAgents.length - countOfRow && (
                      <Button
                        className={`${styles.navArrow} ${styles.rightArrow}`}
                        icon={<RightOutlined />}
                        onClick={() => {
                          recentAgentSlideRef.current?.next();
                        }}
                      />
                    )}
                  </div>
                )}
              </section>
            )}

            <div
              className={classNames(styles.allAgents, {
                'dip-flex-item-full-height dip-flex-column': !showRecent,
              })}
            >
              <div ref={stickySentinelRef} style={{ height: '1px', marginTop: '-1px' }} />
              <div
                className={classNames(styles.categoryContainer, {
                  [styles.stickyHeader]: showRecent,
                  [styles.isStuck]: isStuck,
                })}
              >
                <div className={classNames(styles.categoryHeader)}>
                  {[ModeEnum.MyAgent, ModeEnum.MyTemplate].includes(mode) ? (
                    <div className={classNames('dip-font-16 dip-c-black dip-font-weight-700')}>
                      {showMyCreatedTab ? (
                        <MyCreatedTab
                          activeKey={mode}
                          onChange={(mode: any) => {
                            nextPaginationMarkerStrRef.current = '';
                            // 切换tab时，同步重置筛选条件，避免 reloadDeps 触发时仍携带旧的筛选值
                            setPublishStatusFilter(PublishStatusEnum.All);

                            setPublishCategoryFilter(AgentPublishToBeEnum.All);

                            setSearchName('');
                            // 切换 我的agent、我的模板
                            setMode(mode);
                            // 切换tab时，退出导出模式
                            setIsExportMode(false);
                            setSelectedIdsForExport([]);
                          }}
                        />
                      ) : (
                        <div style={{ fontWeight: 400 }}>{intl.get('dataAgent.all')}</div>
                      )}
                    </div>
                  ) : (
                    <div className={classNames(styles.sectionTitle, 'dip-mb-16')}>
                      {showCategory ? intl.get('dataAgent.browseByCategory') : ''}
                    </div>
                  )}

                  <div className={styles.searchWrapper}>
                    {[ModeEnum.MyAgent, ModeEnum.MyTemplate].includes(mode) && (
                      <>
                        <span>
                          <span className="dip-mr-6">{intl.get('dataAgent.status')}</span>
                          <Select
                            style={{ width: 140 }}
                            options={filterStatusOptions}
                            value={publishStatusFilter}
                            onChange={value => {
                              nextPaginationMarkerStrRef.current = '';

                              setPublishStatusFilter(value);
                            }}
                          />
                        </span>
                        {/* 只有我的创建页面有 发布类型过滤项 */}
                        {mode === ModeEnum.MyAgent && (
                          <span>
                            <span className="dip-mr-6">{intl.get('dataAgent.config.type')}</span>
                            <Select
                              style={{ width: 120 }}
                              options={[
                                { label: intl.get('dataAgent.all'), value: AgentPublishToBeEnum.All },
                                { label: 'API', value: AgentPublishToBeEnum.ApiAgent },
                                { label: intl.get('dataAgent.config.skill'), value: AgentPublishToBeEnum.SkillAgent },
                              ]}
                              value={publishCategoryFilter}
                              onChange={value => {
                                nextPaginationMarkerStrRef.current = '';

                                setPublishCategoryFilter(value);
                              }}
                            />
                          </span>
                        )}
                      </>
                    )}
                    {!showCategory && (
                      <div className="dip-flex dip-gap-8 dip-position-r">
                        <SearchInput
                          value={searchName}
                          onChange={(e: any) => {
                            nextPaginationMarkerStrRef.current = '';
                            setSearchName(e.target.value);
                          }}
                          placeholder={searchPlaceholder}
                          debounce
                          style={{ width: 260 }}
                        />
                        <div>
                          {mode === ModeEnum.MyAgent && (
                            <>
                              <Tooltip title={intl.get('dataAgent.import')}>
                                <Button
                                  icon={<ImportIcon />}
                                  className={classNames(styles['icon-btn'], {
                                    [styles['icon-btn-disabled']]: isExportMode,
                                  })}
                                  disabled={isExportMode}
                                  onClick={() =>
                                    handleImportAgent(modal, () => {
                                      nextPaginationMarkerStrRef.current = '';
                                      refreshCategoryList();
                                    })
                                  }
                                />
                              </Tooltip>
                              {isExportMode ? (
                                <Button
                                  onClick={() => {
                                    setIsExportMode(false);
                                    setSelectedIdsForExport([]);
                                  }}
                                >
                                  {intl.get('dataAgent.cancelExport')}
                                </Button>
                              ) : (
                                <Tooltip title={intl.get('dataAgent.exportWithBuiltInAgentRestriction')}>
                                  <Button
                                    icon={<ExportIcon />}
                                    className={styles['icon-btn']}
                                    onClick={() => {
                                      setIsExportMode(true);
                                    }}
                                  />
                                </Tooltip>
                              )}
                            </>
                          )}
                          <Tooltip title={intl.get('dataAgent.reloadData')}>
                            <Button
                              icon={<ReloadOutlined spin={agentListInitLoading} />}
                              className={classNames(styles['icon-btn'], {
                                [styles['icon-btn-disabled']]: isExportMode,
                              })}
                              disabled={isExportMode}
                              onClick={handleRefresh}
                            />
                          </Tooltip>
                        </div>
                      </div>
                    )}
                  </div>

                  {/* 自定义空间-agent列表页面，添加agent(只有空间的创建者才可以) */}
                  {mode === ModeEnum.CustomSpace &&
                    Boolean(customSpaceId) &&
                    currentUserId === customSpaceInfo?.created_by && (
                      <SpaceAgentAddButton
                        customSpaceId={customSpaceId!}
                        onAddSuccess={() => {
                          nextPaginationMarkerStrRef.current = '';
                          refreshCategoryList();
                        }}
                      />
                    )}
                </div>
                {showCategory && (
                  <div className={classNames(styles.categoriesWrapper, `dip-w-100 dip-gap-10`)}>
                    <div
                      className={classNames(
                        styles.categories,
                        'dip-flex-item-full-width dip-1-line dip-position-r dip-overflow-hidden'
                      )}
                      ref={containerRef}
                    >
                      <div ref={scrollableRef} style={{ width: 'fit-content' }}>
                        <Button
                          type={selectedCategory.category_id === '' ? 'primary' : 'default'}
                          className={styles.categoryTag}
                          onClick={() => handleCategoryClick({ category_id: '', name: '全部' })}
                        >
                          {intl.get('dataAgent.all')}
                        </Button>
                        {categories.map((category, index) => (
                          <Button
                            key={category.category_id}
                            type={selectedCategory.category_id === category.category_id ? 'primary' : 'default'}
                            className={styles.categoryTag}
                            style={index === categories.length - 1 ? { marginRight: 0 } : {}}
                            onClick={() => handleCategoryClick(category)}
                          >
                            {category.name}
                          </Button>
                        ))}
                      </div>
                      {showScrollArrows.left && (
                        <div className={classNames(styles['arrow-icon-wrapper'], styles['left-arrow-wrapper'])}>
                          <Button
                            className={classNames(styles['arrow-icon'])}
                            icon={<LeftOutlined className="dip-font-12" />}
                            onClick={scrollLeft}
                          />
                        </div>
                      )}
                      {showScrollArrows.right && (
                        <div className={classNames(styles['arrow-icon-wrapper'], styles['right-arrow-wrapper'])}>
                          <Button
                            className={classNames(styles['arrow-icon'])}
                            icon={<RightOutlined className="dip-font-12" />}
                            onClick={scrollRight}
                          />
                        </div>
                      )}
                    </div>
                    <div className="dip-flex" style={{ gap: 8 }}>
                      <SearchInput
                        value={searchName}
                        style={{ width: 260 }}
                        onChange={(e: any) => {
                          nextPaginationMarkerStrRef.current = '';
                          setSearchName(e.target.value);
                        }}
                        placeholder={searchPlaceholder}
                        debounce
                      />
                      <Tooltip title={intl.get('dataAgent.reloadData')}>
                        <Button
                          icon={<ReloadOutlined spin={agentListInitLoading} />}
                          onClick={handleRefresh}
                          className={classNames(styles['icon-btn'])}
                        />
                      </Tooltip>
                      {mode === ModeEnum.MyAgent && (
                        <Button type="primary" onClick={handleCreateClick}>
                          <PlusIcon />
                          <span style={{ color: 'white' }}>{intl.get('dataAgent.createNew')}</span>
                        </Button>
                      )}
                    </div>
                  </div>
                )}
              </div>

              {isExportMode && (
                <div
                  className={classNames(
                    'dip-flex-space-between dip-mb-16 dip-border-radius-8',
                    styles['export-header']
                  )}
                >
                  <span>{intl.get('dataAgent.itemsSelected', { count: selectedIdsForExport.length })}</span>
                  <div>
                    <Button
                      type="link"
                      disabled={!selectedIdsForExport.length}
                      className="dip-p-0"
                      onClick={() => {
                        setSelectedIdsForExport([]);
                      }}
                    >
                      {intl.get('dataAgent.clearAll')}
                    </Button>
                    <Button
                      type="link"
                      disabled={!selectedIdsForExport.length}
                      className="dip-p-0 dip-ml-12"
                      onClick={handleExportAgent}
                    >
                      {intl.get('dataAgent.exportSelectedItems')}
                    </Button>
                  </div>
                </div>
              )}
              <div
                ref={!showRecent ? pageContainerRef : null}
                className={classNames('dip-pl-16 dip-pr-16 dip-pb-16', {
                  'dip-flex-item-full-height dip-overflow-auto': !showRecent,
                })}
              >
                {agentListInitLoading ? (
                  <SkeletonGrid countOfRow={countOfRow} avatarShape="square" />
                ) : categoryError ? (
                  <LoadFailed className={styles.emptyStateContainer} onRetry={retryCategoryAgents} />
                ) : (
                  renderCategoryAgents()
                )}
              </div>
            </div>

            {/* 发布设置弹窗 */}
            {publishModalVisible && (
              <PublishSettingsModal
                onCancel={handlePublishCancel}
                onOk={handlePublishSubmit}
                agent={selectedAgent}
                mode={publishModeRef.current}
              />
            )}

            {contextHolder}
          </div>
        </div>
      </ResizeObserver>

      {showRecent && isStuck && (
        <Tooltip title="返回顶部">
          <div className={styles.backTop}>
            <Button
              onClick={() => {
                pageContainerRef.current?.scrollTo({ top: 0, behavior: 'smooth' });
              }}
              size="large"
              shape="circle"
              icon={<VerticalAlignTopOutlined />}
            />
          </div>
        </Tooltip>
      )}
    </GradientContainer>
  );
};

export default DecisionAgent;
