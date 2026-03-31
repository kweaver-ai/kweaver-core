import React, { useState, useEffect } from 'react';
import { Select } from 'antd';
import { useAgentConfig } from '../../AgentConfigContext';
import outputIcon from '@/assets/icons/base-info.png';
import SectionPanel from '../../common/SectionPanel';

const { Option } = Select;

interface SelectOption {
  value: string;
  label: string;
}

const OutputFormatSection: React.FC = () => {
  const { state, actions } = useAgentConfig();
  const [selectedFormat, setSelectedFormat] = useState<string>(state.config.output?.default_format || 'markdown');

  // 输出格式选项
  const formatOptions: SelectOption[] = [
    { value: 'markdown', label: 'Markdown' },
    { value: 'json', label: 'JSON' },
  ];

  // 处理输出格式变更
  const handleFormatChange = (value: string) => {
    setSelectedFormat(value);
    // actions.updateOutput(value);
  };

  // 当Context状态更新时，同步本地状态
  useEffect(() => {
    setSelectedFormat(state.config.output?.default_format || 'markdown');
  }, [state.config.output?.default_format]);

  return (
    <SectionPanel
      title="默认输出格式"
      description="设置智能体回复的默认格式"
      icon={<img src={outputIcon} alt="默认输出格式" width={20} height={20} />}
      className="dip-border-b"
    >
      <div className="format-config">
        <Select style={{ width: '100%' }} value={selectedFormat} onChange={handleFormatChange}>
          {formatOptions.map(option => (
            <Option key={option.value} value={option.value}>
              {option.label}
            </Option>
          ))}
        </Select>
      </div>
    </SectionPanel>
  );
};

export default OutputFormatSection;
