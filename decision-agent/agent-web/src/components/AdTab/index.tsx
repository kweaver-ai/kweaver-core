import React from 'react';
import classnames from 'classnames';
import { Tabs, type TabsProps } from 'antd';
import './style.less';

export interface AdTabType extends TabsProps {
  className?: string;
}

const AdTab: React.FC<AdTabType> = props => {
  const { className, ...restTabsProps } = props;
  return <Tabs className={classnames(className, 'ad-tab dip-w-100 dip-h-100')} {...restTabsProps} />;
};

export default AdTab;
