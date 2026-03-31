import React, { createContext, useContext, useState, type ReactNode, useCallback, useRef, useMemo } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import intl from 'react-intl-universal';
import { debounce, omit } from 'lodash';
import { createAgent, editAgent, editTemplate, getDolphinTemplateList } from '@/apis/agent-factory';
import { message } from 'antd';
import type {
  AgentConfig,
  LLMConfig,
  PrePostDolphinType,
  DolphinTemplateType,
  AgentDetailType,
} from '@/apis/agent-factory/type';
import { useDeepCompareEffect } from '@/hooks';
import { validateName } from '@/utils/validators';
import _ from 'lodash';

// 定义 AgentConfigState 接口，与 API 参数结构保持一致
export interface AgentConfigState extends AgentConfig {
  // Agent ID (用于编辑模式)
  id?: string | null;
  // 在内部使用的Agent ID
  agentId?: string;

  // 所属产品id
  product_key: number;

  // 是否是系统Agent
  is_system_agent?: number;

  // 由于 API 层面没有的字段或命名不一致的字段，单独处理
  key?: string;

  // 标记配置是否有未保存的更改
  isDirty?: boolean;

  // dolphin的变量
  dolphinVars: string[];

  // dolphin模板列表
  dolphinTemplateList: {
    pre_dolphin: DolphinTemplateType[];
    post_dolphin: DolphinTemplateType[];
  };
}

// 从旧数据中查找匹配项，找不到则使用模板+默认值
const mergeWithOldData = (newItems: any[], oldItems?: any[], defaultValues = { enabled: true, edited: false }) => {
  return newItems.map(item => {
    const findOld = oldItems?.find(({ key }) => key === item.key);

    // 1. 如果有旧数据 且已经编辑过内容，则仍然使用旧数据；2. 如果有旧数据 但没有编辑过内容，则使用模板内容 + 旧数据的enabled；3.如果没有旧数据，则使用模板内容 + 默认值
    return findOld?.edited
      ? findOld
      : { ...item, ...defaultValues, enabled: findOld ? findOld.enabled : defaultValues.enabled };
  });
};

// 规范化数据源配置
const normalizeDataSourceConfig = (config: AgentConfig['config']) => {
  return config;
};

// Context actions interface
interface AgentConfigActions {
  getSpecificField: (field: string, value: any) => any;
  updateDolphinVars: (vars: string[]) => void;
  updateSpecificField: (field: string, value: any) => void;
  updateMultipleFields: (updates: Record<string, any>) => void;
  updateBasicInfo: (name: string, profile: string, avatarUrl: string, avatarType?: number) => void;
  updateRoleInstruction: (systemPrompt: string, isDolphin: boolean, dolphinPrompts: string) => void;
  updateProductId: (product_key: string) => void;
  updateInputConfig: (fields: Array<{ name: string; type: string }>, tempZoneConfig?: any) => void;
  updateKnowledgeSources: (sources: {
    knowledge_network?: Array<{
      knowledge_network_id: string;
    }>;
  }) => void;
  updateSkills: (skills: {
    tools: Array<{
      tool_id: string;
      tool_box_id: string;
      tool_input?: Array<{
        enable: boolean;
        input_name: string;
        input_type: string;
        map_type: string;
        map_value: any;
      }>;
      intervention: boolean;
    }>;
    agents: Array<{
      agent_key: string;
      agent_version: string;
      agent_input: Array<{
        enable: boolean;
        input_name: string;
        input_type: string;
        map_type: string;
        map_value: any;
      }>;
      intervention: boolean;
    }>;
    mcps: Array<{
      mcp_server_id: string;
    }>;
  }) => void;
  updateModels: (models: { is_default: boolean; llm_config: LLMConfig }[]) => void;
  updateDataFlow: (enabled: boolean) => void;
  updateLongTermMemory: (enabled: boolean) => void;
  // 更新相关问题开关
  updateRelatedQuestion: (enabled: boolean) => void;
  // 更新任务规划模式配置开关
  updatPlanMode: (enabled: boolean) => void;
  updateOutput: (variables: { answer_var: string; doc_retrieval_var: string }) => void;
  updateWelcomeConfig: (welcomeMsg: string) => void;
  updatePresetQuestions: (questions: { question: string }[]) => void;
  updateAgentId: (agentId: string) => void;
  updateConfig: (config: AgentConfigState['config']) => void;
  updatePreAndPostDolphin: (params: { pre_dolphin: PrePostDolphinType[]; post_dolphin: PrePostDolphinType[] }) => void;
  // 更新pre_dolphin、post_dolphin里的itemKey对应的字段值
  updatePreAndPostDolphinByKey: (itemKey: string, updates: { enabled?: boolean; value?: string }) => void;
  saveAgent: (params?: {
    showSuccess?: boolean;
    isEditTemplate?: boolean;
    onSuccess?: (id: string) => void;
  }) => Promise<string | null>;

