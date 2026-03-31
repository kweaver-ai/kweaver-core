import React, { useState, useEffect } from 'react';
import intl from 'react-intl-universal';
import { Select, Slider, InputNumber, Row, Col, Typography, Spin, message } from 'antd';
import { get } from '@/utils/http';

const { Option } = Select;
const { Text } = Typography;

const modelBaseUrl = '/api/agent-factory/v3/model';

interface ModelConfigProps {
  llmConfig: any;
  onUpdateSelectModel: (name: string, data: any) => void;
  getError?: () => string | undefined;
  disabled?: boolean;
  defaultSelectedModel?: string;
}

interface ModelType {
  model_id: string;
  model_name: string;
  model: string;
  icon?: string;
  model_para?: {
    temperature?: [number, number, number];
    top_p?: [number, number, number];
    top_k?: [number, number, number];
    frequency_penalty?: [number, number, number];
    presence_penalty?: [number, number, number];
    max_tokens?: [number, number, number];
  };
}

const ModelConfig: React.FC<ModelConfigProps> = ({
  llmConfig,
  onUpdateSelectModel,
  getError,
  disabled = false,
  defaultSelectedModel = 'gpt-4',
}) => {
  const [loading, setLoading] = useState(false);
  const [modelList, setModelList] = useState<ModelType[]>([]);
  const [selectedModel, setSelectedModel] = useState<ModelType | null>(null);

  // 获取模型列表
  const fetchModelList = async () => {
    try {
      setLoading(true);
      const response = await get(`${modelBaseUrl}/list`);
      if (response && Array.isArray(response)) {
        setModelList(response);
        // 查找默认模型或已配置的模型
        const modelName = llmConfig?.llm_model_name || defaultSelectedModel;
        const model = response.find(m => m.model === modelName) || response[0];
        if (model) {
          setSelectedModel(model);
          if (!llmConfig) {
            handleModelChange(model.model);
          }
        }
      }
    } catch (error) {
      console.error('获取模型列表失败：', error);
      message.error(intl.get('dataAgent.fetchModelListFail'));
    } finally {
      setLoading(false);
    }
  };

  // 初始化时获取模型列表
  useEffect(() => {
    fetchModelList();
  }, []);

  // 处理模型选择变化
  const handleModelChange = (model: string) => {
    const selectedModel = modelList.find(m => m.model === model);
    if (!selectedModel) return;

    setSelectedModel(selectedModel);

    const params = selectedModel.model_para || {};
    const newConfig = {
      llm_id: selectedModel.model_id,
      llm_model_name: selectedModel.model_name,
      llm_icon: selectedModel.icon,
      temperature: params.temperature?.[2] || 0.7,
      top_p: params.top_p?.[2] || 0.9,
      top_k: params.top_k?.[2] || 40,
      frequency_penalty: params.frequency_penalty?.[2] || 0.9,
      presence_penalty: params.presence_penalty?.[2] || 0.9,
      max_tokens: params.max_tokens?.[2] || 1000,
    };

    onUpdateSelectModel(model, newConfig);
  };

  // 更新参数
  const updateParameter = (paramName: string, value: number) => {
    if (!llmConfig) return;

    const newConfig = { ...llmConfig, [paramName]: value };
    onUpdateSelectModel(selectedModel?.model || '', newConfig);
  };

  // 获取参数范围
  const getParameterRange = (paramName: string) => {
    if (!selectedModel || !selectedModel.model_para || !selectedModel.model_para[paramName]) {
      // 默认范围
      if (paramName === 'temperature') return [0, 2, 0.7];
      if (paramName === 'top_p') return [0, 1, 0.9];
      if (paramName === 'top_k') return [1, 100, 40];
      if (paramName.includes('penalty')) return [0, 2, 0.9];
      if (paramName === 'max_tokens') return [100, 8000, 1000];
      return [0, 1, 0.5];
    }
    return selectedModel.model_para[paramName];
  };

  // 渲染参数控制器
  const renderParameterControl = (label: string, paramName: string, step: number = 0.1, precision: number = 1) => {
    if (!selectedModel) return null;

    const [min, max, defaultValue] = getParameterRange(paramName);
    const value = llmConfig?.[paramName] ?? defaultValue;

    return (
      <div style={{ marginBottom: 16 }}>
        <Text>{label}</Text>
        <Row gutter={16} align="middle">
          <Col span={16}>
            <Slider
              min={min}
              max={max}
              step={step}
              value={value}
              onChange={val => updateParameter(paramName, val)}
              disabled={disabled}
            />
          </Col>
          <Col span={8}>
            <InputNumber
              min={min}
              max={max}
              step={step}
              value={value}
              onChange={val => val !== null && updateParameter(paramName, val)}
              style={{ width: '100%' }}
              disabled={disabled}
              precision={precision}
            />
          </Col>
        </Row>
      </div>
    );
  };

  const errorMessage = getError?.();

  return (
    <Spin spinning={loading}>
      <div>
        <div style={{ marginBottom: 16 }}>
          <div style={{ marginBottom: 8 }}>
            <Text>{intl.get('dataAgent.largeLanguageModel')}</Text>
            {errorMessage && (
              <Text type="danger" style={{ marginLeft: 8 }}>
                {errorMessage}
              </Text>
            )}
          </div>
          <Select
            style={{ width: '100%' }}
            placeholder={intl.get('dataAgent.pleaseSelectModel')}
            value={llmConfig?.llm_model_name || selectedModel?.model_name}
            onChange={handleModelChange}
            disabled={disabled}
          >
            {modelList.map(model => (
              <Option key={model.model_id} value={model.model}>
                {model.model_name}
              </Option>
            ))}
          </Select>
        </div>

        {selectedModel && (
          <div>
            {renderParameterControl('Temperature', 'temperature', 0.1, 1)}
            {renderParameterControl('Top P', 'top_p', 0.01, 2)}
            {renderParameterControl('Top K', 'top_k', 1, 0)}
            {renderParameterControl('Frequency Penalty', 'frequency_penalty', 0.1, 1)}
            {renderParameterControl('Presence Penalty', 'presence_penalty', 0.1, 1)}
            {renderParameterControl('Max Tokens', 'max_tokens', 100, 0)}
          </div>
        )}
      </div>
    </Spin>
  );
};

export default ModelConfig;
