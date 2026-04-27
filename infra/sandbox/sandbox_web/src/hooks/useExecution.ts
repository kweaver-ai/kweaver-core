/**
 * 执行管理 Hook
 */
import { useState, useCallback } from 'react';
import { message } from 'antd';
import * as executionsApi from '@apis/executions';
import type {
  ExecutionResponse,
  ExecuteCodeRequest,
  ExecuteCodeResponse,
} from '@apis/executions';

export function useExecution() {
  const [executions, setExecutions] = useState<ExecutionResponse[]>([]);
  const [loading, setLoading] = useState(false);
  const [currentExecution, setCurrentExecution] = useState<ExecutionResponse | null>(null);

  // 提交代码执行
  const executeCode = useCallback(async (
    sessionId: string,
    data: ExecuteCodeRequest
  ): Promise<ExecuteCodeResponse> => {
    setLoading(true);
    try {
      const result = await executionsApi.executeCode(sessionId, data);
      // 轮询获取执行结果
      pollExecutionResult(result.execution_id);
      return result;
    } catch (error) {
      message.error('代码执行提交失败');
      console.error(error);
      throw error;
    } finally {
      setLoading(false);
    }
  }, []);

  // 轮询执行结果
  const pollExecutionResult = useCallback(async (executionId: string) => {
    const maxAttempts = 60; // 最多轮询 60 次 (约 1 分钟)
    let attempts = 0;

    const poll = async () => {
      attempts++;
      try {
        const result = await executionsApi.getExecutionResult(executionId);
        setCurrentExecution(result);

        // 如果执行还在进行中，继续轮询
        if (
          result.status === 'pending' ||
          result.status === 'PENDING' ||
          result.status === 'running' ||
          result.status === 'RUNNING'
        ) {
          if (attempts < maxAttempts) {
            setTimeout(poll, 1000); // 1 秒后重试
          }
        } else {
          // 执行完成，添加到历史记录并清除当前执行
          setExecutions((prev) => {
            // 检查是否已存在，避免重复添加
            const exists = prev.some(e => e.id === result.id);
            if (exists) {
              return prev;
            }
            return [result, ...prev];
          });
          // 延迟清除 currentExecution，让用户看到最终状态
          setTimeout(() => setCurrentExecution(null), 100);
          if (result.status === 'completed' || result.status === 'COMPLETED') {
            message.success('代码执行成功');
          } else if (result.status === 'failed') {
            message.error('代码执行失败');
          } else if (result.status === 'timeout') {
            message.error('代码执行超时');
          }
        }
      } catch (error) {
        console.error('获取执行结果失败', error);
        if (attempts < maxAttempts) {
          setTimeout(poll, 1000);
        }
      }
    };

    poll();
  }, []);

  // 获取会话的执行列表
  const fetchSessionExecutions = useCallback(async (sessionId: string) => {
    setLoading(true);
    try {
      const result = await executionsApi.listSessionExecutions(sessionId);
      setExecutions(result.items);
    } catch (error) {
      message.error('获取执行历史失败');
      console.error(error);
    } finally {
      setLoading(false);
    }
  }, []);

  // 清空当前执行
  const clearCurrentExecution = useCallback(() => {
    setCurrentExecution(null);
  }, []);

  return {
    executions,
    currentExecution,
    loading,
    executeCode,
    fetchSessionExecutions,
    clearCurrentExecution,
  };
}
