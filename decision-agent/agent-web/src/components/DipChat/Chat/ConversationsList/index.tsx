import styles from './index.module.less';
import { useEffect, useRef, useState } from 'react';
import { Button, Dropdown, message, Modal, Popover, Spin, Tooltip } from 'antd';
import {
  ClockCircleOutlined,
  CloseCircleFilled,
  EditOutlined,
  LeftOutlined,
  LoadingOutlined,
  MessageOutlined,
} from '@ant-design/icons';
import classNames from 'classnames';
import { useDipChatStore } from '@/components/DipChat/store';
import { useNavigate, useSearchParams } from 'react-router-dom';
import DipIcon from '@/components/DipIcon';
import DipButton from '@/components/DipButton';
import { deleteConversationById, markReadConversation } from '@/apis/super-assistant';
import ScrollBarContainer from '@/components/ScrollBarContainer';
import type { ConversationItemType } from '@/components/DipChat/interface';
import EllipsisOutlined from '@/assets/icons/ellipsis.svg';
import EditNameModal from './EditNameModal';
import { getParam } from '@/utils/handle-function';
import { scrollIntoViewForContainer } from '@/utils/handle-function';
import intl from 'react-intl-universal';
import { useMicroWidgetProps } from '@/hooks';
import ConversationListModal from '../ConversationListModal';
import dayjs from 'dayjs';

