import React, { useState, useRef, useCallback, useEffect, useMemo } from 'react';
import intl from 'react-intl-universal';
import { Button, message, Select, Form, Tabs } from 'antd';
import _ from 'lodash';
import { useAgentConfig } from '../AgentConfigContext';
import { streamingOutHttp } from '@/utils/http';
import { EventSourceMessage } from '@microsoft/fetch-event-source';
import RoleInstructionIcon from '@/assets/icons/role-instruction.svg';
import AiIcon from '@/assets/icons/ai-generate.svg';
import OutputIcon from '@/assets/icons/output.svg';
import AdPromptInput from '@/components/AdPromptInput';
import Format from '@/components/Format';
import IconFont from '@/components/IconFont';
import { useDeepCompareEffect, useMicroWidgetProps, useBusinessDomain } from '@/hooks';
import { getAgentsByPost } from '@/apis/agent-factory';
import { getToolBoxMarketList, getBoxToolList } from '@/apis/agent-operator-integration';
import { getMCPServerDetail } from '@/apis/agent-operator-integration/mcp';
import AdDolphinEditor from '../../../components/Editor/AdDolphinEditor';
import SectionPanel from '../common/SectionPanel';
import { extractCustomVarFromBeforeCursor } from './utils';
import DolphinModule from './DolphinModule';
import './RoleInstruction.css';

enum TabKeyEnum {
  Prompt = 'prompt',
  Dolphin = 'dolphin',
}

const ROLE_INSTRUCTION_PLACEHOLDER = {
  'zh-cn': `设定AI应答规范请参照以下格式指南：

# 角色任务
描述决策智能体的角色人设，期望完成的主要任务或目标。
# 使用技能
描述决策智能体可用的组件，并说明如何使用这些技能。
# 要求与限制
指定回答的输入格式、结果内容、风格要求或字数限制等。`,

  'zh-tw': `設定 AI 應答規範請參照以下格式指南：

# 角色任務
描述決策智能體的角色人設，期望完成的主要任務或目標。
# 使用技能
描述決策智能體可用的元件，並說明如何使用這些技能。
# 要求與限制
指定回答的輸入格式、結果內容、風格要求或字數限制等。`,

  'en-us': `To set AI response specifications, please refer to the following format guide:

# Role Task
Describe the Decision Agent's persona and the main tasks or goals it is expected to accomplish.
# Skill Usage
Describe the components available to the Decision Agent and explain how to use these skills.
# Requirements and Restrictions
Specify the input format, result content, style requirements, word count limits, etc. for the responses.`,
};


