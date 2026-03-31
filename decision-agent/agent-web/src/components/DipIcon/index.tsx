import React, { type CSSProperties } from 'react';
import { createFromIconfontCN } from '@ant-design/icons';
import classNames from 'classnames';
import './index.less';
import type { IconFontProps } from '@ant-design/icons/es/components/IconFont';
import '@/assets/fonts/iconfont.js';
import '@/assets/fonts/color-iconfont.js';

const IconBaseComponent = createFromIconfontCN({
  scriptUrl: [],
});

export interface IconFontType extends IconFontProps {
  className?: string;
  style?: CSSProperties;
}

const DipIcon: React.FC<IconFontType> = props => {
  const { className, ...restProps } = props;
  const prefixCls = 'dip-icon';
  return <IconBaseComponent className={classNames(prefixCls, className)} {...restProps} />;
};

export default DipIcon;
