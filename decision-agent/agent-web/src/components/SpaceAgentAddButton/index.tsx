import React, { useState } from 'react';
import { Button, message } from 'antd';
import intl from 'react-intl-universal';
import PlusIcon from '@/assets/icons/plus.svg';
import { addSpaceResource, SpaceResourceEnum } from '@/apis/agent-factory';
import AgentSelectionModal from './AgentSelectionModal';

interface SpaceAgentAddButtonProps {
  // 按钮的文字
  text?: string;

  // 自定义空间的id
  customSpaceId: string;

  onAddSuccess?: (agents: any[]) => void;

  children?: React.ReactElement;
}

// 将agents数据转换成后端接口需要的格式
const transformAgentToBackend = (data: any[]) => {
  return data.map(agent => ({
    resource_type: SpaceResourceEnum.DataAgent,
    resource_id: agent.id,
    resource_name: agent.name,
  }));
};

const SpaceAgentAddButton = ({ text, customSpaceId, onAddSuccess, children }: SpaceAgentAddButtonProps) => {
  const [showAgentSelectionModal, setShowAgentSelectionModal] = useState<boolean>(false);

  // 添加Agent
  const addAgent = async (selections: any[]) => {
    try {
      const formatSelections = transformAgentToBackend(selections);
      await addSpaceResource({ id: customSpaceId, resources: formatSelections });
      message.success(intl.get('dataAgent.addSuccess'));

      onAddSuccess?.(selections);
    } catch (ex: any) {
      if (ex?.description) {
        message.error(ex.description);
      }
    }
  };

  return (
    <>
      {children ? (
        <span onClick={() => setShowAgentSelectionModal(true)}>{children}</span>
      ) : (
        <Button type="primary" className="dip-mb-14" onClick={() => setShowAgentSelectionModal(true)}>
          <PlusIcon />
          <span style={{ color: 'white' }}>{text || intl.get('dataAgent.config.add')}</span>
        </Button>
      )}

      {showAgentSelectionModal && (
        <AgentSelectionModal
          onCancel={() => setShowAgentSelectionModal(false)}
          onConfirm={data => {
            setShowAgentSelectionModal(false);
            addAgent(data);
          }}
        />
      )}
    </>
  );
};

export default SpaceAgentAddButton;
