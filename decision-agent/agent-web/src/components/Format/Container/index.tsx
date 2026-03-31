import React from 'react';
import classnames from 'classnames';

import { ContainerType } from '../type';

const Container = (props: ContainerType) => {
  const { children, className, span } = props;
  return (
    <div className={classnames('containerRoot', className)} style={{ padding: span }}>
      {children}
    </div>
  );
};

export default Container;
