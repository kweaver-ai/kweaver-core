import DipModal from '@/components/DipModal';
import { useEffect, useRef, useState } from 'react';
import { getConversationList, markReadConversation } from '@/apis/super-assistant';
import { handleConversationGroup } from '@/components/DipChat/assistant';
import styles from './index.module.less';
import classNames from 'classnames';
import SearchInput from '@/components/SearchInput';
import ScrollBarContainer from '@/components/ScrollBarContainer';
import type { ConversationItemType, DipChatItem } from '@/components/DipChat/interface';
import { Button, Spin, Splitter } from 'antd';
import { LoadingOutlined } from '@ant-design/icons';
import { useDipChatStore } from '@/components/DipChat/store';
import { useDeepCompareMemo } from '@/hooks';
import NoSearchResult from '@/components/NoSearchResult';
import BubbleList from '../BubbleList';
import GradientContainer from '@/components/GradientContainer';
import ClickView from '@/assets/icons/clickView.svg';
import ConversationItem from './ConversationItem';
import intl from 'react-intl-universal';

const ConversationListModal = ({ onClose, agentAppKey, startNewConversation }: any) => {
  const {
    dipChatStore: { activeConversationKey, chatList },
    cancelChat,
    setDipChatStore,
    resetDipChatStore,
    getConversationDetailsByKey,
    sendChat,
  } = useDipChatStore();
  const timer = useRef<any>(null);
  const [conversationList, setConversationList] = useState<any[]>([]);
  const [activeKey, setActiveKey] = useState('');
  const [searchValue, setSearchValue] = useState('');
  const [loading, setLoading] = useState(true);
  const cacheChatList = useRef<DipChatItem[]>([]);

  useEffect(() => {
    cacheChatList.current = chatList;
    setActiveKey(activeConversationKey);
    getList();
    return () => {
      if (timer.current) {
        clearTimeout(timer.current);
      }
    };
  }, []);

  useEffect(() => {
    if (activeKey) {
      getDetailsData();
    }
  }, [activeKey]);

  const getList = async () => {
    const res: any = await getConversationList(agentAppKey, { size: 1000 });
    setLoading(false);
    if (res) {
      const { entries } = res;
      setConversationList(entries);
      const hasProcessing = entries.some((item: any) => item.status === 'processing');
      if (hasProcessing) {
        if (timer.current) {
          clearTimeout(timer.current);
        }
        timer.current = setTimeout(() => {
          getList();
        }, 2000);
      } else {
        if (timer.current) {
          clearTimeout(timer.current);
          timer.current = null;
        }
      }
    }
  };

  const getDetailsData = async () => {
    const res = await getConversationDetailsByKey(activeKey);
    if (res) {
      const { chatList, read_message_index, message_index } = res;
      // console.log('getDetailsData', chatList);
      setDipChatStore({ chatList });
      if (read_message_index !== message_index) {
        // 标记会话已读
        await markReadConversation(agentAppKey, activeKey, message_index);
        getList();
      }
    }
  };

  const goToConversation = async () => {
    cancelChat();
    const url = new URL(window.location.href);
    url.searchParams.set('conversation_id', activeKey);
    // 使用history API更新URL而不刷新页面
    window.history.pushState({}, '', url.toString());
    setDipChatStore({
      activeConversationKey: activeKey,
    });
    resetDipChatStore(['activeChatItemIndex', 'chatListAutoScroll', 'activeProgressIndex', 'streamGenerating']);
    const res: any = await getConversationDetailsByKey(activeKey);
    if (res) {
      const { recoverConversation, chatList, read_message_index, message_index } = res;
      if (recoverConversation) {
        setDipChatStore({
          activeConversationKey: activeKey,
        });
        sendChat({
          chatList,
          body: { conversation_id: activeKey, interruptAction: 'confirm' },
          recoverConversation: true,
        });
      } else {
        setDipChatStore({ chatList });
        if (read_message_index !== message_index) {
          // 标记会话已读
          await markReadConversation(agentAppKey, activeKey, message_index);
        }
      }
      onClose();
    }
  };

  const data = useDeepCompareMemo(() => {
    let filterList = conversationList;
    if (searchValue) {
      filterList = conversationList.filter(item => item.title?.includes(searchValue));
    }
    const groupData = handleConversationGroup(filterList);
    const list: any = groupData.filter(ii => ii.children.length > 0);
    return list;
  }, [conversationList, searchValue]);

  const renderItems = (data: ConversationItemType[]) => {
    if (data.length === 0 && searchValue) {
      return (
        <div className="dip-full dip-column-center">
          <div style={{ transform: 'translateY(-50%)' }}>
            <NoSearchResult />
          </div>
        </div>
      );
    }
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
        <ConversationItem
          onClick={() => {
            if (item.key !== activeConversationKey) {
              setActiveKey(item.key);
            }
          }}
          className={classNames({
            [styles.active]: item.key === activeKey,
          })}
          key={item.key}
          id={item.key}
          timestamp={item.timestamp}
          activeKey={activeKey}
          unRead={item.unRead}
          label={item.label}
          agentAppKey={agentAppKey}
          status={item.status}
          refreshList={getList}
          onClearSelectedConversation={() => {
            if (item.key === activeKey) {
              // 说明删除的是当前弹框选中的会话
              setDipChatStore({ chatList: [] });
              setActiveKey('');
            }
            if (item.key === activeConversationKey) {
              // 说明删除的是外面正在选中的会话
              startNewConversation();
              cacheChatList.current = [];
            }
          }}
        />
      );
    });
  };

  const renderContent = () => {
    if (loading) {
      return (
        <div className="dip-full dip-flex-center">
          <Spin size="small" indicator={<LoadingOutlined style={{ fontSize: 16 }} spin />} />
        </div>
      );
    }
    return (
      <Splitter>
        <Splitter.Panel defaultSize="30%" max="60%" min="20%">
          <div className="dip-full dip-flex-column">
            <div className="dip-pl-12 dip-pr-12">
              <SearchInput
                autoWidth
                onChange={(e: any) => {
                  setSearchValue(e.target.value);
                }}
                placeholder={intl.get('dipChat.searchConversationPlaceholder')}
              />
            </div>
            <ScrollBarContainer className="dip-flex-item-full-height dip-pl-12 dip-pr-12">
              {renderItems(data)}
            </ScrollBarContainer>
          </div>
        </Splitter.Panel>
        <Splitter.Panel>
          {activeKey ? (
            <div className="dip-full dip-flex-column">
              <GradientContainer className="dip-flex-item-full-height">
                <ScrollBarContainer className="dip-full dip-p-24">
                  <BubbleList readOnly={true} />
                </ScrollBarContainer>
              </GradientContainer>
              <div className="dip-flex-content-end dip-p-8">
                <Button variant="filled" color="primary" onClick={goToConversation}>
                  {intl.get('dipChat.goTo')}
                </Button>
              </div>
            </div>
          ) : (
            <div className="dip-full dip-flex-center">
              <div className="dip-column-center" style={{ transform: 'translateY(-50%)' }}>
                <ClickView />
                <div className="dip-text-color-65">{intl.get('dipChat.selectConversationToPreview')}</div>
              </div>
            </div>
          )}
        </Splitter.Panel>
      </Splitter>
    );
  };

  return (
    <DipModal
      footer={false}
      fullScreen
      open
      title={intl.get('dipChat.historyConversations')}
      onCancel={() => {
        onClose();
        setDipChatStore({ chatList: cacheChatList.current });
      }}
      rootClassName={styles.modal}
    >
      {renderContent()}
    </DipModal>
  );
};

export default ({ open, ...restProps }: any) => {
  return open && <ConversationListModal {...restProps} />;
};
