import intl from 'react-intl-universal';

/**
 * 大模型默认配置
 */
export const getLLM_CONFIG_PARAM = () => [
  {
    key: 'temperature',
    label: intl.get('prompt.temperature'),
    tip: intl.get('prompt.temperatureTip'),
    step: 0.1,
    precision: 2,
  },
  {
    key: 'top_p',
    label: intl.get('prompt.top_p'),
    tip: intl.get('prompt.top_pTip'),
    step: 0.1,
    precision: 2,
  },
  {
    key: 'max_tokens',
    label: intl.get('prompt.max_tokens'),
    tip: intl.get('prompt.max_tokensTip'),
    step: 1,
    precision: 0,
  },
  {
    key: 'top_k',
    label: intl.get('prompt.top_k'),
    tip: intl.get('prompt.top_k_tip'),
    step: 1,
    precision: 0,
  },
  {
    key: 'presence_penalty',
    label: intl.get('prompt.presence_penalty'),
    tip: intl.get('prompt.presence_penaltyTip'),
    step: 0.1,
    precision: 2,
  },
  {
    key: 'frequency_penalty',
    label: intl.get('prompt.frequency_penalty'),
    tip: intl.get('prompt.frequency_penaltyTip'),
    step: 0.1,
    precision: 2,
  },
];
