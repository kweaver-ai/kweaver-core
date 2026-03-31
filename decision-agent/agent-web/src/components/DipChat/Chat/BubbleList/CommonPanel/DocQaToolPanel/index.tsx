import styles from './index.module.less';
import React from 'react';
import { DipChatItemContentProgressType } from '@/components/DipChat/interface';
import DipIcon from '@/components/DipIcon';
import { useDipChatStore } from '@/components/DipChat/store';
import SkillBar from '@/components/DipChat/components/SkillBar';

type CommonToolPanelProps = {
  progressItem: DipChatItemContentProgressType;
  chatItemIndex: number;
  progressIndex: number;
  readOnly: boolean;
};

const DocQaToolPanel = ({ progressItem, chatItemIndex, progressIndex, readOnly }: CommonToolPanelProps) => {
  const {
    dipChatStore: { chatList, streamGenerating, activeProgressIndex },
    openSideBar,
    setDipChatStore,
  } = useDipChatStore();

  const view = () => {
    openSideBar(chatItemIndex);
    setDipChatStore({
      activeProgressIndex: progressIndex,
    });
  };

  const loading = streamGenerating && chatItemIndex === chatList.length - 1;

  return (
    <div className={styles.container}>
      <SkillBar
        icon={<DipIcon className="dip-font-16" type="icon-dip-color-docQa" />}
        title={progressItem.title}
        status={progressItem.status}
        readOnly={readOnly}
        loading={loading}
        consumeTime={progressItem.consumeTime}
        onView={view}
        active={progressIndex === activeProgressIndex}
      />
    </div>
  );
};

export default DocQaToolPanel;
