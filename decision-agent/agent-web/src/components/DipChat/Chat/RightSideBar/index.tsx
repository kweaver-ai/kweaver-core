import styles from './index.module.less';
import classNames from 'classnames';
import { useDipChatStore } from '@/components/DipChat/store';
import { useState } from 'react';
import NetSearchList from './NetSearchList';
import _ from 'lodash';
import FadeInFromRight from '@/components/animation/FadeInFromRight';
import FilePreview from '../../components/FilePreview';
import SqlToolSide from './SqlToolSide';
import ChartToolSide from './ChartToolSide';
import CodeToolSide from './CodeToolSide';
import NGQLToolSide from './NGQLToolSide';
import CommonToolSide from './CommonToolSide';
import DocQAToolSide from './DocQAToolSide';
import MetricToolSide from './MetricToolSide';
import type { PreviewFileType } from '@/components/DipChat/interface';

const RightSideBar = () => {
  const {
    dipChatStore: { activeChatItemIndex, chatList, activeProgressIndex },
  } = useDipChatStore();
  const chatItem = chatList[activeChatItemIndex];
  const activeContent = chatItem?.content || {};
  const isOpen = activeChatItemIndex !== -1;
  const [previewFile, setPreviewFile] = useState<PreviewFileType>();

  const renderContent = () => {
    if (!_.isEmpty(activeContent.cites)) {
      return <NetSearchList />;
    }
    if (activeProgressIndex !== -1) {
      const activeProgressItem = chatItem.content.progress[activeProgressIndex];
      const renderSide = () => {
        if (activeProgressItem.type === 'common_tool') {
          return <CommonToolSide />;
        }
        if (activeProgressItem.type === 'metric_tool') {
          return <MetricToolSide />;
        }
        if (activeProgressItem.type === 'sql_tool') {
          return <SqlToolSide />;
        }
        if (activeProgressItem.type === 'chart_tool') {
          return <ChartToolSide />;
        }
        if (activeProgressItem.type === 'code_tool') {
          return <CodeToolSide />;
        }
        if (activeProgressItem.type === 'ngql_tool') {
          return <NGQLToolSide />;
        }
        if (activeProgressItem.type === 'docQa_tool') {
          return (
            <DocQAToolSide
              onPreview={(file: PreviewFileType) => {
                setPreviewFile(file);
              }}
            />
          );
        }
        if (activeProgressItem.type === 'net_search_tool') {
          return <NetSearchList citesList={activeProgressItem.netSearchResult?.cites} />;
        }
      };
      return <div className={styles.bg}>{activeProgressItem && renderSide()}</div>;
    }
  };

  const renderPreviewFile = () => {
    if (!_.isEmpty(previewFile)) {
      return (
        <FadeInFromRight className={styles.previewFile}>
          <FilePreview
            file={previewFile}
            onClose={() => {
              setPreviewFile(undefined);
            }}
          />
        </FadeInFromRight>
      );
    }
  };

  return (
    <div
      className={classNames(styles.container, {
        [styles.open]: isOpen,
      })}
    >
      {isOpen && (
        <div className="dip-full dip-position-r">
          {renderContent()}
          {renderPreviewFile()}
        </div>
      )}
    </div>
  );
};

export default RightSideBar;
