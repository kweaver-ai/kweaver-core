import React, { createContext, useContext } from 'react';
import { noop } from 'lodash';
import { type MetricModalGroupType, type MetricModelType } from '@/apis/data-model';
import { MetricConstants } from './types';

export interface MetricSelectorState {
  allMetricGroup: MetricModalGroupType;
  selectedGroup: MetricModalGroupType | undefined;
  selectedMetrics: MetricModelType[];
}

export interface MetricSelectorContextType {
  metricSelectorStore: MetricSelectorState;
  setMetricSelectorStore: React.Dispatch<React.SetStateAction<MetricSelectorState>>;
  isAllMetricGroupSetted: boolean;
  setSelectedGroup: (group: MetricModalGroupType) => void;
  setSelectedMetrics: (metrics: MetricModelType[]) => void;
  appendSelectedMetric: (metric: MetricModelType) => void;
  removeSelectedMetric: (metric: MetricModelType) => void;
  updateAllMetricGroup: (updates: Record<MetricConstants.MetricModelCount, number>) => void;
}

export const MetricSelectorContext = createContext<MetricSelectorContextType>({
  metricSelectorStore: {
    allMetricGroup: { id: '__all', name: '' } as MetricModalGroupType,
    selectedGroup: undefined,
    selectedMetrics: [],
  },
  setMetricSelectorStore: noop,
  isAllMetricGroupSetted: false,
  setSelectedGroup: noop,
  setSelectedMetrics: noop,
  appendSelectedMetric: noop,
  removeSelectedMetric: noop,
  updateAllMetricGroup: noop,
});

export const useMetricSelectorStore = () => useContext(MetricSelectorContext);
