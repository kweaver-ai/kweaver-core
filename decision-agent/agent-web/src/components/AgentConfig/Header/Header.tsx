import React, { useState, useCallback, useMemo } from 'react';
import { useSearchParams, useNavigate, useBlocker } from 'react-router-dom';
import classNames from 'classnames';
import intl from 'react-intl-universal';
import { LeftOutlined } from '@ant-design/icons';
import { Button, Modal } from 'antd';
import { useMicroWidgetProps, useNavigationBlocker } from '@/hooks';
import { useAgentConfig } from '../AgentConfigContext';
import { PublishSettingsModal, PublishModeEnum } from '@/components/AgentPublish';
import AgentIcon from '@/components/AgentIcon';
import styles from './Header.module.less';
import { getParam } from '@/utils/handle-function';
import qs from 'qs';

interface HeaderProps {
  // 是否是模板的编辑页面
  isEditTemplate?: boolean;
  showDebugger: boolean;
  onToggleDebugger: (show: boolean) => void;
  activeTab: string;
  onTabChange: (tab: string) => void;
  refreshTraceAnalysis: () => void;
  userPermissions: any;
  isSkillAgent: boolean;
}

const Header: React.FC<HeaderProps> = ({
  isEditTemplate = false,
  showDebugger,
  onToggleDebugger,
  activeTab,
  onTabChange,
  refreshTraceAnalysis,
  userPermissions,
  isSkillAgent,
}) => {
  const microWidgetProps = useMicroWidgetProps();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const { state, actions } = useAgentConfig();
  const [isSaving, setIsSaving] = useState(false);
  const [modal, contextHolder] = Modal.useModal();
  const [publishModalVisible, setPublishModalVisible] = useState(false);

  const isTemplate = useMemo(() => searchParams.get('mode') === 'editTemplate', [searchParams]);

  const handleSave = useCallback(
    async ({ onSuccess }: { onSuccess?: any } = {}) => {
      setIsSaving(true);
      try {
        const result = await actions.saveAgent({ isEditTemplate, onSuccess });
        if (!result) {
          // 保存失败，由saveAgent内部处理错误消息
          console.log('保存失败');
          return false;
        } else {
          return result || true;
        }
      } catch (error) {
        console.error('保存Agent出错:', error);
        return false;
      } finally {
        setIsSaving(false);
      }
    },
    [actions]
  );

  const handleNavigation = useCallback(
    (blocker: ReturnType<typeof useBlocker>) => {
      modal.confirm({
        centered: true,
        title: intl.get('global.existTitle'),
        content: intl.get('benchmarkTask.exitContent'),
        okText: intl.get('prompt.saveClose'),
        onOk: async () => {
          // 调用保存接口。保存并关闭，成功后无需做操作
          if (await handleSave({ onSuccess: () => {} })) {
            blocker.proceed!();
          } else {
            blocker.reset!();
          }
        },
        cancelText: intl.get('prompt.abandon'),
        onCancel: () => {
          blocker.proceed!();
        },
      });
    },
    [handleSave, modal]
  );

  useNavigationBlocker({
    shouldBlock: !!state.isDirty,
    handleNavigation,
  });

  const redirect = () => {
    // 优先使用url的redirect进行跳转-山东大数据局
    const redirectUrl = searchParams.get('redirect');
    if (redirectUrl) {
      location.replace(redirectUrl);
    } else {
      const filterParams = getParam('filterParams');
      const mode = getParam('mode');
      const templateId = getParam('templateId');
      if (mode === 'createAgent' && !!templateId) {
        // 说明是模版创建agent跳转过来的
        microWidgetProps?.history.navigateToMicroWidget({
          name: 'my-agent-list',
          path: '/',
        });
      } else {
        navigate(filterParams ? `/?filterParams=${filterParams}` : '/');
      }
    }
  };

  const handlePublishClick = async () => {
    let agentId = state.id;
    if (!agentId) {
      // If agent hasn't been saved yet, save it first
      const id = await actions.saveAgent({ isEditTemplate });
      if (!id) {
        return;
      }
      agentId = id;
    } else {
      // 如果agentId存在，则先更新
      const id = await actions.saveAgent({ isEditTemplate });
      if (!id) return;
    }

    setPublishModalVisible(true);
  };

  const handlePublishSubmit = async () => {
    setPublishModalVisible(false);
    redirect();
  };

  const handlePublishCancel = () => {
    setPublishModalVisible(false);
  };

  const handleToggleDebugger = async () => {
    if (showDebugger === false && !state.id) {
      const res = await handleSave();
      if (!res) {
        return;
      }
    }
    onToggleDebugger(!showDebugger);
  };

  const renderRightBtn = () => {
    if (activeTab === 'trace-analysis') {
      return (
        <div style={{ minWidth: 360, display: 'flex', gap: 12, justifyContent: 'flex-end' }}>
          <Button type="primary" onClick={refreshTraceAnalysis}>
            刷新
          </Button>
        </div>
      );
    }
    return (
      <div style={{ minWidth: 360, display: 'flex', gap: 12, justifyContent: 'flex-end' }}>
        {/* 编辑模板页面 屏蔽调试 */}
        <Button
          className="dip-min-width-72"
          onClick={async () => {
            const res = await handleSave();
            if (res) {
              const params: any = {
                id: res,
                version: 'v0',
                agentAppType: 'common',
              };
              window.open(`${microWidgetProps.history.getBasePath}/usage?${qs.stringify(params)}`);
            }
          }}
          disabled={isSaving}
        >
          去使用
        </Button>
        {!isTemplate && (
          <Button className="dip-min-width-72" onClick={handleToggleDebugger}>
            {showDebugger ? intl.get('dataAgent.config.closeDebugger') : intl.get('dataAgent.config.openDebugger')}
          </Button>
        )}
        <Button
          className="dip-min-width-72"
          onClick={() => {
            handleSave();
          }}
          disabled={isSaving}
        >
          {isSaving ? intl.get('dataAgent.config.saving') : intl.get('dataAgent.config.save')}
        </Button>
        {userPermissions.hasPublishPermission && (
          <Button type="primary" className="dip-min-width-72" onClick={handlePublishClick}>
            {intl.get('dataAgent.publish')}
          </Button>
        )}
      </div>
    );
  };

  const renderTab = () => {
    // 有查看轨迹分析权限
    // 不是模板
    // 只要发布过 不是发布为技能Agent
    if (!isTemplate && userPermissions.hasTraceAnalysisPermission && state.is_published && !isSkillAgent) {
      return (
        <div className="dip-flex-align-center">
          <div
            className={classNames(styles.tabItem, {
              [styles.activeTab]: activeTab === 'agent-config',
            })}
            onClick={() => onTabChange('agent-config')}
          >
            {intl.get('dataAgent.config.agentConfig')}
          </div>
          <div
            className={classNames(styles.tabItem, {
              [styles.activeTab]: activeTab === 'trace-analysis',
              [styles.disabledTab]: !state.id,
            })}
            onClick={() => {
              if (state.id) {
                onTabChange('trace-analysis');
              }
            }}
          >
            {intl.get('dataAgent.config.traceAnalysis')}
          </div>
        </div>
      );
    }
  };

  return (
    <>
      <header
        className={classNames(
          styles.appHeader,
          'dip-flex-space-between',
          'dip-bg-white',
          'dip-pr-24',
          'dip-pl-24',
          'dip-border-line-b'
        )}
      >
        <div className={styles.headerLeft}>
          <LeftOutlined className={styles.backIcon} onClick={redirect} />
          <AgentIcon avatar_type={state.avatar_type} avatar={state.avatar} name={state.name} size={24} />
          <div>{state.name || intl.get('dataAgent.config.untitled')}</div>
        </div>
        {/*{renderTab()}*/}
        <div className={styles.headerRight}>
          {process.env.NODE_ENV === 'development' && (
            <Button
              className="dip-min-width-72"
              onClick={() => {
                console.log(state);
              }}
            >
              查看配置Store
            </Button>
          )}
          {renderRightBtn()}
        </div>
      </header>

      {/* 发布设置弹窗 */}
      {publishModalVisible && (
        <PublishSettingsModal
          mode={isEditTemplate ? PublishModeEnum.PublishTemplate : PublishModeEnum.PublishAgent}
          agent={state}
          onCancel={handlePublishCancel}
          onOk={handlePublishSubmit}
        />
      )}
      {contextHolder}
    </>
  );
};

export default Header;
