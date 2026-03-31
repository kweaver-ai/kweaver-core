import React, { useState, useEffect, useRef, lazy, Suspense } from 'react';
import intl from 'react-intl-universal';
import _, { omit } from 'lodash';
import classNames from 'classnames';
import { useSearchParams } from 'react-router-dom';
import { message, Spin, Splitter } from 'antd';
import { AgentConfigProvider } from './AgentConfigContext';
import {
  getAgentDetail,
  getPublishedTemplateDetail,
  getPublishedAgentInfoList,
  getTemplateDetail,
  getAgentManagementPerm,
  getPublishedAgentInfo,
} from '@/apis/agent-factory';
import { getToolBoxMarketList } from '@/apis/agent-operator-integration';
import { getMCPServerTools } from '@/apis/agent-operator-integration/mcp';
import { useMicroWidgetProps, useBusinessDomain } from '@/hooks';
import Sidebar from './Sidebar/Sidebar';
import ConfigSection from './ConfigSection/ConfigSection';
import Header from './Header/Header';
import { hiddenBuildInFields, transformAgentInput } from './ConfigSection/utils';
import styles from './AgentConfig.module.less';
import TraceAnalysis from './TraceAnalysis';

const asyncComponents = {
  AgentDebuggerArea: lazy(() => import('./AgentDebuggerArea')),
};

// 预加载组件的函数
const preloadAgentDebugger = () => {
  // 触发import但不立即使用，浏览器会在空闲时加载
  import('./AgentDebuggerArea');
};

