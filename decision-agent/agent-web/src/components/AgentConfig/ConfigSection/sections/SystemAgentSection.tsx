import React, { useState, useMemo, useEffect } from 'react';
import { useSearchParams } from 'react-router-dom';
import { Radio } from 'antd';
import intl from 'react-intl-universal';
import { getAgentManagementPerm } from '@/apis/agent-factory';
import { useAgentConfig } from '../../AgentConfigContext';
import LlmIcon from '@/assets/icons/model.svg';
import SectionPanel from '../../common/SectionPanel';

const SystemAgentSection: React.FC = () => {
  const [searchParams] = useSearchParams();
  const { state, actions } = useAgentConfig();
  const [isExpanded, setIsExpanded] = useState(false);
  const [canCreateSystemAgent, setCanCreateSystemAgent] = useState<boolean>(false);

  const [show, editable] = useMemo(() => {
    const templateId = searchParams.get('templateId');
    const mode = searchParams.get('mode');
    const agentId = searchParams.get('agentId');

    // 编辑模板页面，不显示系统agent; 新建/编辑agent页面，如果create_system_agent为fasle，则不显示系统agent；编辑agent页面，不可修改系统agent
    return [!(templateId && mode === 'editTemplate') && canCreateSystemAgent, !agentId];
  }, [searchParams, canCreateSystemAgent]);

  // 检查是否可编辑系统智能体配置
  const canEditSystemAgent = actions.canEditField('is_system_agent');

  // 处理系统智能体状态变更
  const handleSystemAgentChange = (e: any) => {
    if (!canEditSystemAgent) return;
    const value = e.target.value;
    actions.updateSpecificField('is_system_agent', value);
  };

  useEffect(() => {
    // 获取权限
    const fetchPerm = async () => {
      try {
        const {
          agent: { create_system_agent },
        } = await getAgentManagementPerm();

        setCanCreateSystemAgent(create_system_agent);
      } catch {}
    };

    fetchPerm();
  }, []);

  return show ? (
    <SectionPanel
      title={intl.get('dataAgent.config.systemAgentConfig')}
      description={intl.get('dataAgent.config.systemAgentConfigDescription')}
      isExpanded={isExpanded}
      onToggle={() => setIsExpanded(!isExpanded)}
      icon={<LlmIcon />}
      className="dip-border-line-b"
    >
      <div className="dip-mt-16">
        <Radio.Group
          value={state.is_system_agent || 0}
          onChange={handleSystemAgentChange}
          disabled={!canEditSystemAgent || !editable}
        >
          <Radio value={0}>{intl.get('dataAgent.config.regularAgent')}</Radio>
          <Radio value={1}>{intl.get('dataAgent.config.systemAgent')}</Radio>
        </Radio.Group>
        <div className="dip-mt-8 dip-text-color-45 dip-text-12">
          {canEditSystemAgent
            ? intl.get('dataAgent.config.systemAgentHasPerm')
            : intl.get('dataAgent.config.systemAgentCannotEdit')}
        </div>
      </div>
    </SectionPanel>
  ) : null;
};

export default SystemAgentSection;
