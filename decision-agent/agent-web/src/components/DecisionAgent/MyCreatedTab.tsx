import { memo } from 'react';
import intl from 'react-intl-universal';
import { Tabs } from 'antd';
import { ModeEnum } from './types';

interface MyCreatedTabProps {
  activeKey?: ModeEnum.MyAgent | ModeEnum.MyTemplate;
  onChange: (key: string) => void;
}

const MyCreatedTab = ({ activeKey = ModeEnum.MyAgent, onChange }: MyCreatedTabProps) => {
  return (
    <Tabs
      style={{ marginBottom: '-16px' }}
      items={[
        {
          key: ModeEnum.MyAgent,
          label: intl.get('dataAgent.agent'),
        },
        {
          key: ModeEnum.MyTemplate,
          label: intl.get('dataAgent.myTemplate'),
        },
      ]}
      activeKey={activeKey}
      onChange={onChange}
    />
  );
};

export default memo(MyCreatedTab);
