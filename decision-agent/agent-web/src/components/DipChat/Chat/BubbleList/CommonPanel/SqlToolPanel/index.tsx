import styles from './index.module.less';
import { useDipChatStore } from '@/components/DipChat/store';
import React from 'react';
import DipIcon from '@/components/DipIcon';
import { DipChatItemContentProgressType } from '@/components/DipChat/interface';
import SkillBar from '@/components/DipChat/components/SkillBar';

type SqlToolPanelProps = {
  progressItem: DipChatItemContentProgressType;
  chatItemIndex: number;
  progressIndex: number;
  readOnly: boolean;
};

const SqlToolPanel = ({ progressItem, chatItemIndex, progressIndex, readOnly }: SqlToolPanelProps) => {
  const {
    dipChatStore: { streamGenerating, chatList, activeProgressIndex },
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
        icon={<DipIcon className="dip-font-16" type="icon-dip-color-sql" />}
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

export default SqlToolPanel;
