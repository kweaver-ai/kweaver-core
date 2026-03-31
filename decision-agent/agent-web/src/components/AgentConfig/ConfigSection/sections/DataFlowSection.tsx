import React, { useState, useEffect } from 'react';
import { Switch, Tooltip } from 'antd';
import { QuestionCircleOutlined } from '@ant-design/icons';
import dataFlowIcon from '@/assets/images/dataflow.png';
import { useAgentConfig } from '../../AgentConfigContext';
import SectionPanel from '../../common/SectionPanel';

const DataFlowSection: React.FC = () => {
  const { state, actions } = useAgentConfig();
  const [dataFlowEnabled, setDataFlowEnabled] = useState(state.config.is_data_flow_set_enabled || false);

  // 处理数据流启用状态变更
  const handleDataFlowEnabledChange = (checked: boolean) => {
    setDataFlowEnabled(checked);
    actions.updateDataFlow(checked);
  };

  // 当Context状态更新时，同步本地状态
  useEffect(() => {
    setDataFlowEnabled(state.config.is_data_flow_set_enabled || false);
  }, [state.config.is_data_flow_set_enabled]);

  return (
    <SectionPanel
      title={
        <>
          <div>数据流</div>
          <Tooltip title="允许用户创建和管理数据流">
            <QuestionCircleOutlined className="tooltip-icon" />
          </Tooltip>
        </>
      }
      description="启用后，用户在使用时可创建数据流"
      rightElement={<Switch checked={!!dataFlowEnabled} onChange={handleDataFlowEnabledChange} />}
      showCollapseArrow={false}
      icon={<img src={dataFlowIcon} alt="数据流" width={20} height={20} />}
    >
      <div className="dataflow-config">
        {dataFlowEnabled && <div className="dataflow-content">{/* 这里可以添加更多的数据流配置选项 */}</div>}
      </div>
    </SectionPanel>
  );
};

export default DataFlowSection;
