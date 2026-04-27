/**
 * Health API 实现
 */
import { API_ENDPOINTS } from '@/constants/api';
import { get, post } from '@/utils/http/request';
import type { HealthResponse } from '@/types/api';

/**
 * 健康检查
 */
export function healthCheck(): Promise<HealthResponse> {
  return get<HealthResponse>(API_ENDPOINTS.HEALTH);
}

/**
 * 详细健康检查
 */
export function detailedHealthCheck(): Promise<Record<string, unknown>> {
  return get<Record<string, unknown>>(API_ENDPOINTS.HEALTH_DETAILED);
}

/**
 * 手动触发状态同步
 */
export function triggerStateSync(): Promise<Record<string, unknown>> {
  return post<Record<string, unknown>>(API_ENDPOINTS.HEALTH + '/sync', {});
}
