import { useEffect, useState, useRef, useMemo } from 'react';
import UniversalModal from '@/components/UniversalModal';
import intl from 'react-intl-universal';
import { Divider, Input, InputNumber, message, Switch, type TableColumnsType, Tooltip } from 'antd';
import ADTable from '@/components/ADTable';
import classNames from 'classnames';
import ErrorTip from '@/components/ErrorTip';
import { getAgentDetailInUsagePage, type ResultProcessStrategyType } from '@/apis/agent-factory';
import { getInputParamsFromOpenAPISpec } from '../utils';
import _ from 'lodash';
import './style.less';
import ParamMapSelect from './ParamMapSelect';
import { adTreeUtils } from '@/utils/handle-function';
import LoadingMask from '@/components/LoadingMask';
import { QuestionCircleFilled } from '@ant-design/icons';
import { useLatestState, useBusinessDomain } from '@/hooks';
import { getToolById } from '@/apis/agent-operator-integration';
import AgentDataSourceSettings from './AgentDataSourceSettings';
import ResultHandlingStrategy from './ResultHandlingStrategy';
import DipModal from '@/components/DipModal';
import DipButton from '@/components/DipButton';
import DipIcon from '@/components/DipIcon';

// 用于拼接字符串的特殊字段，以下特殊字符并不能直接在键盘上输入
const stringSeparator1 = '※';

enum ShowTypeEnum {
  // 输入参数
  Inputs = 'Inputs',

  // 输出参数
  Outputs = 'Outputs',

  // 结果处理策略
  Result = 'Result',

  // 超时时间
  Timeout = 'Timeout',
}

