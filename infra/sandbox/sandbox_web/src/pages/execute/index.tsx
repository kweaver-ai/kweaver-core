/**
 * 代码执行页面
 */
import { useState, useEffect } from 'react';
import { useSearchParams } from 'react-router-dom';
import { Button, Select, Card, Tag, Space, Empty, Spin } from 'antd';
import { PlayCircleFilled, CaretRightOutlined } from '@ant-design/icons';
import Editor from '@monaco-editor/react';
import { useExecution } from '@hooks/useExecution';
import { useSessions } from '@hooks/useSessions';
import { EXECUTION_STATUS_LABELS } from '@constants/runtime';
import type { ExecuteCodeRequest, ExecutionResponse } from '@apis/executions';

// 示例代码
const DEFAULT_CODE = `def handler(event):
    name = event.get("name", "World")
    return {"message": f"Hello, {name}!"}
`;

const DEFAULT_EVENT = `{
  "name": "Sandbox Platform"
}`;

export default function ExecutePage() {
  const [searchParams] = useSearchParams();
  const { sessions, loading: sessionsLoading, fetchSessions } = useSessions();
  const [selectedSession, setSelectedSession] = useState<string>('');
  const [code, setCode] = useState(DEFAULT_CODE);
  const [eventData, setEventData] = useState(DEFAULT_EVENT);
  const [eventError, setEventError] = useState('');

  const { executions, currentExecution, loading, executeCode, fetchSessionExecutions } = useExecution();

  // 从 URL 参数获取 sessionId 并自动选中
  useEffect(() => {
    fetchSessions();
  }, [fetchSessions]);

  useEffect(() => {
    const sessionId = searchParams.get('sessionId');
    if (sessionId) {
      setSelectedSession(sessionId);
    } else if (sessions.length > 0) {
      // 过滤掉已终止的会话
      const activeSessions = sessions.filter(
        (s) => s.status !== 'terminated' && s.status !== 'TERMINATED'
      );
      if (activeSessions.length > 0) {
        // 默认选择第一个 running 状态的会话
        const firstRunning = activeSessions.find((s) => s.status === 'running' || s.status === 'RUNNING');
        setSelectedSession(firstRunning?.id || activeSessions[0].id);
      }
    }
  }, [searchParams, sessions]);

  // 当选择的会话改变时，加载该会话的执行历史
  useEffect(() => {
    if (selectedSession) {
      fetchSessionExecutions(selectedSession);
    }
  }, [selectedSession, fetchSessionExecutions]);

  // 执行代码
  const handleExecute = async () => {
    // 验证 Event 数据是否为有效 JSON
    try {
      JSON.parse(eventData);
      setEventError('');
    } catch {
      setEventError('Event 数据必须是有效的 JSON 格式');
      return;
    }

    const request: ExecuteCodeRequest = {
      code,
      language: 'python',
      event: JSON.parse(eventData),
      timeout: 30,
    };

    await executeCode(selectedSession, request);
  };

  // 状态配置 - 支持小写状态值（API 返回小写）
  const getStatusConfig = (status: string) => {
    const configs: Record<
      string,
      { color: string; icon: string; label: string }
    > = {
      PENDING: { color: 'warning', icon: '⏱', label: EXECUTION_STATUS_LABELS.PENDING },
      RUNNING: { color: 'processing', icon: '⚡', label: EXECUTION_STATUS_LABELS.RUNNING },
      COMPLETED: { color: 'success', icon: '✓', label: EXECUTION_STATUS_LABELS.COMPLETED },
      FAILED: { color: 'error', icon: '✗', label: EXECUTION_STATUS_LABELS.FAILED },
      TIMEOUT: { color: 'error', icon: '⏱', label: EXECUTION_STATUS_LABELS.TIMEOUT },
      CRASHED: { color: 'error', icon: '💥', label: EXECUTION_STATUS_LABELS.CRASHED },
      // 支持小写（API 返回）
      pending: { color: 'warning', icon: '⏱', label: EXECUTION_STATUS_LABELS.PENDING },
      running: { color: 'processing', icon: '⚡', label: EXECUTION_STATUS_LABELS.RUNNING },
      completed: { color: 'success', icon: '✓', label: EXECUTION_STATUS_LABELS.COMPLETED },
      failed: { color: 'error', icon: '✗', label: EXECUTION_STATUS_LABELS.FAILED },
      timeout: { color: 'error', icon: '⏱', label: EXECUTION_STATUS_LABELS.TIMEOUT },
      crashed: { color: 'error', icon: '💥', label: EXECUTION_STATUS_LABELS.CRASHED },
    };
    return configs[status] || configs.PENDING;
  };

  return (
    <>
      {/* 页面标题 */}
      <div style={{ marginBottom: 24 }}>
        <div style={{ display: 'flex', alignItems: 'center', marginBottom: 8 }}>
          <div
            style={{
              width: 2,
              height: 18,
              backgroundColor: '#126ee3',
              borderRadius: 4,
              marginRight: 8,
            }}
          />
          <h2
            style={{
              fontSize: 15,
              fontWeight: 500,
              margin: 0,
              color: '#000000',
            }}
          >
            代码执行
          </h2>
        </div>
        <p style={{ fontSize: 13, color: '#677489', marginLeft: 12, marginTop: 0, marginBottom: 0 }}>
          在选定的会话中执行 Python 代码
        </p>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 24 }}>
        {/* 左侧：代码编辑器 */}
        <div
          style={{
            display: 'flex',
            flexDirection: 'column',
            gap: 16,
          }}
        >
          {/* 会话选择和代码编辑器 */}
          <div
            style={{
              backgroundColor: '#ffffff',
              borderRadius: 12,
              border: '1px solid #e7edf7',
              padding: 24,
            }}
          >
            {/* 会话选择 */}
            <div style={{ marginBottom: 16 }}>
              <label
                style={{
                  display: 'block',
                  fontSize: 14,
                  color: 'rgba(0,0,0,0.85)',
                  marginBottom: 8,
                }}
              >
                选择会话
              </label>
              <Select
                value={selectedSession}
                onChange={setSelectedSession}
                style={{ width: '100%' }}
                loading={sessionsLoading}
                placeholder="请选择会话"
              >
                {sessions
                  .filter((s) => s.status !== 'terminated' && s.status !== 'TERMINATED')
                  .map((s) => (
                    <Select.Option key={s.id} value={s.id}>
                      {s.id} ({s.runtime_type}) - {s.status}
                    </Select.Option>
                  ))}
              </Select>
            </div>

            {/* Python 代码编辑器 */}
            <div style={{ marginBottom: 16 }}>
              <label
                style={{
                  display: 'block',
                  fontSize: 14,
                  color: 'rgba(0,0,0,0.85)',
                  marginBottom: 8,
                }}
              >
                Python 代码 (Lambda Handler 格式)
              </label>
              <div
                style={{
                  border: '1px solid #d9d9d9',
                  borderRadius: 4,
                  overflow: 'hidden',
                }}
              >
                <Editor
                  height={280}
                  defaultLanguage="python"
                  value={code}
                  onChange={(value) => setCode(value || '')}
                  theme="vs-light"
                  options={{
                    minimap: { enabled: false },
                    fontSize: 13,
                    lineNumbers: 'on' as const,
                    scrollBeyondLastLine: false,
                    automaticLayout: true,
                  }}
                />
              </div>
            </div>

            {/* Event 数据编辑器 */}
            <div style={{ marginBottom: 16 }}>
              <label
                style={{
                  display: 'block',
                  fontSize: 14,
                  color: 'rgba(0,0,0,0.85)',
                  marginBottom: 8,
                }}
              >
                Event 数据 (JSON)
              </label>
              <div
                style={{
                  border: '1px solid #d9d9d9',
                  borderRadius: 4,
                  overflow: 'hidden',
                }}
              >
                <Editor
                  height={120}
                  defaultLanguage="json"
                  value={eventData}
                  onChange={(value) => setEventData(value || '')}
                  theme="vs-light"
                  options={{
                    minimap: { enabled: false },
                    fontSize: 13,
                    lineNumbers: 'on' as const,
                    scrollBeyondLastLine: false,
                    automaticLayout: true,
                  }}
                />
              </div>
              {eventError && (
                <div style={{ color: '#ff4d4f', fontSize: 12, marginTop: 4 }}>
                  {eventError}
                </div>
              )}
            </div>

            {/* 执行按钮 */}
            <Button
              type="primary"
              icon={<PlayCircleFilled />}
              onClick={handleExecute}
              disabled={loading}
              size="large"
              style={{ width: '100%' }}
            >
              {loading ? '执行中...' : '执行代码'}
            </Button>
          </div>
        </div>

        {/* 右侧：执行历史 */}
        <div
          style={{
            backgroundColor: '#ffffff',
            borderRadius: 12,
            border: '1px solid #e7edf7',
            padding: 24,
          }}
        >
          <h3
            style={{
              fontSize: 15,
              fontWeight: 500,
              marginTop: 0,
              marginBottom: 16,
              color: '#000000',
            }}
          >
            执行历史
          </h3>

          <div
            style={{
              maxHeight: 600,
              overflowY: 'auto',
            }}
          >
            {executions.length === 0 && !currentExecution ? (
              <Empty
                image={Empty.PRESENTED_IMAGE_SIMPLE}
                description="暂无执行记录"
                style={{ padding: '40px 0' }}
              />
            ) : (
              <>
                {/* 当前执行 */}
                {currentExecution && (
                  <ExecutionItem execution={currentExecution} getStatusConfig={getStatusConfig} />
                )}
                {/* 历史执行 */}
                {executions.map((exec) => (
                  <ExecutionItem
                    key={exec.id}
                    execution={exec}
                    getStatusConfig={getStatusConfig}
                  />
                ))}
              </>
            )}
          </div>
        </div>
      </div>
    </>
  );
}

