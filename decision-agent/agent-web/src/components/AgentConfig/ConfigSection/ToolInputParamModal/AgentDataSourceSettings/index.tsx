import { useState, useEffect } from 'react';
import intl from 'react-intl-universal';
import { Radio, Select } from 'antd';
import {
  SkillAgentDataSourceConfig,
  SkillAgentLLMConfig,
  DatasourceConfigTypeEnum,
  DatasourceConfigSpecificInheritEnum,
  LLMConfigTypeEnum,
} from '@/apis/agent-factory/type';

interface AgentDataSourceSettingsProps {
  readonly?: boolean;
  dataSourceConfig: SkillAgentDataSourceConfig;
  llmConfig: SkillAgentLLMConfig;
  onUpdateDataSourceConfig: (dataSourceConfig: SkillAgentDataSourceConfig) => void;
  onUpdateLLMConfig: (llmConfig: SkillAgentLLMConfig) => void;
}

const AgentDataSourceSettings = ({
  readonly = false,
  dataSourceConfig: datasourceConfigFromProps,
  llmConfig: llmConfigFromProps,
  onUpdateDataSourceConfig,
  onUpdateLLMConfig,
}: AgentDataSourceSettingsProps) => {
  const [dataSourceConfig, setDatasourceConfig] = useState<SkillAgentDataSourceConfig>(
    datasourceConfigFromProps || {
      type: DatasourceConfigTypeEnum.SelfConfigured,
    }
  );
  const [llmConfig, setLLMConfig] = useState<SkillAgentLLMConfig>(
    llmConfigFromProps || {
      type: LLMConfigTypeEnum.SelfConfigured,
    }
  );

  useEffect(() => {
    onUpdateDataSourceConfig(dataSourceConfig);
  }, [dataSourceConfig]);

  useEffect(() => {
    onUpdateLLMConfig(llmConfig);
  }, [llmConfig]);

  return (
    <div className="dip-mb-10 dip-mt-14">
      <div className="dip-mb-20">
        <div className="dip-font-weight-700 dip-mb-4">{intl.get('dataAgent.config.dataSourceSettings')}</div>
        <Radio.Group
          className="dip-flex-column dip-ml-10"
          disabled={readonly}
          options={[
            {
              value: DatasourceConfigTypeEnum.SelfConfigured,
              label: intl.get('dataAgent.config.useSelfConfiguredDataSource'),
            },
            { value: DatasourceConfigTypeEnum.InheritMain, label: intl.get('dataAgent.config.inheritMasterAgent') },
          ]}
          value={dataSourceConfig.type}
          onChange={e => {
            const value = e.target.value;
            setDatasourceConfig({
              type: value,
              ...(value === DatasourceConfigTypeEnum.SelfConfigured
                ? {}
                : { specific_inherit: DatasourceConfigSpecificInheritEnum.All }),
            });
          }}
        />
        {dataSourceConfig.type === DatasourceConfigTypeEnum.InheritMain && (
          <Select
            options={[
              {
                label: intl.get('dataAgent.config.inheritOnlyDocDataSource'),
                value: DatasourceConfigSpecificInheritEnum.DocsOnly,
              },
              {
                label: intl.get('dataAgent.config.inheritOnlyGraphDataSource'),
                value: DatasourceConfigSpecificInheritEnum.GraphOnly,
              },
              {
                label: intl.get('dataAgent.config.inheritAllDataSources'),
                value: DatasourceConfigSpecificInheritEnum.All,
              },
            ]}
            disabled={readonly}
            value={dataSourceConfig.specific_inherit}
            style={{ width: 300 }}
            className="dip-mt-4 dip-ml-32"
            onChange={value => {
              setDatasourceConfig(prev => ({
                type: prev.type,
                specific_inherit: value,
              }));
            }}
          />
        )}
      </div>
      <div>
        <div className="dip-font-weight-700 dip-mb-4">{intl.get('dataAgent.config.defaultLLM')}</div>
        <Radio.Group
          className="dip-flex-column dip-ml-10"
          disabled={readonly}
          options={[
            { value: LLMConfigTypeEnum.SelfConfigured, label: intl.get('dataAgent.config.useSelfConfiguredLLM') },
            { value: LLMConfigTypeEnum.InheritMain, label: intl.get('dataAgent.config.inheritMasterAgent') },
          ]}
          defaultValue={llmConfig.type}
          onChange={e => {
            const value = e.target.value;
            setLLMConfig({
              type: value,
            });
          }}
        />
      </div>
    </div>
  );
};

export default AgentDataSourceSettings;
