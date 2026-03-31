import { useEffect, useState } from 'react';
import styles from './index.module.less';
import TraceAnalysisTitle from '../Title';

import { DatePicker } from 'antd';
import dayjs from 'dayjs';
import { getAgentObservabilityMetrics } from '@/apis/agent-app';
import { useAgentConfig } from '../../AgentConfigContext';

const { RangePicker } = DatePicker;

interface AgentMetric {
  key: string;
  label: string;
  unit?: string;
  value: any;
  width?: number;
}

const AgentPerformanceOverview = ({ dateValue, onDateChange, refreshTraceAnalysis }: any) => {
  const { state } = useAgentConfig();
  const [agentMetrics, setAgentMetrics] = useState<AgentMetric[]>([
    { key: 'total_requests', label: '总请求次数', unit: '次', value: 0 },
    { key: 'total_sessions', label: '总会话数', unit: '个', value: 0 },
    { key: 'avg_session_rounds', label: '平均会话轮次', unit: '轮', value: 0 },
    { key: 'run_success_rate', label: '任务成功率', unit: '%', value: 0 },
    { key: 'avg_execute_duration', label: '平均执行耗时', unit: 's', value: 0 },
    { key: 'avg_ttft_duration', label: '平均首Token响应耗时', unit: 'ms', value: 0, width: 192 },
    { key: 'tool_success_rate', label: '工具成功率', unit: '%', value: 0 },
  ]);

  useEffect(() => {
    onDateChange([dateValue[0], dayjs()]);
  }, [refreshTraceAnalysis]);

  useEffect(() => {
    getData();
  }, [dateValue, state.id]);
  const getData = async () => {
    const res = await getAgentObservabilityMetrics(state.id!, {
      agent_version: 'v0',
      start_time: dateValue[0].valueOf(),
      end_time: dateValue[1].valueOf(),
    });
    if (res) {
      agentMetrics.forEach(item => {
        if (item.key === 'avg_execute_duration') {
          item.value = (res[item.key] ?? 0) / 1000;
          item.value = item.value > 0 ? item.value.toFixed(3) : 0;
        } else {
          item.value = res[item.key] ?? 0;
        }
      });
      setAgentMetrics([...agentMetrics]);
    }
  };
  return (
    <div className={styles.container}>
      <TraceAnalysisTitle
        title="性能概览"
        description="智能体核心指标实时监控"
        extra={
          <div className="dip-flex-align-center">
            <span>按会话时间搜索：</span>
            <RangePicker
              value={dateValue}
              onChange={dates => {
                onDateChange(dates || []);
              }}
              showTime
              allowClear={false}
              presets={[
                { label: '最近1小时', value: [dayjs().subtract(1, 'h'), dayjs()] },
                { label: '最近3小时', value: [dayjs().subtract(3, 'h'), dayjs()] },
                { label: '最近1天', value: [dayjs().subtract(1, 'd'), dayjs()] },
                { label: '最近7天', value: [dayjs().subtract(7, 'd'), dayjs()] },
                { label: '最近一个月', value: [dayjs().subtract(1, 'months'), dayjs()] },
              ]}
            />
          </div>
        }
      />
      <div className="dip-flex-space-between dip-mt-16" style={{ gap: 12 }}>
        {agentMetrics.map(metric => {
          const rateArr = ['run_success_rate', 'tool_success_rate'];
          return (
            <div style={{ minWidth: metric.width || 'auto' }} key={metric.key} className={styles.card}>
              <div className={styles.label}>
                {metric.label}
                {metric.unit ? `（${metric.unit}）` : null}
              </div>
              <div className={styles.value}>
                {rateArr.includes(metric.key) ? metric.value.toFixed(2) : metric.value}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};
export default AgentPerformanceOverview;
