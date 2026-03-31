import intl from 'react-intl-universal';

// 输入字段类型
export const getInputTypes = () => [
  { value: 'string', label: intl.get('dataAgent.config.string') },
  // { value: 'file', label: intl.get('dataAgent.config.file') },
  { value: 'object', label: intl.get('dataAgent.config.object') },
];

// 默认临时区配置
export const defaultTempZoneConfig = {
  name: '临时上传区',
  max_file_count: 50,
  single_chat_max_select_file_count: 5,
  single_file_size_limit: 100,
  single_file_size_limit_unit: 'MB',
  support_data_type: ['file'],
  allowed_file_categories: ['document'],
  allowed_file_types: ['*'],
  tmp_file_use_type: 'upload',
};
