import React, { type CSSProperties } from 'react';
import './index.less';
import DipLogo from '@/assets/icons/logo.svg';
import classNames from 'classnames';
type LogoProps = {
  style?: CSSProperties;
  loading?: boolean;
};
const Logo: React.FC<LogoProps> = ({ style = { width: 16, height: 16 }, loading = false }) => {
  const prefixCls = 'dip-logo';
  return (
    <div
      className={classNames(prefixCls, {
        [`${prefixCls}-loading`]: loading,
      })}
    >
      <DipLogo style={style} />
    </div>
  );
};

export default Logo;
