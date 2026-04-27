/**
 * Session API 实现
 */
import { API_ENDPOINTS, LONG_TASK_HTTP_TIMEOUT } from '@/constants/api';
import { get, post, del } from '@/utils/http/request';
import type {
  SessionResponse,
  CreateSessionRequest,
  InstallSessionDependenciesRequest,
  SessionListResponse,
  ListSessionsParams,
} from './types';

/**
 * 获取会话列表
 */
export function listSessions(params?: ListSessionsParams): Promise<SessionListResponse> {
  return get<SessionListResponse>(API_ENDPOINTS.SESSIONS, { params });
}

/**
 * 获取会话详情
 */
export function getSession(sessionId: string): Promise<SessionResponse> {
  return get<SessionResponse>(API_ENDPOINTS.SESSION(sessionId));
}

/**
 * 创建会话
 */
export function createSession(data: CreateSessionRequest): Promise<SessionResponse> {
  return post<SessionResponse>(API_ENDPOINTS.SESSIONS, data, {
    timeout: LONG_TASK_HTTP_TIMEOUT,
  });
}

/**
 * 安装会话依赖
 */
export function installSessionDependencies(
  sessionId: string,
  data: InstallSessionDependenciesRequest,
): Promise<SessionResponse> {
  return post<SessionResponse>(API_ENDPOINTS.INSTALL_SESSION_DEPENDENCIES(sessionId), data, {
    timeout: LONG_TASK_HTTP_TIMEOUT,
  });
}

/**
 * 终止会话
 */
export function terminateSession(sessionId: string): Promise<SessionResponse> {
  return del<SessionResponse>(API_ENDPOINTS.SESSION(sessionId));
}
