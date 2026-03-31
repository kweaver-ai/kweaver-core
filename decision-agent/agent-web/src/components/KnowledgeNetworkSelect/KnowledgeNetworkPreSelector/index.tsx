import React, { useState } from 'react';
import intl from 'react-intl-universal';
import { Modal } from 'antd';
import type { TreeDataNode } from 'antd';
import NetworkSelect from '../NetworkSelect';

interface KnowledgeNetworkPreSelectorProps {
  treeData: TreeDataNode[];
  onCancel: () => void;
  onConfirm: (networkId: string, newWorkName: string, treeData: TreeDataNode[]) => void;
}

const KnowledgeNetworkPreSelector: React.FC<KnowledgeNetworkPreSelectorProps> = ({ onCancel, onConfirm }) => {
  const [selectedNetwork, setSelectedNetwork] = useState<any>(undefined);
  const [treeData, setTreeData] = useState<TreeDataNode[]>([]);

  const handleConfirm = () => {
    onConfirm(selectedNetwork.value, selectedNetwork.label, treeData);
  };

  const handleCancel = () => {
    onCancel();
  };

  const handleSelectChange = (value: string, label: string) => {
    setSelectedNetwork({ value, label });
  };

  return (
    <Modal
      title={intl.get('dataAgent.config.selectKnowledgeNetwork')}
      open
      centered
      maskClosable={false}
      width={500}
      footer={(_, { OkBtn, CancelBtn }) => (
        <>
          <OkBtn />
          <CancelBtn />
        </>
      )}
      okButtonProps={{ disabled: !selectedNetwork, className: 'dip-min-width-72' }}
      cancelButtonProps={{ className: 'dip-min-width-72' }}
      onOk={handleConfirm}
      onCancel={handleCancel}
    >
      <div className="dip-pt-16 dip-pb-16">
        <NetworkSelect
          networkId={selectedNetwork}
          networkTreeData={treeData}
          placeholder={intl.get('dataAgent.config.pleaseSelectKnowledgeNetwork')}
          onChange={handleSelectChange}
          onUpdateTreeData={setTreeData}
        />
      </div>
    </Modal>
  );
};

export default KnowledgeNetworkPreSelector;
