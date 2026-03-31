import React from 'react';
import classnames from 'classnames';
import { Select as AntdSelect } from 'antd';

import { SelectType } from '../type';

import './style.less';

const SIZE = {
  large: 'ad-format-select-large',
  middle: 'ad-format-select-middle',
  small: 'ad-format-select-small',
};

const Select = (props: SelectType) => {
  const { size = 'large', children, className, ...othersProps } = props;
  const extendSize = SIZE[size as keyof typeof SIZE] || '';
  return (
    <AntdSelect className={classnames('ad-format-select', extendSize, className)} {...othersProps}>
      {children}
    </AntdSelect>
  );
};

export default Select;