const ToolInputParamModal = ({
  readonly = false,
  state,
  onClose,
  tool,
  disabled,
  onChange,
  allPreviousBlockVars,
  varOptions,
  disabledIntervention,
}: any) => {
  const { publicAndCurrentDomainIds } = useBusinessDomain();
  const [showType, setShowType] = useState<ShowTypeEnum>(ShowTypeEnum.Inputs);
  const [inputList, setInputList] = useState<any>([]);
  const [expandedRowKeys, setExpandedRowKeys] = useState<any>([]);
  const [outputList, setOutputList] = useState<any>([]);
  const [loading, setLoading] = useState(true);
  const [errorData, setErrorData, getErrorData] = useLatestState([]);
  const [intervention, setIntervention] = useState(false);
  const [timeout, setTimeout] = useState<number>(0);
  // intervention 确认 Modal 相关状态
  const [interventionModalVisible, setInterventionModalVisible] = useState(false);
  const [interventionConfirmationMessage, setInterventionConfirmationMessage] = useState<string | null>(null);
  const [tempInterventionMessage, setTempInterventionMessage] = useState<string>('');

  const dataSourceConfigRef = useRef(tool?.data_source_config);
  const llmConfigRef = useRef(tool?.llm_config);
  const inputListRef = useRef<any[]>([]);
  const resultProcessStrategiesRef = useRef<Array<ResultProcessStrategyType> | undefined>(
    tool?.result_process_strategies
  );

  const isAgent = useMemo(() => tool?.tool_type === 'agent', [tool]);

  useEffect(() => {
    inputListRef.current = inputList;
  }, [inputList]);

  const validateError = () => {
    const toolInputError: any = [];
    const inputVarNames = allPreviousBlockVars.map(({ name }: any) => name);
    const vars = state.config?.is_dolphin_mode ? [...inputVarNames, ...state.dolphinVars] : inputVarNames;

    const loop = (toolInput: any) => {
      toolInput.forEach((item: any) => {
        if (item.children && item.children.length > 0) {
          loop(item.children);
        } else if (item.enable) {
          if (item.map_type !== 'auto') {
            if (!item.map_value) {
              toolInputError.push({
                name: item.input_name,
                error: intl.get('global.noNull'),
              });
            } else if (item.map_type === 'var') {
              const varName = item.map_value;
              const tempName = varName.includes('.') ? varName.split('.')[0] : varName;

              if (!vars.includes(tempName)) {
                toolInputError.push({
                  name: item.input_name,
                  error: intl.get('agentCommonConfig.variableNotExist', {
                    varName,
                  }),
                });
              }
            }
          }
        }
      });
    };
    loop(inputList);
    setErrorData(toolInputError);
  };

  useEffect(() => {
    if (!publicAndCurrentDomainIds) return;
    getToolParam(publicAndCurrentDomainIds);
  }, [publicAndCurrentDomainIds]);

  const flatToolInput = (inputData: any, fieldName: string = 'input_name') => {
    const cloneInputData = _.cloneDeep(inputData);
    const flatToolInputDataSource: any = [];
    const loop = (data: any, parentKey: string) => {
      data.forEach((item: any) => {
        const key = parentKey ? `${parentKey}${stringSeparator1}${item[fieldName]}` : item[fieldName];
        const tempObj = {
          ...item,
          key,
          parentKey,
          type: item.children && item.children.length > 0 ? 'group' : 'item',
        };
        delete tempObj.children;
        flatToolInputDataSource.push(tempObj);
        if (item.children && item.children.length > 0) {
          loop(item.children, key);
        }
      });
    };
    loop(cloneInputData, '');
    return flatToolInputDataSource;
  };

  const removeChildren = (dataSource: any) => {
    return dataSource.map((oldItem: any) => {
      const tempObj = {
        ...oldItem,
      };
      if (tempObj.children && tempObj.children.length > 0) {
        return {
          ...tempObj,
          children: removeChildren(tempObj.children),
        };
      } else {
        delete tempObj.children;
        return tempObj;
      }
    });
  };

  const getToolParam = async (publicAndCurrentDomainIds: string[]) => {
    try {
      setLoading(false);

      // 如果是agent类型
      if (tool.tool_type === 'agent') {
        // 如果tool已有配置的tool_input
        if (tool.tool_input && tool.tool_input.length > 0) {
          const params = tool.tool_input.map((item: any) => ({
            ...item,
            in: 'variable', // 标识为变量来源
            key: item.input_name,
            map_value: item.map_value,
            required: item.input_name === 'query',
          }));

          // 转换为树状结构
          const treeStructure = removeChildren(adTreeUtils.flatToTreeData(flatToolInput(params)));
          setInputList(treeStructure);
        } else {
          setInputList([]);
        }

        const agentDetail = await getAgentDetailInUsagePage({ id: tool.tool_id, version: 'latest', is_visit: false });

        // agent的输出参数，使用answer_var对应的value作为变量名
        const answerItem = agentDetail.config.output?.variables?.['answer_var'];

        if (answerItem) {
          setOutputList([
            {
              key: answerItem,
              output_desc: intl.get('dataAgent.config.finalOutput'),
              output_name: answerItem,
              output_type: '---',
              parentKey: '',
              type: 'item',
            },
          ]);
        } else {
          setOutputList([]);
        }

        setIntervention(tool.intervention || false);
        setInterventionConfirmationMessage(tool.intervention_confirmation_message || null);
        setTimeout(tool.agent_timeout || 1800);
        return;
      }

      // 对于普通工具，使用原有的API调用逻辑
      const toolResponse = await getToolById(tool.tool_box_id, tool.tool_id, publicAndCurrentDomainIds);
      if (toolResponse) {
        // 新接口直接返回工具对象，不需要从tools数组中查找
        const targetTool = toolResponse;
        const apiSpec: any = targetTool.metadata?.api_spec;

        // 处理输入参数：合并 parameters 和 request_body 中的参数
        const inputParams: any[] = getInputParamsFromOpenAPISpec(apiSpec);

        // 如果tool.tool_input为空或者结构不完整，直接使用新生成的inputParams
        if (!tool.tool_input || tool.tool_input.length === 0) {
          // 为新参数添加默认配置
          const addDefaultConfig = (params: any[], parentKey: string = ''): any[] => {
            return params.map((param: any) => {
              const key = parentKey ? `${parentKey}${stringSeparator1}${param.input_name}` : param.input_name;
              const newParam = {
                ...param,
                key: key, // 使用层级结构生成key
                map_type: 'auto', // 不管参数是否必填、是否启用，默认值都得是"模型生成"
                map_value: undefined,
                enable: param.required || false,
              };

              if (param.children && param.children.length > 0) {
                newParam.children = addDefaultConfig(param.children, key);
              }

              return newParam;
            });
          };

          const configuredInputs = addDefaultConfig(inputParams);

          // 直接使用配置好的嵌套结构，不经过flat转换
          setInputList(configuredInputs);

          // 设置所有object类型的项目为展开状态
          const collectExpandedKeys = (items: any[], keys: string[] = []): string[] => {
            items.forEach(item => {
              if (item.children && item.children.length > 0) {
                keys.push(item.key);
                collectExpandedKeys(item.children, keys);
              }
            });
            return keys;
          };

          const expandedKeys = collectExpandedKeys(configuredInputs);
          setExpandedRowKeys(expandedKeys);
        } else {
          // 只读模式下：保留用户已保存的全部参数（即使已删除该字段也继续展示），
          // 不追加后端返回的新增参数字段，仅同步描述、类型等元信息，确保界面稳定。
          const updateInputMetadataInReadonly = (
            existingInput: any[],
            newParams: any[],
            parentKey: string = ''
          ): any[] => {
            return existingInput.map((item: any) => {
              // 生成key，保持与表格rowKey一致
              const key = parentKey ? `${parentKey}${stringSeparator1}${item.input_name}` : item.input_name;

              // 从新参数中查找对应的参数定义
              const findMatchingNewParam = (params: any[], targetName: string): any => {
                for (const param of params) {
                  if (param.input_name === targetName) {
                    return param;
                  }
                  if (param.children && param.children.length > 0) {
                    const found = findMatchingNewParam(param.children, targetName);
                    if (found) return found;
                  }
                }
                return null;
              };

              const matchingParam = findMatchingNewParam(newParams, item.input_name);

              const updatedItem = {
                ...item,
                key, // 添加key属性
              };

              // 更新元信息（但保留用户配置）
              if (matchingParam) {
                updatedItem.input_desc = matchingParam.input_desc || item.input_desc;
                updatedItem.required = matchingParam.required ?? item.required;
                updatedItem.in = matchingParam.in || item.in;
                updatedItem.input_type = matchingParam.input_type || item.input_type;
                updatedItem.defaultValue = matchingParam.defaultValue || item.defaultValue;
              }

              // 处理children - 优先使用已有的children，如果没有则使用新参数定义中的children
              if (item.children && item.children.length > 0) {
                // 如果已有children，递归更新
                updatedItem.children = updateInputMetadataInReadonly(item.children, newParams, key);
              } else if (matchingParam && matchingParam.children && matchingParam.children.length > 0) {
                // 如果没有已有children但新参数定义中有children，则使用新的children并保持配置结构
                updatedItem.children = matchingParam.children.map((child: any) => ({
                  ...child,
                  key: `${key}${stringSeparator1}${child.input_name}`,
                  map_type: 'auto', // 不管参数是否必填、是否启用，默认值都得是"模型生成"
                  map_value: undefined,
                  enable: child.required || false,
                }));
              }

              return updatedItem;
            });
          };

          // 编辑模式下：以新参数为准，仅保留已存在参数的用户配置（enable / map_type / map_value），
          // 不存在的参数直接丢弃；同时同步更新参数的元信息（描述、类型、是否必填等）。
          const updateInputMetadataInEdit = (existingInput: any[], newParams: any[], parentKey: string = ''): any[] => {
            // 首先遍历新参数，确保所有新参数都在结果中
            return newParams.map((newParam: any) => {
              const key = parentKey ? `${parentKey}${stringSeparator1}${newParam.input_name}` : newParam.input_name;

              // 在已有配置中查找对应的参数
              const findMatchingExistingParam = (params: any[], targetName: string): any => {
                for (const param of params) {
                  if (param.input_name === targetName) {
                    return param;
                  }
                  if (param.children && param.children.length > 0) {
                    const found = findMatchingExistingParam(param.children, targetName);
                    if (found) return found;
                  }
                }
                return null;
              };

              const existingParam = findMatchingExistingParam(existingInput, newParam.input_name);

              // 合并逻辑：以新参数为基础，保留用户配置
              const mergedParam = {
                ...newParam,
                key,
                // 保留用户设置
                enable: existingParam ? existingParam.enable : newParam.required || false,
                map_type: existingParam ? existingParam.map_type : 'auto',
                map_value: existingParam ? existingParam.map_value : undefined,
              };

              // 递归处理children
              if (newParam.children && newParam.children.length > 0) {
                const existingChildren = existingParam?.children || [];
                mergedParam.children = updateInputMetadataInEdit(existingChildren, newParam.children, key);
              }

              return mergedParam;
            });
          };

          const updatedInputList = readonly
            ? updateInputMetadataInReadonly(tool.tool_input, inputParams)
            : updateInputMetadataInEdit(tool.tool_input, inputParams);
          setInputList(updatedInputList);

          // 设置展开的行keys
          const collectExpandedKeys = (items: any[], keys: string[] = []): string[] => {
            items.forEach(item => {
              if (item.children && item.children.length > 0) {
                keys.push(item.key); // 使用key而不是input_name
                collectExpandedKeys(item.children, keys);
              }
            });
            return keys;
          };

          const expandedKeys = collectExpandedKeys(updatedInputList);
          setExpandedRowKeys(expandedKeys);
        }

        // 处理输出参数：从 responses 中获取
        if (apiSpec?.responses) {
          let outputData: any = [];
          const getSchemaFromComponents = (key: string) => {
            const schema = apiSpec?.components?.schemas?.[key];
            if (!schema) return null;

            const result: any = [];
            _.forEach(schema?.properties, (item: any, key: string) => {
              if (item?.items) {
                const refs = _.split(item?.items['$ref'], '/');
                result.push({
                  output_name: key,
                  output_type: schema.type,
                  output_desc: schema.description || '',
                  children: getSchemaFromComponents(refs?.[refs.length - 1]),
                });
              } else {
                result.push({
                  output_name: key,
                  output_type: item.type || 'string',
                  output_desc: item.description || '',
                });
              }
            });
            return result;
          };
          const constructOutput = (data: any) => {
            const schema = data?.content['application/json']?.schema;
            if (schema.properties) {
              _.forEach(schema.properties, (item: any, key: string) => {
                outputData.push({
                  output_name: key,
                  output_type: item.type || 'string',
                  output_desc: item.description || '',
                });
              });
            } else if (schema['$ref']) {
              const refs = _.split(schema['$ref'], '/');
              outputData = getSchemaFromComponents(refs?.[refs.length - 1]);
            }
          };

          const data = _.filter(apiSpec?.responses, (item: any) => item?.status_code === '200')?.[0];
          if (data) constructOutput(data);
          const __outputParam = removeChildren(adTreeUtils.flatToTreeData(flatToolInput(outputData, 'output_name')));
          setOutputList(__outputParam);
        } else {
          setOutputList([]);
        }

        setIntervention(tool.intervention);
        setInterventionConfirmationMessage(tool.intervention_confirmation_message || null);
        setTimeout(tool.tool_timeout || 300);
      }
    } catch (error: any) {
      setLoading(false);
      const { Description, ErrorDetails } = error?.response || error?.data || error || {};
      (ErrorDetails || Description) && message.error(ErrorDetails || Description);
    }
  };
  const onOk = () => {
    validateError();
    if (!_.isEmpty(getErrorData())) {
      return;
    }
    const loop = (dataSource: any) => {
      return dataSource.map((oldItem: any) => {
        const tempObj = {
          input_name: oldItem?.input_name,
          input_type: oldItem?.input_type,
          input_desc: oldItem?.input_desc,
        };
        if (oldItem.children && oldItem.children.length > 0) {
          return {
            ...tempObj,
            children: loop(oldItem.children),
          };
        }
        return {
          ...tempObj,
          map_type: oldItem?.map_type,
          map_value: oldItem?.map_value,
          enable: oldItem?.enable ?? false,
        };
      });
    };
    const cloneTool = _.cloneDeep(tool);
    cloneTool.tool_input = loop(inputList);
    cloneTool.intervention = intervention;
    cloneTool.intervention_confirmation_message = intervention ? interventionConfirmationMessage : null;
    if (isAgent) {
      cloneTool.data_source_config = dataSourceConfigRef.current;
      cloneTool.llm_config = llmConfigRef.current;
      cloneTool.agent_timeout = timeout || 1800;
    } else if (tool?.tool_type === 'tool') {
      cloneTool.result_process_strategies = resultProcessStrategiesRef.current;
      cloneTool.tool_timeout = timeout || 300;
    }
    onChange?.(cloneTool);
  };

  const getErrorText = (input_name: string) => {
    const target: any = errorData.find((item: any) => item.name === input_name);
    if (target) {
      return target.error;
    }
  };

  const inputColumns: TableColumnsType = [
    {
      title: intl.get('agentTool.paramName'),
      dataIndex: 'input_name',
      render: (value: string, record: any) => {
        return (
          <div style={{ display: 'inline-block', width: '80%' }}>
            <div className="dip-ellipsis" title={value}>
              {value}
            </div>
            <div className="dip-c-subtext dip-font-12 dip-ellipsis" title={record.input_desc}>
              {record.input_desc || intl.get('global.notDes')}
            </div>
          </div>
        );
      },
    },
    {
      title: intl.get('agentTool.paramType'),
      dataIndex: 'input_type',
      width: 100,
      ellipsis: true,
    },
    {
      title: intl.get('dataAgent.config.source'),
      dataIndex: 'in',
      width: 80,
      ellipsis: true,
      align: 'center',
      render: (value: any, record: any) => {
        if (record.children) {
          return null;
        }
        // 根据in字段值显示变量来源
        if (value === 'variable') {
          return intl.get('dataAgent.config.variable');
        } else if (value === 'body') {
          return 'Body';
        } else if (value === 'path') {
          return 'Path';
        } else if (value === 'query') {
          return 'Query';
        } else {
          return 'Header';
        }
      },
    },
    {
      title: intl.get('global.required'),
      dataIndex: 'required',
      width: 100,
      ellipsis: true,
      align: 'center',
      render: (value: any, record: any) => {
        if (record.children) {
          return null;
        }
        return value ? intl.get('global.required') : intl.get('global.unRequired');
      },
    },
    {
      title: '默认值',
      dataIndex: 'defaultValue',
      width: 100,
      ellipsis: true,
      align: 'center',
      render: (value: any) => {
        return value || '---';
      },
    },
    {
      title: (
        <span>
          <span>{intl.get('global.enable')}</span>
          <Tooltip title={<div>{intl.get('dataAgent.config.enableToolParams')}</div>} placement="top">
            <QuestionCircleFilled className="dip-c-watermark dip-ml-8" />
          </Tooltip>
        </span>
      ),
      dataIndex: 'enable',
      width: 100,
      ellipsis: true,
      align: 'center',
      render: (value: any, record: any) => {
        if (record.children) {
          return null;
        }
        return (
          <Switch
            disabled={record.required || readonly}
            // disabled={readonly}
            size="small"
            checked={value ?? false}
            onChange={checked => {
              const cloneInputList = flatToolInput(inputListRef.current);
              cloneInputList.forEach((item: any) => {
                if (item.key === record.key) {
                  item.enable = checked;
                  if (checked && record.defaultValue && item.map_type === 'fixedValue' && !item.map_value) {
                    item.map_value = record.defaultValue;
                  }
                }
              });
              const newInputList = removeChildren(adTreeUtils.flatToTreeData(cloneInputList));
              setInputList(newInputList);
            }}
          />
        );
      },
    },
    {
      title: (
        <span>
          <span>{intl.get('agentCommonConfig.llm.value')}</span>
          <Tooltip
            title={
              <div>
                <div>{intl.get('agentCommonConfig.llm.fixedValueTip')}</div>
                <div>{intl.get('agentCommonConfig.llm.modelAutoTip')}</div>
                <div>{intl.get('agentCommonConfig.llm.useVariableTip')}</div>
                <div>{intl.get('agentCommonConfig.llm.selectModelTip')}</div>
              </div>
            }
            placement="top"
          >
            <QuestionCircleFilled className="dip-c-watermark dip-ml-8" />
          </Tooltip>
        </span>
      ),
      dataIndex: 'value',
      ellipsis: true,
      render: (value: string, record: any) => {
        if (record.children) {
          return null;
        }
        return (
          <ErrorTip
            errorText={getErrorText(record.input_name)}
            getPopupContainer={() => document.querySelector('.ToolParamTable .agent-web-table-body > table')}
          >
            <ParamMapSelect
              readonly={readonly}
              varOptions={varOptions}
              value={{ map_type: record.map_type, map_value: record.map_value }}
              onChange={(selectValue: any) => {
                // 必须使用inputListRef，否则拿不到最新的值
                const cloneInputList = flatToolInput(inputListRef.current);
                cloneInputList.forEach((item: any) => {
                  if (item.key === record.key) {
                    item.map_type = selectValue.map_type;
                    if (
                      record.defaultValue &&
                      record.map_type !== item.map_type &&
                      item.map_type === 'fixedValue' &&
                      !item.map_value
                    ) {
                      item.map_value = record.defaultValue;
                    } else {
                      item.map_value = selectValue.map_value || undefined;
                    }
                  }
                });
                const newInputList = removeChildren(adTreeUtils.flatToTreeData(cloneInputList));
                setInputList(newInputList);
              }}
              disabled={disabled || !record.enable}
              toolType={tool.tool_type}
            />
          </ErrorTip>
        );
      },
    },
  ];

  const outputColumns = [
    {
      title: intl.get('agentTool.paramName'),
      dataIndex: 'output_name',
      // width: 140,
      ellipsis: true,
      render: (value: string, record: any) => {
        return (
          <div title={value} style={{ display: 'inline-block' }}>
            <div>{value}</div>
            <div className="dip-c-subtext dip-font-12" title={record.output_desc}>
              {record.output_desc || intl.get('global.notDes')}
            </div>
          </div>
        );
      },
    },
    {
      title: intl.get('agentTool.paramType'),
      dataIndex: 'output_type',
      width: 100,
      ellipsis: true,
    },
  ];

  const title = tool?.details?.name || tool?.tool_name || '';

  return (
    <UniversalModal
      width={1200}
      open
      centered
      title={`${title} ${intl.get('agentCommonConfig.llm.parameterSetting')}`}
      className="ToolParamModal"
      onCancel={onClose}
      footerData={
        disabled
          ? false
          : [
              { label: intl.get('global.ok'), type: 'primary', onHandle: onOk },
              { label: intl.get('global.cancel'), onHandle: () => onClose() },
            ]
      }
    >
      {loading ? (
        <LoadingMask loading />
      ) : (
        <div>
          <div className="dip-flex-space-between">
            <div className="dip-border dip-flex-align-center" style={{ borderRadius: 2 }}>
              <span
                className={classNames(
                  {
                    'dip-c-primary': showType === ShowTypeEnum.Inputs,
                    'dip-bg-selected': showType === ShowTypeEnum.Inputs,
                  },
                  'dip-pointer dip-flex-center dip-m-4'
                )}
                style={{ padding: '0 22px', height: 32 }}
                onClick={() => setShowType(ShowTypeEnum.Inputs)}
              >
                {intl.get('agentCommonConfig.llm.inputParameters')}
              </span>
              <Divider className="dip-ml-0 dip-mr-0" type="vertical" />
              <span
                className={classNames(
                  {
                    'dip-c-primary': showType === ShowTypeEnum.Outputs,
                    'dip-bg-selected': showType === ShowTypeEnum.Outputs,
                  },
                  'dip-pointer dip-flex-center dip-m-4'
                )}
                style={{ padding: '0 22px', height: 32 }}
                onClick={() => setShowType(ShowTypeEnum.Outputs)}
              >
                {intl.get('agentCommonConfig.llm.outputParameters')}
              </span>
              {tool?.tool_type === 'tool' && (
                <>
                  <Divider className="dip-ml-0 dip-mr-0" type="vertical" />
                  <span
                    className={classNames(
                      {
                        'dip-c-primary': showType === ShowTypeEnum.Result,
                        'dip-bg-selected': showType === ShowTypeEnum.Result,
                      },
                      'dip-pointer dip-flex-center dip-m-4'
                    )}
                    style={{ padding: '0 22px', height: 32 }}
                    onClick={() => setShowType(ShowTypeEnum.Result)}
                  >
                    {intl.get('dataAgent.resultHandlingStrategy')}
                  </span>
                </>
              )}
              <Divider className="dip-ml-0 dip-mr-0" type="vertical" />
              <span
                className={classNames(
                  {
                    'dip-c-primary': showType === ShowTypeEnum.Timeout,
                    'dip-bg-selected': showType === ShowTypeEnum.Timeout,
                  },
                  'dip-pointer dip-flex-center dip-m-4'
                )}
                style={{ padding: '0 22px', height: 32 }}
                onClick={() => setShowType(ShowTypeEnum.Timeout)}
              >
                {intl.get('agentCommonConfig.llm.timeoutParameters')}
              </span>
            </div>
            <div>
              <div className="dip-flex-align-center">
                {disabledIntervention || readonly ? (
                  <Tooltip title={readonly ? '' : intl.get('agentCommonConfig.llm.disabledInterventionTip')}>
                    <Switch
                      size="small"
                      checked={intervention}
                      onChange={checked => {
                        if (checked) {
                          setTempInterventionMessage(interventionConfirmationMessage || '是否确认参数并继续执行?');
                          setInterventionModalVisible(true);
                        } else {
                          setIntervention(false);
                          setInterventionConfirmationMessage(null);
                        }
                      }}
                      disabled
                    />
                  </Tooltip>
                ) : (
                  <Switch
                    size="small"
                    checked={intervention}
                    onChange={checked => {
                      if (checked) {
                        // 打开确认 Modal，设置默认值
                        setTempInterventionMessage(interventionConfirmationMessage || '是否确认参数并继续执行?');
                        setInterventionModalVisible(true);
                      } else {
                        // 关闭时清空确认消息
                        setIntervention(false);
                        setInterventionConfirmationMessage(null);
                      }
                    }}
                  />
                )}
                <span className="dip-ml-12">{intl.get('agentCommonConfig.llm.intervention')}</span>
              </div>
              {intervention && (
                <div className="dip-mt-4">
                  <DipButton
                    icon={<DipIcon type="icon-dip-bianji" />}
                    size="small"
                    type="link"
                    onClick={() => {
                      setTempInterventionMessage(interventionConfirmationMessage || '是否确认参数并继续执行?');
                      setInterventionModalVisible(true);
                    }}
                  >
                    编辑确认提示
                  </DipButton>
                </div>
              )}
            </div>
          </div>
          <div className="dip-mt-20" style={isAgent ? { maxHeight: 238, minHeight: 95 } : { height: 368 }}>
            {showType === ShowTypeEnum.Outputs ? (
              <ADTable
                className={classNames('ToolParamTable', {
                  'tool-input-param-modal-no-data': isAgent && outputList?.length === 0,
                })}
                showHeader={false}
                rowKey={'key'}
                columns={outputColumns}
                dataSource={outputList}
                pagination={false}
                scroll={{ y: isAgent ? 200 : 330 }}
                expandable={{
                  expandRowByClick: true,
                }}
              />
            ) : showType === ShowTypeEnum.Inputs ? (
              <ADTable
                className={classNames('ToolParamTable', {
                  'tool-input-param-modal-no-data': isAgent && inputList?.length === 0,
                })}
                showHeader={false}
                rowKey={'key'}
                columns={inputColumns}
                dataSource={inputList}
                pagination={false}
                scroll={{ y: isAgent ? 200 : 330 }}
                expandable={{
                  expandRowByClick: true,
                  expandedRowKeys: expandedRowKeys,
                  onExpandedRowsChange: expandedRows => {
                    setExpandedRowKeys(expandedRows);
                  },
                  expandIconColumnIndex: 0,
                }}
              />
            ) : showType === ShowTypeEnum.Result ? (
              <ResultHandlingStrategy
                readonly={readonly}
                defaultResultProcessStrategies={resultProcessStrategiesRef.current}
                onUpdateResultProcessStrategies={param => {
                  resultProcessStrategiesRef.current = param;
                }}
              />
            ) : (
              <div className="dip-flex-align-center">
                <InputNumber
                  changeOnWheel
                  min={0}
                  max={86400}
                  value={timeout}
                  onChange={value => {
                    setTimeout(value ?? 0);
                  }}
                  precision={0}
                />
                <span className="dip-ml-8">秒</span>
              </div>
            )}
          </div>

          {isAgent && (
            <AgentDataSourceSettings
              readonly={readonly}
              dataSourceConfig={tool?.data_source_config}
              llmConfig={tool?.llm_config}
              onUpdateDataSourceConfig={dataSourceConfig => {
                dataSourceConfigRef.current = dataSourceConfig;
              }}
              onUpdateLLMConfig={llmConfig => {
                llmConfigRef.current = llmConfig;
              }}
            />
          )}
        </div>
      )}

      {/* intervention 确认提示配置 Modal */}
      <DipModal
        title="确认提示配置"
        open={interventionModalVisible}
        onCancel={() => {
          setInterventionModalVisible(false);
          setTempInterventionMessage('');
        }}
        onOk={() => {
          setIntervention(true);
          setInterventionConfirmationMessage(tempInterventionMessage || null);
          setInterventionModalVisible(false);
        }}
        okText="确认"
        cancelText="取消"
        destroyOnHidden
      >
        <div className="dip-mb-8">确认提示文案</div>
        <Input.TextArea
          autoSize={{ minRows: 2, maxRows: 6 }}
          value={tempInterventionMessage}
          onChange={e => setTempInterventionMessage(e.target.value)}
        />
        {/*<AdPromptInput*/}
        {/*  style={{ minHeight: 80 }}*/}
        {/*  placeholder="请输入确认提示文案"*/}
        {/*  value={tempInterventionMessage}*/}
        {/*  onChange={(val: string) => setTempInterventionMessage(val)}*/}
        {/*  trigger={[*/}
        {/*    {*/}
        {/*      character: '$',*/}
        {/*      options: varOptions.map((field: any) => ({*/}
        {/*        ...field,*/}
        {/*        value: `$${field.value}`, // 使用{{field.name}}格式*/}
        {/*        // type: 'text',*/}
        {/*      })),*/}
        {/*    },*/}
        {/*  ]}*/}
        {/*/>*/}
        {/*<div className="dip-font-12 dip-text-color-45 dip-mt-8">输入 $ 可快速插入变量</div>*/}
      </DipModal>
    </UniversalModal>
  );
};

export default ({ open, ...restProps }: any) => {
  return open && <ToolInputParamModal {...restProps} />;
};
