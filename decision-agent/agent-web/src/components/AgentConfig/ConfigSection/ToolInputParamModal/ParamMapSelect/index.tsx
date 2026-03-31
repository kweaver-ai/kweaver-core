import { Select, Input } from 'antd';
import intl from 'react-intl-universal';
import AgentVarSelect from '@/components/AgentVarSelect';
import AdLLMSelect from '@/components/AdLLMSelect';
import './style.less';
const ParamMapSelect = ({ value, onChange, varOptions, disabled, toolType, readonly = false }: any) => {
  const renderMapType = () => {
    switch (value.map_type) {
      case 'fixedValue':
        return (
          <Input
            style={{ background: disabled ? '#f5f5f5' : '#fff' }}
            variant="borderless"
            disabled={disabled}
            value={value.map_value}
            onChange={e => {
              value.map_value = e.target.value;
              onChange(value);
            }}
            placeholder={intl.get('global.pleaseEnter')}
          />
        );
      case 'var':
        return (
          <AgentVarSelect
            className="dip-w-100"
            disabled={disabled}
            value={value.map_value}
            onChange={data => {
              value.map_value = data;
              onChange(value);
            }}
            options={varOptions}
            placeholder={intl.get('global.pleaseSelect')}
          />
        );
      case 'model':
        return (
          <AdLLMSelect
            readonly={readonly}
            bordered={false}
            testModel
            disabled={disabled}
            queryOnFocus
            defaultSelectedFirstOption
            placeholder={intl.get('agentCommonConfig.llm.bigModelPlaceholder')}
            value={value.map_value}
            onChange={llm => {
              value.map_value = llm;
              onChange(value);
            }}
          />
        );
      case 'auto':
        return (
          <Input
            style={{ background: '#f5f5f5' }}
            variant="borderless"
            disabled
            placeholder={intl.get('agentCommonConfig.llm.modelAutoPlaceholder')}
          />
        );
    }
  };

  return (
    <div className="ParamMapSelect dip-border-form-item dip-flex-align-center">
      <Select
        style={{ width: 100, background: disabled ? '#f5f5f5' : '#fff' }}
        disabled={disabled}
        variant="borderless"
        value={value.map_type}
        onChange={data => {
          value.map_type = data;
          value.map_value = undefined;
          onChange(value);
        }}
        options={[
          { label: intl.get('agentCommonConfig.llm.fixedValue'), value: 'fixedValue' },
          { label: intl.get('agentCommonConfig.llm.useVariable'), value: 'var' },
          ...(toolType !== 'agent' ? [{ label: intl.get('agentCommonConfig.llm.selectModel'), value: 'model' }] : []),
          { label: intl.get('agentCommonConfig.llm.modelAuto'), value: 'auto' },
        ]}
      />
      <div className="dip-flex-item-full-width dip-border-l">{renderMapType()}</div>
    </div>
  );
};

export default ParamMapSelect;
