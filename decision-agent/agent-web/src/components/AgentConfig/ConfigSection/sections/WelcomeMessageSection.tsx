import React, { useState, useEffect, useRef } from 'react';
import { Button, Input, message } from 'antd';
import intl from 'react-intl-universal';
import { useAgentConfig } from '../../AgentConfigContext';
import { streamingOutHttp } from '@/utils/http';
import { EventSourceMessage } from '@microsoft/fetch-event-source';
import WelcomeIcon from '@/assets/icons/welcome.svg';
import AiIcon from '@/assets/icons/ai-generate.svg';
import SectionPanel from '../../common/SectionPanel';
import styles from '../ConfigSection.module.less';

const { TextArea } = Input;

const WelcomeMessageSection: React.FC = () => {
  const { state, actions } = useAgentConfig();
  const abortControllerRef = useRef<AbortController | null>(null);
  const streamTextRef = useRef<string>('');

  // 检查是否可编辑开场白配置
  const canEditOpeningRemarkConfig = actions.canEditField('opening_remark_config');

  // 安全访问配置信息，处理可能的undefined情况
  const openingRemarkConfig = state.config?.opening_remark_config || { type: 'fixed', fixed_opening_remark: '' };
  const welcomeMessageFromState = openingRemarkConfig?.fixed_opening_remark || '';

  const [welcomeMessage, setWelcomeMessage] = useState<string>(welcomeMessageFromState);
  const [generating, setGenerating] = useState<boolean>(false);
  const [isExpanded, setIsExpanded] = useState(false);

  // 处理欢迎信息变更
  const handleWelcomeMessageChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    if (!canEditOpeningRemarkConfig) return;
    const value = e.target.value;
    setWelcomeMessage(value);
    actions.updateWelcomeConfig(value);
  };

  // 处理流式消息
  const handleStreamMessage = (event: EventSourceMessage) => {
    if (event.data) {
      const text = event.data.trim();
      // 跳过特殊标记
      if (text === '#' || text === '[DONE]' || text === '') {
        return;
      }

      // 累积到流式文本
      setWelcomeMessage(prev => {
        // 添加新文本
        let newText = prev + text;

        // 如果接收到换行相关的特殊字符，处理格式化
        if (text.includes('##') || text.includes('-') || text.includes('：')) {
          // 处理标题
          newText = newText.replace(/##([^#\n]+)/g, '\n\n## $1\n');

          // 处理列表项，确保每个列表项前有换行
          newText = newText.replace(/([^\n])-(\s)([^-])/g, '$1\n- $3');

          // 处理冒号后的换行
          newText = newText.replace(/：(\s*)([^\n])/g, '：\n$2');
        }

        // 处理欢迎语和结束语部分
        if (text.includes('👋') || text.includes('！') || text.includes('。')) {
          // 在句号和感叹号后添加换行
          newText = newText.replace(/([！。])\s*([^！。\n])/g, '$1\n$2');
        }

        streamTextRef.current = newText;

        return newText;
      });
    }
  };

  // 流式请求结束或出错的处理
  const handleStreamClose = () => {
    // 最后再设置生成状态为false
    setGenerating(false);
    actions.updateWelcomeConfig(streamTextRef.current);
    // 清理引用
    abortControllerRef.current = null;
  };

  const handleStreamError = (error: any) => {
    setGenerating(false);
    message.error({
      content:
        intl.get('dataAgent.config.generationFailed') + (error?.error || intl.get('dataAgent.config.unknownError')),
      key: 'generateWelcome',
    });

    // 清理引用
    abortControllerRef.current = null;
  };

  // 处理生成请求的中止
  const abortGeneration = () => {
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
      abortControllerRef.current = null;
      setGenerating(false);
      message.info(intl.get('dataAgent.config.generationCancelled'));
    }
  };

  // 在组件卸载时中止请求
  useEffect(() => {
    return () => {
      if (abortControllerRef.current) {
        abortControllerRef.current.abort();
      }
    };
  }, []);

  // AI生成开场白
  const generateWelcomeMessage = () => {
    if (!canEditOpeningRemarkConfig) return;
    if (generating) {
      // 如果正在生成，点击按钮则中止
      abortGeneration();
      return;
    }

    setGenerating(true);
    setWelcomeMessage(''); // 重置流式文本
    streamTextRef.current = '';

    message.loading({ content: intl.get('dataAgent.config.generatingGreeting'), key: 'generateWelcome' });

    // 发起流式请求
    const controller = streamingOutHttp({
      url: '/api/agent-factory/v3/agent/ai-autogen',
      method: 'POST',
      body: {
        params: {
          name: state.name || '',
          profile: state.profile || '',
          skills: [],
          sources: [],
        },
        from: 'opening_remarks',
      },
      onMessage: handleStreamMessage,
      onClose: handleStreamClose,
      onError: handleStreamError,
      onOpen: controller => {
        abortControllerRef.current = controller;
      },
    });

    setIsExpanded(true);

    // 存储controller以便后续可能的中止操作
    abortControllerRef.current = controller;
  };

  return (
    <SectionPanel
      title={intl.get('dataAgent.config.defaultGreeting')}
      rightElement={
        <Button
          icon={<AiIcon />}
          type="text"
          onClick={generateWelcomeMessage}
          className="dip-c-link-75"
          loading={generating}
          disabled={!canEditOpeningRemarkConfig}
        >
          {intl.get('dataAgent.config.AIGenerate')}
        </Button>
      }
      isExpanded={isExpanded}
      onToggle={() => setIsExpanded(!isExpanded)}
      className="dip-pb-0"
      icon={<WelcomeIcon />}
    >
      <div className={styles['welcome-config']}>
        <TextArea
          value={welcomeMessage}
          onChange={handleWelcomeMessageChange}
          placeholder={intl.get('dataAgent.config.enterYourGreeting')}
          autoSize={{ minRows: 6, maxRows: 10 }}
          className={styles['welcome-textarea']}
          disabled={generating || !canEditOpeningRemarkConfig}
          maxLength={500}
        />
      </div>
    </SectionPanel>
  );
};

export default WelcomeMessageSection;
