import { useEffect, useState } from 'react';
import { Button, Drawer } from 'antd';
import { CloseOutlined } from '@ant-design/icons';
import TraceAnalysisTitle from '@/components/AgentConfig/TraceAnalysis/Title';
import ExecutionTimeline from './ExecutionTimeline';
import ExecutionDetails from './ExecutionDetails';
import { getAgentObservabilityRunDetail } from '@/apis/agent-app';
import { useAgentConfig } from '../../../../../AgentConfigContext';
import _ from 'lodash';

const RunAnalysis = ({ onClose, sessionId, conversationId, dateValue, activeRun, qualityInsightAgentDetails }: any) => {
  const runId = activeRun?.run_id;
  const { state } = useAgentConfig();
  const [selectedStep, setSelectedStep] = useState<any>(null);

  const [runDetails, setRunDetails] = useState({});

  useEffect(() => {
    getRunDetails();
  }, [sessionId, conversationId, dateValue, runId]);

  const getRunDetails = async () => {
    const res = await getAgentObservabilityRunDetail(state.id!, conversationId, sessionId, runId, {
      agent_id: state.id!,
      agent_version: 'v0',
      start_time: dateValue[0].valueOf(),
      end_time: dateValue[1].valueOf(),
    });
    if (res) {
      console.log(res, 'run详情');
      setRunDetails(res);
    }
  };

  const handleStepClick = (step: any) => {
    console.log(step, '点击步骤');
    setSelectedStep(step);
  };

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
      maskStyle={{ background: 'rgba(0,0,0,0.05)' }}
      onClose={onClose}
    >
      <div className="dip-full dip-position-r dip-p-24 dip-flex-column">
        <div className="dip-position-f" style={{ top: 12, right: 12 }}>
          <Button onClick={onClose} type="text" icon={<CloseOutlined />} />
        </div>
        <TraceAnalysisTitle
          title="执行记录详情"
          description={
            <div className="dip-ellipsis" title={activeRun.input_message}>
              {activeRun.input_message}
            </div>
          }
        />
        <div className="dip-flex-item-full-height dip-flex dip-mt-16">
          <div className="dip-flex-item-full-width">
            {!_.isEmpty(runDetails) && <ExecutionTimeline runDetails={runDetails} onStepClick={handleStepClick} />}
          </div>
          <div className="dip-flex-item-full-width dip-ml-10">
            {!_.isEmpty(selectedStep) && (
              <ExecutionDetails
                dateValue={dateValue}
                runId={runId}
                step={selectedStep}
                qualityInsightAgentDetails={qualityInsightAgentDetails}
              />
            )}
          </div>
        </div>
      </div>
    </Drawer>
  );
};

export default ({ open, ...restProp }: any) => {
  return open && <RunAnalysis {...restProp} />;
};
