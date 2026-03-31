import React, { useState, useEffect, useMemo } from 'react';
import intl from 'react-intl-universal';
import { Table, Input, Select, Button, message, Tooltip } from 'antd';
import { PlusOutlined, QuestionCircleOutlined, SettingOutlined } from '@ant-design/icons';
import { uniqBy } from 'lodash';
import InputIcon from '@/assets/icons/input.svg';
import CollapseArrow from '@/assets/icons/collapse-arrow.svg';
import { useAgentConfig } from '../../AgentConfigContext';
import DipIcon from '@/components/DipIcon';
import SectionPanel from '../../common/SectionPanel';
import FileSettingsModal from '../FileSettingsModal';
import { defaultTempZoneConfig, getInputTypes } from './constants';
import styles from './InputConfig.module.less';

const { Option } = Select;

interface InputField {
  name: string;
  type: string;
  is_built_in?: boolean;
}

// 临时区配置类型
interface TempZoneConfig {
  name: string;
  max_file_count: number;
  single_file_size_limit: number;
  single_file_size_limit_unit: string;
  support_data_type: string[];
  allowed_file_categories: string[];
  allowed_file_types: string[];
  tmp_file_use_type: string;
}
const getBuiltInVariableDescriptions = () => ({
  history: {
    title: 'history - 历史对话信息',
    content: `**${intl.get('dataAgent.config.variableMeaning')}** ${intl.get('dataAgent.config.chatHistory')}

**${intl.get('dataAgent.config.dataStructure')}** ${intl.get('dataAgent.config.historyDataStructure')}

**${intl.get('dataAgent.config.example')}**
\`\`\`json
[
  {
    "role": "user",
    "content": "你好，我想了解一下Python编程"
  },
  {
    "role": "assistant",
    "content": "你好！很高兴为您介绍Python编程。Python是一门非常受欢迎的编程语言，适合初学者。您想从哪个方面开始了解呢？"
  },
  {
    "role": "user",
    "content": "请详细介绍一下字符串类型，包括如何创建和操作字符串"
  }
]
\`\`\``,
  },
  tool: {
    title: 'tool - 工具执行参数',
    content: `**${intl.get('dataAgent.config.variableMeaning')}** ${intl.get('dataAgent.config.toolParamDes')}

**${intl.get('dataAgent.config.dataStructure')}** ${intl.get('dataAgent.config.toolDataStructure')}

**${intl.get('dataAgent.config.example')}**
\`\`\`json
{
  "session_id": "agent-session-1897901852071170048",
  "tool_name": "tool_name",
  "tool_args": [
    {
      "key": "param1",
      "value": "value1",
      "type": "string"
    },
    {
      "key": "param2",
      "value": "value2",
      "type": "string"
    }
  ]
}
\`\`\``,
  },
  header: {
    title: 'header - 请求头信息',
    content: `**${intl.get('dataAgent.config.variableMeaning')}** ${intl.get('dataAgent.config.headerParamDes')}

**${intl.get('dataAgent.config.dataStructure')}** ${intl.get('dataAgent.config.headerDataStructure')}

**${intl.get('dataAgent.config.example')}**
\`\`\`json
{
  "authorization": "Bearer xxxxx",
  "userid": "user_id"
}
\`\`\``,
  },
  self_config: {
    title: 'self_config - Agent配置信息',
    content: `**${intl.get('dataAgent.config.variableMeaning')}** ${intl.get('dataAgent.config.selfConfigDes')}

**${intl.get('dataAgent.config.dataStructure')}** ${intl.get('dataAgent.config.SelfConfigDataStructure')}

**${intl.get('dataAgent.config.example')}**
\`\`\`json
{
  "data_source": {
    "metric": [
      {
        "metric_model_id": "523210081068677639"
      }
    ],
    "kn_entry": [
      {
        "kn_entry_id": "d28q82kgpm5637kisosg"
      }
    ]
  }
}
\`\`\``,
  },
});

// 默认的query变量
const defaultQueryField: InputField = { name: 'query', type: 'string', is_built_in: true };
// 定义4个固定的对象类型变量
const fixedVariables: InputField[] = [
  { name: 'history', type: 'object', is_built_in: true },
  // { name: 'tool', type: 'object', is_built_in: true },
  { name: 'header', type: 'object', is_built_in: true },
  { name: 'self_config', type: 'object', is_built_in: true },
];
// 确保inputConfig包含固定变量和query字段
const initializeInputConfig = (originalConfig: any) => ({
  ...originalConfig,
  fields: uniqBy([defaultQueryField, ...fixedVariables, ...originalConfig.fields], 'name'),
});

