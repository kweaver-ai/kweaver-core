import { useState, useRef, useEffect, memo, useMemo } from 'react';
import intl from 'react-intl-universal';
import { Table } from 'antd';
import { getKnExperimentDetailsById } from '@/apis/knowledge-data';
import { getMetricInfoByIds, getDataDictInfoByIds } from '@/apis/data-model';
import { AgentDetailType } from '@/apis/agent-factory/type';

interface Props {
  config: AgentDetailType | null;
}

const DataSourceViewerTable = memo(({ config }: Props) => {
  const metricNamesRef = useRef<Record<string, string>>({});
  const knEntryNamesRef = useRef<Record<string, string>>({});

  const [metricNames, setMetricNames] = useState<Record<string, string>>({});
  const [knEntryNames, setKnEntryNames] = useState<Record<string, string>>({});
  const [knExperimentNames, setKnExperimentNames] = useState<Record<string, string>>({});

  const dataSource = useMemo(() => {
    const arr: any[] = [];

    if (config?.config?.data_source?.knowledge_network) {
      config.config.data_source.knowledge_network.forEach((network: any) => {
        arr.push({
          key: network.knowledge_network_id,
          type: `${intl.get('dataAgent.config.businessKnowledgeNetwork')}(实验版)`,
          name: knExperimentNames[network.knowledge_network_id] || '---',
          fields: ['---'],
        });
      });
    }

    if (config?.config?.data_source?.metric?.length) {
      const fields = config.config.data_source.metric.map(
        ({ metric_model_id }) => metricNames[metric_model_id] || '---'
      );
      arr.push({
        key: 'metric',
        type: intl.get('dataAgent.indicator'),
        name: intl.get('dataAgent.indicator'),
        fields,
      });
    }

    if (config?.config?.data_source?.kn_entry?.length) {
      const fields = config.config.data_source.kn_entry.map(({ kn_entry_id }) => knEntryNames[kn_entry_id] || '---');
      arr.push({
        key: 'kn_entry',
        type: intl.get('dataAgent.knowledgeEntry'),
        name: intl.get('dataAgent.knowledgeEntry'),
        fields,
      });
    }

    return arr;
  }, [config, metricNames, knEntryNames, knExperimentNames]);

  const columns = useMemo(
    () => [
      {
        title: intl.get('dataAgent.config.type'),
        dataIndex: 'type',
        key: 'type',
        ellipsis: true,
      },
      {
        title: intl.get('dataAgent.config.knowledgeSourceName'),
        dataIndex: 'name',
        key: 'name',
        ellipsis: true,
      },
      {
        title: intl.get('dataAgent.config.searchScope'),
        dataIndex: 'fields',
        key: 'fields',
        ellipsis: true,
        render: (fields: any) => {
          if (!fields || !Array.isArray(fields)) return '';
          return fields.map((field: any) => field.name || field).join(', ');
        },
      },
    ],
    []
  );

  const fetchMetricNames = async (metricIds: string[]) => {
    try {
      const result = await getMetricInfoByIds({ ids: metricIds });
      const names: Record<string, string> = result.reduce((prev, { id, name }) => {
        return {
          ...prev,
          [id]: name,
        };
      }, {});
      setMetricNames(names);
    } catch {
      if (metricIds.length === 1) return;
      metricIds.forEach(async id => {
        try {
          const [{ name }] = await getMetricInfoByIds({ ids: [id] });
          metricNamesRef.current = { ...metricNamesRef.current, [id]: name };
          setMetricNames(metricNamesRef.current);
        } catch {}
      });
    }
  };

  const fetchKnExperimentNames = async (knIds: string[]) => {
    if (knIds.length > 0) {
      const res = await getKnExperimentDetailsById(knIds[0]);
      if (res) {
        setKnExperimentNames({ [res.id]: res.name });
      }
    }
  };

  const fetchKnEntryNames = async (knEntryIds: string[]) => {
    try {
      const result = await getDataDictInfoByIds(knEntryIds);
      const names: Record<string, string> = result.reduce((prev, { id, name }) => {
        return {
          ...prev,
          [id]: name,
        };
      }, {});
      setKnEntryNames(names);
    } catch {
      if (knEntryIds.length === 1) return;
      knEntryIds.forEach(async id => {
        try {
          const [{ name }] = await getDataDictInfoByIds([id]);
          knEntryNamesRef.current = { ...knEntryNamesRef.current, [id]: name };
          setKnEntryNames(knEntryNamesRef.current);
        } catch {}
      });
    }
  };

  const getDataSourceInfo = async (agent: AgentDetailType | null) => {
    if (agent?.config?.data_source?.metric) {
      const metricIds = agent.config.data_source.metric.map(({ metric_model_id }) => metric_model_id);
      if (metricIds.length) {
        fetchMetricNames(metricIds);
      }
    }

    if (agent?.config?.data_source?.kn_entry) {
      const knEntryIds = agent.config.data_source.kn_entry.map(({ kn_entry_id }) => kn_entry_id);
      if (knEntryIds.length) {
        fetchKnEntryNames(knEntryIds);
      }
    }

    if (agent?.config?.data_source?.knowledge_network) {
      const knIds = agent.config.data_source.knowledge_network.map(item => item.knowledge_network_id);
      if (knIds.length > 0) {
        await fetchKnExperimentNames(knIds);
      }
    }
  };

  useEffect(() => {
    getDataSourceInfo(config);
  }, [config]);

  return dataSource.length === 0 ? (
    <div style={{ textAlign: 'center', padding: '40px 0', color: '#999' }}>
      {intl.get('dataAgent.noKnowledgeSourceConfig')}
    </div>
  ) : (
    <Table columns={columns} dataSource={dataSource} pagination={false} size="small" />
  );
});

export default DataSourceViewerTable;
