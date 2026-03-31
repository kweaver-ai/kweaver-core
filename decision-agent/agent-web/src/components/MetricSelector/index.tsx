import { useState } from 'react';
import intl from 'react-intl-universal';
import { Modal, Button } from 'antd';
import classNames from 'classnames';
import { type MetricModelType, type MetricModalGroupType } from '@/apis/data-model';
import {
  MetricSelectorContext,
  type MetricSelectorState,
  type MetricSelectorContextType,
} from './store';
import { MetricConstants } from './types';
import GroupList from './GroupList';
import MetricList from './MetricList';
import styles from './index.module.less';

interface MetricSelectorProps {
  onCancel: () => void;
  onConfirm: (metrics: Array<MetricModelType>) => void;
}

const MetricSelector = ({ onCancel, onConfirm }: MetricSelectorProps) => {
  const [metricSelectorStore, setMetricSelectorStore] = useState<MetricSelectorState>({
    allMetricGroup: {
      id: '__all',
      name: intl.get('dataAgent.allIndicatorModels'),
    } as MetricModalGroupType,
    selectedGroup: undefined,
    selectedMetrics: [],
  });

  const updateAllMetricGroup = (updates: Record<MetricConstants.MetricModelCount, number>) => {
    setMetricSelectorStore(prev => ({
      ...prev,
      allMetricGroup: {
        ...prev.allMetricGroup,
        ...updates,
      },
    }));
  };

  const setSelectedGroup = (group: MetricModalGroupType) => {
    setMetricSelectorStore(prev => ({ ...prev, selectedGroup: group }));
  };

  const setSelectedMetrics = (metrics: MetricModelType[]) => {
    setMetricSelectorStore(prev => ({ ...prev, selectedMetrics: metrics }));
  };

  const appendSelectedMetric = (metric: MetricModelType) => {
    setMetricSelectorStore(prev => ({ ...prev, selectedMetrics: [...prev.selectedMetrics, metric] }));
  };

  const removeSelectedMetric = (metric: MetricModelType) => {
    setMetricSelectorStore(prev => ({
      ...prev,
      selectedMetrics: prev.selectedMetrics.filter(item => item.id !== metric.id),
    }));
  };

  const contextValue: MetricSelectorContextType = {
    metricSelectorStore,
    setMetricSelectorStore,
    isAllMetricGroupSetted: MetricConstants.MetricModelCount in metricSelectorStore.allMetricGroup,
    updateAllMetricGroup,
    setSelectedGroup,
    setSelectedMetrics,
    appendSelectedMetric,
    removeSelectedMetric,
  };

  return (
    <MetricSelectorContext.Provider value={contextValue}>
      <Modal
        title={intl.get('dataAgent.selectIndicator')}
        open
        centered
        destroyOnHidden
        maskClosable={false}
        width={800}
        onCancel={onCancel}
        footer={[
          <Button
            key="submit"
            type="primary"
            className="dip-min-width-72"
            disabled={!metricSelectorStore.selectedMetrics.length}
            onClick={() => onConfirm(metricSelectorStore.selectedMetrics)}
          >
            {intl.get('dataAgent.ok')}
          </Button>,
          <Button key="cancel" className="dip-min-width-72" onClick={onCancel}>
            {intl.get('dataAgent.cancel')}
          </Button>,
        ]}
      >
        <div className={classNames('dip-flex', styles['container'])}>
          <div className={styles['group']}>
            <GroupList />
          </div>
          <div className="dip-flex-item-full-width">
            <MetricList />
          </div>
        </div>
      </Modal>
    </MetricSelectorContext.Provider>
  );
};

export default MetricSelector;
