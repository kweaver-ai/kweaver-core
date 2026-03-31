import React, { useEffect, useRef, useState } from 'react';
import { Popover } from 'antd';
import './style.less';
import classNames from 'classnames';

export type ErrorTipProps = {
  errorText?: string;
  getPopupContainer?: any;
  className?: string;
};

const ErrorTip = (props: React.PropsWithChildren<ErrorTipProps>) => {
  const { errorText, children, getPopupContainer, className } = props;
  const isFocus = useRef(false);
  const [visible, setVisible] = useState(!!errorText);

  useEffect(() => {
    if (!isFocus.current) return;
    setVisible(!!errorText);
  }, [errorText]);

  /**
   * 劫持focus
   */
  const handleFocus = (e: any) => {
    isFocus.current = true;
    errorText && setVisible(true);
  };

  /**
   * 劫持blur
   */
  const handleBlur = (e: any) => {
    isFocus.current = false;
  };

  /**
   * 开关控制
   */
  const onVisibleChange = (isOpen: boolean) => {
    setVisible(isOpen);
  };

  return (
    <Popover
      style={{ width: '100%' }}
      overlayClassName="onto-pro-error-tip"
      trigger={['hover']}
      content={errorText}
      placement="bottomLeft"
      destroyTooltipOnHide
      open={visible}
      getPopupContainer={getPopupContainer}
      onOpenChange={onVisibleChange}
    >
      <div className={classNames(className, { 'ad-error-tip': !!errorText })} onFocus={handleFocus} onBlur={handleBlur}>
        {children}
      </div>
    </Popover>
  );
};

export default ErrorTip;
