import React, { useEffect, useRef, useState } from 'react';
import styles from './index.module.less';
import classNames from 'classnames';
import AiInput from '../components/AiInput';
import GradientContainer from '../../../components/GradientContainer';
import BubbleList from './BubbleList';
import ConversationList from './ConversationsList';
import ResizeObserver from '@/components/ResizeObserver';
import { useDipChatStore } from '../store';
import _ from 'lodash';
import { nanoid } from 'nanoid';
import type { AiInputRef, AiInputValue } from '../components/AiInput/interface';
import ScrollBarContainer from '@/components/ScrollBarContainer';
import { getAgentInputDisplayFields } from '../utils';
import DipIcon from '@/components/DipIcon';
import DipButton from '@/components/DipButton';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { ArrowDownOutlined, LeftOutlined, MenuFoldOutlined, MenuUnfoldOutlined } from '@ant-design/icons';
import { Button, Splitter, Tooltip } from 'antd';
import TempArea from './TempArea';
import RightSideBar from './RightSideBar';
import { useMicroWidgetProps } from '@/hooks';
import AgentDescription from './AgentDescription';
import type { ChatBody } from '../interface';
import DebuggerProcess from './DebuggerProcess';
import AgentInputParamDrawer from './AgentInputParamDrawer';
import FilePreview from '../components/FilePreview';
import ColorLoading from '@/components/DipChat/components/ColorLoading';
import intl from 'react-intl-universal';
import { getParam } from '@/utils/handle-function';
import dayjs from 'dayjs';