const InputConfig: React.FC = () => {
  const { state, actions } = useAgentConfig();

  // 检查是否可编辑输入配置
  const canEditInputConfig = actions.canEditField('input_config');

  const [inputConfig, setInputConfig] = useState(() => initializeInputConfig(state.config?.input));
  const [fileSettingsVisible, setFileSettingsVisible] = useState(false);
  const [isExpanded, setIsExpanded] = useState(false);
  const [builtInVariablesExpanded, setBuiltInVariablesExpanded] = useState(false);

  // 获取要显示的字段列表
  const displayFields = useMemo(() => {
    const builtInFields = inputConfig.fields.filter(
      (field: InputField) => fixedVariables.find(v => v.name === field.name) && field.is_built_in
    );
    const userDefinedFields = inputConfig.fields.filter((field: InputField) => !field.is_built_in);

    // 默认只显示第一个内置变量，如果展开则显示所有内置变量
    const visibleBuiltInFields = builtInVariablesExpanded ? builtInFields : builtInFields.slice(0, 1);

    return [defaultQueryField, ...visibleBuiltInFields, ...userDefinedFields];
  }, [inputConfig.fields, builtInVariablesExpanded]);
  // 内置变量的说明信息
  const builtInVariableDescriptions = useMemo(getBuiltInVariableDescriptions, []);
  const inputTypes = useMemo(getInputTypes, []);

  // 确保在组件初始化时就将包含固定变量的完整字段列表同步到context中
  useEffect(() => {
    if (inputConfig.fields) {
      updateContext(inputConfig.fields);
    }
  }, []);

  // 创建新的输入字段
  const createNewInputField = (): InputField => {
    // 生成一个新的唯一名称，从query1开始
    let newName = 'query1';
    let counter = 1;

    // 找到一个不重复的名称
    while (inputConfig.fields.some((field: InputField) => field.name === newName)) {
      counter++;
      newName = `query${counter}`;
    }

    return {
      name: newName,
      type: 'string', // 默认类型为string
    };
  };

  // 添加新字段
  const handleAddField = () => {
    if (!canEditInputConfig) return;
    const newField = createNewInputField();
    const updatedFields = [...inputConfig.fields, newField];
    setInputConfig((prev: any) => ({
      ...prev,
      fields: updatedFields,
    }));
    updateContext(updatedFields);
    setIsExpanded(true);
  };

  // 更新字段名称
  const handleNameChange = (value: string, record: any, index: number) => {
    if (!canEditInputConfig) return;

    if (record.is_built_in) {
      message.warning(intl.get('dataAgent.cannotModifyFixedVariableName'));
      return;
    }

    // 检查是否有重复的名称
    if (inputConfig.fields.some((field: any) => field.name === value)) {
      message.warning(intl.get('dataAgent.variableNameExistsModifyBeforeSubmit'));
    }

    // 内置变量的展开/收起，之间的index差了 fixedVariables.length - 1
    const actualIndex = builtInVariablesExpanded ? index : index + (fixedVariables.length - 1);

    // 允许用户输入任何值（包括重复的名称），但会给出提示
    const updatedFields = inputConfig.fields.map((field: any, i: number) => {
      // 需要找到原始字段在完整列表中的索引
      return actualIndex === i ? { ...field, name: value } : field;
    });

    setInputConfig((prev: any) => ({
      ...prev,
      fields: updatedFields,
    }));

    updateContext(updatedFields);
  };

  // 更新字段类型
  const handleTypeChange = (value: string, _record: any, index: number) => {
    if (!canEditInputConfig) return;

    // 内置变量的展开/收起，之间的index差了 fixedVariables.length - 1
    const actualIndex = builtInVariablesExpanded ? index : index + (fixedVariables.length - 1);
    const updatedFields = inputConfig.fields.map((field: any, i: number) => {
      return i === actualIndex ? { ...field, type: value } : field;
    });

    setInputConfig((prev: any) => ({
      ...prev,
      fields: updatedFields,
    }));

    updateContext(updatedFields);
  };

  // 打开文件配置弹窗
  const openFileSettings = () => {
    if (!canEditInputConfig) return;
    setFileSettingsVisible(true);
  };

  // 保存文件配置
  const handleFileSettingsSave = (config: TempZoneConfig) => {
    const updatedConfig = {
      ...inputConfig,
      temp_zone_config: config,
    };

    setInputConfig((prev: any) => ({
      ...prev,
      temp_zone_config: config,
    }));

    actions.updateInputConfig(updatedConfig.fields, config);

    setFileSettingsVisible(false);
  };

  // 删除字段
  const handleDeleteField = (record: any, index: number) => {
    if (!canEditInputConfig) return;
    // 检查是否为固定变量，不允许删除固定变量
    if (record.is_built_in) {
      message.warning(intl.get('dataAgent.cannotDeleteFixedVariable'));
      return;
    }

    // 内置变量的展开/收起，之间的index差了 fixedVariables.length - 1
    const actualIndex = builtInVariablesExpanded ? index : index + (fixedVariables.length - 1);
    const updatedFields = inputConfig.fields.filter((_field: any, i: number) => i !== actualIndex);

    setInputConfig((prev: any) => ({
      ...prev,
      fields: updatedFields,
    }));

    updateContext(updatedFields);
  };

  // 更新上下文
  const updateContext = (fields: InputField[]) => {
    // 转换为API需要的格式，保留所有字段（包括固定变量），只保留name和type字段
    const allFields = fields.map(field => ({
      name: field.name,
      type: field.type,
    }));

    // 检查是否有file类型字段（包括固定变量和用户自定义字段）
    const hasFileType = fields.some((field: InputField) => field.type === 'file');

    // 临时区配置
    const tempZoneConfig = hasFileType ? inputConfig.temp_zone_config || defaultTempZoneConfig : null;

    // 使用上下文的更新方法，传入所有字段
    actions.updateInputConfig(allFields, tempZoneConfig);
  };

  // 检查指定选项是否应该被禁用
  const isOptionDisabled = (optionValue: string) => {
    return optionValue === 'file' && inputConfig.fields.some((field: InputField) => field.type === 'file');
  };

  // 检查是否应该显示展开按钮（在第一个内置变量的操作列）
  const shouldShowExpandButton = (record: any, index: number) => {
    const fixedVariableNames = fixedVariables.map(v => v.name);
    const isFirstBuiltIn = fixedVariableNames.includes(record.name) && index === 1;
    const hasMoreBuiltInVariables = fixedVariables.length > 1;
    return isFirstBuiltIn && hasMoreBuiltInVariables;
  };

  // 检查字段是否不可编辑
  const isFieldDisabled = (record: any) => {
    return !canEditInputConfig || record.is_built_in;
  };

  // 表格列配置
  const columns = [
    {
      title: intl.get('dataAgent.config.variableName'),
      dataIndex: 'name',
      key: 'name',
      width: '50%',
      render: (text: string, record: any, index: number) => {
        const isDisabled = isFieldDisabled(record);

        return (
          <div style={{ display: 'flex', alignItems: 'center' }}>
            {shouldShowExpandButton(record, index) && (
              <span
                style={{
                  // marginRight: '8px',
                  fontSize: '14px',
                  cursor: 'pointer',
                  color: '#666',
                  userSelect: 'none',
                  fontWeight: 'bold',
                  position: 'absolute',
                  left: '-8px',
                }}
                className={`${styles['input-config-arrow']} ${builtInVariablesExpanded ? '' : styles.collapsed}`}
                onClick={e => {
                  e.stopPropagation();
                  setBuiltInVariablesExpanded(!builtInVariablesExpanded);
                }}
              >
                <CollapseArrow />
              </span>
            )}
            <Input
              style={{ width: '100%' }}
              value={text}
              onChange={e => {
                const filteredValue = e.target.value.replace(/[^\x00-\xff]/g, '');
                handleNameChange(filteredValue, record, index);
              }}
              placeholder={intl.get('dataAgent.config.pleaseTypeVariableName')}
              disabled={isDisabled}
            />
          </div>
        );
      },
    },
    {
      title: intl.get('dataAgent.config.type'),
      dataIndex: 'type',
      key: 'type',
      width: 'calc(50% - 80px)',
      render: (text: string, record: any, index: number) => {
        const isDisabled = isFieldDisabled(record);

        return (
          <Select
            value={text}
            style={{ width: '100%' }}
            onChange={value => handleTypeChange(value, record, index)}
            disabled={isDisabled}
          >
            {inputTypes.map(option => (
              <Option key={option.value} value={option.value} disabled={isOptionDisabled(option.value)}>
                {option.label}
              </Option>
            ))}
          </Select>
        );
      },
    },
    {
      title: intl.get('dataAgent.config.operation'),
      key: 'action',
      width: '80px',
      render: (_: any, record: any, index: number) => {
        const fixedVariableNames = fixedVariables.map(v => v.name);
        const isFixedVariable = fixedVariableNames.includes(record.name) && record.is_built_in;
        const isQueryField = record.name === 'query' && record.is_built_in;

        return (
          <div className="dip-1-line">
            {record.type === 'file' && (
              <div style={{ display: 'inline-block', width: '24px', textAlign: 'center' }}>
                <SettingOutlined className="dip-c-subtext" size={20} onClick={() => openFileSettings()} />
              </div>
            )}
            {isFixedVariable && (
              <Tooltip
                title={
                  <div
                    style={{
                      maxWidth: '400px',
                      whiteSpace: 'pre-wrap',
                      maxHeight: '500px',
                      overflowY: 'auto',
                      marginRight: '-8px',
                      paddingRight: '8px',
                    }}
                  >
                    <div
                      dangerouslySetInnerHTML={{
                        __html: builtInVariableDescriptions[
                          record.name as keyof typeof builtInVariableDescriptions
                        ]?.content
                          .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
                          .replace(/`([^`]+)`/g, '<code>$1</code>'),
                      }}
                    />
                  </div>
                }
                placement="left"
                styles={{
                  root: {
                    maxWidth: '450px',
                  },
                  body: {
                    backgroundColor: '#fff',
                    color: '#333',
                    border: '1px solid #d9d9d9',
                    borderRadius: '6px',
                    boxShadow:
                      '0 6px 16px 0 rgba(0, 0, 0, 0.08), 0 3px 6px -4px rgba(0, 0, 0, 0.12), 0 9px 28px 8px rgba(0, 0, 0, 0.05)',
                  },
                }}
              >
                <QuestionCircleOutlined className="dip-c-subtext" style={{ marginLeft: '8px', cursor: 'pointer' }} />
              </Tooltip>
            )}
            {!isFixedVariable && !isQueryField && (
              <Button
                size="small"
                type="text"
                className={styles['delete-button']}
                onClick={() => handleDeleteField(record, index)}
                disabled={!canEditInputConfig}
                icon={<DipIcon type="icon-dip-trash" />}
              />
            )}
          </div>
        );
      },
    },
  ];

  return (
    <SectionPanel
      title={
        <>
          <div>{intl.get('dataAgent.config.inputConfig')}</div>
          <Tooltip title={intl.get('dataAgent.config.inputConfigTip')}>
            <QuestionCircleOutlined className="dip-font-14" />
          </Tooltip>
        </>
      }
      description={intl.get('dataAgent.config.inputConfigDescription')}
      icon={<InputIcon />}
      className="dip-border-line-b"
      isExpanded={isExpanded}
      onToggle={() => setIsExpanded(!isExpanded)}
      rightElement={
        <Button
          icon={<PlusOutlined />}
          type="text"
          onClick={handleAddField}
          disabled={!canEditInputConfig}
          className="dip-c-link-75"
        >
          {intl.get('dataAgent.config.add')}
        </Button>
      }
    >
      <div className={styles['input-config']}>
        <div>
          <div className={styles['input-config']}>
            <Table
              dataSource={displayFields}
              columns={columns}
              pagination={false}
              rowKey={(_record, index) => `${index}`}
              size="middle"
            />
            {fileSettingsVisible && (
              <FileSettingsModal
                isEditable={true}
                onCancel={() => setFileSettingsVisible(false)}
                onOk={handleFileSettingsSave}
                initialConfig={
                  inputConfig.temp_zone_config
                    ? {
                        ...defaultTempZoneConfig,
                        ...inputConfig.temp_zone_config,
                      }
                    : defaultTempZoneConfig
                }
              />
            )}
          </div>
        </div>
      </div>
    </SectionPanel>
  );
};

export default InputConfig;
