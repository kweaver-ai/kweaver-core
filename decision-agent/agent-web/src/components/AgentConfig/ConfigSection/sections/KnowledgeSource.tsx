import React, { useEffect, useRef, useState } from 'react';
import classNames from 'classnames';
import _, { uniqBy } from 'lodash';
import intl from 'react-intl-universal';
import { Collapse, Button, Space, List, Tooltip } from 'antd';
import { QuestionCircleOutlined, PlusOutlined, DownOutlined } from '@ant-design/icons';
import KnEntryIcon from '@/assets/icons/kn-entry.svg';
import NetworkIcon from '@/assets/icons/network.svg';
import KnowledgeIcon from '@/assets/icons/knowledge.svg';
import { getMetricInfoByIds, getDataDictInfoByIds } from '@/apis/data-model';
import { getKnExperimentDetailsById } from '@/apis/knowledge-data';
import { useAgentConfig } from '../../AgentConfigContext';
import KnEntrySelector from '@/components/KnEntrySelector';
import DipIcon from '@/components/DipIcon';
import MetricSelector from '@/components/MetricSelector';
import SectionPanel from '../../common/SectionPanel';
import styles from '../ConfigSection.module.less';
import KnowledgeNetworkExperimentalSelect from '@/components/KnowledgeNetworkExperimentalSelect';
import { useDeepCompareEffect } from '@/hooks';

const { Panel } = Collapse;

interface KnowledgeSourceProps {
  style?: React.CSSProperties;
}

