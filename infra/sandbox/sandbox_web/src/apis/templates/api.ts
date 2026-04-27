/**
 * Template API 实现
 */
import { API_ENDPOINTS, DEFAULT_PAGINATION } from '@/constants/api';
import { get, post, put, del } from '@/utils/http/request';
import type {
  TemplateResponse,
  CreateTemplateRequest,
  UpdateTemplateRequest,
} from './types';

/**
 * 获取模板列表
 */
export function listTemplates(params?: { limit?: number; offset?: number }): Promise<TemplateResponse[]> {
  return get<TemplateResponse[]>(API_ENDPOINTS.TEMPLATES, {
    params: { ...DEFAULT_PAGINATION, ...params },
  });
}

/**
 * 获取模板详情
 */
export function getTemplate(templateId: string): Promise<TemplateResponse> {
  return get<TemplateResponse>(API_ENDPOINTS.TEMPLATE(templateId));
}

/**
 * 创建模板
 */
export function createTemplate(data: CreateTemplateRequest): Promise<TemplateResponse> {
  return post<TemplateResponse>(API_ENDPOINTS.TEMPLATES, data);
}

/**
 * 更新模板
 */
export function updateTemplate(templateId: string, data: UpdateTemplateRequest): Promise<TemplateResponse> {
  return put<TemplateResponse>(API_ENDPOINTS.TEMPLATE(templateId), data);
}

/**
 * 删除模板
 */
export function deleteTemplate(templateId: string): Promise<{ message: string }> {
  return del<{ message: string }>(API_ENDPOINTS.TEMPLATE(templateId));
}
