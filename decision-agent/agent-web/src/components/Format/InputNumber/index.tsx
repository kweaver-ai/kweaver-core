import React, { useEffect, useState } from 'react';
import classnames from 'classnames';
import { Input as AntInputNumber } from 'antd';

import { InputType } from '../type';

import './style.less';

const InputNumber = (props: InputType) => {
  const { min, max, size = 'middle', value, className, ...othersProps } = props;

  const [number, setNumber] = useState<any>(null);
  useEffect(() => {
    setNumber(value);
  }, [value]);

  const onChangeNumber = (value: any) => {
    if (!value) {
      setNumber(value);
      if (othersProps.onChange) othersProps.onChange(value);
      return;
    }
    const reg = /^\d+(\.\d+)?$/;
    if (reg.test(value)) {
      const num: any = parseInt(value);
      if (max) {
        if (num <= max) {
          setNumber(num);
          if (othersProps.onChange) othersProps.onChange(num);
        } else {
          setNumber(max);
          if (othersProps.onChange) othersProps.onChange(max);
        }
      }
    }
  };
  const onInputBlur = (e: any) => {
    if (!number && min) setNumber(min);
    if (othersProps.onBlur) othersProps.onBlur(e);
  };

  return (
    <AntInputNumber
      className={classnames(className)}
      {...othersProps}
      value={number}
      onBlur={onInputBlur}
      onChange={onChangeNumber}
    />
  );
};

export default InputNumber;