const RoleInstruction: React.FC = () => {
  const microWidgetProps = useMicroWidgetProps();
  const { publicAndCurrentDomainIds } = useBusinessDomain();
  const lang = microWidgetProps.language.getLanguage || 'zh-cn';
  const { state, actions } = useAgentConfig();

  // 检查是否可编辑系统提示词配置
  const canEditSystemPrompt = actions.canEditField('system_prompt');

  const [isGenerating, setIsGenerating] = useState(false);
  const [streamText, setStreamText] = useState<string>(''); // 用于累积流式文本
  const streamTextRef = useRef<string>('');
  const editorArrRef = useRef<any[]>([]);
  const monacoRef = useRef<any>(null);
  const cursorPositionRef = useRef<any>(null);
  const abortControllerRef = useRef<AbortController | null>(null);
  const [form] = Form.useForm();

  // 存储工具和智能体的名称映射
  const [skillNamesMap, setSkillNamesMap] = useState<Record<string, string>>({});

  const dolphinValue = useMemo(() => {
    const dolphin = state.config.dolphin;
    if (typeof dolphin === 'string') {
      return dolphin;
    }
    // 简化处理，如果不是字符串就返回空字符串
    return '';
  }, [state.config.dolphin]);

  // 输入配置变量组成的选项数组
  const inputVarOptions = useMemo(
    () => state.config.input.fields.map(field => ({ label: field.name, value: field.name })) || [],
    [state.config.input.fields]
  );
  // dolphin变量组成的选项数组
  const dolphinVarOptions = useMemo(
    () => state.dolphinVars?.map?.(item => ({ label: item, value: item })) || [],
    [state.dolphinVars]
  );

  // dolphin模式下，编辑框里的变量选项
  const inputAndDolphinVarOptions = useMemo(
    () => _.uniqBy([...inputVarOptions, ...dolphinVarOptions], 'value'),
    [inputVarOptions, dolphinVarOptions]
  );

  // 获取技能名称映射
  const fetchSkillNames = useCallback(async () => {
    if (!publicAndCurrentDomainIds) return;

    const namesMap: Record<string, string> = {};

    try {
      // 获取智能体名称
      if (state.config.skills?.agents && state.config.skills.agents.length > 0) {
        const agentKeys = state.config.skills.agents.map(agent => agent.agent_key);
        const agentsResponse = await getAgentsByPost({
          size: 1000,
          agent_keys: agentKeys,
          business_domain_ids: publicAndCurrentDomainIds,
        });

        agentsResponse.entries.forEach(agent => {
          const key = agent.key;
          namesMap[key] = `.${agent.name}`;
        });
      }

      // 获取工具名称
      if (state.config.skills?.tools && state.config.skills.tools.length > 0) {
        // 获取所有工具箱ID
        const toolBoxIds = Array.from(new Set(state.config.skills.tools.map(tool => tool.tool_box_id).filter(Boolean)));

        if (toolBoxIds.length > 0) {
          // 获取工具箱信息
          const toolBoxMarketResponse = await getToolBoxMarketList(
            {
              box_ids: toolBoxIds,
              fields: 'box_name',
            },
            publicAndCurrentDomainIds
          );

          // 为每个工具箱获取详细的工具列表
          const toolBoxPromises = toolBoxIds.map(async boxId => {
            try {
              const boxToolsResponse = await getBoxToolList(boxId, { all: true }, publicAndCurrentDomainIds);
              const boxInfo = toolBoxMarketResponse.find(box => box.box_id === boxId);
              const boxName = boxInfo?.box_name || boxId;

              // 为该工具箱下的每个工具设置名称映射
              state.config.skills?.tools?.forEach(tool => {
                if (tool.tool_box_id === boxId) {
                  const toolInfo = boxToolsResponse.tools.find(t => t.tool_id === tool.tool_id);
                  if (toolInfo) {
                    const key = `${boxId}.${tool.tool_id}`;
                    namesMap[key] = `${boxName}.${toolInfo.name}`;
                  }
                }
              });
            } catch (error) {
              console.error('获取工具箱详情失败:', error);
            }
          });

          await Promise.all(toolBoxPromises);
        }
      }

      // 获取MCP名称
      if (state.config.skills?.mcps && state.config.skills.mcps.length > 0) {
        const mcpPromises = state.config.skills.mcps.map(async mcp => {
          try {
            const mcpDetailResponse = await getMCPServerDetail(mcp.mcp_server_id, publicAndCurrentDomainIds);
            const key = mcp.mcp_server_id;
            namesMap[key] = `.${mcpDetailResponse.base_info.name}`;
          } catch (error) {
            console.error('获取MCP详情失败:', error);
          }
        });

        await Promise.all(mcpPromises);
      }

      setSkillNamesMap(namesMap);
    } catch (error) {
      console.error('获取技能名称失败:', error);
    }
  }, [state.config.skills, publicAndCurrentDomainIds]);

  // 当技能配置发生变化时，重新获取名称映射
  useEffect(() => {
    fetchSkillNames();
  }, [fetchSkillNames]);

  // 生成带有名称的工具选项
  const generateToolOptions = useCallback(() => {
    const options: string[] = [];

    // 添加工具选项
    if (state.config.skills?.tools) {
      state.config.skills.tools.forEach(tool => {
        const key = `${tool.tool_box_id || ''}.${tool.tool_id}`;
        const name = skillNamesMap[key];
        // 如果有名称则显示名称，否则显示原始key
        options.push(name || key);
      });
    }

    // 添加智能体选项
    if (state.config.skills?.agents) {
      state.config.skills.agents.forEach(agent => {
        const key = agent.agent_key;
        const name = skillNamesMap[key];
        // 如果有名称则显示名称，否则显示原始key
        options.push(name || key);
      });
    }

    // 添加MCP选项
    if (state.config.skills?.mcps) {
      state.config.skills.mcps.forEach(mcp => {
        const key = mcp.mcp_server_id;
        const name = skillNamesMap[key];
        // 如果有名称则显示名称，否则显示原始key
        options.push(name || key);
      });
    }

    return options;
  }, [state.config.skills, skillNamesMap]);

  const toolOptions = generateToolOptions();

  const promptTrigger = useMemo(() => {
    const triggerToolOptions = toolOptions.map(option => {
      const [toolBoxName, toolName] = option.split('.');
      return {
        label: toolName + (toolBoxName ? `(${toolBoxName})` : ''),
        value: toolName,
        type: 'text',
      };
    });
    return [
      {
        character: '$',
        options: state.config.input.fields.map(field => ({
          label: field.name,
          value: `$${field.name}`, // 使用{{field.name}}格式
          type: 'text',
        })),
      },
      {
        character: '@',
        options: triggerToolOptions,
      },
    ];
  }, [toolOptions, state.config.input.fields]);

  // Function to switch to dolphin editor
  const handleUseDolphin = () => {
    if (!canEditSystemPrompt) return;
    actions.updateRoleInstruction(state.config.system_prompt || '', true, state.config.dolphin || '');
    message.success(intl.get('dataAgent.switchedToDolphinEditor'));
  };

  // Function to switch to text editor
  const handleUseTextEditor = () => {
    if (!canEditSystemPrompt) return;
    actions.updateRoleInstruction(state.config.system_prompt || '', false, state.config.dolphin || '');
    message.success(intl.get('dataAgent.switchedToNormalEditor'));
  };

  // 缓存当前编辑器光标位置
  const cacheCursorPosition = useCallback(() => {
    try {
      const editor = editorArrRef.current[0];
      if (editor) {
        const position = editor.getPosition();
        if (position) {
          cursorPositionRef.current = position;
        }
      }
    } catch (error) {
      console.error('缓存光标位置失败:', error);
    }
  }, []);

  // 处理文本区域变化
  const handleTextAreaChange = (value: string) => {
    if (!canEditSystemPrompt) return;
    setStreamText(value); // 更新显示的文本
    streamTextRef.current = value;
    actions.updateRoleInstruction(value, state.config.is_dolphin_mode === 1, state.config.dolphin || '');
  };

  // 处理流式消息
  const handleStreamMessage = (event: EventSourceMessage) => {
    if (event.data) {
      const text = event.data.trim();
      // 跳过特殊标记
      if (text === '#' || text === '[DONE]' || text === '') {
        return;
      }

      // 累积到流式文本
      setStreamText(prev => {
        const cleanedText = text
          .replace(/\s+/g, ' ') // 将多个空格合并为一个
          .trim();

        let newText: string = text;

        // 参考AS，仅处理###、##、-、:
        if ((cleanedText === '###' || cleanedText === '##') && prev) {
          newText = `\n\n${cleanedText} `;
        } else if (cleanedText === '-' && prev) {
          if (prev.endsWith('\n')) {
            // 上面的末尾已经有空行，则这里不加空行了
            newText = `${cleanedText} `;
          } else {
            newText = `\n${cleanedText} `;
          }
        } else if (cleanedText === ':') {
          newText = `${cleanedText}\n`;
        }

        // 处理有序列表：在前面加换行符
        const updateText = prev.replace(/(?<!\n)(\d+\.)/g, '\n$1') + newText;
        streamTextRef.current = updateText;
        return updateText;
      });
    }
  };

  // 流式请求结束或出错的处理
  const handleStreamClose = () => {
    // 将累积的文本作为最终结果
    if (streamTextRef.current) {
      // 格式化最终文本
      const formattedText = streamTextRef.current;

      // 更新到context
      actions.updateRoleInstruction(formattedText, state.config.is_dolphin_mode === 1, state.config.dolphin || '');

      message.success({ content: intl.get('dataAgent.roleInstructionGenerateSuccess'), key: 'generatePrompt' });
    }

    // 最后再设置生成状态为false
    setIsGenerating(false);
    // 清理引用
    abortControllerRef.current = null;
  };

  const handleStreamError = (error: any) => {
    setIsGenerating(false);
    message.error({
      content: intl.get('dataAgent.config.generationFailed') + (error?.error || intl.get('dataAgent.unknownError')),
      key: 'generatePrompt',
    });

    // 清理引用
    abortControllerRef.current = null;
  };

  // 处理生成请求的中止
  const abortGeneration = () => {
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
      abortControllerRef.current = null;
      setIsGenerating(false);
      message.info(intl.get('dataAgent.generateCancelled'));
    }
  };

  // 在组件卸载时中止请求
  useEffect(() => {
    return () => {
      if (abortControllerRef.current) {
        abortControllerRef.current.abort();
      }
    };
  }, []);

  const openDolphinDoc = () => {
    microWidgetProps?.history.navigateToMicroWidget({
      name: 'agent-list',
      path: `/dolphin-language-doc?hidesidebar=true&hideHeaderPath=true`,
      isNewTab: true,
    });
  };

  // AI生成角色指令
  const generateRoleInstruction = () => {
    if (!canEditSystemPrompt) return;
    if (isGenerating) {
      // 如果正在生成，点击按钮则中止
      abortGeneration();
      return;
    }

    setIsGenerating(true);
    setStreamText(''); // 重置流式文本
    message.loading({ content: intl.get('dataAgent.generatingRoleInstruction'), key: 'generatePrompt' });

    let sources: string[] = [];

    const { kn_entry: knEntryNameMapping, metric: metricNameMapping } = actions.getDataSourceNameMapping();

    // 获取指标的名称
    state?.config?.data_source?.metric?.forEach(item => {
      const name = metricNameMapping[item?.metric_model_id];

      if (name) {
        sources = [...sources, name];
      }
    });

    // 获取kn_entry的名称
    state?.config?.data_source?.kn_entry?.forEach(item => {
      const name = knEntryNameMapping[item?.kn_entry_id];

      if (name) {
        sources = [...sources, name];
      }
    });

    // 获取技能列表用于生成角色指令
    const skills: string[] = [];
    if (state?.config?.skills?.tools) {
      const names =
        state.config.skills.tools.map(tool => tool.details?.name || tool.details?.tool_name).filter(Boolean) || [];
      skills.push(...names);
    }
    if (state?.config?.skills?.agents) {
      const names =
        state.config.skills.agents.map(agent => agent.details?.name || agent.details?.tool_name).filter(Boolean) || [];
      skills.push(...names);
    }
    if (state?.config?.skills?.mcps) {
      state.config.skills.mcps.forEach(mcp => {
        const names = mcp.details?.tools?.map?.(({ tool_name, name }) => name || tool_name).filter(Boolean) || [];
        skills.push(...names);
      });
    }

    // 发起流式请求
    const controller = streamingOutHttp({
      url: '/api/agent-factory/v3/agent/ai-autogen',
      method: 'POST',
      body: {
        params: {
          name: state.name || '',
          profile: state.profile || '',
          skills,
          sources,
        },
        from: 'system_prompt',
      },
      onMessage: handleStreamMessage,
      onClose: handleStreamClose,
      onError: handleStreamError,
      onOpen: controller => {
        abortControllerRef.current = controller;
      },
    });

    // 存储controller以便后续可能的中止操作
    abortControllerRef.current = controller;
  };

  // 判断是否是海豚编辑器模式
  const isDolphinMode = state.config.is_dolphin_mode === 1;

  // 处理用户编辑器的内容变化
  const handleUserEditorChange = useCallback(
    (value: string | undefined) => {
      cacheCursorPosition();
      actions.updateRoleInstruction(state.config.system_prompt || '', state.config.is_dolphin_mode === 1, value || '');
    },
    [cacheCursorPosition, actions, state.config.is_dolphin_mode, state.config.system_prompt]
  );

  const allOptions = useMemo(() => {
    const preAndPostDolphinStr = [...(state.config?.pre_dolphin || []), ...(state.config?.post_dolphin || [])]
      .filter(item => item.enabled)
      .map(item => item.value)
      .join('\n');
    return extractCustomVarFromBeforeCursor(dolphinValue + '\n' + preAndPostDolphinStr);
  }, [dolphinValue, state.config?.pre_dolphin, state.config?.post_dolphin]);
  const options = _.uniqWith(allOptions, _.isEqual);

  useDeepCompareEffect(() => {
    actions.updateDolphinVars(options.map(({ value }) => value));
  }, [options]);

  // 当 dolphin 值变化时，同步检查表单中已选值是否仍然有效
  useEffect(() => {
    const fieldsToUpdate: Record<string, any> = { ...state.config.output?.variables };
    if (isDolphinMode) {
      const formValues = form.getFieldsValue();
      const optionValues = options.map(opt => opt.value);
      let needUpdate = false;
      // 检查最终输出结果
      if (fieldsToUpdate.answer_var && !optionValues.includes(fieldsToUpdate.answer_var)) {
        fieldsToUpdate.answer_var = undefined;
        needUpdate = true;
      }

      // 检查文档召回结果
      if (formValues.doc_retrieval_var && !optionValues.includes(formValues.doc_retrieval_var)) {
        fieldsToUpdate.doc_retrieval_var = undefined;
        needUpdate = true;
      }

      // 如果有字段需要更新，调用 setFieldsValue
      if (needUpdate) {
        form.setFieldsValue(fieldsToUpdate);
        // 更新到 context
        actions.updateOutput({ ...formValues, ...fieldsToUpdate } as any);
      }
    } else {
      // 说明切换到角色指令模式
      if (fieldsToUpdate.answer_var !== 'answer') {
        fieldsToUpdate.answer_var = 'answer';
        form.setFieldsValue({ answer_var: undefined });
        actions.updateOutput({ ...fieldsToUpdate } as any);
      }
    }
  }, [isDolphinMode, options, actions]);

  // 在值变化时更新到context
  const handleFormValuesChange = useCallback(
    (changedValues: Record<string, any>, allValues: Record<string, any>) => {
      // 更新dolphin中的值
      if (state.config.is_dolphin_mode === 1) {
        actions.updateOutput(allValues as any);
      }
    },
    [actions, state.config]
  );

  return (
    <>
      <SectionPanel
        className="dip-pt-12"
        title={
          <Tabs
            className="role-instruction-tabs"
            items={[
              {
                key: TabKeyEnum.Prompt,
                label: intl.get('dataAgent.config.roleCommands'),
                disabled: isGenerating || !canEditSystemPrompt,
              },
              {
                key: TabKeyEnum.Dolphin,
                label: 'Dolphin',
                disabled: isGenerating || !canEditSystemPrompt,
              },
            ]}
            activeKey={isDolphinMode ? TabKeyEnum.Dolphin : TabKeyEnum.Prompt}
            onChange={key => {
              if (key === TabKeyEnum.Dolphin) {
                handleUseDolphin();
              } else {
                handleUseTextEditor();
              }
            }}
          />
        }
        isExpanded={true}
        showCollapseArrow={false}
        icon={isDolphinMode ? <RoleInstructionIcon /> : <RoleInstructionIcon />}
        rightElement={
          <div className="instruction-buttons">
            {!isDolphinMode ? (
              <>
                <Button
                  type="text"
                  icon={<AiIcon />}
                  onClick={generateRoleInstruction}
                  className="dip-c-link-75 dip-pl-4 dip-pr-6"
                  loading={isGenerating}
                  disabled={!canEditSystemPrompt}
                >
                  {intl.get('dataAgent.config.AIGenerate')}
                </Button>
              </>
            ) : (
              <>
                <Format.Button
                  tip={intl.get('dataAgent.config.Documentation', { language: 'Dolphin Language' })}
                  tipPosition="top"
                  type="icon"
                  onClick={openDolphinDoc}
                >
                  <IconFont style={{ fontSize: 14 }} type="icon-wendang-xianxing" />
                </Format.Button>
              </>
            )}
          </div>
        }
      >
        <div className="dip-pt-8">
          {isDolphinMode ? (
            <>
              {/* dolphin模块 */}
              <DolphinModule toolOptions={toolOptions} inputAndDolphinVarOptions={inputAndDolphinVarOptions} />

              <div className="dip-position-r">
                <AdDolphinEditor
                  value={dolphinValue}
                  onChange={handleUserEditorChange}
                  placeholder={intl.get('dataAgent.config.inputEditorPlaceholder', { language: 'Dolphin language' })}
                  placeholderStyle={{ zIndex: 1 }}
                  disabled={isGenerating || !canEditSystemPrompt}
                  promptVarOptions={inputAndDolphinVarOptions}
                  toolOptions={toolOptions}
                  minHeight={364}
                  onMount={(editor, monaco) => {
                    editorArrRef.current[1] = editor;
                    monacoRef.current = monaco;
                  }}
                  onMouseUp={() => {
                    cacheCursorPosition();
                  }}
                  onError={errors => {
                    // 可以处理错误提示
                    console.log('编辑器错误:', errors);
                  }}
                />
              </div>
              <div className="dip-pt-0">
                <div className="dip-flex-align-center dip-mb-12 dip-mt-12">
                  <OutputIcon />
                  <div className="dip-ml-8 dip-font-14">{intl.get('dataAgent.config.outputConfig')}</div>
                </div>
                <Form
                  layout="vertical"
                  form={form}
                  onValuesChange={handleFormValuesChange}
                  initialValues={{
                    ...state.config.output?.variables,
                    answer_var: state.config.output?.variables.answer_var || undefined,
                    doc_retrieval_var: state.config.output?.variables.doc_retrieval_var || undefined,
                  }}
                >
                  <Form.Item
                    label={intl.get('dataAgent.config.finalOutputResult')}
                    name="answer_var"
                    required={true}
                    className="dip-mb-12"
                    rules={[
                      {
                        required: true,
                        message: intl.get('dataAgent.config.pleaseSelectFinalOutputResult'),
                      },
                    ]}
                    validateTrigger={['onBlur']}
                  >
                    <Select
                      style={{ width: '100%' }}
                      allowClear
                      options={options}
                      placeholder={intl.get('dataAgent.config.finaloutputResultPlaceholder')}
                    />
                  </Form.Item>
                </Form>
              </div>
            </>
          ) : (
            <AdPromptInput
              value={streamText || state.config.system_prompt}
              bordered={false}
              onChange={inputValue => {
                handleTextAreaChange(inputValue);
              }}
              placeholder={ROLE_INSTRUCTION_PLACEHOLDER[lang] || ROLE_INSTRUCTION_PLACEHOLDER['zh-cn']}
              style={{ minHeight: 200 }}
              trigger={promptTrigger}
              disabled={isGenerating || !canEditSystemPrompt}
              className="instruction-textarea"
            />
          )}
        </div>
      </SectionPanel>
    </>
  );
};

export default RoleInstruction;
