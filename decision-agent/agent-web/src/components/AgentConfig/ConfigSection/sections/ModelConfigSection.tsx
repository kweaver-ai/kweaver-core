import React, { useState, useEffect, useRef } from 'react';
import intl from 'react-intl-universal';
import classNames from 'classnames';
import { Button, Select, Spin, message, Radio } from 'antd';
import { PlusOutlined, SettingOutlined } from '@ant-design/icons';
import { useAgentConfig } from '../../AgentConfigContext';
import { getModelList, testModelConnection } from '@/apis/model-manager';
import { LLMConfig } from '@/apis/agent-factory/type';
import LlmIcon from '@/assets/icons/model.svg';
import DipIcon from '@/components/DipIcon';
import SectionPanel from '../../common/SectionPanel';
import ModelSettingsPopover, { ModelSettings } from '../ModelSettingsPopover';
import styles from '../ConfigSection.module.less';

interface SelectOption {
  value: string;
  label: string;
  type: 'llm' | 'rlm';
}

interface Model {
  is_default: boolean;
  llm_config: LLMConfig;
}

// 默认模型配置常量
const DEFAULT_MODEL_CONFIG: Omit<LLMConfig, 'id' | 'name'> = {
  model_type: 'llm',
  temperature: 1,
  top_p: 1,
  top_k: 1,
  frequency_penalty: 0,
  presence_penalty: 0,
  max_tokens: 1000,
  retrieval_max_tokens: 32,
};

