import styles from './index.module.less';
import { useDipChatStore } from '@/components/DipChat/store';
import React from 'react';
import DipIcon from '@/components/DipIcon';
import SkillBar from '@/components/DipChat/components/SkillBar';

const NGQLToolPanel = ({ progressItem, chatItemIndex, progressIndex, readOnly }: any) => {
  const {
    dipChatStore: { streamGenerating, chatList, activeProgressIndex },
    openSideBar,
    setDipChatStore,
  } = useDipChatStore();
  const loading = streamGenerating && chatItemIndex === chatList.length - 1;
  const view = () => {
    openSideBar(chatItemIndex);
    setDipChatStore({
      activeProgressIndex: progressIndex,
    });
  };
  return (
    <div className={styles.container}>
      <SkillBar
        icon={<DipIcon className="dip-font-16" type="icon-dip-color-ngl" />}
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

export default NGQLToolPanel;
