import { get, post } from '@/utils/http';

/**
 * 获取大语言模型列表
 * @param params 分页参数
 * @returns Promise 返回模型列表数据
 */
export const getModelList = (params: { page: number; size: number }) => {
  return get('/api/mf-model-manager/v1/llm/list', { params });
};

/**
 * 测试模型连接状态
 */
export const testModelConnection = (params: {
  model_id?: string;
  model_series?: string;
  model_config?: Record<string, any>;
}) => {
  return post('/api/mf-model-manager/v1/llm/test', { body: params });
};
