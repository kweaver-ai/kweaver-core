import React from 'react';
import classnames from 'classnames';
import { Input as AntdInput } from 'antd';

import { InputType } from '../type';

import './style.less';

const SIZE = {
  large: 'ad-format-input-large',
  middle: 'ad-format-input-middle',
  small: 'ad-format-input-small',
};

const Input = (props: InputType) => {
  const { size = 'middle', children, className, ...othersProps } = props;
  const extendSize = SIZE[size as keyof typeof SIZE] || '';
  return <AntdInput className={classnames('ad-format-input', extendSize, className)} {...othersProps} />;
};

export default Input;
