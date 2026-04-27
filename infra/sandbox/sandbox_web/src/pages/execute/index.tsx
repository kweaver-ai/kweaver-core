/**
 * 代码执行页面
 */
import { useEffect, useState } from 'react';
import { useSearchParams } from 'react-router-dom';
import { Alert, Button, Empty, Input, Select, Tag } from 'antd';
import { CaretRightOutlined, PlayCircleFilled } from '@ant-design/icons';
import Editor from '@monaco-editor/react';
import { useExecution } from '@hooks/useExecution';
import { useSessions } from '@hooks/useSessions';
import { EXECUTION_STATUS_LABELS } from '@constants/runtime';
import type { CodeLanguage, ExecuteCodeRequest, ExecutionResponse } from '@apis/executions';

const CODE_TEMPLATES: Record<CodeLanguage, string> = {
  python: `def handler(event):
    name = event.get("name", "World")
    return {"message": f"Hello, {name}!"}
`,
  javascript: `function handler(event) {
  const name = event?.name || 'World';
  return { message: \`Hello, \${name}!\` };
}`,
  shell: `pwd
ls -la
# 也支持:
# bash run.sh
# python scripts/analyze_project.py
# bash python scripts/analyze_project.py`,
};

const DEFAULT_EVENT = `{
  "name": "Sandbox Platform"
}`;

