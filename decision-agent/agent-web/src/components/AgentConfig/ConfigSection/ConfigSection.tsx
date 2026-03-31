import React from 'react';
import classNames from 'classnames';
import intl from 'react-intl-universal';
import { useAgentConfig } from '../AgentConfigContext';
// 导入拆分出的各个子组件
import ProductSection from './sections/ProductSection';
import InputConfig from './InputConfig/InputConfig';
import KnowledgeSource from './sections/KnowledgeSource';
import SkillsSection from './sections/SkillsSection';
import ModelConfigSection from './sections/ModelConfigSection';
import WelcomeMessageSection from './sections/WelcomeMessageSection';
import PresetQuestionSection from './sections/PresetQuestionSection';
import SystemAgentSection from './sections/SystemAgentSection';
import LongTermMemorySection from './sections/LongTermMemorySection';
import RelatedQuestionSection from './sections/RelatedQuestionSection';
import PlanModeSection from './sections/PlanModeSection';
import styles from './ConfigSection.module.less';

// 主组件 - 仅负责组织子组件和主要布局
const ConfigSection: React.FC = () => {
  const { state, actions } = useAgentConfig();

  return (
    <>
      <ProductSection />
      <SystemAgentSection />
      <InputConfig />
      <KnowledgeSource />
      <SkillsSection state={state} actions={actions} />
      <ModelConfigSection />
      <LongTermMemorySection />
      <RelatedQuestionSection />
      <PlanModeSection />
      {/* <DataFlowSection /> */}
      {/* <OutputFormatSection /> */}
      <div className="dip-flex-align-center dip-mr-24 dip-ml-24">
        <div className={classNames(styles['divider-with-text'], 'dip-text-color-45')}>
          {intl.get('dataAgent.config.otherExperiences')}
        </div>
        <div className="dip-border-b dip-w-100"></div>
      </div>
      <WelcomeMessageSection />
      <PresetQuestionSection />
    </>
  );
};

export default ConfigSection;
