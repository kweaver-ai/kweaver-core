import DipModal from '@/components/DipModal';
import intl from 'react-intl-universal';
import { Select } from 'antd';
import React, { useEffect, useState } from 'react';
import { getKnExperimentList } from '@/apis/knowledge-data';
import _ from 'lodash';
import { useDeepCompareMemo } from '@/hooks';

const KnowledgeNetworkExperimentalSelect = ({ onClose, handleSaveKNExperiment }: any) => {
  const [netList, setNetList] = useState([]);
  const [selectedNet, setSelectedNet] = useState<any>();
  useEffect(() => {
    getNetList();
  }, []);
  const getNetList = async () => {
    const res = await getKnExperimentList();
    if (res) {
      const list = _.get(res, 'entries') || [];
      setNetList(list);
    }
  };
  const netOptions = useDeepCompareMemo(() => {
    return netList.map((item: any) => {
      return {
        label: item.name,
        value: item.id,
      };
    });
  }, [netList]);
  return (
    <DipModal
      width={500}
      centered
      title={intl.get('dataAgent.config.selectKnowledgeNetwork')}
      open
      onCancel={onClose}
      okButtonProps={{ disabled: !selectedNet }}
      onOk={() => {
        handleSaveKNExperiment(selectedNet);
      }}
      footer={(_, { OkBtn, CancelBtn }) => (
        <>
          <OkBtn />
          <CancelBtn />
        </>
      )}
    >
      <Select
        showSearch
        placeholder={intl.get('dataAgent.config.pleaseSelectKnowledgeNetwork')}
        className="dip-w-100"
        options={netOptions}
        value={selectedNet}
        onChange={(value, option) => {
          setSelectedNet(option);
        }}
      />
    </DipModal>
  );
};

export default ({ open, ...restProps }: any) => {
  return open && <KnowledgeNetworkExperimentalSelect {...restProps} />;
};
