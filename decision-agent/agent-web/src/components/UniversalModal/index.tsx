import React, { useEffect } from 'react';
import type { ModalProps } from 'antd';
import { Modal } from 'antd';
import classNames from 'classnames';
import Header from './Header';
import Footer, { FooterModalSourceType } from './Footer';
import './style.less';
import { useWindowSize, useMicroWidgetProps } from '@/hooks';

export interface TemplateModalProps extends ModalProps {
  children?: React.ReactNode;
  title?: string | React.ReactNode; // 标题
  footerExtra?: React.ReactNode; // 底部左侧额外元素
  isDisabled?: boolean; // 按钮是否灰置
  footerData?: FooterModalSourceType[] | React.ReactNode; // 底部元素
  fullScreen?: boolean; // 是否全屏 默认 false
  adaptive?: boolean; // 全屏条件下是否自适应高度 默认 false
}

/**
 * 常用于 新建、编辑、添加数据 等操作的弹窗模板, 包含标题和底部按钮
 * `title`和`footer`为`null`则不渲染
 */
const UniversalModal = (props: TemplateModalProps) => {
  const {
    className,
    children,
    title = null,
    okText,
    cancelText,
    footerExtra,
    isDisabled,
    width = 640,
    footerData,
    fullScreen = false,
    adaptive = false,
    ...reset
  } = props;

  const windowSize = useWindowSize();
  const microWidgetProps = useMicroWidgetProps();

  useEffect(() => {
    if (fullScreen && adaptive) {
      const headerDOM = document.querySelector('.ad-universal-modal-header')!;
      const contentDOM: HTMLDivElement = document.querySelector('.ad-universal-modal-content')!;
      const footerDOM = document.querySelector('.ad-universal-modal-footer');
      let height = windowSize.height - 48 - headerDOM.clientHeight;
      if (footerDOM) {
        height -= footerDOM.clientHeight;
      }
      contentDOM.style.maxHeight = `${height}px`;
    }
  }, [windowSize.height, footerData]);

  return (
    <Modal
      wrapClassName={classNames({
        'ad-modal-fullScreen': fullScreen,
      })}
      className={classNames('ad-universal-modal', className, {
        'ad-modal-adaptive': fullScreen && adaptive,
      })}
      classNames={{
        content: 'ad-universal-modal-custom-content',
      }}
      focusTriggerAfterClose={false}
      destroyOnClose
      maskClosable={false}
      width={fullScreen && !width ? '100%' : width}
      footer={null}
      getContainer={() => microWidgetProps?.container}
      {...reset}
    >
      <Header visible={!!title} title={title}></Header>

      <div className="ad-universal-modal-content">{children}</div>

      <Footer
        visible={footerData !== null || JSON.stringify(footerData) !== '{}'}
        source={footerData}
        footerExtra={footerExtra}
      />
    </Modal>
  );
};

UniversalModal.displayName = 'UniversalModal';
UniversalModal.Footer = Footer;
export default UniversalModal;
