import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import classNames from 'classnames';
import _, { debounce } from 'lodash';
import { message, Popover, Checkbox, Tooltip, Spin, Tag } from 'antd';
import intl from 'react-intl-universal';
import SearchInput from '@/components/SearchInput';
import AdTree from '@/components/AdTree';
import LoadingMask from '@/components/LoadingMask';
import AdTab from '@/components/AdTab';
import LoadFailed from '@/components/LoadFailed';
import UniversalModal from '@/components/UniversalModal';
import { AdTreeDataNode, adTreeUtils } from '@/utils/handle-function';
import { getToolBoxListFromMarks, getBoxToolList, getGlobalMarketToolList } from '@/apis/agent-operator-integration';
import {
  getMCPServerList,
  getMCPServerTools,
  MCPServerReleaseInfo,
  MCPTool,
} from '@/apis/agent-operator-integration/mcp';
import { useMicroWidgetProps, useBusinessDomain } from '@/hooks';
import { getModelList as getModelListReq } from '@/apis/model-manager';
import {
  AgentPublishToBeEnum,
  getAgentDetailInUsagePage,
  getAgentsByPost,
  getPublishedAgentInfoList,
} from '@/apis/agent-factory';
import { DatasourceConfigTypeEnum, LLMConfigTypeEnum } from '@/apis/agent-factory/type';
import { useLatestState } from '@/hooks';
import AgentIcon from '@/assets/icons/agent3.svg';
import ToolBoxIcon from '@/assets/images/tool.svg';
import ToolIcon from '@/assets/icons/tool.svg';
import MCPIcon from '@/assets/icons/mcp.svg';
import NoResultIcon from '@/assets/icons/no-result.svg';
import {
  getInputParamsFromOpenAPISpec,
  hiddenBuildInFields,
  transformAgentInput,
  getMCPInputParamsFromOpenAPISpec,
} from '../utils';
import './style.less';

const getBusinessDomainName = (businessDomainMap: Record<string, any>, businessDomainId: string) => {
  const currentDomainName = intl.get('businessDomain.currentDomain');
  return businessDomainMap?.[businessDomainId]?.isCurrent
    ? currentDomainName
    : businessDomainMap?.[businessDomainId]?.name || currentDomainName;
};

