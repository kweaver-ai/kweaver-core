import styles from './index.module.less';
import { useDipChatStore } from '@/components/DipChat/store';
import classNames from 'classnames';
import React, { useState } from 'react';
import DipIcon from '@/components/DipIcon';
import { Spin } from 'antd';
import { CloseCircleFilled, LoadingOutlined, RightOutlined } from '@ant-design/icons';
import DipButton from '@/components/DipButton';
import { DipChatItemContentProgressType } from '@/components/DipChat/interface';
import DipEcharts from '@/components/DipEcharts';
import _ from 'lodash';
import intl from 'react-intl-universal';
import SkillBar from '@/components/DipChat/components/SkillBar';
type ChartToolPanelProps = {
  progressItem: DipChatItemContentProgressType;
  chatItemIndex: number;
  progressIndex: number;
  readOnly: boolean;
};
const ChartToolPanel = ({ progressItem, chatItemIndex, progressIndex, readOnly }: ChartToolPanelProps) => {
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

  const echartsOptions = progressItem.chartResult?.echartsOptions || {};

  return (
    <div className={styles.container}>
      <SkillBar
        className="dip-mb-8"
        icon={<DipIcon className="dip-font-16" type="icon-dip-color-echarts" />}
        title={progressItem.title}
        status={progressItem.status}
        readOnly={readOnly}
        loading={loading}
        consumeTime={progressItem.consumeTime}
        onView={view}
        active={progressIndex === activeProgressIndex}
      />
      {!_.isEmpty(echartsOptions) && (
        <div className="dip-mb-8">
          <div className={styles.charts}>
            <DipEcharts options={echartsOptions} />
          </div>
        </div>
      )}
    </div>
  );
};

export default ChartToolPanel;
