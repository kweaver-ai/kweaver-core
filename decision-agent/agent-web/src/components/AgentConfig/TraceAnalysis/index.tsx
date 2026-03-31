import styles from './index.module.less';

import AgentPerformanceOverview from './AgentPerformanceOverview';
import ConversationLogList from './ConversationLogList';
import OptimizationSuggestion from './OptimizationSuggestion';
import { useEffect, useState } from 'react';
import { getAgentsByPost } from '@/apis/agent-factory';
import _ from 'lodash';
import dayjs, { Dayjs } from 'dayjs';
import { useAgentConfig } from '../AgentConfigContext.tsx';
import { useBusinessDomain } from '@/hooks';
import { useSearchParams } from 'react-router-dom';

const TraceAnalysis = ({ refreshTraceAnalysis }: any) => {
  const { publicBusinessDomain } = useBusinessDomain();
  const { state } = useAgentConfig();
  const [qualityInsightAgentDetails, setQualityInsightAgentDetails] = useState<any>({});
  const [dateValue, setDateValue] = useState<Dayjs[]>([dayjs().subtract(1, 'd'), dayjs()]);
  useEffect(() => {
    if (publicBusinessDomain?.id) {
      getQualityInsightAgent();
    }
  }, [publicBusinessDomain]);
  const getQualityInsightAgent = async () => {
    const res: any = await getAgentsByPost({
      pagination_marker_str: '',
      category_id: '',
      size: 120,
      agent_keys: ['QualityInsight_Agent'],
      custom_space_id: '',
      is_to_square: 1,
      business_domain_ids: [publicBusinessDomain!.id], // 获取公共业务域的
    });
    if (res) {
      const target = _.get(res, 'entries', [])[0];
      if (target) {
        setQualityInsightAgentDetails(target);
      }
    }
  };
  return (
    <div className={styles.container}>
      <AgentPerformanceOverview
        dateValue={dateValue}
        onDateChange={(value: Dayjs[]) => {
          setDateValue(value);
        }}
        refreshTraceAnalysis={refreshTraceAnalysis}
      />
      <ConversationLogList
        refreshTraceAnalysis={refreshTraceAnalysis}
        dateValue={dateValue}
        qualityInsightAgentDetails={qualityInsightAgentDetails}
      />
      <div className="dip-mt-24">
        <OptimizationSuggestion
          id={state.id!}
          analysisLevel="agent"
          qualityInsightAgentDetails={qualityInsightAgentDetails}
          startTime={dateValue[0].valueOf()}
          endTime={dateValue[1].valueOf()}
        />
      </div>
    </div>
  );
};

export default (props: any) => {
  const [searchParams] = useSearchParams();
  const { state } = useAgentConfig();
  const isTemplate = searchParams.get('mode') === 'editTemplate';
  return (
    !isTemplate &&
    props.userPermissions.hasTraceAnalysisPermission &&
    !props.isSkillAgent &&
    state.is_published && <TraceAnalysis {...props} />
  );
};
