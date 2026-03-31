import styles from './index.module.less';
import React from 'react';
import { DipChatItemContentProgressType } from '@/components/DipChat/interface';
import _ from 'lodash';
import classNames from 'classnames';
import DipIcon from '@/components/DipIcon';
import ShinyText from '@/components/animation/ShinyText';
import { useDipChatStore } from '@/components/DipChat/store';
import { RightOutlined } from '@ant-design/icons';
import DipButton from '@/components/DipButton';
import intl from 'react-intl-universal';
import { useDeepCompareMemo } from '@/hooks';

type CommonToolPanelProps = {
  progressItem: DipChatItemContentProgressType;
  chatItemIndex: number;
  progressIndex: number;
  readOnly: boolean;
};

const NetSearchToolPanel = ({ progressItem, chatItemIndex, progressIndex, readOnly }: CommonToolPanelProps) => {
  const {
    dipChatStore: { chatList, streamGenerating, activeProgressIndex },
    openSideBar,
    setDipChatStore,
  } = useDipChatStore();

  const total = useDeepCompareMemo(() => {
    const cites = progressItem?.netSearchResult?.cites || [];
    let total: number = 0;
    cites.forEach((item: any) => {
      if ('children' in item) {
        if (item.children && Array.isArray(item.children)) {
          total += item.children.length;
        }
      } else {
        total += 1;
      }
    });
    return total;
  }, [progressItem?.netSearchResult?.cites]);

  const renderCites = () => {
    const loading = progressItem.status === 'processing' && streamGenerating && chatItemIndex === chatList.length - 1;
    return (
      <div
        className={classNames(styles.title, 'dip-flex-align-center dip-flex-item-full-width', {
          [styles.active]: progressIndex === activeProgressIndex,
        })}
      >
        <span className="dip-flex-align-center dip-flex-item-full-width">
          <DipIcon className="dip-font-16" type="icon-dip-net" />
          <ShinyText
            loading={loading}
            className="dip-ml-8 dip-flex-item-full-width dip-ellipsis"
            text={
              loading
                ? intl.get('dipChat.readingDocs', { count: total })
                : intl.get('dipChat.foundDocs', { count: total })
            }
          />
        </span>
        {progressItem.status === 'completed' && (
          <span className="dip-flex-align-center dip-text-color-45 dip-font-12 dip-ml-12 dip-mr-12">
            <span>耗时：{progressItem.consumeTime}s</span>
          </span>
        )}
        {!readOnly && total > 0 && (
          <DipButton
            onClick={() => {
              if (total > 0) {
                openSideBar(chatItemIndex);
                setDipChatStore({
                  activeProgressIndex: progressIndex,
                });
              }
            }}
            type="text"
            icon={<RightOutlined />}
            iconPosition="end"
          >
            {intl.get('dipChat.view')}
          </DipButton>
        )}
      </div>
    );
  };
  return total > 0 && <div className={styles.container}>{renderCites()}</div>;
};

export default NetSearchToolPanel;