const DipChat = () => {
  const microWidgetProps = useMicroWidgetProps();
  const {
    dipChatStore: {
      chatList,
      chatListAutoScroll,
      aiInputValue,
      streamGenerating,
      conversationCollapsed,
      agentAppType,
      agentDetails,
      debug,
      showDebuggerArea,
      agentInputParamsFormErrorFields,
      agentInputParamForm,
      previewFile,
      tempFileList,
      tempAreaOpen,
    },
    setDipChatStore,
    getDipChatStore,
    sendChat,
    stopChat,
    cancelChat,
    resetDipChatStore,
  } = useDipChatStore();
  const conversationWrapperRef = useRef<HTMLDivElement>(null);
  const collapseBtnWrapperRef = useRef<HTMLDivElement>(null);
  const chatWrapperRef = useRef<HTMLDivElement>(null);
  const aiInputRef = useRef<AiInputRef | null>(null);
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const [showBackBottom, setShowBackBottom] = useState(false);
  const [size, setSize] = React.useState<number>(280);

  useEffect(() => {
    const handleScroll = () => {
      if (!chatWrapperRef.current) return;
      const { scrollTop, scrollHeight, clientHeight } = chatWrapperRef.current;
      const isNearBottom = scrollHeight - scrollTop - clientHeight > 50;
      setShowBackBottom(isNearBottom);
    };

    const chatWrapper = chatWrapperRef.current;
    if (chatWrapper) {
      chatWrapper.addEventListener('scroll', handleScroll);
    }
    return () => {
      if (chatWrapper) {
        chatWrapper.removeEventListener('scroll', handleScroll);
      }
    };
  }, [showDebuggerArea]);
  useEffect(() => {
    return () => {
      cancelChat();
      resetDipChatStore([
        'activeConversationKey',
        'conversationItems',
        'chatList',
        'activeChatItemIndex',
        'streamGenerating',
        'conversationCollapsed',
        'activeProgressIndex',
      ]);
    };
  }, []);

  const onSubmit = (value: AiInputValue) => {
    const cloneChatList = _.cloneDeep(chatList);
    cloneChatList.push({
      key: nanoid(),
      role: 'user',
      content: value.inputValue,
      loading: false,
      updateTime: dayjs().valueOf(),
    });
    cloneChatList.push({
      key: nanoid(),
      role: 'common',
      content: '',
      loading: true,
    });
    const body: ChatBody = { query: value.inputValue };
    sendChat({
      chatList: cloneChatList,
      body,
      activeChatItemIndex: -1,
    });
  };

  const renderCollapseBtn = () => {
    if (conversationCollapsed) {
      return (
        <div className={classNames(styles.collapseBtn)} ref={collapseBtnWrapperRef}>
          <Tooltip
            title={intl.get('dipChat.return')}
            getTooltipContainer={node => node.parentElement!}
            placement="right"
          >
            <DipButton
              color="default"
              variant="link"
              icon={<LeftOutlined className="dip-font-16" />}
              onClick={() => {
                // 优先使用url的redirect进行跳转-山东大数据局
                const redirectUrl = searchParams.get('redirect');
                if (redirectUrl) {
                  location.replace(redirectUrl);
                } else {
                  const preRoute = searchParams.get('preRoute');
                  const preRouteIsMicroApp = searchParams.get('preRouteIsMicroApp');
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
            />
          </Tooltip>
          <div className={styles.divider} />
          <Tooltip
            title={intl.get('dipChat.expandSidebar')}
            getTooltipContainer={node => node.parentElement!}
            placement="right"
          >
            <DipButton
              variant="link"
              color="default"
              onClick={() =>
                setDipChatStore({
                  conversationCollapsed: !conversationCollapsed,
                })
              }
            >
              <DipIcon type="icon-dip-cebianlan" className="dip-font-16" />
            </DipButton>
          </Tooltip>
          <div className={styles.divider} />
          <Tooltip
            title={tempAreaOpen ? '收起临时区' : '展开临时区'}
            getTooltipContainer={node => node.parentElement!}
            placement="right"
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
          <div className={styles.divider} />
          <Tooltip
            title={intl.get('dipChat.startNewConversation')}
            getTooltipContainer={node => node.parentElement!}
            placement="right"
          >
            <DipButton
              color="default"
              variant="link"
              icon={<DipIcon type="icon-dip-chat1" className="dip-font-16" />}
              onClick={startNewConversation}
            />
          </Tooltip>
        </div>
      );
    }
  };

  const startNewConversation = () => {
    cancelChat();
    // 获取当前URL并删除conversation_id参数（不刷新页面）
    const url = new URL(window.location.href);
    if (url.searchParams.has('conversation_id')) {
      url.searchParams.delete('conversation_id');
      // 使用history API更新URL而不刷新页面
      window.history.pushState({}, '', url.toString());
    }
    setDipChatStore({ chatList: [] });
    resetDipChatStore([
      'activeConversationKey',
      'activeChatItemIndex',
      'agentInputParamsFormErrorFields',
      'agentInputParamsFormValue',
      'activeProgressIndex',
      'streamGenerating',
    ]);
    agentInputParamForm?.resetFields();
    setShowBackBottom(false);
  };

  const renderChatList = () => {
    if (chatList.length > 0) {
      return <BubbleList />;
    }
    return <AgentDescription />;
  };

  const renderAgentInputParams = () => {
    const fields = getAgentInputDisplayFields(agentDetails?.config);
    if (fields.length > 0) {
      const renderFields = fields.slice(0, 2);
      return (
        <>
          {renderFields.map((field: any, index: number) => {
            const fieldError = agentInputParamsFormErrorFields.find((item: any) => item.name.includes(field.name));
            return (
              <Tooltip open={fieldError ? undefined : false} title={fieldError?.errors?.join('；')} placement="top">
                <div
                  key={index}
                  className={classNames(styles.inputParamItem, 'dip-ellipsis', {
                    [styles.errorStatus]: !!fieldError,
                  })}
                  onClick={() => {
                    setDipChatStore({ showAgentInputParamsDrawer: true });
                  }}
                  title={field.name}
                >
                  {field.name}
                </div>
              </Tooltip>
            );
          })}
          {fields.length > 2 && (
            <DipButton
              onClick={() => setDipChatStore({ showAgentInputParamsDrawer: true })}
              type="text"
              size="small"
              className="dip-ml-8 dip-font-12 dip-text-color-45"
            >
              {intl.get('dipChat.moreParams')}
            </DipButton>
          )}
        </>
      );
    }
    return '';
  };

  const renderChatArea = () => {
    return (
      <div className="dip-flex-item-full-width dip-h-100 dip-flex">
        <div id="adp-chat-wrapper" className="dip-h-100 dip-overflow-hidden dip-position-r dip-flex-item-full-width">
          <ScrollBarContainer
            onWheelCapture={() => {
              if (chatListAutoScroll) {
                setDipChatStore({
                  chatListAutoScroll: false,
                });
              }
            }}
            ref={chatWrapperRef}
            className={classNames(styles.chatWrapper, 'dip-full dip-pl-24 dip-pr-24')}
          >
            <ResizeObserver
              onResize={() => {
                if (getDipChatStore().chatListAutoScroll) {
                  chatWrapperRef.current!.scrollTop = chatWrapperRef.current!.scrollHeight;
                }
              }}
            >
              <div
                className={classNames(styles.chat, {
                  'dip-flex-column-center': chatList.length === 0,
                })}
              >
                {renderChatList()}
              </div>
            </ResizeObserver>
          </ScrollBarContainer>
          <div className={styles.aiInputWrapper}>
            <div className={styles.aiInput}>
              <div className={styles.aiInputToolBar}>
                <div className="dip-flex-space-between dip-position-r">
                  {streamGenerating && (
                    <div className={styles.loading}>
                      <span className="dip-mr-8">{intl.get('dipChat.taskInProgress')}</span>
                      <ColorLoading />
                    </div>
                  )}
                  <div className={classNames(styles.newConversation)}>
                    <DipButton
                      icon={<DipIcon type="icon-dip-chat1" className="dip-font-16" />}
                      type="text"
                      onClick={startNewConversation}
                    >
                      {intl.get('dipChat.startNewConversation')}
                    </DipButton>
                  </div>
                  <div className="dip-flex-item-full-width dip-position-r dip-ml-2">
                    <div className="dip-full dip-flex-align-center" style={{ justifyContent: 'flex-end' }}>
                      {renderAgentInputParams()}
                    </div>
                    {showBackBottom && (
                      <Tooltip title={intl.get('dipChat.backToBottom')} placement="top">
                        <div
                          className={classNames(styles.backBottom, 'dip-flex-center')}
                          onClick={() => {
                            chatWrapperRef.current!.scrollTop = chatWrapperRef.current!.scrollHeight;
                            if (streamGenerating && !chatListAutoScroll) {
                              setDipChatStore({
                                chatListAutoScroll: true,
                              });
                            }
                          }}
                        >
                          <ArrowDownOutlined />
                        </div>
                      </Tooltip>
                    )}
                  </div>
                </div>
              </div>
              <AiInput
                // loading={streamGenerating || activeConversation?.status === 'processing'}
                loading={streamGenerating}
                value={aiInputValue}
                ref={aiInputRef}
                onSubmit={onSubmit}
                onChange={value => {
                  setDipChatStore({
                    aiInputValue: value,
                  });
                }}
                placeholder={intl.get('dipChat.placeholder')}
                onCancel={stopChat}
                agentAppType={agentAppType}
                agentConfig={{
                  ...agentDetails.config,
                  debug,
                }}
                autoSize={{ minRows: 1, maxRows: 6 }}
                tempFileList={tempFileList}
              />
            </div>
          </div>
        </div>
        <AgentInputParamDrawer />
      </div>
    );
  };

  const renderPreviewFile = () => {
    if (previewFile) {
      return (
        <div style={{ width: '50%' }}>
          <FilePreview
            file={previewFile}
            onClose={() => {
              setDipChatStore({ previewFile: undefined });
            }}
          />
        </div>
      );
    }
  };

  if (debug) {
    if (!showDebuggerArea) {
      return (
        <div className="dip-full dip-flex-column">
          <div className="dip-pl-24 dip-pr-24 dip-pt-8 dip-pb-8">
            <div className="dip-flex-align-center dip-full">
              <span className="dip-font-weight-700 dip-font-16">{intl.get('dipChat.previewAndDebug')}</span>
              <Button
                className="dip-ml-12"
                onClick={() => setDipChatStore({ showDebuggerArea: true })}
                size="small"
                type="text"
                icon={<MenuFoldOutlined />}
              />
            </div>
          </div>
          <div className="dip-flex-item-full-height dip-flex">{renderChatArea()}</div>
        </div>
      );
    }
    return (
      <div className={styles.debugger}>
        <div
          className="dip-flex-item-full-width"
          onClick={() => {
            setDipChatStore({ showDebuggerArea: !showDebuggerArea });
          }}
        />
        <GradientContainer className={styles.debuggerContent}>
          <Splitter>
            <Splitter.Panel>
              <div className="dip-full dip-flex-column dip-overflowY-hidden dip-bg-white">
                <div className="dip-pl-12 dip-pr-12 dip-pt-8 dip-pb-8">
                  <div className="dip-flex-align-center">
                    <Button
                      onClick={() => setDipChatStore({ showDebuggerArea: false })}
                      size="small"
                      type="text"
                      icon={<DipIcon type="icon-dip-close" />}
                    />
                    <span className="dip-font-16 dip-ml-12">{intl.get('dipChat.debugger')}</span>
                  </div>
                </div>
                <div className="dip-flex-item-full-height">
                  <DebuggerProcess />
                </div>
              </div>
            </Splitter.Panel>
            <Splitter.Panel>
              <div className="dip-full dip-flex-column">
                <div className="dip-pl-24 dip-pr-24 dip-pt-8 dip-pb-8">
                  <div className="dip-flex-align-center dip-full">
                    <span className="dip-font-weight-700 dip-font-16">{intl.get('dipChat.previewAndDebug')}</span>
                    <Button
                      className="dip-ml-12"
                      onClick={() => setDipChatStore({ showDebuggerArea: false })}
                      size="small"
                      type="text"
                      icon={<MenuUnfoldOutlined />}
                    />
                  </div>
                </div>
                <div className="dip-flex-item-full-height dip-flex">{renderChatArea()}</div>
              </div>
            </Splitter.Panel>
          </Splitter>
        </GradientContainer>
      </div>
    );
  }

  return (
    <div className={classNames(styles['chat-container'], 'dip-flex dip-full')}>
      <div
        ref={conversationWrapperRef}
        className={classNames(styles.conversation, {
          [styles.conversationCollapsed]: conversationCollapsed,
        })}
      >
        <ConversationList startNewConversation={startNewConversation} />
      </div>
      <GradientContainer
        className={classNames(styles.rightWrapper, 'dip-flex-item-full-width dip-h-100 dip-flex', {
          [styles.mr240]: !conversationCollapsed,
          [styles.mr68]: conversationCollapsed,
        })}
      >
        {renderCollapseBtn()}
        <div className="dip-flex-item-full-width dip-h-100 dip-flex">
          <Splitter
            className={styles.split}
            onResize={sizes => {
              setSize(sizes[0]);
            }}
          >
            <Splitter.Panel min={200} size={previewFile || !tempAreaOpen ? 0 : size}>
              <div
                className={classNames('dip-full dip-pt-8 dip-pb-8', {
                  'dip-pl-8': !conversationCollapsed,
                })}
              >
                <TempArea />
              </div>
            </Splitter.Panel>
            <Splitter.Panel>
              <div className="dip-full dip-flex">
                {renderPreviewFile()}
                {renderChatArea()}
                <RightSideBar />
              </div>
            </Splitter.Panel>
          </Splitter>
        </div>
      </GradientContainer>
    </div>
  );
};

export default DipChat;