const ConversationList = ({ startNewConversation, className }: any) => {
  const microWidgetProps = useMicroWidgetProps();
  const {
    dipChatStore: {
      conversationItems,
      conversationItemsTotal,
      activeConversationKey,
      agentAppKey,
      conversationListModalOpen,
      tempAreaOpen,
    },
    setDipChatStore,
    resetDipChatStore,
    getConversationData,
    sendChat,
    cancelChat,
    getConversationDetailsByKey,
  } = useDipChatStore();
  const [modal, contextHolder] = Modal.useModal();
  const [messageApi, messageContextHolder] = message.useMessage();
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const [nameModal, setNameModal] = useState({ open: false, name: '', key: '' });
  const scrollContainer = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    getConversationData();
  }, []);

  useEffect(() => {
    const key = getParam('conversation_id');
    if (key) {
      getDetailsById(key);
    }
  }, []);

  useEffect(() => {
    if (activeConversationKey && scrollContainer.current) {
      const ele = document.getElementById(activeConversationKey);
      scrollIntoViewForContainer(scrollContainer.current!, ele!);
    }
  }, [activeConversationKey]);

  const deleteConversation = async (id: string) => {
    const confirmed = await modal.confirm({
      title: intl.get('dipChat.permanentDeleteConversation'),
      content: intl.get('dipChat.deleteConversationConfirm'),
      okButtonProps: {
        danger: true,
      },
      okText: intl.get('dipChat.delete'),
    });
    if (confirmed) {
      const res = await deleteConversationById(agentAppKey, id);
      if (res) {
        messageApi.success(intl.get('dipChat.deleteSuccess'));
        getConversationData();
        if (id === activeConversationKey) {
          startNewConversation();
        }
      }
    }
  };

  const getDetailsById = async (id: string) => {
    const res: any = await getConversationDetailsByKey(id);
    if (res) {
      const { recoverConversation, chatList, read_message_index, message_index, conversationLoading } = res;
      if (recoverConversation) {
        setDipChatStore({
          activeConversationKey: id,
        });
        sendChat({
          chatList,
          body: { conversation_id: id, interruptAction: 'confirm' },
          recoverConversation: true,
        });
      } else {
        setDipChatStore({ chatList, streamGenerating: conversationLoading });
        if (read_message_index !== message_index) {
          // 标记会话已读
          await markReadConversation(agentAppKey, id, message_index);
          getConversationData();
        }
      }
    }
  };

  const renderItems = (data: ConversationItemType[]) => {
    return data.map(item => {
      if (item.children && item.children.length > 0) {
        return (
          <div className={styles.group} key={item.key}>
            <div className="dip-mb-10 dip-text-color-45">{item.label}</div>
            {renderItems(item.children)}
          </div>
        );
      }
      return (
        <div
          onClick={() => {
            if (item.key !== activeConversationKey) {
              cancelChat();
              // 此处用 setTimeout 的目的是：要等cancelChat涉及的状态全部执行完毕
              setTimeout(() => {
                console.log('取消会话了哈哈哈哈 ++++ cancelChat');
                const url = new URL(window.location.href);
                url.searchParams.set('conversation_id', item.key);
                // 使用history API更新URL而不刷新页面
                window.history.pushState({}, '', url.toString());
                setDipChatStore({
                  activeConversationKey: item.key,
                });
                resetDipChatStore([
                  'activeChatItemIndex',
                  'chatListAutoScroll',
                  'activeProgressIndex',
                  'streamGenerating',
                ]);
                getDetailsById(item.key);
              }, 0);
            }
          }}
          className={classNames(styles.item, 'dip-flex-align-center', {
            [styles.active]: item.key === activeConversationKey,
          })}
          key={item.key}
          id={item.key}
        >
          {item.unRead && <div className={styles.unRead} />}
          <MessageOutlined className="dip-text-color-45" />
          <div
            className={classNames('dip-ml-8 dip-flex-item-full-width dip-ellipsis dip-text-color-85')}
            title={item.label}
          >
            {item.label}
          </div>
          {item.status === 'processing' && (
            <Tooltip title={intl.get('dipChat.taskInProgress')}>
              <Spin className="dip-ml-8" size="small" indicator={<LoadingOutlined style={{ fontSize: 14 }} spin />} />
            </Tooltip>
          )}
          {item.status === 'failed' && <CloseCircleFilled className="dip-text-color-error" />}
          <Dropdown
            trigger={['click']}
            menu={{
              items: [
                { label: intl.get('dipChat.edit'), key: 'edit', icon: <EditOutlined /> },
                { label: intl.get('dipChat.delete'), key: 'delete', icon: <DipIcon type="icon-dip-trash" /> },
              ],
              onClick: ({ key, domEvent }) => {
                domEvent.stopPropagation();
                if (key === 'edit') {
                  setNameModal(prevState => ({
                    ...prevState,
                    open: true,
                    name: item.label,
                    key: item.key,
                  }));
                }
                if (key === 'delete') {
                  deleteConversation(item.key);
                }
              },
            }}
          >
            <Button
              onClick={e => e.stopPropagation()}
              className={classNames(styles.btn)}
              size="small"
              type="text"
              icon={<EllipsisOutlined />}
            />
          </Dropdown>
          <Popover
            placement="right"
            content={<div className="dip-p-12">更新时间：{dayjs(item.timestamp).format('YYYY-MM-DD HH:mm:ss')}</div>}
          >
            <Button size="small" type="text" className={classNames(styles.btn)} icon={<ClockCircleOutlined />} />
          </Popover>
        </div>
      );
    });
  };

  const renderContent = () => {
    return (
      <>
        <div className="dip-pt-12 dip-pl-8 dip-pr-16 dip-flex-space-between">
          <DipButton
            color="default"
            variant="link"
            icon={<LeftOutlined />}
            onClick={() => {
              // 优先使用url的redirect进行跳转-山东大数据局
              const redirectUrl = searchParams.get('redirect');
              if (redirectUrl) {
                location.replace(redirectUrl);
              } else {
                const preRoute = searchParams.get('preRoute');
                const preRouteIsMicroApp = searchParams.get('preRouteIsMicroApp');
                // 说明之前的路由是其他微应用的一个页面
                if (preRouteIsMicroApp === 'true') {
                  microWidgetProps.navigate(preRoute ?? '/studio/home');
                } else {
                  let url = preRoute ?? '/';
                  const filterParams = getParam('filterParams');
                  if (filterParams) {
                    url += `?filterParams=${filterParams}`;
                  }
                  navigate(url);
                }
              }
            }}
            size="small"
          >
            {intl.get('dipChat.return')}
          </DipButton>
          <span>
            <Tooltip
              fresh
              getTooltipContainer={node => node.parentElement!}
              placement="bottom"
              title={tempAreaOpen ? '收起临时区' : '展开临时区'}
            >
              <DipButton
                variant="link"
                color="default"
                onClick={() =>
                  setDipChatStore({
                    tempAreaOpen: !tempAreaOpen,
                  })
                }
              >
                <DipIcon type="icon-dip-tempArea" className="dip-font-16" />
              </DipButton>
            </Tooltip>
            <Tooltip fresh getTooltipContainer={node => node.parentElement!} placement="bottom" title={'收起侧边栏'}>
              <DipButton
                className="dip-ml-8"
                variant="link"
                color="default"
                onClick={() =>
                  setDipChatStore({
                    conversationCollapsed: true,
                  })
                }
              >
                <DipIcon type="icon-dip-cebianlan" className="dip-font-16" />
              </DipButton>
            </Tooltip>
          </span>
        </div>
        <div className="dip-mt-8 dip-pl-8 dip-pr-8">
          <div
            onClick={startNewConversation}
            className={classNames(styles.newConversation, 'dip-flex-align-center dip-border')}
          >
            <DipButton icon={<DipIcon type="icon-dip-chat1" className="dip-font-16" />} type="text">
              {intl.get('dipChat.startNewConversation')}
            </DipButton>
          </div>
        </div>

        <ScrollBarContainer ref={scrollContainer} className="dip-flex-item-full-height dip-pl-8 dip-pr-8">
          {renderItems(conversationItems)}
          {conversationItemsTotal > 10 && (
            <DipButton
              onClick={() => {
                setDipChatStore({
                  conversationListModalOpen: true,
                });
              }}
              type="text"
              // className="dip-text-color"
            >
              {intl.get('dipChat.viewAll')}
            </DipButton>
          )}
        </ScrollBarContainer>
      </>
    );
  };
  return (
    <div className={classNames(styles.container, 'dip-full dip-flex-column', className)}>
      {renderContent()}
      {contextHolder}
      {messageContextHolder}
      <EditNameModal
        conversationKey={nameModal.key}
        open={nameModal.open}
        name={nameModal.name}
        onClose={() => {
          setNameModal(prevState => ({
            ...prevState,
            open: false,
          }));
        }}
        refreshList={getConversationData}
      />
      <ConversationListModal
        agentAppKey={agentAppKey}
        open={conversationListModalOpen}
        onClose={() => {
          setDipChatStore({
            conversationListModalOpen: false,
          });
          getConversationData();
        }}
        startNewConversation={startNewConversation}
      />
    </div>
  );
};

export default ConversationList;
