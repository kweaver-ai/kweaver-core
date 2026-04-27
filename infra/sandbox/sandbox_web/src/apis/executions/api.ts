/**
 * Execution API 实现
 */
import { API_ENDPOINTS, DEFAULT_PAGINATION } from '@/constants/api';
import { get, post } from '@/utils/http/request';
import type {
  ExecutionResponse,
  ExecuteCodeRequest,
  ExecuteCodeResponse,
  ExecutionListResponse,
} from './types';

/**
 * 提交代码执行
 */
export function executeCode(sessionId: string, data: ExecuteCodeRequest): Promise<ExecuteCodeResponse> {
  return post<ExecuteCodeResponse>(API_ENDPOINTS.EXECUTE(sessionId), data);
}

/**
 * 获取执行状态
 */
export function getExecutionStatus(executionId: string): Promise<ExecutionResponse> {
  return get<ExecutionResponse>(API_ENDPOINTS.EXECUTION_STATUS(executionId));
}

/**
 * 获取执行结果
 */
export function getExecutionResult(executionId: string): Promise<ExecutionResponse> {
  return get<ExecutionResponse>(API_ENDPOINTS.EXECUTION_RESULT(executionId));
}

/**
 * 获取会话的执行列表
 */
export function listSessionExecutions(
  sessionId: string,
  params?: { limit?: number; offset?: number }
): Promise<ExecutionListResponse> {
  return get<ExecutionListResponse>(API_ENDPOINTS.EXECUTIONS_BY_SESSION(sessionId), {
    params: { ...DEFAULT_PAGINATION, ...params },
  });
}