const AgentConfig: React.FC = () => {
  const [searchParams] = useSearchParams();
  const { publicAndCurrentDomainIds } = useBusinessDomain();
  const agentId = searchParams.get('agentId');
  const templateId = searchParams.get('templateId');
  const mode = searchParams.get('mode');
  const [showDebugger, setShowDebugger] = useState(false);
  const microWidgetProps = useMicroWidgetProps();
  const containerRef = useRef<HTMLDivElement>(null);

  // 如果有agentId或templateId，则加载详细配置
  const [initialAgentData, setInitialAgentData] = React.useState<any | null>(null);
  const [loading, setLoading] = React.useState(!!(agentId || templateId));
  const [sizes, setSizes] = useState<Array<string | number>>(['50%', '50%']);

  const [activeTab, setActiveTab] = useState('agent-config');
  const [refreshTraceAnalysis, setRefreshTraceAnalysis] = useState(false);

  const [userPermissions, setUserPermissions] = useState({
    hasPublishPermission: false,
    hasTraceAnalysisPermission: false,
  });
  const [isSkillAgent, setIsSkillAgent] = useState(true); // 当前Agent是否技能Agent

  const isEditTemplate = Boolean(templateId) && mode === 'editTemplate';

  // 获取是不是技能Agent
  useEffect(() => {
    const fetchAgentPublishedInfo = async (id: string) => {
      const info: any = await getPublishedAgentInfo(id);
      const { publish_to_where, publish_to_bes } = info || {};
      if (_.isEmpty(publish_to_where) && publish_to_bes.length === 1 && publish_to_bes[0] === 'skill_agent') {
        // 说明当前Agent只发布为技能Agent
        setIsSkillAgent(true);
      } else {
        setIsSkillAgent(false);
      }
    };
    if (agentId) {
      fetchAgentPublishedInfo(agentId);
    }
  }, []);

  // 获取用户权限
  useEffect(() => {
    const fetchPublishPerm = async () => {
      const {
        agent: { publish: publishAgentPerm, see_trajectory_analysis },
        agent_tpl: { publish: publishTemplatePerm },
      } = await getAgentManagementPerm();
      setUserPermissions(prevState => ({
        ...prevState,
        hasPublishPermission: isEditTemplate ? publishTemplatePerm : publishAgentPerm,
        hasTraceAnalysisPermission: see_trajectory_analysis ?? false,
      }));
    };

    fetchPublishPerm();
  }, []);

  // 获取skillAgent最新的输入参数
  // 逻辑：遍历新的input，判断旧的input中是否存在 name & type相同的参数，如果有 则enable、map_type、map_value、required、input_name、input_type 都使用旧的值，其它使用新值；如果不存在，则使用新值；
  const getSkillAgentDetail = async (skillAgents: any) => {
    const agent_keys = skillAgents.map(({ agent_key }: any) => agent_key);
    const { entries } = await getPublishedAgentInfoList(agent_keys);

    return skillAgents.map((skillAgent: any) => {
      // 历史数据中的agent_input
      const oldInputs = skillAgent.agent_input;
      // 通过接口查询到的最新的agent_input(注意：需要过滤掉几个内置输入参数)
      const findSkillAgent = entries.find(item => item.key === skillAgent.agent_key);
      const latestInputs =
        findSkillAgent?.config?.input?.fields?.filter(({ name }) => !hiddenBuildInFields.includes(name)) || [];

      const agent_input = latestInputs?.map(newInput => {
        const oldInput = oldInputs.find((i: any) => i?.input_name === newInput.name && i?.input_type === newInput.type);

        if (oldInput) {
          return {
            enable: oldInput?.enable,
            map_type: oldInput?.map_type,
            map_value: oldInput?.map_value,
            input_name: oldInput.input_name,
            input_type: oldInput.input_type,
            input_desc: newInput?.desc || `${newInput.name}变量`,
          };
        } else {
          return transformAgentInput(newInput);
        }
      });

      return {
        ...skillAgent,
        agent_input,
        details: findSkillAgent,
      };
    });
  };

  // 获取工具的name等信息
  const getToolDetails = async (tools: any[], publicAndCurrentDomainIds: string[]) => {
    const response = await getToolBoxMarketList(
      {
        box_ids: tools.map(({ tool_box_id }) => tool_box_id),
        fields: 'tools',
      },
      publicAndCurrentDomainIds
    );

    return tools.map(item => {
      const findToolBox = response.find(box => box.box_id === item.tool_box_id);
      const findTool = findToolBox?.tools?.find?.((tool: any) => tool.tool_id === item.tool_id);

      return {
        ...item,
        details: findTool,
      };
    });
  };

  // 获取mcp的tools信息
  const getMCPDetails = async (mcps: any[], publicAndCurrentDomainIds: string[]) => {
    const response = await Promise.all(
      mcps.map(({ mcp_server_id }) => getMCPServerTools(mcp_server_id, publicAndCurrentDomainIds))
    );

    return mcps.map((mcp, index) => ({
      ...mcp,
      details: {
        tools: response[index]?.tools,
      },
    }));
  };

  // 获取配置，用于回显在界面上
  const fetchInitialAgentData = async (
    { agentId, templateId }: { agentId?: string; templateId?: string },
    publicAndCurrentDomainIds: string[]
  ) => {
    setLoading(true);

    try {
      let detail;

      if (agentId) {
        // 编辑agent页面，获取agent配置
        detail = await getAgentDetail(agentId);
      } else {
        if (mode === 'editTemplate') {
          // 编辑模板页面，获取模板最新配置
          detail = await getTemplateDetail(templateId!);
        } else {
          // 使用已发布的模板创建agent页面，获取已发布的模板配置
          detail = await getPublishedTemplateDetail(templateId!);
        }
      }

      // const detail = agentId ? await getAgentDetail(agentId) : await getPublishedTemplateDetail(templateId!);
      const setInitialData = (config: any) => {
        if (templateId && mode === 'createAgent') {
          // 使用模板创建agent时，获取到的配置中 需去除id、key（因为id和key都是模板的，创建agent时 无需传递）
          setInitialAgentData(omit(config, ['id', 'key']));
        } else {
          setInitialAgentData(config);
        }
      };
      const skillAgents = detail.config?.skills?.agents;

      if (skillAgents?.length) {
        try {
          (detail.config.skills as { agents: any }).agents = await getSkillAgentDetail(skillAgents);
        } catch {}
      }

      const tools = detail.config?.skills?.tools;

      if (tools?.length) {
        try {
          (detail.config.skills as { tools: any }).tools = await getToolDetails(tools, publicAndCurrentDomainIds);
        } catch {}
      }

      const mcps = detail.config?.skills?.mcps;

      if (mcps?.length) {
        try {
          (detail.config.skills as { mcps: any }).mcps = await getMCPDetails(mcps, publicAndCurrentDomainIds);
        } catch {}
      }

      setInitialData(detail);
    } catch (ex: any) {
      if (ex?.description) {
        message.error(ex.description);
      }
    } finally {
      setLoading(false);
    }
  };

  const handleToggleDebugger = (show: boolean) => {
    setShowDebugger(show);

    if (show) {
      setSizes(['33%', '34%', '33%']);
    } else {
      setSizes(['50%', '50%']);
    }
  };

  useEffect(() => {
    if (!publicAndCurrentDomainIds) return;

    if (agentId) {
      fetchInitialAgentData({ agentId }, publicAndCurrentDomainIds);
    } else if (templateId) {
      fetchInitialAgentData({ templateId }, publicAndCurrentDomainIds);
    }
  }, [publicAndCurrentDomainIds]);

  useEffect(() => {
    // 配置页面，隐藏侧边栏
    microWidgetProps?.toggleSideBarShow?.(false);

    return () => {
      microWidgetProps?.toggleSideBarShow?.(true);
    };
  }, []);

  // 非模板页面，预加载AgentDebugger组件
  useEffect(() => {
    if (Boolean(templateId) && mode === 'editTemplate') return;

    preloadAgentDebugger();
  }, []);

  if (loading) {
    return (
      <div className={styles.loadingContainer}>
        <Spin size="large" tip={intl.get('dataAgent.loadingAgentConfig')} />
      </div>
    );
  }

  return (
    <AgentConfigProvider initialData={initialAgentData}>
      <div className={styles.pageLayout} ref={containerRef}>
        <Header
          isEditTemplate={isEditTemplate}
          showDebugger={showDebugger}
          onToggleDebugger={handleToggleDebugger}
          activeTab={activeTab}
          onTabChange={tab => {
            setActiveTab(tab);
          }}
          refreshTraceAnalysis={() => {
            setRefreshTraceAnalysis(!refreshTraceAnalysis);
          }}
          userPermissions={userPermissions}
          isSkillAgent={isSkillAgent}
        />
        <div className={styles.pageContent} style={{ display: activeTab === 'agent-config' ? 'flex' : 'none' }}>
          <Splitter onResize={setSizes}>
            <Splitter.Panel size={sizes[0]} min={showDebugger ? '25%' : '30%'}>
              <Sidebar />
            </Splitter.Panel>

            <Splitter.Panel size={sizes[1]} min="30%">
              <div className={classNames('dip-h-100', styles.mainContent)}>
                <ConfigSection />
              </div>
            </Splitter.Panel>

            {showDebugger && (
              <Splitter.Panel size={sizes[2] || '33%'} min="25%">
                <Suspense>
                  <asyncComponents.AgentDebuggerArea />
                </Suspense>
              </Splitter.Panel>
            )}
          </Splitter>
        </div>
        <div
          className="dip-flex-item-full-height dip-w-100"
          style={{ display: activeTab === 'trace-analysis' ? 'block' : 'none' }}
        >
          {/*<TraceAnalysis*/}
          {/*  refreshTraceAnalysis={refreshTraceAnalysis}*/}
          {/*  userPermissions={userPermissions}*/}
          {/*  isSkillAgent={isSkillAgent}*/}
          {/*/>*/}
        </div>
      </div>
    </AgentConfigProvider>
  );
};

export default AgentConfig;