export default function ExecutePage() {
  const [searchParams] = useSearchParams();
  const { sessions, loading: sessionsLoading, fetchSessions } = useSessions();
  const [selectedSession, setSelectedSession] = useState<string>('');
  const [language, setLanguage] = useState<CodeLanguage>('python');
  const [code, setCode] = useState(CODE_TEMPLATES.python);
  const [eventData, setEventData] = useState(DEFAULT_EVENT);
  const [eventError, setEventError] = useState('');
  const [workingDirectory, setWorkingDirectory] = useState('');

  const { executions, currentExecution, loading, executeCode, fetchSessionExecutions } = useExecution();

  useEffect(() => {
    fetchSessions();
  }, [fetchSessions]);

  useEffect(() => {
    const sessionId = searchParams.get('sessionId');
    if (sessionId) {
      setSelectedSession(sessionId);
    } else if (sessions.length > 0) {
      const activeSessions = sessions.filter(
        (s) => s.status !== 'terminated' && s.status !== 'TERMINATED'
      );
      if (activeSessions.length > 0) {
        const firstRunning = activeSessions.find(
          (s) => s.status === 'running' || s.status === 'RUNNING'
        );
        setSelectedSession(firstRunning?.id || activeSessions[0].id);
      }
    }
  }, [searchParams, sessions]);

  useEffect(() => {
    if (selectedSession) {
      fetchSessionExecutions(selectedSession);
    }
  }, [selectedSession, fetchSessionExecutions]);

  const handleLanguageChange = (value: CodeLanguage) => {
    setLanguage(value);
    setCode(CODE_TEMPLATES[value]);
    setEventError('');
    setWorkingDirectory('');
    setEventData(value === 'shell' ? '{}' : DEFAULT_EVENT);
  };

  const handleExecute = async () => {
    let parsedEvent: Record<string, unknown> = {};
    if (language !== 'shell' || eventData.trim()) {
      try {
        parsedEvent = JSON.parse(eventData);
        setEventError('');
      } catch {
        setEventError('Event 数据必须是有效的 JSON 格式');
        return;
      }
    }

    const request: ExecuteCodeRequest = {
      code,
      language,
      event: parsedEvent,
      timeout: 30,
      working_directory:
        language === 'shell' && workingDirectory.trim()
          ? workingDirectory.trim()
          : undefined,
    };

    await executeCode(selectedSession, request);
  };

  const getStatusConfig = (status: string) => {
    const configs: Record<string, { color: string; icon: string; label: string }> = {
      PENDING: { color: 'warning', icon: '⏱', label: EXECUTION_STATUS_LABELS.PENDING },
      RUNNING: { color: 'processing', icon: '⚡', label: EXECUTION_STATUS_LABELS.RUNNING },
      COMPLETED: { color: 'success', icon: '✓', label: EXECUTION_STATUS_LABELS.COMPLETED },
      FAILED: { color: 'error', icon: '✗', label: EXECUTION_STATUS_LABELS.FAILED },
      TIMEOUT: { color: 'error', icon: '⏱', label: EXECUTION_STATUS_LABELS.TIMEOUT },
      CRASHED: { color: 'error', icon: '💥', label: EXECUTION_STATUS_LABELS.CRASHED },
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
          在选定的会话中执行 Python、JavaScript 或 Shell 脚本
        </p>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 24 }}>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
          <div
            style={{
              backgroundColor: '#ffffff',
              borderRadius: 12,
              border: '1px solid #e7edf7',
              padding: 24,
            }}
          >
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

            <div style={{ marginBottom: 16 }}>
              <label
                style={{
                  display: 'block',
                  fontSize: 14,
                  color: 'rgba(0,0,0,0.85)',
                  marginBottom: 8,
                }}
              >
                执行语言
              </label>
              <Select
                value={language}
                onChange={handleLanguageChange}
                style={{ width: '100%' }}
                options={[
                  { label: 'Python', value: 'python' },
                  { label: 'JavaScript', value: 'javascript' },
                  { label: 'Shell', value: 'shell' },
                ]}
              />
            </div>

            {language === 'shell' && (
              <>
                <div style={{ marginBottom: 16 }}>
                  <label
                    style={{
                      display: 'block',
                      fontSize: 14,
                      color: 'rgba(0,0,0,0.85)',
                      marginBottom: 8,
                    }}
                  >
                    工作目录 <span style={{ color: '#8c8c8c' }}>(可选，相对于 workspace 根目录)</span>
                  </label>
                  <Input
                    placeholder="例如: skill/mini-wiki"
                    value={workingDirectory}
                    onChange={(e) => setWorkingDirectory(e.target.value)}
                  />
                </div>
                <Alert
                  type="info"
                  showIcon
                  style={{ marginBottom: 16 }}
                  message="Shell 执行说明"
                  description="未填写工作目录时默认在 /workspace 执行。支持 bash run.sh、python scripts/analyze_project.py，也兼容 bash python scripts/analyze_project.py 这类上层调用写法。"
                />
              </>
            )}

            <div style={{ marginBottom: 16 }}>
              <label
                style={{
                  display: 'block',
                  fontSize: 14,
                  color: 'rgba(0,0,0,0.85)',
                  marginBottom: 8,
                }}
              >
                {language === 'python'
                  ? 'Python 代码 (Lambda Handler 格式)'
                  : language === 'javascript'
                    ? 'JavaScript 代码'
                    : 'Shell 脚本'}
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
                  language={language === 'shell' ? 'shell' : language}
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
              <div style={{ fontSize: 12, color: '#677489', marginTop: 8 }}>
                {language === 'shell'
                  ? 'Shell 模式下通常可留空为 {}，如无需要可不传业务事件。'
                  : 'Python 和 JavaScript 模式下会作为 handler 的 event 参数传入。'}
              </div>
              {eventError && (
                <div style={{ color: '#ff4d4f', fontSize: 12, marginTop: 4 }}>{eventError}</div>
              )}
            </div>

            <Button
              type="primary"
              icon={<PlayCircleFilled />}
              onClick={handleExecute}
              disabled={loading || !selectedSession}
              size="large"
              style={{ width: '100%' }}
            >
              {loading ? '执行中...' : language === 'shell' ? '执行 Shell' : '执行代码'}
            </Button>
          </div>
        </div>

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

          <div style={{ maxHeight: 600, overflowY: 'auto' }}>
            {executions.length === 0 && !currentExecution ? (
              <Empty
                image={Empty.PRESENTED_IMAGE_SIMPLE}
                description="暂无执行记录"
                style={{ padding: '40px 0' }}
              />
            ) : (
              <>
                {currentExecution && (
                  <ExecutionItem execution={currentExecution} getStatusConfig={getStatusConfig} />
                )}
                {executions.map((exec) => (
                  <ExecutionItem key={exec.id} execution={exec} getStatusConfig={getStatusConfig} />
                ))}
              </>
            )}
          </div>
        </div>
      </div>
    </>
  );
}

interface ExecutionItemProps {
  execution: ExecutionResponse;
  getStatusConfig: (status: string) => {
    color: string;
    icon: string;
    label: string;
  };
}

function ExecutionItem({ execution, getStatusConfig }: ExecutionItemProps) {
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
          {execution.language && (
            <p style={{ fontSize: 12, color: '#677489', margin: '4px 0 0 0' }}>
              语言：{execution.language}
            </p>
          )}
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