const ModelConfigSection: React.FC = () => {
  const { state, actions } = useAgentConfig();
  const [selectedModels, setSelectedModels] = useState<Model[]>(state.config.llms || []);
  const [isLoadingModels, setIsLoadingModels] = useState<boolean>(false);
  const [testingModel, setTestingModel] = useState<string | null>(null);
  const [modelOptions, setModelOptions] = useState<SelectOption[]>([]);
  const [isExpanded, setIsExpanded] = useState(false);

  // 引用模型设置按钮元素
  const modelRefs = useRef<(HTMLElement | null)[]>([]);

  // --------- 模型状态管理函数 ---------

  // 更新模型状态（本地状态和上下文状态）
  const updateModelsState = (models: Model[]) => {
    // 更新本地状态
    setSelectedModels(models);

    // 更新上下文状态
    actions.updateModels(models);

    return models;
  };

  // 获取可添加的模型选项
  const getAvailableModelOptions = () => {
    return modelOptions.filter(option => !selectedModels.some(model => model.llm_config?.id === option.value));
  };

  // --------- 模型操作函数 ---------

  // 测试选定的模型
  const testModel = async (modelId: string) => {
    setTestingModel(modelId);
    const tip = intl.get('dataAgent.config.testingConnection', {
      name: modelOptions.find(option => option.value === modelId)?.label,
    });
    const hide = message.loading(tip, 0);

    try {
      await testModelConnection({ model_id: modelId });
      hide();
      message.success(intl.get('dataAgent.config.modelConnectionTestSuccess'));
      setTestingModel(null);
    } catch (error) {
      hide();
      message.error(intl.get('dataAgent.config.modelConnectionTestFailed'));
      setTestingModel(null);
    }
  };

  // 添加新模型
  const handleAddModel = () => {
    if (modelOptions.length === 0) {
      message.warning(intl.get('dataAgent.config.noAvailableModel'));
      return;
    }

    // 找出未添加的模型选项
    const availableModels = getAvailableModelOptions();

    if (availableModels.length === 0) {
      message.warning(intl.get('dataAgent.config.allModelsAdded'));
      return;
    }

    // 添加第一个未添加的模型
    const modelToAdd = availableModels[0];
    const newModel: Model = {
      is_default: false,
      llm_config: {
        ...DEFAULT_MODEL_CONFIG,
        id: modelToAdd.value,
        name: modelToAdd.label,
        model_type: modelToAdd.type as 'llm' | 'rlm',
      },
    };

    // 添加新模型并更新状态
    const updatedModels = [...selectedModels, newModel];
    updateModelsState(updatedModels);
    setIsExpanded(true);
  };

  // 删除指定模型
  const handleDeleteModel = (id: string) => {
    const modelToDelete = selectedModels.find(model => model.llm_config?.id === id);
    const updatedModels = selectedModels.filter(model => model.llm_config?.id !== id);

    // 如果删除的是默认模型，将第一个模型设为默认
    if (modelToDelete?.is_default) {
      updatedModels[0].is_default = true;
    }

    updateModelsState(updatedModels);
  };

  // 设置默认模型
  const handleSetDefaultModel = (modelId: string) => {
    // 确保每个模型都设置了正确的default状态
    const updatedModels = selectedModels.map(model => ({
      ...model,
      is_default: model.llm_config?.id === modelId,
    }));

    updateModelsState(updatedModels);
    message.success(intl.get('dataAgent.config.defaultModelSet'));
  };

  // 更新模型选择
  const handleModelChange = (modelId: string, newModelId: string) => {
    // 获取新模型的名称 (必然存在于选项中)
    const newModelName = modelOptions.find(option => option.value === newModelId)?.label as string;

    const updatedModels = selectedModels.map(model =>
      model.llm_config.id === modelId
        ? {
            ...model,
            llm_config: {
              ...DEFAULT_MODEL_CONFIG,
              id: newModelId,
              name: newModelName,
              model_type: modelOptions.find(option => option.value === newModelId)?.type,
            },
          }
        : model
    );

    updateModelsState(updatedModels);
    message.success(intl.get('dataAgent.config.modelUpdatedAndReset'));
  };

  // 更新模型设置
  const handleModelSettingsUpdate = (modelId: string, settings: ModelSettings) => {
    // 使用map直接更新匹配ID的模型设置
    const updatedModels = selectedModels.map(model =>
      model.llm_config.id === modelId
        ? {
            ...model,
            llm_config: {
              ...model.llm_config,
              ...settings, // 只覆盖用户修改的设置
            },
          }
        : model
    );

    updateModelsState(updatedModels);
  };

  // 组件挂载时获取模型
  useEffect(() => {
    // 获取模型列表
    const fetchModels = async () => {
      setIsLoadingModels(true);
      try {
        const { data } = await getModelList({ page: 1, size: 100 });
        if (Array.isArray(data)) {
          // 处理获取到的模型数据
          // 这里要过滤掉vu类型的模型
          handleFetchedModels(data?.filter(model => model.model_type !== 'vu'));
        } else {
          message.error(intl.get('dataAgent.config.fetchModelListFailed'));
        }
      } catch {
        message.error(intl.get('dataAgent.config.fetchModelListFailed'));
      } finally {
        setIsLoadingModels(false);
      }
    };

    // 处理获取到的模型数据
    const handleFetchedModels = (data: any[]) => {
      // 1. 转换API响应为SelectOption格式
      const options = data.map((model: any) => ({
        value: model.model_id,
        label: model.model_name,
        type: model.model_type,
      }));
      setModelOptions(options);

      // 2. 如果没有选择模型但从API收到了模型，选择第一个
      if (selectedModels.length === 0 && data.length > 0) {
        const firstModelConfig: Model = {
          is_default: true,
          llm_config: {
            ...DEFAULT_MODEL_CONFIG,
            id: data[0].model_id,
            name: data[0].model_name,
            model_type: data[0].model_type,
          },
        };

        updateModelsState([firstModelConfig]);
        actions.resetDirtyState();
      }

      // 3. 如果选择了模型，则根据接口获取的列表更新已选模型信息
      if (selectedModels.length > 0) {
        const validModels = selectedModels.filter(model => data.some(item => item.model_id === model.llm_config.id));

        if (validModels.length) {
          const updatedModels = validModels.map(model => ({
            ...model,
            llm_config: {
              ...model.llm_config,
              name: options.find(option => option.value === model.llm_config.id)?.label,
            },
          }));

          updateModelsState(updatedModels);
          actions.resetDirtyState();
        } else {
          // 选中的模型不存在了，增加标识 invalid: true
          const updatedModels = selectedModels.map(model => ({
            ...model,
            llm_config: {
              ...model.llm_config,
              invalid: true,
            },
          }));

          updateModelsState(updatedModels);
          actions.resetDirtyState();
        }
      }
    };

    fetchModels();
  }, []); // eslint-disable-line

  return (
    <SectionPanel
      title={intl.get('dataAgent.config.defaultModelConfig')}
      description={intl.get('dataAgent.config.setDefaultLLMForAgent')}
      icon={<LlmIcon />}
      className="dip-border-line-b"
      isExpanded={isExpanded}
      onToggle={() => setIsExpanded(!isExpanded)}
      rightElement={
        <Button icon={<PlusOutlined />} type="text" onClick={handleAddModel} className="dip-c-link-75">
          {intl.get('dataAgent.config.add')}
        </Button>
      }
    >
      <div className={styles['model-config']}>
        {isLoadingModels ? (
          <div className={styles['loading-models']}>
            <Spin />
            <span>{intl.get('dataAgent.config.loadingModel')}</span>
          </div>
        ) : (
          <>
            {selectedModels.length > 0 && (
              <div className={styles['models-list']}>
                {selectedModels.map((model, index) => (
                  <div key={model.llm_config.id} className={classNames(styles['model-item'], 'dip-border-b')}>
                    <div className={styles['model-content']}>
                      <div className="dip-overflow-hidden dip-flex-align-center dip-flex-item-full-width dip-mr-8">
                        <Radio checked={model.is_default} onChange={() => handleSetDefaultModel(model.llm_config.id)} />
                        <span className="dip-1-line dip-mr-8 dip-ml-4">{intl.get('dataAgent.config.default')}</span>
                        <Select
                          style={{ width: 300, maxWidth: 'calc(100% - 66px)' }}
                          value={model.llm_config.invalid ? '---' : model.llm_config.id}
                          status={model.llm_config.invalid ? 'error' : ''}
                          defaultActiveFirstOption={false}
                          onChange={value => handleModelChange(model.llm_config.id, value)}
                        >
                          {/* 只显示自己或未添加的模型 */}
                          {modelOptions
                            .filter(
                              option =>
                                option.value === model.llm_config.id ||
                                !selectedModels.some(m => m.llm_config.id === option.value)
                            )
                            .map(option => (
                              <Select.Option key={option.value} value={option.value}>
                                {option.label}
                              </Select.Option>
                            ))}
                        </Select>
                      </div>
                      <div className={styles['model-actions']}>
                        {Boolean(model.llm_config.id) && (
                          <>
                            <Button
                              type="text"
                              onClick={() => testModel(model.llm_config.id)}
                              className={styles['test-button']}
                              loading={testingModel === model.llm_config.id}
                            >
                              {intl.get('dataAgent.config.test')}
                            </Button>
                            <ModelSettingsPopover
                              initialSettings={model.llm_config}
                              onSettingsChange={settings => handleModelSettingsUpdate(model.llm_config.id, settings)}
                            >
                              <SettingOutlined
                                className="dip-c-subtext"
                                size={20}
                                ref={el => (modelRefs.current[index] = el)}
                              />
                            </ModelSettingsPopover>
                          </>
                        )}
                        {selectedModels.length > 1 && (
                          <Button
                            size="small"
                            type="text"
                            className="dip-c-subtext"
                            onClick={() => handleDeleteModel(model.llm_config.id)}
                            icon={<DipIcon type="icon-dip-trash" />}
                          />
                        )}
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </>
        )}
      </div>
    </SectionPanel>
  );
};

export default ModelConfigSection;
