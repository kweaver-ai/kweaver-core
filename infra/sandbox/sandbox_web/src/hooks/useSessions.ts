/**
 * 会话管理 Hook
 */
import { useState, useCallback, useRef } from 'react';
import { message } from 'antd';
import * as sessionsApi from '@apis/sessions';
import type {
  SessionResponse,
  CreateSessionRequest,
  InstallSessionDependenciesRequest,
  SessionStatus,
} from '@apis/sessions';

/** 会话统计 */
export interface SessionStats {
  total: number;
  starting: number;
  running: number;
  terminated: number;
}

export function useSessions() {
  const [sessions, setSessions] = useState<SessionResponse[]>([]);
  const [loading, setLoading] = useState(false);
  const isLoadingRef = useRef(false);

  // 获取统计 - 支持小写状态值
  const stats: SessionStats = {
    total: sessions.length,
    starting: sessions.filter((s) =>
      s.status === 'CREATING' || s.status === 'creating' ||
      s.status === 'STARTING' || s.status === 'starting'
    ).length,
    running: sessions.filter((s) => s.status === 'RUNNING' || s.status === 'running').length,
    terminated: sessions.filter((s) =>
      s.status === 'TERMINATED' || s.status === 'terminated' ||
      s.status === 'COMPLETED' || s.status === 'completed'
    ).length,
  };

  // 获取会话列表
  const fetchSessions = useCallback(async () => {
    // 使用 ref 防止重复请求
    if (isLoadingRef.current) return;
    isLoadingRef.current = true;
    setLoading(true);
    try {
      const response = await sessionsApi.listSessions();
      setSessions(response.items);
    } catch (error) {
      message.error('获取会话列表失败');
      console.error(error);
    } finally {
      isLoadingRef.current = false;
      setLoading(false);
    }
  }, []);

  // 创建会话
  const createSession = useCallback(async (data: CreateSessionRequest) => {
    setLoading(true);
    try {
      const newSession = await sessionsApi.createSession(data);
      setSessions((prev) => [...prev, newSession]);
      message.success('会话创建成功');
      return newSession;
    } catch (error) {
      message.error('会话创建失败');
      console.error(error);
      throw error;
    } finally {
      setLoading(false);
    }
  }, []);

  // 终止会话
  const terminateSession = useCallback(async (id: string) => {
    setLoading(true);
    try {
      const terminated = await sessionsApi.terminateSession(id);
      setSessions((prev) =>
        prev.map((s) => (s.id === id ? terminated : s))
      );
      message.success('会话已终止');
      return terminated;
    } catch (error) {
      message.error('会话终止失败');
      console.error(error);
      throw error;
    } finally {
      setLoading(false);
    }
  }, []);

  const installSessionDependencies = useCallback(
    async (id: string, data: InstallSessionDependenciesRequest) => {
      setLoading(true);
      try {
        const updatedSession = await sessionsApi.installSessionDependencies(id, data);
        setSessions((prev) => prev.map((s) => (s.id === id ? updatedSession : s)));
        message.success('依赖安装任务已提交');
        return updatedSession;
      } catch (error) {
        message.error('安装依赖失败');
        console.error(error);
        throw error;
      } finally {
        setLoading(false);
      }
    },
    [],
  );

  return {
    sessions,
    stats,
    loading,
    fetchSessions,
    createSession,
    installSessionDependencies,
    terminateSession,
  };
}
