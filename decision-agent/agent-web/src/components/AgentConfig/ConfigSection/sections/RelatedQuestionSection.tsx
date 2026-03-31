import React, { useState, useEffect } from 'react';
import intl from 'react-intl-universal';
import { Switch } from 'antd';
import { useAgentConfig } from '../../AgentConfigContext';
import RelatedQuestionIcon from '@/assets/icons/related-question.svg';
import SectionPanel from '../../common/SectionPanel';

const RelatedQuestionSection: React.FC = () => {
  const { state, actions } = useAgentConfig();
  const [enabled, setEnabled] = useState(state.config.related_question?.is_enabled || false);

  // 检查是否可编辑长期记忆配置
  const canEdit = actions.canEditField('related_question');

  // 处理长期记忆启用状态变更
  const handleEnabledChange = (checked: boolean) => {
    if (!canEdit) return;

    actions.updateRelatedQuestion(checked);
  };

  // 当Context状态更新时，同步本地状态
  useEffect(() => {
    setEnabled(state.config.related_question?.is_enabled || false);
  }, [state.config.related_question?.is_enabled]);

  return (
    <SectionPanel
      className={state.config.is_dolphin_mode ? '' : 'dip-border-line-b'}
      title={intl.get('dataAgent.config.relatedQuestions')}
      description={intl.get('dataAgent.config.enableRelatedQuestions')}
      rightElement={<Switch checked={!!enabled} onChange={handleEnabledChange} disabled={!canEdit} />}
      showCollapseArrow={false}
      icon={<RelatedQuestionIcon style={{ color: '#1E2C72' }} className="dip-font-16" />}
    />
  );
};

export default RelatedQuestionSection;