/** 执行记录项组件 */
interface ExecutionItemProps {
  execution: ExecutionResponse;
  getStatusConfig: (status: string) => {
    color: string;
    icon: string;
    label: string;
  };
}

function ExecutionItem({ execution, getStatusConfig }: ExecutionItemProps) {
  const [expanded, setExpanded] = useState(false);

  const statusConfig = getStatusConfig(execution.status);

  return (
    <div
      key={execution.id}
      style={{
        border: '1px solid #e7edf7',
        borderRadius: 8,
        padding: 16,
        marginBottom: 12,
        transition: 'border-color 0.2s',
      }}
    >
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 12 }}>
        <div style={{ flex: 1 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 4 }}>
            <span style={{ fontSize: 13, fontWeight: 500, color: 'rgba(0,0,0,0.85)' }}>
              {execution.id}
            </span>
            <Tag color={statusConfig.color}>
              {statusConfig.icon} {statusConfig.label}
            </Tag>
          </div>
          <p style={{ fontSize: 12, color: '#677489', margin: 0 }}>{execution.created_at}</p>
        </div>

        {execution.execution_time && (
          <div style={{ textAlign: 'right' }}>
            <p style={{ fontSize: 12, color: '#677489', margin: 0 }}>耗时</p>
            <p style={{ fontSize: 13, fontWeight: 500, color: 'rgba(0,0,0,0.85)', margin: 0 }}>
              {(execution.execution_time * 1000).toFixed(0)}ms
            </p>
          </div>
        )}
      </div>

      {/* 返回值 */}
      {(execution.status === 'COMPLETED' || execution.status === 'completed') && execution.return_value && (
        <div style={{ marginBottom: 8 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 4, marginBottom: 4 }}>
            <CaretRightOutlined style={{ fontSize: 12, color: '#52c41a' }} />
            <p style={{ fontSize: 12, fontWeight: 500, color: 'rgba(0,0,0,0.85)', margin: 0 }}>
              返回值
            </p>
          </div>
          <pre
            style={{
              backgroundColor: '#f6ffed',
              border: '1px solid #b7eb8f',
              borderRadius: 4,
              padding: 8,
              fontSize: 11,
              fontFamily: 'monospace',
              overflow: 'auto',
              margin: 0,
            }}
          >
            {JSON.stringify(execution.return_value, null, 2)}
          </pre>
        </div>
      )}

      {/* 标准输出 */}
      {execution.stdout && (
        <div style={{ marginBottom: 8 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 4, marginBottom: 4 }}>
            <CaretRightOutlined style={{ fontSize: 12, color: '#1890ff' }} />
            <p style={{ fontSize: 12, fontWeight: 500, color: 'rgba(0,0,0,0.85)', margin: 0 }}>
              标准输出
            </p>
          </div>
          <pre
            style={{
              backgroundColor: '#fafafa',
              border: '1px solid #e7edf7',
              borderRadius: 4,
              padding: 8,
              fontSize: 11,
              fontFamily: 'monospace',
              overflow: 'auto',
              margin: 0,
            }}
          >
            {execution.stdout}
          </pre>
        </div>
      )}

      {/* 错误信息 */}
      {execution.stderr && (
        <div style={{ marginBottom: 8 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 4, marginBottom: 4 }}>
            <CaretRightOutlined style={{ fontSize: 12, color: '#ff4d4f' }} />
            <p style={{ fontSize: 12, fontWeight: 500, color: 'rgba(0,0,0,0.85)', margin: 0 }}>
              错误信息
            </p>
          </div>
          <pre
            style={{
              backgroundColor: '#fff1f0',
              border: '1px solid #ffccc7',
              borderRadius: 4,
              padding: 8,
              fontSize: 11,
              fontFamily: 'monospace',
              overflow: 'auto',
              color: '#ff4d4f',
              margin: 0,
            }}
          >
            {execution.stderr}
          </pre>
        </div>
      )}

      {/* 查看代码 */}
      <details style={{ marginTop: 8 }}>
        <summary
          style={{
            fontSize: 12,
            color: '#126ee3',
            cursor: 'pointer',
            userSelect: 'none',
          }}
        >
          查看代码
        </summary>
        <pre
          style={{
            marginTop: 8,
            backgroundColor: '#fafafa',
            border: '1px solid #e7edf7',
            borderRadius: 4,
            padding: 8,
            fontSize: 11,
            fontFamily: 'monospace',
            overflow: 'auto',
            margin: 0,
          }}
        >
          {execution.code}
        </pre>
      </details>
    </div>
  );
}