const KnowledgeSource: React.FC<KnowledgeSourceProps> = () => {
  const { state, actions } = useAgentConfig();

  const canEditDataSourceKgExperiment = actions.canEditField('data_source.knowledge_network');
  const canEditDataSourceMetric = actions.canEditField('data_source.metric');
  const canEditDataSourceKnEntry = actions.canEditField('data_source.kn_entry');

  const [activeKey, setActiveKey] = useState<string[]>([]);
  const [isExpanded, setIsExpanded] = useState(true);

  const [metricSelectorVisible, setMetricSelectorVisible] = useState<boolean>(false);
  const [metricNames, setMetricNames] = useState<Record<string, string>>({});
  const invalidMetricIds = useRef<string[]>([]);
  const metricNamesRef = useRef<Record<string, string>>({});

  const [knEntrySelectorVisible, setKnEntrySelectorVisible] = useState<boolean>(false);
  const [knEntryNames, setKnEntryNames] = useState<Record<string, string>>({});
  const inValidKnEntryIds = useRef<string[]>([]);
  const knEntryNamesRef = useRef<Record<string, string>>({});

  const [knowledgeNetworkExperimentalOpen, setKnowledgeNetworkExperimentalOpen] = useState(false);
  const [knExperimentNames, setKnExperimentNames] = useState<Record<string, string>>({});
  const invalidKgExperimentIds = useRef<string[]>([]);

  const getDataSourceItemId = (item: any, idKey: 'metric_model_id' | 'kn_entry_id') => item?.[idKey] || item?.id;

  const normalizeDataSourceItems = (items: any[], idKey: 'metric_model_id' | 'kn_entry_id') =>
    items.map(item => {
      const normalizedId = getDataSourceItemId(item, idKey);
      if (!normalizedId || item?.[idKey]) return item;

      const normalizedItem = { ...item, [idKey]: normalizedId };
      delete normalizedItem.id;
      return normalizedItem;
    });

  const deleteDataSource = (
    item: any,
    options: {
      canEdit: boolean;
      statePath: 'metric' | 'kn_entry';
      idKey: 'metric_model_id' | 'kn_entry_id';
      invalidIds: React.MutableRefObject<string[]>;
      updateInvalidFn: (key: 'kg-experiment' | 'metric' | 'kn_entry', invalid: boolean) => void;
    }
  ) => {
    const { canEdit, statePath, idKey, invalidIds, updateInvalidFn } = options;
    if (!canEdit) return;

    const currentDataSource = state.config?.data_source?.[statePath] || [];
    const currentDataSourceNormalized = normalizeDataSourceItems(currentDataSource, idKey);
    const itemId = getDataSourceItemId(item, idKey);
    const newDataSource = currentDataSourceNormalized.filter(
      dataSourceItem => getDataSourceItemId(dataSourceItem, idKey) !== itemId
    );
    actions.updateSpecificField(`config.data_source.${statePath}`, newDataSource);

    invalidIds.current = invalidIds.current.filter(id => id !== itemId);
    updateInvalidFn(statePath, invalidIds.current.length > 0);
  };

  const addDataSource = (
    item: any | any[],
    options: {
      selectorVisibleSetter: React.Dispatch<React.SetStateAction<boolean>>;
      statePath: 'metric' | 'kn_entry';
      activeKeyValue: 'metric' | 'kn_entry';
      idKey: 'metric_model_id' | 'kn_entry_id';
      namesSetter?: React.Dispatch<React.SetStateAction<Record<string, string>>>;
    }
  ) => {
    const { selectorVisibleSetter, statePath, activeKeyValue, idKey, namesSetter } = options;
    const data = normalizeDataSourceItems(Array.isArray(item) ? item : [item], idKey);
    const currentDataSource = normalizeDataSourceItems(state.config?.data_source?.[statePath] || [], idKey);
    const newDataSource = uniqBy([...currentDataSource, ...data], idKey);

    actions.updateSpecificField(`config.data_source.${statePath}`, newDataSource);
    selectorVisibleSetter(false);

    if (namesSetter) {
      const namesMap = data.reduce((prev, current) => ({ ...prev, [current[idKey]]: current.name }), {});
      namesSetter(prev => ({ ...prev, ...namesMap }));
    }

    setActiveKey(prev => {
      if (prev.includes(activeKeyValue)) return prev;
      return [...prev, activeKeyValue];
    });
  };

  const handleDeleteKnowledgeNetworkExperimental = (id: string) => {
    const newNetworkList = _.filter(
      state.config?.data_source?.knowledge_network,
      item => item.knowledge_network_id !== id
    );
    actions.updateMultipleFields({
      'config.data_source.knowledge_network': newNetworkList,
    });
  };

  const handleSaveKNExperiment = (selectedNet: any) => {
    const networkList: any[] = _.cloneDeep(state.config?.data_source?.knowledge_network) || [];
    actions.updateMultipleFields({
      'config.data_source.knowledge_network': [
        ...networkList,
        {
          knowledge_network_id: selectedNet.value,
        },
      ],
    });
    setKnowledgeNetworkExperimentalOpen(false);
    setActiveKey(prev => {
      if (prev.includes('knowledgeNetwork-experimental')) return prev;
      return [...prev, 'knowledgeNetwork-experimental'];
    });
  };

  const customExpandIcon = ({ isActive }: { isActive?: boolean }) => (
    <DownOutlined style={{ fontSize: '12px', color: '#8c8c8c' }} rotate={isActive ? 180 : 0} />
  );

  const renderKnowledgeNetworkExperimentalHeader = () => (
    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', width: '100%' }}>
      <Space>
        <span>{intl.get('dataAgent.config.businessKnowledgeNetwork')}</span>
      </Space>
      {(!state.config?.data_source?.knowledge_network ||
        state.config?.data_source?.knowledge_network?.length === 0) && (
        <Button
          type="text"
          icon={<PlusOutlined />}
          size="small"
          disabled={!canEditDataSourceKgExperiment}
          onClick={e => {
            e.stopPropagation();
            if (canEditDataSourceKgExperiment) {
              setKnowledgeNetworkExperimentalOpen(true);
            }
          }}
          className="dip-c-link-75"
        >
          {intl.get('dataAgent.config.add')}
        </Button>
      )}
    </div>
  );

  const renderMetricHeader = () => (
    <div className="dip-w-100 dip-flex-space-between">
      <Space>
        <span>{intl.get('dataAgent.indicator')}</span>
      </Space>
      <Button
        type="text"
        icon={<PlusOutlined />}
        size="small"
        disabled={!canEditDataSourceMetric}
        onClick={e => {
          e.stopPropagation();
          if (canEditDataSourceMetric) {
            setMetricSelectorVisible(true);
          }
        }}
        className="dip-c-link-75"
      >
        {intl.get('dataAgent.config.add')}
      </Button>
    </div>
  );

  useEffect(() => {
    const getMetricName = async () => {
      const metric = state.config?.data_source?.metric;
      if (!metric?.length) return;

      const metricIds = metric.map(({ metric_model_id }) => metric_model_id);
      try {
        const result = await getMetricInfoByIds({ ids: metricIds });
        const metricNameMap: Record<string, string> = result.reduce((prev, { id, name }) => {
          return {
            ...prev,
            [id]: name,
          };
        }, {});

        setMetricNames(metricNameMap);
        metricIds.forEach(id => {
          if (!metricNameMap[id]) {
            invalidMetricIds.current = [...invalidMetricIds.current, id];
          }
        });
        actions.updateDataSourceInvalid('metric', invalidMetricIds.current.length > 0);
      } catch {
        if (metricIds.length === 1) {
          invalidMetricIds.current = metricIds;
          actions.updateDataSourceInvalid('metric', invalidMetricIds.current.length > 0);
          return;
        }

        metricIds.forEach(async id => {
          try {
            const [{ name }] = await getMetricInfoByIds({ ids: [id] });
            metricNamesRef.current = { ...metricNamesRef.current, [id]: name };
          } catch {
            invalidMetricIds.current = [...invalidMetricIds.current, id];
          } finally {
            setMetricNames(metricNamesRef.current);
            actions.updateDataSourceInvalid('metric', invalidMetricIds.current.length > 0);
          }
        });
      }
    };

    getMetricName();
  }, []);

  useEffect(() => {
    const getKnEntryName = async () => {
      const knEntry = state.config?.data_source?.kn_entry;
      if (!knEntry?.length) return;

      const knEntryIds = knEntry.map(({ kn_entry_id }) => kn_entry_id);
      try {
        const result = await getDataDictInfoByIds(knEntryIds);
        const knEntryNameMap: Record<string, string> = result.reduce((prev, { id, name }) => {
          return {
            ...prev,
            [id]: name,
          };
        }, {});

        setKnEntryNames(knEntryNameMap);
        knEntryIds.forEach(id => {
          if (!knEntryNameMap[id]) {
            inValidKnEntryIds.current = [...inValidKnEntryIds.current, id];
          }
        });
        actions.updateDataSourceInvalid('kn_entry', inValidKnEntryIds.current.length > 0);
      } catch {
        if (knEntryIds.length === 1) {
          inValidKnEntryIds.current = knEntryIds;
          actions.updateDataSourceInvalid('kn_entry', inValidKnEntryIds.current.length > 0);
          return;
        }

        knEntryIds.forEach(async id => {
          try {
            const [{ name }] = await getDataDictInfoByIds([id]);
            knEntryNamesRef.current = { ...knEntryNamesRef.current, [id]: name };
          } catch {
            inValidKnEntryIds.current = [...inValidKnEntryIds.current, id];
          } finally {
            setKnEntryNames(knEntryNamesRef.current);
            actions.updateDataSourceInvalid('kn_entry', inValidKnEntryIds.current.length > 0);
          }
        });
      }
    };

    getKnEntryName();
  }, []);

  useDeepCompareEffect(() => {
    const getData = async () => {
      const knData = state.config?.data_source?.knowledge_network || [];
      const knIds = knData.map(item => item.knowledge_network_id);
      if (knIds.length > 0) {
        const res = await getKnExperimentDetailsById(knIds[0]);
        if (res) {
          setKnExperimentNames({ [res.id]: res.name });
          invalidKgExperimentIds.current = [];
          actions.updateDataSourceInvalid('kg-experiment', false);
        } else {
          setKnExperimentNames({});
          invalidKgExperimentIds.current = knIds;
          actions.updateDataSourceInvalid('kg-experiment', invalidKgExperimentIds.current.length > 0);
        }
      } else {
        setKnExperimentNames({});
      }
    };

    getData();
    return () => {
      invalidKgExperimentIds.current = [];
      actions.updateDataSourceInvalid('kg-experiment', false);
    };
  }, [state.config?.data_source?.knowledge_network]);

  useEffect(() => {
    actions.updateDataSourceNameMapping('metric', metricNames);
  }, [metricNames]);

  useEffect(() => {
    actions.updateDataSourceNameMapping('kn_entry', knEntryNames);
  }, [knEntryNames]);

  return (
    <SectionPanel
      title={
        <>
          <div>{intl.get('dataAgent.config.knowledgeSource')}</div>
          <Tooltip title={intl.get('dataAgent.config.knowledgeSourceTip')}>
            <QuestionCircleOutlined className="dip-font-14" />
          </Tooltip>
        </>
      }
      isExpanded={isExpanded}
      onToggle={() => setIsExpanded(!isExpanded)}
      icon={<KnowledgeIcon />}
      description={intl.get('dataAgent.config.knowledgeDes')}
      className={'dip-border-line-b'}
    >
      <div className={styles['knowledge-config']}>
        <Collapse
          ghost
          expandIcon={customExpandIcon}
          activeKey={activeKey}
          style={{ background: 'transparent' }}
          onChange={(key: string | string[]) => {
            setActiveKey(key);
          }}
        >
          <Panel
            header={renderKnowledgeNetworkExperimentalHeader()}
            key="knowledgeNetwork-experimental"
            style={{ border: 'none' }}
          >
            <List
              size="small"
              dataSource={state.config?.data_source?.knowledge_network || []}
              renderItem={(item: any) => {
                const isInvalid = invalidKgExperimentIds.current.includes(item.knowledge_network_id);
                const name = knExperimentNames[item.knowledge_network_id] || '---';

                return (
                  <List.Item
                    key={item.knowledge_network_id}
                    style={{
                      padding: '8px 0',
                      border: 'none',
                    }}
                    actions={[
                      <Button
                        type="text"
                        icon={<DipIcon type="icon-dip-trash" />}
                        size="small"
                        disabled={!canEditDataSourceKgExperiment}
                        onClick={() => {
                          if (canEditDataSourceKgExperiment) {
                            handleDeleteKnowledgeNetworkExperimental(item.knowledge_network_id);
                          }
                        }}
                        className="dip-c-subtext"
                      />,
                    ].filter(Boolean)}
                  >
                    <div className="dip-flex-align-center dip-ml-24 dip-gap-8 dip-overflow-hidden">
                      <NetworkIcon style={{ color: '#1890ff', width: 16, height: 16 }} className="dip-flex-shrink-0" />
                      <span
                        title={name}
                        className={classNames('dip-ellipsis', {
                          'dip-text-color-error': isInvalid,
                        })}
                      >
                        {isInvalid ? '---' : name}
                      </span>
                    </div>
                  </List.Item>
                );
              }}
            />
          </Panel>

          <Panel header={renderMetricHeader()} key="metric" style={{ border: 'none' }}>
            <List
              size="small"
              dataSource={state.config?.data_source?.metric || []}
              renderItem={(item: any) => {
                const isInvalid = invalidMetricIds.current.includes(item.metric_model_id);
                const name = item.name || metricNames[item.metric_model_id];

                return (
                  <List.Item
                    key={item.metric_model_id}
                    className="dip-pt-8 dip-pb-8 dip-pl-0 dip-pr-0"
                    style={{ border: 'none' }}
                    actions={[
                      <Button
                        type="text"
                        icon={<DipIcon type="icon-dip-trash" />}
                        size="small"
                        disabled={!canEditDataSourceMetric}
                        onClick={() => {
                          deleteDataSource(item, {
                            canEdit: canEditDataSourceMetric,
                            statePath: 'metric',
                            idKey: 'metric_model_id',
                            invalidIds: invalidMetricIds,
                            updateInvalidFn: actions.updateDataSourceInvalid,
                          });
                        }}
                        className="dip-c-subtext"
                      />,
                    ].filter(Boolean)}
                  >
                    <div className="dip-flex-align-center dip-ml-24 dip-gap-8 dip-overflow-hidden">
                      <DipIcon type={'icon-dip-color-zhibiaometirc'} className="dip-font-16" />
                      <span
                        title={name}
                        className={classNames('dip-ellipsis', {
                          'dip-text-color-error': isInvalid,
                        })}
                      >
                        {isInvalid ? '---' : name}
                      </span>
                    </div>
                  </List.Item>
                );
              }}
            />
          </Panel>

          <Panel
            header={
              <div className="dip-w-100 dip-flex-space-between">
                <Space>
                  <span>{intl.get('dataAgent.knowledgeEntry')}</span>
                </Space>
                <Button
                  type="text"
                  icon={<PlusOutlined />}
                  size="small"
                  disabled={!canEditDataSourceKnEntry}
                  onClick={e => {
                    e.stopPropagation();
                    if (canEditDataSourceKnEntry) {
                      setKnEntrySelectorVisible(true);
                    }
                  }}
                  className="dip-c-link-75"
                >
                  {intl.get('dataAgent.config.add')}
                </Button>
              </div>
            }
            key="kn_entry"
            style={{ border: 'none' }}
          >
            <List
              size="small"
              dataSource={state.config?.data_source?.kn_entry || []}
              renderItem={(item: any) => {
                const isInvalid = inValidKnEntryIds.current.includes(item.kn_entry_id);
                const name = item.name || knEntryNames[item.kn_entry_id];

                return (
                  <List.Item
                    key={item.kn_entry_id}
                    className="dip-pt-8 dip-pb-8 dip-pl-0 dip-pr-0"
                    style={{ border: 'none' }}
                    actions={[
                      <Button
                        type="text"
                        icon={<DipIcon type="icon-dip-trash" />}
                        size="small"
                        disabled={!canEditDataSourceKnEntry}
                        onClick={() => {
                          deleteDataSource(item, {
                            canEdit: canEditDataSourceKnEntry,
                            statePath: 'kn_entry',
                            idKey: 'kn_entry_id',
                            invalidIds: inValidKnEntryIds,
                            updateInvalidFn: actions.updateDataSourceInvalid,
                          });
                        }}
                        className="dip-c-subtext"
                      />,
                    ].filter(Boolean)}
                  >
                    <div className="dip-flex-align-center dip-ml-24 dip-gap-8 dip-overflow-hidden">
                      <KnEntryIcon className="dip-flex dip-font-16 dip-flex-shrink-0" />
                      <span
                        title={name}
                        className={classNames('dip-ellipsis', {
                          'dip-text-color-error': isInvalid,
                        })}
                      >
                        {isInvalid ? '---' : name}
                      </span>
                    </div>
                  </List.Item>
                );
              }}
            />
          </Panel>
        </Collapse>

        {metricSelectorVisible && (
          <MetricSelector
            onConfirm={metric =>
              addDataSource(metric, {
                selectorVisibleSetter: setMetricSelectorVisible,
                statePath: 'metric',
                activeKeyValue: 'metric',
                idKey: 'metric_model_id',
                namesSetter: setMetricNames,
              })
            }
            onCancel={() => setMetricSelectorVisible(false)}
          />
        )}

        {knEntrySelectorVisible && (
          <KnEntrySelector
            onCancel={() => setKnEntrySelectorVisible(false)}
            onConfirm={knEntry =>
              addDataSource(knEntry, {
                selectorVisibleSetter: setKnEntrySelectorVisible,
                statePath: 'kn_entry',
                activeKeyValue: 'kn_entry',
                idKey: 'kn_entry_id',
                namesSetter: setKnEntryNames,
              })
            }
          />
        )}

        <KnowledgeNetworkExperimentalSelect
          open={knowledgeNetworkExperimentalOpen}
          onClose={() => {
            setKnowledgeNetworkExperimentalOpen(false);
          }}
          handleSaveKNExperiment={handleSaveKNExperiment}
        />
      </div>
    </SectionPanel>
  );
};

export default KnowledgeSource;
