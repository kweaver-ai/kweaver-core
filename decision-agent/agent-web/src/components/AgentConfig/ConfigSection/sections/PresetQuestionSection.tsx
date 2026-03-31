import React, { useState, useEffect, useRef } from 'react';
import intl from 'react-intl-universal';
import { Button, Input, message } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import { useAgentConfig } from '../../AgentConfigContext';
import DipIcon from '@/components/DipIcon';
import { aiGeneratePresetQuestions } from '@/apis/agent-factory';
import { AgentGenerationParams } from '@/apis/agent-factory/type';
import PresetQuestionIcon from '@/assets/icons/preset-question.svg';
import AiIcon from '@/assets/icons/ai-generate.svg';
import SectionPanel from '../../common/SectionPanel';
import styles from '../ConfigSection.module.less';

// 预设问题接口定义
interface PresetQuestion {
  question: string;
}

const PresetQuestionSection: React.FC = () => {
  const { state, actions } = useAgentConfig();
  const abortControllerRef = useRef<AbortController | null>(null);

  // 检查是否可编辑预设问题配置
  const canEditPresetQuestions = actions.canEditField('preset_questions');

  // 安全访问配置信息，处理可能的undefined情况
  const questionsFromState = state.config?.preset_questions?.length ? state.config?.preset_questions : [];

  const [presetQuestions, setPresetQuestions] = useState<PresetQuestion[]>(questionsFromState);
  const [generatingQuestions, setGeneratingQuestions] = useState<boolean>(false);
  const [isExpanded, setIsExpanded] = useState(false);

  // 添加空的预设问题
  const addPresetQuestion = () => {
    if (!canEditPresetQuestions) return;
    if (presetQuestions.length >= 4) {
      message.warning(intl.get('dataAgent.config.maxPresetQuestionsLimit'));
      return;
    }

    const updatedQuestions = [...presetQuestions, { question: '' }];
    setPresetQuestions(updatedQuestions);
    setIsExpanded(true);
  };

  // 更新预设问题
  const updatePresetQuestion = (index: number, value: string) => {
    if (!canEditPresetQuestions) return;
    const updatedQuestions = [...presetQuestions];
    updatedQuestions[index] = { question: value };
    setPresetQuestions(updatedQuestions);
    actions.updatePresetQuestions(updatedQuestions.filter(q => q.question.trim() !== ''));
  };

  // 移除预设问题
  const removePresetQuestion = (index: number) => {
    if (!canEditPresetQuestions) return;
    const updatedQuestions = [...presetQuestions];
    updatedQuestions.splice(index, 1);
    setPresetQuestions(updatedQuestions);
    actions.updatePresetQuestions(updatedQuestions.filter(q => q.question.trim() !== ''));
  };

  // AI生成预设问题
  const generatePresetQuestions = async () => {
    if (!canEditPresetQuestions) return;
    if (generatingQuestions) {
      return;
    }

    if (presetQuestions.length >= 4) {
      message.warning(intl.get('dataAgent.config.maxPresetQuestionsLimit'));
      return;
    }

    setGeneratingQuestions(true);
    message.loading({ content: intl.get('dataAgent.config.generatingPresetQuestions'), key: 'generateQuestions' });

    try {
      // 准备API请求参数
      const params: AgentGenerationParams = {
        params: {
          name: state.name || '',
          description: state.description || '',
          skills: [],
          sources: [],
        },
        from: 'preset_question',
      };

      // 调用API获取预设问题
      const questions = await aiGeneratePresetQuestions(params);

      // 转换并添加到现有问题中
      const questionObjects = questions.map(q => ({ question: q }));

      // 计算可以添加的问题数量
      const availableSlots = 4 - presetQuestions.length;
      const questionsToAdd = questionObjects.slice(0, availableSlots);

      // 添加生成的问题到现有问题
      const updatedQuestions = [...presetQuestions, ...questionsToAdd];
      setPresetQuestions(updatedQuestions);
      actions.updatePresetQuestions(updatedQuestions);

      message.success({ content: intl.get('dataAgent.config.presetQuestionsGenerated'), key: 'generateQuestions' });
    } catch (error) {
      message.error({
        content:
          intl.get('dataAgent.config.generationFailed') + (error?.error || intl.get('dataAgent.config.unknownError')),
        key: 'generateQuestions',
      });
    } finally {
      setGeneratingQuestions(false);
      setIsExpanded(true);
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

  return (
    <SectionPanel
      title={intl.get('dataAgent.config.presetQuestionIndex', { index: presetQuestions.length })}
      isExpanded={isExpanded}
      onToggle={() => setIsExpanded(!isExpanded)}
      rightElement={
        <div className="dip-flex-align-center">
          <Button
            icon={<AiIcon />}
            type="text"
            onClick={generatePresetQuestions}
            className="dip-c-link-75"
            loading={generatingQuestions}
            disabled={!canEditPresetQuestions}
          >
            {intl.get('dataAgent.config.AIGenerate')}
          </Button>
          {presetQuestions.length < 4 && (
            <Button
              icon={<PlusOutlined />}
              type="text"
              onClick={addPresetQuestion}
              disabled={presetQuestions.length >= 4 || !canEditPresetQuestions}
              className="dip-c-link-75"
            >
              {intl.get('dataAgent.config.add')}
            </Button>
          )}
        </div>
      }
      icon={<PresetQuestionIcon className="dip-flex-shrink-0" />}
    >
      <div className={styles['welcome-config']}>
        <div className={styles['preset-questions']}>
          {/* 显示预设问题输入框 */}
          <div className="dip-flex-column-center">
            {presetQuestions.map((item, index) => (
              <div key={index} className="dip-w-100 dip-mb-12">
                <Input
                  value={item.question}
                  onChange={e => updatePresetQuestion(index, e.target.value)}
                  placeholder={intl.get('dataAgent.config.inputGreetingPrompt')}
                  maxLength={50}
                  disabled={!canEditPresetQuestions}
                />
                <Button
                  type="text"
                  className={styles['delete-button']}
                  onClick={() => removePresetQuestion(index)}
                  disabled={!canEditPresetQuestions}
                  icon={<DipIcon type="icon-dip-trash" />}
                />
              </div>
            ))}
          </div>
        </div>
      </div>
    </SectionPanel>
  );
};

export default PresetQuestionSection;
