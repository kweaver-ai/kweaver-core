import React, { useState, useEffect } from 'react';
import intl from 'react-intl-universal';
import { Switch } from 'antd';
import { useAgentConfig } from '../../AgentConfigContext';
import PlanModeIcon from '@/assets/icons/plan-mode.svg';
import SectionPanel from '../../common/SectionPanel';

const PlanModeSection: React.FC = () => {
  const { state, actions } = useAgentConfig();
  const [enabled, setEnabled] = useState(state.config.plan_mode?.is_enabled || false);

  // 检查是否可编辑任务规划
  const canEdit = actions.canEditField('plan_mode');

  // 处理状态变更
  const handleEnabledChange = (checked: boolean) => {
    if (!canEdit) return;

    actions.updatPlanMode(checked);
  };

  // 当Context状态更新时，同步本地状态
  useEffect(() => {
    setEnabled(state.config.plan_mode?.is_enabled || false);
  }, [state.config.plan_mode?.is_enabled]);

  // dolphin模式下，隐藏任务规划
  return state.config.is_dolphin_mode ? null : (
    <SectionPanel
      title={intl.get('dataAgent.taskPlanningTitle')}
      description={intl.get('dataAgent.taskPlanningDescription')}
      rightElement={<Switch checked={!!enabled} onChange={handleEnabledChange} disabled={!canEdit} />}
      showCollapseArrow={false}
      icon={<PlanModeIcon style={{ color: '#1E2C72' }} className="dip-font-16" />}
    />
  );
};

export default PlanModeSection;
