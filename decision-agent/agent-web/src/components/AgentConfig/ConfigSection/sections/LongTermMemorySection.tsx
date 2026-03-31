import React, { useState, useEffect } from 'react';
import intl from 'react-intl-universal';
import { Switch, Tooltip } from 'antd';
import { QuestionCircleOutlined } from '@ant-design/icons';
import { useAgentConfig } from '../../AgentConfigContext';
import MemoryIcon from '@/assets/icons/memory.svg';
import SectionPanel from '../../common/SectionPanel';

const LongTermMemorySection: React.FC = () => {
  const { state, actions } = useAgentConfig();
  const [longTermMemoryEnabled, setLongTermMemoryEnabled] = useState(state.config.memory?.is_enabled || false);

  // 检查是否可编辑长期记忆配置
  const canEditLongTermMemory = actions.canEditField('memory');

  // 处理长期记忆启用状态变更
  const handleLongTermMemoryEnabledChange = (checked: boolean) => {
    if (!canEditLongTermMemory) return;
    setLongTermMemoryEnabled(checked);
    actions.updateLongTermMemory(checked);
  };

  // 当Context状态更新时，同步本地状态
  useEffect(() => {
    setLongTermMemoryEnabled(state.config.memory?.is_enabled || false);
  }, [state.config.memory?.is_enabled]);

  return (
    <SectionPanel
      className="dip-border-line-b"
      title={
        <>
          <div>{intl.get('dataAgent.config.longTermMemory')}</div>
          <Tooltip title={intl.get('dataAgent.config.enableLongTermMemory')}>
            <QuestionCircleOutlined className="dip-pointer dip-font-14" />
          </Tooltip>
        </>
      }
      description={intl.get('dataAgent.config.enableMemoryFeature')}
      rightElement={
        <Switch
          checked={!!longTermMemoryEnabled}
          onChange={handleLongTermMemoryEnabledChange}
          disabled={!canEditLongTermMemory}
        />
      }
      showCollapseArrow={false}
      icon={<MemoryIcon />}
    >
      <div className="long-term-memory-config">
        {longTermMemoryEnabled ? (
          <div className="dip-mt-16 dip-text-color-45 dip-text-12">
            {canEditLongTermMemory
              ? intl.get('dataAgent.config.longTermMemoryEnabled')
              : intl.get('dataAgent.config.longTermMemoryNotEditable')}
          </div>
        ) : (
          <div className="dip-mt-16 dip-text-color-45 dip-text-12">
            {intl.get('dataAgent.config.longTermMemoryDisabled')}
          </div>
        )}
      </div>
    </SectionPanel>
  );
};

export default LongTermMemorySection;