  // 重置isDirty状态
  resetDirtyState: () => void;

  // 检查字段是否可编辑（针对内置Agent）
  canEditField: (fieldName: string) => boolean;

  // 更新dolphin模板列表
  updateDolphinTemplateList: (params: {
    pre_dolphin: DolphinTemplateType[];
    post_dolphin: DolphinTemplateType[];
  }) => void;

  // 验证并修复技能输入引用变量
  validateAndFixInputReferences: (config: AgentConfigState['config']) => AgentConfigState['config'];

  // 更新数据源下的有效性，用于保存时校验
  updateDataSourceInvalid: (key: 'kg-experiment' | 'metric' | 'kn_entry', invalid: boolean) => void;

  // 更新数据源
  updateDataSourceNameMapping: (key: 'metric' | 'kn_entry', nameMapping: Record<string, string>) => void;

  // 获取dataSourceNameMapping
  getDataSourceNameMapping: () => {
    metric: Record<string, string>;
    kn_entry: Record<string, string>;
  };
}

// Create context with initial state and actions
const AgentConfigContext = createContext<
  | {
      state: AgentConfigState;
      actions: AgentConfigActions;
    }
  | undefined
>(undefined);

// Initial state
const initialState: AgentConfigState = {
  name: '',
  profile: '',
  avatar_type: 1,
  avatar: '1',
  product_key: 0,
  is_built_in: 0,
  is_system_agent: 0,
  isDirty: false,
  config: {
    input: {
      fields: [
        {
          name: 'query',
          type: 'string',
        },
      ],
      augment: {
        enable: false,
        data_source: {},
      },
      rewrite: {
        enable: false,
        llm_config: {
          id: '',
          name: 'test',
          max_tokens: 1000,
          retrieval_max_tokens: 1000,
          model_type: 'llm',
          temperature: 0.5,
          top_p: 0.5,
          top_k: 0,
          frequency_penalty: 0,
          presence_penalty: 0,
        },
      },
    },
    system_prompt: '',
    is_dolphin_mode: 0,
    dolphin: '',
    pre_dolphin: [],
    post_dolphin: [],
    memory: {
      is_enabled: false,
    },
    related_question: {
      is_enabled: false,
    },
    plan_mode: {
      is_enabled: false,
    },
    data_source: {
      knowledge_network: [],
    },
    skills: {
      tools: [],
      agents: [],
      mcps: [],
    },
    llms: [],
    output: {
      default_format: 'markdown',
      variables: {},
    },
    preset_questions: [],
  },
  dolphinVars: [],
  dolphinTemplateList: {
    pre_dolphin: [],
    post_dolphin: [],
  },
};

