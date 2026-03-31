import React, { useState, useEffect, useMemo, useCallback } from 'react';
import intl from 'react-intl-universal';
import classNames from 'classnames';
import { Button, Table, Space, message } from 'antd';
import { uniqBy, forEach, keyBy } from 'lodash';
import { PlusOutlined, SettingOutlined, RightOutlined } from '@ant-design/icons';
import { type ResultProcessStrategyType } from '@/apis/agent-factory';
import AgentIcon from '@/assets/icons/agent3.svg';
import ToolBoxIcon from '@/assets/images/tool.svg';
import ToolIcon from '@/assets/icons/tool.svg';
import MCPIcon from '@/assets/icons/mcp.svg';
import { useDeepCompareMemo, useBusinessDomain } from '@/hooks';
import DipIcon from '@/components/DipIcon';
import SkillsIcon from '@/assets/icons/skill.svg';
import {
  getMCPServerDetail,
  getMCPServerTools,
  type MCPServerReleaseInfo,
  type MCPTool,
} from '@/apis/agent-operator-integration/mcp';
import { getToolBoxMarketList, getBoxToolList } from '@/apis/agent-operator-integration';
import { getAgentsByPost, getPublishedAgentInfoList } from '@/apis/agent-factory';
import {
  type Agent,
  type SkillAgentDataSourceConfig,
  type SkillAgentLLMConfig,
  DatasourceConfigTypeEnum,
  LLMConfigTypeEnum,
} from '@/apis/agent-factory/type';
import styles from '../ConfigSection.module.less';
import SectionPanel from '../../common/SectionPanel';
import AddToolModal from '../AddToolModal';
import ToolInputParamModal from '../ToolInputParamModal';

// 工具箱详情接口
interface ToolBoxInfo {
  box_name: string;
  box_desc: string;
  metadata_type: string;
}

// 扩展技能项接口，匹配实际使用场景
interface SkillItem {
  // 基本属性 - 与API契约一致
  tool_type: string;
  metadata_type: string;
  tool_id: string;
  tool_box_id: string;
  tool_input?: Array<{
    enable: boolean;
    input_name: string;
    input_type: string;
    map_type: string;
    map_value: any;
  }>;
  intervention?: boolean;
  data_source_config?: SkillAgentDataSourceConfig;
  llm_config?: SkillAgentLLMConfig;

  result_process_strategies?: Array<ResultProcessStrategyType>;

  // 用于UI展示的扩展属性
  id?: string;
  tool_name?: string; // 技能名称
  tool_desc?: string; // 技能描述
  tool_box_name?: string; // 工具箱名称
  icon?: React.ReactNode;
  agent_version?: string; // Agent版本
  agent_timeout?: number; // Agent超时时间,

  // 树状结构支持
  children?: SkillItem[];
  isServerNode?: boolean; // 标识是否为MCP服务器节点
  isToolBoxNode?: boolean; // 标识是否为工具箱节点

  details?: any;
}

interface SkillsSectionProps {
  state?: any;
  actions?: any;
  // 只读模式 —— 查看agent的配置页面使用
  readonly?: boolean;
  // 只读模式下的技能
  viewSkills?: any;
}