const ToolModal = ({ agentKey, onCancel, value, onConfirm, retrieverBlockOptions, allPreviousBlockVars }: any) => {
  const microWidgetProps = useMicroWidgetProps();
  const { publicAndCurrentDomainIds, businessDomainMap } = useBusinessDomain();
  const [treeProps, setTreeProps, getTreeProps] = useLatestState({
    treeData: [] as AdTreeDataNode[],
    checkedKeys: [] as any,
    checkedNodes: [] as AdTreeDataNode[],
    loadedKeys: [] as any,
    expandedKeys: [] as any,
    // 搜索用到的属性
    searchText: '' as string,
  });
  const [searchTreeProps, setSearchTreeProps, getSearchTreeProps, resetSearchTreeProps] = useLatestState({
    treeData: [] as AdTreeDataNode[],
    expandedKeys: [] as any,
    checkedKeys: [] as any,
    checkedNodes: [] as AdTreeDataNode[],
    loadedKeys: [] as any,
  });
  const publicAndCurrentDomainIdsRef = useRef<string[]>([]);
  const cacheLLMList = useRef<any[]>([]);
  const nextPaginationMarkerStrRef = useRef<string>(''); // agent分页marker
  const isAgentLoadMoreRef = useRef<boolean>(false); // agent是否正在加载更多
  const [loading, setLoading] = useState(true);
  const [agentTool, setAgentTool] = useState({
    data: [],
    checkedKeys: [] as any,
    searchText: '' as string,
  });
  const agentBoxId = 'built-in-agent';
  // 缓存工具箱信息，用于后续获取工具时使用
  const [toolBoxCache, setToolBoxCache] = useState<Record<string, any>>({});
  // agent的输入参数集合，诸如: { id1: { loading: true }, id2: { loading: false, details: { ...}}}
  const [agentInputs, setAgentInputs] = useState<Record<string, { loading: boolean; details?: any[] }>>({});

  // 新增MCP相关状态
  const [mcpProps, setMcpProps, getMcpProps] = useLatestState({
    treeData: [] as AdTreeDataNode[],
    checkedKeys: [] as any,
    checkedNodes: [] as AdTreeDataNode[],
    loadedKeys: [] as any,
    expandedKeys: [] as any,
    searchText: '' as string,
  });
  const [searchMcpProps, setSearchMcpProps, getSearchMcpProps, resetSearchMcpProps] = useLatestState({
    treeData: [] as AdTreeDataNode[],
    expandedKeys: [] as any,
    checkedKeys: [] as any,
    checkedNodes: [] as AdTreeDataNode[],
    loadedKeys: [] as any,
  });

  useEffect(() => {
    getModelList();
  }, []);

  useEffect(() => {
    if (!publicAndCurrentDomainIds) return;

    publicAndCurrentDomainIdsRef.current = publicAndCurrentDomainIds;
    getAgentTool();
    getTooBox();
    getMcpServerList();
  }, [publicAndCurrentDomainIds]);

  // 获取agent的input参数
  const fetchAgentInput = async (agentId: string) => {
    try {
      const result = await getAgentDetailInUsagePage({ id: agentId, version: 'latest', is_visit: false });
      const inputs = result?.config?.input?.fields
        .filter(input => !hiddenBuildInFields.includes(input?.name))
        .map(input => ({
          input_name: input?.name,
          input_desc: input?.desc || '',
          input_type: input?.type || 'unknown',
          enable: input?.name === 'query',
          map_type: 'auto',
          required: input?.name === 'query',
        }));
      setAgentInputs(prev => ({
        ...prev,
        [agentId]: {
          loading: false,
          details: inputs,
        },
      }));
    } catch (ex: any) {
      if (ex?.description) {
        message.error(ex.description);
      }

      setAgentInputs(prev => ({
        ...prev,
        [agentId]: {
          loading: false,
          details: [],
        },
      }));
    }
  };

  const getAgentTool = useCallback(
    async (searchKey: string = '') => {
      try {
        const isFirst = !nextPaginationMarkerStrRef.current;
        const postData: any = {
          pagination_marker_str: nextPaginationMarkerStrRef.current,
          size: 10,
          publish_to_be: AgentPublishToBeEnum.SkillAgent,
          name: searchKey,
          business_domain_ids: publicAndCurrentDomainIdsRef.current,
        };
        if (agentKey) postData.exclude_agent_keys = [agentKey];

        const { entries, pagination_marker_str } = await getAgentsByPost(postData);
        nextPaginationMarkerStrRef.current = pagination_marker_str;

        setAgentTool((prevState: any) => {
          const checkedKeys =
            value?.filter((item: any) => item.tool_type === 'agent')?.map((item: any) => item.tool_id) || [];
          return {
            ...prevState,
            data: isFirst ? entries : [...prevState.data, ...entries],
            checkedKeys: [...prevState.checkedKeys, ...checkedKeys],
          };
        });
      } catch (error: any) {
        const { Description, ErrorDetails } = error?.response || error?.data || error || {};
        (ErrorDetails || Description) && message.error(ErrorDetails || Description);
        return false;
      } finally {
        isAgentLoadMoreRef.current = false;
      }
    },
    [value]
  );

  const debounceGetAgentTool = useMemo(() => debounce(getAgentTool, 300), [getAgentTool]);

  const getModelList = async () => {
    const { data } = await getModelListReq({ page: 1, size: 100 });
    cacheLLMList.current = data || [];
  };

  const getTooBox = async () => {
    try {
      const response = await getToolBoxListFromMarks(
        {
          page: 1,
          all: true,
          status: 'published',
        },
        publicAndCurrentDomainIdsRef.current
      );

      if (response && response.data) {
        // 将新接口返回的数据转换为组件期望的格式
        const toolBoxData = response.data.map((item: any) => ({
          ...item,
          type: 'tool-box',
          // 兼容旧字段名
          box_id: item.box_id,
          box_name: item.box_name,
          box_desc: item.box_desc,
          box_svc_url: item.box_svc_url,
          create_time: item.create_time,
          update_time: item.update_time,
          create_user: item.create_user,
          update_user: item.update_user,
        }));

        // 缓存工具箱信息
        const cache: Record<string, any> = {};
        toolBoxData.forEach((item: any) => {
          cache[item.box_id] = item;
        });
        setToolBoxCache(cache);

        let treeData = adTreeUtils.createAdTreeNodeData(toolBoxData, {
          titleField: 'box_name',
          keyField: 'box_id',
          isLeaf: false,
        });

        if (value) {
          const checkedKeys = value
            .filter((item: any) => item.tool_box_id !== agentBoxId)
            .map((item: any) => item.tool_id);
          const expandedKeys = _.uniq(value.map((item: any) => item.tool_box_id));
          const toolBoxNodes = treeData.filter(item => expandedKeys.includes(item.key));

          for (let i = 0; i < toolBoxNodes.length; i++) {
            const toolBoxNode = toolBoxNodes[i];
            const childTreeData = await getToolTreeNode(toolBoxNode);
            if (childTreeData) {
              treeData = adTreeUtils.addTreeNode(treeData, childTreeData as any);
            }
          }

          const checkedNodes = adTreeUtils.getTreeNodeByKey(treeData, checkedKeys);

          setTreeProps(prevState => ({
            ...prevState,
            checkedKeys,
            expandedKeys,
            loadedKeys: expandedKeys,
            treeData,
            checkedNodes,
          }));
          setLoading(false);
          return;
        }
        setLoading(false);
        setTreeProps(prevState => ({
          ...prevState,
          treeData,
        }));
      }
    } catch (error: any) {
      setLoading(false);
      const { Description, ErrorDetails } = error?.response || error?.data || error || {};
      (ErrorDetails || Description) && message.error(ErrorDetails || Description);
    }
  };

  const selectedTools = useMemo(() => {
    const flatTreeData = adTreeUtils.flatTreeData(treeProps.treeData);
    const toolBox = treeProps.checkedNodes.filter(item => item.sourceData?.type === 'tool-box');
    let tools = treeProps.checkedNodes.filter(item => item.sourceData?.type === 'tool');
    if (toolBox.length > 0) {
      const toolBoxKeys = toolBox.map(item => item.key);
      flatTreeData.forEach(item => {
        if (toolBoxKeys.includes(item.parentKey!) && !treeProps.searchText) {
          tools = [...tools, item];
        }
      });
    }
    const agentTools = agentTool.data.filter((tool: any) => agentTool.checkedKeys.includes(tool.key));
    const tempArr: any = agentTools.map((item: any) => ({
      key: item.key,
      title: item.name,
      sourceData: {
        tool_desc: item.profile,
        box_name: 'Agent',
        box_id: agentBoxId,
        is_build_in: false,
        tool_input: item.tool_input || [],
        agent_version: item.version,
      },
    }));

    // 添加MCP工具到选中列表
    const flatMcpTreeData = adTreeUtils.flatTreeData(mcpProps.treeData);
    const mcpServerBox = mcpProps.checkedNodes.filter(item => item.sourceData?.type === 'mcp-server');
    let mcpTools = mcpProps.checkedNodes.filter(item => item.sourceData?.type === 'mcp-tool');
    if (mcpServerBox.length > 0) {
      const mcpServerKeys = mcpServerBox.map(item => item.key);
      flatMcpTreeData.forEach(item => {
        if (mcpServerKeys.includes(item.parentKey!) && !mcpProps.searchText) {
          mcpTools = [...mcpTools, item];
        }
      });
    }

    tools = [...tools, ...tempArr, ...mcpTools];
    return _.uniqBy(tools, 'key');
  }, [treeProps.treeData, treeProps.checkedNodes, agentTool.checkedKeys, mcpProps.treeData, mcpProps.checkedNodes]);

  // 批量获取skillAgent的input
  const multiFetchSkillAgentInput = async (agents: any[]) => {
    try {
      const agent_keys = agents.map(({ agent_key }) => agent_key);
      const { entries } = await getPublishedAgentInfoList(agent_keys);

      return agents.map(agent => {
        const input = entries
          .find(item => item.key === agent.agent_key)
          ?.config?.input?.fields?.filter(({ name }) => !hiddenBuildInFields.includes(name));

        return {
          ...agent,
          tool_input: input?.map(transformAgentInput),
        };
      });
    } catch {
      return agents;
    }
  };

  const onOk = async () => {
    if (selectedTools.length === 0) {
      message.error(intl.get('agentCommonConfig.llm.selectAtLeastOneTool'));
      return;
    }

    let defaultModel = {};
    if (cacheLLMList.current.length > 0) {
      let model = cacheLLMList.current.find(item => item.model === 'Qwen2-72B-Chat');
      if (!model) {
        model = cacheLLMList.current[0];
      }
      const model_para = model?.model_para;
      defaultModel = {
        id: model.model_id,
        name: model.model_name,
        temperature: model_para?.temperature[2],
        top_p: model_para?.top_p[2],
        max_tokens: model_para?.max_tokens[2],
        top_k: model_para?.top_k[2],
        presence_penalty: model_para?.presence_penalty[2],
        frequency_penalty: model_para?.frequency_penalty[2],
      };
    }

    const newValue = selectedTools.map(item => {
      const target = value.find((v: any) => v.tool_id === item.key);
      //  MCP工具标记为'mcp'类型
      const tool_type = item.sourceData?.type === 'mcp-tool' ? 'mcp' : item.sourceData?.type || 'agent';
      const tempTool: any = {
        tool_type,
        tool_id: item.key,
        tool_name: item.title,
        tool_box_id: item.sourceData?.box_id,
        tool_box_name: item.sourceData?.box_name,
        tool_desc: item.sourceData?.tool_desc,
        intervention: target?.intervention ?? false,
        tool_input: item.sourceData?.tool_input,
        ...(tool_type === 'agent'
          ? {
              agent_version: item.sourceData.agent_version,
              agent_key: item.key,
              data_source_config: item.sourceData.data_source_config || {
                type: DatasourceConfigTypeEnum.SelfConfigured,
              },
              llm_config: item.sourceData.llm_config || {
                type: LLMConfigTypeEnum.SelfConfigured,
              },
            }
          : {}),
      };
      if (item.sourceData?.is_build_in && item.sourceData?.box_id === 'built-in') {
        // 内置工具 中的搜索工具下面的所有工具 要有config  添加默认参数  (搜索工具 箱的id是built-in， 和后端沟通暂时先用固定的id判断)
        tempTool.config = target?.config ?? {
          llm_config: defaultModel,
          schema_linking_res: {
            name: retrieverBlockOptions?.[0]?.value,
          },
          query: {
            name: allPreviousBlockVars?.[0]?.name,
          },
        };
      }
      const toolInput = item.sourceData?.tool_input;
      const getToolInput = (toolInput: any) => {
        return toolInput.map((item: any) => {
          if (item.children) {
            return {
              input_name: item.input_name,
              input_type: item.input_type,
              children: getToolInput(item.children),
            };
          }
          return {
            input_name: item.input_name,
            input_type: item.input_type,
            map_type: 'auto',
            map_value: '',
            enable: item.required ?? false,
          };
        });
      };
      tempTool.tool_input = target?.tool_input ?? getToolInput(toolInput);
      return tempTool;
    });

    // 从newValue中过滤出无tool_input的技能agent,然后调用接口获取它们的input
    const skillAgentsWithoutInput = newValue.filter(item => item.tool_type === 'agent' && !item?.tool_input?.length);
    let skillAgentsWithInput = [];

    if (skillAgentsWithoutInput?.length) {
      skillAgentsWithInput = await multiFetchSkillAgentInput(skillAgentsWithoutInput);
    }

    onConfirm(
      newValue.map(item => {
        const findAgent = skillAgentsWithInput.find(agent => agent.tool_id === item.tool_id);

        if (findAgent) return findAgent;

        return item;
      })
    );
  };

  const renderParam = (param: any) => {
    return (
      <div style={{ width: 428, padding: '16px', maxHeight: '400px', overflow: 'auto' }}>
        {param?.map((item: any, index: number) => (
          <div
            key={item.input_name}
            className={classNames('dip-font-12', {
              'dip-mb-16': index !== param.length - 1,
            })}
          >
            <div>
              <span className="dip-c-bold">{item.input_name}</span>
              <span className="dip-ml-8 dip-c-text-lower">{item.input_type}</span>
              {item.required && (
                <span className="dip-ml-8" style={{ color: '#ff7a45' }}>
                  {intl.get('agentCommonConfig.llm.required')}
                </span>
              )}
            </div>
            {item.input_desc && (
              <div className="dip-ellipsis-2 dip-mt-8 dip-c-text-lower" title={item.input_desc}>
                {item.input_desc}
              </div>
            )}
          </div>
        ))}
      </div>
    );
  };

  const titleRender = (nodeData: AdTreeDataNode) => {
    const desc = nodeData.sourceData?.box_desc || nodeData.sourceData?.tool_desc;

    // 根据节点类型确定图标
    let IconComponent = null;
    if (nodeData.sourceData?.type === 'tool-box') {
      IconComponent = (
        <div className="dip-position-r" style={{ width: 38, height: 38 }}>
          <ToolBoxIcon style={{ width: 38, height: 38, borderRadius: 8 }} />
          <div className="toolBoxLabel">
            {nodeData.sourceData?.metadata_type === 'openapi' ? 'OpenAPI' : '函数计算'}
          </div>
        </div>
      );
    } else if (nodeData.sourceData?.type === 'tool') {
      IconComponent = <ToolIcon style={{ width: 32, height: 32 }} />;
    } else if (nodeData.sourceData?.type === 'mcp-server') {
      IconComponent = <MCPIcon style={{ width: 38, height: 38, minWidth: '32px' }} />;
    }

    const businessDomainName =
      IconComponent && nodeData.sourceData?.business_domain_id
        ? getBusinessDomainName(businessDomainMap as Record<string, any>, nodeData.sourceData?.business_domain_id)
        : '';

    return (
      <span className="dip-flex-align-center dip-mt-8">
        {IconComponent}
        <span className="dip-flex-column dip-ml-8 dip-overflow-hidden dip-flex-1">
          <span className="dip-ellipsis" title={nodeData.title}>
            {nodeData.title}
          </span>
          <span style={{ fontSize: 12 }} className="dip-c-text-lower dip-ellipsis" title={desc}>
            {desc || <span className="dip-c-subtext">{intl.get('global.notDes')}</span>}
          </span>
          {(nodeData.sourceData?.type === 'tool' || nodeData.sourceData?.type === 'mcp-tool') && (
            <div className="dip-flex dip-mt-8" style={{ gap: 8 }}>
              {nodeData.sourceData?.tool_input?.slice(0, 3).map((inputItem: any) => (
                <span
                  style={{ background: 'rgba(0, 0, 0, 0.04)' }}
                  title={inputItem.input_name}
                  className="dip-pl-8 dip-pr-8 dip-font-12"
                  key={inputItem.input_name}
                >
                  {inputItem.input_name}
                </span>
              ))}
              {nodeData.sourceData?.tool_input?.length > 0 && (
                <Popover
                  overlayClassName="ToolModal-param-tip"
                  content={renderParam(nodeData.sourceData?.tool_input)}
                  trigger={['hover']}
                  destroyOnHidden
                  placement="bottomLeft"
                  getPopupContainer={() => microWidgetProps?.container}
                >
                  <span className="dip-c-link dip-font-12">{intl.get('agentCommonConfig.llm.parameter')}</span>
                </Popover>
              )}
            </div>
          )}
        </span>
        {businessDomainName && (
          <Tag color="blue" bordered={false} className="dip-ml-4">
            {businessDomainName}
          </Tag>
        )}
      </span>
    );
  };

  const getToolTreeNode = async (nodeData: AdTreeDataNode) => {
    try {
      const boxId = nodeData.key as string;
      const response = await getBoxToolList(
        boxId,
        {
          page: 1,
          page_size: 100,
          status: 'enabled', // 只获取启用的工具
          all: true, // 获取所有工具
        },
        publicAndCurrentDomainIdsRef.current
      );

      if (response && response.tools) {
        // 从缓存中获取工具箱信息
        const toolBoxInfo = toolBoxCache[boxId] || nodeData.sourceData;
        const global_headers = toolBoxInfo.global_headers || {};

        const headers = Object.keys(global_headers).map(headerItem => ({
          input_name: headerItem,
          input_type: 'string',
        }));

        // 将新接口返回的工具数据转换为组件期望的格式
        const toolData = response.tools.map((item: any) => {
          const allInputs = getInputParamsFromOpenAPISpec(item.metadata?.api_spec);

          return {
            // 新接口字段映射到旧格式
            tool_id: item.tool_id,
            tool_name: item.name,
            tool_desc: item.description,
            tool_path: item.metadata?.path || '',
            tool_method: item.metadata?.method || 'GET',
            tool_input: allInputs,
            // 兼容字段
            type: 'tool',
            box_name: nodeData.title,
            box_id: nodeData.key as string,
            create_time: item.create_time,
            update_time: item.update_time,
            is_build_in: false,
          };
        });

        // 合并headers到tool_input
        const finalToolData = toolData.map((tool: any) => ({
          ...tool,
          tool_input: _.uniqBy([...tool.tool_input, ...headers], 'input_name'),
        }));

        const childTreeData = adTreeUtils.createAdTreeNodeData(finalToolData, {
          titleField: 'tool_name',
          keyField: 'tool_id',
          parentKey: nodeData.key as string,
          keyPath: nodeData.keyPath,
        });
        return childTreeData;
      }
    } catch (error: any) {
      const { Description, ErrorDetails } = error?.response || error?.data || error || {};
      (ErrorDetails || Description) && message.error(ErrorDetails || Description);
      return false;
    }
  };

  const loadData = (nodeData: AdTreeDataNode) =>
    new Promise(resolve => {
      const getTreeData = async () => {
        const childTreeData = await getToolTreeNode(nodeData);

        if (childTreeData) {
          const treeData = adTreeUtils.addTreeNode(treeProps.treeData, childTreeData as any);
          setTreeProps(prevState => ({
            ...prevState,
            treeData,
          }));
          resolve(true);
        } else {
          resolve(false);
        }
      };
      getTreeData();
    });

  const handleSearch = async (e: any) => {
    const value = e.target.value;
    if (value) {
      try {
        // 使用新接口进行搜索
        const response = await getGlobalMarketToolList(
          {
            tool_name: value, // 按工具箱名称搜索
            all: true,
            sort_by: 'create_time',
            sort_order: 'desc',
            status: 'enabled', // 只获取启用的工具
          },
          publicAndCurrentDomainIdsRef.current
        );

        if (response && response.data) {
          const toolBoxData = response.data.map((item: any) => ({
            ...item,
            type: 'tool-box',
            // 保持字段兼容性
            box_id: item.box_id,
            box_name: item.box_name,
            box_desc: item.box_desc,
            box_svc_url: item.box_svc_url,
            create_time: item.create_time,
            update_time: item.update_time,
            create_user: item.create_user,
            update_user: item.update_user,
          }));

          const treeData = adTreeUtils.createAdTreeNodeData(toolBoxData, {
            titleField: 'box_name',
            keyField: 'box_id',
            isLeaf: false,
          });

          const boxIds = response.data.map((item: any) => item.box_id);

          // 为每个工具箱加载工具
          for (const nodeData of treeData) {
            const toolBoxInfo = nodeData?.sourceData;
            const toolData = toolBoxInfo?.tools.map((item: any) => {
              const allInputs = getInputParamsFromOpenAPISpec(item.metadata?.api_spec);
              return {
                // 新接口字段映射到旧格式
                tool_id: item.tool_id,
                tool_name: item.name,
                tool_desc: item.description,
                tool_path: item.metadata?.path || '',
                tool_method: item.metadata?.method || 'GET',
                tool_input: allInputs,
                // 兼容字段
                type: 'tool',
                box_name: nodeData.title,
                box_id: nodeData.key as string,
                create_time: item.create_time,
                update_time: item.update_time,
                is_build_in: false,
              };
            });

            const global_headers = toolBoxInfo.global_headers || {};
            const headers = Object.keys(global_headers).map(headerItem => ({
              input_name: headerItem,
              input_type: 'string',
            }));
            // 合并headers到tool_input
            const finalToolData = toolData.map((tool: any) => ({
              ...tool,
              tool_input: _.uniqBy([...tool.tool_input, ...headers], 'input_name'),
            }));
            const childTreeData = adTreeUtils.createAdTreeNodeData(finalToolData, {
              titleField: 'tool_name',
              keyField: 'tool_id',
              parentKey: nodeData.key as string,
              keyPath: nodeData.keyPath,
            });
            nodeData.children = childTreeData;
          }

          // 获取之前选中的节点
          const previousCheckedKeys = getTreeProps().checkedKeys;
          const previousCheckedNodes = getTreeProps().checkedNodes;

          setTreeProps(prevState => ({
            ...prevState,
            searchText: value,
          }));
          setSearchTreeProps(prevState => ({
            ...prevState,
            treeData: treeData,
            expandedKeys: boxIds,
            loadedKeys: boxIds,
            // 保留所有之前选中的节点状态
            checkedKeys: previousCheckedKeys,
            checkedNodes: previousCheckedNodes,
          }));
        }
      } catch (error) {
        console.error('搜索失败:', error);
        // 搜索失败时显示空结果
        setTreeProps(prevState => ({
          ...prevState,
          searchText: value,
        }));
        setSearchTreeProps(prevState => ({
          ...prevState,
          treeData: [],
          expandedKeys: [],
          loadedKeys: [],
          checkedKeys: getTreeProps().checkedKeys,
          checkedNodes: getTreeProps().checkedNodes,
        }));
      }
    } else {
      setTreeProps(prevState => ({
        ...prevState,
        searchText: '',
      }));
      resetSearchTreeProps();
    }
  };

  const handleSearchAgent = (e: any) => {
    const value = e.target.value;
    nextPaginationMarkerStrRef.current = '';
    setAgentTool(prevState => ({
      ...prevState,
      searchText: value,
    }));

    nextPaginationMarkerStrRef.current = '';
    debounceGetAgentTool(value);
  };

  const checkboxChange = (tool_id: string) => {
    if (agentTool.checkedKeys.includes(tool_id)) {
      setAgentTool(prevState => ({
        ...prevState,
        checkedKeys: prevState.checkedKeys.filter((ii: any) => ii !== tool_id),
      }));
    } else {
      setAgentTool(prevState => ({
        ...prevState,
        checkedKeys: [...prevState.checkedKeys, tool_id],
      }));
    }
  };

  const handleScroll = (e: any) => {
    // 检查是否滚动到距离底部100px以内
    if (
      e.target.scrollHeight - (e.target.scrollTop + e.target.clientHeight) < 100 &&
      nextPaginationMarkerStrRef.current &&
      !isAgentLoadMoreRef.current
    ) {
      // 触发懒加载
      isAgentLoadMoreRef.current = true;
      debounceGetAgentTool(agentTool.searchText);
    }
  };

  const renderToolTree = () => {
    return (
      <div className="dip-h-100 dip-flex-column">
        <SearchInput
          style={{ width: '100%' }}
          placeholder={intl.get('agentTool.searchPlaceholder')}
          onChange={handleSearch}
          debounce
        />
        <div className="dip-flex-item-full-height" style={{ overflowY: 'auto' }}>
          {treeProps.searchText && searchTreeProps.treeData.length == 0 ? (
            <div className="dip-column-center">
              <NoResultIcon />
              <div className="dip-mt-8 dip-c-text-lower">{intl.get('global.noResult')}</div>
            </div>
          ) : (
            <AdTree
              className="ToolModal-tree"
              expandAction={false}
              selectable={false}
              treeData={treeProps.searchText ? searchTreeProps.treeData : treeProps.treeData}
              checkable
              titleRender={titleRender as any}
              checkedKeys={treeProps.checkedKeys}
              onCheck={(checkedKeys, { node, checkedNodes }) => {
                if (!treeProps.searchText && (node as AdTreeDataNode).sourceData?.type === 'tool-box') {
                  const loadedKeys = getTreeProps().loadedKeys;
                  if (!loadedKeys.includes(node.key)) {
                    // 树节点的展开会自动触发loadData方法dd
                    setTreeProps(prevState => ({
                      ...prevState,
                      expandedKeys: [...prevState.expandedKeys, node.key],
                    }));
                  }
                }

                // 获取之前选中的节点
                const previousCheckedNodes = getTreeProps().checkedNodes;

                // 在搜索状态下，合并之前选中的节点
                if (treeProps.searchText) {
                  // 获取搜索结果树数据
                  const searchTreeData = getSearchTreeProps().treeData;
                  // 找出不在搜索结果中的之前选中的节点
                  const previousNodesNotInSearch = previousCheckedNodes.filter(
                    prevNode =>
                      !searchTreeData.some(
                        (searchNode: AdTreeDataNode) =>
                          searchNode.key === prevNode.key ||
                          searchNode.children?.some((child: AdTreeDataNode) => child.key === prevNode.key)
                      )
                  );

                  // 合并当前选中的节点和之前选中的节点
                  let mergedCheckedKeys = Array.isArray(checkedKeys)
                    ? [...checkedKeys, ...previousNodesNotInSearch.map(node => node.key)]
                    : checkedKeys;
                  let mergedCheckedNodes: any = [...(checkedNodes as AdTreeDataNode[]), ...previousNodesNotInSearch];

                  mergedCheckedNodes = _.filter(mergedCheckedNodes, (item: any) => {
                    if (!item?.isLeaf) mergedCheckedKeys = _.filter(mergedCheckedKeys, key => key !== item.value);
                    return item?.isLeaf;
                  });

                  setTreeProps(prevState => {
                    const data = {
                      ...prevState,
                      checkedKeys: mergedCheckedKeys,
                      checkedNodes: mergedCheckedNodes,
                    };
                    return data;
                  });
                  setSearchTreeProps(prevState => {
                    const data = {
                      ...prevState,
                      checkedKeys: mergedCheckedKeys,
                      checkedNodes: mergedCheckedNodes,
                    };
                    return data;
                  });
                } else {
                  setTreeProps(prevState => ({
                    ...prevState,
                    checkedKeys,
                    checkedNodes: checkedNodes as AdTreeDataNode[],
                  }));
                }
              }}
              loadData={loadData as any}
              loadedKeys={treeProps.searchText ? searchTreeProps.loadedKeys : treeProps.loadedKeys}
              onLoad={loadedKeys => {
                setTreeProps(prevState => ({ ...prevState, loadedKeys }));
              }}
              expandedKeys={treeProps.searchText ? searchTreeProps.expandedKeys : treeProps.expandedKeys}
              onExpand={expandedKeys => {
                if (treeProps.searchText) {
                  setSearchTreeProps(prevState => ({ ...prevState, expandedKeys }));
                } else {
                  setTreeProps(prevState => ({ ...prevState, expandedKeys }));
                }
              }}
            />
          )}
        </div>
      </div>
    );
  };

  const renderAgentItem = (item: any) => {
    const { loading, details: inputList } = agentInputs[item?.id] || {};
    const businessDomainName = item.business_domain_id
      ? getBusinessDomainName(businessDomainMap as Record<string, any>, item.business_domain_id)
      : ''; // 所属业务域名称

    return (
      <div className="dip-flex-align-center ">
        <Checkbox
          // className="dip-flex-align-center"
          disabled={item.intervention}
          checked={agentTool.checkedKeys.includes(item.key)}
          onChange={e => {
            if (e.target.checked) {
              setAgentTool(prevState => ({
                ...prevState,
                checkedKeys: [...prevState.checkedKeys, item.key],
              }));
            } else {
              setAgentTool(prevState => ({
                ...prevState,
                checkedKeys: prevState.checkedKeys.filter((ii: any) => ii !== item.key),
              }));
            }
          }}
        />
        <div
          className="dip-ml-8 dip-flex-align-center dip-flex-1 dip-overflow-hidden dip-gap-4"
          style={item.intervention ? { cursor: 'not-allowed' } : {}}
          onClick={() => {
            if (!item.intervention) {
              checkboxChange(item.id);
            }
          }}
        >
          {/* <img style={{ width: 32, height: 32 }} src={agentIcon} alt="" /> */}
          <AgentIcon style={{ width: '32px', height: '32px' }} className="dip-flex-shrink-0" />
          <div className="dip-ml-8 dip-flex-column dip-flex-1 dip-overflow-hidden">
            <div className="dip-ellipsis" title={item.name} style={{ width: 'fit-content', maxWidth: '100%' }}>
              {item.name}
            </div>
            <div
              className="dip-font-12 dip-c-text-lower dip-ellipsis dip-w-fit-content dip-m-w-100"
              title={item.profile}
            >
              {item.profile || <span className="dip-c-subtext">{intl.get('global.notDes')}</span>}
            </div>
            <div className="dip-flex dip-mt-8" style={{ gap: 8 }}>
              {item?.tool_input?.slice(0, 3).map((inputItem: any) => (
                <span
                  style={{ background: 'rgba(0, 0, 0, 0.04)' }}
                  title={inputItem.input_name}
                  className="dip-pl-8 dip-pr-8 dip-font-12"
                  key={inputItem.input_name}
                >
                  {inputItem.input_name}
                </span>
              ))}
              <Popover
                overlayClassName="ToolModal-param-tip"
                content={
                  inputList?.length ? (
                    renderParam(inputList)
                  ) : loading ? (
                    <Spin className="ToolModal-param-tip-param-loading" />
                  ) : (
                    <LoadFailed className="ToolModal-param-tip-param-failed" />
                  )
                }
                destroyOnHidden
                placement="bottomLeft"
                getPopupContainer={() => microWidgetProps?.container}
                onOpenChange={open => {
                  if (open && item?.id) {
                    // 如果之前没请求过，则往agentInputs里插入一条，并且fetch请求
                    if (!agentInputs[item.id]) {
                      setAgentInputs(prev => ({
                        ...prev,
                        [item.id]: {
                          loading: true,
                        },
                      }));

                      fetchAgentInput(item?.id);
                    }
                  }
                }}
              >
                <span className="dip-c-link dip-font-12 dip-pointer">
                  {intl.get('agentCommonConfig.llm.parameter')}
                </span>
              </Popover>
            </div>
          </div>
          {businessDomainName && (
            <Tag color="blue" bordered={false} style={{ marginTop: '-20px' }}>
              {businessDomainName}
            </Tag>
          )}
        </div>
      </div>
    );
  };

  const renderAgentList = () => {
    return (
      <div className="dip-h-100 dip-flex-column">
        <SearchInput
          style={{ width: '100%' }}
          placeholder={intl.get('dataAgent.searchSkillName')}
          onChange={handleSearchAgent}
          debounce
          className="dip-mb-8"
        />
        <div className="dip-flex-item-full-height" style={{ overflowY: 'auto' }} onScroll={handleScroll}>
          {agentTool.searchText && agentTool.data.length === 0 ? (
            <div className="dip-column-center">
              <NoResultIcon />
              <div className="dip-mt-8 dip-c-text-lower">{intl.get('global.noResult')}</div>
            </div>
          ) : (
            agentTool.data.map((item: any) => (
              <div
                key={item.tool_id}
                className="dip-mb-8 dip-mb-12"
                style={{ background: item.intervention ? '#f5f5f5' : '#fff' }}
              >
                {item.intervention ? (
                  <Tooltip title={intl.get('dataAgent.config.agentWithHumanIntervention')}>
                    {renderAgentItem(item)}
                  </Tooltip>
                ) : (
                  renderAgentItem(item)
                )}
              </div>
            ))
          )}
        </div>
      </div>
    );
  };

  const getMcpServerList = async () => {
    try {
      const response = await getMCPServerList(
        {
          page: 1,
          page_size: 100,
          status: 'published', // 只获取已发布的MCP服务
        },
        publicAndCurrentDomainIdsRef.current
      );

      if (response && response.data) {
        // 将MCP服务数据转换为组件期望的格式
        const mcpServerData = response.data.map((item: MCPServerReleaseInfo) => ({
          ...item,
          type: 'mcp-server',
          // 兼容树形结构字段
          box_id: item.mcp_id,
          box_name: item.name,
          box_desc: item.description,
          create_time: item.create_time,
        }));

        let treeData = adTreeUtils.createAdTreeNodeData(mcpServerData, {
          titleField: 'box_name',
          keyField: 'box_id',
          isLeaf: false,
        });

        if (value) {
          const checkedKeys = value.filter((item: any) => item.tool_type === 'mcp').map((item: any) => item.tool_id);
          const expandedKeys = _.uniq(
            value.filter((item: any) => item.tool_type === 'mcp').map((item: any) => item.tool_box_id)
          );
          const mcpServerNodes = treeData.filter(item => expandedKeys.includes(item.key));

          for (let i = 0; i < mcpServerNodes.length; i++) {
            const mcpServerNode = mcpServerNodes[i];
            const childTreeData = await getMcpToolTreeNode(mcpServerNode);
            if (childTreeData) {
              treeData = adTreeUtils.addTreeNode(treeData, childTreeData as any);
            }
          }

          const checkedNodes = adTreeUtils.getTreeNodeByKey(treeData, checkedKeys);

          setMcpProps(prevState => ({
            ...prevState,
            checkedKeys,
            expandedKeys,
            loadedKeys: expandedKeys,
            treeData,
            checkedNodes,
          }));
          return;
        }

        setMcpProps(prevState => ({
          ...prevState,
          treeData,
        }));
      }
    } catch (error: any) {
      const { Description, ErrorDetails } = error?.response || error?.data || error || {};
      (ErrorDetails || Description) && message.error(ErrorDetails || Description);
    }
  };

  // 获取MCP服务下的工具节点
  const getMcpToolTreeNode = async (nodeData: AdTreeDataNode, handleError?: any) => {
    try {
      const mcpId = nodeData.sourceData.mcp_id as string;
      const response = await getMCPServerTools(mcpId, publicAndCurrentDomainIdsRef.current);
      if (response && response.tools) {
        // 将MCP工具数据转换为组件期望的格式
        const toolData = response.tools.map((item: MCPTool) => {
          // 处理inputSchema参数为tool_input格式
          const toolInputs = getMCPInputParamsFromOpenAPISpec(item.inputSchema);

          return {
            tool_id: item.name,
            tool_name: item.name,
            tool_desc: item.description,
            tool_input: toolInputs,
            type: 'mcp-tool',
            box_name: nodeData.title,
            box_id: nodeData.key as string,
            is_build_in: false,
            checkable: false, // MCP工具节点不显示复选框
          };
        });

        const childTreeData = adTreeUtils.createAdTreeNodeData(toolData, {
          titleField: 'tool_name',
          keyField: 'tool_id',
          parentKey: nodeData.key as string,
          keyPath: nodeData.keyPath,
        });
        console.log('getMcpToolTreeNode ---- childTreeData', childTreeData);
        return childTreeData;
      }
    } catch (error: any) {
      const { Description, ErrorDetails, description, errorDetails } = error?.response || error?.data || error || {};
      const messageString = ErrorDetails || Description || errorDetails || description;
      if (messageString) message.error(messageString);
      if (handleError) handleError(nodeData?.sourceData?.mcp_id);
      return false;
    }
  };

  // MCP数据加载
  const loadMcpData = (nodeData: AdTreeDataNode) =>
    new Promise(resolve => {
      const getTreeData = async () => {
        const childTreeData = await getMcpToolTreeNode(nodeData);
        if (childTreeData) {
          const treeData = adTreeUtils.addTreeNode(mcpProps.treeData, childTreeData as any);
          setMcpProps(prevState => ({
            ...prevState,
            treeData,
          }));
          resolve(true);
        } else {
          resolve(false);
        }
      };
      getTreeData();
    });

  // MCP搜索处理
  const handleSearchMcp = async (e: any) => {
    const value = e.target.value;
    if (value) {
      try {
        const response = await getMCPServerList(
          {
            name: value,
            page: 1,
            page_size: 100,
            status: 'published',
          },
          publicAndCurrentDomainIdsRef.current
        );

        if (response && response.data) {
          const mcpServerData = response.data.map((item: MCPServerReleaseInfo) => ({
            ...item,
            type: 'mcp-server',
            box_id: item.mcp_id,
            box_name: item.name,
            box_desc: item.description,
          }));

          let treeData = adTreeUtils.createAdTreeNodeData(mcpServerData, {
            titleField: 'box_name',
            keyField: 'box_id',
            isLeaf: false,
          });

          const serverIds = response.data.map((item: MCPServerReleaseInfo) => item.mcp_id);

          // 为每个MCP服务加载工具
          for (const nodeData of treeData) {
            const childTreeData = await getMcpToolTreeNode(nodeData);
            if (childTreeData) {
              treeData = adTreeUtils.addTreeNode(treeData, childTreeData as any);
            }
          }

          const previousCheckedKeys = getMcpProps().checkedKeys;
          const previousCheckedNodes = getMcpProps().checkedNodes;

          setMcpProps(prevState => ({
            ...prevState,
            searchText: value,
          }));
          setSearchMcpProps(prevState => ({
            ...prevState,
            treeData: treeData,
            expandedKeys: serverIds,
            loadedKeys: serverIds,
            checkedKeys: previousCheckedKeys,
            checkedNodes: previousCheckedNodes,
          }));
        }
      } catch (error) {
        console.error('MCP搜索失败:', error);
        setMcpProps(prevState => ({
          ...prevState,
          searchText: value,
        }));
        setSearchMcpProps(prevState => ({
          ...prevState,
          treeData: [],
          expandedKeys: [],
          loadedKeys: [],
          checkedKeys: getMcpProps().checkedKeys,
          checkedNodes: getMcpProps().checkedNodes,
        }));
      }
    } else {
      setMcpProps(prevState => ({
        ...prevState,
        searchText: '',
      }));
      resetSearchMcpProps();
    }
  };

  // 渲染MCP工具树
  const renderMcpTree = () => {
    return (
      <div className="dip-h-100 dip-flex-column">
        <SearchInput
          style={{ width: '100%' }}
          placeholder={intl.get('dataAgent.searchMCPName')}
          onChange={handleSearchMcp}
          debounce
        />
        <div className="dip-flex-item-full-height" style={{ overflowY: 'auto' }}>
          {mcpProps.searchText && searchMcpProps.treeData.length == 0 ? (
            <div className="dip-column-center">
              <NoResultIcon />
              <div className="dip-mt-8 dip-c-text-lower">{intl.get('global.noResult')}</div>
            </div>
          ) : (
            <AdTree
              className="ToolModal-tree"
              expandAction={false}
              selectable={false}
              treeData={mcpProps.searchText ? searchMcpProps.treeData : mcpProps.treeData}
              checkable
              titleRender={titleRender as any}
              checkedKeys={mcpProps.checkedKeys}
              onCheck={(checkedKeys, { node, checkedNodes }) => {
                if (!mcpProps.searchText && (node as AdTreeDataNode).sourceData?.type === 'mcp-server') {
                  const loadedKeys = getMcpProps().loadedKeys;
                  if (!loadedKeys.includes(node.key)) {
                    setMcpProps(prevState => ({
                      ...prevState,
                      expandedKeys: [...prevState.expandedKeys, node.key],
                    }));
                  }
                }
                const previousCheckedNodes = getMcpProps().checkedNodes;

                if (mcpProps.searchText) {
                  const searchTreeData = getSearchMcpProps().treeData;
                  const previousNodesNotInSearch = previousCheckedNodes.filter(
                    prevNode =>
                      !searchTreeData.some(
                        (searchNode: AdTreeDataNode) =>
                          searchNode.key === prevNode.key ||
                          searchNode.children?.some((child: AdTreeDataNode) => child.key === prevNode.key)
                      )
                  );

                  let mergedCheckedKeys = Array.isArray(checkedKeys)
                    ? [...checkedKeys, ...previousNodesNotInSearch.map(node => node.key)]
                    : checkedKeys;
                  let mergedCheckedNodes: any = [...(checkedNodes as AdTreeDataNode[]), ...previousNodesNotInSearch];

                  mergedCheckedNodes = _.filter(mergedCheckedNodes, (item: any) => {
                    if (!item?.isLeaf) mergedCheckedKeys = _.filter(mergedCheckedKeys, key => key !== item.value);
                    return item?.isLeaf;
                  });

                  setMcpProps(prevState => ({
                    ...prevState,
                    checkedKeys: mergedCheckedKeys,
                    checkedNodes: mergedCheckedNodes,
                  }));
                  setSearchMcpProps(prevState => ({
                    ...prevState,
                    checkedKeys: mergedCheckedKeys,
                    checkedNodes: mergedCheckedNodes,
                  }));
                } else {
                  setMcpProps(prevState => ({
                    ...prevState,
                    checkedKeys,
                    checkedNodes: checkedNodes as AdTreeDataNode[],
                  }));
                }
              }}
              loadData={loadMcpData as any}
              loadedKeys={mcpProps.searchText ? searchMcpProps.loadedKeys : mcpProps.loadedKeys}
              onLoad={loadedKeys => {
                setMcpProps(prevState => ({ ...prevState, loadedKeys }));
              }}
              expandedKeys={mcpProps.searchText ? searchMcpProps.expandedKeys : mcpProps.expandedKeys}
              onExpand={expandedKeys => {
                if (mcpProps.searchText) {
                  setSearchMcpProps(prevState => ({ ...prevState, expandedKeys }));
                } else {
                  setMcpProps(prevState => ({ ...prevState, expandedKeys }));
                }
              }}
            />
          )}
        </div>
      </div>
    );
  };

  return (
    <UniversalModal
      width={843}
      onCancel={onCancel}
      className="ToolModal"
      open
      centered
      title={intl.get('agentCommonConfig.llm.addTool')}
      footerData={[
        { label: intl.get('global.ok'), type: 'primary', onHandle: onOk, isDisabled: loading },
        { label: intl.get('global.cancel'), onHandle: () => onCancel() },
      ]}
      footerExtra={
        <span>
          {intl.get('agentCommonConfig.llm.selected', {
            count: selectedTools.length,
          })}
        </span>
      }
    >
      <div className="dip-w-100" style={{ minHeight: '400px' }}>
        {loading ? (
          <div className="dip-position-r" style={{ height: 300 }}>
            <LoadingMask loading />
          </div>
        ) : (
          <AdTab
            className="dip-flex-column ToolModal-tab"
            items={[
              {
                label: intl.get('agent.tool'),
                key: 'tool',
                children: renderToolTree(),
              },
              {
                label: 'MCP',
                key: 'mcp',
                children: renderMcpTree(),
              },
              {
                label: intl.get('dataAgent.config.skill'),
                key: 'agent',
                children: renderAgentList(),
              },
            ]}
          />
        )}
      </div>
    </UniversalModal>
  );
};

export default ({ visible, ...restProps }: any) => {
  return visible && <ToolModal {...restProps} />;
};
