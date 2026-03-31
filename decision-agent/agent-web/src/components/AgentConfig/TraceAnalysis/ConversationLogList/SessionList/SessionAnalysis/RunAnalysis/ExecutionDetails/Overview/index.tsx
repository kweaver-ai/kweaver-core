import styles from './index.module.less';
import TraceAnalysisTitle from '../../../../../../Title';
import { Alert, Col, Row } from 'antd';
import dayjs from 'dayjs';
import OptimizationSuggestion from '@/components/AgentConfig/TraceAnalysis/OptimizationSuggestion';
const Overview = ({ step, qualityInsightAgentDetails, dateValue, runId }: any) => {
  const progressItem = step.sourceData;
  const executeInfo = [
    {
      label: '执行开始时间',
      value: dayjs.unix(progressItem.start_time).format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      label: '执行结束时间',
      value: dayjs.unix(progressItem.end_time).format('YYYY/MM/DD HH:mm:ss'),
    },
    {
      label: '执行耗时',
      value: `${(progressItem.end_time - progressItem.start_time).toFixed(3)}s`,
    },
    {
      label: 'Token消耗',
      value: `${progressItem.token_usage.total_tokens}`,
    },
  ];

  return (
    <div className="dip-full dip-flex-column dip-overflowY-auto dip-overflowX-hidden">
      <Alert message="执行成功" type="success" showIcon />
      <TraceAnalysisTitle className="dip-mt-16 dip-mb-16" title="执行信息" />
      <Row gutter={[8, 8]}>
        {executeInfo.map((item, index) => (
          <Col span={12} key={index}>
            <div className={styles.card}>
              <div className="dip-font-12 dip-text-color-45">{item.label}</div>
              <div className="dip-text-color dip-font-16">{item.value}</div>
            </div>
          </Col>
        ))}
      </Row>
      <div className="dip-mt-24">
        <OptimizationSuggestion
          id={runId}
          analysisLevel="run"
          qualityInsightAgentDetails={qualityInsightAgentDetails}
          startTime={dateValue[0].valueOf()}
          endTime={dateValue[1].valueOf()}
        />
      </div>
    </div>
  );
};

export default Overview;
