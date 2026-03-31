import React, { useEffect } from 'react';
import { Modal } from 'antd';
import type { ModalProps } from 'antd';
import './index.less';
import classNames from 'classnames';
import { useWindowSize } from '@/hooks';

export interface DipModalProps extends ModalProps {
  fullScreen?: boolean; // 是否全屏 默认 false
  adaptive?: boolean; // 全屏条件下是否自适应高度 默认 false
}

const DipModal: React.FC<DipModalProps> = props => {
  const { className, fullScreen = false, adaptive = false, width, ...restProps } = props;
  const windowSize = useWindowSize();
  useEffect(() => {
    if (fullScreen && adaptive) {
      const headerDOM = document.querySelector('.ant-modal-header')!;
      const contentDOM: HTMLDivElement = document.querySelector('.ant-modal-body')!;
      const footerDOM = document.querySelector('.ant-modal-footer');
      let height = windowSize.height - 48 - headerDOM.clientHeight;
      if (footerDOM) {
        height -= footerDOM.clientHeight;
      }
      contentDOM.style.maxHeight = `${height}px`;
    }
  }, [windowSize.height]);
  const prefixCls = 'dip-modal';
  return (
    <Modal
      className={classNames(prefixCls, className, {
        [`${prefixCls}-adaptive`]: fullScreen && adaptive,
      })}
      wrapClassName={classNames(`${prefixCls}-wrapper`, {
        [`${prefixCls}-fullScreen`]: fullScreen,
      })}
      maskClosable={false}
      width={fullScreen && !width ? '100%' : width}
      {...restProps}
    />
  );
};

export default DipModal;