const SkillsSection = (props: SkillsSectionProps) => {
  const { state, actions, readonly = false, viewSkills } = props;
  const { publicAndCurrentDomainIds } = useBusinessDomain();
  // 检查是否可编辑技能配置
  const canEditSkills = actions?.canEditField('skills');
  const canEditToolInput = actions?.canEditField('skills.tools.tool_input');

  // 添加MCP服务器详情的状态管理
  const [mcpServerDetails, setMcpServerDetails] = useState<Record<string, MCPServerReleaseInfo>>({});
  // 添加MCP工具详情的状态管理
  const [mcpToolsDetails, setMcpToolsDetails] = useState<Record<string, MCPTool[]>>({});
  // 添加工具箱详情的状态管理
  const [toolBoxDetails, setToolBoxDetails] = useState<Record<string, ToolBoxInfo>>({});
  // 添加Agent详情的状态管理
  const [agentDetails, setAgentDetails] = useState<Record<string, Agent>>({});
  // 添加工具详情的状态管理
  const [toolDetails, setToolDetails] = useState<Record<string, any>>({});

  // 将skills的各个部分合并为一个用于显示的数组
  const allSkills = useMemo(() => {
    if (readonly) {
      const agents =
        viewSkills?.agents?.map((agent: any) => ({
          ...agent,
          tool_type: 'agent',
          tool_input: agent?.agent_input,
          tool_box_id: agent?.agent_version,
          tool_id: agent?.agent_key,
          agent_timeout: agent?.agent_timeout,
        })) || [];

      const tools =
        viewSkills?.tools?.map((tool: any) => ({
          ...tool,
          tool_type: 'tool',
        })) || [];

      const mcps =
        viewSkills?.mcps?.map((mcp: any) => ({
          tool_type: 'mcp',
          tool_id: mcp.mcp_server_id,
          tool_box_id: mcp.mcp_server_id,
          intervention: false,
        })) || [];

      return [...agents, ...tools, ...mcps];
    }

    const result: SkillItem[] = [];

    // 添加工具
    if (state?.config?.skills?.tools) {
      state.config.skills.tools.forEach((tool: any) => {
        result.push({
          ...tool,
          tool_type: 'tool',
        });
      });
    }

    // 添加智能体
    if (state?.config?.skills?.agents) {
      state.config.skills.agents.forEach((agent: any) => {
        result.push({
          tool_type: 'agent',
          tool_id: agent.agent_key,
          tool_box_id: agent.agent_version || 'default',
          tool_input: agent.agent_input,
          intervention: agent.intervention,
          agent_version: agent.agent_version,
          data_source_config: agent.data_source_config || {
            type: DatasourceConfigTypeEnum.SelfConfigured,
          },
          llm_config: agent.llm_config || {
            type: LLMConfigTypeEnum.SelfConfigured,
          },
          details: agent.details,
          agent_timeout: agent?.agent_timeout,
        });
      });
    }

    // 添加MCP服务器 - 这里只是标记，实际的工具会在processedSkills中展开
    if (state?.config?.skills?.mcps) {
      state.config.skills.mcps.forEach((mcp: any) => {
        result.push({
          tool_type: 'mcp',
          tool_id: mcp.mcp_server_id,
          tool_box_id: mcp.mcp_server_id,
          intervention: false,
          details: mcp.details,
        });
      });
    }

    return result;
  }, [state?.config?.skills, viewSkills, readonly]);

  const [skills, setSkills] = useState<SkillItem[]>(allSkills);
  const [toolModalVisible, setToolModalVisible] = useState(false);
  const [toolInputParamModal, setToolInputParamModal] = useState({
    open: false,
    tool: null as any,
  });
  const [isExpanded, setIsExpanded] = useState(false);

  // 获取MCP服务器详情的函数
  const fetchMCPServerDetails = useCallback(async (serverIds: string[], publicAndCurrentDomainIds: string[]) => {
    const detailsPromises = serverIds.map(async serverId => {
      try {
        const response = await getMCPServerDetail(serverId, publicAndCurrentDomainIds);
        return { serverId, details: response.base_info };
      } catch (error) {
        console.error(`获取MCP服务器详情失败 (${serverId}):`, error);
        return { serverId, details: null };
      }
    });

    const results = await Promise.all(detailsPromises);
    const detailsMap: Record<string, MCPServerReleaseInfo> = {};

    results.forEach(({ serverId, details }) => {
      if (details) {
        detailsMap[serverId] = details;
      }
    });

    setMcpServerDetails(prev => ({ ...prev, ...detailsMap }));
  }, []);

  // 获取MCP工具详情的函数
  const fetchMCPToolsDetails = useCallback(async (serverIds: string[], publicAndCurrentDomainIds: string[]) => {
    const toolsPromises = serverIds.map(async serverId => {
      try {
        const response = await getMCPServerTools(serverId, publicAndCurrentDomainIds);
        return { serverId, tools: response.tools || [] };
      } catch (error) {
        console.error(`获取MCP工具详情失败 (${serverId}):`, error);
        return { serverId, tools: [] };
      }
    });

    const results = await Promise.all(toolsPromises);
    const toolsMap: Record<string, MCPTool[]> = {};

    results.forEach(({ serverId, tools }) => {
      toolsMap[serverId] = tools;
    });

    setMcpToolsDetails(prev => ({ ...prev, ...toolsMap }));
  }, []);

  // 获取工具箱详情的函数
  const fetchToolBoxDetails = useCallback(async (toolBoxIds: string[], publicAndCurrentDomainIds: string[]) => {
    try {
      // 使用批量接口获取所有工具箱信息
      const response = await getToolBoxMarketList(
        {
          box_ids: toolBoxIds,
          fields: 'box_name,box_desc',
        },
        publicAndCurrentDomainIds
      );

      const detailsMap: Record<string, ToolBoxInfo> = {};

      response.forEach((toolBox: any) => {
        detailsMap[toolBox.box_id] = {
          box_name: toolBox.box_name || intl.get('dataAgent.config.toolboxWithId', { id: toolBox.box_id }),
          box_desc: toolBox.box_desc || intl.get('dataAgent.config.toolboxDescription'),
          metadata_type: toolBox.metadata_type,
        };
      });

      setToolBoxDetails(prev => ({ ...prev, ...detailsMap }));
    } catch (error) {
      console.error('获取工具箱详情失败:', error);

      // 失败时使用默认值
      const detailsMap: Record<string, ToolBoxInfo> = {};
      toolBoxIds.forEach(toolBoxId => {
        detailsMap[toolBoxId] = {
          box_name: intl.get('dataAgent.config.toolboxWithId', { id: toolBoxId }),
          box_desc: intl.get('dataAgent.config.toolboxDescription'),
          metadata_type: '',
        };
      });
      setToolBoxDetails(prev => ({ ...prev, ...detailsMap }));
    }
  }, []);

  // 获取Agent详情的函数
  const fetchAgentDetails = useCallback(async (agent_keys: string[]) => {
    try {
      const response = await getPublishedAgentInfoList(agent_keys);

      const detailsMap: Record<string, Agent> = {};
      response.entries.forEach(agent => {
        (detailsMap as any)[agent.key] = agent;
      });

      setAgentDetails(detailsMap);
    } catch (error) {
      console.error('获取Agent详情失败:', error);
    }
  }, []);

  // 获取工具箱内工具详情的函数
  const fetchToolBoxToolDetails = useCallback(
    async (toolBoxId: string) => {
      try {
        const response = await getBoxToolList(
          toolBoxId,
          {
            all: true, // 获取所有工具
          },
          publicAndCurrentDomainIds
        );

        const toolsMap: Record<string, any> = {};
        response.tools.forEach(tool => {
          toolsMap[tool.tool_id] = {
            tool_name: tool.name,
            tool_desc: tool.description,
          };
        });

        setToolDetails(prev => ({ ...prev, ...toolsMap }));
      } catch (error) {
        console.error(`获取工具箱 ${toolBoxId} 工具详情失败:`, error);
      }
    },
    [publicAndCurrentDomainIds]
  );

  // 监听MCP服务器变化，获取详情信息
  useEffect(() => {
    if (!publicAndCurrentDomainIds?.length) return;

    const mcpServerIds = Array.from(
      new Set(skills.filter(skill => skill.tool_type === 'mcp').map(skill => skill.tool_box_id))
    );

    if (mcpServerIds.length > 0) {
      // 使用回调函数形式获取最新状态，避免闭包问题
      setMcpServerDetails(currentDetails => {
        const missingServerIds = mcpServerIds.filter(id => !currentDetails[id]);
        if (missingServerIds.length > 0) {
          fetchMCPServerDetails(missingServerIds, publicAndCurrentDomainIds);
        }
        return currentDetails;
      });

      setMcpToolsDetails(currentDetails => {
        const missingToolsServerIds = mcpServerIds.filter(id => !currentDetails[id]);
        if (missingToolsServerIds.length > 0) {
          fetchMCPToolsDetails(missingToolsServerIds, publicAndCurrentDomainIds);
        }
        return currentDetails;
      });
    }

    // 获取工具箱详情
    const toolBoxIds = Array.from(
      new Set(skills.filter(skill => skill.tool_type === 'tool').map(skill => skill.tool_box_id))
    );

    if (toolBoxIds.length > 0) {
      setToolBoxDetails(currentDetails => {
        const missingToolBoxIds = toolBoxIds.filter(id => !currentDetails[id]);
        if (missingToolBoxIds.length > 0) {
          fetchToolBoxDetails(missingToolBoxIds, publicAndCurrentDomainIds);
        }
        return currentDetails;
      });
    }

    // 获取Agent详情
    const agentIds = Array.from(
      new Set(skills.filter(skill => skill.tool_type === 'agent').map(skill => skill.tool_id))
    );

    if (agentIds.length > 0) {
      fetchAgentDetails(agentIds);
    }
  }, [
    skills,
    fetchMCPServerDetails,
    fetchMCPToolsDetails,
    fetchToolBoxDetails,
    fetchAgentDetails,
    publicAndCurrentDomainIds,
  ]);

  // 处理技能数据，将工具按工具箱分组、MCP工具按服务器分组为树状结构
  const processedSkills = useMemo(() => {
    if (!publicAndCurrentDomainIds) return [];
    if (!skills?.length) return [];

    // 分离不同类型的工具
    const tools = skills.filter(skill => skill.tool_type === 'tool');
    const agents = skills.filter(skill => skill.tool_type === 'agent');
    const mcpTools = skills.filter(skill => skill.tool_type === 'mcp');

    const result: SkillItem[] = [];

    // 处理普通工具 - 按工具箱分组
    if (tools.length > 0) {
      const toolBoxes = new Map<string, SkillItem[]>();

      tools.forEach(tool => {
        const toolBoxId = tool.tool_box_id;
        if (!toolBoxes.has(toolBoxId)) {
          toolBoxes.set(toolBoxId, []);
        }

        // 使用工具详情信息（如果已获取）
        const toolDetail = toolDetails[tool.tool_id];
        const enhancedTool = {
          ...tool,
          tool_name: toolDetail?.tool_name || tool.tool_name,
          tool_desc: toolDetail?.tool_desc || tool.tool_desc,
        };

        toolBoxes.get(toolBoxId)!.push(enhancedTool);
      });

      // 创建工具箱树状节点
      const toolBoxNodes: SkillItem[] = Array.from(toolBoxes.entries()).map(([toolBoxId, toolList]) => {
        // 使用获取到的工具箱详情，如果没有则使用默认值
        const toolBoxInfo = toolBoxDetails[toolBoxId];
        const toolBoxName = toolBoxInfo?.box_name || toolList[0]?.tool_box_name;
        const toolBoxDesc = toolBoxInfo?.box_desc;
        const metadata_type = toolBoxInfo?.metadata_type;
        return {
          tool_type: 'tool-box',
          tool_id: `tool-box-${toolBoxId}`,
          tool_box_id: toolBoxId,
          tool_name: toolBoxName,
          tool_desc: toolBoxDesc,
          metadata_type: metadata_type,
          isToolBoxNode: true,
          // 名称不存在时，children设置为undefined
          children: toolBoxName ? toolList : undefined,
        };
      });

      result.push(...toolBoxNodes);
    }

    // 处理Agent - 直接添加，不分组，使用获取到的详情信息
    const agentsWithDetails = agents.map(agent => {
      const agentDetail = agentDetails[agent.tool_id];
      return {
        ...agent,
        tool_name: agentDetail?.name || agent.tool_name,
        tool_desc: agentDetail?.profile || agent.tool_desc,
      };
    });
    result.push(...agentsWithDetails);

    // 处理MCP工具 - 按服务器分组
    if (mcpTools.length > 0) {
      const mcpServers = new Map<string, SkillItem[]>();
      mcpTools.forEach(tool => {
        const serverId = tool.tool_box_id;

        if (!mcpServers.has(serverId)) {
          mcpServers.set(serverId, []);
        }

        // 获取该服务器的工具详情
        const serverTools = mcpToolsDetails[serverId] || [];
        const serverDetails = mcpServerDetails[serverId];

        // 为每个MCP工具创建SkillItem，包含完整的显示信息
        const toolItems: SkillItem[] = serverTools.map(mcpTool => ({
          tool_type: 'mcp',
          tool_id: mcpTool.name,
          tool_box_id: serverId,
          tool_name: mcpTool.name,
          tool_desc: mcpTool.description,
          tool_box_name: serverDetails?.name,
          intervention: false,
        }));

        mcpServers.set(serverId, toolItems);
      });

      // 创建MCP服务器树状节点
      const mcpServerNodes: SkillItem[] = Array.from(mcpServers.entries()).map(([serverId, tools]) => {
        // 使用获取到的详情信息，如果没有则使用默认值
        const serverDetails = mcpServerDetails[serverId];
        const serverName = serverDetails?.name;
        const serverDescription = serverDetails?.description;

        return {
          tool_type: 'mcp-server',
          tool_id: `mcp-server-${serverId}`,
          tool_box_id: serverId,
          tool_name: serverName,
          tool_desc: serverDescription,
          isServerNode: true,
          // 名称不存在，则children设置为undefined
          children: serverName ? tools : undefined,
        };
      });

      result.push(...mcpServerNodes);
    }

    const setName = async () => {
      const agentIds: string[] = [];
      forEach(result, item => {
        if (item.tool_type === 'agent') agentIds.push(item.tool_id);
      });
      try {
        const { entries } = await getAgentsByPost({
          size: 1000,
          agent_keys: agentIds,
          business_domain_ids: publicAndCurrentDomainIds,
        });
        if (entries) {
          const entriesKV = keyBy(entries, 'key');
          forEach(result, item => {
            if (entriesKV[item?.tool_id]) item.tool_name = entriesKV[item?.tool_id]?.name;
          });
        }
      } catch (error) {
        console.log('error', error);
      }
    };
    setName();
    return result;
  }, [skills, mcpServerDetails, mcpToolsDetails, toolBoxDetails, agentDetails, toolDetails, publicAndCurrentDomainIds]);

  // 技能表格列定义
  const skillColumns = [
    {
      title: intl.get('dataAgent.config.skillName'),
      dataIndex: 'tool_name',
      key: 'tool_name',
      // 只读模式，名称和描述各占一半:50% * (100% - 100px)；编辑模式，名称占剩余的60%: 60% * (100% - 80px)
      width: readonly ? 'calc(50% - 50px)' : 'calc(60% - 48px)',
      onCell: () => ({
        style: { display: 'flex', alignItems: 'center' },
      }),
      render: (text: string, record: SkillItem) => {
        let Icon;
        if (record.tool_type === 'mcp-server') {
          Icon = <MCPIcon style={{ width: 32, height: 32, borderRadius: 8 }} />; // MCP服务器使用MCP图标
        } else if (record.tool_type === 'tool-box') {
          Icon = (
            <div className="dip-position-r" style={{ width: 32, height: 32 }}>
              <ToolBoxIcon style={{ width: 32, height: 32, borderRadius: 8 }} />
              <div className="toolBoxLabel">{record?.metadata_type === 'openapi' ? 'OpenAPI' : '函数计算'}</div>
            </div>
          ); // 工具箱使用工具箱图标
        } else if (record.tool_type === 'mcp' || record.tool_type === 'tool') {
          Icon = <ToolIcon style={{ width: 24, height: 24 }} />; // MCP工具和普通工具使用工具图标
        } else {
          Icon = <AgentIcon style={{ width: 32, height: 32, borderRadius: 8 }} />; // Agent工具使用Agent图标
        }

        return (
          <div className={classNames(styles['skill-name-cell'], 'dip-ellipsis')}>
            {Icon}
            <span
              className={classNames('dip-ellipsis', {
                'dip-text-color-error': !text,
              })}
              title={text}
              style={{ maxWidth: '200px' }}
            >
              {text || '---'}
            </span>
          </div>
        );
      },
    },
    {
      title: intl.get('dataAgent.config.functionDescription'),
      dataIndex: 'tool_desc',
      key: 'tool_desc',
      render: (text: string, record: SkillItem) => {
        // 名称不存在，描述显示---
        const desc = record.tool_name ? text : '---';
        return (
          <div className={styles['skill-role']}>
            <div
              className={classNames(styles['skill-description'], 'dip-ellipsis', {
                'dip-text-color-error': desc === '---',
              })}
              title={desc}
            >
              {desc || intl.get('dataAgent.config.noSkillDescription')}
            </div>
          </div>
        );
      },
    },
    {
      title: intl.get('dataAgent.config.operation'),
      key: 'action',
      width: readonly ? 140 : 80,
      render: (_: any, record: SkillItem) => {
        // MCP工具（子节点）不显示任何操作按钮
        if (record.tool_type === 'mcp') {
          return null;
        }

        return (
          <Space size="middle" className={styles['skill-actions']}>
            {/* 普通工具和Agent显示配置按钮，工具箱和MCP服务器不显示 */}
            {(record.tool_type === 'tool' && !record.isToolBoxNode) || record.tool_type === 'agent' ? (
              readonly ? (
                /* 查看配置页面，显示【查看配置】 */
                <Button
                  type="link"
                  className="dip-p-0"
                  onClick={e => {
                    e.stopPropagation();
                    configureSkill(record);
                  }}
                >
                  {intl.get('dataAgent.config.viewConfiguration')}
                </Button>
              ) : /* 编辑agent页面，显示【设置】按钮 */
              /** 名称为空时，视为不存在，则设置按钮不显示 */
              record.tool_name ? (
                <SettingOutlined
                  className="dip-c-subtext"
                  size={20}
                  onClick={e => {
                    e.stopPropagation();
                    canEditToolInput && configureSkill(record);
                  }}
                  style={{
                    cursor: canEditToolInput ? 'pointer' : 'not-allowed',
                    opacity: canEditToolInput ? 1 : 0.5,
                  }}
                />
              ) : null
            ) : null}
            {/* 所有类型都显示删除按钮，除了MCP工具（子节点）。注：查看agent配置页面，不显示删除按钮 */}
            {!readonly && (
              <Button
                className="dip-c-subtext"
                type="text"
                disabled={!canEditSkills}
                icon={<DipIcon type="icon-dip-trash" />}
                onClick={e => {
                  e.stopPropagation();
                  deleteSkill(record.tool_id);
                }}
              />
            )}
          </Space>
        );
      },
    },
  ];

  // 处理添加技能
  const handleAddSkill = () => {
    if (!canEditSkills) return;
    setToolModalVisible(true);
    setIsExpanded(true);
  };

  // 配置技能设置
  const configureSkill = (record: SkillItem) => {
    if (!canEditToolInput && !readonly) return;
    const skill = skills.find(skill => skill.tool_id === record.tool_id);
    if (skill) {
      setToolInputParamModal({
        open: true,
        tool: skill,
      });
    }
  };

  // 删除技能
  const deleteSkill = (id: string) => {
    if (!canEditSkills) return;
    let updatedSkills;

    // 如果删除的是MCP服务器节点，需要删除该服务器下的所有工具
    if (id.startsWith('mcp-server-')) {
      const serverId = id.replace('mcp-server-', '');
      updatedSkills = skills.filter(skill => !(skill.tool_type === 'mcp' && skill.tool_box_id === serverId));
      message.success(intl.get('dataAgent.config.mcpServerAndToolsDeleted'));
    }
    // 如果删除的是工具箱节点，需要删除该工具箱下的所有工具
    else if (id.startsWith('tool-box-')) {
      const toolBoxId = id.replace('tool-box-', '');
      updatedSkills = skills.filter(skill => !(skill.tool_type === 'tool' && skill.tool_box_id === toolBoxId));
      message.success(intl.get('dataAgent.config.toolboxAndToolsDeleted'));
    } else {
      // 删除其他类型的工具（MCP工具不会到这里，因为它们没有删除按钮）
      updatedSkills = skills.filter(skill => skill.tool_id !== id);
      message.success(intl.get('dataAgent.config.skillDeleted'));
    }

    // 更新本地状态
    setSkills(updatedSkills);

    // 将更新后的技能转换为新的skills结构
    const newSkills = {
      tools: updatedSkills
        .filter(skill => skill.tool_type === 'tool')
        .map(skill => ({
          tool_id: skill.tool_id,
          tool_box_id: skill.tool_box_id,
          tool_input: skill.tool_input,
          intervention: skill.intervention || false,
          details: skill.details || skill,
          result_process_strategies: skill.result_process_strategies,
        })),
      agents: updatedSkills
        .filter(skill => skill.tool_type === 'agent')
        .map(skill => ({
          agent_key: skill.tool_id,
          agent_version: skill.agent_version || skill.tool_box_id,
          agent_input: skill.tool_input || [],
          intervention: skill.intervention || false,
          details: skill.details || skill,
        })),
      mcps: updatedSkills
        .filter(skill => skill.tool_type === 'mcp')
        .reduce((acc: Array<{ mcp_server_id: string; details?: any }>, skill) => {
          // 只保留唯一的MCP服务器ID，避免重复
          const serverId = skill.tool_box_id;
          const findMCP = acc.find(mcp => mcp.mcp_server_id === serverId);
          if (!findMCP) {
            acc.push({
              mcp_server_id: serverId,
              details: {
                tools: skill.details?.tools || [skill],
              },
            });
          } else {
            findMCP.details.tools.push(skill);
          }
          return acc;
        }, []),
    };

    actions.updateSkills(newSkills);
  };

  // 处理工具选择完成
  const handleToolSelectComplete = (tools: SkillItem[]) => {
    if (!canEditSkills) return;
    if (!tools || tools.length === 0) {
      setToolModalVisible(false);
      return;
    }

    const updatedSkills = uniqBy([...skills, ...tools], 'tool_id');
    setSkills(updatedSkills);
    // 关闭模态框
    setToolModalVisible(false);

    // 将工具数组转换为新的skills结构
    const newSkills = {
      tools: updatedSkills
        .filter(tool => tool.tool_type === 'tool')
        .map(tool => ({
          tool_id: tool.tool_id,
          tool_box_id: tool.tool_box_id,
          tool_input: tool.tool_input,
          intervention: tool.intervention || false,
          details: tool.details || tool,
          result_process_strategies: tool.result_process_strategies,
        })),
      agents: updatedSkills
        .filter(tool => tool.tool_type === 'agent')
        .map(tool => ({
          agent_key: tool.tool_id,
          agent_version: tool.agent_version || tool.tool_box_id,
          agent_input: tool.tool_input || [],
          intervention: tool.intervention || false,
          data_source_config: tool.data_source_config,
          llm_config: tool.llm_config,
          details: tool.details || tool,
        })),
      mcps: updatedSkills
        .filter(tool => tool.tool_type === 'mcp')
        .reduce((acc: Array<{ mcp_server_id: string; details: any }>, tool) => {
          // 只保留唯一的MCP服务器ID，避免重复
          const serverId = tool.tool_box_id;
          const findMCP = acc.find(mcp => mcp.mcp_server_id === serverId);
          if (!findMCP) {
            acc.push({
              mcp_server_id: serverId,
              details: {
                tools: tool.details?.tools || [tool],
              },
            });
          } else {
            findMCP.details.tools.push(tool);
          }

          return acc;
        }, []),
    };

    actions.updateSkills(newSkills);
  };

  // 处理工具输入完成
  const handleToolInputComplete = (updatedTool: any) => {
    if (!canEditToolInput) return;
    // 更新本地状态
    const updatedSkills = skills.map(skill => (skill.tool_id === updatedTool?.tool_id ? updatedTool : skill));

    setSkills(updatedSkills);

    // 将更新后的技能转换为新的skills结构
    const newSkills = {
      tools: updatedSkills
        .filter(skill => skill.tool_type === 'tool')
        .map(skill => ({
          tool_id: skill.tool_id,
          tool_box_id: skill.tool_box_id,
          tool_input: skill.tool_input,
          intervention: skill.intervention || false,
          intervention_confirmation_message: skill.intervention ? skill.intervention_confirmation_message : null,
          tool_timeout: skill.tool_timeout || 300,
          details: skill.details || skill,
          result_process_strategies: skill.result_process_strategies,
        })),
      agents: updatedSkills
        .filter(skill => skill.tool_type === 'agent')
        .map(skill => ({
          agent_key: skill.tool_id,
          agent_version: skill.agent_version || skill.tool_box_id,
          agent_input: skill.tool_input || [],
          intervention: skill.intervention || false,
          intervention_confirmation_message: skill.intervention ? skill.intervention_confirmation_message : null,
          agent_timeout: skill.agent_timeout || 1800,
          data_source_config: skill?.data_source_config,
          llm_config: skill?.llm_config,
          details: skill.details || skill,
        })),
      mcps: updatedSkills
        .filter(skill => skill.tool_type === 'mcp')
        .reduce((acc: Array<{ mcp_server_id: string; details?: any }>, skill) => {
          // 只保留唯一的MCP服务器ID，避免重复
          const serverId = skill.tool_box_id;
          const findMCP = acc.find(mcp => mcp.mcp_server_id === serverId);
          if (!findMCP) {
            acc.push({
              mcp_server_id: serverId,
              details: {
                tools: skill.details?.tools || [skill],
              },
            });
          } else {
            findMCP.details.tools.push(skill);
          }
          return acc;
        }, []),
    };

    actions.updateSkills(newSkills);
    setToolInputParamModal({
      open: false,
      tool: null as any,
    });
    message.success(intl.get('dataAgent.config.toolParametersSaved'));
  };

  const varOptions = useDeepCompareMemo(() => {
    let options: any = [];
    state?.config?.input?.fields?.forEach((field: any) => {
      // 不过滤任何字段，包括query和内置变量
      let tempArr: any = [
        {
          label: field.name,
          value: field.name,
          editable: false,
          type: field.type,
        },
      ];
      if (field.type === 'file') {
        tempArr = [
          ...tempArr,
          {
            label: `${field.name}.name`,
            value: `${field.name}.name`,
            type: field.type,
          },
          {
            label: `${field.name}.content`,
            value: `${field.name}.content`,
            type: field.type,
          },
        ];
      }
      if (field.type === 'object' && field.name !== 'self_config') {
        tempArr = [
          ...tempArr,
          {
            label: `${field.name}.`,
            value: `${field.name}.`,
            editable: true,
            type: 'object',
          },
        ];
      }
      options = [...options, ...tempArr];
    });

    // 使用dolphin变量处理的选项
    const dolphinVarOptions =
      state?.dolphinVars?.map((dolphinVar: any) => ({
        editable: false,
        type: 'string',
        label: dolphinVar,
        value: dolphinVar,
      })) || [];

    return uniqBy([...options, ...dolphinVarOptions], 'value');
  }, [state?.config?.input?.fields, state?.dolphinVars]);

  const allPreviousBlockVars = useDeepCompareMemo(() => {
    return (
      (state?.config?.input?.fields || [])
        // 不过滤任何字段，包括query和内置变量
        .map((item: any) => ({
          name: item?.name,
          desc: item?.desc,
          type: item.type,
        }))
    );
  }, [state?.config?.input?.fields]);

  const SkillTable = (
    <div className={styles['skills-config']}>
      <Table
        dataSource={processedSkills}
        columns={skillColumns}
        pagination={false}
        className={styles['skills-table']}
        rowKey={record => record.tool_id}
        bordered={false}
        scroll={{ y: 300 }}
        showHeader={Boolean(processedSkills?.length)}
        expandable={{
          expandRowByClick: true,
          expandedRowRender: undefined, // 使用默认的children展开
          childrenColumnName: 'children',
          expandIcon: ({ expanded, expandable }) =>
            expandable ? (
              <RightOutlined
                className={classNames('dip-mr-12 dip-pointer dip-font-12 dip-transition-transform-30', {
                  'dip-rotate-90': expanded,
                })}
              />
            ) : null,
          onExpand: (expanded, record) => {
            // 当展开工具箱时，获取该工具箱下的工具详情
            if (expanded && record.isToolBoxNode) {
              fetchToolBoxToolDetails(record.tool_box_id);
            }
          },
        }}
      />
    </div>
  );

  return (
    <div>
      {readonly ? (
        SkillTable
      ) : (
        <SectionPanel
          title={intl.get('dataAgent.config.skill')}
          description={intl.get('dataAgent.config.agentCapabilityExtension')}
          rightElement={
            readonly ? null : (
              <Button
                icon={<PlusOutlined />}
                type="text"
                onClick={handleAddSkill}
                disabled={!canEditSkills}
                className="dip-c-link-75"
              >
                {intl.get('dataAgent.config.add')}
              </Button>
            )
          }
          isExpanded={isExpanded}
          onToggle={() => setIsExpanded(!isExpanded)}
          icon={<SkillsIcon />}
          className="dip-border-line-b"
        >
          {SkillTable}

          {/* 添加工具弹窗 */}
          <AddToolModal
            agentKey={state?.key}
            visible={toolModalVisible}
            onCancel={() => setToolModalVisible(false)}
            onConfirm={handleToolSelectComplete}
            allPreviousBlockVars={allPreviousBlockVars}
            retrieverBlockOptions={[]}
            value={skills}
          />
        </SectionPanel>
      )}

      {/* 工具参数配置弹窗 */}
      <ToolInputParamModal
        state={state}
        readonly={readonly}
        allPreviousBlockVars={allPreviousBlockVars}
        disabledIntervention={false}
        open={toolInputParamModal.open}
        tool={toolInputParamModal.tool}
        disabled={!canEditToolInput}
        varOptions={varOptions}
        onClose={() => setToolInputParamModal({ open: false, tool: null })}
        onChange={handleToolInputComplete}
      />
    </div>
  );
};

export default SkillsSection;