// Provider component
export const AgentConfigProvider: React.FC<{
  children: ReactNode;
  initialData?: any;
}> = ({ children, initialData }) => {
  const getInitialState = (): AgentConfigState => {
    if (!initialData) return initialState;

    try {
      return { ...initialData, isDirty: false };
    } catch {
      return initialState;
    }
  };
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const [state, setState] = useState<AgentConfigState>(getInitialState());
  const pollingIntervalRef = useRef<NodeJS.Timeout | null>(null);
  // 存储dataSource下 id 到 name的映射关系（用于AI生成的传参）
  const dataSourceNameMappingRef = useRef<{
    metric: Record<string, string>;
    kn_entry: Record<string, string>;
  }>({
    metric: {},
    kn_entry: {},
  });
  // 数据源是否无效，默认均有效
  const dataSourceInvalidRef = useRef<{ metric: boolean; kn_entry: boolean }>({
    metric: false,
    kn_entry: false,
  });

  // 保存的成功事件
  const handleSaveSuccess = (id: string, options?: { isEditTemplate: boolean }) => {
    // 新建agent成功后，使用模板创建agent成功后，给query加上agentId=xxx，这样后续刷新页面都能保留agent的信息了
    if (!options?.isEditTemplate && !searchParams.get('agentId')) {
      setTimeout(() => {
        navigate(
          {
            search: `?agentId=${id}`,
          },
          { replace: true }
        );
      }, 50);
    }
  };

  // 组件卸载时清除轮询
  React.useEffect(() => {
    return () => {
      if (pollingIntervalRef.current) {
        clearInterval(pollingIntervalRef.current);
        pollingIntervalRef.current = null;
      }
    };
  }, []);

  // Actions
  const actions: AgentConfigActions = {
    getSpecificField: (field: string, value: any) => {
      return _.get(state, field, value);
    },

    updateDolphinVars: (vars: string[]) => {
      setState(prev => ({
        ...prev,
        dolphinVars: vars,
      }));
    },

    updateMultipleFields: (updates: Record<string, any>) => {
      setState(prev => {
        const result = _.cloneDeep(prev);

        Object.entries(updates).forEach(([field, value]) => {
          _.set(result, field, value);
        });

        return {
          ...result,
          isDirty: true,
        };
      });
    },

    updateSpecificField: (field: string, value: any) => {
      const result = _.cloneDeep(state);

      _.set(result, field, value);

      setState(prev => ({
        ...prev,
        ...result,
        isDirty: true,
      }));
    },

    updateBasicInfo: (name, profile, avatarUrl, avatarType) => {
      setState(prev => ({
        ...prev,
        name,
        profile,
        avatar: avatarUrl,
        avatar_type: avatarType || prev.avatar_type,
        isDirty: true,
      }));
    },

    updateRoleInstruction: (systemPrompt, isDolphin, dolphinPrompts) => {
      setState(prev => ({
        ...prev,
        config: {
          ...prev.config,
          system_prompt: systemPrompt,
          is_dolphin_mode: isDolphin ? 1 : 0,
          dolphin: dolphinPrompts,
        },
        isDirty: true,
      }));
    },

    updateProductId: product_key => {
      setState(prev => ({
        ...prev,
        product_key,
        isDirty: true,
      }));
    },

    updateInputConfig: (fields, tempZoneConfig) => {
      setState(prev => ({
        ...prev,
        config: {
          ...prev.config,
          input: {
            ...prev.config.input,
            fields,
            temp_zone_config: tempZoneConfig,
          },
        },
        isDirty: true,
      }));
    },

    updateKnowledgeSources: sources => {
      setState(prev => ({
        ...prev,
        config: {
          ...prev.config,
          data_source: {
            ...prev.config.data_source,
            ...(sources.knowledge_network ? { knowledge_network: sources.knowledge_network } : {}),
          },
        },
        isDirty: true,
      }));
    },

    updateSkills: skills => {
      setState(prev => ({
        ...prev,
        config: {
          ...prev.config,
          skills: skills,
        },
        isDirty: true,
      }));
    },

    updateModels: models => {
      setState(prev => ({
        ...prev,
        config: {
          ...prev.config,
          llms: models,
        },
        isDirty: true,
      }));
    },

    updateDataFlow: enabled => {
      setState(prev => ({
        ...prev,
        config: {
          ...prev.config,
          is_data_flow_set_enabled: enabled ? 1 : 0,
        },
        isDirty: true,
      }));
    },

    updateLongTermMemory: enabled => {
      setState(prev => ({
        ...prev,
        config: {
          ...prev.config,
          memory: {
            ...prev.config.memory,
            is_enabled: enabled,
          },
        },
        isDirty: true,
      }));
    },

    updateRelatedQuestion: enabled => {
      setState(prev => ({
        ...prev,
        config: {
          ...prev.config,
          related_question: {
            ...prev.config.related_question,
            is_enabled: enabled,
          },
        },
        isDirty: true,
      }));
    },

    updatPlanMode: enabled => {
      setState(prev => ({
        ...prev,
        config: {
          ...prev.config,
          plan_mode: {
            ...prev.config.plan_mode,
            is_enabled: enabled,
          },
        },
        isDirty: true,
      }));
    },

    updateOutput: variables => {
      setState(prev => ({
        ...prev,
        config: {
          ...prev.config,
          output: {
            ...prev.config.output,
            variables: {
              ...prev.config.output!.variables,
              ...variables,
            },
          },
        },
        isDirty: true,
      }));
    },

    updateWelcomeConfig: welcomeMsg => {
      setState(prev => ({
        ...prev,
        config: {
          ...prev.config,
          ...(welcomeMsg
            ? {
                opening_remark_config: {
                  type: 'fixed',
                  fixed_opening_remark: welcomeMsg,
                },
              }
            : { opening_remark_config: null }),
        },
        isDirty: true,
      }));
    },

    updatePresetQuestions: questions => {
      setState(prev => ({
        ...prev,
        config: { ...prev.config, preset_questions: questions },
        isDirty: true,
      }));
    },

    updateAgentId: agentId => {
      setState(prev => ({ ...prev, agentId }));
    },

    resetDirtyState: () => {
      setState(prev => ({ ...prev, isDirty: false }));
    },

    updateConfig: config => {
      setState(prevState => ({
        ...prevState,
        config: {
          ...prevState.config,
          data_source: {
            ...prevState.config.data_source,
            advanced_config: config,
          },
        },
        isDirty: true,
      }));
    },

    updatePreAndPostDolphin: ({ pre_dolphin, post_dolphin }) => {
      setState(prev => ({
        ...prev,
        config: {
          ...prev.config,
          pre_dolphin,
          post_dolphin,
        },
      }));
    },

    updatePreAndPostDolphinByKey: (itemKey, updates) => {
      setState(prev => {
        const newConfig = {
          ...prev.config,
          pre_dolphin:
            prev.config.pre_dolphin?.map(item => (item.key === itemKey ? { ...item, ...updates } : item)) || [],
          post_dolphin:
            prev.config.post_dolphin?.map(item => (item.key === itemKey ? { ...item, ...updates } : item)) || [],
        };

        if ('enabled' in updates) {
          // 启用/禁用时，重新获取dolphin模板
          fetchDolphinTemplateDebounce({ config: newConfig });
        }

        return {
          ...prev,
          config: newConfig,
          isDirty: true,
        };
      });
    },

    saveAgent: async ({ showSuccess = true, isEditTemplate = false, onSuccess = handleSaveSuccess } = {}) => {
      // Get current state to avoid stale closure values
      const currentData = { ...state };

      try {
        // Validation for required inputs
        if (!currentData.name) {
          message.error(intl.get('dataAgent.enterAgentName'));
          return null;
        }

        // 验证名称格式
        if (!validateName(currentData.name)) {
          message.error(intl.get('dataAgent.config.nameInputRule'));
          return null;
        }

        // 验证简介是否存在
        if (!currentData.profile) {
          message.error(intl.get('dataAgent.enterDescription'));
          return null;
        }

        // 验证产品ID
        if (!currentData.product_key) {
          message.error(intl.get('dataAgent.selectProduct'));
          return null;
        }

        // 验证技能是否有效（通过名称校验：名称为空，视为无效）
        if (
          currentData.config?.skills?.tools?.some?.(item => !(item.details?.name || item.details?.tool_name)) ||
          currentData.config?.skills?.agents?.some?.(item => !(item.details?.name || item.details?.tool_name)) ||
          currentData.config?.skills?.mcps?.some?.(
            item => !item.details || item.details?.tools?.some(t => !(t.name || t.tool_name))
          )
        ) {
          message.error(intl.get('dataAgent.skillHasInvalidItems'));
          return null;
        }

        // 验证模型配置
        if (!currentData.config.llms || currentData.config.llms.length === 0) {
          message.error(intl.get('dataAgent.configureModel'));
          return null;
        }

        // 验证模型是否有效
        if (currentData.config.llms.find(({ llm_config }) => llm_config?.invalid)) {
          message.error(intl.get('dataAgent.modelHasInvalidItems'));
          return null;
        }

        // 验证指标是否有效
        if (dataSourceInvalidRef.current.metric) {
          message.error(intl.get('dataAgent.indicatorHasInvalidItems'));
          return null;
        }

        // 验证kn_entry是否有效
        if (dataSourceInvalidRef.current.kn_entry) {
          message.error(intl.get('dataAgent.knowledgeEntryHasInvalidItems'));
          return null;
        }

        if (currentData.config?.is_dolphin_mode === 1 && !currentData.config?.output?.variables.answer_var) {
          message.error(intl.get('dataAgent.config.finaloutputResultPlaceholder'));
          return null;
        }
        // 验证输入字段是否有重名
        if (currentData.config?.input?.fields) {
          const fieldNames = currentData.config.input.fields.map(field => field.name);
          const uniqueNames = new Set(fieldNames);

          if (uniqueNames.size !== fieldNames.length) {
            message.error(intl.get('dataAgent.inputFieldNameDuplicate'));
            return null;
          }
        }

        const newConfig = normalizeDataSourceConfig(actions.validateAndFixInputReferences(currentData.config));

        // dolphin模式，不传递plan_mode参数给后端
        if (newConfig.is_dolphin_mode) {
          newConfig.plan_mode = undefined;
        }

        // 将数据转换为API需要的格式
        const configData: AgentConfig = {
          name: currentData.name,
          profile: currentData.profile,
          avatar_type: currentData.avatar_type,
          avatar: currentData.avatar,
          product_key: currentData.product_key,
          ...(currentData.key && { key: currentData.key }),
          config: newConfig,
          ...(isEditTemplate ? {} : { is_system_agent: currentData.is_system_agent }), // 编辑模板页面，无需传递is_system_agent
        };

        let response;
        let successMessage;

        // 根据id判断是创建还是编辑操作
        if (currentData.id) {
          if (isEditTemplate) {
            // 编辑模板
            await editTemplate(currentData.tpl_id ?? currentData.id, configData);
            successMessage = intl.get('dataAgent.templateUpdateSuccess');
          } else {
            // 编辑agent
            await editAgent(currentData.id, configData);
            successMessage = intl.get('dataAgent.agentUpdateSuccess');
          }
          response = { id: currentData.id };
        } else {
          // 创建新的Agent
          response = await createAgent(configData);
          successMessage = intl.get('dataAgent.agentCreateSuccess');
        }
        if (response && response.id) {
          showSuccess && message.success(successMessage);
          // 更新id而不是agentId，并重置isDirty状态
          setState(prev => ({ ...prev, id: response.id, isDirty: false }));

          onSuccess?.(response.id, { isEditTemplate });

          return response.id;
        } else {
          const actionType = currentData.id ? '更新' : '创建';
          console.error(`${actionType}智能体失败`);
          // 增加duration以确保消息显示
          message.error(
            currentData.id
              ? intl.get('dataAgent.updateFailCheckNetworkOrAdmin')
              : intl.get('dataAgent.createFailCheckNetworkOrAdmin'),
            4
          );
          return null;
        }
      } catch (error) {
        console.error('智能体配置处理错误:', error);
        // 确保错误消息显示足够长时间
        message.error(`${(error as Error)?.description || intl.get('dataAgent.unknownError')}`, 5);
        return null;
      }
    },

    // 检查字段是否可编辑（针对内置Agent）
    canEditField: (fieldName: string) => {
      // 如果不是内置Agent，所有字段都可编辑
      if (!state.is_built_in) {
        return true;
      }

      // 如果是内置Agent但没有配置built_in_can_edit_fields，默认不可编辑
      const editableFields = state.config?.built_in_can_edit_fields;
      if (!editableFields) {
        return false;
      }

      // 检查指定字段是否允许编辑
      const fieldValue = editableFields[fieldName as keyof typeof editableFields];
      return !!fieldValue;
    },

    updateDolphinTemplateList: ({ pre_dolphin, post_dolphin }) => {
      setState(prev => ({
        ...prev,
        dolphinTemplateList: {
          pre_dolphin,
          post_dolphin,
        },
      }));
    },

    validateAndFixInputReferences: config => {
      const inputVarNames = config?.input?.fields?.map(({ name }: any) => name) || [];
      const vars = config?.is_dolphin_mode ? [...inputVarNames, ...state.dolphinVars] : inputVarNames;
      const agents = [...(config?.skills?.agents || [])];
      const tools = [...(config?.skills?.tools || [])];

      const loop = (toolInput: any) => {
        toolInput.forEach((item: any) => {
          if (item.children && item.children.length > 0) {
            loop(item.children);
          } else if (item.enable && item.map_type === 'var') {
            const varName = item.map_value;
            const tempName = varName.includes('.') ? varName.split('.')[0] : varName;

            if (!vars.includes(tempName)) {
              // 如果变量不存在了，则设置为默认值
              item.map_type = 'auto';
              item.map_value = null;
            }
          }
        });
      };
      agents.forEach(agent => loop(agent.agent_input));
      tools.forEach(tool => loop(tool.tool_input));

      return {
        ...config,
        skills: {
          ...config?.skills,
          agents,
          tools,
        },
      };
    },

    updateDataSourceInvalid: (key, invalid) => {
      dataSourceInvalidRef.current = {
        ...dataSourceInvalidRef.current,
        [key]: invalid,
      };
    },

    updateDataSourceNameMapping: (key: string, nameMapping) => {
      dataSourceNameMappingRef.current = {
        ...dataSourceNameMappingRef.current,
        [key]: nameMapping,
      };
    },

    getDataSourceNameMapping: () => {
      return dataSourceNameMappingRef.current;
    },
  };

  // 获取dolphin模板列表 -> 更新state.config
  const fetchDolphinTemplate = useCallback(
    async (params: { config: AgentDetailType['config']; built_in_agent_key?: string }) => {
      try {
        const { config } = params;
        // config中，需要去除 pre_dolphin和post_dolphin
        const dolphinTemplateList = await getDolphinTemplateList(params);
        // 更新模板列表
        actions.updateDolphinTemplateList(dolphinTemplateList);

        const pre_dolphin = mergeWithOldData(dolphinTemplateList.pre_dolphin, config?.pre_dolphin);
        const post_dolphin = mergeWithOldData(dolphinTemplateList.post_dolphin, config?.post_dolphin);
        // 更新config中的pre_dolphin和post_dolphin
        actions.updatePreAndPostDolphin({ pre_dolphin, post_dolphin });
      } catch (ex: any) {
        if (ex.description) {
          message.error(ex.description);
        }
      }
    },
    []
  );

  // 必须使用防抖，不然config变化时，会一直触发
  const fetchDolphinTemplateDebounce = useMemo(() => debounce(fetchDolphinTemplate, 500), []);

  useDeepCompareEffect(() => {
    // 当state.config有变更 且 是dolphin模式时，获取dolphin模板 -> 更新
    if (state.config?.is_dolphin_mode === 1) {
      fetchDolphinTemplateDebounce({
        config: state.config,
        // 内置agent的编辑页面，在获取dolphinTemplate时，需要传递key（后端有相关逻辑）；新建agent、编辑模板，无需传递key
        ...(state.is_built_in && state.key && !searchParams.get('templateId') ? { built_in_agent_key: state.key } : {}),
      });
    }
  }, [
    omit(state.config, [
      'pre_dolphin',
      'post_dolphin',
      'dolphin',
      'output',
      'preset_questions',
      'opening_remark_config',
    ]),
  ]); // 注意：这里必须剔除config.pre_dolphin和config.post_dolphin，不然会形成死循环

  return <AgentConfigContext.Provider value={{ state, actions }}>{children}</AgentConfigContext.Provider>;
};

// Hook for easier context usage
export const useAgentConfig = () => {
  const context = useContext(AgentConfigContext);
  if (context === undefined) {
    throw new Error('useAgentConfig must be used within an AgentConfigProvider');
  }
  return context;
};
