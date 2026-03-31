import React from 'react';
import styles from './index.module.less';
import { CloseOutlined } from '@ant-design/icons';
import { useDipChatStore } from '../../../store';
import classNames from 'classnames';
import ScrollBarContainer from '@/components/ScrollBarContainer';
import DipButton from '@/components/DipButton';
import { Badge, Modal } from 'antd';
import DipIcon from '@/components/DipIcon';
import { useDeepCompareMemo } from '@/hooks';
import { DipChatItemContentType } from '@/components/DipChat/interface';

const NetSearchList = ({ citesList }: any) => {
  const {
    dipChatStore: { chatList, activeChatItemIndex },
    closeSideBar,
  } = useDipChatStore();
  const [modal, contextHolder] = Modal.useModal();
  const chatItem = chatList[activeChatItemIndex];
  const activeContent: DipChatItemContentType = chatItem.content || {};
  const cites = citesList || activeContent?.cites || [];
  const total = useDeepCompareMemo(() => {
    let num: number = 0;
    cites.forEach((item: any) => {
      if ('children' in item) {
        if (item.children && Array.isArray(item.children)) {
          num += item.children.length;
        }
      } else {
        num += 1;
      }
    });
    return num;
  }, [cites]);

  const renderCites = (listData: any) => {
    return listData.map((item: any) => {
      if ('children' in item) {
        // 说明要渲染标题
        return (
          <div key={item.id} className={classNames(styles.item)}>
            <Badge status="processing" text={<span className="dip-font-weight-700">{item.title}</span>} />
            <div>{renderCites(item.children)}</div>
          </div>
        );
      }
      return (
        <div className={styles.processChild} key={item.title}>
          <div className="dip-flex-align-center">
            {item.icon && <img className={classNames(styles.icon, 'dip-mr-8')} src={item.icon} alt="" />}
            <span
              onClick={() => {
                modal.confirm({
                  closable: true,
                  width: 424,
                  title: '您即将离开超级助手，跳转到其他网站',
                  content: '超级助手不会对其他网站的内容及其真实性负责，请注意上网安全，保护好您的个人信息及财产安全。',
                  onOk: () => {
                    window.open(item.link);
                  },
                  okText: '继续访问',
                  footer: (_, { OkBtn, CancelBtn }) => (
                    <>
                      <OkBtn />
                      <CancelBtn />
                    </>
                  ),
                });
              }}
              className={classNames(styles.link, 'dip-flex-item-full-width dip-ellipsis')}
              title={item.title}
            >
              {item.title}
            </span>
          </div>
          <div className="dip-mt-8 dip-text-color-45 dip-ellipsis-2">{item.content}</div>
        </div>
      );
    });
  };

  return (
    <div className={classNames(styles.container, 'dip-full dip-pt-20 dip-flex-column')}>
      <div className="dip-flex-space-between dip-pl-24 dip-pr-24">
        <span className="dip-flex-align-center dip-font-16 dip-text-color dip-font-weight-700">
          <DipIcon style={{ fontSize: 26 }} type="icon-dip-net" />
          <span className="dip-ml-8 dip-flex-item-full-width dip-ellipsis dip-font-16 dip-text-color dip-font-weight-700">
            网络搜索
            <span className="dip-ml-8">{total}</span>
          </span>
        </span>
        <DipButton
          size="small"
          type="text"
          onClick={() => {
            closeSideBar();
          }}
        >
          <CloseOutlined className="dip-text-color-45 dip-font-16" />
        </DipButton>
      </div>
      <ScrollBarContainer className="dip-flex-item-full-height dip-pl-24 dip-pr-24 dip-pb-24">
        {renderCites(cites)}
      </ScrollBarContainer>
      {contextHolder}
    </div>
  );
};

export default NetSearchList;
