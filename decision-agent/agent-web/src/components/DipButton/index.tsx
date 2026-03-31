import React from 'react';
import { Button } from 'antd';
import type { ButtonProps } from 'antd';
import classNames from 'classnames';
import './index.less';

const DipButton: React.FC<ButtonProps> = ({ className, ...restProps }) => {
  return <Button className={classNames('dip-button', className)} {...restProps} />;
};

export default DipButton;
