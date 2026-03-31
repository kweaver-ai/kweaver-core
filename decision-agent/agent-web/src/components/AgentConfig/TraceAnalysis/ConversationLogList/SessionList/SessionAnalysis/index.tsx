import { Button, Drawer, message, type TableColumnsType, Tag } from 'antd';
import { CloseOutlined } from '@ant-design/icons';
import TraceAnalysisTitle from '../../../Title';
import styles from './index.module.less';
import { useEffect, useState } from 'react';
import ADTable from '@/components/ADTable';
import DipButton from '@/components/DipButton';
import RunAnalysis from './RunAnalysis';
import { useAgentConfig } from '../../../../AgentConfigContext';
import {
  getAgentObservabilityRunDetail,
  getAgentObservabilityRunList,
  getAgentObservabilitySessionMetric,
} from '@/apis/agent-app';
import OptimizationSuggestion from '@/components/AgentConfig/TraceAnalysis/OptimizationSuggestion';
import _ from 'lodash';

const SessionAnalysis = ({ onClose, activeSession, conversationId, dateValue, qualityInsightAgentDetails }: any) => {
  const { state } = useAgentConfig();
  const [open, setOpen] = useState(false);
  const [listData, setListData] = useState<any>([]);
  const [activeRun, setActiveRun] = useState<any>(null);
  const [messageApi, contextHolder] = message.useMessage();

  const sessionId = activeSession?.session_id;

  const [sessionMetrics, setSessionMetrics] = useState<any>([
    { key: 'session_run_count', label: '会话总轮数', unit: '轮', value: 0 },
    { key: 'session_duration', label: '会话时长', unit: 's', value: 0 },
    { key: 'avg_run_execute_duration', label: '平均执行耗时', unit: 's', value: 0 },
    { key: 'avg_run_ttft_duration', label: '平均首Token响应耗时', unit: 'ms', value: 0, width: 192 },
    { key: 'run_error_count', label: 'Run错误次数', unit: '次', value: 0 },
    { key: 'tool_fail_count', label: '工具错误次数', unit: '次', value: 0 },
  ]);

  useEffect(() => {
    getRunList();
    getMetrics();
  }, [sessionId, conversationId, dateValue]);

  const getRunList = async () => {
    const res = await getAgentObservabilityRunList(state.id!, conversationId, sessionId, {
      page: 1,
      size: 10000,
      agent_id: state.id!,
      agent_version: 'v0',
      start_time: dateValue[0].valueOf(),
      end_time: dateValue[1].valueOf(),
    });
    if (res) {
      console.log(res, 'run列表');
      setListData(res.entries);
    }
  };

  const getMetrics = async () => {
    const res = await getAgentObservabilitySessionMetric(state.id!, conversationId, sessionId, {
      agent_id: state.id!,
      agent_version: 'v0',
      start_time: dateValue[0].valueOf(),
      end_time: dateValue[1].valueOf(),
    });
    if (res) {
      sessionMetrics.forEach((item: any) => {
        if (['session_duration', 'avg_run_execute_duration'].includes(item.key)) {
          item.value = (res[item.key] ?? 0) / 1000;
          item.value = item.value > 0 ? item.value.toFixed(3) : 0;
        } else {
          item.value = res[item.key] ?? 0;
        }
      });
      setSessionMetrics([...sessionMetrics]);
    }
  };

  const columns: TableColumnsType = [
    {
      title: '问题',
      dataIndex: 'input_message',
    },
    {
      title: '运行状态',
      dataIndex: 'status',
      render: value => {
        if (value === 'success') {
          return <Tag color="success">成功</Tag>;
        }
        return <Tag color="error">失败</Tag>;
      },
    },
    {
      title: '运行时长',
      dataIndex: 'total_time',
      render: value => {
        return `${(value / 1000).toFixed(3)}秒`;
      },
    },
    {
      title: '操作',
      dataIndex: 'operate',
      render: (_value, record: any) => {
        return (
          <DipButton
            type="link"
            onClick={async () => {
              const res = await getAgentObservabilityRunDetail(state.id!, conversationId, sessionId, record.run_id, {
                agent_id: state.id!,
                agent_version: 'v0',
                start_time: dateValue[0].valueOf(),
                end_time: dateValue[1].valueOf(),
              });
              if (res) {
                if (!_.isEmpty(res.progress)) {
                  setOpen(true);
                  setActiveRun(record);
                } else {
                  messageApi.warning('无执行进度数据');
                }
              }
            }}
          >
            查看详情
          </DipButton>
        );
      },
    },
  ];
  return (
    <Drawer
      styles={{
        body: {
          padding: 0,
        },
      }}
      closable={false}
      width="80vw"
      open
      onClose={onClose}
      maskStyle={{ background: 'rgba(0,0,0,0.05)' }}
    >
      <div className="dip-full dip-position-r dip-p-24 dip-overflowY-auto">
        <div className="dip-position-f" style={{ top: 12, right: 12 }}>
          <Button onClick={onClose} type="text" icon={<CloseOutlined />} />
        </div>
        <TraceAnalysisTitle title="会话分析" description={sessionId} />
        <div className="dip-flex-space-between dip-mt-16" style={{ gap: 12 }}>
          {sessionMetrics.map((metric: any) => (
            <div style={{ minWidth: metric.width || 'auto' }} key={metric.key} className={styles.card}>
              <div className={styles.label}>
                {metric.label}
                {metric.unit ? `（${metric.unit}）` : null}
              </div>
              <div className={styles.value}>{metric.value}</div>
            </div>
          ))}
        </div>
        <TraceAnalysisTitle className="dip-mt-16" title="执行记录" description="本次会话中所有执行单元的详细信息" />
        <div>
          <ADTable
            rowKey="run_id"
            className="dip-mt-16"
            showHeader={false}
            showSearch={false}
            size="small"
            columns={columns}
            dataSource={listData}
          />
        </div>
        <div className="dip-mt-24">
          <OptimizationSuggestion
            id={sessionId}
            analysisLevel="session"
            qualityInsightAgentDetails={qualityInsightAgentDetails}
            startTime={dateValue[0].valueOf()}
            endTime={dateValue[1].valueOf()}
          />
        </div>
        <RunAnalysis
          open={open}
          onClose={() => {
            setOpen(false);
            setActiveRun(null);
          }}
          activeRun={activeRun}
          conversationId={conversationId}
          sessionId={sessionId}
          dateValue={dateValue}
          qualityInsightAgentDetails={qualityInsightAgentDetails}
        />
        {contextHolder}
      </div>
    </Drawer>
  );
};

export default ({ open, ...restProp }: any) => {
  return open && <SessionAnalysis {...restProp} />;
};
